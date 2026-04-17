package commands

import (
	"context"
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

type LibgenCollectionsCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*LibgenCollectionsCommand)(nil)
var _ cmds.WriterCommand = (*LibgenCollectionsCommand)(nil)

type LibgenCollectionsSettings struct {
	ID          string `glazed:"id"`
	URL         string `glazed:"url"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

func NewLibgenCollectionsCommand() (*LibgenCollectionsCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"collections",
		cmds.WithShort("Get collections a book belongs to on 1lib.sk"),
		cmds.WithLong("Get the booklists/collections that a book appears in on 1lib.sk."),
		cmds.WithFlags(
			fields.New("id", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Book ID")),
			fields.New("url", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Full 1lib.sk book URL")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the tab open")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Tab ID override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Window ID override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Debug socket traffic")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &LibgenCollectionsCommand{CommandDescription: desc}, nil
}

func buildLibgenCollectionsCode() string {
	return `
var result = {
  book: {
    title: '',
    author: ''
  },
  collections: []
};

// Get current book info
var heading = document.querySelector('h1') || document.querySelector('h2');
if (heading) result.book.title = heading.textContent.trim();

var authorLink = document.querySelector('a[href*="/author/"]');
if (authorLink) result.book.author = authorLink.textContent.trim();

// Look for z-booklist custom elements (used by 1lib.sk for booklists)
var booklistElements = document.querySelectorAll('z-booklist');
booklistElements.forEach(function(el) {
  var href = el.getAttribute('href');
  var id = el.getAttribute('id') || '';
  var topic = el.getAttribute('topic') || el.getAttribute('name') || '';
  var quantity = el.getAttribute('quantity') || '';
  
  // Get title from nested z-cover or img
  var cover = el.querySelector('z-cover');
  var title = '';
  if (cover) {
    title = cover.getAttribute('title') || '';
    if (!title) {
      var img = cover.querySelector('img');
      if (img) title = img.getAttribute('alt') || '';
    }
  }
  
  if (href && href.includes('/booklist/')) {
    result.collections.push({
      title: topic || title || 'Untitled',
      url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
      id: id,
      bookCount: quantity
    });
  }
});

// Also look for standard booklist links
if (result.collections.length === 0) {
  var allBooklistLinks = document.querySelectorAll('a[href*="/booklist/"]');
  allBooklistLinks.forEach(function(link) {
    var href = link.getAttribute('href');
    var idMatch = href.match(/\/booklist\/([^\/]+)/);
    var id = idMatch ? idMatch[1] : '';
    var title = link.textContent.trim();
    
    if (title && id && !title.includes('Login') && !title.includes('Sign')) {
      var exists = result.collections.some(function(c) { return c.id === id; });
      if (!exists) {
        result.collections.push({
          title: title,
          url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
          id: id,
          bookCount: ''
        });
      }
    }
  });
}

return result;
`
}

func (c *LibgenCollectionsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &LibgenCollectionsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenCollections(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range libgenCollectionsToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *LibgenCollectionsCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &LibgenCollectionsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenCollections(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderLibgenCollectionsMarkdown(data))
	return err
}

func fetchLibgenCollections(ctx context.Context, s *LibgenCollectionsSettings) (*libgenCollectionsData, error) {
	bookURL := s.URL
	if bookURL == "" && s.ID != "" {
		bookURL = "https://1lib.sk/book/" + s.ID
	}
	if bookURL == "" {
		return nil, fmt.Errorf("--id or --url is required")
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
		resolvedTabID, err := openOwnedTab(ctx, client, bookURL, tabReadyOptions{URLPrefix: "https://1lib.sk/book/"})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": bookURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: "https://1lib.sk/book/"}); err != nil {
				return nil, err
			}
		}
	}

	defer func() {
		if s.KeepTabOpen {
			return
		}
		closeOwnedTab(ctx, client, ownedTabID)
	}()

	// Wait for page to load and scroll to trigger lazy loading
	time.Sleep(5 * time.Second)

	// Scroll to trigger lazy loading for collections
	for i := 0; i < 8; i++ {
		_, _ = ExecuteTool(ctx, client, "js", map[string]any{"code": `window.scrollTo(0, document.body.scrollHeight); return document.body.scrollHeight;`}, tabID, windowID)
		time.Sleep(2 * time.Second)
		// Try to click "Load more" buttons
		_, _ = ExecuteTool(ctx, client, "js", map[string]any{"code": `
			var buttons = document.querySelectorAll('button');
			buttons.forEach(function(btn) {
				if (btn.textContent.includes('Load more') || btn.textContent.includes('Show more')) {
					btn.click();
				}
			});
			return 'clicked';
		`}, tabID, windowID)
	}
	_, _ = ExecuteTool(ctx, client, "js", map[string]any{"code": `window.scrollTo(0, 0);`}, tabID, windowID)
	time.Sleep(2 * time.Second)

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenCollectionsCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	return parseLibgenCollectionsResponse(resp), nil
}

type libgenCollectionsData struct {
	Book        libgenBookInfo
	Collections []libgenCollectionResult
}

type libgenCollectionResult struct {
	Title     string
	URL       string
	ID        string
	BookCount string
}

func parseLibgenCollectionsResponse(resp map[string]any) *libgenCollectionsData {
	result := &libgenCollectionsData{}
	parsed := parseResult(resp)
	if data, ok := parsed.Data.(map[string]any); ok {
		if book, ok := data["book"].(map[string]any); ok {
			result.Book.Title = getString(book, "title")
			result.Book.Author = getString(book, "author")
		}
		if collections, ok := data["collections"].([]any); ok {
			for _, c := range collections {
				if m, ok := c.(map[string]any); ok {
					result.Collections = append(result.Collections, libgenCollectionResult{
						Title:     getString(m, "title"),
						URL:       getString(m, "url"),
						ID:        getString(m, "id"),
						BookCount: getString(m, "bookCount"),
					})
				}
			}
		}
	}
	return result
}

func libgenCollectionsToRows(data *libgenCollectionsData) []types.Row {
	rows := []types.Row{}
	for _, c := range data.Collections {
		row := map[string]any{
			"title":      c.Title,
			"url":        c.URL,
			"id":         c.ID,
			"book_count": c.BookCount,
		}
		rows = append(rows, types.NewRowFromMap(row))
	}
	return rows
}

func renderLibgenCollectionsMarkdown(data *libgenCollectionsData) string {
	var b strings.Builder
	b.WriteString("# Book Collections\n\n")
	b.WriteString(fmt.Sprintf("Book: **%s**", data.Book.Title))
	if data.Book.Author != "" {
		b.WriteString(fmt.Sprintf(" by %s", data.Book.Author))
	}
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Appears in %d collections\n\n", len(data.Collections)))

	for i, c := range data.Collections {
		b.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, c.Title))
		b.WriteString(fmt.Sprintf("- [View collection](%s)\n", c.URL))
		b.WriteString(fmt.Sprintf("- **ID:** `%s`\n", c.ID))
		if c.BookCount != "" {
			b.WriteString(fmt.Sprintf("- **Books:** %s\n", c.BookCount))
		}
		b.WriteString("\n")
	}

	return b.String()
}
