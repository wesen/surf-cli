---
Title: Go Native Host Lite (Core Browser Only) Implementation Plan with Glazed Commands
Ticket: SURF-20260225-R2
Status: active
Topics:
    - go
    - native-messaging
    - chromium
    - architecture
    - glazed
    - migration
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md
      Note: |-
        Primary implementation plan for Go Host Lite and Glazed command architecture
        Primary plan doc
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/01-implementation-diary.md
      Note: |-
        Chronological record of investigation and planning steps
        Chronological planning diary
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/02-phase-0-contract-inventory-and-freeze.md
      Note: |-
        Phase 0 contract freeze and machine-readable fixture references
        Phase 0 contract freeze summary
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md
      Note: |-
        Granular implementation backlog for execution
        Granular implementation backlog
ExternalSources: []
Summary: Ticket for planning and executing a Go-native host + Glazed CLI migration focused on core browser functionality and explicitly excluding provider-specific site logic in v1.
LastUpdated: 2026-02-25T17:37:00-05:00
WhatFor: Define and track an implementation-ready plan to ship Go Host Lite safely
WhenToUse: Use when executing migration work from Node native host core flows to Go
---



# Go Native Host Lite (Core Browser Only) Implementation Plan with Glazed Commands

## Overview

This ticket captures an implementation-ready migration plan for a **Go Host Lite** runtime that preserves core browser automation capabilities while excluding provider-specific site logic from v1.

Deliverables in this ticket:

1. primary design plan with architecture and phased execution guidance,
2. granular implementation backlog suitable for task-by-task delivery,
3. chronological diary of the planning process and evidence used.

## Key Links

1. Plan: `design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md`
2. Granular tasks: `tasks.md`
3. Phase 0 contract freeze: `reference/02-phase-0-contract-inventory-and-freeze.md`
4. Diary: `reference/01-implementation-diary.md`
5. Changelog: `changelog.md`

## Status

Current status: **active**

## Topics

- go
- native-messaging
- chromium
- architecture
- glazed
- migration

## Tasks

See [tasks.md](./tasks.md) for the current implementation backlog.

## Changelog

See [changelog.md](./changelog.md) for updates and delivery milestones.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
