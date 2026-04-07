---
Title: 'Go Host Provider Compatibility: Exhaustive Architecture and Migration Research'
Ticket: SURF-20260225-R3
Status: active
Topics:
    - go
    - native-messaging
    - architecture
    - migration
    - chatgpt
    - providers
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-host-go/main.go
      Note: Go host runtime bridging and pending correlation
    - Path: go/internal/host/router/toolmap.go
      Note: Go router provider blocklist and mapping
    - Path: native/cli.cjs
      Note: CLI output expectations for provider commands
    - Path: native/host-helpers.cjs
      Note: Provider tool mapping contract
    - Path: native/host.cjs
      Note: Node provider orchestration and response shaping
    - Path: scripts/install-native-host.cjs
      Note: Runtime profile wrapper and snap install behavior
    - Path: src/native/port-manager.ts
      Note: Extension native messaging lifecycle
    - Path: src/service-worker/index.ts
      Note: Provider primitive handlers and command routing
ExternalSources:
    - https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging
    - https://developer.chrome.com/docs/extensions/develop/concepts/service-workers/lifecycle
    - https://developer.chrome.com/docs/extensions/reference/runtime
    - https://snapcraft.io/docs/environment-variables
Summary: Exhaustive architecture and compatibility research for adding provider tooling (ChatGPT/Gemini/Perplexity/Grok/AI Studio) to the Go native host runtime while preserving Node parity and Snap Chromium reliability.
LastUpdated: 2026-02-25T19:43:23-05:00
WhatFor: Decide architecture and phased implementation strategy for provider parity in the Go host runtime.
WhenToUse: Use when implementing or reviewing provider command support in go-core and when debugging cross-runtime (Node/Go) behavior.
---


# Go Host Provider Compatibility: Exhaustive Architecture and Migration Research

## Executive Summary

The current `core-go` runtime deliberately rejects all provider tools (`chatgpt`, `gemini`, `perplexity`, `grok`, `aistudio`, `aistudio.build`) even though the service worker already implements the low-level primitives those providers need. This is the main compatibility gap.

Provider logic today is split across:
1. Node host orchestration (`native/host.cjs`) for provider query workflows, retries, queueing, and response shaping.
2. Service worker primitives (`src/service-worker/index.ts`) for cookie reads, provider tab lifecycle, CDP command/eval execution, and downloads lookup.

The Go host stack already has the right transport pattern (socket bridge + pending correlation + native stdio codec), so provider support can be added incrementally by porting Node orchestration logic and unblocking router mappings, without changing extension primitives first.

Recommended direction:
1. Keep Node as compatibility reference.
2. Add a provider dispatcher in Go with one provider at a time.
3. Reuse existing service-worker message contracts first, then optimize/cleanup.

## Problem Statement And Scope

### Problem

User-facing behavior requires Node runtime for provider commands. `SURF_HOST_PROFILE=core-go` cannot run AI provider commands yet, creating a two-runtime operational split and confusing install/runtime behavior.

### In Scope

1. Command and message compatibility for `chatgpt`, `gemini`, `perplexity`, `grok`, `aistudio`, `aistudio.build`.
2. Host/runtime architecture and migration path (Node -> Go parity).
3. Contract matrix for request/response fields and provider primitives.
4. Risks and validation strategy.

### Out Of Scope

1. Reworking provider site-specific selector heuristics beyond parity requirements.
2. Replacing extension CDP/cookie primitives.
3. Replacing browser extension architecture.

## Current-State Architecture (Evidence-Based)

### End-to-End Transport Layers

1. Browser extension service worker connects with `chrome.runtime.connectNative("surf.browser.host")` (`src/native/port-manager.ts:56`).
2. On disconnect, it logs and reconnects unless host-not-found (`src/native/port-manager.ts:101-112`).
3. Node host and Go host both expose a local socket for CLI clients and bridge to extension native messaging stdio.

### Node Runtime Provider Flow

#### Tool-to-message mapping

