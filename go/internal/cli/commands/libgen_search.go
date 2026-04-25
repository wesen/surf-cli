package commands

import (
	"context"
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
	Query       string `glazed:"query"`
	Limit       int    `glazed:"limit"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
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

func buildLibgenSearchCode(limit int) string {
	return fmt.Sprintf(`var results = [];
var maxResults = %d;

var cards = document.querySelectorAll('z-bookcard.ready');

cards.forEach(function(card) {
  if (results.length >= maxResults) return;

  var href = card.getAttribute('href') || '';
  var numericId = card.getAttribute('id') || '';

  // Extract the hash ID from the URL (e.g., oOY31bWjPK from /book/oOY31bWjPK/...)
  var urlIdMatch = href.match(/\/book\/([^\/]+)/);
  var urlId = urlIdMatch ? urlIdMatch[1] : '';

  // Title and author from light DOM slots (most reliable)
  var titleSlot = card.querySelector('[slot="title"]');
  var authorSlot = card.querySelector('[slot="author"]');
  var title = titleSlot ? titleSlot.textContent.trim() : '';
  var author = authorSlot ? authorSlot.textContent.trim() : '';

  // Fallback to z-cover attributes inside shadow DOM
  if (!title) {
    var shadow = card.shadowRoot;
    if (shadow) {
      var cover = shadow.querySelector('z-cover');
      if (cover) {
        title = cover.getAttribute('title') || '';
        author = cover.getAttribute('author') || '';
      }
    }
  }

  var url = href.startsWith('http') ? href : 'https://1lib.sk' + href;
  var downloadPath = card.getAttribute('download') || '';
  var downloadUrl = downloadPath ? ('https://1lib.sk' + downloadPath) : '';

  results.push({
    title: title,
    author: author,
    url: url,
    id: urlId,
    numericId: numericId,
    downloadUrl: downloadUrl,
    year: card.getAttribute('year') || '',
    format: (card.getAttribute('extension') || '').toUpperCase(),
    size: card.getAttribute('filesize') || '',
    publisher: card.getAttribute('publisher') || '',
    language: card.getAttribute('language') || ''
  });
});

return {
  count: results.length,
  results: results
};
`, limit)
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
	client := transport.NewClient(s.Socket, time.Duration(s.TimeoutMS)*time.Millisecond)
	client.Debug = s.DebugSocket

	var tabID *int64
	var windowID *int64
	var ownedTabID *int64

	if s.TabID != -1 {
		id := s.TabID
		tabID = &id
	}
	if s.WindowID != -1 {
		id := s.WindowID
		windowID = &id
	}

	searchURL := fmt.Sprintf("https://1lib.sk/s/%s", s.Query)

	if tabID == nil {
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

	// Wait for z-bookcard custom elements to hydrate (they render via shadow DOM)
	// Poll briefly for z-bookcard.ready to appear
	cardReadyDeadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(cardReadyDeadline) {
		probeResp, probeErr := ExecuteTool(ctx, client, "js", map[string]any{"code": `return document.querySelectorAll('z-bookcard.ready').length;`}, tabID, windowID)
		if probeErr == nil {
			parsed := parseResult(probeResp)
			if count, ok := parsed.Data.(float64); ok && count > 0 {
				if s.DebugSocket {
					fmt.Fprintf(os.Stderr, "DEBUG: Found %d z-bookcard.ready elements\n", int(count))
				}
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	finalResp, finalErr := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenSearchCode(s.Limit)}, tabID, windowID)
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
	Title       string
	Author      string
	URL         string
	ID          string
	NumericID   string
	DownloadURL string
	Year        string
	Size        string
	Format      string
	Publisher   string
	Language    string
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
				if rm, ok := r.(map[string]any); ok {
					book := libgenBookResult{}
					if title, ok := rm["title"].(string); ok {
						book.Title = title
					}
					if author, ok := rm["author"].(string); ok {
						book.Author = author
					}
					if url, ok := rm["url"].(string); ok {
						book.URL = url
					}
					if id, ok := rm["id"].(string); ok {
						book.ID = id
					}
					if numericId, ok := rm["numericId"].(string); ok {
						book.NumericID = numericId
					}
					if downloadUrl, ok := rm["downloadUrl"].(string); ok {
						book.DownloadURL = downloadUrl
					}
					if year, ok := rm["year"].(string); ok {
						book.Year = year
					}
					if size, ok := rm["size"].(string); ok {
						book.Size = size
					}
					if format, ok := rm["format"].(string); ok {
						book.Format = format
					}
					if publisher, ok := rm["publisher"].(string); ok {
						book.Publisher = publisher
					}
					if language, ok := rm["language"].(string); ok {
						book.Language = language
					}
					result.Results = append(result.Results, book)
				}
			}
		}
	}
	return result
}

func libgenSearchToRows(data *libgenSearchData) []types.Row {
	rows := []types.Row{}
	for _, r := range data.Results {
		row := map[string]any{
			"title":        r.Title,
			"author":       r.Author,
			"url":          r.URL,
			"id":           r.ID,
			"numeric-id":   r.NumericID,
			"download-url": r.DownloadURL,
			"year":         r.Year,
			"size":         r.Size,
			"format":       r.Format,
			"publisher":    r.Publisher,
			"language":     r.Language,
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
		if r.ID != "" {
			b.WriteString(fmt.Sprintf("- **ID:** `%s`\n", r.ID))
		}
		if r.Year != "" {
			b.WriteString(fmt.Sprintf("- **Year:** %s\n", r.Year))
		}
		if r.Format != "" {
			b.WriteString(fmt.Sprintf("- **Format:** %s\n", r.Format))
		}
		if r.Size != "" {
			b.WriteString(fmt.Sprintf("- **Size:** %s\n", r.Size))
		}
		if r.Publisher != "" {
			b.WriteString(fmt.Sprintf("- **Publisher:** %s\n", r.Publisher))
		}
		if r.Language != "" {
			b.WriteString(fmt.Sprintf("- **Language:** %s\n", r.Language))
		}
		b.WriteString("\n")
	}

	return b.String()
}
