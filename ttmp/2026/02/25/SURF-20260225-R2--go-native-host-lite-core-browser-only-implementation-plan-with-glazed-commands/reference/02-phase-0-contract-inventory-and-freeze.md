---
Title: Phase 0 Contract Inventory and Freeze
Ticket: SURF-20260225-R2
Status: active
Topics:
    - go
    - native-messaging
    - chromium
    - architecture
    - glazed
    - migration
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: native/cli.cjs
      Note: |-
        Source of socket request envelope shapes
        Socket request envelope shape baseline
    - Path: native/host-helpers.cjs
      Note: |-
        Source of tool mapping inventory and alias behavior
        Tool mapping inventory and alias behavior baseline
    - Path: native/host.cjs
      Note: |-
        Source of host-side response normalization and lifecycle behavior
        Response normalization and disconnect behavior baseline
    - Path: src/native/port-manager.ts
      Note: |-
        Source of extension-side native host message expectations
        Extension native messaging expectations baseline
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/01-tool-inventory.json
      Note: Extracted tool inventory artifact
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/02-go-core-v1-classification.yaml
      Note: Core/provider/defer classification artifact
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/03-core-v1-envelope-contract.yaml
      Note: Frozen envelope contract artifact
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/04-go-v1-unsupported-tools.json
      Note: Explicit unsupported command list artifact
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/05-core-request-fixtures.json
      Note: Request->extension fixture table artifact
    - Path: ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/06-response-normalization-fixtures.json
      Note: Extension->socket normalization fixture table artifact
ExternalSources: []
Summary: Frozen Phase 0 artifacts for Go Host Lite v1 including tool inventory, v1 scope classification, unsupported list, envelope contract, and machine-readable request/response fixtures.
LastUpdated: 2026-02-25T17:47:00-05:00
WhatFor: Lock protocol and scope assumptions before implementing Go host transport and router
WhenToUse: Use this document when building or testing Go host parity against the Node baseline
---


# Phase 0 Contract Inventory and Freeze

## Goal

Provide a concrete, testable baseline for Phase 0 tasks (`T0.1` to `T0.8`) so Go Host Lite implementation starts from a frozen command/protocol contract rather than ad-hoc behavior.

## Context

The current Node host and CLI protocols are stable but broad. This freeze extracts the exact tool surface and response behavior needed for Go v1 core-browser implementation while explicitly identifying unsupported/deferred commands.

## Quick Reference

### Produced artifacts

1. `sources/01-tool-inventory.json`
2. `sources/02-go-core-v1-classification.yaml`
3. `sources/03-core-v1-envelope-contract.yaml`
4. `sources/04-go-v1-unsupported-tools.json`
5. `sources/05-core-request-fixtures.json`
6. `sources/06-response-normalization-fixtures.json`

### Phase 0 completion mapping

1. `T0.1` Extract tool names: `sources/01-tool-inventory.json`
2. `T0.2` Classify tools (`core-v1/provider/defer`): `sources/02-go-core-v1-classification.yaml`
3. `T0.3` Unsupported list for Go v1: `sources/04-go-v1-unsupported-tools.json`
4. `T0.4` Socket/native envelope freeze: `sources/03-core-v1-envelope-contract.yaml`
5. `T0.5` Response envelope rules freeze: `sources/03-core-v1-envelope-contract.yaml`
6. `T0.6` Disconnect parity requirements: `sources/03-core-v1-envelope-contract.yaml`
7. `T0.7` Request-to-extension fixture table: `sources/05-core-request-fixtures.json`
8. `T0.8` Extension-to-socket normalization fixtures: `sources/06-response-normalization-fixtures.json`

### Core policy for Go Host Lite v1

1. Provider/site-specific commands are unsupported in Go v1.
2. Deferred convenience commands are unsupported in Go v1.
3. Go host must preserve Node transport and lifecycle semantics (`HOST_READY`, frame decoding, disconnect broadcast).

## Usage Examples

### Generate a test list from fixtures

```bash
jq -r '.fixtures[].name' ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/05-core-request-fixtures.json
```

### Inspect unsupported commands

```bash
jq '.unsupported_provider_tools + .unsupported_deferred_tools' ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/04-go-v1-unsupported-tools.json
```

## Related

1. `design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md`
2. `tasks.md`
3. `reference/01-implementation-diary.md`
