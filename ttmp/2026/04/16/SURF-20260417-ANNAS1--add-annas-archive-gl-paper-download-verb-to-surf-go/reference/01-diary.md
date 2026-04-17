---
Title: ""
Ticket: ""
Status: ""
Topics:
    - surf-go
    - browser-automation
    - annas-archive
    - cli
    - glazed
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: Command registration
    - Path: go/internal/cli/commands/annas_archive.go
      Note: Implementation diary
    - Path: go/internal/cli/commands/scripts/annas_archive.js
      Note: Research and production JS scripts
    - Path: ttmp/2026/04/16/SURF-20260417-ANNAS1--add-annas-archive-gl-paper-download-verb-to-surf-go/scripts/01-page-shape-probe.js
      Note: Browser probe script
    - Path: ttmp/2026/04/16/SURF-20260417-ANNAS1--add-annas-archive-gl-paper-download-verb-to-surf-go/scripts/02-search-result-extractor.js
      Note: Browser probe script
    - Path: ttmp/2026/04/16/SURF-20260417-ANNAS1--add-annas-archive-gl-paper-download-verb-to-surf-go/scripts/03-paper-detail-extractor.js
      Note: Browser probe script
    - Path: ttmp/2026/04/16/SURF-20260417-ANNAS1--add-annas-archive-gl-paper-download-verb-to-surf-go/scripts/04-download-url-extractor.js
      Note: Browser probe script
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---








# Diary

## Goal

Add annas-archive.gl paper download verb to surf-go following the browser-side verb tutorial pattern.

## Step 1: Initial Research - Understanding Anna's Archive Site Structure

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to add annas archive support to surf-go. Read @go/pkg/doc/tutorials/01-building-browser-side-verbs.md and then create a verb to download a paper/scientific article from annas-archive.gl . Keep a detailed diary as you work. Save every single script you create in the ticket's scripts/ folder, with an incrementing numerical prefix."

**Assistant interpretation:** Build a new surf-go verb that downloads scientific papers/articles from annas-archive.gl using the browser automation pattern.

**Inferred user intent:** Add paper download capability to surf-go via Anna's Archive, following the existing tutorial patterns.

### What I did
- Read the browser-side verb tutorial (`01-building-browser-side-verbs.md`)
- Created docmgr ticket SURF-20260417-ANNAS1
- Explored annas-archive.gl homepage, search page, and paper detail page
- Identified key URL patterns and page structures

### What I learned

**Anna's Archive Site Structure:**

1. **Main search URL:** `https://annas-archive.gl/search`
2. **Journal articles search:** `https://annas-archive.gl/search?index=journals`
3. **Paper detail pages:** `https://annas-archive.gl/md5/{md5hash}`
4. **SciDB (papers):** `https://annas-archive.gl/scidb?doi={doi}`

**Key Page Elements:**

From search page:
- Search input: `input[placeholder*="Title, author, DOI"]`
- Result containers: `div.generic` with links to `/md5/{hash}`
- Result metadata: language, format (PDF), size, source

