---
title: "Investigation Diary"
doc_type: reference
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
---

# Investigation Diary

## 2026-04-25 — Initial Investigation

### Problem Statement
User reports: when downloading a book from 1lib.sk, even though the download succeeds in the browser, surf-go reports a "login required failure". Suspected cause: latching onto a DOM element that is not visible when properly logged in. Chromium exposes the logged-in section.

### Files Examined
- `go/internal/cli/commands/libgen_download.go` — contains `buildLibgenDownloadCheckCode()`
- `go/internal/cli/commands/libgen.go` — parent command definition
- `src/service-worker/index.ts` — how JS execution works via CDP

### Root Cause Identified

`buildLibgenDownloadCheckCode()` uses:
```javascript
var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink) {
  result.hasError = true;
  result.errorType = 'login_required';
}
```

This does NOT check visibility. When logged in, 1lib.sk may still have `a[data-mode="singlelogin"]` in the DOM (hidden), and `querySelector` finds it anyway.

### Reproduction Experiment

Created `scripts/01-reproduce-bug.html` with:
- Visible download button
- Hidden login links (`display: none` and `visibility: hidden`)

Ran buggy JS via Playwright:
```json
{
  "buggy": { "hasError": true, "errorType": "login_required", "hasDownloadButton": true },
  "fixed": { "hasError": false, "errorType": "", "hasDownloadButton": true },
  "totalLoginLinks": 2,
  "visibleLoginLinks": 0
}
```

Bug confirmed: buggy version reports login_required, fixed version does not.

### Verification on Real Site

Navigated to `https://1lib.sk/` and ran fixed check:
```json
{
  "hasError": true,
  "errorType": "login_required",
  "loginLinkVisible": true,
  "totalLoginLinks": 3,
  "visibleLoginLinks": 2
}
```

Fix still correctly detects login required when the link IS visible.

### Next Steps
- [x] Create docmgr ticket
- [x] Write analysis and implementation guide
- [x] Apply fix to `libgen_download.go`
- [ ] Build and verify Go binary compiles
- [ ] Update changelog

## 2026-04-25 — Fix Applied

Applied the visibility-aware check to `buildLibgenDownloadCheckCode()` in `go/internal/cli/commands/libgen_download.go`.

```
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
```

Build verified with `go build`.

## 2026-04-25 — Real-World Test on 1lib.sk Book Page

Navigated to a real book page: `https://1lib.sk/book/rO7GeyL4O2/black-hat-go-go-programming-for-hackers-and-pentesters.html`

### Test 1: Not logged in (login links visible)
Fixed check correctly reports:
```json
{
  "hasDownloadButton": true,
  "hasError": true,
  "errorType": "login_required",
  "visibleLoginLinks": 2
}
```

### Test 2: Simulated logged-in state (hidden login links)
Used JS to hide all `a[data-mode="singlelogin"]` elements (`display: none`), simulating how 1lib.sk behaves when a user is logged in.

**Fixed version result:**
```json
{
  "hasDownloadButton": true,
  "hasError": false,
  "errorType": "",
  "visibleLoginLinks": 0
}
```
✅ No false positive!

**Buggy version result (same page state):**
```json
{
  "hasDownloadButton": true,
  "hasError": true,
  "errorType": "login_required"
}
```
❌ False positive — the old code would abort the download even though the user is logged in.

## 2026-04-25 — End-to-End surf-go CLI Test on Logged-In Chromium

The stale surf-host-go process was killed, the browser auto-restarted it, and the socket was re-created at `/home/manuel/snap/chromium/common/surf-cli/surf.sock`.

Confirmed control of the user's actual Chromium by navigating to `https://1lib.sk/` via `surf-go navigate`.

Ran full download command:
```bash
./surf-go libgen download --id rO7GeyL4O2 --save-to /tmp/test-download.pdf --debug-socket
```

**Result:** The download proceeded past the login check correctly. The fixed JS code (with `isVisible`) was executed via the socket:

```javascript
// Check for login requirement — only if the link is actually visible
function isVisible(el) { ... }
var loginLink = document.querySelector('a[data-mode="singlelogin"]');
if (loginLink && isVisible(loginLink)) { ... }
```

It did **NOT** report `login_required`. Instead, it correctly navigated to the download URL and got a legitimate network response:

```json
{
  "errorType": "not_found",
  "hasDownloadButton": false,
  "hasError": true,
  "title": "Connection timed out\n                Error code 522",
  "url": "https://dlaswz.ncdn.ec/books-files/..."
}
```

The "Connection timed out / Error code 522" is a real Cloudflare CDN error — **not a false-positive login failure**. With the old buggy code, this same flow would have aborted early with `login_required` even though the user is properly logged in.

**Fix confirmed working on real logged-in browser.**
