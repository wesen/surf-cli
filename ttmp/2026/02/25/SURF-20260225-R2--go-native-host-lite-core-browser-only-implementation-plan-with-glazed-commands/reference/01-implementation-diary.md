---
Title: Implementation Diary
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
    - Path: native/host.cjs
      Note: Host runtime behavior and lifecycle parity baseline
    - Path: native/host-helpers.cjs
      Note: Command/tool mapping inventory baseline
    - Path: src/native/port-manager.ts
      Note: Extension-native host connect/reconnect and request matching baseline
    - Path: src/service-worker/index.ts
      Note: Extension command handling baseline
ExternalSources:
    - /home/manuel/.codex/skills/glazed-command-authoring/SKILL.md
    - /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md
Summary: Chronological diary for creating the Go Host Lite + Glazed command implementation plan ticket and delivery bundle.
LastUpdated: 2026-02-25T17:39:00-05:00
WhatFor: Preserve exact planning context, evidence, and command trace for continuation
WhenToUse: Use when continuing implementation from this planning ticket
---

# Implementation Diary

## Goal

Create a new ticket containing a detailed implementation plan and granular task backlog to build a Go version of the core browser host/CLI stack, excluding provider-specific site logic, and publish it to reMarkable.

## Context

User request:

1. focus on core browser flows only,
2. create a new ticket with a detailed plan,
3. produce granular tasks,
4. use Glazed command authoring conventions,
5. upload the final plan bundle to reMarkable.

## Chronological Log

## Phase 1 - Skill selection and workflow setup

### Skills used

1. `glazed-command-authoring` for CLI command architecture and conventions.
2. `ticket-research-docmgr-remarkable` for ticket/docmgr/reMarkable delivery workflow.

### Commands run

```bash
sed -n '1,260p' /home/manuel/.codex/skills/glazed-command-authoring/SKILL.md
sed -n '1,260p' /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md
docmgr status --summary-only
```

### Result

Confirmed both workflows and established docmgr root health before creating a new ticket.

## Phase 2 - Evidence collection for core contract boundaries

### Commands run

```bash
rg -n "connectNative|HOST_READY|tool_request|stream_request|onDisconnect" src native -S
rg -n "function mapToolToMessage|case \"[a-zA-Z0-9._-]+\"" native/host-helpers.cjs
nl -ba src/native/port-manager.ts | sed -n '1,220p'
rg -n "async function handleMessage|switch \(message.type\)|case \"EXECUTE_|case \"READ_PAGE|case \"TABS_|case \"CHATGPT_|case \"AISTUDIO_" src/service-worker/index.ts -S
nl -ba src/service-worker/index.ts | sed -n '2160,2345p'
nl -ba native/protocol.cjs
nl -ba native/host.cjs | sed -n '1280,1668p'
nl -ba native/host-helpers.cjs | sed -n '380,1125p'
nl -ba native/cli.cjs | sed -n '1970,2004p'
nl -ba native/cli.cjs | sed -n '2741,2867p'
```

### Findings

1. Native messaging transport contract is stable and simple (length-prefixed JSON + `HOST_READY`).
2. Core and provider mappings are currently mixed in one switch in `mapToolToMessage`.
3. Service worker already has broad core-handler coverage for action/page/tab/network/stream flows.
4. CLI socket protocol is NDJSON with explicit `tool_request` and `stream_request` envelopes.

### Interpretation

A contract-preserving Go core host is feasible with lower risk if provider commands are explicitly excluded in v1.

## Phase 3 - New ticket and document scaffolding

### Commands run

```bash
docmgr ticket create-ticket --ticket SURF-20260225-R2 --title "Go Native Host Lite (Core Browser Only) Implementation Plan with Glazed Commands" --topics go,native-messaging,chromium,architecture,glazed,migration
docmgr doc add --ticket SURF-20260225-R2 --doc-type design-doc --title "Go Native Host Lite: Core Browser + Glazed Command Plan"
docmgr doc add --ticket SURF-20260225-R2 --doc-type reference --title "Implementation Diary"
```

