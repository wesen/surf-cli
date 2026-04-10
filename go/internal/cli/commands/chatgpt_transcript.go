package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

var _ cmds.GlazeCommand = (*ChatGPTTranscriptCommand)(nil)
var _ cmds.WriterCommand = (*ChatGPTTranscriptCommand)(nil)

type ChatGPTTranscriptSettings struct {
	WithActivity  bool   `glazed:"with-activity"`
	ActivityLimit int    `glazed:"activity-limit"`
	ExportFile    string `glazed:"export-file"`
	ExportFormat  string `glazed:"export-format"`
	Socket        string `glazed:"socket-path"`
	TimeoutMS     int    `glazed:"timeout-ms"`
	TabID         int64  `glazed:"tab-id"`
	WindowID      int64  `glazed:"window-id"`
	DebugSocket   bool   `glazed:"debug-socket"`
}

type chatGPTTranscriptData struct {
	Raw map[string]any
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
		cmds.WithShort("Export the current ChatGPT conversation"),
		cmds.WithLong("Extracts the current chatgpt.com conversation from the active page DOM. By default it prints a human-readable Markdown transcript. Use --with-glaze-output for structured row output. Optionally opens 'Thought for ...' Activity flyouts and attaches the scraped content to assistant turns."),
		cmds.WithFlags(
			fields.New("with-activity", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Open and scrape ChatGPT Activity flyouts for assistant turns that have thought traces")),
			fields.New("activity-limit", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Maximum number of Activity flyouts to open; 0 means all assistant turns with thought traces")),
			fields.New("export-file", fields.TypeString, fields.WithHelp("Optional file path to write the exported transcript artifact")),
			fields.New("export-format", fields.TypeChoice, fields.WithChoices("markdown", "json"), fields.WithDefault("markdown"), fields.WithHelp("Export artifact format used with --export-file")),
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

	data, err := fetchChatGPTTranscript(ctx, s)
	if err != nil {
		return err
	}
	if s.ExportFile != "" {
		if err := writeChatGPTTranscriptExport(s.ExportFile, s.ExportFormat, data); err != nil {
			return err
		}
	}

	for _, row := range chatGPTTranscriptDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *ChatGPTTranscriptCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &ChatGPTTranscriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchChatGPTTranscript(ctx, s)
	if err != nil {
		return err
	}
	if s.ExportFile != "" {
		if err := writeChatGPTTranscriptExport(s.ExportFile, s.ExportFormat, data); err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, renderChatGPTTranscriptMarkdown(data.Raw))
	return err
}

func fetchChatGPTTranscript(ctx context.Context, s *ChatGPTTranscriptSettings) (*chatGPTTranscriptData, error) {
	code, err := buildChatGPTTranscriptCode(s)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return parseChatGPTTranscriptResponse(resp)
}

func parseChatGPTTranscriptResponse(resp map[string]any) (*chatGPTTranscriptData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &chatGPTTranscriptData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &chatGPTTranscriptData{Raw: map[string]any{"content": nil}}, nil
	}

	return &chatGPTTranscriptData{Raw: dataMap}, nil
}

func chatGPTTranscriptResponseToRows(resp map[string]any) ([]types.Row, error) {
	data, err := parseChatGPTTranscriptResponse(resp)
	if err != nil {
		return nil, err
	}
	return chatGPTTranscriptDataToRows(data), nil
}

func chatGPTTranscriptDataToRows(data *chatGPTTranscriptData) []types.Row {
	dataMap := data.Raw
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
		return rows
	}

	return []types.Row{types.NewRowFromMap(dataMap)}
}

func writeChatGPTTranscriptExport(path string, format string, data *chatGPTTranscriptData) error {
	var body []byte
	var err error

	switch format {
	case "json":
		body, err = json.MarshalIndent(data.Raw, "", "  ")
	case "markdown":
		body = []byte(renderChatGPTTranscriptMarkdown(data.Raw))
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	return nil
}

func renderChatGPTTranscriptMarkdown(dataMap map[string]any) string {
	var b strings.Builder

	title, _ := dataMap["title"].(string)
	href, _ := dataMap["href"].(string)
	withActivity, _ := dataMap["withActivity"].(bool)
	activityExported, _ := dataMap["activityExported"]

	if strings.TrimSpace(title) == "" {
		b.WriteString("# ChatGPT Transcript\n\n")
	} else {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if href != "" {
		b.WriteString("- URL: ")
		b.WriteString(href)
		b.WriteString("\n")
	}
	if withActivity {
		b.WriteString("- With Activity: yes\n")
	} else {
		b.WriteString("- With Activity: no\n")
	}
	if activityExported != nil {
		b.WriteString("- Activity Exported: ")
		b.WriteString(fmt.Sprintf("%v", activityExported))
		b.WriteString("\n")
	}

	items, _ := dataMap["transcript"].([]any)
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		index := m["index"]
		role, _ := m["role"].(string)
		messageID, _ := m["messageId"].(string)
		model, _ := m["model"].(string)
		text, _ := m["text"].(string)
		thoughtButtonText, _ := m["thoughtButtonText"].(string)
		activityText, _ := m["activityText"].(string)

		b.WriteString("\n## Turn ")
		b.WriteString(fmt.Sprintf("%v", index))
		if role != "" {
			b.WriteString(" - ")
			b.WriteString(role)
		}
		b.WriteString("\n\n")
		if messageID != "" {
			b.WriteString("- Message ID: `")
			b.WriteString(messageID)
			b.WriteString("`\n")
		}
		if model != "" {
			b.WriteString("- Model: `")
			b.WriteString(model)
			b.WriteString("`\n")
		}
		if thoughtButtonText != "" {
			b.WriteString("- Thought Trace: ")
			b.WriteString(thoughtButtonText)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(text)
		b.WriteString("\n")
		if strings.TrimSpace(activityText) != "" {
			b.WriteString("\n### Activity\n\n```text\n")
			b.WriteString(activityText)
			b.WriteString("\n```\n")
		}
	}

	return b.String()
}
