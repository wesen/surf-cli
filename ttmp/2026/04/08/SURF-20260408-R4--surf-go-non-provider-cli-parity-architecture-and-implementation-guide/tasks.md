# Tasks

## Phase 0 - Lock the Command Authoring Contract

- [x] Generate a machine-readable inventory of missing non-provider Node-vs-Go commands.
- [x] Write a detailed architecture/design/implementation guide for new engineers.
- [x] Relate key source files to the ticket design doc and diary.
- [x] Validate ticket vocabulary and metadata with `docmgr doctor`.
- [x] Update the design doc to make "all new public verbs must be Glazed commands" explicit.
- [x] Add a Glazed command authoring checklist to the design doc.
- [ ] Define which common Glazed sections every new command should expose.
- [ ] Define which shared helper functions are acceptable below the command boundary.
- [ ] Decide the naming convention for new command files under `go/internal/cli/commands`.

## Phase 1 - Shared Glazed Infrastructure

- [ ] Review `go/internal/cli/commands/base.go` for reusable helpers that should be extracted for new command implementations.
- [ ] Review `go/internal/cli/commands/tool_simple.go` and decide what pieces can be reused without making the public command surface generic.
- [ ] Add or refine shared helpers for decoding typed settings structs from Glazed values.
- [ ] Add or refine shared helpers for command-level tool execution so new commands do not duplicate transport boilerplate.
- [ ] Add or refine shared helpers for row emission so structured outputs stay consistent.
- [ ] Add or refine shared helpers for common browser-targeting fields such as tab ID and window ID.
- [ ] Add or refine shared helpers for common timeout settings where commands need them.
- [ ] Add or refine shared helpers for command examples and section wiring where repetition would otherwise be high.
- [ ] Add tests for any new shared command helpers introduced in this phase.

## Phase 2 - Workflow Unlocker Commands

### `js`

- [ ] Create `go/internal/cli/commands/js.go`.
- [ ] Define Glazed arguments for inline JavaScript source.
- [ ] Define Glazed flags for `--file` and any browser-targeting settings.
- [ ] Define a typed settings struct for `js`.
- [ ] Implement input normalization so `--file` and inline code cannot conflict silently.
- [ ] Implement request shaping for tool `js`.
- [ ] Preserve structured response output rather than flattening everything to text.
- [ ] Add command-specific help examples for inline code and file-based code.
- [ ] Add unit tests for validation and request shaping.
- [ ] Validate `js` against `tool-raw`.
- [ ] Validate `js` against a real browser session.

### `upload`

- [ ] Create `go/internal/cli/commands/upload.go`.
- [ ] Define Glazed flags for upload target ref and file list input.
- [ ] Define a typed settings struct for `upload`.
- [ ] Normalize comma-separated and repeated file inputs into a stable list.
- [ ] Enforce non-empty file input validation.
- [ ] Implement request shaping for tool `upload`.
- [ ] Add help examples for single-file and multi-file uploads.
- [ ] Add unit tests for normalization and validation.
- [ ] Validate `upload` against `tool-raw`.
- [ ] Validate `upload` against a real browser session.

### `form fill`

- [ ] Create `go/internal/cli/commands/form_fill.go`.
- [ ] Decide the Glazed interface for structured fill data.
- [ ] Define typed flags and arguments for form payload input.
- [ ] Define a typed settings struct for `form fill`.
- [ ] Implement payload parsing and validation.
- [ ] Implement request shaping for tool `form.fill`.
- [ ] Preserve structured response output.
- [ ] Add help examples for common fill patterns.
- [ ] Add unit tests for payload parsing and request shaping.
- [ ] Validate `form fill` against `tool-raw`.
- [ ] Validate `form fill` against a real browser session.

### `locate role`, `locate text`, `locate label`