### Result

Created workspace and primary docs under:

`ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands`

## Phase 4 - Authoring deliverables

### Files authored

1. `design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md`
2. `tasks.md`
3. `index.md`
4. `reference/01-implementation-diary.md` (this file)

### Content decisions

1. Core-only scope was made explicit and provider commands listed as out-of-scope.
2. Glazed conventions were translated into concrete Go command scaffolding requirements.
3. Backlog was broken into phased, granular tasks (`T0.x` through `T7.x`) to support stepwise execution.

## Quick Reference

### Core protocol anchors

1. Extension native connection and pending request handling:
   - `src/native/port-manager.ts:56-112`
2. Native host framing and lifecycle:
   - `native/protocol.cjs:1-27`
   - `native/host.cjs:1288-1668`
3. Core command mapping and provider boundaries:
   - `native/host-helpers.cjs:421-521`
   - `native/host-helpers.cjs:527-1125`
4. Service worker core handlers:
   - `src/service-worker/index.ts:313-2335`
5. CLI socket protocol shapes:
   - `native/cli.cjs:1970-2004`
   - `native/cli.cjs:2741-2867`

### Plan artifacts

1. Primary plan doc:
   - `design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md`
2. Granular tasks:
   - `tasks.md`

## Usage Examples

### Continue execution from this plan

1. Start with scope freeze tasks (`T0.x`) and contract fixture generation.
2. Implement transport foundation (`T1.x`) before any command-level work.
3. Add Glazed CLI skeleton (`T3.x`) before porting command groups (`T4.x`).
4. Keep provider commands rejected in Go profile until a separate provider plan is approved.

### Validate ticket health

```bash
docmgr doctor --ticket SURF-20260225-R2 --stale-after 30
```

## Related

1. `design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md`
2. `tasks.md`
3. `changelog.md`

## Phase 5 - Validation and reMarkable publication

### Commands run

```bash
docmgr doctor --ticket SURF-20260225-R2 --stale-after 30
docmgr vocab add --category topics --slug go --description "Go language tooling and runtime implementation"
docmgr vocab add --category topics --slug glazed --description "Glazed command authoring and structured CLI output"
docmgr vocab add --category topics --slug migration --description "Incremental transition from one runtime/architecture to another"
docmgr doctor --ticket SURF-20260225-R2 --stale-after 30
remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/index.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/01-implementation-diary.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/changelog.md \
  --name "SURF-20260225-R2-go-core-host-lite-plan" \
  --remote-dir "/ai/2026/02/25/SURF-20260225-R2" \
  --toc-depth 2
remarquee upload bundle \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/index.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/01-implementation-diary.md \
  ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/changelog.md \
  --name "SURF-20260225-R2-go-core-host-lite-plan" \
  --remote-dir "/ai/2026/02/25/SURF-20260225-R2" \
  --toc-depth 2
remarquee cloud ls /ai/2026/02/25/SURF-20260225-R2 --long --non-interactive
```

### Results

1. `docmgr doctor` passed after adding missing vocabulary entries (`go`, `glazed`, `migration`).
2. Dry-run upload passed.
3. Real upload succeeded:
   - `OK: uploaded SURF-20260225-R2-go-core-host-lite-plan.pdf -> /ai/2026/02/25/SURF-20260225-R2`
4. Cloud listing verification succeeded:
   - `[f] SURF-20260225-R2-go-core-host-lite-plan`

### Delivery status

1. Ticket plan and tasks are complete and published.
2. Implementation execution has not started in this ticket (planning-only milestone complete).

## Phase 6 - Phase 0 implementation artifacts completed

### Commands run

