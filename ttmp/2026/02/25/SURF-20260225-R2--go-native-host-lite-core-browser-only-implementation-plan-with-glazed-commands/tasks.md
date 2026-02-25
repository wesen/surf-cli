# Tasks

## Planning Deliverables

- [x] P1: Create ticket workspace and baseline docs.
- [x] P2: Produce design doc for Go Host Lite (core browser only).
- [x] P3: Produce granular implementation backlog.
- [ ] P4: Review plan and lock v1 scope with stakeholders.

## Phase 0 - Contract Inventory and Scope Freeze

- [x] T0.1: Extract all tool names currently routed by `mapToolToMessage`.
- [x] T0.2: Tag each tool as `core-v1`, `provider`, or `defer`.
- [x] T0.3: Define explicit unsupported command list for Go v1.
- [x] T0.4: Freeze core-v1 request envelope schema (`tool_request`, `stream_request`, `stream_stop`).
- [x] T0.5: Freeze Native Messaging response envelope rules (`id`, `type`, `error`, payload).
- [x] T0.6: Document disconnect behavior parity requirements.
- [x] T0.7: Build fixture table: request -> expected extension message.
- [x] T0.8: Build fixture table: extension response -> CLI response normalization.

## Phase 1 - Go Host Transport Foundation

- [x] T1.1: Create Go module scaffold (`cmd`, `internal`).
- [x] T1.2: Implement Native Messaging frame reader (length-prefixed JSON).
- [x] T1.3: Implement Native Messaging frame writer.
- [x] T1.4: Add malformed-frame handling and logging.
- [x] T1.5: Implement socket path resolver with `SURF_SOCKET_PATH` override parity.
- [x] T1.6: Implement Unix socket / Windows pipe listener abstraction.
- [x] T1.7: Implement socket client session manager.
- [x] T1.8: Implement request ID allocator and pending map.
- [x] T1.9: Implement stream registry (`streamId -> socket`).
- [x] T1.10: Emit `HOST_READY` after socket server is active.
- [x] T1.11: Handle `stdin EOF` by notifying connected clients and exiting.
- [x] T1.12: Add SIGINT/SIGTERM cleanup (remove socket path on unix).

## Phase 2 - Core Router and Host Runtime Behavior

- [x] T2.1: Parse and validate incoming socket `tool_request` payloads.
- [x] T2.2: Parse and validate incoming socket `stream_request` payloads.
- [x] T2.3: Parse and validate incoming socket `stream_stop` payloads.
- [ ] T2.4: Implement host-side routing for core-v1 tools.
- [ ] T2.5: Forward extension-bound messages preserving IDs.
- [ ] T2.6: Correlate extension responses to pending socket clients.
- [ ] T2.7: Implement stream event forwarding (`STREAM_EVENT`, `STREAM_ERROR`).
- [ ] T2.8: Implement provider-command rejection path with clear error text.
- [ ] T2.9: Implement parity handling for `GET_AUTH` passthrough.
- [ ] T2.10: Implement parity handling for `API_REQUEST` passthrough.

## Phase 3 - Glazed CLI Skeleton

- [ ] T3.1: Create `surf-go` root Cobra command.
- [ ] T3.2: Add Glazed output section via `settings.NewGlazedSchema()`.
- [ ] T3.3: Add command settings section via `cli.NewCommandSettingsSection()`.
- [ ] T3.4: Add logging section and `PersistentPreRunE` logger init.
- [ ] T3.5: Implement shared transport client (socket connect/write/read/timeout).
- [ ] T3.6: Implement base command helper for building `tool_request` envelopes.
- [ ] T3.7: Implement shared response-to-row formatter utility.
- [ ] T3.8: Add root help wiring with Glazed help system.

## Phase 4 - Glazed Core Commands (Group by Group)

### Page

- [ ] T4.1: Implement `page read` command.
- [ ] T4.2: Implement `page text` command.
- [ ] T4.3: Implement `page state` command.
- [ ] T4.4: Implement `page search` command.
- [ ] T4.5: Implement `wait element` command.
- [ ] T4.6: Implement `wait url` command.
- [ ] T4.7: Implement `wait network` command.
- [ ] T4.8: Implement `wait dom` command.

### Input / Interaction

- [ ] T4.9: Implement `click` command.
- [ ] T4.10: Implement `type` command.
- [ ] T4.11: Implement `key` command.
- [ ] T4.12: Implement `scroll` command.
- [ ] T4.13: Implement `hover` command.
- [ ] T4.14: Implement `drag` command.
- [ ] T4.15: Implement `select` command.
- [ ] T4.16: Implement `screenshot` command.

### Tabs / Windows / Frames / Dialog

- [ ] T4.17: Implement `tab list` command.
- [ ] T4.18: Implement `tab new` command.
- [ ] T4.19: Implement `tab switch` command.
- [ ] T4.20: Implement `tab close` command.
- [ ] T4.21: Implement `tab name` command.
- [ ] T4.22: Implement `tab named` command.
- [ ] T4.23: Implement `window list/new/focus/close/resize` commands.
- [ ] T4.24: Implement `frame list/switch/main/eval` commands.
- [ ] T4.25: Implement `dialog accept/dismiss/info` commands.

### Network / Console / Cookies / Emulation

- [ ] T4.26: Implement `network list` command.
- [ ] T4.27: Implement `network get` command.
- [ ] T4.28: Implement `network body` command.
- [ ] T4.29: Implement `network origins/stats/clear/export` commands.
- [ ] T4.30: Implement `network stream` command.
- [ ] T4.31: Implement `console read` command.
- [ ] T4.32: Implement `console stream` command.
- [ ] T4.33: Implement `cookie list/get/set/clear` commands.
- [ ] T4.34: Implement `emulate network/cpu/geo/device/viewport/touch` commands.

## Phase 5 - Compatibility, Packaging, and Installer

- [ ] T5.1: Add Go host binary build target (linux/mac/windows).
- [ ] T5.2: Add host profile flag/environment (`SURF_HOST_PROFILE=core-go|node-full`).
- [ ] T5.3: Update installer to install and reference Go host wrapper.
- [ ] T5.4: Keep Node host fallback path selectable.
- [ ] T5.5: Add Snap-aware Go host install target parity.
- [ ] T5.6: Update uninstall script for Go artifacts.
- [ ] T5.7: Document migration and fallback behavior in README.

## Phase 6 - Tests and Validation

- [ ] T6.1: Unit tests for Native Messaging framing.
- [ ] T6.2: Unit tests for socket transport and disconnect behavior.
- [ ] T6.3: Unit tests for request correlation/pending maps.
- [ ] T6.4: Contract tests for core tool mapping parity vs Node.
- [ ] T6.5: CLI integration tests for representative commands.
- [ ] T6.6: Stream integration tests (console/network start-stop).
- [ ] T6.7: Installer tests for standard Linux and Snap targets.
- [ ] T6.8: Manual real-browser validation checklist execution.

## Phase 7 - Rollout

- [ ] T7.1: Enable Go host as opt-in beta profile.
- [ ] T7.2: Collect regression reports and fix parity gaps.
- [ ] T7.3: Promote Go host to default for core commands.
- [ ] T7.4: Keep provider commands routed to Node profile until separate plan lands.
