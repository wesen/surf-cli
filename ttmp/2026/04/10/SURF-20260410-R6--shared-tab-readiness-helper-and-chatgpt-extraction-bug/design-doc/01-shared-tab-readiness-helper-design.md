---
Title: Shared tab readiness helper design
Ticket: SURF-20260410-R6
Status: active
Topics:
  - surf-go
  - glazed
  - cli
  - browser-automation
  - chatgpt
  - kagi
  - gmail
  - debugging
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Unify fresh-tab creation around exact tab-id ownership and a reusable readiness handshake before running page JS.
LastUpdated: 2026-04-10T10:45:00-04:00
WhatFor: Define the reusable helper for commands that create a tab and immediately run JS against it.
WhenToUse: Use when implementing or reviewing surf-go commands that open tabs and need deterministic page readiness before extraction.
---

# Shared tab readiness helper design

## Executive Summary

Several browser-side commands in `surf-go` create a new tab and then immediately execute JavaScript in that tab. That pattern is currently open-coded. It produces inconsistent behavior: ad hoc sleeps in one command, retry loops in another, and no shared contract for when a command may safely assume the newly created tab is executable and on the intended page.

The solution is to introduce a shared helper in `go/internal/cli/commands` that:

- opens a tab and captures the exact returned `tabId`
- probes that exact `tabId` until JavaScript execution works reliably
- verifies the tab is no longer `about:blank`
- optionally verifies exact or prefix URL matching
- returns control only when the tab is genuinely ready for extraction or interaction

This helper should replace command-local `tab.new` plus hand-written sleep/retry logic in Kagi and future Gmail commands.

## Problem Statement

The current fresh-tab pattern has three defects.

First, readiness is not modeled explicitly. A command may get a valid `tabId` from `tab.new` and still fail to execute JavaScript because the page has not finished creating an execution context.

Second, the logic is duplicated. `kagi-search` and `kagi-assistant` each evolved their own workaround paths.

Third, the behavior is hard to test because the transport sequence is implicit rather than standardized.

## Proposed Solution

Create a shared helper in `go/internal/cli/commands` with two responsibilities:

1. `openOwnedTab(...)`
   - call `tab.new`
   - extract `tabId`
   - call `waitForTabReady(...)`

2. `waitForTabReady(...)`
   - execute a small JS probe against the exact `tabId`
   - retry while the browser reports `Cannot find default execution context`
   - require `document.readyState === "complete"`
   - require `location.href` to be non-empty and not `about:blank`
   - optionally require exact URL match or prefix match

The shared probe returns:

- `href`
- `title`
- `readyState`

That is enough for deterministic gating without committing to page-specific selectors.

## Design Decisions

### Decision 1: Use exact `tabId` ownership, not title or URL matching, as the primary identity

The unique identifier is the `tabId` returned by `tab.new`. URL checks are secondary validation, not the primary way to find a tab.

### Decision 2: Keep page-specific readiness separate from transport readiness

The shared helper should only guarantee that JS can run in the correct tab and that the page is no longer blank. Page-specific extraction readiness should remain inside the command's own embedded JS.

### Decision 3: Commands that create tabs own those tabs

If a command creates a tab via the shared helper, that command owns the tab and may close it by default unless `--keep-tab-open` is set.

## Expected Tool Sequence

For commands that create tabs, the normalized tool sequence should become:

1. `tab.new`
2. `js` readiness probe on the exact `tabId`
3. `js` page-specific extractor or interaction script
4. `tab.close` if the command owns the tab and cleanup is enabled

This sequence should be reflected in integration tests.

## Commands That Should Use the Helper

Current commands:

- `kagi-search`
- `kagi-assistant`

Planned commands:

- `gmail list`
- `gmail search`

Commands intentionally operating on the current page, such as `chatgpt-transcript`, do not need the fresh-tab helper.

## Implementation Plan

1. Add the shared helper file under `go/internal/cli/commands/`.
2. Update `kagi-search` to use the helper instead of ad hoc retries.
3. Update `kagi-assistant` to use the helper instead of ad hoc retries.
4. Update integration tests to expect the readiness probe between `tab.new` and the main JS script.
5. Reuse the helper in future Gmail commands.