```bash
python - <<'PY'
# extracted mapToolToMessage / mapComputerAction case labels to sources/01-tool-inventory.json
PY
cat > sources/02-go-core-v1-classification.yaml
cat > sources/03-core-v1-envelope-contract.yaml
cat > sources/04-go-v1-unsupported-tools.json
cat > sources/05-core-request-fixtures.json
cat > sources/06-response-normalization-fixtures.json
docmgr doc add --ticket SURF-20260225-R2 --doc-type reference --title "Phase 0 Contract Inventory and Freeze"
```

### Outputs created

1. `sources/01-tool-inventory.json`
2. `sources/02-go-core-v1-classification.yaml`
3. `sources/03-core-v1-envelope-contract.yaml`
4. `sources/04-go-v1-unsupported-tools.json`
5. `sources/05-core-request-fixtures.json`
6. `sources/06-response-normalization-fixtures.json`
7. `reference/02-phase-0-contract-inventory-and-freeze.md`

### Results

1. Phase 0 tasks (`T0.1` through `T0.8`) are now backed by concrete artifacts.
2. `tasks.md` was updated to mark all Phase 0 items complete.
3. `index.md` was updated with a direct link to the Phase 0 freeze reference.

### Notes

1. The extracted inventory reports `132` `mapToolToMessage` case labels and `18` `mapComputerAction` labels.
2. Go v1 unsupported list was explicitly frozen into provider-specific and deferred sets to avoid ambiguous runtime behavior.

## Phase 7 - Go transport foundation scaffold implemented

### Commands run

```bash
mkdir -p go/cmd/surf-host-go go/cmd/surf-go \
  go/internal/host/nativeio go/internal/host/config go/internal/host/socketbridge \
  go/internal/host/router go/internal/host/pending go/internal/host/lifecycle \
  go/internal/cli/commands go/internal/cli/transport
# wrote go/go.mod, command mains, doc.go placeholders, codec and socket path packages
cd go && go test ./...
```

### Outputs created

1. `go/go.mod`
2. `go/cmd/surf-host-go/main.go`
3. `go/cmd/surf-go/main.go`
4. `go/internal/host/nativeio/codec.go`
5. `go/internal/host/nativeio/codec_test.go`
6. `go/internal/host/config/socket_path.go`
7. `go/internal/host/config/socket_path_test.go`
8. `go/internal/host/socketbridge/doc.go`
9. `go/internal/host/router/doc.go`
10. `go/internal/host/pending/doc.go`
11. `go/internal/host/lifecycle/doc.go`
12. `go/internal/cli/commands/doc.go`
13. `go/internal/cli/transport/doc.go`

### Results

1. Implemented length-prefixed Native Messaging framing read/write helpers with size limits and invalid JSON handling.
2. Added unit tests for round-trip framing, malformed JSON, oversized payloads, and EOF behavior.
3. Implemented socket path parity helper with `SURF_SOCKET_PATH` override and OS defaults (`/tmp/surf.sock`, `//./pipe/surf`).
4. `go test ./...` passed for the current scaffold and tests.
5. `tasks.md` Phase 1 items `T1.1` through `T1.5` were marked complete.

## Phase 8 - Completed host transport lifecycle foundation (T1.6-T1.12)

### Commands run

```bash
# wrote listener/session manager, pending/stream registries, and runtime wiring
cd go && gofmt -w ./cmd ./internal
cd go && go test ./...
```

### Outputs created

1. `go/internal/host/socketbridge/listener.go`
2. `go/internal/host/socketbridge/listener_unix.go`
3. `go/internal/host/socketbridge/listener_windows.go`
4. `go/internal/host/socketbridge/session.go`
5. `go/internal/host/socketbridge/listener_unix_test.go`
6. `go/internal/host/socketbridge/session_test.go`
7. `go/internal/host/pending/id_allocator.go`
8. `go/internal/host/pending/store.go`
9. `go/internal/host/pending/store_test.go`
10. `go/internal/host/router/stream_registry.go`
11. `go/internal/host/router/stream_registry_test.go`
12. `go/cmd/surf-host-go/main.go`

### Results

