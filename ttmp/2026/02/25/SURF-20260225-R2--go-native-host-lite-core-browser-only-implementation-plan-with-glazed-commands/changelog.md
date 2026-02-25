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
