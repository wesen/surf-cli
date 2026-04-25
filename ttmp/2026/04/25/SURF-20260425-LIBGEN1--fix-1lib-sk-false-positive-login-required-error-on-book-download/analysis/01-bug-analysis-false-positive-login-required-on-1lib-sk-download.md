---
title: "Bug Analysis: False-Positive Login Required on 1lib.sk Download"
doc_type: analysis
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
  - /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/libgen_download.go:Contains the buggy login check JavaScript
---

# Bug Analysis: False-Positive Login Required on 1lib.sk Download

## Summary

When downloading a book from 1lib.sk via `surf-go libgen download`, the command reports **"login required"** even when the user is properly logged in and the download would otherwise succeed. The browser successfully downloads the file, but the Go CLI aborts with a false-positive error.

## Root Cause

In `go/internal/cli/commands/libgen_download.go`, the function `buildLibgenDownloadCheckCode()` executes this JavaScript to detect login requirements:

```javascript
// Check for login requirement
var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink) {
  result.hasError = true;
  result.errorType = 'login_required';
}
```

**The problem:** `document.querySelector()` finds the first matching element in the DOM, regardless of whether it is **visible** to the user. When a user is logged in to 1lib.sk, the page may still contain `a[data-mode="singlelogin"]` elements in the DOM (e.g., hidden by `display: none`, `visibility: hidden`, or `opacity: 0`), but the code treats their mere presence as a login-required condition.

## Evidence

### Reproduction on Simulated Page

A test page was created with:
- A visible download button (`a[href="/dl/12345"]`)
- Two hidden login links (`a[data-mode="singlelogin"]`) — one via `display: none`, one via `visibility: hidden`

**Buggy check result:**
```json
{
  "hasError": true,
  "errorType": "login_required",
  "hasDownloadButton": true
}
```

**Expected result:**
```json
{
  "hasError": false,
  "errorType": "",
  "hasDownloadButton": true
}
```

### Verification on Real 1lib.sk

On the 1lib.sk homepage (where login IS actually required and the link IS visible):
- Total `a[data-mode="singlelogin"]` elements: 3
- Visible elements: 2

The fix correctly still reports `login_required` here because the link is actually visible.

### Real Book Page Test

Tested on an actual book page: `https://1lib.sk/book/rO7GeyL4O2/black-hat-go-go-programming-for-hackers-and-pentesters.html`

**Not logged in (login links visible):**
```json
{
  "hasDownloadButton": true,
  "hasError": true,
  "errorType": "login_required",
  "visibleLoginLinks": 2
}
```

**Simulated logged-in state (hidden login links via `display: none`):**

*Fixed version:*
```json
{
  "hasDownloadButton": true,
  "hasError": false,
  "errorType": "",
  "visibleLoginLinks": 0
}
```

*Buggy version (same state):*
```json
{
  "hasDownloadButton": true,
  "hasError": true,
  "errorType": "login_required"
}
```

This conclusively demonstrates the false positive on a real 1lib.sk page.

## Why Chromium Exposes a Logged-In Section

The browser extension / native host executes JavaScript via CDP (`Runtime.evaluate`), which has full access to the DOM — including hidden elements. This is different from what a user sees or what a visual scraper would observe. The Go code was written as if `querySelector` reflected visual state, but it does not.

## Impact

- Users who are logged in cannot use `--save-to` auto-download because the CLI aborts prematurely
- The false error message is confusing: "login required: please log in to your Z-Library account in the browser first"
- Manual workaround: omit `--save-to` and click the download URL manually

## Fix Strategy

Add a visibility check before treating the login link as an error condition:

```javascript
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
```

This mirrors the `isVis` helper already present in the surf extension's `piHelpers` (`src/service-worker/index.ts`).

## Related Code

- `go/internal/cli/commands/libgen_download.go` — `buildLibgenDownloadCheckCode()` (lines ~207-220)
- `src/service-worker/index.ts` — `piHelpers.waitForSelector()` with `isVis` helper (lines ~1920+)