1. Implemented local IPC listener abstraction with Unix socket implementation and Windows pipe placeholder (`ErrWindowsPipeUnsupported`) for explicit behavior.
2. Implemented thread-safe socket session manager with extension disconnect notification broadcast behavior.
3. Implemented request ID allocator and pending-request correlation store.
4. Implemented stream registry (`streamId -> session`) with per-session teardown support.
5. Wired host runtime to:
   - start socket listener and emit `HOST_READY`,
   - forward socket requests to native messaging with generated IDs,
   - correlate native responses back to original socket sessions and IDs,
   - handle `STREAM_EVENT` / `STREAM_ERROR` forwarding,
   - notify socket clients on native stdin EOF (`extension_disconnected`),
   - handle SIGINT/SIGTERM with listener shutdown and unix socket cleanup.
6. `go test ./...` passed after formatting.
7. `tasks.md` was updated to mark `T1.6` through `T1.12` complete.

## Phase 9 - Implemented strict socket ingress validation (T2.1-T2.3)

### Commands run

```bash
# wrote router ingress validation parser + tests and wired it into host runtime
cd go && gofmt -w ./cmd ./internal
cd go && go test ./...
```

### Outputs created

1. `go/internal/host/router/ingress.go`
2. `go/internal/host/router/ingress_test.go`

### Files modified

1. `go/cmd/surf-host-go/main.go`

### Results

1. Added explicit request parsing and validation for socket ingress message types:
   - `tool_request` requires `method=execute_tool` and `params.tool`.
   - `stream_request` requires `streamType` in `{STREAM_CONSOLE, STREAM_NETWORK}`.
   - `stream_stop` requires `type=stream_stop`.
2. Wired validation into host runtime before request forwarding, returning structured CLI error lines on malformed requests.
3. `go test ./...` passed.
4. `tasks.md` updated to mark `T2.1`, `T2.2`, and `T2.3` complete.

## Phase 10 - Implemented core-v1 routing and response correlation runtime (T2.4-T2.10)

### Commands run

```bash
# wrote core tool mapping + unsupported guards and rewired runtime request handling
cd go && gofmt -w ./cmd ./internal
cd go && go test ./...
```

### Outputs created

1. `go/internal/host/router/toolmap.go`
2. `go/internal/host/router/toolmap_test.go`

### Files modified

1. `go/cmd/surf-host-go/main.go`
2. `go/internal/host/router/ingress.go`
3. `go/internal/host/router/ingress_test.go`
4. `go/internal/host/pending/store.go`

### Results

1. Added core-v1 tool routing in Go host for the browser primitives and aliases, including `page.*`, `click/type/key/hover/drag/scroll`, tabs/windows/frames/dialog, network/console/cookie/emulation, and utility aliases (`back`, `forward`, `zoom`, `tab.reload`).
2. Implemented explicit provider/deferred command rejection with v1 error text:
   - `Command '<tool>' is not supported in go-core profile`.
3. Added `computer` action mapping parity path with action-level translation (`CLICK_REF`, `EXECUTE_KEY_REPEAT`, `EXECUTE_SCROLL`, etc.).
4. Updated runtime to route `tool_request` through the mapper and to emit `tool_response` envelopes (error or result) back to socket clients.
5. Updated stream handling parity:
   - `stream_request` now allocates stream IDs host-side, forwards to extension, and immediately returns `stream_started`.
   - `STREAM_EVENT` and `STREAM_ERROR` are forwarded to stream-owning sockets.
6. Preserved passthrough flow for non-tool requests (`GET_AUTH`, `API_REQUEST`) through the pending correlation path with original IDs restored.
7. `go test ./...` passed.
8. `tasks.md` updated to mark `T2.4` through `T2.10` complete.

## Phase 11 - Implemented Glazed CLI skeleton and transport utilities (T3.1-T3.8)

### Skills applied

1. `glazed-command-authoring` for command struct/section/wiring conventions.

### Commands run

