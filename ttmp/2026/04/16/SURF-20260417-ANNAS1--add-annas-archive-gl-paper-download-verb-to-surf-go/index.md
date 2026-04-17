---
Title: Add annas-archive.gl paper download verb to surf-go
Ticket: SURF-20260417-ANNAS1
Status: active
Topics:
    - surf-go
    - browser-automation
    - annas-archive
    - cli
    - glazed
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: Added annas-archive command registration
    - Path: go/internal/cli/commands/annas_archive.go
      Note: Main command implementation following kagi_search pattern
    - Path: go/internal/cli/commands/scripts/annas_archive.js
      Note: Production JS extractor for SciDB and search pages
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-16T20:37:21.527364676-04:00
WhatFor: ""
WhenToUse: ""
---




# Add annas-archive.gl paper download verb to surf-go

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- surf-go
- browser-automation
- annas-archive
- cli
- glazed

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
