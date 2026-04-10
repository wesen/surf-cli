---
Title: Shared tab readiness helper and ChatGPT extraction bug
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
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/integration_test.go
      Note: Integration tests updated for tab.new -> readiness probe -> js -> tab.close
    - Path: go/internal/cli/commands/kagi_assistant.go
      Note: Adopts the shared tab readiness helper for fresh Kagi assistant tabs
    - Path: go/internal/cli/commands/kagi_search.go
      Note: Adopts the shared tab readiness helper for fresh Kagi search tabs
    - Path: go/internal/cli/commands/tab_ready.go
      Note: Shared exact-tab readiness helper for fresh-tab commands
    - Path: go/internal/host/providers/chatgpt.go
      Note: Turn-based assistant extraction in interactive ChatGPT polling
    - Path: go/internal/host/providers/chatgpt_test.go
      Note: Provider tests updated for turn-based extraction polling
ExternalSources: []
Summary: Track the shared fresh-tab readiness helper and the ChatGPT interactive extraction bug fix.
LastUpdated: 2026-04-10T10:45:00-04:00
WhatFor: Coordinate implementation and documentation for exact tab readiness and ChatGPT response extraction fixes.
WhenToUse: Use when working on fresh-tab command stability or ChatGPT extraction correctness.
---


# Shared tab readiness helper and ChatGPT extraction bug

## Overview

This ticket tracks two linked fixes:

- unify fresh-tab creation around an exact-tab readiness helper
- fix the ChatGPT interactive response extractor so it returns the same substantive answer that `chatgpt-transcript` sees on the final conversation page

## Key Links

- Design: [Shared tab readiness helper design](./design-doc/01-shared-tab-readiness-helper-design.md)
- Implementation guide: [Shared tab readiness helper implementation guide](./design-doc/02-implementation-guide.md)
- Bug report: [ChatGPT extraction bug report](./reference/01-chatgpt-extraction-bug-report.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**