```bash
cd go && go mod tidy
cd go && go test ./...
cd go && go run ./cmd/surf-go --help
```

### Outputs created

1. `go/internal/cli/transport/client.go`
2. `go/internal/cli/transport/client_test.go`
3. `go/internal/cli/commands/base.go`
4. `go/internal/cli/commands/base_test.go`
5. `go/internal/cli/commands/format.go`
6. `go/internal/cli/commands/format_test.go`
7. `go/internal/cli/commands/tool_raw.go`

### Files modified

1. `go/cmd/surf-go/main.go`
2. `go/go.mod`
3. `go/go.sum`

### Results

1. Added `surf-go` root command with:
   - Glazed help-system wiring (`help.NewHelpSystem`, `help_cmd.SetupCobraRootCommand`),
   - logging section and root logger init (`logging.AddLoggingSectionToRootCommand`, `logging.InitLoggerFromCobra`).
2. Added first Glazed command (`tool-raw`) as a skeleton command surface:
   - includes `settings.NewGlazedSchema()` output section,
   - includes `cli.NewCommandSettingsSection()` command metadata/debug section,
   - decodes settings via `vals.DecodeSectionInto(schema.DefaultSlug, settings)`.
3. Implemented shared socket transport client utility with request timeout and one-response parsing.
4. Implemented shared helper for constructing `tool_request` envelopes.
5. Implemented shared response-to-row formatter utility for Glazed output.
6. Added unit tests for base envelope builder, formatter extraction, and transport roundtrip.
7. Full package tests pass and `surf-go --help` renders correctly.
8. `tasks.md` updated to mark `T3.1` through `T3.8` complete.

## Phase 12 - Implemented Page + Input command wrappers (T4.1-T4.16)

### Commands run

```bash
cd go && gofmt -w ./cmd ./internal
cd go && go test ./...
cd go && go run ./cmd/surf-go --help
```

### Outputs created

1. `go/internal/cli/commands/tool_simple.go`

### Files modified

1. `go/cmd/surf-go/main.go`

### Results

1. Added reusable `SimpleToolCommand` Glazed wrapper that:
   - decodes common transport flags,
   - accepts command-specific args through `--args-json`,
   - merges command defaults,
   - sends `tool_request` via shared transport and emits normalized rows.
2. Registered first Phase 4 command surfaces in `surf-go`:
   - `page read|text|state|search`
   - `wait element|url|network|dom`
   - `click`, `type`, `key`, `scroll`, `hover`, `drag`, `select`, `screenshot`
3. Verified CLI wiring by running `surf-go --help` and ensuring commands are visible.
4. `go test ./...` passed.
5. `tasks.md` updated to mark `T4.1` through `T4.16` complete.

## Phase 13 - Implemented remaining core command groups + stream commands (T4.17-T4.34)

### Commands run

```bash
cd go && gofmt -w ./cmd ./internal
cd go && go test ./...
cd go && go run ./cmd/surf-go --help
cd go && go run ./cmd/surf-go network stream --help
```

### Outputs created

1. `go/internal/cli/commands/stream_simple.go`

### Files modified

1. `go/internal/cli/transport/client.go`
2. `go/internal/cli/transport/client_test.go`
3. `go/cmd/surf-go/main.go`

### Results

1. Extended transport with stream support (`Client.Stream`) including:
   - `stream_request` send,
   - `stream_started` handshake handling,
   - event loop callback,
   - `stream_stop` send on context cancellation/timeout.
2. Added reusable Glazed stream command wrapper (`StreamCommand`) with duration/options settings.
3. Wired remaining command groups in `surf-go`:
   - `tab` (`list/new/switch/close/name/named`)
   - `window` (`list/new/focus/close/resize`)
   - `frame` (`list/switch/main/eval`)
   - `dialog` (`accept/dismiss/info`)
   - `network` (`list/get/body/origins/stats/clear/export/stream`)
   - `console` (`read/stream`)
   - `cookie` (`list/get/set/clear`)
   - `emulate` (`network/cpu/geo/device/viewport/touch`)