- [ ] Create `go/internal/cli/commands/locate.go`.
- [ ] Create a Glazed parent group for `locate`.
- [ ] Define a Glazed command constructor for `locate role`.
- [ ] Define a typed settings struct for `locate role`.
- [ ] Wire flags for name, action, value, and selection behavior for `locate role`.
- [ ] Implement request shaping for tool `locate.role`.
- [ ] Define a Glazed command constructor for `locate text`.
- [ ] Define a typed settings struct for `locate text`.
- [ ] Wire flags for exact matching, action, and value for `locate text`.
- [ ] Implement request shaping for tool `locate.text`.
- [ ] Define a Glazed command constructor for `locate label`.
- [ ] Define a typed settings struct for `locate label`.
- [ ] Wire flags for action and value for `locate label`.
- [ ] Implement request shaping for tool `locate.label`.
- [ ] Add command-specific help examples for find-only and action-taking flows.
- [ ] Add unit tests for all three subcommands.
- [ ] Validate all three locate subcommands against `tool-raw`.
- [ ] Validate all three locate subcommands against a real browser session.

### `wait load`

- [ ] Create `go/internal/cli/commands/wait_load.go`.
- [ ] Register `wait load` as a Glazed subcommand under the existing `wait` group.
- [ ] Define typed timeout-related fields for `wait load`.
- [ ] Define a typed settings struct for `wait load`.
- [ ] Implement request shaping for tool `wait.load`.
- [ ] Add help examples for default and explicit timeout usage.
- [ ] Add unit tests for validation and request shaping.
- [ ] Validate `wait load` against `tool-raw`.
- [ ] Validate `wait load` against a real browser session.

## Phase 3 - Existing Routed Utility Commands

### `network curl` and `network path`

- [ ] Create `go/internal/cli/commands/network_curl.go`.
- [ ] Create `go/internal/cli/commands/network_path.go`.
- [ ] Register both as Glazed subcommands under `network`.
- [ ] Define typed request-ID input for both commands.
- [ ] Implement request shaping for tool `network.curl`.
- [ ] Implement request shaping for tool `network.path`.
- [ ] Preserve structured response output for path metadata and request details.
- [ ] Add command-specific help examples.
- [ ] Add unit tests for both commands.
- [ ] Validate both commands against `tool-raw`.

### `tab unname`, `tab group`, `tab ungroup`, `tab groups`

- [ ] Create `go/internal/cli/commands/tab_metadata.go` or a similarly named Glazed command file.
- [ ] Register the new tab metadata commands under `tab`.
- [ ] Define typed fields for group names, tab IDs, and selection behavior where required.
- [ ] Implement request shaping for tool `tab.unname`.
- [ ] Implement request shaping for tool `tab.group`.
- [ ] Implement request shaping for tool `tab.ungroup`.
- [ ] Implement request shaping for tool `tab.groups`.
- [ ] Add help examples for common tab-grouping flows.
- [ ] Add unit tests for the tab metadata commands.
- [ ] Validate the tab metadata commands against `tool-raw`.

### `scroll top`, `scroll bottom`, `scroll to`, `scroll info`

- [ ] Decide whether to introduce a dedicated `scroll` Glazed group or expand the existing root-level scroll surface carefully.
- [ ] Create `go/internal/cli/commands/scroll_group.go` or similarly scoped files.
- [ ] Define a Glazed command constructor for `scroll top`.
- [ ] Define a Glazed command constructor for `scroll bottom`.
- [ ] Define a Glazed command constructor for `scroll to`.
- [ ] Define a Glazed command constructor for `scroll info`.
- [ ] Define typed fields for target position and behavior where required.
- [ ] Implement request shaping for tools `scroll.top`, `scroll.bottom`, `scroll.to`, and `scroll.info`.
- [ ] Add help examples for relative and absolute scrolling patterns.
- [ ] Add unit tests for the scroll command family.
- [ ] Validate the scroll command family against `tool-raw`.

## Phase 4 - Browser Utility and Bookkeeping Commands

### `perf start`, `perf stop`, `perf metrics`

