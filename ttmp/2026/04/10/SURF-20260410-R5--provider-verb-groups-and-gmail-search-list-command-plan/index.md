---
Title: Provider verb groups and Gmail search/list command plan
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
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: Grouped provider parents and Gmail subcommand registration
    - Path: go/internal/cli/commands/gmail_list.go
      Note: Gmail inbox list command and Markdown/row output shaping
    - Path: go/internal/cli/commands/gmail_search.go
      Note: Gmail search command and Gmail search-state handling
    - Path: go/internal/cli/commands/scripts/gmail_list.js
      Note: Embedded Gmail inbox DOM extraction logic
    - Path: go/internal/cli/commands/scripts/gmail_search.js
      Note: Embedded Gmail search interaction and result extraction logic
    - Path: go/pkg/doc/tutorials/01-building-browser-side-verbs.md
      Note: Updated grouped command examples and Gmail tutorial link
    - Path: go/pkg/doc/tutorials/02-building-stateful-gmail-verbs.md
      Note: New stateful Gmail workflow tutorial
ExternalSources: []
Summary: Plan for grouping provider-specific verbs under provider parents and implementing the first Gmail command family.
LastUpdated: 2026-04-10T10:00:00-04:00
WhatFor: Track the namespace refactor for provider verbs and the first Gmail list/search implementation.
WhenToUse: Use when planning or implementing grouped provider commands and Gmail browser-side verbs in surf-go.
---


# Provider verb groups and Gmail search/list command plan

## Overview

This ticket covers two linked changes in `surf-go`:

- reorganize provider-specific commands into grouped provider namespaces
- implement the first Gmail command family under that grouped structure

The intended end state is a cleaner CLI tree such as `surf-go chatgpt ...`, `surf-go kagi ...`, and `surf-go gmail ...`, with `gmail list --inbox` and `gmail search` as the first Gmail read-oriented commands.

## Key Links

- Design: [Provider groups and Gmail command architecture](./design-doc/01-provider-groups-and-gmail-command-architecture.md)
- Planning diary: [Implementation planning diary](./reference/01-implementation-planning-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

## Scope Summary

Included:

- provider grouping for ChatGPT and Kagi commands
- Gmail list/search planning and implementation
- help/doc updates for grouped provider verbs and stateful Gmail workflows

Excluded:

- Gmail attachment download
- Gmail message export beyond list/search
- compatibility aliases for old flat command names unless explicitly requested later

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
