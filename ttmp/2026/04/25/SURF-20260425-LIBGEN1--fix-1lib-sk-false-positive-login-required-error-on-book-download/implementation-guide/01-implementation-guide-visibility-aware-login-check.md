---
title: "Implementation Guide: Visibility-Aware Login Check"
doc_type: implementation-guide
status: active
intent: long-term
topics:
  - browser-automation
  - cli
  - surf-go
  - libgen
  - 1lib
  - bugfix
created_date: 2026-04-25
updated_date: 2026-04-25
related_files:
  - /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/libgen_download.go:Target file for the fix
---

# Implementation Guide: Visibility-Aware Login Check

## Goal

Modify `buildLibgenDownloadCheckCode()` in `libgen_download.go` so that it only reports `login_required` when a `a[data-mode="singlelogin"]` element is **actually visible** on the page.

## Change Details

### File

`go/internal/cli/commands/libgen_download.go`

### Function to modify

`buildLibgenDownloadCheckCode()` (currently returns a raw JavaScript string)

### Current code (buggy)

```go
func buildLibgenDownloadCheckCode() string {
	return `
var h1 = document.querySelector('h1');
var result = {
  url: location.href,
  title: h1 ? h1.textContent.trim() : '',
  hasDownloadButton: !!document.querySelector('a[href*="/dl/"]'),
  hasError: false,
  errorType: ''
};

if (h1) {
  var text = h1.textContent.trim().toLowerCase();
  if (text.includes('daily limit') || text.includes('limit reached')) {
    result.hasError = true;
    result.errorType = 'daily_limit';
  } else if (text.includes('not found') || text.includes('error')) {
    result.hasError = true;
    result.errorType = 'not_found';
  }
}

// Check for login requirement
var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink) {
  result.hasError = true;
  result.errorType = 'login_required';
}

return result;
`
}
```

### New code (fixed)

```go
func buildLibgenDownloadCheckCode() string {
	return `
var h1 = document.querySelector('h1');
var result = {
  url: location.href,
  title: h1 ? h1.textContent.trim() : '',
  hasDownloadButton: !!document.querySelector('a[href*="/dl/"]'),
  hasError: false,
  errorType: ''
};

if (h1) {
  var text = h1.textContent.trim().toLowerCase();
  if (text.includes('daily limit') || text.includes('limit reached')) {
    result.hasError = true;
    result.errorType = 'daily_limit';
  } else if (text.includes('not found') || text.includes('error')) {
    result.hasError = true;
    result.errorType = 'not_found';
  }
}

// Check for login requirement — only if the link is actually visible
function isVisible(el) {
  if (!el) return false;
  var style = getComputedStyle(el);
  return style.display !== 'none' &&
         style.visibility !== 'hidden' &&
         style.opacity !== '0' &&
         el.offsetWidth > 0 &&
         el.offsetHeight > 0;
}

var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink && isVisible(loginLink)) {
  result.hasError = true;
  result.errorType = 'login_required';
}

return result;
`
}
```

## Testing Steps

1. **Build the Go binary:**
   ```bash
   cd go
   go build -o surf-go ./cmd/surf-go
   ```

2. **Test against the simulated page (automated):**
   Use Playwright or the browser's devtools console to evaluate the new JS on a page with hidden login links. Expect `hasError: false`.

3. **Test against 1lib.sk homepage (real page):**
   Evaluate on `https://1lib.sk/` without being logged in. Expect `hasError: true`, `errorType: 'login_required'`.

4. **Integration test (if you have a 1lib.sk account):**
   ```bash
   ./surf-go libgen download --id <book-id> --save-to ~/Downloads/test.pdf
   ```
   With the fix, a logged-in user should no longer see the false login error.

## Backwards Compatibility

- The change only affects the login-detection branch
- Daily-limit and not-found detection are unchanged
- If no login link exists, behavior is identical
- If a visible login link exists, behavior is identical

## Risk Assessment

**Low risk.** The change is a strict narrowing of the error condition. The only behavioral change is that hidden elements no longer trigger the error, which is the desired behavior.
