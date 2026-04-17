package commands

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
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

//go:embed scripts/annas_archive_download.js
var annasArchiveDownloadScript string

type AnnasArchiveDownloadCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*AnnasArchiveDownloadCommand)(nil)
var _ cmds.WriterCommand = (*AnnasArchiveDownloadCommand)(nil)

type AnnasArchiveDownloadSettings struct {
	DOI         string `glazed:"doi"`
	MirrorType  string `glazed:"mirror"`        // "slow" only (fast mirrors require membership)
	MirrorIndex int    `glazed:"mirror-index"` // specific server index (0-7 for slow)
	ListMirrors bool   `glazed:"list-mirrors"`
Output      string `glazed:"save-to"`      // output file path for PDF download
	KeepTabOpen bool   `glazed:"keep-tab-open"`
	Socket      string `glazed:"socket-path"`
	TimeoutMS   int    `glazed:"timeout-ms"`
	TabID       int64  `glazed:"tab-id"`
	WindowID    int64  `glazed:"window-id"`
	DebugSocket bool   `glazed:"debug-socket"`
}

type MirrorInfo struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	URL   string `json:"url"`
	Label string `json:"label"`
}

type annasArchiveDownloadData struct {
	Raw         map[string]any
	MD5         string
	Title       string
	Metadata    map[string]any
	Mirrors     []MirrorInfo
	DownloadURL string `json:"download_url"`
}

