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
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/nicobailon/surf-cli/gohost/internal/cli/transport"
	"github.com/nicobailon/surf-cli/gohost/internal/host/config"
)

type ChatGPTCommand struct {
	*cmds.CommandDescription
}

type ChatGPTSettings struct {
	Query       string `glazed:"query"`
	Model       string `glazed:"model"`
	File        string `glazed:"file"`
	WithPage    bool   `glazed:"with-page"`
	ListModels  bool   `glazed:"list-models"`
	TimeoutSec  int    `glazed:"timeout"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

func NewChatGPTCommand() (*ChatGPTCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"chatgpt",
		cmds.WithShort("Send a prompt to ChatGPT using your browser session"),
		cmds.WithLong("Uses the Go native host ChatGPT provider through the local surf socket. Requires an active ChatGPT browser login."),
		cmds.WithFlags(
			fields.New("model", fields.TypeString, fields.WithHelp("ChatGPT model to select")),
			fields.New("file", fields.TypeString, fields.WithHelp("File to attach before sending the prompt")),
			fields.New("with-page", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Include current page content in the prompt")),
			fields.New("list-models", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List available ChatGPT models without sending a prompt")),
			fields.New("timeout", fields.TypeInteger, fields.WithDefault(2700), fields.WithHelp("Request timeout in seconds")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(30000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override for page-context reads")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithArguments(
			fields.New("query", fields.TypeString, fields.WithHelp("Prompt to send to ChatGPT")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &ChatGPTCommand{CommandDescription: desc}, nil
}

func (c *ChatGPTCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &ChatGPTSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if s.Query == "" && !s.ListModels {
		return fmt.Errorf("query required unless --list-models is set")
	}

	toolArgs := map[string]any{}
	if s.Query != "" {
		toolArgs["query"] = s.Query
	}
	if s.Model != "" {
		toolArgs["model"] = s.Model
	}
	if s.File != "" {
		toolArgs["file"] = s.File
	}
	if s.WithPage {
		toolArgs["with-page"] = true
	}
	if s.ListModels {
		toolArgs["list-models"] = true
	}
	if s.TimeoutSec > 0 {
		toolArgs["timeout"] = s.TimeoutSec
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

	resp, err := ExecuteTool(ctx, client, "chatgpt", toolArgs, tabID, windowID)
	if err != nil {
		return err
	}

	rows, err := chatGPTResponseToRows(resp)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func chatGPTResponseToRows(resp map[string]any) ([]types.Row, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}

	parsed := parseResult(resp)
	if dataMap, ok := parsed.Data.(map[string]any); ok {
		if models, ok := dataMap["models"]; ok {
			items, _ := models.([]any)
			selected, _ := dataMap["selected"].(string)
			rows := make([]types.Row, 0, len(items))
			for _, item := range items {
				name, _ := item.(string)
				rows = append(rows, types.NewRow(
					types.MRP("model", name),
					types.MRP("selected", selected == name),
				))
			}
			if len(rows) > 0 {
				return rows, nil
			}
		}
	}

	switch data := parsed.Data.(type) {
	case map[string]any:
		return []types.Row{types.NewRowFromMap(data)}, nil
	case []any:
		rows := make([]types.Row, 0, len(data))
		for _, item := range data {
			switch v := item.(type) {
			case map[string]any:
				rows = append(rows, types.NewRowFromMap(v))
			default:
				rows = append(rows, types.NewRow(types.MRP("content", v)))
			}
		}
		return rows, nil
	}

	if parsed.Text != "" {
		var fallback map[string]any
		if err := json.Unmarshal([]byte(parsed.Text), &fallback); err == nil {
			return []types.Row{types.NewRowFromMap(fallback)}, nil
		}
		return []types.Row{types.NewRow(types.MRP("response", parsed.Text))}, nil
	}

	return []types.Row{types.NewRow(types.MRP("content", nil))}, nil
}
