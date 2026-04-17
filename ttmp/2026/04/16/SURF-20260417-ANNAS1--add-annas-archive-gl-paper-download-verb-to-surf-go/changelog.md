# Changelog

## 2026-04-17

### Added
- `surf-go annas-archive` parent command
- `surf-go annas-archive search` subcommand - search papers by query
- `surf-go annas-archive download` subcommand - download with mirror selection
  - `--doi DOI` - paper DOI (required)
  - `--list-mirrors` - show all available mirrors
  - `--mirror fast|slow` - select mirror type (default: slow)
  - `--mirror-index N` - select specific mirror (0-based)
  - `--keep-tab-open` - keep tab open after download
  - Handles wait cycles for slow download initiation (120s timeout)

### Technical details
- Fast mirrors: 12 servers (0-11)
- Slow mirrors: 8 servers (0-7)
- Default: random slow mirror
- Download URL: external (e.g., `wbsg8v.xyz/d3/...`)

## 2026-04-16

- Initial workspace created

