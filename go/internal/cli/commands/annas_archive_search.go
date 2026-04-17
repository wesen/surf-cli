package commands

import (
	"context"
	_ "embed"
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

//go:embed scripts/annas_archive_search.js
var annasArchiveSearchScript string

type AnnasArchiveSearchCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*AnnasArchiveSearchCommand)(nil)
var _ cmds.WriterCommand = (*AnnasArchiveSearchCommand)(nil)

type AnnasArchiveSearchSettings struct {
	Query       string `glazed:"query"`
	MaxResults  int    `glazed:"max-results"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type annasArchiveSearchData struct {
	Raw map[string]any
}

func NewAnnasArchiveSearchCommand() (*AnnasArchiveSearchCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Search Anna's Archive for papers"),
		cmds.WithLong("Searches Anna's Archive for papers by title, author, DOI, ISBN, or other queries. Navigates to the search results page and extracts paper metadata including titles, authors, formats, and download sources."),
		cmds.WithFlags(
			fields.New("query", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Search query (title, author, DOI, ISBN, etc.)")),
			fields.New("max-results", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Maximum number of results to return")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the search tab open instead of closing it when finished")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &AnnasArchiveSearchCommand{CommandDescription: desc}, nil
}

func buildAnnasArchiveSearchURL(query string) string {
	return "https://annas-archive.gl/search?q=" + url.QueryEscape(query) + "&index=journals"
}

func buildAnnasArchiveSearchCode(s *AnnasArchiveSearchSettings) string {
	return annasArchiveSearchScript
}

func (c *AnnasArchiveSearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &AnnasArchiveSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchAnnasArchiveSearch(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range annasArchiveSearchDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *AnnasArchiveSearchCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &AnnasArchiveSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchAnnasArchiveSearch(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderAnnasArchiveSearchMarkdown(data.Raw))
	return err
}

func fetchAnnasArchiveSearch(ctx context.Context, s *AnnasArchiveSearchSettings) (data *annasArchiveSearchData, retErr error) {
	if strings.TrimSpace(s.Query) == "" {
		return nil, fmt.Errorf("--query is required")
	}

	searchURL := buildAnnasArchiveSearchURL(s.Query)
	code := buildAnnasArchiveSearchCode(s)

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
			URLPrefix: "https://annas-archive.gl/search",
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
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: "https://annas-archive.gl/search"}); err != nil {
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

	return parseAnnasArchiveSearchResponse(resp)
}

func parseAnnasArchiveSearchResponse(resp map[string]any) (*annasArchiveSearchData, error) {
	if e := extractErrorText(resp); e != "" {
		return nil, fmt.Errorf("%s", e)
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		if parsed.Text != "" {
			return &annasArchiveSearchData{Raw: map[string]any{"content": parsed.Text}}, nil
		}
		return &annasArchiveSearchData{Raw: map[string]any{"content": nil}}, nil
	}

	return &annasArchiveSearchData{Raw: dataMap}, nil
}

func annasArchiveSearchDataToRows(data *annasArchiveSearchData) []types.Row {
	dataMap := data.Raw
	rows := []types.Row{}

	query, _ := dataMap["query"].(string)
	totalResults, _ := dataMap["totalResults"]

	items, _ := dataMap["results"].([]any)
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rowMap := map[string]any{
			"query":         query,
			"totalResults":   totalResults,
		}
		for k, v := range m {
			rowMap[k] = v
		}
		rows = append(rows, types.NewRowFromMap(rowMap))
	}

	if len(rows) == 0 {
		rows = append(rows, types.NewRowFromMap(dataMap))
	}

	return rows
}

func renderAnnasArchiveSearchMarkdown(data map[string]any) string {
	var b strings.Builder

	query, _ := data["query"].(string)
	totalResults := fmt.Sprintf("%v", data["totalResults"])

	b.WriteString("# Anna's Archive Search\n\n")
	if query != "" {
		b.WriteString(fmt.Sprintf("- Query: `%s`\n", query))
	}
	if totalResults != "<nil>" {
		b.WriteString(fmt.Sprintf("- Total results: %s\n", totalResults))
	}

	items, _ := data["results"].([]any)
	if len(items) == 0 {
		b.WriteString("\n_No results found._\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("\n## Results (%d found)\n\n", len(items)))
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		title, _ := m["title"].(string)
		href, _ := m["href"].(string)
		md5, _ := m["md5"].(string)
		format, _ := m["format"].(string)
		size, _ := m["size"].(string)
		authors, _ := m["authors"].(string)
		year, _ := m["year"].(string)

		b.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, title))
		if href != "" {
			b.WriteString(fmt.Sprintf("URL: https://annas-archive.gl%s\n", href))
		}
		if md5 != "" {
			b.WriteString(fmt.Sprintf("MD5: `%s`\n", md5))
		}
		if authors != "" {
			b.WriteString(fmt.Sprintf("Authors: %s\n", authors))
		}
		if year != "" {
			b.WriteString(fmt.Sprintf("Year: %s\n", year))
		}

		var metaParts []string
		if format != "" {
			metaParts = append(metaParts, format)
		}
		if size != "" {
			metaParts = append(metaParts, size)
		}
		if len(metaParts) > 0 {
			b.WriteString(fmt.Sprintf("Info: %s\n", strings.Join(metaParts, ", ")))
		}
		b.WriteString("\n")
	}

	return b.String()
}
