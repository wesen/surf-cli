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

//go:embed scripts/claude_transcript.js
var claudeTranscriptScript string

type ClaudeTranscriptCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ClaudeTranscriptCommand)(nil)
var _ cmds.WriterCommand = (*ClaudeTranscriptCommand)(nil)

type ClaudeTranscriptSettings struct {
	ExportFile   string `glazed:"export-file"`
	ExportFormat string `glazed:"export-format"`
	Socket       string `glazed:"socket-path"`
	TimeoutMS    int    `glazed:"timeout-ms"`
	TabID        int64  `glazed:"tab-id"`
	WindowID     int64  `glazed:"window-id"`
	DebugSocket  bool   `glazed:"debug-socket"`
}

type claudeTranscriptData struct {
	Raw map[string]any
}

func NewClaudeTranscriptCommand() (*ClaudeTranscriptCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"claude-transcript",
		cmds.WithShort("Export the current Claude conversation"),
		cmds.WithLong("Extracts the current claude.ai conversation from the active page DOM. By default it prints a human-readable Markdown transcript. Use --with-glaze-output for structured row output."),
		cmds.WithFlags(
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

	return &ClaudeTranscriptCommand{CommandDescription: desc}, nil
}

func buildClaudeTranscriptCode() string {
	return claudeTranscriptScript
}

func (c *ClaudeTranscriptCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ClaudeTranscriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchClaudeTranscript(ctx, s)
	if err != nil {
		return err
	}
	if s.ExportFile != "" {
		if err := writeClaudeTranscriptExport(s.ExportFile, s.ExportFormat, data); err != nil {
			return err
		}
	}
	for _, row := range claudeTranscriptDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *ClaudeTranscriptCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := &ClaudeTranscriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchClaudeTranscript(ctx, s)
	if err != nil {
		return err
	}
	if s.ExportFile != "" {
		if err := writeClaudeTranscriptExport(s.ExportFile, s.ExportFormat, data); err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, renderClaudeTranscriptMarkdown(data.Raw))
	return err
}

func fetchClaudeTranscript(ctx context.Context, s *ClaudeTranscriptSettings) (*claudeTranscriptData, error) {
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

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildClaudeTranscriptCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	return parseClaudeTranscriptResponse(resp)
}

func parseClaudeTranscriptResponse(resp map[string]any) (*claudeTranscriptData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &claudeTranscriptData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &claudeTranscriptData{Raw: map[string]any{"content": nil}}, nil
	}
	return &claudeTranscriptData{Raw: dataMap}, nil
}

func claudeTranscriptDataToRows(data *claudeTranscriptData) []types.Row {
	dataMap := data.Raw
	items, _ := dataMap["transcript"].([]any)
	rows := make([]types.Row, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rowMap := map[string]any{
			"href":              dataMap["href"],
			"title":             dataMap["title"],
			"conversationTitle": dataMap["conversationTitle"],
			"currentModel":      dataMap["currentModel"],
			"turnCount":         dataMap["turnCount"],
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

func writeClaudeTranscriptExport(path string, format string, data *claudeTranscriptData) error {
	var body []byte
	var err error
	switch format {
	case "json":
		body, err = json.MarshalIndent(data.Raw, "", "  ")
	case "markdown":
		body = []byte(renderClaudeTranscriptMarkdown(data.Raw))
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
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func renderClaudeTranscriptMarkdown(data map[string]any) string {
	var b strings.Builder
	title, _ := data["conversationTitle"].(string)
	if title == "" {
		title, _ = data["title"].(string)
	}
	if title == "" {
		title = "Claude Conversation"
	}
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	if href, _ := data["href"].(string); href != "" {
		b.WriteString("- URL: ")
		b.WriteString(href)
		b.WriteString("\n")
	}
	if model, _ := data["currentModel"].(string); model != "" {
		b.WriteString("- Model: ")
		b.WriteString(model)
		b.WriteString("\n")
	}
	if thinkingMode, _ := data["currentThinkingMode"].(string); thinkingMode != "" {
		b.WriteString("- Thinking mode: ")
		b.WriteString(thinkingMode)
		b.WriteString("\n")
	}
	items, _ := data["transcript"].([]any)
	if len(items) == 0 {
		b.WriteString("\n_No transcript turns found._\n")
		return b.String()
	}
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		role, _ := m["role"].(string)
		text, _ := m["text"].(string)
		index := fmt.Sprintf("%v", m["index"])
		b.WriteString("\n## ")
		if role != "" {
			b.WriteString(strings.ToUpper(role[:1]))
			if len(role) > 1 {
				b.WriteString(role[1:])
			}
		} else {
			b.WriteString("Turn")
		}
		b.WriteString(" ")
		b.WriteString(index)
		b.WriteString("\n\n")
		b.WriteString(text)
		b.WriteString("\n")
		if citations, _ := m["citations"].([]any); len(citations) > 0 {
			b.WriteString("\n### Citations\n\n")
			for _, citation := range citations {
				cm, ok := citation.(map[string]any)
				if !ok {
					continue
				}
				label, _ := cm["text"].(string)
				href, _ := cm["href"].(string)
				if label == "" {
					label = href
				}
				b.WriteString("- ")
				if href != "" {
					b.WriteString(label)
					b.WriteString(": ")
					b.WriteString(href)
				} else {
					b.WriteString(label)
				}
				b.WriteString("\n")
			}
		}
		if searchWeb, _ := m["searchWeb"].(map[string]any); searchWeb != nil {
			b.WriteString("\n### Searched The Web\n\n")
			if label, _ := searchWeb["label"].(string); label != "" {
				b.WriteString("- Label: ")
				b.WriteString(label)
				b.WriteString("\n")
			}
			if queries, _ := searchWeb["queries"].([]any); len(queries) > 0 {
				b.WriteString("- Queries:\n")
				for _, query := range queries {
					b.WriteString("  - ")
					b.WriteString(fmt.Sprintf("%v", query))
					b.WriteString("\n")
				}
			}
			if results, _ := searchWeb["results"].([]any); len(results) > 0 {
				b.WriteString("- Results:\n")
				for _, result := range results {
					rm, ok := result.(map[string]any)
					if !ok {
						continue
					}
					label, _ := rm["text"].(string)
					href, _ := rm["href"].(string)
					host, _ := rm["host"].(string)
					b.WriteString("  - ")
					if label != "" {
						b.WriteString(label)
					} else {
						b.WriteString(href)
					}
					if host != "" {
						b.WriteString(" [")
						b.WriteString(host)
						b.WriteString("]")
					}
					if href != "" {
						b.WriteString(": ")
						b.WriteString(href)
					}
					b.WriteString("\n")
				}
			}
		}
	}
	return b.String()
}
