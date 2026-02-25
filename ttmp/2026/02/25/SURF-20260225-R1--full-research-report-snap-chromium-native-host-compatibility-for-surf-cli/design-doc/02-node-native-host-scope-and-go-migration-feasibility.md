---
Title: Node Native Host Scope and Go Migration Feasibility
Ticket: SURF-20260225-R1
Status: active
Topics:
    - linux
    - chromium
    - snap
    - native-messaging
    - debugging
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: native/host.cjs
      Note: Current native host runtime responsibilities and lifecycle behavior
    - Path: native/socket-path.cjs
      Note: Cross-runtime socket path override contract
    - Path: scripts/install-native-host.cjs
      Note: Snap runtime staging and node binary copy logic
ExternalSources: []
Summary: Evidence-based assessment of how much the Node native host does today and a phased plan for replacing Node runtime dependency with a Go implementation.
LastUpdated: 2026-02-25T17:12:00-05:00
WhatFor: Decide whether to keep Node runtime packaging or invest in a Go host migration
WhenToUse: Use when planning runtime simplification and Linux Snap compatibility improvements
---


# Node Native Host Scope and Go Migration Feasibility

## Executive Summary

The current Node native host is not a thin transport shim. It is a large runtime controller (~1,668 LOC in `native/host.cjs`) that multiplexes CLI commands, extension messages, streaming channels, screenshot/file workflows, and multiple AI provider automation clients.

This means a full Go rewrite is feasible but non-trivial. The practical path is a phased migration:

1. isolate and lock protocol contracts,
2. build a minimal Go native host for core browser actions,
3. keep advanced AI-provider automation behind a compatibility boundary until feature parity is achieved.

If the immediate goal is to avoid requiring a user-managed Node install on Linux/Snap, we can solve that sooner with packaging/runtime changes than with a full host rewrite.

## Problem Statement and Scope

### Problem

Current Linux Snap troubleshooting raised the question: can we remove the Node runtime dependency entirely by replacing the native host with Go?

### Scope of this report

1. quantify what the current Node host does,
2. identify which responsibilities are mandatory for extension connectivity,
3. estimate migration effort/risk for a Go replacement,
4. propose an implementation path that reduces risk while improving deployability.

## Current-State Architecture (Evidence-Based)

### Native host responsibilities today

Observed in `native/host.cjs`:

1. Native Messaging stdio framing + lifecycle handling.
   - Writes/reads length-prefixed messages (`writeMessage`, `processInput`) and emits `HOST_READY` after server start.
   - Evidence: `native/host.cjs:1288-1295`, `native/host.cjs:1495-1500`, `native/host.cjs:1638-1643`.

2. Local socket server for CLI clients.
   - Creates a local `net` server on configured socket path and bridges CLI JSON-lines requests to extension messages.
   - Evidence: `native/host.cjs:1530-1636`, `native/host.cjs:1638-1643`.

3. Tool router and request correlation.
   - Maps CLI tool requests to extension actions and tracks pending requests/streams.
   - Evidence: `native/host.cjs:300-399`, `native/host.cjs:1366-1488`.

4. AI-provider workflow orchestration (not just API calls).
   - Integrates ChatGPT, Gemini, Perplexity, Grok, AI Studio clients with cookie retrieval, tab lifecycle, CDP evaluate/command flows.
   - Evidence: imports at `native/host.cjs:10-15`; handlers begin at `native/host.cjs:399`, `:454`, `:555`, `:649`, `:722`, `:842`, `:919`, `:1034`.

5. Screenshot and file post-processing.
   - Saves base64 screenshots, runs resize via `sips` / ImageMagick, cleans temp artifacts, attaches autoscreenshot metadata.
   - Evidence: `native/host.cjs:23-57`, `native/host.cjs:1376-1448`.

6. Disconnect semantics.
   - On extension disconnect (`stdin end`), notifies connected CLI clients and exits.
   - Evidence: `native/host.cjs:1506-1520`.

### Installer/runtime behavior relevant to Node dependency

Observed in `scripts/install-native-host.cjs`:

1. For Linux Chromium Snap target, installer stages a runtime under snap-accessible paths and copies Node there.
   - `prepareSnapRuntime` copies package + Node binary.
   - Evidence: `scripts/install-native-host.cjs:134-171`, `:281-297`.

2. Wrapper exports `SURF_SOCKET_PATH` and launches Node + host script.
   - Evidence: `scripts/install-native-host.cjs:173-199`.

3. Socket path is centrally configurable via `SURF_SOCKET_PATH`.
   - Evidence: `native/socket-path.cjs:1-15`.

## Gap Analysis for a Go Replacement

### What must exist for baseline functionality

A minimum viable host must provide:

