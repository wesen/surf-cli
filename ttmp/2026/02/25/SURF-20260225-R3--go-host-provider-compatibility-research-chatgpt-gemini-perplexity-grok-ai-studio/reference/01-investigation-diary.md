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
LastUpdated: 2026-04-07T22:05:00-04:00
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

### Entry 3 - Tasks 2/4 completed: `chatgpt.go` provider module + unit tests (2026-02-25 20:1x EST)

Implementation:
1. Added new package `go/internal/host/providers`.
2. Added `chatgpt.go` with:
   - request parsing (`query`, `model`, `with-page/withPage`, `timeout`, `file`),
   - provider orchestration (`GET_CHATGPT_COOKIES`, optional `GET_PAGE_TEXT`, `CHATGPT_NEW_TAB`, `CHATGPT_EVALUATE`, `CHATGPT_CDP_COMMAND`, `CHATGPT_CLOSE_TAB`),
   - page/login/prompt/send/response wait flow,
   - structured return payload (`response`, `model`, `tookMs`).
3. Added `chatgpt_test.go` with mocked native caller tests, including `with-page` prompt context path.

Initial failure encountered:

```bash
go test ./internal/host/providers
```

Error:
1. `non-constant format string in call to fmt.Errorf` at multiple lines.

Fix:
1. Replaced `fmt.Errorf(variable)` with `errors.New(variable)` in provider module.

Validation commands:

```bash
go test ./internal/host/providers
gofmt -w go/internal/host/providers/chatgpt.go go/internal/host/providers/chatgpt_test.go
go test ./internal/host/providers
```

Result:
1. Provider package tests pass.

### Entry 4 - Tasks 3/5/6 completed: host integration + dispatch tests + full go test (2026-02-25 20:2x EST)

Implementation:
1. Integrated ChatGPT dispatch in `go/cmd/surf-host-go/main.go` before router mapping.
2. Added injectable `runChatGPTTool` function in `hostRuntime` for testability.
3. Added `providerNativeCaller` adapter using `requestNativeForProvider` bridge.
4. Added `cmd/surf-host-go/main_test.go` to verify:
   - `chatgpt` uses provider runner and returns tool response,
   - provider errors are surfaced as tool errors,
   - other providers (example: `gemini`) remain blocked by router behavior.

Issue encountered during tests:

```bash
go test ./cmd/surf-host-go ./internal/host/providers ./internal/host/router
```

Behavior:
1. Test process hung (no output).

Root cause:
1. `net.Pipe` write path blocked because `handleSessionLine` was called synchronously and response read happened only after call return.

Fix:
1. Updated tests to call `handleSessionLine` in goroutines and read from the paired connection concurrently.

Validation commands:

```bash
gofmt -w go/cmd/surf-host-go/main.go go/cmd/surf-host-go/main_test.go go/internal/host/providers/chatgpt.go go/internal/host/providers/chatgpt_test.go
go test ./cmd/surf-host-go ./internal/host/providers ./internal/host/router
go test ./...
```

Result:
1. Targeted suites pass.
2. Full Go module test suite passes.
3. Tasks 2-6 marked complete.

### Entry 5 - Task 7 validation attempt and environment handoff (2026-02-25 20:3x EST)

Goal:
1. Validate ChatGPT integration against a live browser-connected host socket.

Commands run:

```bash
ls -l /tmp/surf.sock ~/snap/chromium/common/surf-cli/surf.sock
go run ./cmd/surf-go tool-raw --tool chatgpt --args-json '{"query":"say ping"}' --socket-path /tmp/surf.sock
go run ./cmd/surf-go tool-raw --tool chatgpt --args-json '{"query":"say ping"}' --socket-path /home/manuel/snap/chromium/common/surf-cli/surf.sock
go run ./cmd/surf-go tool-raw --tool gemini --args-json '{"query":"say ping"}' --socket-path /home/manuel/snap/chromium/common/surf-cli/surf.sock
```

