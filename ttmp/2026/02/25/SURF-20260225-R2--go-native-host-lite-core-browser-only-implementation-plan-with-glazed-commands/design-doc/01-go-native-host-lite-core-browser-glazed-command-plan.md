---
Title: 'Go Native Host Lite: Core Browser + Glazed Command Plan'
Ticket: SURF-20260225-R2
Status: active
Topics:
    - go
    - native-messaging
    - chromium
    - architecture
    - glazed
    - migration
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: native/cli.cjs
      Note: |-
        Current local socket NDJSON request protocol
        Current socket NDJSON request protocol baseline
    - Path: native/host-helpers.cjs
      Note: |-
        Current tool-to-message mapping surface (core + provider-specific)
        Current core/provider tool mapping baseline
    - Path: native/host.cjs
      Note: |-
        Current Node native host runtime and socket bridge lifecycle
        Current host transport/lifecycle baseline
    - Path: native/protocol.cjs
      Note: |-
        Native Messaging length-prefixed envelope contract
        Length-prefixed native messaging protocol baseline
    - Path: src/native/port-manager.ts
      Note: |-
        Extension native messaging connect/reconnect and response correlation
        Extension native port contract baseline
    - Path: src/service-worker/index.ts
      Note: |-
        Extension message handler contract for browser actions
        Core message handler contract baseline
ExternalSources:
    - /home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/tutorials/05-build-first-command.md
Summary: Detailed implementation plan for replacing Node native-host core browser flows with a Go Host Lite and Glazed-based Go CLI command surface, explicitly excluding provider-specific site automation.
LastUpdated: 2026-02-25T17:33:00-05:00
WhatFor: Plan and execute a low-risk migration of core browser automation from Node host/CLI to Go
WhenToUse: Use when implementing Go Host Lite and prioritizing core browser operations over provider-specific site logic
---


# Go Native Host Lite: Core Browser + Glazed Command Plan

## Executive Summary

This plan defines a phased migration from the current Node native host to a Go-native host for **core browser automation only**. Provider-specific flows (ChatGPT, Gemini, Perplexity, Grok, AI Studio) are explicitly out of scope for the first Go release.

The target is:

1. Go Native Messaging host process (`surf-host-go`) with contract parity for core message handling.
2. Go CLI (`surf-go`) using **Glazed command authoring conventions** for structured output and stable command schemas.
3. Incremental rollout with dual-host support and side-by-side verification.

## Problem Statement

The current native host stack is Node-based and mixes core transport/routing responsibilities with provider-specific browser automation logic. For Linux Snap and constrained runtime environments, this increases packaging and runtime complexity.

Observed evidence:

1. Native host transport + socket bridge + request correlation are centralized in `native/host.cjs` (`writeMessage/processInput/server.listen`, `stdin end` lifecycle) at `native/host.cjs:1288-1668`.
2. Tool routing surface is broad in `mapToolToMessage` and mixes core actions with provider commands at `native/host-helpers.cjs:527-1125`.
3. Extension connects to host via `chrome.runtime.connectNative("surf.browser.host")` and expects `HOST_READY`/request-id semantics in `src/native/port-manager.ts:56-99`.
4. Service worker executes core browser messages via `handleMessage` switch in `src/service-worker/index.ts:313-2335`.

## Scope

### In scope (Go v1)

1. Native Messaging framing and lifecycle parity.
2. Local socket server protocol parity (`tool_request`, `stream_request`, `stream_stop`).
3. Core browser command routing and result forwarding.
4. Glazed-based Go CLI command groups for core actions.
5. Installer integration for selecting Go host runtime.

### Out of scope (Go v1)

1. Provider-specific commands and adapters:
   - `chatgpt`, `gemini`, `perplexity`, `grok`, `aistudio`, `aistudio.build`.
2. Host-side provider orchestration/retry heuristics.
3. Model selection and site-specific DOM automation.

## Current-State Architecture (Evidence-Based)

### Extension/native contract

1. Extension manages native port, reconnect behavior, and ID-based response matching (`src/native/port-manager.ts:30-39`, `:56-112`).
2. Host emits `HOST_READY` on startup (`native/host.cjs:1638-1643`).
3. Host consumes Native Messaging length-prefixed JSON (`native/protocol.cjs:1-27`, `native/host.cjs:1299-1306`).

### Core command surface

Core message mapping lives in `native/host-helpers.cjs`:

