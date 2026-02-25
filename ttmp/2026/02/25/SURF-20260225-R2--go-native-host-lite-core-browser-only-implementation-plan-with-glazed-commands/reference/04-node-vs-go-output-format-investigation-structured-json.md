---
Title: Node vs Go Output Format Investigation (Structured JSON)
Ticket: SURF-20260225-R2
Status: active
Topics:
    - go
    - native-messaging
    - chromium
    - architecture
    - glazed
    - migration
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/internal/cli/commands/format.go
      Note: |-
        New response parser and structured row shaping logic
        Structured response parsing and row schema
    - Path: go/internal/cli/commands/format_test.go
      Note: |-
        Unit tests validating structured text parsing
        Structured parsing unit tests
    - Path: go/internal/cli/commands/navigate.go
      Note: |-
        Passes tool name to shared formatter
        Formatter tool-name callsite update
    - Path: go/internal/cli/commands/stream_simple.go
      Note: |-
        Emits stream metadata fields in output rows
        Stream row metadata expansion
    - Path: go/internal/cli/commands/tool_raw.go
      Note: |-
        Passes tool name to shared formatter
        Formatter tool-name callsite update
    - Path: go/internal/cli/commands/tool_simple.go
      Note: |-
        Passes tool name to shared formatter
        Formatter tool-name callsite update
    - Path: scripts/compare-go-node-output.cjs
      Note: |-
        Live comparison harness for Node vs Go output shape and payloads
        Live node-vs-go comparison harness
    - Path: scripts/diff-go-node-summary.cjs
      Note: |-
        Shape-diff helper across two comparison runs
        Run-to-run shape diff helper
ExternalSources: []
Summary: Evidence-driven comparison of Node CLI and Go CLI output formats, with formatter changes that add structured data extraction and richer JSON rows.
LastUpdated: 2026-02-25T18:56:00-05:00
WhatFor: Track concrete output-parity progress and remaining schema gaps during Go CLI migration
WhenToUse: Use when validating JSON output quality/parity or debugging formatter regressions
---


# Node vs Go Output Format Investigation (Structured JSON)

## Goal

Validate and improve Go CLI JSON output usefulness by parsing tool-response payloads into structured fields instead of only returning `status/message/error/response`.

## Context

The active migration issue was that `surf-go --output json` emitted rows like:

```json
[{"status":"ok","message":"...","error":"","response":"{...}"}]
```

That shape forced downstream consumers to re-parse stringified JSON from `message`/`response` and made command-specific automation brittle.

This investigation compared live Node and Go CLI outputs against a running Snap Chromium extension socket:

- Socket path: `/home/manuel/snap/chromium/common/surf-cli/surf.sock`
- Baseline run: `2026-02-25T23-45-57-027Z`
- Post-change run: `2026-02-25T23-48-23-122Z`

## Quick Reference

### Reproduction commands

```bash
SURF_SOCKET_PATH="$HOME/snap/chromium/common/surf-cli/surf.sock" \
  node scripts/compare-go-node-output.cjs

node scripts/diff-go-node-summary.cjs \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/output-compare/2026-02-25T23-45-57-027Z \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/output-compare/2026-02-25T23-48-23-122Z \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/output-compare/2026-02-25T23-48-23-122Z/SHAPE-DIFF-vs-2026-02-25T23-45-57-027Z.md

cd go && go test ./...
```

### New Go row schema

Each non-stream tool command now returns one row with these key fields:

- `tool`: requested tool name (for example `tab.list`, `page.state`)
- `status`: `ok` or `error`
- `id`: native response id
- `error`: extracted error message when present
- `text`: joined text blocks from `result.content`
- `data_kind`: `object`, `array`, or `none`
- `data_count`: element/property count for structured payloads
- `data`: parsed JSON object/array when `text` is valid JSON
- `content`: original `result.content`
- `result`: original `result` object

### Before vs after shape summary

| Case | Go shape before | Go shape after |
|---|---|---|
| `tab-list` | `array(1)<object{error,message,response,status}>` | `array(1)<object{content,data,data_count,data_kind,error,id,result,status}>` |
| `page-state` | `array(1)<object{error,message,response,status}>` | `array(1)<object{content,data,data_count,data_kind,error,id,result,status}>` |
| `page-read` | `array(1)<object{error,message,response,status}>` | `array(1)<object{content,data,data_count,data_kind,error,id,result,status}>` |
| `network-list` | `array(1)<object{error,message,response,status}>` | `array(1)<object{content,data,data_count,data_kind,error,id,result,status}>` |

Full table: `sources/output-compare/2026-02-25T23-48-23-122Z/SHAPE-DIFF-vs-2026-02-25T23-45-57-027Z.md`

### Concrete parsing outcomes

- `tab.list`
  - `data_kind=array`
  - `data_count=3`
  - `data` contains parsed tab objects.
- `page.state`
  - `data_kind=object`
  - `data_count=8`
  - `data` contains parsed state object (`title`, `url`, `focusedElement`, modal/dropdown flags).
- `page.read`, `page.text`, `network`, `console`, `navigate`
  - `data_kind=none`
  - textual responses still preserved in `text` and `content`.

### Stream row improvement

Stream commands now emit:

- `stream_type`
- `event_type`
- `event` (raw event payload)

This avoids requiring consumers to inspect only one untyped `event` field.

### Remaining parity gap

Node CLI often returns raw JSON objects/arrays/strings directly, while `surf-go` still returns array-of-row objects (Glazed row model). The new Go rows are now structured and machine-parseable, but not identical to Node top-level shapes.

## Usage Examples

### Extract tab ids from Go output directly

```bash
cd go && go run ./cmd/surf-go tab list --output json \
  | jq '.[0].data[] | {id, title, active}'
```

### Read page state title/url without re-parsing escaped strings

```bash
cd go && go run ./cmd/surf-go page state --output json \
  | jq '.[0].data | {title, url, hasModal, hasDropdown}'
```

### Inspect textual responses when no structured JSON is available

```bash
cd go && go run ./cmd/surf-go page read --output json \
  | jq '.[0] | {tool, status, data_kind, text}'
```

## Related

- `reference/01-implementation-diary.md`
- `reference/03-manual-browser-validation-checklist.md`
- `sources/output-compare/2026-02-25T23-45-57-027Z`
- `sources/output-compare/2026-02-25T23-48-23-122Z`
