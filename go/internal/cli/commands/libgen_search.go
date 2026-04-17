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
	Query        string `glazed:"query"`
	Limit        int    `glazed:"limit"`
	KeepTabOpen  bool   `glazed:"keep-tab-open"`
	Socket       string `glazed:"socket-path"`
	TimeoutMS    int    `glazed:"timeout-ms"`
	TabID        int64  `glazed:"tab-id"`
	WindowID     int64  `glazed:"window-id"`
	DebugSocket  bool   `glazed:"debug-socket"`
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

// Parse titles from body text - the search results are numbered
var bodyText = document.body.innerText;

// Pattern: look for numbered entries followed by titles
// The page shows "1\nTitle" format
var numberPattern = /(\d+)\s*\n([^\n]{5,300})/g;
var match;
var count = 0;
while ((match = numberPattern.exec(bodyText)) !== null && count < maxResults) {
  var num = parseInt(match[1]);
  var title = match[2].trim();
  
  // Skip headers, footers, and navigation
  if (title.length < 5 || title.length > 300) continue;
  if (title.match(/^(Z-Library|Home|My Library|Donate|Log In|Search|General|Full-Text|Your gateway|Are you familiar)/)) continue;
  if (title.match(/^\d+\s*$/)) continue;
  // Skip noise: email addresses, single words, short titles
  if (title.match(/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/)) continue;
  if (title.match(/^[a-zA-Z]+$/)) continue;
  if (title.match(/^(By signing|To build|Sign in|Blog|Official)/)) continue;
  
  // Skip if already seen
  var titleKey = title.toLowerCase().substring(0, 50);
  if (seen[titleKey]) continue;
  seen[titleKey] = true;
  
  // Extract metadata from surrounding text (look ahead to next few lines)
  var startIdx = match.index;
  var endIdx = Math.min(bodyText.length, match.index + 500);
  var metaText = bodyText.substring(startIdx, endIdx);
  
  // Look for author (usually in "by Author" format after title)
  var author = '';
  var authorMatch = metaText.match(/by\s+([^\n,]+)/i);
  if (authorMatch) author = authorMatch[1].trim();
  
  // Look for format, size, year - usually in same line or next few lines
  // Format often appears as "PDF" or "EPUB" before size
  var formatMatch = metaText.match(/\b(PDF|EPUB|MOBI|AZW3|FB2|DOCX?)\b/i);
  var sizeMatch = metaText.match(/(\d+\.?\d*)\s*(MB|KB|GB)/i);
  var yearMatch = metaText.match(/\b(19\d{2}|20\d{2})\b/);
  
  // Build URL - prefer actual book link if found nearby
  var url = 'https://1lib.sk/s/' + encodeURIComponent(title);
  var bookLinkMatch = metaText.match(/\/book\/([a-zA-Z0-9]+)/);
  if (bookLinkMatch) {
    url = 'https://1lib.sk/book/' + bookLinkMatch[1];
  }
  
  results.push({
    num: num,
    title: title,
    author: author,
    url: url,
    id: bookLinkMatch ? bookLinkMatch[1] : '',
    year: yearMatch ? yearMatch[1] : '',
    size: sizeMatch ? sizeMatch[0] : '',
    format: formatMatch ? formatMatch[0] : ''
  });
  count++;
}

// Also look for book items with known selectors
var resItems = document.querySelectorAll('.res-item');
resItems.forEach(function(item) {
  if (results.length >= maxResults) return;
  
  var link = item.querySelector('a[href*="/book/"]');
  if (!link) return;
  
  var href = link.getAttribute('href');
  var idMatch = href.match(/\/book\/([^\/]+)/);
  var id = idMatch ? idMatch[1] : '';
  
  var title = link.textContent.trim() || link.querySelector('span')?.textContent?.trim() || '';
  if (!title) {
    var h3 = item.querySelector('h3');
    if (h3) title = h3.textContent.trim();
  }
  
  if (!title || title.length < 3) return;
  
  var titleKey = title.toLowerCase().substring(0, 50);
  if (seen[titleKey]) return;
  seen[titleKey] = true;
  
  var authorLink = item.querySelector('a[href*="/author/"]');
  var author = authorLink ? authorLink.textContent.trim() : '';
  
  var metaDiv = item.querySelector('.res-add') || item.querySelector('.bookMeta');
  var meta = metaDiv ? metaDiv.textContent : '';
  var yearMatch = meta.match(/(\d{4})/);
  var sizeMatch = meta.match(/(\d+\.?\d*)\s*(MB|KB|GB)/i);
  var formatMatch = meta.match(/\b(PDF|EPUB|MOBI|AZW3|FB2|DOCX?)\b/i);
  
  results.push({
    title: title,
    author: author,
    url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
    id: id,
    year: yearMatch ? yearMatch[1] : '',
    size: sizeMatch ? sizeMatch[0] : '',
    format: formatMatch ? formatMatch[0] : ''
  });
});

// Look for any book links
if (results.length === 0) {
  var links = document.querySelectorAll('a[href*="/book/"]');
  links.forEach(function(link) {
    if (results.length >= maxResults) return;
    
    var href = link.getAttribute('href');
    var idMatch = href.match(/\/book\/([^\/]+)/);
    var id = idMatch ? idMatch[1] : '';
    
    var title = link.textContent.trim();
    var titleSpan = link.querySelector('span');
    if (titleSpan) {
      var spanText = titleSpan.textContent.trim();
      if (spanText) title = spanText;
    }
    
    if (!title || title.length < 5 || title.length > 400) return;
    if (title.match(/^(Log In|Home|Library|Donate|Search|Z-Library|\+[\d]+|Results?|Most Popular|Booklists?)/)) return;
    
    results.push({
      title: title,
      author: '',
      url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
      id: id
    });
  });
}

return {
  count: results.length,
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

	// Wait for page to load - longer for potential CloudFlare challenge
	time.Sleep(8 * time.Second)

	// Scroll to trigger lazy loading and wait for results
	for i := 0; i < 10; i++ {
		// Scroll down to trigger lazy loading
		_, _ = ExecuteTool(ctx, client, "js", map[string]any{"code": `window.scrollTo(0, document.body.scrollHeight); window.scrollTo(0, document.body.scrollHeight / 2); return document.body.scrollHeight;`}, tabID, windowID)
		time.Sleep(2 * time.Second)

		// Check if results are visible
		resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": `return { bookLinks: document.querySelectorAll('a[href*="/book/"]').length, resItems: document.querySelectorAll('.res-item').length };`}, tabID, windowID)
		if err == nil {
			parsed := parseResult(resp)
			if data, ok := parsed.Data.(map[string]any); ok {
				if bookLinks, ok := data["bookLinks"].(float64); ok && bookLinks > 0 {
					if s.DebugSocket {
						fmt.Fprintf(os.Stderr, "DEBUG: Found %d book links after scroll\n", int(bookLinks))
					}
					break
				}
				if resItems, ok := data["resItems"].(float64); ok && resItems > 0 {
					if s.DebugSocket {
						fmt.Fprintf(os.Stderr, "DEBUG: Found %d res-items after scroll\n", int(resItems))
					}
					break
				}
			}
		}
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
					if year, ok := rm["year"].(string); ok {
						book.Year = year
					}
					if size, ok := rm["size"].(string); ok {
						book.Size = size
					}
					if format, ok := rm["format"].(string); ok {
						book.Format = format
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
		if r.ID != "" {
			b.WriteString(fmt.Sprintf("- **ID:** `%s`\n\n", r.ID))
		}
	}

	return b.String()
}
