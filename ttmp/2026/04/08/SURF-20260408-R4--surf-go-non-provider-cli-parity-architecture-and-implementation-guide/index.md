---
Title: surf-go Non-Provider CLI Parity Architecture and Implementation Guide
Ticket: SURF-20260408-R4
Status: active
Topics:
    - go
    - glazed
    - chromium
    - cli
    - native-messaging
    - architecture
    - onboarding
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: design-doc/01-surf-go-non-provider-cli-parity-detailed-architecture-and-implementation-guide.md
      Note: Primary intern-oriented architecture and implementation guide
    - Path: reference/01-investigation-diary.md
      Note: Chronological evidence log for this ticket
    - Path: tasks.md
      Note: Granular phased backlog for parity implementation
    - Path: sources/01-node-vs-go-non-provider-command-gap.json
      Note: Generated missing-command inventory
ExternalSources: []
Summary: Ticket workspace for the detailed non-provider surf-go parity guide and implementation backlog.
LastUpdated: 2026-04-08T21:28:00-04:00
WhatFor: Organize the architecture guide, evidence, backlog, and delivery artifacts for non-provider surf-go CLI parity work.
WhenToUse: Use when starting or reviewing implementation of missing non-provider surf-go commands.
---

# surf-go Non-Provider CLI Parity Architecture and Implementation Guide

## Overview

This ticket documents how the Surf command stack is structured and how to close the non-provider parity gap between the Node `surf` CLI and `surf-go`. The intended audience is a new engineer who needs architectural orientation, implementation guidance, and a concrete phased backlog.

## Primary Deliverables

- `design-doc/01-surf-go-non-provider-cli-parity-detailed-architecture-and-implementation-guide.md`
  - Main architecture, design, and implementation guide.
- `reference/01-investigation-diary.md`
  - Chronological record of evidence gathering and ticket creation.
- `tasks.md`
  - Granular phased implementation backlog.
- `sources/01-node-vs-go-non-provider-command-gap.json`
  - Generated inventory of missing non-provider commands.

## Status

Current status: **active**

The analysis and design deliverables are complete. Implementation work has not started in this ticket yet.

## Recommended Reading Order

1. Read the design doc.
2. Review the generated gap inventory.
3. Review the task backlog.
4. Use the investigation diary if you need to understand why a recommendation was made.

## Structure

- design-doc/ - main long-form architecture/design analysis
- reference/ - diary and future quick-reference material
- scripts/ - helper scripts if inventory generation becomes scripted beyond one-off commands
- sources/ - machine-generated or durable evidence artifacts
- various/ - scratch notes if needed later
- archive/ - deprecated or superseded material