4. Added transport stream unit test.
5. Verified command surfacing via help output.
6. `go test ./...` passed.
7. `tasks.md` updated to mark `T4.17` through `T4.34` complete.

## Phase 14 - Installer/profile packaging updates for Go host rollout (T5.1-T5.7)

### Commands run

```bash
node scripts/install-native-host.cjs --help
node scripts/uninstall-native-host.cjs --help
node scripts/build-go-host-binaries.cjs
cd go && go test ./...
node --check scripts/install-native-host.cjs
node --check scripts/uninstall-native-host.cjs
node --check scripts/build-go-host-binaries.cjs
```

### Outputs created

1. `scripts/build-go-host-binaries.cjs`

### Files modified

1. `scripts/install-native-host.cjs`
2. `scripts/uninstall-native-host.cjs`
3. `package.json`
4. `README.md`

### Results

1. Added cross-platform Go host build target script (`build-go-host-binaries.cjs`) building:
   - `linux/$GOARCH`,
   - `darwin/$GOARCH`,
   - `windows/$GOARCH`.
2. Installer now supports optional Go host runtime build/install:
   - detects Go toolchain via `SURF_GO_PATH` or PATH,
   - builds `surf-host-go` into wrapper dirs when source/toolchain are available,
   - wraps runtime with profile selection.
3. Added runtime profile switch in generated wrappers:
   - `SURF_HOST_PROFILE=node-full` (default fallback),
   - `SURF_HOST_PROFILE=core-go` (prefer `surf-host-go` when installed).
4. Preserved node fallback selection path when Go host is unavailable or fails to build.
5. Added Snap-aware parity for Go host wrapper:
   - builds snap-local `surf-host-go` in snap wrapper dir,
   - keeps snap socket export behavior intact.
6. Uninstaller now removes Go host artifacts (`surf-host-go`, `surf-host-go.exe`) from wrapper dirs.
7. README updated with:
   - `SURF_HOST_PROFILE` and `SURF_GO_PATH`,
   - build command `npm run build:go-host`,
   - migration/fallback usage notes for Linux/Snap.
8. `tasks.md` updated to mark `T5.1` through `T5.7` complete.

## Phase 15 - Added verification suite and manual validation pack (T6.1-T6.7 complete)

### Commands run

```bash
cd go && go test ./...
node scripts/tests/native-host-installer-smoke.cjs
```

### Outputs created

1. `go/internal/host/router/toolmap_contract_test.go`
2. `go/cmd/surf-go/integration_test.go`
3. `scripts/tests/native-host-installer-smoke.cjs`
4. `ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/03-manual-browser-validation-checklist.md`

### Files modified

1. `go/internal/cli/transport/client.go`
2. `go/internal/cli/transport/client_test.go`
3. `package.json`
4. `ttmp/.../index.md`

### Results

1. Confirmed framing tests, socket transport tests, pending correlation tests are in place (`T6.1`-`T6.3`).
2. Added core mapping contract tests for representative tool routes against frozen parity expectations (`T6.4`).
3. Added `surf-go` CLI integration tests with a mock unix socket server:
   - `page read` request/response path,
   - `network stream` start/stop lifecycle (`T6.5`, `T6.6`).
4. Added installer smoke test covering Linux standard + Snap targets and uninstall cleanup (`T6.7`).
5. Added a dedicated manual real-browser checklist doc for `T6.8` execution.
6. `tasks.md` updated: `T6.1`-`T6.7` complete, `T6.8` pending user-run browser validation.

## Phase 16 - Marked rollout opt-in milestone complete (T7.1)

### Result

1. `SURF_HOST_PROFILE=core-go|node-full` runtime switch is implemented and documented.
2. Go host is now available as opt-in beta profile (`T7.1` complete).
3. Remaining rollout tasks (`T7.2`-`T7.4`) depend on real-browser regression feedback and rollout decisions.
