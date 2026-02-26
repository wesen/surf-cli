---
Title: Investigation Diary
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
    - Path: src/native/port-manager.ts
      Note: Disconnect/reconnect behavior evidence
    - Path: ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/scripts/provider-compat-inventory.cjs
      Note: Reproducibility script for inventory generation
    - Path: ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/sources/01-provider-compat-inventory.json
      Note: Generated inventory artifact
ExternalSources: []
Summary: Chronological log of all investigation commands, findings, fixes, and delivery steps for provider compatibility research.
LastUpdated: 2026-02-25T19:43:23-05:00
WhatFor: Preserve reproducible investigation history for continuation and review.
WhenToUse: Use when implementing provider support or auditing why design decisions were made.
---


# Investigation Diary

## Goal

Create an exhaustive, evidence-backed research package for Go-host provider compatibility (ChatGPT, Gemini, Perplexity, Grok, AI Studio), with complete ticket bookkeeping and reMarkable delivery.

## Chronological Log

### Phase 1 - Ticket initialization

Commands:

```bash
docmgr ticket create-ticket --ticket SURF-20260225-R3 --title "Go Host Provider Compatibility Research (ChatGPT/Gemini/Perplexity/Grok/AI Studio)" --topics go,native-messaging,architecture,migration,chatgpt,providers
docmgr doc add --ticket SURF-20260225-R3 --doc-type design-doc --title "Go Host Provider Compatibility: Exhaustive Architecture and Migration Research"
docmgr doc add --ticket SURF-20260225-R3 --doc-type reference --title "Investigation diary"
docmgr doc add --ticket SURF-20260225-R3 --doc-type reference --title "Provider Compatibility Matrix and Contracts"
```

Result:
1. Ticket workspace created at `ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio`.
2. Core ticket files and docs scaffolded.

### Phase 2 - Evidence sweep: Node provider command and host orchestration

Commands:

```bash
nl -ba native/host-helpers.cjs | sed -n '980,1140p'
nl -ba native/host.cjs | sed -n '430,760p'
nl -ba native/host.cjs | sed -n '760,1160p'
nl -ba native/host.cjs | sed -n '1320,1450p'
```

Findings:
1. Provider command mapping and args are explicit in `host-helpers.cjs`.
2. `host.cjs` orchestrates provider workflows and calls extension primitives (`GET_*_COOKIES`, `*_NEW_TAB`, `*_EVALUATE`, `*_CDP_COMMAND`).
3. `pendingToolRequests` correlates asynchronous extension roundtrips.
4. AI requests are serialized with queue and 2s pacing.

### Phase 3 - Evidence sweep: service worker provider primitive handlers

Commands:

```bash
nl -ba src/service-worker/index.ts | sed -n '2580,3060p'
rg -n "GET_CHATGPT_COOKIES|CHATGPT_NEW_TAB|PERPLEXITY_NEW_TAB|GROK_NEW_TAB|AISTUDIO_NEW_TAB|DOWNLOADS_SEARCH" src/service-worker/index.ts
```

Findings:
1. Service worker already implements the provider primitive API needed by host orchestration.
2. Commands are marked as `COMMANDS_WITHOUT_TAB`, so they bypass active-tab auto-resolution.

### Phase 4 - Evidence sweep: Go host behavior and intentional provider block

Commands:

```bash
nl -ba go/internal/host/router/toolmap.go | sed -n '1,260p'
nl -ba go/internal/host/router/toolmap.go | sed -n '260,660p'
nl -ba go/internal/host/router/toolmap_test.go | sed -n '1,120p'
nl -ba go/cmd/surf-host-go/main.go | sed -n '140,260p'
```

Findings:
1. `providerPrefixes` blocklist rejects provider tools in go-core.
2. `UnsupportedToolError` explicitly documents unsupported-go-core behavior.
3. Test suite enforces provider rejection path.
4. Go host runtime transport and pending correlation are implemented and stable for core tools.

### Phase 5 - Installer/runtime profile and Snap path analysis

Commands:

```bash
nl -ba scripts/install-native-host.cjs | sed -n '200,520p'
nl -ba README.md | sed -n '303,352p'
nl -ba README.md | sed -n '544,577p'
```

Findings:
1. Installer supports dual runtime wrapper (`node-full` default, `core-go` optional).
2. Linux Chromium Snap target installs an isolated runtime copy and writes a snap-specific socket-path hint.
3. README documents `SURF_HOST_PROFILE` and `SURF_SOCKET_PATH` behavior.

### Phase 6 - Reproducible compatibility inventory script

Actions:
1. Added script: `scripts/provider-compat-inventory.cjs` in ticket workspace.
2. Output file: `sources/01-provider-compat-inventory.json`.

Commands:

```bash
node ttmp/2026/02/25/SURF-20260225-R3--.../scripts/provider-compat-inventory.cjs
cat ttmp/2026/02/25/SURF-20260225-R3--.../sources/01-provider-compat-inventory.json
```

Tricky points/failures:
1. Initial repo-root detection failed from script directory.
Resolution: added upward directory walk `resolveRepoRoot()`.
2. Initial parser missed some `case` branches with `{` style.
Resolution: updated regex to support `case "...": {`.

