# Tasks

## Phase 1: Provider Group Refactor

- [x] Create a `chatgpt` Cobra parent command in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go).
- [x] Create a `kagi` Cobra parent command in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go).
- [x] Create a placeholder `gmail` Cobra parent command in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go) before Gmail subcommands exist.
- [x] Register the current ChatGPT ask implementation as `surf-go chatgpt ask`.
- [x] Register the current ChatGPT transcript implementation as `surf-go chatgpt transcript`.
- [x] Register the current Kagi search implementation as `surf-go kagi search`.
- [x] Register the current Kagi assistant implementation as `surf-go kagi assistant`.
- [x] Remove the old flat root registrations for `chatgpt`, `chatgpt-transcript`, `kagi-search`, and `kagi-assistant`.
- [x] Decide whether `chatgpt models` should be introduced now or deferred; record the decision in the design doc or changelog.
- [x] Update root-command integration tests so they assert the grouped command paths rather than the old flat names.
- [x] Validate `go run ./cmd/surf-go --help`, `go run ./cmd/surf-go chatgpt --help`, and `go run ./cmd/surf-go kagi --help`.

## Phase 2: Help and Discoverability Updates

- [x] Update any existing help docs or examples that still reference flat provider commands.
- [x] Update [01-building-browser-side-verbs.md](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/pkg/doc/tutorials/01-building-browser-side-verbs.md) examples so they use grouped command paths where relevant.
- [ ] Add grouped command examples to root help output expectations if there are tests or snapshots covering help text.
- [ ] Verify `go run ./cmd/surf-go help build-browser-side-verbs` still renders cleanly after grouped command changes.

## Phase 3: Gmail Research Setup

- [x] Create the first Gmail research diary document under this ticket's `reference/` directory.
- [x] Use ordered numeric prefixes for all Gmail research scripts placed under this ticket's `scripts/` directory.
- [x] Add an initial script that confirms current Gmail page URL, title, login state, and high-level DOM markers.
- [x] Add an initial script that inventories likely inbox row selectors and row attributes.
- [x] Add an initial script that inventories Gmail search box selectors and submit paths.
- [x] Record the first pass findings in the Gmail research diary before writing command code.

## Phase 4: Gmail Inbox Research

- [x] Probe inbox row selectors in the real Gmail session.
- [x] Determine whether inbox rows expose stable thread identifiers in DOM attributes, links, or both.
- [x] Determine how unread state is represented in row classes, ARIA attributes, or text markers.
- [x] Determine how starred state is represented in row DOM.
- [x] Determine how attachment presence is represented in row DOM.
- [x] Determine whether participants, subject, snippet, and timestamp can be extracted reliably across multiple rows.
- [ ] Determine whether inbox density or preview-pane settings materially change the row selectors.
- [ ] Write down accepted selectors and rejected selectors in the Gmail research diary.

## Phase 5: Gmail Search Research

- [x] Probe the Gmail search input selector and confirm whether simple `.value` assignment is sufficient or whether synthetic input events are required.
- [x] Probe the submit path for Gmail search: Enter key, submit button, or both.
- [x] Determine whether Gmail search results reuse inbox row DOM or require separate selectors.
- [ ] Determine whether labels are exposed in result rows.
- [ ] Determine whether query state is visible in DOM after submission and can be echoed back in output.
- [ ] Determine whether search result rows expose stable thread links or identifiers.
- [ ] Record all accepted and rejected search selectors in the Gmail research diary.

## Phase 6: Gmail Command Surface Design

- [ ] Finalize the Glazed settings struct for `gmail list`.
- [ ] Finalize the Glazed settings struct for `gmail search`.
- [ ] Decide whether both Gmail commands should be dual-mode or whether one should be glaze-first only; record the decision.
- [ ] Decide whether `gmail list` should require `--inbox` in v1 or default to inbox when no mode is specified; record the decision.
- [x] Ensure both command surfaces include explicit owned-tab cleanup behavior via `--keep-tab-open`.
- [x] Ensure both command surfaces support `--tab-id`, `--window-id`, and `--debug-socket`.

