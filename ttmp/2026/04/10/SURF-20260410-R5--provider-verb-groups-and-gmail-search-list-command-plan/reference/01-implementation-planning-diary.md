---
Title: Implementation planning diary
Ticket: SURF-20260410-R5
Status: active
Topics:
  - surf-go
  - glazed
  - cli
  - browser-automation
  - gmail
  - chatgpt
  - kagi
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Planning diary for grouped provider verbs and the first Gmail command family.
LastUpdated: 2026-04-10T15:05:00-04:00
WhatFor: Record the planning decisions, scope boundaries, and implementation sequencing for grouped provider verbs and Gmail commands.
WhenToUse: Use when resuming the ticket or handing off implementation to another developer.
---

# Implementation planning diary

## Goal

Provide a concise but concrete handoff reference for the next implementation pass: grouped provider verbs plus the first Gmail list/search commands.

## Context

The current surf-go CLI already has several provider-specific browser verbs, but they are still flat at the root. The codebase now has enough working precedent to standardize provider grouping and apply the same pattern to Gmail.

## Quick Reference

### Current root-level provider commands

- `chatgpt`
- `chatgpt-transcript`
- `kagi-search`
- `kagi-assistant`

### Target grouped structure

- `chatgpt ask`
- `chatgpt transcript`
- `kagi search`
- `kagi assistant`
- `gmail list`
- `gmail search`

### Core files to read first

- [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go)
- [chatgpt.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt.go)
- [chatgpt_transcript.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt_transcript.go)
- [kagi_search.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_search.go)
- [kagi_assistant.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_assistant.go)
- [01-building-browser-side-verbs.md](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/pkg/doc/tutorials/01-building-browser-side-verbs.md)

### Implementation constraints

- keep new public commands as real Glazed commands
- do not add backwards-compatibility aliases unless explicitly requested
- use `go:embed` for production JS
- store Gmail research probes under this ticket's `scripts/` directory with numeric prefixes
- model owned-tab cleanup explicitly and use `--keep-tab-open` where needed

## Planning notes

### Why the grouping refactor comes first

Gmail should not be added as another top-level verb. Grouping is not cosmetic here; it defines the stable namespace that the Gmail family should live under.

### Why Gmail starts with list/search

Those are read-only enough to validate selectors and flow safely. Attachments, message download, or mutating commands should wait until the first read surface is stable.

### Why a new Gmail/stateful help page is part of the scope

The existing browser-side playbook is strong for extraction-oriented verbs, but Gmail introduces more page-state management. The help surface should reflect that before more stateful commands get added ad hoc.

## Implementation notes

### Grouped provider tree

The grouped command refactor landed as real Cobra parents:

- `surf-go chatgpt ask`
- `surf-go chatgpt transcript`
- `surf-go kagi search`
- `surf-go kagi assistant`
- `surf-go gmail list`
- `surf-go gmail search`

The old flat provider paths were removed from the root command registration.

### Gmail research scripts

Ordered Gmail probes created under this ticket:

- `01-gmail-page-markers.js`
- `02-gmail-inbox-row-inventory.js`
- `03-gmail-search-box-inventory.js`
- `04-gmail-search-submit-probe.js`
- `05-gmail-thread-row-detail.js`
- `06-gmail-semantic-field-probe.js`

These probes established the selectors used in the production implementation.

### Accepted Gmail selectors

Inbox and search rows:

- `tr.zA`

Semantic fields:

- participant: `.yP, .yW span[email], .yW`
- subject: `.bog, .y6 span[id]`
- snippet: `.y2`
- timestamp: `.xW span, .xW .xS`

Thread identity:

- descendant nodes with `data-thread-id`
- descendant nodes with `data-legacy-thread-id`

State:

- unread: row class `zE`
- star state: tooltip/title equality check for `Starred`
- attachment presence: `Attachment:` marker in text or attachment-like nodes

Search controls:

- search input: `input[name="q"], input[aria-label*="Search mail"]`
- search button: `button[aria-label="Search mail"]`

### Gmail search-specific pitfall

The first production `gmail search` implementation returned inbox rows because it treated the existence of `tr.zA` rows as sufficient readiness. That was incorrect because inbox rows already exist before Gmail finishes activating the search view.

The corrected search script now waits for both:

- a Gmail search route such as `#search/...`
- and a changed thread snapshot relative to the initial inbox state

That change was necessary to make the command return search-state data instead of the inbox baseline.

## Usage examples

These are target examples after implementation, not current commands:

```bash
surf-go chatgpt ask "hello"
surf-go chatgpt transcript --with-activity
surf-go kagi search --query "bovine flatulence"
surf-go kagi assistant "hello" --assistant Quick
surf-go gmail list --inbox --max-results 25
surf-go gmail search --query "from:boss has:attachment"
```
