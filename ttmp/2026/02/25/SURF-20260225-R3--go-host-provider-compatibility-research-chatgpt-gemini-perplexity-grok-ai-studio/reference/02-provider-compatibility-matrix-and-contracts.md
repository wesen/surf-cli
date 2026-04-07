---
Title: Provider Compatibility Matrix and Contracts
Ticket: SURF-20260225-R3
Status: active
Topics:
    - go
    - native-messaging
    - architecture
    - migration
    - chatgpt
    - providers
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/internal/host/router/toolmap.go
      Note: Go provider block status source
    - Path: native/cli.cjs
      Note: Node CLI output branch expectations
    - Path: native/host-helpers.cjs
      Note: Provider command mapping matrix source
    - Path: native/host.cjs
      Note: Provider response shape contract source
    - Path: src/service-worker/index.ts
      Note: Provider primitive command availability
    - Path: ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/sources/01-provider-compat-inventory.json
      Note: Generated contract inventory
ExternalSources: []
Summary: Canonical provider command/message/response contract matrix for Node parity and Go-host implementation.
LastUpdated: 2026-02-25T19:43:23-05:00
WhatFor: Act as implementation contract for provider support in go-core.
WhenToUse: Use while implementing provider handlers, writing tests, and reviewing compatibility.
---


# Provider Compatibility Matrix and Contracts

## Goal

Define exact provider command contracts and parity requirements between current Node runtime and target Go runtime.

## Source of truth

1. `native/host-helpers.cjs` tool->message mappings.
2. `native/host.cjs` provider orchestration and response shaping.
3. `src/service-worker/index.ts` provider primitive handlers.
4. `native/cli.cjs` tool-specific output expectations.
5. `go/internal/host/router/toolmap.go` current go-core restrictions.
6. `sources/01-provider-compat-inventory.json` generated inventory.

## Current compatibility status

| Provider tool | Node runtime | Go runtime (`core-go`) | Reason |
|---|---|---|---|
| `chatgpt` | Supported | Rejected | Blocked by provider prefix check |
| `gemini` | Supported | Rejected | Blocked by provider prefix check |
| `perplexity` | Supported | Rejected | Blocked by provider prefix check |
| `grok` | Supported (`query`, `validate`) | Rejected | Blocked by provider prefix check |
| `aistudio` | Supported | Rejected | Blocked by provider prefix check |
| `aistudio.build` | Supported | Rejected | Blocked by provider prefix check |

## Tool -> native message contract

| Tool | Native message type(s) | Key args |
|---|---|---|
| `chatgpt` | `CHATGPT_QUERY` | `query`, `model`, `withPage`, `file`, `timeout` |
| `gemini` | `GEMINI_QUERY` | `query`, `model`, `withPage`, `file`, `generateImage`, `editImage`, `output`, `youtube`, `aspectRatio`, `timeout` |
| `perplexity` | `PERPLEXITY_QUERY` | `query`, `mode`, `model`, `withPage`, `timeout` |
| `grok` | `GROK_QUERY`, `GROK_VALIDATE` | `query`, `model`, `deepSearch`, `withPage`, `timeout`, `saveModels` |
| `aistudio` | `AISTUDIO_QUERY` | `query`, `model`, `withPage`, `timeout` |
| `aistudio.build` | `AISTUDIO_BUILD` | `query`, `model`, `output`, `keepOpen`, `timeout` |

## Provider primitive dependency matrix (extension side)

| Primitive | ChatGPT | Gemini | Perplexity | Grok | AI Studio | AI Studio Build |
|---|---:|---:|---:|---:|---:|---:|
| `GET_PAGE_TEXT` | Optional (`withPage`) | Optional | Optional | Optional | Optional | No |
| `GET_CHATGPT_COOKIES` | Yes | No | No | No | No | No |
| `GET_GOOGLE_COOKIES` | No | Yes | No | No | Yes | Yes |
| `GET_TWITTER_COOKIES` | No | No | No | Yes | No | No |
| `*_NEW_TAB` | Yes | No | Yes | Yes | Yes | Yes |
| `*_CLOSE_TAB` | Yes | No | Yes | Yes | Yes | Yes |
| `*_EVALUATE` | Yes | No | Yes | Yes | Yes | Yes |
| `*_CDP_COMMAND` | Yes | No | Yes | Yes | Yes | Yes |
| `READ_NETWORK_REQUESTS` | No | No | No | No | Optional | No |
| `DOWNLOADS_SEARCH` | No | No | No | No | No | Yes |

## Response shape contracts (Node baseline)

| Tool | Node host result payload keys |
|---|---|
| `chatgpt` | `response`, `model`, `tookMs` |
| `gemini` | `response`, `model`, `tookMs`, optional `imagePath` |
| `perplexity` | `response`, `sources`, `url`, `mode`, `model`, `tookMs` |
| `grok` | `response`, `model`, `tookMs`, optional `thinkingTime`, `deepSearch`, `partial`, `warnings`, `modelSelectionFailed` |
| `grok --validate` | validation struct (`authenticated`, `premium`, `inputFound`, `sendButtonFound`, `models`, `expectedModels`, `modelMismatch`, `errors`, etc.) |
| `aistudio` | serialized JSON string in `output` containing `response`, `model`, `thinkingTime`, `tookMs` |
| `aistudio.build` | serialized JSON string in `output` containing `zipPath`, optional `extractedPath`, `model`, `buildDuration`, `tookMs` |

## CLI expectation matrix (important for parity)

| Tool | Node CLI expectation |
|---|---|
| `chatgpt`, `gemini` | `data.response` text + stderr metadata (`model`, `tookMs`) |
| `perplexity` | `data.response` + sources/mode/model/time metadata |
| `aistudio` | `data.response` after decoding `output` payload |
| `aistudio.build` | `data.zipPath` and optional `data.extractedPath` after decoding `output` payload |

## Go runtime blockers

1. Provider prefix blocklist: `go/internal/host/router/toolmap.go` (`providerPrefixes`).
2. Provider rejection enforced by tests: `go/internal/host/router/toolmap_test.go`.

## Parity requirements for Go implementation

### Required

1. Same tool names and arg keys accepted.
2. Same provider primitive command usage to service worker (initial phase).
3. Response key compatibility with Node for each provider command.
4. Error text actionable and equivalent in meaning.

### Nice-to-have (can phase later)

1. Identical metadata formatting in human CLI output.
2. Shared cross-runtime structured schema normalization.

## Suggested conformance test cases

1. Happy path per provider tool.
2. Missing login cookie per provider.
3. Model not found/unavailable.
4. Timeout behavior.
5. Extension disconnect mid-request.
6. Snap socket path override (`SURF_SOCKET_PATH`) path correctness.

## Quick commands

```bash
# Rebuild provider inventory artifact
node ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/scripts/provider-compat-inventory.cjs

# Inspect current inventory
cat ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/sources/01-provider-compat-inventory.json
```
