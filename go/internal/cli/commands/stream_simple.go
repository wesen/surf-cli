package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

type StreamCommand struct {
	*cmds.CommandDescription
	streamType string
}

type StreamSettings struct {
	OptionsJSON string `glazed:"options-json"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	DurationSec int    `glazed:"duration-sec"`
	TabID       int64  `glazed:"tab-id"`
}

func NewStreamCommand(name, short, streamType string) (*StreamCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		name,
		cmds.WithShort(short),
		cmds.WithLong(fmt.Sprintf("Starts %s streaming through local host and emits rows until timeout.", streamType)),
		cmds.WithFlags(
			fields.New("options-json", fields.TypeString, fields.WithDefault("{}"), fields.WithHelp("Stream options as JSON object")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Connect timeout in milliseconds")),
			fields.New("duration-sec", fields.TypeInteger, fields.WithDefault(30), fields.WithHelp("How long to stream before stopping")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &StreamCommand{CommandDescription: desc, streamType: streamType}, nil
}

func (c *StreamCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &StreamSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	options := map[string]any{}
	if err := json.Unmarshal([]byte(s.OptionsJSON), &options); err != nil {
		return fmt.Errorf("invalid --options-json: %w", err)
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)

	streamCtx := ctx
	cancel := func() {}
	if s.DurationSec > 0 {
		streamCtx, cancel = context.WithTimeout(ctx, time.Duration(s.DurationSec)*time.Second)
	}
	defer cancel()

	var tabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}

	return client.Stream(streamCtx, c.streamType, options, tabID, func(msg map[string]any) error {
		eventType, _ := msg["type"].(string)
		row := types.NewRow(
			types.MRP("stream_type", c.streamType),
			types.MRP("event_type", eventType),
			types.MRP("event", msg),
		)
		return gp.AddRow(streamCtx, row)
	})
}
