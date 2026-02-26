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
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

type SimpleToolCommand struct {
	*cmds.CommandDescription
	tool        string
	defaultArgs map[string]any
}

type SimpleToolSettings struct {
	ArgsJSON  string `glazed:"args-json"`
	Socket    string `glazed:"socket-path"`
	TimeoutMS int    `glazed:"timeout-ms"`
	TabID     int64  `glazed:"tab-id"`
	WindowID  int64  `glazed:"window-id"`
}

func NewSimpleToolCommand(name string, short string, tool string, defaultArgs map[string]any) (*SimpleToolCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
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
		cmds.WithLong(fmt.Sprintf("Executes the %s tool through the local surf host.", tool)),
		cmds.WithFlags(
			fields.New("args-json", fields.TypeString, fields.WithDefault("{}"), fields.WithHelp("Additional tool args as JSON object")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	defaults := map[string]any{}
	for k, v := range defaultArgs {
		defaults[k] = v
	}

	return &SimpleToolCommand{
		CommandDescription: desc,
		tool:               tool,
		defaultArgs:        defaults,
	}, nil
}

func (c *SimpleToolCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &SimpleToolSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	args := map[string]any{}
	if err := json.Unmarshal([]byte(s.ArgsJSON), &args); err != nil {
		return fmt.Errorf("invalid --args-json: %w", err)
	}
	for k, v := range c.defaultArgs {
		if _, ok := args[k]; !ok {
			args[k] = v
		}
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)

	var tabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	resp, err := ExecuteTool(ctx, client, c.tool, args, tabID, windowID)
	if err != nil {
		return err
	}

	for _, row := range ToolResponseToRows(resp) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
