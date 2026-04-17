---
title: Tasks
slug: annas-archive-paper-download-tasks
Topics:
- surf-go
- browser-automation
- annas-archive
- cli
- glazed
---

# Tasks

## Research Phase

- [x] 1. Explore annas-archive.gl site structure and understand the paper download flow
- [x] 2. Create browser probes to identify page selectors for search results and paper downloads
- [x] 3. Document annas-archive.gl site quirks and anti-bot measures

## Design Phase

- [x] 4. Write the user contract for the annas-archive paper download verb
- [x] 5. Decide tab strategy (create new vs reuse vs current tab)
- [x] 6. Define browser-state ownership and cleanup rules

## Implementation Phase

- [x] 7. Create the Go command file (annas_archive.go)
- [x] 8. Create the production JavaScript extractor script
- [x] 9. Implement dual-mode output (Markdown + Glaze rows)
- [x] 10. Wire the command into main.go

## Testing Phase

- [x] 11. Real-browser validation pass (DOI mode works)
- [x] 12. Fix search mode extraction

## Documentation Phase

- [ ] 13. Add embedded help documentation
- [ ] 14. Update diary with final findings

## Completed Scripts

- `scripts/01-page-shape-probe.js` - Page structure exploration
- `scripts/02-search-result-extractor.js` - Search result extraction  
- `scripts/03-paper-detail-extractor.js` - Paper detail extraction
- `scripts/04-download-url-extractor.js` - Download URL extraction from SciDB
