package commands

import (
	"context"
	_ "embed"
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

//go:embed scripts/chatgpt_transcript.js
var chatGPTTranscriptScript string

type ChatGPTTranscriptCommand struct {
	*cmds.CommandDescription
}

type ChatGPTTranscriptSettings struct {
	WithActivity  bool   `glazed:"with-activity"`
	ActivityLimit int    `glazed:"activity-limit"`
	Socket        string `glazed:"socket-path"`
	TimeoutMS     int    `glazed:"timeout-ms"`
	TabID         int64  `glazed:"tab-id"`
	WindowID      int64  `glazed:"window-id"`
	DebugSocket   bool   `glazed:"debug-socket"`
}

func NewChatGPTTranscriptCommand() (*ChatGPTTranscriptCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"chatgpt-transcript",
		cmds.WithShort("Export the current ChatGPT conversation as structured rows"),
		cmds.WithLong("Extracts the current chatgpt.com conversation from the active page DOM. Optionally opens 'Thought for ...' Activity flyouts and attaches the scraped content to assistant turns."),
		cmds.WithFlags(
			fields.New("with-activity", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Open and scrape ChatGPT Activity flyouts for assistant turns that have thought traces")),
			fields.New("activity-limit", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Maximum number of Activity flyouts to open; 0 means all assistant turns with thought traces")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &ChatGPTTranscriptCommand{CommandDescription: desc}, nil
}

func buildChatGPTTranscriptCode(s *ChatGPTTranscriptSettings) (string, error) {
	options := map[string]any{
		"withActivity":  s.WithActivity,
		"activityLimit": s.ActivityLimit,
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal transcript options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), chatGPTTranscriptScript), nil
}

func (c *ChatGPTTranscriptCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &ChatGPTTranscriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	code, err := buildChatGPTTranscriptCode(s)
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

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": code}, tabID, windowID)
	if err != nil {
		return err
	}

	rows, err := chatGPTTranscriptResponseToRows(resp)
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

func chatGPTTranscriptResponseToRows(resp map[string]any) ([]types.Row, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return []types.Row{types.NewRow(types.MRP("content", parsed.Text))}, nil
		}
		return []types.Row{types.NewRow(types.MRP("content", nil))}, nil
	}

	href, _ := dataMap["href"].(string)
	title, _ := dataMap["title"].(string)
	turnCount, _ := dataMap["turnCount"]
	withActivity, _ := dataMap["withActivity"]
	activityLimit, _ := dataMap["activityLimit"]
	activityExported, _ := dataMap["activityExported"]

	items, _ := dataMap["transcript"].([]any)
	rows := make([]types.Row, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rowMap := map[string]any{
			"href":             href,
			"title":            title,
			"turnCount":        turnCount,
			"withActivity":     withActivity,
			"activityLimit":    activityLimit,
			"activityExported": activityExported,
		}
		for k, v := range m {
			rowMap[k] = v
		}
		rows = append(rows, types.NewRowFromMap(rowMap))
	}
	if len(rows) > 0 {
		return rows, nil
	}

	return []types.Row{types.NewRowFromMap(dataMap)}, nil
}
