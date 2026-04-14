---
Title: Implementation Guide - ChatGPT File StringList
Ticket: SURF-20260414-R7
---

# Implementation Guide: --file StringList for surf chatgpt ask

## Overview

Change the `--file` flag from `fields.TypeString` to `fields.TypeStringList`, allowing users to specify multiple files via repeated flags: `--file file1.txt --file file2.txt`.

## Current State

### Flag Definition (CLI)
```go
// chatgpt.go
type ChatGPTSettings struct {
    File string `glazed:"file"`  // single string
}

fields.New("file", fields.TypeString, ...)  // in NewChatGPTCommand()
```

### Message Passing (CLI → Provider)
```go
// chatgpt.go RunIntoGlazeProcessor
toolArgs["file"] = s.File  // single string
```

### Provider Parsing
```go
// providers/chatgpt.go parseChatGPTRequest
File: strings.TrimSpace(asString(args["file"])),  // single string

// providers/chatgpt.go uploadFiles
files := splitFileList(rawFiles)  // already splits by comma internally
// But: this is called with a single string, so comma-split is the only path
```

### Provider Upload
```go
// providers/chatgpt.go splitFileList
func splitFileList(raw string) []string {
    parts := strings.Split(raw, ",")
    // ... already handles comma-separated input
}
```

## Target State

The flag accepts multiple values, each being a file path. The provider receives a list of file paths.

### Flag Definition (CLI)
```go
type ChatGPTSettings struct {
    Files []string `glazed:"file"`  // string list
}

fields.New("file", fields.TypeStringList, ...)  // in NewChatGPTCommand()
```

### Message Passing (CLI → Provider)
```go
// chatgpt.go RunIntoGlazeProcessor
if len(s.Files) > 0 {
    toolArgs["files"] = s.Files  // pass as list
}
```

### Provider Parsing
```go
// providers/chatgpt.go parseChatGPTRequest
Files: toStringArray(args["files"]),  // handles []string, []any, string
```

### Provider Upload (no change needed)
```go
// The existing uploadFiles and splitFileList already handle []string
func (b *chatGPTBridge) uploadFiles(ctx context.Context, files []string) error {
    // No change needed — already accepts []string
}

// The rawFiles param becomes files []string, we remove splitFileList call
```

## Tasks

### Task 1: Update CLI Settings Struct
**File**: `go/internal/cli/commands/chatgpt.go`
**Change**: `File string` → `Files []string`
**Breaking**: No — callers using `--file path.txt` still work with Glazed's string-to-stringlist coercion

### Task 2: Update Flag Definition
**File**: `go/internal/cli/commands/chatgpt.go`
**Change**: `fields.TypeString` → `fields.TypeStringList`
**Help text**: Update to indicate multiple files supported

### Task 3: Update Tool Args Construction
**File**: `go/internal/cli/commands/chatgpt.go`
**Change**: Pass `Files` as list to tool args (key change: `"file"` → `"files"` for clarity)

### Task 4: Update Provider Request Struct
**File**: `go/internal/host/providers/chatgpt.go`
**Change**: `File string` → `Files []string`

### Task 5: Update Provider Parsing
**File**: `go/internal/host/providers/chatgpt.go`
**Change**: Use `toStringArray` to handle both legacy single-string and new list format

### Task 6: Update Provider Upload Call
**File**: `go/internal/cli/commands/chatgpt.go`
**Change**: Remove `splitFileList` call (no longer needed — list is already split)

### Task 7: Update Provider uploadFiles Signature
**File**: `go/internal/host/providers/chatgpt.go`
**Change**: `uploadFiles(ctx context.Context, rawFiles string)` → `uploadFiles(ctx context.Context, files []string)`  
Remove internal `splitFileList` call, pass slice directly

### Task 8: Update Tests
**Files**: `go/internal/host/providers/chatgpt_test.go`
**Changes**: Update test fixtures to use `Files []string` instead of `File string`

### Task 9: Verify Backward Compatibility
**Check**: Does `--file a.txt,b.txt` still work? (Yes, `toStringArray` handles comma-separated string)
**Check**: Does `--file a.txt --file b.txt` work? (Yes, new syntax)
**Check**: Does single `--file a.txt` work? (Yes, Glazed coerces string to []string)

## Implementation Notes

### Glazed StringList Behavior
Glazed's `TypeStringList` accepts:
- Repeated flags: `--file a.txt --file b.txt` → `[]string{"a.txt", "b.txt"}`
- Single flag: `--file a.txt` → Glazed coerces to `[]string{"a.txt"}`

This means backward compatibility is maintained for existing single-file usage.

### toStringArray already handles everything we need
```go
// From toolmap.go — already in the codebase
func toStringArray(v any) []string {
    switch raw := v.(type) {
    case nil:       return nil
    case string:    // ← handles legacy comma-separated
        if raw == "" { return nil }
        return strings.Split(raw, ",")
    case []string:  // ← handles new list format
        return raw
    case []any:    // ← handles []any from JSON
        // ...
    }
}
```

We can reuse this function in the provider.

## Files to Modify

1. `go/internal/cli/commands/chatgpt.go` — Tasks 1-3, 6
2. `go/internal/host/providers/chatgpt.go` — Tasks 4-5, 7
3. `go/internal/host/providers/chatgpt_test.go` — Task 8
4. `ttmp/2026/04/14/SURF-20260414-R7--chatgpt-file-stringlist/design-doc/01-implementation-guide.md` — This file
