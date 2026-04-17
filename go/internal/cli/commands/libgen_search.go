package commands

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
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

type LibgenSearchCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*LibgenSearchCommand)(nil)
var _ cmds.WriterCommand = (*LibgenSearchCommand)(nil)

type LibgenSearchSettings struct {
	Query      string `glazed:"query"`
	Limit      int    `glazed:"limit"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket     string `glazed:"socket-path"`
	TimeoutMS  int    `glazed:"timeout-ms"`
	TabID      int64  `glazed:"tab-id"`
	WindowID   int64  `glazed:"window-id"`
	DebugSocket bool  `glazed:"debug-socket"`
}

func NewLibgenSearchCommand() (*LibgenSearchCommand, error) {
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
		cmds.WithShort("Search for books on 1lib.sk"),
		cmds.WithLong("Search for books by title, author, ISBN, DOI, or any keyword."),
		cmds.WithFlags(
			fields.New("query", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Search query (title, author, ISBN, DOI, etc.)")),
			fields.New("limit", fields.TypeInteger, fields.WithDefault(20), fields.WithHelp("Maximum number of results")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the tab open when finished")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Tab ID override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Window ID override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Debug socket traffic")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &LibgenSearchCommand{CommandDescription: desc}, nil
}

func buildLibgenSearchCode() string {
	return `
var results = [];
var maxResults = 50;
var seen = {};

// Debug: count all links
var allLinks = document.querySelectorAll('a');
var bookLinks = [];
allLinks.forEach(function(link) {
  var href = link.getAttribute('href');
  if (href && href.match(/\/book\//)) {
    bookLinks.push(link);
  }
});

bookLinks.forEach(function(link) {
  if (results.length >= maxResults) return;
  
  var href = link.getAttribute('href');
  var idMatch = href.match(/\/book\/([^\/]+)/);
  var id = idMatch ? idMatch[1] : '';
  
  if (seen[id]) return;
  seen[id] = true;
  
  // Get title - try the link text first
  var title = link.textContent.trim();
  
  // If title is too short or looks like navigation, try parent
  if (title.length < 5) {
    var parent = link.closest('.item') || link.closest('div');
    if (parent) {
      var h3 = parent.querySelector('h3');
      if (h3) title = h3.textContent.trim();
    }
  }
  
  // Skip UI elements
  if (title.length < 5 || title.length > 400) return;
  if (title.match(/^(Log In|Home|Library|Donate|Search|Z-Library|\+[\d]+|Results?|Most Popular|Booklists?|\d+ Books)/)) return;
  
  // Get author from nearby element
  var author = '';
  var container = link.closest('.item') || link.closest('div');
  if (container) {
    var authorLink = container.querySelector('a[href*="/author/"]');
    if (authorLink) author = authorLink.textContent.trim();
  }
  
  results.push({
    title: title,
    author: author,
    url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
    id: id
  });
});

return {
  count: results.length,
  debug: 'Found ' + bookLinks.length + ' book links',
  results: results
};
`
}

func (c *LibgenSearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &LibgenSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenSearch(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range libgenSearchToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *LibgenSearchCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &LibgenSearchSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenSearch(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderLibgenSearchMarkdown(data))
	return err
}

func fetchLibgenSearch(ctx context.Context, s *LibgenSearchSettings) (*libgenSearchData, error) {
	if strings.TrimSpace(s.Query) == "" {
		return nil, fmt.Errorf("--query is required")
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

	searchURL := "https://1lib.sk/s/" + strings.ReplaceAll(s.Query, " ", "+")

	if tabID == nil && windowID == nil {
		resolvedTabID, err := openOwnedTab(ctx, client, searchURL, tabReadyOptions{
			URLPrefix: "https://1lib.sk/s/",
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
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: "https://1lib.sk/s/"}); err != nil {
				return nil, err
			}
		}
	}

	defer func() {
		if s.KeepTabOpen {
			return
		}
		if err := closeOwnedTab(ctx, client, ownedTabID); err != nil {
			// ignore
		}
	}()

	// Wait for page to load results (site uses lazy loading)
	time.Sleep(5 * time.Second)

	// Wait for results to appear
	for i := 0; i < 15; i++ {
		resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": `return { href: location.href, title: document.title, readyState: document.readyState, bookLinks: document.querySelectorAll('a[href*="/book/"]').length };`}, tabID, windowID)
		if err == nil {
			parsed := parseResult(resp)
			if data, ok := parsed.Data.(map[string]any); ok {
				if bookLinks, ok := data["bookLinks"].(float64); ok && bookLinks > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Found %d book links\n", int(bookLinks))
					break
				}
			}
		}
		time.Sleep(2 * time.Second)
	}

	finalResp, finalErr := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenSearchCode()}, tabID, windowID)
	if finalErr != nil {
		return nil, finalErr
	}

	return parseLibgenSearchResponse(finalResp), nil
}

type libgenSearchData struct {
	Count   int
	Results []libgenBookResult
}

type libgenBookResult struct {
	Title  string
	Author string
	URL    string
	ID     string
	Year   string
	Size   string
	Format string
}

func parseLibgenSearchResponse(resp map[string]any) *libgenSearchData {
	result := &libgenSearchData{}
	if resp == nil {
		return result
	}

	parsed := parseResult(resp)
	if data, ok := parsed.Data.(map[string]any); ok {
		if count, ok := data["count"].(float64); ok {
			result.Count = int(count)
		}
		if results, ok := data["results"].([]any); ok {
			for _, r := range results {
				if m, ok := r.(map[string]any); ok {
					result.Results = append(result.Results, libgenBookResult{
						Title:  getString(m, "title"),
						Author: getString(m, "author"),
						URL:    getString(m, "url"),
						ID:     getString(m, "id"),
						Year:   getString(m, "year"),
						Size:   getString(m, "size"),
						Format: getString(m, "format"),
					})
				}
			}
		}
	}

	return result
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func libgenSearchToRows(data *libgenSearchData) []types.Row {
	rows := []types.Row{}
	for _, r := range data.Results {
		row := map[string]any{
			"title":  r.Title,
			"author": r.Author,
			"url":    r.URL,
			"id":     r.ID,
			"year":   r.Year,
			"size":   r.Size,
			"format": r.Format,
		}
		rows = append(rows, types.NewRowFromMap(row))
	}
	return rows
}

func renderLibgenSearchMarkdown(data *libgenSearchData) string {
	var b strings.Builder
	b.WriteString("# 1lib.sk Search Results\n\n")
	b.WriteString(fmt.Sprintf("- **Results:** %d found\n\n", data.Count))

	for i, r := range data.Results {
		if r.Title == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, r.Title))
		if r.Author != "" {
			b.WriteString(fmt.Sprintf("**Author:** %s\n\n", r.Author))
		}
		b.WriteString(fmt.Sprintf("- [View on 1lib.sk](%s)\n", r.URL))
		if r.Year != "" {
			b.WriteString(fmt.Sprintf("- **Year:** %s\n", r.Year))
		}
		if r.Format != "" {
			b.WriteString(fmt.Sprintf("- **Format:** %s\n", r.Format))
		}
		if r.Size != "" {
			b.WriteString(fmt.Sprintf("- **Size:** %s\n", r.Size))
		}
		b.WriteString(fmt.Sprintf("- **ID:** `%s`\n\n", r.ID))
	}

	return b.String()
}