Result:
1. JSON inventory produced with provider mappings, Go blocked prefixes, and service-worker provider handler set.

### Phase 7 - Provider-specific logic and dependency inventory

Commands:

```bash
cat package.json
nl -ba native/chatgpt-client.cjs | sed -n '1,320p'
nl -ba native/gemini-client.cjs | sed -n '1,360p'
nl -ba native/perplexity-client.cjs | sed -n '1,360p'
nl -ba native/grok-client.cjs | sed -n '1,420p'
nl -ba native/aistudio-client.cjs | sed -n '1,420p'
nl -ba native/aistudio-build.cjs | sed -n '1,420p'
```

Findings:
1. ChatGPT/Perplexity/Grok/AI Studio rely mostly on CDP + DOM automation via extension primitives.
2. Gemini client uses cookie-authenticated HTTPS flow.
3. Third-party package usage in provider path is minimal; runtime orchestration is custom.

### Phase 8 - Native messaging lifecycle evidence for disconnect behavior

Commands:

```bash
nl -ba src/native/port-manager.ts | sed -n '1,180p'
```

Findings:
1. Extension connects native host using `connectNative("surf.browser.host")`.
2. On disconnect it logs lastError and retries.
3. User-provided logs (`HOST_READY` then immediate stdin end) match extension port disconnect semantics.

### Phase 9 - External references

Commands:

```bash
# Searched and captured primary docs links:
# - developer.chrome.com native messaging
# - developer.chrome.com extension service worker lifecycle
# - developer.chrome.com runtime API
# - snapcraft environment variables
```

Findings:
1. External docs align with observed lifecycle and installation constraints.
2. No contradiction found with local code-level evidence.

## User-Reported Runtime Evidence (captured in thread)

Recorded user log sample:

```text
2026-02-25T21:44:18.507Z Host starting...
2026-02-25T21:44:18.511Z Host initialization complete, waiting for connections...
2026-02-25T21:44:18.511Z Socket server listening on /tmp/surf.sock
2026-02-25T21:44:18.512Z Sent HOST_READY to extension
2026-02-25T21:44:18.513Z stdin ended (extension disconnected), notifying clients
```

Interpretation:
1. Host startup path is healthy.
2. Disconnect happens after extension connection lifecycle event, not pre-start host crash.

## Final Artifacts Produced

1. Design doc: exhaustive architecture and migration report.
2. Reference matrix: provider contracts and parity requirements.
3. Sources artifact: provider inventory JSON.
4. Ticket bookkeeping updates (index/tasks/changelog/relations/doctor).
5. reMarkable bundle upload verification.

## Verification Checklist

1. Evidence anchored to file/line where possible.
2. Scripted inventory reproducible.
3. docmgr doctor run before delivery.
4. reMarkable dry-run + real upload + remote listing verification completed.

## Implementation Diary - ChatGPT Go Integration (2026-02-25 onward)

### Entry 1 - Task framing and baseline inspection (2026-02-25 19:5x EST)

Goal for this phase:
1. Implement only ChatGPT provider integration in Go host.
2. Execute task-by-task with tests and incremental commits.
3. Keep other providers blocked for now.

Commands run:

```bash
git status --short
rg -n "tool_request|MapToolToMessage|UnsupportedToolError|providerPrefixes|sendToolResult|sendToolError" go -S
wc -l native/chatgpt-client.cjs
nl -ba native/chatgpt-client.cjs | sed -n '1,260p'
nl -ba native/chatgpt-client.cjs | sed -n '260,620p'
nl -ba go/cmd/surf-host-go/main.go | sed -n '260,760p'
```

Observed:
1. Worktree already contains unrelated changes from previous sessions (left untouched).
2. `chatgpt-client.cjs` is 462 lines and is the reference flow for Go port.
3. Go host runtime has pending request correlation for client->extension but no internal provider subrequest bridge yet.
4. Router currently blocks all provider prefixes, so ChatGPT must be intercepted before `MapToolToMessage` or routing behavior must be changed.

Decision:
1. Keep router provider blocklist unchanged for non-ChatGPT providers.
2. Implement ChatGPT via dedicated provider dispatch in host runtime before router mapping.
3. Add an internal bridge in host runtime for provider subrequest roundtrips.

### Entry 2 - Task 1 completed: host provider bridge primitives (2026-02-25 20:0x EST)

Changes made:
1. Extended `hostRuntime` with `providerPending` map + mutex for internal provider subrequest correlation.
2. Added `requestNativeForProvider(ctx, msg, timeout)` helper to send extension messages and await correlated response.
3. Updated native message handling to resolve provider-pending IDs before normal client pending store.
4. Added `cloneMap` utility to avoid mutating caller maps.

Files changed:
1. `go/cmd/surf-host-go/main.go`
2. `ttmp/.../tasks.md`
3. `ttmp/.../reference/01-investigation-diary.md`

Commands run:

```bash
go test ./cmd/surf-host-go
```

Result:
1. Build/test pass for host command package.
2. Task 1 marked complete.