func NewAnnasArchiveDownloadCommand() (*AnnasArchiveDownloadCommand, error) {
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
		cmds.WithShort("Download a paper from Anna's Archive by DOI"),
		cmds.WithLong("Downloads a paper from Anna's Archive by DOI. Use --list-mirrors to see available mirrors and --mirror to select slow mirror. By default selects a random slow mirror."),
		cmds.WithFlags(
			fields.New("doi", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Paper DOI (e.g., 10.1038/nature12373)")),
			fields.New("mirror", fields.TypeString, fields.WithDefault("slow"), fields.WithHelp("Mirror type: 'slow' only (fast mirrors require membership)")),
			fields.New("mirror-index", fields.TypeInteger, fields.WithDefault(-1), fields.WithHelp("Specific mirror index (-1 = random). Slow: 0-7")),
			fields.New("list-mirrors", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("List available mirrors for this paper")),
				fields.New("save-to", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Save downloaded PDF to this file path")),
			fields.New("keep-tab-open", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Keep the tab open when finished")),
			fields.New("socket-path", fields.TypeString, fields.WithDefault(config.CurrentSocketPath()), fields.WithHelp("Host socket path")),
			fields.New("timeout-ms", fields.TypeInteger, fields.WithDefault(120000), fields.WithHelp("Socket request timeout in milliseconds")),
			fields.New("tab-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional tab id override")),
			fields.New("window-id", fields.TypeInteger, fields.WithDefault(int64(-1)), fields.WithHelp("Optional window id override")),
			fields.New("debug-socket", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Log socket request/response frames to stderr")),
		),
		cmds.WithSections(glazedSection, commandSection),
	)

	return &AnnasArchiveDownloadCommand{CommandDescription: desc}, nil
}

func buildAnnasArchiveMD5URL(doi string) string {
	doi = strings.TrimPrefix(doi, "/")
	doi = strings.TrimPrefix(doi, "https://doi.org/")
	doi = strings.TrimPrefix(doi, "doi.org/")
	return "https://annas-archive.gl/md5/" + url.PathEscape(doi)
}

// getMirrorURL returns the mirror URL for the given MD5, type, and optional index
func getMirrorURL(md5, mirrorType string, mirrorIndex int) string {
	downloadType := "slow"
	numServers := 8 // slow has 8 servers (0-7)

	if mirrorType == "fast" {
		downloadType = "fast"
		numServers = 12 // fast has 12 servers (0-11)
	}

	serverIndex := mirrorIndex
	if serverIndex < 0 {
		serverIndex = rand.Intn(numServers)
	}

	if serverIndex >= numServers {
		serverIndex = numServers - 1
	}

	return fmt.Sprintf("https://annas-archive.gl/%s_download/%s/0/%d", downloadType, md5, serverIndex)
}

// buildMetadataCode returns JS to extract paper metadata from MD5 page
func buildMetadataCode() string {
	return `
var result = {
  title: '',
  authors: '',
  doi: '',
  year: '',
  publisher: '',
  format: '',
  size: ''
};

// Get title from h1 or h2
var heading = document.querySelector('h1') || document.querySelector('h2');
if (heading) {
  result.title = heading.textContent.trim();
}

// Get DOI
var bodyText = document.body.innerText;
var doiMatch = bodyText.match(/(?:doi[:\s]*|https:\/\/doi\.org\/)([\d.]+\/[\w./%-]+)/i);
if (doiMatch) {
  result.doi = doiMatch[1];
}

// Get year
var yearMatch = bodyText.match(/\b(19|20)\d{2}\b/);
if (yearMatch) {
  result.year = yearMatch[0];
}

// Get format and size
var formatMatch = bodyText.match(/\b(PDF|EPUB|CBZ)\b/);
if (formatMatch) {
  result.format = formatMatch[0];
}
var sizeMatch = bodyText.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
if (sizeMatch) {
  result.size = sizeMatch[0];
}

return result;
`
}

// buildMirrorListCode returns JS to extract all available mirrors
func buildMirrorListCode() string {
	return `
var result = {
  md5: '',
  recordUrl: '',
  mirrors: { fast: [], slow: [] },
  metadata: {}
};

var md5Match = location.pathname.match(/\/md5\/([a-f0-9]{32})/);
if (md5Match) {
  result.md5 = md5Match[1];
}

if (location.pathname.startsWith('/scidb/')) {
  var links = document.querySelectorAll('a[href]');
  for (var i = 0; i < links.length; i++) {
    var href = links[i].getAttribute('href');
    if (href && href.match(/^\/md5\/[a-f0-9]{32}$/)) {
      result.recordUrl = href;
      var match = href.match(/\/md5\/([a-f0-9]{32})/);
      if (match) {
        result.md5 = match[1];
      }
      break;
    }
  }
}

var allLinks = document.querySelectorAll('a[href]');
for (var i = 0; i < allLinks.length; i++) {
  var link = allLinks[i];
  var href = link.getAttribute('href');
  var text = link.textContent.trim();

  if (href && href.includes('/fast_download/')) {
    var match = href.match(/\/fast_download\/[^\/]+\/\d+\/(\d+)/);
    if (match) {
      var idx = parseInt(match[1], 10);
      var exists = result.mirrors.fast.some(function(m) { return m.index === idx; });
      if (!exists) {
        result.mirrors.fast.push({
          type: 'fast',
          index: idx,
          url: href,
          label: text || 'Fast Partner Server #' + (idx + 1)
        });
      }
    }
  }

  if (href && href.includes('/slow_download/')) {
    var match = href.match(/\/slow_download\/[^\/]+\/\d+\/(\d+)/);
    if (match) {
      var idx = parseInt(match[1], 10);
      var exists = result.mirrors.slow.some(function(m) { return m.index === idx; });
      if (!exists) {
        result.mirrors.slow.push({
          type: 'slow',
          index: idx,
          url: href,
          label: text || 'Slow Partner Server #' + (idx + 1)
        });
      }
    }
  }
}

result.mirrors.fast.sort(function(a, b) { return a.index - b.index; });
result.mirrors.slow.sort(function(a, b) { return a.index - b.index; });

var doiLink = document.querySelector('a[href*="doi.org"]');
if (doiLink) {
  result.metadata = { doi: doiLink.getAttribute('href').replace('https://doi.org/', '') };
}

return result;
`
}

func (c *AnnasArchiveDownloadCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &AnnasArchiveDownloadSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchAnnasArchiveDownload(ctx, s)
	if err != nil {
		return err
	}

	for _, row := range annasArchiveDownloadDataToRows(data) {
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *AnnasArchiveDownloadCommand) RunIntoWriter(
	ctx context.Context,
	vals *values.Values,
	w io.Writer,
) error {
	s := &AnnasArchiveDownloadSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	data, err := fetchAnnasArchiveDownload(ctx, s)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, renderAnnasArchiveDownloadMarkdown(data))
	return err
}

func fetchAnnasArchiveDownload(ctx context.Context, s *AnnasArchiveDownloadSettings) (data *annasArchiveDownloadData, retErr error) {
	if strings.TrimSpace(s.DOI) == "" {
		return nil, fmt.Errorf("--doi is required")
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

	// Navigate to SciDB page for DOI
	scidbURL := "https://annas-archive.gl/scidb/" + url.PathEscape(s.DOI) + "/"

	if tabID == nil && windowID == nil {
		resolvedTabID, err := openOwnedTab(ctx, client, scidbURL, tabReadyOptions{
			URLPrefix: "https://annas-archive.gl/scidb/",
		})
		if err != nil {
			return nil, err
		}
		tabID = &resolvedTabID
		ownedTabID = &resolvedTabID
	} else {
		if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": scidbURL}, tabID, windowID); err != nil {
			return nil, err
		}
		if tabID != nil {
			if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: "https://annas-archive.gl/scidb/"}); err != nil {
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

	time.Sleep(2 * time.Second)

	// Get mirrors from SciDB page first
	mirrorResp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildMirrorListCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}

	mirrorData := parseMirrorResponse(mirrorResp)

	// Navigate to MD5 page for metadata and mirrors
	md5PageURL := "https://annas-archive.gl/md5/" + mirrorData.MD5
	if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": md5PageURL}, tabID, windowID); err != nil {
		return nil, err
	}
	if tabID != nil {
		if err := waitForTabReady(ctx, client, *tabID, tabReadyOptions{URLPrefix: md5PageURL}); err != nil {
			return nil, err
		}
	}
	time.Sleep(2 * time.Second)

	// Extract metadata from MD5 page
	metadataResp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": buildMetadataCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	metadataData := parseMetadataResponse(metadataResp)

	// Re-fetch mirrors from MD5 page (for --list-mirrors)
	mirrorResp, err = ExecuteTool(ctx, client, "js", map[string]any{"code": buildMirrorListCode()}, tabID, windowID)
	if err != nil {
		return nil, err
	}
	mirrorData = parseMirrorResponse(mirrorResp)

	// Build metadata map
	metadata := map[string]any{
		"doi":     metadataData.DOI,
		"title":   metadataData.Title,
		"year":    metadataData.Year,
		"format":  metadataData.Format,
		"size":    metadataData.Size,
		"authors": metadataData.Authors,
	}

	// List mirrors mode
	if s.ListMirrors || s.MirrorType == "list" {
		return &annasArchiveDownloadData{
			MD5:      mirrorData.MD5,
			Metadata: metadata,
			Mirrors:  mirrorData.AllMirrors(),
		}, nil
	}

	// Download mode
	mirrorType := s.MirrorType
	if mirrorType == "" || mirrorType == "list" {
		mirrorType = "slow"
	}

	// Fail fast for fast mirrors (not supported yet)
	if mirrorType == "fast" {
		return nil, fmt.Errorf("--mirror fast is not supported: fast mirrors require Anna's Archive membership. Use --mirror slow (default) or --list-mirrors to see options")
	}

	mirrorURL := getMirrorURL(mirrorData.MD5, mirrorType, s.MirrorIndex)

	// Navigate to mirror
	if _, err := ExecuteTool(ctx, client, "navigate", map[string]any{"url": mirrorURL}, tabID, windowID); err != nil {
		return nil, err
	}

	// Wait for download link
	deadline := time.Now().Add(120 * time.Second)
	var downloadURL string

	for time.Now().Before(deadline) {
		resp, err := ExecuteTool(ctx, client, "js", map[string]any{
			"code": `var r = {
				url: location.href,
				title: document.title,
				downloadUrl: '',
				waiting: false,
				waitMessage: '',
				pageState: 'unknown'
			};
			
			// Determine page state
			var body = document.body.innerText.toLowerCase();
			
			// Check for wait message
			if (body.includes('please wait') || body.includes('generating') || body.includes('preparing')) {
				r.waiting = true;
				r.pageState = 'waiting';
				var match = body.match(/(\d+)\s*(seconds?|minutes?)/i);
				if (match) {
					r.waitMessage = 'Please wait ~' + match[0];
				}
			} else if (location.pathname.includes('/fast_download') || location.pathname.includes('/slow_download')) {
				r.pageState = 'download_page';
			}
			
			// Look for download link
			var links = document.querySelectorAll('a[href]');
			for (var i = 0; i < links.length; i++) {
				var href = links[i].getAttribute('href');
				var text = links[i].textContent.trim();
				
				if (href && href.startsWith('/')) continue;
				
				if (href && (href.includes('.pdf') || href.includes('/d3/'))) {
					r.downloadUrl = href;
					r.waiting = false;
					r.pageState = 'found';
					break;
				}
				
				if (href && text.toLowerCase().includes('download') && !href.startsWith('/')) {
					r.downloadUrl = href;
					r.waiting = false;
					r.pageState = 'found';
					break;
				}
			}
			
			return r;`,
		}, tabID, windowID)
		if err != nil {
			return nil, err
		}

		parsed := parseResult(resp)
		if dataMap, ok := parsed.Data.(map[string]any); ok {
			if url, ok := dataMap["downloadUrl"].(string); ok && url != "" {
				downloadURL = url
				break
			}

			// Log state for debugging
			if state, ok := dataMap["pageState"].(string); ok && state == "waiting" {
				msg := ""
				if m, ok := dataMap["waitMessage"].(string); ok {
					msg = m
				}
				fmt.Fprintf(os.Stderr, "DEBUG: Waiting for download... %s\n", msg)
			}
		}

		time.Sleep(3 * time.Second)
	}

if s.Output != "" {
		if err := downloadFile(ctx, s, downloadURL, metadataData.Title); err != nil {
			return nil, err
		}
	}

	// Ensure output path ends with .pdf
	if s.Output != "" && !strings.HasSuffix(s.Output, ".pdf") {
		s.Output = s.Output + ".pdf"
	}

	return &annasArchiveDownloadData{
		Raw: map[string]any{
			"md5":           mirrorData.MD5,
			"metadata":       metadata,
			"download_url":  downloadURL,
			"mirror_used":    mirrorURL,
		},
		MD5:         mirrorData.MD5,
		Title:       metadataData.Title,
		Metadata:    metadata,
		DownloadURL: downloadURL,
	}, nil
}

type mirrorListData struct {
	MD5       string
	RecordURL string
	Metadata  map[string]any
	Fast      []MirrorInfo
	Slow      []MirrorInfo
}

func (m *mirrorListData) AllMirrors() []MirrorInfo {
	all := make([]MirrorInfo, 0, len(m.Fast)+len(m.Slow))
	all = append(all, m.Fast...)
	all = append(all, m.Slow...)
	return all
}

func parseMirrorResponse(resp map[string]any) *mirrorListData {
	if e := extractErrorText(resp); e != "" {
		return &mirrorListData{}
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		return &mirrorListData{}
	}

	result := &mirrorListData{
		Metadata: make(map[string]any),
	}

	result.MD5, _ = dataMap["md5"].(string)
	result.RecordURL, _ = dataMap["recordUrl"].(string)

	if meta, ok := dataMap["metadata"].(map[string]any); ok {
		result.Metadata = meta
	}

	if mirrors, ok := dataMap["mirrors"].(map[string]any); ok {
		if fast, ok := mirrors["fast"].([]any); ok {
			for _, f := range fast {
				if m, ok := f.(map[string]any); ok {
					result.Fast = append(result.Fast, MirrorInfo{
						Type:    "fast",
						Index:   int(m["index"].(float64)),
						URL:     m["url"].(string),
						Label:   m["label"].(string),
					})
				}
			}
		}
		if slow, ok := mirrors["slow"].([]any); ok {
			for _, s := range slow {
				if m, ok := s.(map[string]any); ok {
					result.Slow = append(result.Slow, MirrorInfo{
						Type:    "slow",
						Index:   int(m["index"].(float64)),
						URL:     m["url"].(string),
						Label:   m["label"].(string),
					})
				}
			}
		}
	}

	return result
}

type metadataResult struct {
	Title     string
	Authors   string
	DOI       string
	Year      string
	Publisher string
	Format    string
	Size      string
}

func parseMetadataResponse(resp map[string]any) *metadataResult {
	if e := extractErrorText(resp); e != "" {
		return &metadataResult{}
	}

	parsed := parseResult(resp)
	dataMap, ok := parsed.Data.(map[string]any)
	if !ok {
		return &metadataResult{}
	}

	result := &metadataResult{}
	result.Title, _ = dataMap["title"].(string)
	result.Authors, _ = dataMap["authors"].(string)
	result.DOI, _ = dataMap["doi"].(string)
	result.Year, _ = dataMap["year"].(string)
	result.Publisher, _ = dataMap["publisher"].(string)
	result.Format, _ = dataMap["format"].(string)
	result.Size, _ = dataMap["size"].(string)

	return result
}

func annasArchiveDownloadDataToRows(data *annasArchiveDownloadData) []types.Row {
	rows := []types.Row{}

	if data.Mirrors != nil && len(data.Mirrors) > 0 {
		for _, m := range data.Mirrors {
			row := map[string]any{
				"type":  m.Type,
				"index": m.Index,
				"label": m.Label,
				"url":   "https://annas-archive.gl" + m.URL,
			}
			rows = append(rows, types.NewRowFromMap(row))
		}
		return rows
	}

	row := map[string]any{
		"md5": data.MD5,
	}
	if data.Metadata != nil {
		for k, v := range data.Metadata {
			row[k] = v
		}
	}
	row["download_url"] = data.DownloadURL
	rows = append(rows, types.NewRowFromMap(row))

	return rows
}

func renderAnnasArchiveDownloadMarkdown(data *annasArchiveDownloadData) string {
	var b strings.Builder

	if data.Mirrors != nil && len(data.Mirrors) > 0 {
		b.WriteString("# Anna's Archive Mirrors\n\n")
		if data.MD5 != "" {
			b.WriteString(fmt.Sprintf("- **MD5:** `%s`\n", data.MD5))
		}
		if doi, ok := data.Metadata["doi"].(string); ok && doi != "" {
			b.WriteString(fmt.Sprintf("- **DOI:** `%s`\n", doi))
		}

		b.WriteString("\n## Fast Mirrors (12 available)\n\n")
		for _, m := range data.Mirrors {
			if m.Type == "fast" {
				b.WriteString(fmt.Sprintf("- %d. [%s](https://annas-archive.gl%s)\n", m.Index+1, m.Label, m.URL))
			}
		}

		b.WriteString("\n## Slow Mirrors (8 available)\n\n")
		for _, m := range data.Mirrors {
			if m.Type == "slow" {
				b.WriteString(fmt.Sprintf("- %d. [%s](https://annas-archive.gl%s)\n", m.Index+1, m.Label, m.URL))
			}
		}

		b.WriteString("\n---\n*Use `--mirror slow --mirror-index N` to select.*\n")
		return b.String()
	}

	b.WriteString("# Anna's Archive Paper Download\n\n")

	if title, ok := data.Metadata["title"].(string); ok && title != "" {
		b.WriteString(fmt.Sprintf("## %s\n\n", title))
	}

	if doi, ok := data.Metadata["doi"].(string); ok && doi != "" {
		b.WriteString(fmt.Sprintf("- **DOI:** `%s`\n", doi))
	}

	if data.MD5 != "" {
		b.WriteString(fmt.Sprintf("- **MD5:** `%s`\n", data.MD5))
	}

	b.WriteString("\n## Download\n\n")
	if data.DownloadURL != "" {
		b.WriteString(fmt.Sprintf("[Download PDF](%s)\n\n", data.DownloadURL))
		if year, ok := data.Metadata["year"].(string); ok && year != "" {
			b.WriteString(fmt.Sprintf("- **Year:** %s\n", year))
		}
		if format, ok := data.Metadata["format"].(string); ok && format != "" {
			b.WriteString(fmt.Sprintf("- **Format:** %s\n", format))
		}
		if size, ok := data.Metadata["size"].(string); ok && size != "" {
			b.WriteString(fmt.Sprintf("- **Size:** %s\n", size))
		}
		b.WriteString(fmt.Sprintf("\n```\n%s\n```\n", data.DownloadURL))
	} else {
		b.WriteString("*Download link not available (try again or use --list-mirrors to see options)*\n")
	}

	return b.String()
}

// downloadFile downloads a file from URL to output path
func downloadFile(ctx context.Context, s *AnnasArchiveDownloadSettings, url, title string) error {
	fmt.Fprintf(os.Stderr, "Downloading: %s\n", url)

	httpClient := &http.Client{Timeout: 300 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Determine output filename
	outputPath := s.Output
	if outputPath == "" {
		if title != "" {
			filename := sanitizeFilename(title) + ".pdf"
			outputPath = filename
		} else {
			outputPath = "download.pdf"
		}
	}

	// Ensure output path ends with .pdf
	if !strings.HasSuffix(outputPath, ".pdf") {
		outputPath = outputPath + ".pdf"
	}

	// Create output directory if needed
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Copy response body to file
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Saved to: %s\n", outputPath)
	return nil
}

// sanitizeFilename makes a title safe for use as filename
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.TrimSpace(name)
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