Results:
1. `/tmp/surf.sock` refused connection (`connect: connection refused`).
2. Snap socket responded to `chatgpt` with `ChatGPT login required`.
3. Same snap socket successfully executed `gemini` end-to-end.

Interpretation:
1. Live socket currently appears to be served by Node runtime profile (or otherwise not strictly `core-go`-limited), because `gemini` succeeds.
2. This environment cannot conclusively verify the new Go-only ChatGPT integration path without forcing browser host profile to `core-go` and reconnecting extension.

Handoff requirement:
1. Browser-side verification by user is required to confirm ChatGPT query flow in real extension-native-messaging runtime under `SURF_HOST_PROFILE=core-go`.

### Entry 6 - Status checkpoint

1. Tasks 1-6 and 8 are complete.
2. Task 7 remains pending user-side browser validation under confirmed `core-go` runtime profile.

### Entry 7 - ChatGPT file upload implementation (2026-02-26 20:0x EST)

Goal:
1. Implement `chatgpt --file` support end-to-end for Node and Go runtimes.
2. Keep behavior aligned by reusing existing extension `UPLOAD_FILE` primitive.

Discovery commands:

```bash
rg -n "File upload not yet implemented|UPLOAD_FILE|GET_FILE_INPUT_SELECTOR|chatgpt" native src go/internal/host/providers go/cmd/surf-host-go
sed -n '1570,1705p' src/service-worker/index.ts
sed -n '400,520p' native/chatgpt-client.cjs
sed -n '430,560p' native/host.cjs
sed -n '1,280p' go/internal/host/providers/chatgpt.go
```

Findings:
1. Both Node and Go ChatGPT paths explicitly returned `File upload not yet implemented`.
2. Service worker already had `UPLOAD_FILE` backed by CDP `setFileInputBySelector`, but required `ref`.
3. Provider automation lacked a stable `ref`, so selector-based upload support was the cleanest bridge.

Implementation changes:
1. `src/service-worker/index.ts`
   - Extended `UPLOAD_FILE` to accept either:
     - `selector` (new path), or
     - `ref` (legacy path via content-script `GET_FILE_INPUT_SELECTOR`).
2. `native/chatgpt-client.cjs`
   - Added file list normalization.
   - Added file input discovery (`waitForFileInputSelector`) that tags discovered input with `data-surf-file-input-id`.
   - Added upload orchestration (`uploadChatGPTFiles`) and integrated into `query()`.
3. `native/host.cjs`
   - Added `uploadFile(tabId, selector, files)` callback for ChatGPT query flow, sending `UPLOAD_FILE`.
4. `go/internal/host/providers/chatgpt.go`
   - Replaced placeholder error with real upload flow:
     - split file list,
     - find file input selector in ChatGPT tab,
     - call extension `UPLOAD_FILE`.

Test additions:
1. New Node unit test file: `test/unit/chatgpt-client.test.ts`.
   - Verifies upload call is issued with selector + file list.
   - Verifies upload error propagation.
2. Added Go unit test: `TestHandleChatGPTToolWithFileUpload` in `go/internal/host/providers/chatgpt_test.go`.

Validation commands and outcomes:

```bash
cd go && go test ./internal/host/providers ./cmd/surf-host-go
npm test -- test/unit/chatgpt-client.test.ts
cd go && go test ./...
```

Result:
1. All listed Go tests passed.
2. New Node ChatGPT unit tests passed (3/3).

Additional check attempted:

```bash
npx biome check native/chatgpt-client.cjs native/host.cjs src/service-worker/index.ts test/unit/chatgpt-client.test.ts
```

Outcome:
1. Failed due existing repository Biome config/version mismatch (`biome.json` schema 2.3.11 vs CLI 2.4.4) and unrelated rule key compatibility.
2. No code changes made for this; issue pre-exists in repo tooling config.

### Entry 8 - Snap Chromium runtime diagnosis, polling fixes, and Go-host confirmation (2026-04-07 evening EDT / 2026-04-08 host-log timestamps)

