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

type LibgenCollectionCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*LibgenCollectionCommand)(nil)
var _ cmds.WriterCommand = (*LibgenCollectionCommand)(nil)

type LibgenCollectionSettings struct {
	ID          string `glazed:"id"`
	URL         string `glazed:"url"`
	Limit       int    `glazed:"limit"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

func NewLibgenCollectionCommand() (*LibgenCollectionCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"collection",
		cmds.WithShort("Get books in a collection on 1lib.sk"),
		cmds.WithLong("Get all books in a booklist/collection on 1lib.sk."),
		cmds.WithFlags(
			fields.New("id", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Collection ID (from URL)")),
			fields.New("url", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Full 1lib.sk collection URL")),
			fields.New("limit", fields.TypeInteger, fields.WithDefault(50), fields.WithHelp("Maximum number of books")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the tab open")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Tab ID override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Window ID override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Debug socket traffic")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &LibgenCollectionCommand{CommandDescription: desc}, nil
}

func buildLibgenCollectionCode() string {
	return `
var result = {
  collection: {
    title: '',
    category: ''
  },
  books: []
};

// Get collection title
var heading = document.querySelector('h1') || document.querySelector('h2');
if (heading) result.collection.title = heading.textContent.trim();

// Try to get category from breadcrumb
var breadcrumbs = document.querySelectorAll('a[class*="breadcrumb"]');
breadcrumbs.forEach(function(crumb) {
  var text = crumb.textContent.trim();
  if (text && text !== result.collection.title) {
    result.collection.category = text;
  }
});

// Find book links
var maxBooks = 100;
var seen = {};

// Look for book items
var bookItems = document.querySelectorAll('a[href*="/book/"]');
bookItems.forEach(function(link) {
  if (result.books.length >= maxBooks) return;
  
  var href = link.getAttribute('href');
  var idMatch = href.match(/\/book\/([^\/]+)/);
  var id = idMatch ? idMatch[1] : '';
  
  if (seen[id]) return;
  seen[id] = true;
  
  // Get title - try to find within the book item container
  var title = '';
  var parent = link.closest('div') || link.parentElement;
  if (parent) {
    var titleEl = parent.querySelector('span') || link;
    title = titleEl.textContent.trim();
  }
  if (!title || title.length > 300) {
    title = link.textContent.trim();
  }
  
  // Skip if looks like UI element
  if (!title || title.length < 3 || title.includes('Load more')) return;
  
  // Get author
  var author = '';
  if (parent) {
    var authorLink = parent.querySelector('a[href*="/author/"]');
    if (authorLink) author = authorLink.textContent.trim();
  }
  
  result.books.push({
    title: title,
    author: author,
    url: href.startsWith('http') ? href : 'https://1lib.sk' + href,
    id: id
  });
});

return result;
`
}

func (c *LibgenCollectionCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &LibgenCollectionSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenCollection(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range libgenCollectionToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *LibgenCollectionCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &LibgenCollectionSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenCollection(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderLibgenCollectionMarkdown(data))
	return err
}

func fetchLibgenCollection(ctx context.Context, s *LibgenCollectionSettings) (*libgenCollectionData, error) {
	collectionURL := s.URL
	if collectionURL == "" && s.ID != "" {
		collectionURL = "https://1lib.sk/booklist/" + s.ID
	}
	if collectionURL == "" {
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
		resolvedTabID, err := openOwnedTab(ctx, client, collectionURL, tabReadyOptions{URLPrefix: "https://1lib.sk/booklist/"})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": collectionURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: "https://1lib.sk/booklist/"}); err != nil {
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

	time.Sleep(2 * time.Second)

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenCollectionCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	return parseLibgenCollectionResponse(resp), nil
}

type libgenCollectionData struct {
	Collection libgenCollectionInfo
	Books      []libgenBookResult
}

type libgenCollectionInfo struct {
	Title    string
	Category string
}

func parseLibgenCollectionResponse(resp map[string]any) *libgenCollectionData {
	result := &libgenCollectionData{}
	parsed := parseResult(resp)
	if data, ok := parsed.Data.(map[string]any); ok {
		if collection, ok := data["collection"].(map[string]any); ok {
			result.Collection.Title = getString(collection, "title")
			result.Collection.Category = getString(collection, "category")
		}
		if books, ok := data["books"].([]any); ok {
			for _, b := range books {
				if m, ok := b.(map[string]any); ok {
					result.Books = append(result.Books, libgenBookResult{
						Title:  getString(m, "title"),
						Author: getString(m, "author"),
						URL:    getString(m, "url"),
						ID:     getString(m, "id"),
					})
				}
			}
		}
	}
	return result
}

func libgenCollectionToRows(data *libgenCollectionData) []types.Row {
	rows := []types.Row{}
	for _, b := range data.Books {
		row := map[string]any{
			"title":  b.Title,
			"author": b.Author,
			"url":    b.URL,
			"id":     b.ID,
		}
		rows = append(rows, types.NewRowFromMap(row))
	}
	return rows
}

func renderLibgenCollectionMarkdown(data *libgenCollectionData) string {
	var b strings.Builder
	b.WriteString("# Collection: ")
	b.WriteString(data.Collection.Title)
	b.WriteString("\n\n")
	if data.Collection.Category != "" {
		b.WriteString(fmt.Sprintf("**Category:** %s\n\n", data.Collection.Category))
	}
	b.WriteString(fmt.Sprintf("Found %d books\n\n", len(data.Books)))

	for i, book := range data.Books {
		b.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, book.Title))
		if book.Author != "" {
			b.WriteString(fmt.Sprintf("**Author:** %s\n\n", book.Author))
		}
		b.WriteString(fmt.Sprintf("- [View](%s)\n", book.URL))
		b.WriteString(fmt.Sprintf("- **ID:** `%s`\n\n", book.ID))
	}

	return b.String()
}