From paper detail page:
- Paper title and authors
- DOI: `doi:10.1016/...` format in metadata
- MD5 visible in URL
- Download options section with:
  - Fast downloads (Anna's Archive SciDB, Fast Partner Servers)
  - Slow downloads (Slow Partner Servers)
- Download URLs: `/fast_download/{md5}/{fileIndex}/{serverIndex}`
- SciDB URL: `/scidb?doi={doi}`

### What was tricky to build
- Understanding the download flow (multiple server options)
- Identifying how to look up papers by DOI vs title vs search

### What should be done in the future
- Create browser probe scripts to validate selectors
- Test DOI lookup flow
- Test download button behavior

### Code review instructions
N/A - still in research phase

### Technical details

**URL Patterns:**
- Search: `https://annas-archive.gl/search?q={query}&index=journals`
- Paper by MD5: `https://annas-archive.gl/md5/{md5}`
- Fast download: `https://annas-archive.gl/fast_download/{md5}/0/{serverIndex}`
- Slow download: `https://annas-archive.gl/slow_download/{md5}/0/{serverIndex}`
- SciDB by DOI: `https://annas-archive.gl/scidb?doi={doi}`

**Verified Selectors:**
- Search input: `input[placeholder*="Title"], input[placeholder*="DOI"]`
- Result count: `main` text matching `/Results \d+-\d+ \((\d+) total\)/`
- Result MD5 links: `main a[href^="/md5/"]` (dedupe by href)
- Paper title: `main h1` or `main h2`
- DOI from page: text matching `/doi[:\s]*([\d.]+\/[\w./%-]+)/i`
- Format/size: text matching `/\b(PDF|EPUB)\b/` and `/(\d+\.?\d*) (MB|GB|KB)/i`
- Download links: all `a[href*="/fast_download/"]`, `a[href*="/slow_download/"]`, `a[href*="/scidb"]`

**Download URL Structure:**
- Fast: `/fast_download/{md5}/0/{serverIndex}` - servers 0-10 (servers 0-5 recommended)
- Slow: `/slow_download/{md5}/0/{serverIndex}` - servers 0-7
- SciDB: `/scidb?doi={doi}` - direct to SciDB

**Validated Flow:**
1. Navigate to search URL with DOI query + `index=journals`
2. Wait for results, extract MD5 from first `/md5/` link
3. Navigate to paper page `/md5/{md5}`
4. Extract metadata (title, DOI, format, size) and download URLs
5. Return structured result with download links

---

## Step 2: Script Creation and Browser Validation

### What I did
- Created script 01: `scripts/01-page-shape-probe.js` - initial page exploration
- Created script 02: `scripts/02-search-result-extractor.js` - search result extraction
- Created script 03: `scripts/03-paper-detail-extractor.js` - paper page extraction
- Validated with real DOI search: `10.1038/nature12373`

### What I learned
- DOI search works: `https://annas-archive.gl/search?q=10.1038/nature12373&index=journals`
- First result: MD5 `d89c394b00116f093b5d9d6a6611f975`
- Paper metadata extraction works
- Multiple download options available (fast, slow, SciDB)

### Scripts created
1. `scripts/01-page-shape-probe.js` - Page structure exploration
2. `scripts/02-search-result-extractor.js` - Search result extraction
3. `scripts/03-paper-detail-extractor.js` - Paper detail extraction
4. `scripts/04-download-url-extractor.js` - Download URL extraction from SciDB

---

## Step 3: Go Command Implementation

### What I did
- Created `go/internal/cli/commands/annas_archive.go` - the main command
- Updated `go/internal/cli/commands/scripts/annas_archive.js` - embedded JS for extraction

### Key findings about download flow
1. Fast downloads require membership - redirects to `/fast_download_not_member`
2. Slow downloads work without membership
3. SciDB page (`/scidb/{doi}`) has a "Download" link to external URL
4. Download URL format: `https://b4mcx2ml.net/d3/x/.../paper.pdf`

### Command design
- **Input:** `--doi` for direct DOI lookup, or `--query` for search
- **Output:** Paper metadata + download URLs (Markdown by default, Glaze rows with `--with-glaze-output`)
- **Tab behavior:** Creates new tab if not specified, closes by default unless `--keep-tab-open`

### Files created
- `go/internal/cli/commands/annas_archive.go` (13KB)
- `go/internal/cli/commands/scripts/annas_archive.js` (updated)

### Completed
- Command wired into `main.go`
- Help system integration works
- Verified with `surf-go help annas-archive`

### Next steps
- Fix search mode extraction (minimal script only supports DOI mode)
- Add unit tests
- Add embedded help documentation

## Testing Results

### Working
- DOI lookup mode: `surf-go annas-archive --doi 10.1038/nature12373`
  - Navigates to SciDB page
  - Extracts DOI
  - Finds download URL
  - Returns Markdown with results

### Not working yet
- Search mode: `surf-go annas-archive --query "title"`
  - Navigates to search page but extraction not implemented in minimal script

### Key discovery about js tool
The `js` tool requires explicit `return` statement to return values:
```js
var result = {test: true};
result;  // doesn't work
return result;  // works
```

Also, IIFEs don't work:
```js
(function() { return {test: true}; })();  // returns undefined
```

---

## Step 4: Refinement - Separate Commands and Mirror Selection

### What I did
- Split single command into subcommand structure:
  - `surf-go annas-archive search` - Search for papers
  - `surf-go annas-archive download` - Get download with mirror selection
- Implemented `--list-mirrors` and `--mirror fast|slow --mirror-index N` flags
- Fixed download polling loop to handle wait cycles properly

### Working features
```bash
# List available mirrors
surf-go annas-archive download --doi 10.1038/nature12373 --list-mirrors

# Download with specific mirror
surf-go annas-archive download --doi 10.1038/nature12373 --mirror slow --mirror-index 0
```

### Mirror types
- **Fast mirrors:** 12 servers (0-11), servers 0-5 marked "recommended"
- **Slow mirrors:** 8 servers (0-7), no membership required
- Default: random slow mirror

### Download wait handling
The download page shows "Please wait" before the PDF link appears. JS code:
1. Checks page state (waiting vs download_page vs found)
2. Polls every 3 seconds
3. Up to 120 second timeout
4. Returns download URL when found

### Key fix: removing emoji check
Emoji literal (`📚`) caused Go string parsing issues. Removed:
```js
// Removed: text.includes('📚')
// Keep: text.toLowerCase().includes('download')
```

### Files modified/created
- `go/internal/cli/commands/annas_archive.go` - Parent command (subcommand structure)
- `go/internal/cli/commands/annas_archive_search.go` - Search subcommand
- `go/internal/cli/commands/annas_archive_download.go` - Download subcommand (refactored)

### Commit
```
feat: add annas-archive command with search and download subcommands
```

### Remaining work
- Add paper title/metadata to download output
- Add actual PDF download option (fetch URL and save to file)
- Add `--output FILE` flag for direct file save

---

## Step 5: Metadata, Output File, and Fast Mirror Rejection

### What I did
- Added paper title, year, format, size extraction from MD5 page
- Added `--save-to` flag for direct PDF download to file
- Rejected fast mirrors with clear error message

### New features
```bash
# Download with metadata
surf-go annas-archive download --doi 10.1038/nature12373
# Output includes: Title, Year (2013), Format (PDF), Size (0.9MB)

# Download to file directly
surf-go annas-archive download --doi 10.1038/nature12373 --save-to ~/papers/thermometry.pdf

# Fast mirrors now fail with clear message
surf-go annas-archive download --doi 10.1038/nature12373 --mirror fast
# Error: --mirror fast is not supported: fast mirrors require Anna's Archive membership
```

### Key fix: --output vs --save-to
Glazed already uses `--output` for output format. Used `--save-to` for PDF file path instead.

### Commit
```
feat: add metadata extraction, --save-to flag, and fail fast for fast mirrors
```

### Done
- [x] Add paper title/metadata to download output
- [x] Add --save-to FILE flag for direct file save
- [x] Fail fast for fast mirrors with clear message