## Phase 7: Implement `gmail list`

- [x] Create [gmail_list.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/gmail_list.go).
- [x] Create embedded production JS at [scripts/gmail_list.js](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/gmail_list.js).
- [x] Build the `SURF_OPTIONS` prelude generator for `gmail list`.
- [x] Implement the shared fetch path for `gmail list`, including target tab resolution, optional tab creation, and owned-tab cleanup.
- [x] Implement DOM extraction for inbox rows in the embedded JS.
- [x] Implement response parsing in Go.
- [x] Implement structured row shaping in Go.
- [x] Implement writer output if `gmail list` is dual-mode.
- [x] Register `gmail list` under the `gmail` provider parent in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go).

## Phase 8: Validate `gmail list`

- [ ] Add unit tests for `gmail list` script-prelude generation.
- [ ] Add unit tests for `gmail list` response parsing.
- [ ] Add unit tests for `gmail list` row shaping and, if applicable, Markdown rendering.
- [x] Add mock-host integration tests for the expected request sequence: `tab.new` or `navigate`, then `js`, then `tab.close` when the tab is command-owned.
- [x] Live-validate `go run ./cmd/surf-go gmail list --inbox` against the real Gmail session.
- [ ] Live-validate `go run ./cmd/surf-go gmail list --inbox --keep-tab-open`.
- [ ] Live-validate explicit target behavior with `--tab-id` or `--window-id` and confirm no user-supplied tab is closed.

## Phase 9: Implement `gmail search`

- [x] Create [gmail_search.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/gmail_search.go).
- [x] Create embedded production JS at [scripts/gmail_search.js](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/gmail_search.js).
- [x] Build the `SURF_OPTIONS` prelude generator for `gmail search`.
- [x] Implement the shared fetch path for `gmail search`, including target tab resolution, optional tab creation, and owned-tab cleanup.
- [x] Implement search-box interaction and result extraction in the embedded JS.
- [x] Implement response parsing in Go.
- [x] Implement structured row shaping in Go.
- [x] Implement writer output if `gmail search` is dual-mode.
- [x] Register `gmail search` under the `gmail` provider parent in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go).

## Phase 10: Validate `gmail search`

- [ ] Add unit tests for `gmail search` script-prelude generation.
- [ ] Add unit tests for `gmail search` response parsing.
- [ ] Add unit tests for `gmail search` row shaping and, if applicable, Markdown rendering.
- [x] Add mock-host integration tests for the expected request sequence: `tab.new` or `navigate`, then `js`, then `tab.close` when the tab is command-owned.
- [x] Live-validate `go run ./cmd/surf-go gmail search --query "from:boss"` against the real Gmail session.
- [ ] Live-validate `go run ./cmd/surf-go gmail search --query "from:boss" --keep-tab-open`.
- [ ] Live-validate explicit target behavior with `--tab-id` or `--window-id` and confirm no user-supplied tab is closed.

## Phase 11: Gmail Workflow Documentation

- [x] Add a new help/tutorial page for stateful browser workflow verbs under `go/pkg/doc/tutorials/`.
- [x] Use Gmail as the motivating example for page-state handling, read-only safety, owned-tab cleanup, and research-script discipline.
- [ ] Link the new Gmail tutorial from the existing browser-side playbook where appropriate.
- [ ] Verify `go run ./cmd/surf-go help <new-gmail-tutorial-slug>` renders correctly.

## Phase 12: Ticket Hygiene and Final Validation

- [ ] Add or update changelog entries in this ticket as implementation work lands.
- [ ] Keep the Gmail research diary current as selectors are accepted or rejected.
- [ ] Relate implementation files and help docs to this ticket with `docmgr doc relate`.
- [ ] Run `go test ./internal/cli/commands ./cmd/surf-go` from the `go/` module root.
- [ ] Run `docmgr doctor --ticket SURF-20260410-R5 --stale-after 30`.
- [ ] Clean up any resulting docmgr vocabulary or metadata issues.
