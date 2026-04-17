package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

type LibgenDownloadCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*LibgenDownloadCommand)(nil)
var _ cmds.WriterCommand = (*LibgenDownloadCommand)(nil)

type LibgenDownloadSettings struct {
	ID         string `glazed:"id"`
	URL        string `glazed:"url"`
	SaveTo     string `glazed:"save-to"`
	KeepTabOpen bool  `glazed:"keep-tab-open"`
	Socket     string `glazed:"socket-path"`
	TimeoutMS  int    `glazed:"timeout-ms"`
	TabID      int64  `glazed:"tab-id"`
	WindowID   int64  `glazed:"window-id"`
	DebugSocket bool  `glazed:"debug-socket"`
}

func NewLibgenDownloadCommand() (*LibgenDownloadCommand, error) {
	glazedSection, err := NewGlazedSchemaWithYAMLDefault()
	if err != nil {
		return nil, err
	}
	commandSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"download",
		cmds.WithShort("Download a book from 1lib.sk"),
		cmds.WithLong("Get download link or download a book from 1lib.sk by ID or URL."),
		cmds.WithFlags(
			fields.New("id", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Book ID (from search results)")),
			fields.New("url", fields.TypeString, fields.WithRequired(false), fields.WithHelp("Full 1lib.sk book URL")),
			fields.New("save-to", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Save downloaded file to this path")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the tab open when finished")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Tab ID override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Window ID override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Debug socket traffic")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &LibgenDownloadCommand{CommandDescription: desc}, nil
}

func buildLibgenDownloadCode() string {
	return `
var result = {
  title: '',
  author: '',
  year: '',
  format: '',
  size: '',
  downloadUrl: '',
  bookId: ''
};

// Get title
var heading = document.querySelector('h1') || document.querySelector('h2');
if (heading) result.title = heading.textContent.trim();

// Get author
var authorLink = document.querySelector('a[href*="/author/"]');
if (authorLink) result.author = authorLink.textContent.trim();

// Get book ID from URL
var idMatch = location.pathname.match(/\/book\/([^\/]+)/);
if (idMatch) result.bookId = idMatch[1];

// Get metadata
var metaItems = document.querySelectorAll('.param');
metaItems.forEach(function(item) {
  var text = item.textContent;
  if (text.includes('Year:')) result.year = text.replace('Year:', '').trim();
  if (text.includes('File:')) {
    var match = text.match(/([^\,]+),\s*([\d.]+\s*(MB|KB|GB))/i);
    if (match) {
      result.format = match[1].trim();
      result.size = match[2].trim();
    }
  }
});

// Get download link
var dlLink = document.querySelector('a[href*="/dl/"]');
if (dlLink) {
  var href = dlLink.getAttribute('href');
  result.downloadUrl = href.startsWith('http') ? href : 'https://1lib.sk' + href;
}

// Fallback: find any /dl/ link
if (!result.downloadUrl) {
  var allLinks = document.querySelectorAll('a[href]');
  for (var i = 0; i < allLinks.length; i++) {
    var href = allLinks[i].getAttribute('href');
    if (href && href.match(/^\/dl\//)) {
      result.downloadUrl = 'https://1lib.sk' + href;
      break;
    }
  }
}

return result;
`
}

func (c *LibgenDownloadCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &LibgenDownloadSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenDownload(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range libgenDownloadToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *LibgenDownloadCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &LibgenDownloadSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchLibgenDownload(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderLibgenDownloadMarkdown(data))
	return err
}

func fetchLibgenDownload(ctx context.Context, s *LibgenDownloadSettings) (*libgenDownloadData, error) {
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
		resolvedTabID, err := openOwnedTab(ctx, client, bookURL, tabReadyOptions{
			URLPrefix: "https://1lib.sk/book/",
		})
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
		if err := closeOwnedTab(ctx, client, ownedTabID); err != nil {
			// ignore
		}
	}()

	time.Sleep(2 * time.Second)

	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenDownloadCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	result := parseLibgenDownloadResponse(resp)

	// Download if save-to is specified
	if s.SaveTo != "" && result.DownloadURL != "" {
		if err := downloadLibgenFile(s.SaveTo, result.DownloadURL, result.Title); err != nil {
			return nil, err
		}
	}

	return result, nil
}

type libgenDownloadData struct {
	Title       string
	Author      string
	Year        string
	Format      string
	Size        string
	DownloadURL string
	BookID      string
}

func parseLibgenDownloadResponse(resp map[string]any) *libgenDownloadData {
	result := &libgenDownloadData{}
	parsed := parseResult(resp)
	if data, ok := parsed.Data.(map[string]any); ok {
		result.Title = getString(data, "title")
		result.Author = getString(data, "author")
		result.Year = getString(data, "year")
		result.Format = getString(data, "format")
		result.Size = getString(data, "size")
		result.DownloadURL = getString(data, "downloadUrl")
		result.BookID = getString(data, "bookId")
	}
	return result
}

func libgenDownloadToRows(data *libgenDownloadData) []types.Row {
	row := map[string]any{
		"title":         data.Title,
		"author":        data.Author,
		"year":          data.Year,
		"format":        data.Format,
		"size":          data.Size,
		"download_url":  data.DownloadURL,
		"id":            data.BookID,
	}
	return []types.Row{types.NewRowFromMap(row)}
}

func renderLibgenDownloadMarkdown(data *libgenDownloadData) string {
	var b strings.Builder
	b.WriteString("# 1lib.sk Book\n\n")
	b.WriteString(fmt.Sprintf("## %s\n\n", data.Title))
	if data.Author != "" {
		b.WriteString(fmt.Sprintf("**Author:** %s\n\n", data.Author))
	}
	b.WriteString(fmt.Sprintf("- **ID:** `%s`\n", data.BookID))
	if data.Year != "" {
		b.WriteString(fmt.Sprintf("- **Year:** %s\n", data.Year))
	}
	if data.Format != "" {
		b.WriteString(fmt.Sprintf("- **Format:** %s\n", data.Format))
	}
	if data.Size != "" {
		b.WriteString(fmt.Sprintf("- **Size:** %s\n", data.Size))
	}
	b.WriteString("\n## Download\n\n")
	if data.DownloadURL != "" {
		b.WriteString(fmt.Sprintf("[Download](%s)\n\n", data.DownloadURL))
		b.WriteString(fmt.Sprintf("```\n%s\n```\n", data.DownloadURL))
	} else {
		b.WriteString("*No download link found*\n")
	}
	return b.String()
}

func downloadLibgenFile(path, url, title string) error {
	fmt.Fprintf(os.Stderr, "Downloading: %s\n", url)

	httpClient := &http.Client{Timeout: 300 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	outputPath := path
	if outputPath == "" {
		if title != "" {
			outputPath = sanitizeFilename(title) + ".pdf"
		} else {
			outputPath = "download"
		}
	}

	if !strings.Contains(outputPath, ".") {
		// Try to determine extension from content-type
		ext := getExtension(resp.Header.Get("Content-Type"))
		outputPath = outputPath + ext
	}

	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Saved to: %s\n", outputPath)
	return nil
}

func getExtension(contentType string) string {
	switch {
	case strings.Contains(contentType, "pdf"):
		return ".pdf"
	case strings.Contains(contentType, "epub"):
		return ".epub"
	case strings.Contains(contentType, "mobi"):
		return ".mobi"
	case strings.Contains(contentType, "zip"):
		return ".zip"
	default:
		return ""
	}
}

