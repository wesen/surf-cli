package commands

import (
	"context"
	"fmt"
	"io"
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
	ID          string `glazed:"id"`
	URL         string `glazed:"url"`
	SaveTo      string `glazed:"save-to"`
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
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

// Get all author links
var authorLinks = document.querySelectorAll('a[href*="/author/"]');
var authors = [];
authorLinks.forEach(function(a) { authors.push(a.textContent.trim()); });
result.author = authors.join(', ');

// Get book ID from URL
var idMatch = location.pathname.match(/\/book\/([^\/]+)/);
if (idMatch) result.bookId = idMatch[1];

// Get download link
var dlLink = document.querySelector('a[href*="/dl/"]');
if (dlLink) {
  var href = dlLink.getAttribute('href');
  result.downloadUrl = href.startsWith('http') ? href : 'https://1lib.sk' + href;
  // Extract format and size from download link text
  var dlText = dlLink.textContent.trim();
  var fmtMatch = dlText.match(/(\w+),\s*([\d.]+\s*(?:MB|KB|GB))/i);
  if (fmtMatch) {
    result.format = fmtMatch[1].toUpperCase();
    result.size = fmtMatch[2];
  }
}

// Get metadata from detail sections
var detailsContainer = document.querySelector('.bookDetailsBox, [class*="book-details"]');
if (detailsContainer) {
  var text = detailsContainer.textContent;
  if (!result.year) {
    var ym = text.match(/Year:\s*(\d{4})/);
    if (ym) result.year = ym[1];
  }
}

// Fallback: scan all text nodes for Year/File metadata
if (!result.year || !result.format) {
  var bodyText = document.body.innerText;
  if (!result.year) {
    var ym = bodyText.match(/Year:\s*(\d{4})/);
    if (ym) result.year = ym[1];
  }
  if (!result.format || !result.size) {
    var fm = bodyText.match(/File:\s*(\w+),\s*([\d.]+\s*(?:MB|KB|GB))/i);
    if (fm) {
      if (!result.format) result.format = fm[1].toUpperCase();
      if (!result.size) result.size = fm[2];
    }
  }
}

// Fallback download link
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

func buildLibgenDownloadCheckCode() string {
	return `
var h1 = document.querySelector('h1');
var result = {
  url: location.href,
  title: h1 ? h1.textContent.trim() : '',
  hasDownloadButton: !!document.querySelector('a[href*="/dl/"]'),
  hasError: false,
  errorType: ''
};

if (h1) {
  var text = h1.textContent.trim().toLowerCase();
  if (text.includes('daily limit') || text.includes('limit reached')) {
    result.hasError = true;
    result.errorType = 'daily_limit';
  } else if (text.includes('not found') || text.includes('error')) {
    result.hasError = true;
    result.errorType = 'not_found';
  }
}

// Check for login requirement — only if the link is actually visible
function isVisible(el) {
  if (!el) return false;
  var style = getComputedStyle(el);
  return style.display !== 'none' &&
         style.visibility !== 'hidden' &&
         style.opacity !== '0' &&
         el.offsetWidth > 0 &&
         el.offsetHeight > 0;
}

var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink && isVisible(loginLink)) {
  result.hasError = true;
  result.errorType = 'login_required';
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

	// Wait for the book detail page to render
	time.Sleep(2 * time.Second)

	// Extract book metadata and download URL
	resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenDownloadCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	result := parseLibgenDownloadResponse(resp)

	// If --save-to is specified, trigger the download via the browser
	if s.SaveTo != "" && result.DownloadURL != "" {
		saveToPath, err := expandDownloadPath(s.SaveTo, result)
		if err != nil {
			return nil, err
		}

		if err := browserDownloadFile(ctx, client, tabID, windowID, result.DownloadURL, saveToPath, s.DebugSocket); err != nil {
			result.DownloadError = err.Error()
		} else {
			result.SavedTo = saveToPath
		}
	}

	return result, nil
}

func expandDownloadPath(path string, result *libgenDownloadData) (string, error) {
	expanded := os.ExpandEnv(path)
	if strings.HasPrefix(expanded, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		expanded = filepath.Join(home, expanded[2:])
	}

	// If the path has no extension, derive one from the format
	if !strings.Contains(filepath.Base(expanded), ".") {
		ext := ".pdf"
		if result.Format != "" {
			ext = "." + strings.ToLower(result.Format)
		}
		expanded = expanded + ext
	}

	// Ensure parent directory exists
	dir := filepath.Dir(expanded)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return expanded, nil
}

func browserDownloadFile(ctx context.Context, client *transport.Client, tabID, windowID *int64, downloadURL, saveToPath string, debug bool) error {
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Navigating browser to download URL: %s\n", downloadURL)
	}

	// Navigate to the download URL - the browser will handle the download
	// using its own session/cookies
	if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": downloadURL}, tabID, windowID); err != nil {
		return fmt.Errorf("failed to navigate to download URL: %w", err)
	}

	// Wait for the download page to load
	time.Sleep(3 * time.Second)

	// Check if the page shows an error (daily limit, login required, etc.)
	checkResp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildLibgenDownloadCheckCode()}, tabID, windowID)
	if err == nil {
		parsed := parseResult(checkResp)
		if data, ok := parsed.Data.(map[string]any); ok {
			if hasError, ok := data["hasError"].(bool); ok && hasError {
				errorType, _ := data["errorType"].(string)
				switch errorType {
				case "daily_limit":
					return fmt.Errorf("daily download limit reached: log in to your Z-Library account in the browser or wait 24 hours")
				case "login_required":
					return fmt.Errorf("login required: please log in to your Z-Library account in the browser first")
				default:
					title, _ := data["title"].(string)
					return fmt.Errorf("download page shows error: %s", title)
				}
			}
		}
	}

	// The browser is downloading the file. Wait for it to appear.
	// Determine the browser's default downloads directory and watch for the file.
	downloadsDir, err := getBrowserDownloadsDir()
	if err != nil {
		return fmt.Errorf("download started in browser but could not locate downloaded file: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Watching for download in: %s\n", downloadsDir)
	}

	// Record time before we start looking
	lookAfter := time.Now().Add(-2 * time.Second)

	// Wait up to 60 seconds for the download to complete
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		// Look for recently created .crdownload files (Chrome partial downloads)
		// When these disappear, the download is complete
		crdownloads, _ := filepath.Glob(filepath.Join(downloadsDir, "*.crdownload"))
		if len(crdownloads) == 0 {
			// No partial downloads - check for the completed file
			candidate := findRecentFile(downloadsDir, lookAfter)
			if candidate != "" {
				if debug {
					fmt.Fprintf(os.Stderr, "DEBUG: Found downloaded file: %s\n", candidate)
				}
				// Move to the target path
				if err := os.Rename(candidate, saveToPath); err != nil {
					// If rename fails (cross-device), copy then remove
					if err := copyFile(candidate, saveToPath); err != nil {
						return fmt.Errorf("downloaded but failed to move to %s: %w", saveToPath, err)
					}
					os.Remove(candidate)
				}
				fmt.Fprintf(os.Stderr, "Saved to: %s\n", saveToPath)
				return nil
			}
		} else if debug {
			fmt.Fprintf(os.Stderr, "DEBUG: Download in progress (%d partial files)\n", len(crdownloads))
		}
	}

	return fmt.Errorf("timed out waiting for browser download to complete (check %s)", downloadsDir)
}

func getBrowserDownloadsDir() (string, error) {
	// Check XDG_DOWNLOAD_DIR first
	if xdgDir := os.Getenv("XDG_DOWNLOAD_DIR"); xdgDir != "" {
		if _, err := os.Stat(xdgDir); err == nil {
			return xdgDir, nil
		}
	}

	// Check user-dirs.dirs
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Common download directories
	candidates := []string{
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "downloads"),
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
	}

	return "", fmt.Errorf("could not find downloads directory")
}

func findRecentFile(dir string, after time.Time) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var bestMatch string
	var bestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		modTime := info.ModTime()
		if modTime.After(after) && modTime.After(bestTime) {
			// Prefer PDF/EPUB files
			name := strings.ToLower(entry.Name())
			if strings.HasSuffix(name, ".pdf") || strings.HasSuffix(name, ".epub") ||
				strings.HasSuffix(name, ".mobi") || strings.HasSuffix(name, ".djvu") {
				bestMatch = filepath.Join(dir, entry.Name())
				bestTime = modTime
			}
		}
	}

	return bestMatch
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

type libgenDownloadData struct {
	Title        string
	Author       string
	Year         string
	Format       string
	Size         string
	DownloadURL  string
	BookID       string
	SavedTo      string
	DownloadError string
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
		"title":        data.Title,
		"author":       data.Author,
		"year":         data.Year,
		"format":       data.Format,
		"size":         data.Size,
		"download_url": data.DownloadURL,
		"id":           data.BookID,
	}
	if data.SavedTo != "" {
		row["saved_to"] = data.SavedTo
	}
	if data.DownloadError != "" {
		row["error"] = data.DownloadError
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
	if data.SavedTo != "" {
		b.WriteString(fmt.Sprintf("✅ Saved to: `%s`\n", data.SavedTo))
	} else if data.DownloadError != "" {
		b.WriteString(fmt.Sprintf("❌ %s\n", data.DownloadError))
	} else if data.DownloadURL != "" {
		b.WriteString(fmt.Sprintf("[Download](%s)\n\n", data.DownloadURL))
		b.WriteString(fmt.Sprintf("```\n%s\n```\n", data.DownloadURL))
		if data.Format == "" {
			b.WriteString("\n*Use `--save-to <path>` to download the file automatically via the browser.*\n")
		}
	} else {
		b.WriteString("*No download link found*\n")
	}
	return b.String()
}