1. Interaction primitives (`EXECUTE_CLICK`, `EXECUTE_TYPE`, `EXECUTE_KEY`, scroll, drag) via `mapComputerAction` (`native/host-helpers.cjs:421-521`).
2. Page/read/eval/wait/network/tab/window/cookie/dialog/frame/emulation mappings (`native/host-helpers.cjs:531-1125`).
3. Provider-specific mappings are mixed into same switch (`native/host-helpers.cjs:1021-1101`).

### Service worker handlers

Core service worker handlers exist for:

1. execute actions (`EXECUTE_*`) and page reads (`READ_PAGE`, `GET_PAGE_TEXT`) at `src/service-worker/index.ts:325-992`.
2. tabs and registry flows (`TABS_*`, `SWITCH_TAB`) at `src/service-worker/index.ts:2055-2207`.
3. stream bridging (`STREAM_CONSOLE`, `STREAM_NETWORK`, `STREAM_STOP`) at `src/service-worker/index.ts:2239-2318`.

## Proposed Solution

## 1) Split responsibilities by profile

Define explicit runtime profiles:

1. `core` profile (Go Host Lite): only core browser routing.
2. `provider` profile (existing Node host): legacy provider-specific commands.

In v1, Go host rejects provider messages with explicit error payloads:

```json
{ "error": { "content": [{"type":"text","text":"Provider command not supported in go-core profile"}] } }
```

## 2) Introduce Go Host Lite modules

Suggested package layout:

```text
go/
  cmd/
    surf-host-go/main.go
    surf-go/main.go
  internal/host/
    nativeio/        # length-prefixed read/write
    socketbridge/    # NDJSON socket protocol (tool_request/stream_request)
    router/          # core message routing
    pending/         # request correlation and timeouts
    lifecycle/       # signal handling, cleanup
  internal/cli/
    commands/        # glazed commands by domain
    transport/       # socket client to host
```

## 3) Preserve extension contract first, improve later

v1 keeps protocol stability:

1. Same Native Messaging framing.
2. Same request/response IDs.
3. Same key message types.
4. Same disconnect semantics (`extension_disconnected` to socket clients on stdin EOF).

## 4) Build Go CLI with Glazed command authoring

Follow `glazed-command-authoring` conventions:

1. Each command struct embeds `*cmds.CommandDescription`.
2. Settings structs use `glazed:"..."` tags.
3. Decode with `vals.DecodeSectionInto(schema.DefaultSlug, settings)`.
4. Build with `cmds.NewCommandDescription + fields.New + settings.NewGlazedSchema + cli.NewCommandSettingsSection`.
5. Cobra wiring via `cli.BuildCobraCommandFromCommand(...)` with explicit parser middleware config.

Reference pattern: `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/tutorials/05-build-first-command.md`.

## Command Architecture (Glazed)

### Command groups (core-only)

1. `page`: `read`, `text`, `state`, `search`, `wait-*`
2. `input`: `click`, `type`, `key`, `scroll`, `hover`, `drag`, `select`
3. `tab`: `list`, `new`, `switch`, `close`, `name`, `named`
4. `network`: `list`, `get`, `body`, `origins`, `stats`, `clear`, `export`, `stream`
5. `console`: `read`, `stream`
6. `cookie`: `list`, `get`, `set`, `clear`
7. `window`: `new`, `list`, `focus`, `close`, `resize`
8. `frame`: `list`, `switch`, `main`, `eval`
9. `dialog`: `accept`, `dismiss`, `info`
10. `emulate`: `network`, `cpu`, `geo`, `device`, `viewport`, `touch`
11. `shot`: `capture` (screenshot)

### Glazed skeleton example

```go
type PageReadCommand struct {
    *cmds.CommandDescription
}

type PageReadSettings struct {
    Filter            string `glazed:"filter"`
    Depth             int    `glazed:"depth"`
    IncludeScreenshot bool   `glazed:"include-screenshot"`
    TabID             int    `glazed:"tab-id"`
}

func (c *PageReadCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    s := &PageReadSettings{}
    if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil { return err }

    req := ToolRequest{Type: "tool_request", Method: "execute_tool", Params: ToolParams{Tool: "read_page", Args: map[string]any{
        "filter": s.Filter,
        "depth": s.Depth,
        "includeScreenshot": s.IncludeScreenshot,
        "tabId": s.TabID,
    }}}

    resp, err := transport.Send(req)
    if err != nil { return err }

    return gp.AddRow(ctx, types.NewRow(types.MRP("result", resp)))
}
```

