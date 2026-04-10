---
Title: Gmail research diary
Ticket: SURF-20260410-R5
Status: active
Topics:
  - surf-go
  - gmail
  - browser-automation
  - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Ordered investigation notes and accepted selectors for the first Gmail list/search command family.
LastUpdated: 2026-04-10T15:05:00-04:00
WhatFor: Preserve the exact Gmail DOM findings and probe sequence used to build the initial Gmail commands.
WhenToUse: Use when modifying Gmail list/search selectors or debugging Gmail page-state issues.
---

# Gmail research diary

## Probe sequence

Ordered scripts created for this ticket:

1. `01-gmail-page-markers.js`
2. `02-gmail-inbox-row-inventory.js`
3. `03-gmail-search-box-inventory.js`
4. `04-gmail-search-submit-probe.js`
5. `05-gmail-thread-row-detail.js`
6. `06-gmail-semantic-field-probe.js`

## Findings

Live Gmail session:

- URL: `https://mail.google.com/mail/u/0/#inbox`
- Title shape: `Inbox (...) - <account> - Gmail`
- Logged-in markers:
  - thread list present
  - Gmail search box present
  - Gmail shell loaded

Accepted row selector:

- `tr.zA`

Accepted semantic selectors:

- participant: `.yP, .yW span[email], .yW`
- subject: `.bog, .y6 span[id]`
- snippet: `.y2`
- timestamp: `.xW span, .xW .xS`

Accepted thread-id source:

- descendant nodes carrying `data-thread-id`
- descendant nodes carrying `data-legacy-thread-id`

Accepted state markers:

- unread: row class `zE`
- star state: exact tooltip/title equality with `Starred`
- attachment presence: `Attachment:` text or attachment-like nodes

Accepted search controls:

- search input: `input[name="q"], input[aria-label*="Search mail"]`
- search button: `button[aria-label="Search mail"]`

## Important pitfall

The first Gmail search implementation returned inbox rows because it treated the existence of `tr.zA` rows as sufficient readiness. That was incorrect because inbox rows exist before search results load.

The corrected implementation waits for:

- a Gmail search route in `location.href` or search-results title text
- and a changed row snapshot relative to the initial inbox rows

That is the key Gmail-specific difference from simpler page scrapers.