- [ ] Create `go/internal/cli/commands/perf.go`.
- [ ] Register a Glazed `perf` group with `start`, `stop`, and `metrics` subcommands.
- [ ] Define typed settings structs for each perf subcommand.
- [ ] Implement request shaping for tools `perf.start`, `perf.stop`, and `perf.metrics`.
- [ ] Preserve structured response output for performance data.
- [ ] Add help examples for performance capture sessions.
- [ ] Add unit tests for the perf command family.
- [ ] Validate the perf command family against `tool-raw`.

### `zoom`

- [ ] Create `go/internal/cli/commands/zoom.go`.
- [ ] Decide the Glazed surface for get, set, and reset behavior.
- [ ] Define typed fields for zoom level selection and reset semantics.
- [ ] Implement request shaping for the relevant zoom tools.
- [ ] Add help examples for reading and changing tab zoom.
- [ ] Add unit tests for zoom request shaping.
- [ ] Validate zoom commands against `tool-raw`.

### `resize`

- [ ] Create `go/internal/cli/commands/resize.go`.
- [ ] Decide whether `resize` should target the current window only or expose explicit window targeting as Glazed fields.
- [ ] Define typed width and height fields.
- [ ] Implement request shaping for tool `resize`.
- [ ] Add help examples for common viewport/window sizes.
- [ ] Add unit tests for resize validation and request shaping.
- [ ] Validate `resize` against `tool-raw`.

### `bookmark add`, `bookmark remove`, `bookmark list`

- [ ] Create `go/internal/cli/commands/bookmark.go`.
- [ ] Register a Glazed `bookmark` group with `add`, `remove`, and `list` subcommands.
- [ ] Define typed settings structs for each bookmark subcommand.
- [ ] Implement request shaping for `bookmark.add`, `bookmark.remove`, and `bookmark.list`.
- [ ] Preserve structured response output for bookmark records.
- [ ] Add help examples for bookmark creation and listing.
- [ ] Add unit tests for the bookmark command family.
- [ ] Validate the bookmark command family against `tool-raw`.

### `history list`, `history search`

- [ ] Create `go/internal/cli/commands/history.go`.
- [ ] Register a Glazed `history` group with `list` and `search` subcommands.
- [ ] Define typed settings structs for `history list` and `history search`.
- [ ] Define typed query and limit fields where appropriate.
- [ ] Implement request shaping for `history.list` and `history.search`.
- [ ] Preserve structured response output for history entries.
- [ ] Add help examples for recent-history and query-based history usage.
- [ ] Add unit tests for the history command family.
- [ ] Validate the history command family against `tool-raw`.

## Phase 5 - Registration, Help, and Reviewability

- [ ] Register every new Glazed command in `go/cmd/surf-go/main.go`.
- [ ] Keep command grouping coherent so related verbs appear together in `surf-go --help`.
- [ ] Add or update help text so descriptions match Node semantics without copying Node quirks blindly.
- [ ] Ensure all new commands default to the intended Glazed output format.
- [ ] Ensure structured outputs land in useful rows rather than opaque text blobs.
- [ ] Add or update tests that cover command registration and help visibility.
- [ ] Run `go test ./internal/cli/commands ./cmd/surf-go`.
- [ ] Run `go test ./internal/host/router` if any router touch becomes necessary.

## Phase 6 - Deferred Design Decisions

- [ ] Decide whether `batch` belongs in this ticket or should stay in a follow-up ticket.
- [ ] Decide whether Node aliases like `read` and `find` should be added to `surf-go`.
- [ ] Decide whether composite helpers such as `smart_type`, `click_type`, and `type_submit` should be added or intentionally omitted.
- [ ] Decide whether `wait` as a local timing helper needs a separate Glazed command or should stay out of scope.
- [ ] Decide whether `computer` belongs in a separate orchestration-focused ticket.

## Phase 7 - Documentation and Delivery

- [ ] Update the design doc after implementation details solidify.
- [ ] Add implementation diary entries as each phase lands.
- [ ] Update README and `README.go.md` for the new first-class Glazed commands.
- [ ] Re-run `docmgr doctor --ticket SURF-20260408-R4 --stale-after 30`.
- [ ] Upload an updated ticket bundle to reMarkable after the next major milestone.
