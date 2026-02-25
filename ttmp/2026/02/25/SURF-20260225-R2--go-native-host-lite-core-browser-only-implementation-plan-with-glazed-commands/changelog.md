# Changelog

## 2026-02-25

- Initial workspace created.
- Added primary design plan for Go Host Lite core-browser migration with Glazed command architecture.
- Added granular phased implementation backlog (`T0.x` through `T7.x`).
- Added implementation diary with command-level investigation record.
- Prepared ticket for reMarkable publication.

## 2026-02-25 - Added Go Host Lite plan and granular implementation backlog

Authored a core-browser-only Go Host Lite implementation plan using Glazed command authoring conventions, added a phase-by-phase granular task list, and documented planning evidence in the implementation diary.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md — Detailed architecture and phased implementation plan
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/01-implementation-diary.md — Chronological command and evidence diary
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md — Granular execution backlog


## 2026-02-25 - Published Go Host Lite plan bundle to reMarkable

Validated ticket docs with docmgr doctor, resolved vocabulary warnings, and uploaded the full planning bundle to /ai/2026/02/25/SURF-20260225-R2 with cloud listing verification.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/design-doc/01-go-native-host-lite-core-browser-glazed-command-plan.md — Included in published bundle
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/index.md — Included in published bundle
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/01-implementation-diary.md — Included in published bundle
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md — Included in published bundle


## 2026-02-25 - Completed Phase 0 contract freeze artifacts

Implemented Phase 0 tasks T0.1-T0.8 by generating tool inventory, v1 scope classification, unsupported list, frozen envelope contract, and machine-readable request/response fixture sets for Go Host Lite parity testing.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/02-phase-0-contract-inventory-and-freeze.md — Phase 0 summary reference
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/01-tool-inventory.json — Extracted tool inventory
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/02-go-core-v1-classification.yaml — Core/provider/defer classification
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/03-core-v1-envelope-contract.yaml — Frozen request/response envelope contract
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/04-go-v1-unsupported-tools.json — Unsupported command list
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/05-core-request-fixtures.json — Request routing fixtures
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/06-response-normalization-fixtures.json — Response normalization fixtures

## 2026-02-25 - Completed Phase 1 transport foundation (T1.1-T1.5)

Created the initial Go host module scaffold, implemented Native Messaging frame encode/decode helpers with error handling, added framing tests, and added socket path resolution with `SURF_SOCKET_PATH` override parity.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/go.mod — New Go module declaration
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/nativeio/codec.go — Native Messaging framing implementation
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/nativeio/codec_test.go — Framing unit tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/config/socket_path.go — Socket path parity logic and env override support
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/config/socket_path_test.go — Socket path unit tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/tasks.md — Marked T1.1-T1.5 complete

## 2026-02-25 - Completed Phase 1 lifecycle + socket runtime tasks (T1.6-T1.12)

Added the socket listener abstraction, client session management, pending/stream registries, and host runtime lifecycle wiring including `HOST_READY`, stdin EOF disconnect notification, and SIGINT/SIGTERM cleanup behavior.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-host-go/main.go — Host runtime wiring for socket/native loops and lifecycle handling
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/listener.go — OS-agnostic listener abstraction
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/listener_unix.go — Unix socket implementation + cleanup behavior
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/listener_windows.go — Explicit Windows named-pipe placeholder behavior
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/session.go — Session manager and extension disconnect broadcast support
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/pending/id_allocator.go — Numeric host request ID allocation
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/pending/store.go — Pending request correlation map
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/stream_registry.go — Stream-to-session registry
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/session_test.go — Session manager tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/socketbridge/listener_unix_test.go — Unix listener test
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/pending/store_test.go — Pending store tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/stream_registry_test.go — Stream registry tests

## 2026-02-25 - Implemented socket ingress validation (T2.1-T2.3)

Added strict parser/validator logic for `tool_request`, `stream_request`, and `stream_stop` socket messages, and integrated validation into runtime before forwarding requests to the extension channel.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/ingress.go — Message-shape validation for socket ingress types
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/ingress_test.go — Validation tests for accepted/rejected payload shapes
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-host-go/main.go — Runtime integration of ingress validation and CLI error response path

## 2026-02-25 - Implemented core-v1 routing and runtime correlation (T2.4-T2.10)

Implemented host-side core tool routing, provider/deferred rejection paths, stream start/event handling parity, and tool-response envelope normalization for socket clients while preserving passthrough behavior for non-tool native commands.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap.go — Core-v1 tool and computer-action message mapping + unsupported command guards
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap_test.go — Mapping and unsupported-path tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/ingress.go — Extended tool request parse output (id/tab/window metadata)
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/pending/store.go — Pending metadata updated to preserve arbitrary original request IDs
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-host-go/main.go — Runtime routing, stream/session behavior, and `tool_response` shaping

## 2026-02-25 - Added Glazed CLI root + shared transport/format utilities (T3.1-T3.8)

Implemented the initial `surf-go` Cobra+Glazed skeleton, including root help/logging wiring, a first raw tool command scaffold, shared socket transport client, request envelope helper, and response row formatting utility.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go — Root Cobra command with help/logging wiring and command registration
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/tool_raw.go — Glazed command scaffold with output + command settings sections
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/transport/client.go — Shared socket transport utility
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/base.go — Shared tool_request envelope builder and execute helper
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/format.go — Shared response-to-row formatter
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/go.mod — Added Glazed/Cobra dependencies
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/go.sum — Dependency lock entries

## 2026-02-25 - Added Page/Input command wrappers in `surf-go` (T4.1-T4.16)

Implemented a reusable simple Glazed command wrapper for tool dispatch and wired the initial command groups and interaction commands for core browsing workflows.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/tool_simple.go — Reusable simple tool wrapper command
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go — Registration of page/wait/input commands

## 2026-02-25 - Added remaining core command groups + stream wrappers (T4.17-T4.34)

Completed surf-go command coverage for tabs/windows/frames/dialog plus network/console/cookie/emulation groups, and added reusable stream command support for `network stream` and `console stream`.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/stream_simple.go — Reusable stream command implementation
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/transport/client.go — Stream transport support (`stream_request`/`stream_stop` lifecycle)
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/transport/client_test.go — Stream transport test coverage
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go — Registration of remaining core command groups

## 2026-02-25 - Added Go host profile packaging + installer/uninstaller updates (T5.1-T5.7)

Implemented Go host build targets, profile-aware wrapper runtime selection (`SURF_HOST_PROFILE=core-go|node-full`), snap-aware Go host install parity, uninstall cleanup for Go artifacts, and migration/fallback documentation updates.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/scripts/build-go-host-binaries.cjs — Cross-platform Go host binary build script
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/scripts/install-native-host.cjs — Profile-aware wrapper generation and Go host build/install integration
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/scripts/uninstall-native-host.cjs — Go host artifact cleanup paths
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/package.json — Added `build:go-host` npm script and packaged `go/` sources
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/README.md — Runtime profile and migration/fallback documentation

## 2026-02-25 - Added Phase 6 automated verification coverage and manual checklist scaffold

Expanded verification with mapping contract tests, CLI integration tests, and installer smoke tests; added manual browser checklist for final human-in-the-loop validation.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap_contract_test.go — Core tool mapping parity contract tests
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/integration_test.go — Representative CLI integration tests against mock socket host
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/scripts/tests/native-host-installer-smoke.cjs — Installer smoke test for standard + snap Linux targets
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/reference/03-manual-browser-validation-checklist.md — T6.8 manual execution checklist
