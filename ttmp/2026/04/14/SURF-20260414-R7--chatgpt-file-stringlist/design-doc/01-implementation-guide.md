---
Title: Implementation Guide - ChatGPT File StringList
Ticket: SURF-20260414-R7
---

# Implementation Guide: --file StringList for surf chatgpt ask

## Overview

Changed the `--file` flag from `fields.TypeString` to `fields.TypeStringList`, allowing users to specify multiple files via repeated flags: `--file file1.txt --file file2.txt`.

## Final Implementation

### CLI: `go/internal/cli/commands/chatgpt.go`

**ChatGPTSettings struct** — single change:
```go
// Before
File string `glazed:"file"`

// After
Files []string `glazed:"file"`  // Note: Glazed tag unchanged so callers use same flag name
```

**Flag definition**:
```go
fields.New("file", fields.TypeStringList, fields.WithHelp(
    "File(s) to attach before sending the prompt. "+
    "Specify multiple files with repeated --file flags "+
    "(e.g., --file a.txt --file b.txt). "+
    "Backward compatible with single file paths."))
```

**Tool args**:
```go
// Before
if s.File != "" {
    toolArgs["file"] = s.File
}

// After
if len(s.Files) > 0 {
    toolArgs["files"] = s.Files  // Key changed to "files" (list)
}
```

### Provider: `go/internal/host/providers/chatgpt.go`

**ChatGPTRequest struct**:
```go
// Before
File string

// After
Files []string
```

**New helpers added** (local duplicates of router helpers):

```go
// Converts any value to []string — handles nil, string (comma-split), []string, []any
func toStringArray(v any) []string { ... }

// Extracts file list from tool args with backward compat for legacy "file" key
func fileListFromArgs(args map[string]any) []string {
    if v, ok := args["files"]; ok && v != nil {
        return toStringArray(v)  // New format: list of strings
    }
    if v, ok := args["file"]; ok && v != nil {
        return toStringArray(v)  // Legacy format: comma-separated string
    }
    return nil
}
```

**parseChatGPTRequest**:
```go
// Before
File: strings.TrimSpace(asString(args["file"])),

// After
Files: fileListFromArgs(args),
```

**uploadFiles signature**:
```go
// Before
func (b *chatGPTBridge) uploadFiles(ctx context.Context, rawFiles string) error {
    files := splitFileList(rawFiles)
    ...

// After
func (b *chatGPTBridge) uploadFiles(ctx context.Context, files []string) error {
    if len(files) == 0 {
        return fmt.Errorf("No files to upload")
    }
    ...
```

**runChatGPTQuery call site**:
```go
// Before
if req.File != "" {
    if err := bridge.uploadFiles(ctx, req.File); err != nil {

// After
if len(req.Files) > 0 {
    if err := bridge.uploadFiles(ctx, req.Files); err != nil {
```

### Tests: `go/internal/host/providers/chatgpt_test.go`

**New test** (`TestHandleChatGPTToolWithMultipleFiles`):
```go
_, err := HandleChatGPTTool(context.Background(), caller, map[string]any{
    "query": "review these files",
    "files": []string{"/tmp/a.txt", "/tmp/b.txt", "/tmp/c.md"},
}, nil, nil)
if len(uploadedFiles) != 3 { ... }
```

**Existing test** (`TestHandleChatGPTToolWithFileUpload`) — unchanged, still passes:
```go
// Uses legacy "file" key — backward compat works
HandleChatGPTTool(context.Background(), caller, map[string]any{
    "query": "review this file",
    "file":  "/tmp/demo.txt",  // legacy string key
}, nil, nil)
```

## Backward Compatibility

| Format | Example | Status |
|--------|---------|--------|
| Single `--file` | `--file a.txt` | ✅ Glazed coerces to `[]string{"a.txt"}` |
| Multiple `--file` | `--file a.txt --file b.txt` | ✅ New syntax |
| Comma-separated | `--file a.txt,b.txt` | ✅ Provider's `toStringArray` splits by comma |
| Legacy `"file"` key | `{"file": "/tmp/demo.txt"}` | ✅ Provider checks both keys |

## Why Duplicate toStringArray?

`router/toolmap.go` has a `toStringArray` function, but it's private (lowercase). Options:
1. Export it to a shared package — more refactoring
2. Duplicate it locally — simpler, isolated change

Chose option 2. The function is small and well-understood.

## Tasks Completed

- [x] Task 1: Update CLI Settings Struct (`File string` → `Files []string`)
- [x] Task 2: Update Flag Definition (`TypeString` → `TypeStringList`)
- [x] Task 3: Update Tool Args Construction
- [x] Task 4: Update Provider Request Struct
- [x] Task 5: Update Provider Parsing with backward compat
- [x] Task 6: Update Provider Upload Call
- [x] Task 7: Update Provider `uploadFiles` signature
- [x] Task 8: Update Tests (legacy + new multi-file)
- [x] Task 9: Verify backward compatibility
- [x] Verify help text

## Commits

- `8487a00` feat(surf chatgpt): change --file flag to stringlist for multi-file upload
- `f6e7c1c` docs(ttmp): add SURF-20260414-R7 ticket and implementation guide
