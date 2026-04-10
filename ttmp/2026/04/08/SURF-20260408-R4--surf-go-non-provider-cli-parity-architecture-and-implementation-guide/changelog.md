# Changelog

## 2026-04-10

- Added an embedded Glazed help tutorial under `go/pkg/doc/tutorials/01-building-browser-side-verbs.md` and wired `surf-go` to load embedded docs into the Glazed help system.
- Implemented `surf-go kagi-search` as a dual-mode browser-side verb with an embedded Kagi scraper, mock-host coverage, and Markdown/Glazed row output.
- Added a Kagi research diary and ticket-side probe script documenting the selectors and extraction strategy used for the first implementation.

## 2026-04-08

- Initial workspace created.
- Added detailed architecture/design/implementation guide for non-provider surf-go parity.
- Added investigation diary documenting ticket setup and evidence collection.
- Added phased task backlog for parity implementation.
- Generated machine-readable Node-vs-Go non-provider command-gap inventory.
- Tightened the design doc so every new public `surf-go` verb in this effort is explicitly required to be a Glazed command.
- Replaced the coarse phase backlog with a detailed Glazed command implementation checklist.
