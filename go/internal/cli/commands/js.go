package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
	"github.com/pkg/errors"
)

type JSCommand struct {
	*cmds.CommandDescription
}

type JSSettings struct {
	Code        string `glazed:"code"`
	File        string `glazed:"file"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

func NewJSCommand() (*JSCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"js",
		cmds.WithShort("Execute JavaScript in the current page"),
		cmds.WithLong("Execute JavaScript through the local surf host. Use 'return' in the script when you want a value back."),
		cmds.WithFlags(
			fields.New("file", fields.TypeString, fields.WithHelp("Read JavaScript from file instead of inline code")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithArguments(
			fields.New("code", fields.TypeString, fields.WithHelp("Inline JavaScript to execute")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &JSCommand{CommandDescription: desc}, nil
}

func buildJSArgs(s *JSSettings) (map[string]any, error) {
	code := s.Code
	if s.File != "" {
		if s.Code != "" {
			return nil, fmt.Errorf("provide inline code or --file, not both")
		}
		b, err := os.ReadFile(s.File)
		if err != nil {
			return nil, errors.Wrap(err, "read --file")
		}
		code = string(b)
	}
	if code == "" {
		return nil, fmt.Errorf("code required unless --file is set")
	}
	return map[string]any{"code": code}, nil
}

func (c *JSCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &JSSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	args, err := buildJSArgs(s)
	if err != nil {
		return err
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)
	client.Debug = s.DebugSocket

	var tabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	resp, err := ExecuteTool(ctx, client, "js", args, tabID, windowID)
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
