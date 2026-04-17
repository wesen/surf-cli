# Changelog

## 2026-04-17

### Added
- `surf-go annas-archive` parent command
- `surf-go annas-archive search` subcommand - search papers by query
- `surf-go annas-archive download` subcommand - download with mirror selection
  - `--doi DOI` - paper DOI (required)
  - `--list-mirrors` - show all available mirrors
  - `--mirror slow` - select mirror type (default: slow)
  - `--mirror-index N` - select specific mirror (0-based, slow only)
  - `--save-to FILE` - download PDF directly to file
  - `--keep-tab-open` - keep tab open after download
  - Handles wait cycles for slow download initiation (120s timeout)

### Changed
- Fast mirrors now rejected with clear error (membership required)
- Added paper metadata to output: title, year, format, size

### Technical details
- Slow mirrors: 8 servers (0-7), no membership required
- Default: random slow mirror
- Download URL: external (e.g., `momot.rs/d3/...`)
- Metadata extracted from MD5 page (title, DOI, year, format, size)

## 2026-04-16

- Initial workspace created