Provider tools map in `native/host-helpers.cjs`:
1. `chatgpt -> CHATGPT_QUERY` (`native/host-helpers.cjs:1021-1031`)
2. `gemini -> GEMINI_QUERY` (`native/host-helpers.cjs:1032-1047`)
3. `perplexity -> PERPLEXITY_QUERY` (`native/host-helpers.cjs:1048-1058`)
4. `grok -> GROK_QUERY | GROK_VALIDATE` (`native/host-helpers.cjs:1059-1076`)
5. `aistudio -> AISTUDIO_QUERY` (`native/host-helpers.cjs:1077-1088`)
6. `aistudio.build -> AISTUDIO_BUILD` (`native/host-helpers.cjs:1089-1101`)

#### Provider orchestration in host

`native/host.cjs` handles provider query messages and drives sub-operations via extension primitives:
1. `CHATGPT_QUERY` flow (`native/host.cjs:454-553`)
2. `PERPLEXITY_QUERY` flow (`native/host.cjs:555-647`)
3. `GEMINI_QUERY` flow (`native/host.cjs:649-720`)
4. `GROK_QUERY` and `GROK_VALIDATE` (`native/host.cjs:722-917`)
5. `AISTUDIO_QUERY` and `AISTUDIO_BUILD` (`native/host.cjs:919-1109`)

Cross-request correlation uses `pendingToolRequests` (`native/host.cjs:301`, `1366-1372`).

#### Queueing and pacing

AI calls are serialized with queue + 2s inter-request delay (`native/host.cjs:59-81`).

### Service Worker Provider Primitives

Service worker implements the provider primitive command set used by Node host:
1. ChatGPT cookies/tab/CDP/eval (`src/service-worker/index.ts:2610-2666`)
2. Perplexity tab/CDP/eval (`src/service-worker/index.ts:2668-2718`)
3. Grok cookies/tab/CDP/eval (`src/service-worker/index.ts:2720-2803`)
4. AI Studio tab/CDP/eval/download search and Google cookies (`src/service-worker/index.ts:2805-2893`)
5. These commands are tabless in dispatcher whitelist (`src/service-worker/index.ts:2999-3012`).

### Go Runtime Boundaries

#### Core runtime and transport are present

1. Stdio framing codec for native messaging is implemented (`go/internal/host/nativeio/codec.go:11-82`).
2. Socket session + pending correlation store are implemented (`go/internal/host/pending/store.go:25-69`, `go/cmd/surf-host-go/main.go:216-283`).
3. Extension-disconnect behavior is explicit and propagates to clients (`go/cmd/surf-host-go/main.go:107-113`, `go/internal/host/socketbridge/session.go:88-98`).

#### Provider tools are intentionally blocked

1. Provider prefix blocklist in router (`go/internal/host/router/toolmap.go:13-21`).
2. Blocked tool rejection path (`go/internal/host/router/toolmap.go:50-52`).
3. Error message is explicit go-core unsupported (`go/internal/host/router/toolmap.go:41-43`).
4. Test enforces this behavior (`go/internal/host/router/toolmap_test.go:34-43`).

So current incompatibility is a design guardrail, not missing transport.

### Installer Runtime Selection

Installer writes wrapper that can switch runtime profile:
1. `node-full` default.
2. `core-go` if Go host binary exists.
3. Linux Chromium Snap target creates isolated runtime copy and snap socket path hint (`scripts/install-native-host.cjs:363-398`).
4. Profile hints are surfaced (`scripts/install-native-host.cjs:339`, `359`, `385`, `465-467`).

## Provider-Specific Logic And Third-Party Packages

### Provider logic location

1. ChatGPT: DOM/CDP workflow in `native/chatgpt-client.cjs`.
2. Perplexity: DOM/CDP workflow in `native/perplexity-client.cjs`.
3. Grok: DOM/CDP workflow + model validation/cache in `native/grok-client.cjs`.
4. Gemini: cookie-authenticated HTTPS flow in `native/gemini-client.cjs`.
5. AI Studio query/build: DOM/CDP + network extraction in `native/aistudio-client.cjs` and `native/aistudio-build.cjs`.

