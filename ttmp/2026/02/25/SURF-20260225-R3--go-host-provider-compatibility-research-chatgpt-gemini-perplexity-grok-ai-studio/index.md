---
Title: Go Host Provider Compatibility Research (ChatGPT/Gemini/Perplexity/Grok/AI Studio)
Ticket: SURF-20260225-R3
Status: active
Topics:
    - go
    - native-messaging
    - architecture
    - migration
    - chatgpt
    - providers
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/internal/host/router/toolmap.go
      Note: Current go-core provider support boundary
    - Path: native/host.cjs
      Note: Reference implementation for provider orchestration
    - Path: src/service-worker/index.ts
      Note: Provider primitive implementation surface
ExternalSources:
    - https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging
    - https://developer.chrome.com/docs/extensions/develop/concepts/service-workers/lifecycle
    - https://developer.chrome.com/docs/extensions/reference/runtime
    - https://snapcraft.io/docs/environment-variables
Summary: Research workspace for achieving provider command parity between Node and Go native host runtimes.
LastUpdated: 2026-02-25T19:43:23-05:00
WhatFor: Track architecture findings, migration plan, contracts, and delivery artifacts.
WhenToUse: Use when implementing or reviewing provider support in go-core runtime.
---


# Go Host Provider Compatibility Research (ChatGPT/Gemini/Perplexity/Grok/AI Studio)

## Overview

This ticket documents the full architecture and migration strategy for bringing provider commands (`chatgpt`, `gemini`, `perplexity`, `grok`, `aistudio`, `aistudio.build`) to the Go host runtime with Node parity.

Core finding:
1. Go transport/runtime is already capable.
2. Provider support is currently blocked intentionally in router mapping.
3. Service worker primitives required by providers already exist.

## Primary Deliverables

1. Design doc:
   `design-doc/01-go-host-provider-compatibility-exhaustive-architecture-and-migration-research.md`
2. Investigation diary:
   `reference/01-investigation-diary.md`
3. Contract matrix:
   `reference/02-provider-compatibility-matrix-and-contracts.md`
4. Reproducible source artifact:
   `sources/01-provider-compat-inventory.json`
5. Repro script:
   `scripts/provider-compat-inventory.cjs`

## Current Status

1. Evidence gathering complete.
2. Architecture and migration recommendation complete.
3. Contract matrix complete.
4. Ticket bookkeeping + validation + reMarkable delivery in progress/completed per `tasks.md` and `changelog.md`.

## Key Internal Evidence Files

1. `native/host-helpers.cjs`
2. `native/host.cjs`
3. `src/native/port-manager.ts`
4. `src/service-worker/index.ts`
5. `go/internal/host/router/toolmap.go`
6. `go/cmd/surf-host-go/main.go`
7. `scripts/install-native-host.cjs`

## Tasks and Change History

1. See `tasks.md` for checklist status.
2. See `changelog.md` for dated updates.