## Design Decisions

1. **Contract-first migration**: preserve extension and socket protocols before optimizing internals.
2. **Core-only boundary**: explicitly exclude provider commands from Go v1.
3. **Dual runtime period**: keep Node host available until Go core reaches parity.
4. **Glazed-first CLI model**: move away from ad-hoc argument parsing in `native/cli.cjs` to typed command settings.
5. **Incremental command porting**: port highest-value command groups first (`page`, `input`, `tab`, `network`).

## Alternatives Considered

1. Full parity Go rewrite including providers in one milestone.
   - Rejected for v1 due to high regression risk and long lead time.
2. Keep Node host and only patch installer/runtime behavior.
   - Rejected as long-term direction because it does not reduce runtime heterogeneity.
3. Replace extension contract while building Go host.
   - Rejected for v1 due to avoidable integration risk; extension contract is stable and already battle-tested.

## Implementation Plan

### Phase 0: Contract Inventory and Freeze

1. Enumerate all core tool mappings from `native/host-helpers.cjs`.
2. Classify each as `core-v1`, `provider`, or `defer`.
3. Write a machine-readable contract fixture set (message-in -> expected message-out).

### Phase 1: Go Host Transport Foundation

1. Implement Native Messaging length-prefix reader/writer.
2. Implement socket server for local CLI (line-delimited JSON).
3. Implement pending request map + stream map + ID correlation.
4. Implement lifecycle parity (`HOST_READY`, stdin end handling, SIGINT/SIGTERM cleanup).

### Phase 2: Core Router and Message Forwarding

1. Implement `tool_request` and `stream_request` ingress handlers.
2. Implement extension-bound message forwarding for core message set.
3. Implement result normalization parity (errors vs result payloads).
4. Add explicit unsupported-provider response behavior.

### Phase 3: Glazed CLI Bootstrap

1. Create root command and grouped subcommand roots.
2. Add glazed output + command settings sections globally.
3. Implement shared transport client with request timeout and stream handling.
4. Implement base command scaffolding pattern for all groups.

### Phase 4: Port Core Command Groups

1. `page` + `input` group commands.
2. `tab` + `window` + `frame` + `dialog` groups.
3. `network` + `console` groups including stream mode.
4. `cookie`, `emulate`, `shot` groups.

### Phase 5: Installer + Packaging + Rollout

1. Add Go binary discovery/install path support in native host installer.
2. Support host profile selection (`core-go`, fallback `node-full`).
3. Add migration docs and compatibility matrix.

### Phase 6: Verification and Cutover

1. Golden protocol tests (Node vs Go for core commands).
2. End-to-end browser tests in Chromium/Snap and non-Snap.
3. Controlled default switch to Go core host.

## Testing and Validation Strategy

1. Unit tests for framing, request correlation, and socket behavior.
2. Contract tests for command mapping parity.
3. Integration tests with mocked extension responses.
4. Manual browser validation checklist:
   - service worker connects and gets `HOST_READY`.
   - representative commands in each core group succeed.
   - extension reload emits client disconnect notification.
5. Installer validation for standard Linux + Snap chromium targets.

## Risks and Mitigations

1. **Risk**: Silent protocol drift.
   - Mitigation: fixed fixture-based contract tests and golden snapshots.
2. **Risk**: Incomplete command coverage in first release.
   - Mitigation: explicit profile gating and fallback to Node full host.
3. **Risk**: CLI usability regressions during migration.
   - Mitigation: keep command aliases compatible where possible, publish mapping table.

## Open Questions

1. Do we want `surf-go` as a separate binary or as the default `surf` command at cutover?
2. Should unsupported provider commands auto-forward to Node host when available?
3. How strict should we be on preserving exact textual error messages for backward compatibility?

## References

1. `native/protocol.cjs:1-27`
2. `native/host.cjs:1288-1668`
3. `native/host-helpers.cjs:421-521`
4. `native/host-helpers.cjs:527-1125`
5. `src/native/port-manager.ts:17-112`
6. `src/service-worker/index.ts:313-2335`
7. `native/cli.cjs:1970-2004`
8. `native/cli.cjs:2741-2867`
9. `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/tutorials/05-build-first-command.md`
