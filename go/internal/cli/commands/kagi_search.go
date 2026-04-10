package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

//go:embed scripts/kagi_search.js
var kagiSearchScript string

type KagiSearchCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*KagiSearchCommand)(nil)
var _ cmds.WriterCommand = (*KagiSearchCommand)(nil)

type KagiSearchSettings struct {
	Query       string `glazed:"query"`
	MaxResults  int    `glazed:"max-results"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type kagiSearchData struct {
	Raw map[string]any
}

func NewKagiSearchCommand() (*KagiSearchCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"kagi-search",
		cmds.WithShort("Run a Kagi search and extract results"),
		cmds.WithLong("Navigates to a Kagi search results page for the given query, waits for the results to stabilize, and extracts structured search results. By default it renders a Markdown report. Use --with-glaze-output for structured row output."),
		cmds.WithFlags(
			fields.New("query", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Search query to run on Kagi")),
			fields.New("max-results", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Maximum number of result rows to extract; 0 means all discovered results")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep a newly created search tab open instead of closing it when the command finishes")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &KagiSearchCommand{CommandDescription: desc}, nil
}

func buildKagiSearchURL(query string) string {
	return "https://kagi.com/search?q=" + url.QueryEscape(query)
}

func buildKagiSearchCode(s *KagiSearchSettings) (string, error) {
	options := map[string]any{
		"maxResults": s.MaxResults,
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal kagi search options: %w", err)
	}
	return fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), kagiSearchScript), nil
}

func (c *KagiSearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &KagiSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchKagiSearch(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range kagiSearchDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *KagiSearchCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &KagiSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchKagiSearch(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderKagiSearchMarkdown(data.Raw))
	return err
}

func fetchKagiSearch(ctx context.Context, s *KagiSearchSettings) (data *kagiSearchData, retErr error) {
	if strings.TrimSpace(s.Query) == "" {
		return nil, fmt.Errorf("--query is required")
	}

	searchURL := buildKagiSearchURL(s.Query)
	code, err := buildKagiSearchCode(s)
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
		resolvedTabID, err := openOwnedTab(ctx, client, searchURL, tabReadyOptions{
			URLExact: searchURL,
		})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": searchURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLExact: searchURL}); err != nil {
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

	return parseKagiSearchResponse(resp)
}

func parseKagiSearchResponse(resp map[string]any) (*kagiSearchData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &kagiSearchData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &kagiSearchData{Raw: map[string]any{"content": nil}}, nil
	}

	return &kagiSearchData{Raw: dataMap}, nil
}

func kagiSearchDataToRows(data *kagiSearchData) []types.Row {
	dataMap := data.Raw
	query, _ := dataMap["query"].(string)
	href, _ := dataMap["href"].(string)
	title, _ := dataMap["title"].(string)
	resultCount, _ := dataMap["resultCount"]
	waitedMs, _ := dataMap["waitedMs"]
	maxResults, _ := dataMap["maxResults"]

	quickAnswerText := ""
	quickAnswerTitle := ""
	if qa, ok := dataMap["quickAnswer"].(map[string]any); ok {
		quickAnswerText, _ = qa["text"].(string)
		quickAnswerTitle, _ = qa["title"].(string)
	}

	items, _ := dataMap["results"].([]any)
	rows := make([]types.Row, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rowMap := map[string]any{
			"query":            query,
			"href":             href,
			"title":            title,
			"resultCount":      resultCount,
			"waitedMs":         waitedMs,
			"maxResults":       maxResults,
			"quickAnswerTitle": quickAnswerTitle,
			"quickAnswerText":  quickAnswerText,
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

func renderKagiSearchMarkdown(data map[string]any) string {
	var b strings.Builder

	query, _ := data["query"].(string)
	href, _ := data["href"].(string)
	title, _ := data["title"].(string)
	resultCount := fmt.Sprintf("%v", data["resultCount"])

	b.WriteString("# Kagi Search\n\n")
	if query != "" {
		b.WriteString(fmt.Sprintf("- Query: `%s`\n", query))
	}
	if href != "" {
		b.WriteString(fmt.Sprintf("- URL: %s\n", href))
	}
	if title != "" {
		b.WriteString(fmt.Sprintf("- Title: %s\n", title))
	}
	if resultCount != "<nil>" {
		b.WriteString(fmt.Sprintf("- Result count: %s\n", resultCount))
	}

	if qa, ok := data["quickAnswer"].(map[string]any); ok {
		if text, _ := qa["text"].(string); text != "" {
			b.WriteString("\n## Quick Answer\n\n")
			b.WriteString(text)
			b.WriteString("\n")
		}
	}

	items, _ := data["results"].([]any)
	if len(items) == 0 {
		b.WriteString("\n_No results found._\n")
		return b.String()
	}

	b.WriteString("\n## Results\n")
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		index := fmt.Sprintf("%v", m["index"])
		resultTitle, _ := m["title"].(string)
		resultURL, _ := m["url"].(string)
		displayURL, _ := m["displayUrl"].(string)
		snippet, _ := m["snippet"].(string)

		b.WriteString("\n")
		if resultURL != "" {
			b.WriteString(fmt.Sprintf("### %s. [%s](%s)\n\n", index, resultTitle, resultURL))
		} else {
			b.WriteString(fmt.Sprintf("### %s. %s\n\n", index, resultTitle))
		}
		if displayURL != "" {
			b.WriteString(fmt.Sprintf("- Display URL: `%s`\n", displayURL))
		}
		if snippet != "" {
			b.WriteString("\n")
			b.WriteString(snippet)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func extractTabIDFromResponse(resp map[string]any) (int64, error) {
	if e := extractErrorText(resp); e != "" {
		return 0, fmt.Errorf("%s", e)
	}
	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("missing structured tab creation response")
	}
	rawID, ok := dataMap["tabId"]
	if !ok {
		return 0, fmt.Errorf("missing tabId in tab creation response")
	}
	switch v := rawID.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unexpected tabId type %T", rawID)
	}
}

func closeOwnedTab(ctx context.Context, client *transport.Client, tabID *int64) error {
	if tabID == nil {
		return nil
	}
	_, err := ExecuteTool(ctx, client, "tab.close", map[string]any{"id": *tabID}, nil, nil)
	return err
}
