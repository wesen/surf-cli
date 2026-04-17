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

type LibgenSuggestionsCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*LibgenSuggestionsCommand)(nil)
var _ cmds.WriterCommand = (*LibgenSuggestionsCommand)(nil)

type LibgenSuggestionsSettings struct {
	ID          string `glazed:"id"`
	URL         string `glazed:"url"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

func NewLibgenSuggestionsCommand() (*LibgenSuggestionsCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"suggestions",
		cmds.WithShort("Get suggested books for a book on 1lib.sk"),
		cmds.WithLong("Get recommended/suggested books based on a book on 1lib.sk."),
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

	return &LibgenSuggestionsCommand{CommandDescription: desc}, nil
}

func buildLibgenSuggestionsCode() string {
	return `
var result = {
  book: {
    title: '',
    author: ''
  },
  suggestions: []
};

// Get current book info
var heading = document.querySelector('h1') || document.querySelector('h2');
if (heading) result.book.title = heading.textContent.trim();

var authorLink = document.querySelector('a[href*="/author/"]');
if (authorLink) result.book.author = authorLink.textContent.trim();

// Find "You may be interested in" section
var sections = document.querySelectorAll('h2');
sections.forEach(function(section) {
  var text = section.textContent.trim();
  if (text.includes('interested') || text.includes('recommend')) {
    var container = section.closest('div');
    if (container) {
      var links = container.querySelectorAll('a[href*="/book/"]');
      links.forEach(function(link) {
        var href = link.getAttribute('href');
        var idMatch = href.match(/\/book\/([^\/]+)/);
        var id = idMatch ? idMatch[1] : '';
        
        // Get title from the element
        var titleEl = link.querySelector('span') || link;
        var title = titleEl.textContent.trim();
        
        // Get author from nearby element
        var parent = link.closest('div') || link.parentElement;
        var author = '';
        var authorLink = parent ? parent.querySelector('a[href*="/author/"]') : null;
        if (authorLink) author = authorLink.textContent.trim();
        
        if (title && !title.includes('Downloaded') && id) {
          result.suggestions.push({
            title: title,
            author: author,
            url: 'https://1lib.sk' + href,
            id: id
          });
        }
      });
    }
  }
});

return result;
`
}

func (c *LibgenSuggestionsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &LibgenSuggestionsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenSuggestions(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range libgenSuggestionsToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *LibgenSuggestionsCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &LibgenSuggestionsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenSuggestions(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderLibgenSuggestionsMarkdown(data))
	return err
}

func fetchLibgenSuggestions(ctx context.Context, s *LibgenSuggestionsSettings) (*libgenSuggestionsData, error) {
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

	time.Sleep(2 * time.Second)

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenSuggestionsCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	return parseLibgenSuggestionsResponse(resp), nil
}

type libgenSuggestionsData struct {
	Book        libgenBookInfo
	Suggestions []libgenBookResult
}

type libgenBookInfo struct {
	Title  string
	Author string
}

func parseLibgenSuggestionsResponse(resp map[string]any) *libgenSuggestionsData {
	result := &libgenSuggestionsData{}
	parsed := parseResult(resp)
	if data, ok := parsed.Data.(map[string]any); ok {
		if book, ok := data["book"].(map[string]any); ok {
			result.Book.Title = getString(book, "title")
			result.Book.Author = getString(book, "author")
		}
		if suggestions, ok := data["suggestions"].([]any); ok {
			for _, s := range suggestions {
				if m, ok := s.(map[string]any); ok {
					result.Suggestions = append(result.Suggestions, libgenBookResult{
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

func libgenSuggestionsToRows(data *libgenSuggestionsData) []types.Row {
	rows := []types.Row{}
	for _, r := range data.Suggestions {
		row := map[string]any{
			"title":  r.Title,
			"author": r.Author,
			"url":    r.URL,
			"id":     r.ID,
		}
		rows = append(rows, types.NewRowFromMap(row))
	}
	return rows
}

func renderLibgenSuggestionsMarkdown(data *libgenSuggestionsData) string {
	var b strings.Builder
	b.WriteString("# Suggested Books\n\n")
	b.WriteString(fmt.Sprintf("Based on: **%s**", data.Book.Title))
	if data.Book.Author != "" {
		b.WriteString(fmt.Sprintf(" by %s", data.Book.Author))
	}
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Found %d suggestions\n\n", len(data.Suggestions)))

	for i, s := range data.Suggestions {
		b.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, s.Title))
		if s.Author != "" {
			b.WriteString(fmt.Sprintf("**Author:** %s\n\n", s.Author))
		}
		b.WriteString(fmt.Sprintf("- [View](%s)\n", s.URL))
		b.WriteString(fmt.Sprintf("- **ID:** `%s`\n\n", s.ID))
	}

	return b.String()
}