### Third-party packages used in provider path

From `package.json`:
1. `@google/generative-ai` is imported in `native/host.cjs` (`native/host.cjs:8`) for the generic `ai` helper mode.
2. Provider-specific clients largely use Node built-ins and custom modules, not large external automation frameworks.
3. MCP path uses `@modelcontextprotocol/sdk` and `zod` (`native/mcp-server.cjs:3-5`) for protocol surface, not provider site automation.

Go side external runtime stack is primarily `glazed` + `cobra` (`go/go.mod`), with provider logic not yet implemented.

## Gap Analysis Against Requested Outcome

Requested outcome: Node CLI compatibility against Go host for provider features.

Gaps:
1. Routing gap: provider tools are rejected before host forwarding.
2. Orchestration gap: Go host does not yet run provider workflows that chain primitive commands.
3. Output compatibility gap: Node provider responses have tool-specific shaping; e.g. AI Studio wraps payload JSON in `output` (`native/host.cjs:1019-1027`) and Node CLI has custom branches (`native/cli.cjs:3171-3210`).
4. Coverage gap: no Go contract tests for provider roundtrips comparable to Node behavior.

## External Runtime Constraints (Primary Docs)

1. Chrome Native Messaging requires host manifest and stdio framed protocol; host exits immediately disconnect the port (`developer.chrome.com/docs/extensions/develop/concepts/native-messaging`).
2. Extension service worker lifecycle is non-persistent; open ports and events affect lifetime, and disconnect handling is required (`developer.chrome.com/docs/extensions/develop/concepts/service-workers/lifecycle`).
3. Snap apps run with snap-specific environment/data paths; socket path selection should rely on explicit environment control (`snapcraft.io/docs/environment-variables`).

Inference from local behavior:
1. User logs show host starts, sends `HOST_READY`, then stdin ends immediately; this matches extension port disconnect sequence and is consistent with native host lifecycle handling.

## Proposed Solution

### Architecture decision

Implement provider compatibility inside Go host as a first-class dispatcher that mirrors Node host behavior while reusing existing service-worker primitive command contracts.

### Why this path

1. Lowest risk: service-worker primitives already work and are battle-tested.
2. Best parity: request/response semantics can be ported provider-by-provider.
3. Avoids dual maintenance of new extension endpoints.

### Proposed Go components

1. `go/internal/host/providers/` package
2. `dispatcher.go`: route provider tool requests to provider handlers.
3. `provider_common.go`: tab/cookie/CDP helper wrappers over `writeNative + pending` model.
4. `chatgpt.go`, `gemini.go`, `perplexity.go`, `grok.go`, `aistudio.go`: provider-specific orchestration.
5. `result_shape.go`: normalize provider outputs to Node-compatible result schemas.

### Router changes

1. Remove provider blocklist rejection in `toolmap.go`.
2. Keep mapping for provider tools and args.
3. For provider tools, dispatch through provider executor path in host runtime before generic forwarding.

### Pseudocode sketch

```go
if req.Type == "tool_request" && isProvider(req.Params.Tool) {
  result, err := providers.Dispatch(ctx, req, providerBridge)
  if err != nil {
    sendToolError(session, reqID, err.Error())
    return
  }
  sendToolResult(session, reqID, result)
  return
}
```

```go
type ProviderBridge interface {
  SendAndWait(msg map[string]any, timeout time.Duration) (map[string]any, error)
}

func (h *hostRuntime) SendAndWait(msg map[string]any, timeout time.Duration) (map[string]any, error) {
  // allocate id, register pending callback, writeNative, wait on channel or timeout
}
```

## Phased Implementation Plan

### Phase 0: Contract locking and fixtures

1. Generate/maintain provider inventory artifact (`scripts/provider-compat-inventory.cjs`).
2. Capture golden JSON outputs for Node provider commands in controlled scenarios.
3. Define parity acceptance table per command/field.

### Phase 1: Router and host scaffolding