Goal:
1. Confirm whether Chromium Snap was actually launching `surf-host-go` or silently falling back to Node.
2. Make ChatGPT list-models/query debugging observable in both the service worker and Go host.
3. Fix the Go ChatGPT response poller so it recognizes the rendered assistant response instead of looping on empty text.

Commands and evidence gathered:

```bash
cat ~/snap/chromium/common/chromium/NativeMessagingHosts/surf.browser.host.json
cat ~/snap/chromium/common/surf-cli/host-wrapper.sh
ls -l ~/snap/chromium/common/surf-cli/
snap run --shell chromium -c 'tail -n 120 /tmp/surf-host-go.log'
go run ./cmd/surf-go chatgpt hello --debug-socket
go run ./cmd/surf-go tool-raw --tool chatgpt --args-json '{"list-models":true}'
```

Service worker evidence added during this phase:
1. Native-host request lifecycle logs (`Handling native host request`, `Native host handler resolved`, `Sent response to native host`).
2. `HOST_READY` runtime diagnostics:
   - `runtime: "go-host" | "node-host"`
   - `socketPath`
3. `CHATGPT_EVALUATE` detail summaries for model parsing and response polling.

Host-side findings:
1. The snap manifest correctly pointed to `~/snap/chromium/common/surf-cli/host-wrapper.sh`.
2. The wrapper defaulted to `profile=core-go` but only used Go if `~/snap/chromium/common/surf-cli/surf-host-go` existed and was executable.
3. Before reinstall, the wrapper fell back to:
   - copied Node binary + `runtime/surf-cli/native/host.cjs`
4. After reinstall/copy, service worker `HOST_READY` explicitly showed:
   - `runtime: "go-host"`
   - `socketPath: "/home/manuel/snap/chromium/common/surf-cli/surf.sock"`

Polling failure evidence:
1. Initial Go-host logs showed:
   - `waitForResponse poll=N len=0 stop=true/false finished=false ...`
2. That proved transport was healthy but the DOM extractor never found the assistant response text.
3. User-provided DOM inspection showed the actual response under:
   - `div[data-message-author-role="assistant"]`
   - inner `.markdown`
   - `<p>...</p>`

Code changes made during this investigation:
1. `go/internal/host/providers/chatgpt.go`
   - switched response extraction to prefer direct assistant nodes (`[data-message-author-role="assistant"], [data-turn="assistant"]`)
   - added assistant node summaries (`textLength`, `textPreview`, `hasMarkdown`, `messageId`)
   - ported/stabilized Node-like completion heuristics
2. `src/native/port-manager.ts`
   - added native-host request/response lifecycle logs
   - added nested result summaries for `CHATGPT_EVALUATE`
3. `native/host.cjs` and `go/cmd/surf-host-go/main.go`
   - added runtime/socket handshake metadata

What worked:
1. `surf-go chatgpt` and `tool-raw --tool chatgpt` eventually returned valid responses through the Go host.
2. `--list-models` resumed reporting canonical IDs instead of UI labels after model parser fixes.
3. Snap-private host logs became accessible through:
   - `snap run --shell chromium -c 'tail -n 120 /tmp/surf-host-go.log'`

What didn't work:
1. Host logs were initially "missing" from `/tmp` in the normal shell.
   - Root cause: Snap uses a private `/tmp`, so the log file only existed inside the Chromium snap namespace.
2. Model-list parsing initially regressed to UI labels such as:
   - `AutoDecides how long to think`
   - `InstantAnswers right away`
   instead of canonical IDs.
   - Root cause: dropdown parsing used visible labels without restoring canonical `gpt-5-2-*` extraction.
3. Legacy submenu enumeration initially failed because submenu items were portal-rendered outside the first menu container.

Outcome:
1. Go-host runtime under Snap Chromium is now confirmed and observable in the service worker.
2. ChatGPT provider response polling now finds rendered assistant content instead of timing out on `len=0`.
3. The remaining operational issue after this entry was local `Ctrl-C` cancellation, which is recorded separately in the rollout ticket.
