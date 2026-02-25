package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

type NavigateCommand struct {
	*cmds.CommandDescription
}

type NavigateSettings struct {
	URL       string `glazed:"url"`
	Socket    string `glazed:"socket-path"`
	TimeoutMS int    `glazed:"timeout-ms"`
	TabID     int64  `glazed:"tab-id"`
	WindowID  int64  `glazed:"window-id"`
}

func NewNavigateCommand() (*NavigateCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"navigate",
		cmds.WithShort("Navigate current tab to a URL"),
		cmds.WithLong("Navigate current tab/window to a URL through the local surf host."),
		cmds.WithFlags(
			fields.New("url", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Target URL")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &NavigateCommand{CommandDescription: desc}, nil
}

func (c *NavigateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &NavigateSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if s.URL == "" {
		return fmt.Errorf("--url is required")
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)

	args := map[string]any{"url": s.URL}
	var tabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	resp, err := ExecuteTool(ctx, client, "navigate", args, tabID, windowID)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, ToolResponseToRow("navigate", resp))
}