1. Introduce provider dispatch hook in Go host runtime.
2. Keep current core behavior unchanged for non-provider tools.
3. Add provider tool request parsing without execution.

### Phase 2: Provider-by-provider parity

Order:
1. ChatGPT
2. Perplexity
3. Gemini
4. Grok
5. AI Studio query
6. AI Studio build

For each provider:
1. Implement minimal successful query flow.
2. Implement model/mode flags.
3. Implement timeout and error mapping.
4. Add parity tests vs captured Node fixtures.

### Phase 3: Output/CLI compatibility hardening

1. Ensure Go responses expose the same payload fields Node CLI expects.
2. Avoid tool-specific wrapper drift (especially `aistudio`/`aistudio.build`).
3. Add structured output snapshots (JSON + YAML views).

### Phase 4: Runtime and docs completion

1. Installer/runtime docs for `SURF_HOST_PROFILE`, `SURF_SOCKET_PATH`, Snap path.
2. Add troubleshooting matrix for disconnect causes.
3. Release checklist and migration notes.

## Testing And Validation Strategy

1. Unit tests: provider request mapping, arg coercion, result shaping.
2. Integration tests: host-runtime pending correlation with mocked extension responses.
3. End-to-end tests (manual + script harness): run same provider command via Node and Go profiles and diff normalized outputs.
4. Regression check: all non-provider core commands unchanged in `core-go`.

Suggested acceptance criteria:
1. For each provider tool, required fields match Node output contract.
2. Error cases map to user-actionable messages (login missing, model unavailable, timeout).
3. Snap Chromium path works with explicit `SURF_SOCKET_PATH` override.

## Alternatives Considered

1. Keep providers Node-only forever.
Reason rejected: persistent dual-runtime complexity and user confusion.

2. Rebuild provider flows entirely in extension service worker.
Reason rejected: increases MV3 complexity and makes extension heavier; host-side orchestration already exists and is portable.

3. Reimplement provider calls via official APIs only.
Reason rejected: violates current no-API-key/session-cookie product behavior and reduces parity.

## Risks And Mitigations

1. Provider DOM churn breaks selectors.
Mitigation: keep provider validation commands and selector smoke checks; add quick-fail diagnostics.

2. Output drift between Node and Go.
Mitigation: golden fixture diff tests and explicit contract matrix.

3. Timeout/retry behavior divergence.
Mitigation: centralize retry/backoff/timeouts in shared provider helper package.

4. Snap/socket path confusion.
Mitigation: preserve explicit install hints and add runtime self-check command.

## Open Questions

1. Should Go provider logic duplicate Node client logic in Go, or embed/reuse existing JS clients via runtime bridge for faster parity?
2. Should `aistudio` keep JSON-in-string `output` behavior for strict Node compatibility, or normalize both runtimes to structured payload now?
3. How strict should parity be for optional metadata fields (e.g., `thinkingTime`, `sources`, warnings arrays)?

## References

### Repository evidence

1. `native/host-helpers.cjs`
2. `native/host.cjs`
3. `src/native/port-manager.ts`
4. `src/service-worker/index.ts`
5. `go/internal/host/router/toolmap.go`
6. `go/internal/host/router/toolmap_test.go`
7. `go/cmd/surf-host-go/main.go`
8. `go/internal/host/nativeio/codec.go`
9. `go/internal/host/pending/store.go`
10. `go/internal/host/router/ingress.go`
11. `scripts/install-native-host.cjs`
12. `native/cli.cjs`
13. `native/chatgpt-client.cjs`
14. `native/gemini-client.cjs`
15. `native/perplexity-client.cjs`
16. `native/grok-client.cjs`
17. `native/aistudio-client.cjs`
18. `native/aistudio-build.cjs`
19. `ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/sources/01-provider-compat-inventory.json`

### External sources

1. https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging
2. https://developer.chrome.com/docs/extensions/develop/concepts/service-workers/lifecycle
3. https://developer.chrome.com/docs/extensions/reference/runtime
4. https://snapcraft.io/docs/environment-variables
