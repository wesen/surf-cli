package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
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

//go:embed scripts/gmail_list.js
var gmailListScript string

const gmailInboxURL = "https://mail.google.com/mail/u/0/#inbox"
const gmailURLPrefix = "https://mail.google.com/mail/"

type GmailListCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*GmailListCommand)(nil)
var _ cmds.WriterCommand = (*GmailListCommand)(nil)

type GmailListSettings struct {
	Inbox       bool   `glazed:"inbox"`
	MaxResults  int    `glazed:"max-results"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type gmailListData struct {
	Raw map[string]any
}

func NewGmailListCommand() (*GmailListCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List Gmail inbox threads"),
		cmds.WithLong("Opens or reuses a Gmail tab, ensures the inbox view is active, and extracts thread summary rows. By default it renders a Markdown report. Use --with-glaze-output for structured rows."),
		cmds.WithFlags(
			fields.New("inbox", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Explicitly list the inbox view; required in v1")),
			fields.New("max-results", fields.TypeInteger, fields.WithDefault(25), fields.WithHelp("Maximum number of inbox rows to extract; 0 means all visible rows")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep a newly created Gmail tab open instead of closing it when the command finishes")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &GmailListCommand{CommandDescription: desc}, nil
}

func buildGmailListCode(s *GmailListSettings) (string, error) {
	options := map[string]any{
		"maxResults": s.MaxResults,
		"mailbox":    "inbox",
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal gmail list options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), gmailListScript), nil
}

func (c *GmailListCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &GmailListSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchGmailList(ctx, s)
	if err != nil {
		return err
	}
	for _, row := range gmailThreadsToRows(data.Raw, "mailbox") {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *GmailListCommand) RunIntoWriter(ctx context.Context, vals *values.Values, w io.Writer) error {
	s := &GmailListSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	data, err := fetchGmailList(ctx, s)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, renderGmailThreadsMarkdown(data.Raw, "Gmail Inbox"))
	return err
}

func fetchGmailList(ctx context.Context, s *GmailListSettings) (data *gmailListData, retErr error) {
	if !s.Inbox {
		return nil, fmt.Errorf("--inbox is required in v1")
	}
	code, err := buildGmailListCode(s)
	if err != nil {
		return nil, err
	}

	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)
	client.Debug = s.DebugSocket

	var tabID *int64
	var ownedTabID *int64
	if s.TabID >= 0 {
		tabID = &s.TabID
	}
	var windowID *int64
	if s.WindowID >= 0 {
		windowID = &s.WindowID
	}

	if tabID == nil && windowID == nil {
		resolvedTabID, err := openOwnedTab(ctx, client, gmailInboxURL, tabReadyOptions{URLPrefix: gmailURLPrefix})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": gmailInboxURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: gmailURLPrefix}); err != nil {
				return nil, err
			}
		}
	}

	defer func() {
		if retErr != nil || s.KeepTabOpen {
			return
		}
		if err := closeOwnedTab(ctx, client, ownedTabID); err != nil {
			retErr = err
		}
	}()

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": code}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	parsed, err := parseGmailDataResponse(resp)
	if err != nil {
		return nil, err
	}
	return &gmailListData{Raw: parsed}, nil
}

func parseGmailDataResponse(resp map[string]any) (map[string]any, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return map[string]any{"content": parsed.Text}, nil
		}
		return map[string]any{"content": nil}, nil
	}
	return dataMap, nil
}

func gmailThreadsToRows(dataMap map[string]any, mode string) []types.Row {
	href, _ := dataMap["href"].(string)
	title, _ := dataMap["title"].(string)
	resultCount, _ := dataMap["resultCount"]
	waitedMs, _ := dataMap["waitedMs"]
	query, _ := dataMap["query"].(string)
	items, _ := dataMap["threads"].([]any)
	rows := make([]types.Row, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rowMap := map[string]any{
			"href":        href,
			"title":       title,
			"resultCount": resultCount,
			"waitedMs":    waitedMs,
			"mode":        mode,
			"query":       query,
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

func renderGmailThreadsMarkdown(dataMap map[string]any, heading string) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(heading)
	b.WriteString("\n\n")
	if href, _ := dataMap["href"].(string); href != "" {
		b.WriteString("- URL: ")
		b.WriteString(href)
		b.WriteString("\n")
	}
	if title, _ := dataMap["title"].(string); title != "" {
		b.WriteString("- Title: ")
		b.WriteString(title)
		b.WriteString("\n")
	}
	if query, _ := dataMap["query"].(string); query != "" {
		b.WriteString("- Query: `")
		b.WriteString(query)
		b.WriteString("`\n")
	}
	if resultCount, ok := dataMap["resultCount"]; ok {
		b.WriteString("- Result count: ")
		b.WriteString(fmt.Sprintf("%v", resultCount))
		b.WriteString("\n")
	}
	b.WriteString("\n## Threads\n\n")

	items, _ := dataMap["threads"].([]any)
	if len(items) == 0 {
		b.WriteString("_No threads found._\n")
		return b.String()
	}

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		index, _ := m["index"]
		subject, _ := m["subject"].(string)
		participant, _ := m["participant"].(string)
		snippet, _ := m["snippet"].(string)
		timestamp, _ := m["timestamp"].(string)
		b.WriteString("### ")
		b.WriteString(fmt.Sprintf("%v", index))
		b.WriteString(". ")
		if subject != "" {
			b.WriteString(subject)
		} else {
			b.WriteString("(no subject)")
		}
		b.WriteString("\n\n")
		if participant != "" {
			b.WriteString("- From: ")
			b.WriteString(participant)
			b.WriteString("\n")
		}
		if timestamp != "" {
			b.WriteString("- Time: ")
			b.WriteString(timestamp)
			b.WriteString("\n")
		}
		if unread, ok := m["unread"].(bool); ok {
			b.WriteString("- Unread: ")
			b.WriteString(fmt.Sprintf("%t", unread))
			b.WriteString("\n")
		}
		if hasAttachment, ok := m["hasAttachment"].(bool); ok {
			b.WriteString("- Attachment: ")
			b.WriteString(fmt.Sprintf("%t", hasAttachment))
			b.WriteString("\n")
		}
		if threadID, _ := m["threadId"].(string); threadID != "" {
			b.WriteString("- Thread ID: `")
			b.WriteString(threadID)
			b.WriteString("`\n")
		}
		if legacyThreadID, _ := m["legacyThreadId"].(string); legacyThreadID != "" {
			b.WriteString("- Legacy thread ID: `")
			b.WriteString(legacyThreadID)
			b.WriteString("`\n")
		}
		if snippet != "" {
			b.WriteString("\n")
			b.WriteString(snippet)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