1. Native Messaging framing and correct process lifecycle with Chromium,
2. local socket listener for `surf` CLI,
3. request/response correlation between CLI and extension,
4. parity for core browser tools currently routed by `mapToolToMessage`.

### What makes full parity expensive

1. AI-provider automation is deeply integrated with extension CDP/cookie flows.
2. Existing Node modules encapsulate provider-specific brittle logic and recovery heuristics.
3. Host currently mixes transport responsibilities with high-level product workflows.

Conclusion: “rewrite host in Go” can mean either:

1. **Go transport host (small/medium effort)**, or
2. **full feature-parity Go runtime (large effort)**.

## Proposed Solution

### Recommendation

Adopt a two-track plan:

1. **Short term:** keep Node behavior, but harden packaging so end users do not need to manually manage Node in problematic environments (Snap).
2. **Medium term:** implement **Go Host Lite** for core transport + browser tools.
3. **Long term:** port provider workflows incrementally or keep them in a Node sidecar behind a stable RPC boundary.

### Target architecture

1. `surf-host-go` (new binary): Native Messaging process + local socket + core tool routing.
2. Optional `surf-provider-node` sidecar (temporary): advanced provider workflows until Go parity.
3. Shared protocol contract tests for framing and message envelopes.

### API sketch

```text
Extension <-> NativeMessaging (len-prefixed JSON) <-> surf-host-go <-> Unix socket / named pipe <-> surf CLI
                                                            |
                                                            +-> optional provider sidecar RPC
```

### Pseudocode (Go Host Lite)

```go
func main() {
  cfg := loadConfigFromEnv() // socket path, log path
  ext := NewNativeMessagingEndpoint(os.Stdin, os.Stdout)
  cli := NewSocketServer(cfg.SocketPath)

  go cli.AcceptLoop(func(req CLIMsg, client net.Conn) {
    id := correlate(req, client)
    ext.Send(mapToExtension(req, id))
  })

  for {
    msg, err := ext.Read()
    if err == io.EOF {
      notifyClientsExtensionDisconnected()
      os.Exit(0)
    }
    routeExtensionReply(msg)
  }
}
```

## Implementation Plan

### Phase 1: Contract extraction and test harness

1. Freeze message envelope schema and required message types.
2. Add protocol fixtures from current Node behavior.
3. Add integration tests that simulate extension <-> host <-> CLI loops.

### Phase 2: Go Host Lite MVP

1. Implement Native Messaging framing.
2. Implement socket server + pending-request correlation.
3. Implement core browser action routing and response handling.
4. Match disconnect behavior (`stdin end` => notify clients + exit).

### Phase 3: Provider strategy

1. Inventory provider features by usage frequency.
2. Decide per provider: Go port now vs sidecar.
3. If sidecar: define strict JSON-RPC contract and timeout/error semantics.

### Phase 4: Packaging and migration

1. Ship binary via npm package assets or GitHub releases.
2. Update installer to prefer Go binary and fallback to Node host where needed.
3. Deprecate Node host gradually after parity and soak tests.

## Testing and Validation Strategy

1. Unit tests for Native Messaging framing and socket-path resolution.
2. Golden tests for request/response envelopes across Node and Go hosts.
3. Linux Snap integration test matrix:
   - native host process startup,
   - extension connect/reload behavior,
   - CLI command round-trips,
   - socket-path mismatch diagnostics.
4. Real-browser manual verification checklist (service worker logs + host logs + CLI command success).

## Effort and Risk Estimate

1. Go Host Lite (core transport + common actions): medium effort.
2. Full parity with all provider workflows: high effort.
3. Primary risk: hidden behavior coupling currently embedded in Node provider clients.
4. Secondary risk: migration regressions in request correlation, stream handling, and disconnect semantics.

## Alternatives Considered

1. Keep full Node host indefinitely and only patch installer/runtime behavior.
   - Pros: fastest, lowest regression risk.
   - Cons: retains Node runtime complexity long-term.

2. Full clean-slate Go rewrite in one step.
   - Pros: single runtime after completion.
   - Cons: highest delivery risk and longest time to stable parity.

3. Hybrid migration (recommended).
   - Pros: reduces immediate user pain while controlling risk.
   - Cons: temporary dual-runtime complexity.

## Open Questions

1. Which provider flows are actually required for first Go-host release?
2. Should provider automation be a permanent pluggable sidecar model?
3. Do we want an official “core-only” host profile for constrained environments?

## References

1. `native/host.cjs:1-21`
2. `native/host.cjs:300-399`
3. `native/host.cjs:399-1034`
4. `native/host.cjs:1366-1520`
5. `native/host.cjs:1530-1643`
6. `scripts/install-native-host.cjs:134-171`
7. `scripts/install-native-host.cjs:271-297`
8. `scripts/install-native-host.cjs:173-199`
9. `native/socket-path.cjs:1-15`
