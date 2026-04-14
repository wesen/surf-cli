---
Title: Make --file Flag a StringList for surf chatgpt ask
Ticket: SURF-20260414-R7
Status: completed
Topics:
  - surf-go
  - glazed
  - chatgpt
  - file-upload
  - cli
DocType: ticket
Intent: implementation
Owners: []
RelatedFiles:
  - Path: go/internal/cli/commands/chatgpt.go
    Note: CLI command with --file flag
  - Path: go/internal/host/providers/chatgpt.go
    Note: Provider that calls uploadFiles
  - Path: go/internal/host/router/toolmap.go
    Note: Tool routing
ExternalSources: []
Summary: Changed the --file flag in surf chatgpt ask from a single string to a string list so users can upload multiple files using --file file1.txt --file file2.txt syntax.
LastUpdated: 2026-04-14T15:30:00-04:00
WhatFor: Currently --file only accepts one file. ChatGPT supports multiple file uploads. This change makes --file accept multiple files while maintaining backward compatibility.
WhenToUse: Use when a user wants to attach multiple files to a ChatGPT prompt.
---

# Ticket: SURF-20260414-R7

## Status: COMPLETED

## Commits

### Commit 1: Implementation (8487a00)
**Files changed**: `chatgpt.go`, `chatgpt.go`, `chatgpt_test.go`

Changes:
- **CLI (`chatgpt.go`)**: `File string` → `Files []string`, `TypeString` → `TypeStringList`
- **Provider (`chatgpt.go`)**: `File string` → `Files []string`, `uploadFiles(string)` → `uploadFiles([]string)`
- **Tests**: Existing test passes (legacy `file` key). New `TestHandleChatGPTToolWithMultipleFiles` verifies new `files` key.
- New `fileListFromArgs` helper with backward compat for legacy `"file"` key
- New `toStringArray` helper (duplicated from router/toolmap.go)

### Commit 2: Ticket docs (b20cb82)
**Files added**: `README.md`, `design-doc/01-implementation-guide.md`

## Verification

```bash
# Help text
$ surf-go chatgpt ask --help
     --file    File(s) to attach before sending the prompt.
               Specify multiple files with repeated --file flags
               (e.g., --file a.txt --file b.txt).
               Backward compatible with single file paths. - <stringList>

# All tests pass
$ go test ./...
ok  github.com/nicobailon/surf-cli/gohost/internal/host/providers  1.205s
```

## Usage

```bash
# Single file (backward compatible)
surf chatgpt ask --file /tmp/report.txt "summarize this"

# Multiple files (new)
surf chatgpt ask --file /tmp/a.txt --file /tmp/b.txt --file /tmp/c.md "review these files"
```

## Backward Compatibility

- `--file a.txt,b.txt` (comma-separated) still works via `toStringArray` legacy path
- `--file a.txt` (single) still works via Glazed coercion (`[]string{"a.txt"}`)
- Provider's `fileListFromArgs` checks `"files"` first (new), falls back to `"file"` (legacy)
