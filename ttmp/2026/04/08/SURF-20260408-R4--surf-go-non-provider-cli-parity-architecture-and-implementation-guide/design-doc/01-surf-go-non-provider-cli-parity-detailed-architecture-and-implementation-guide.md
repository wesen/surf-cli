---
Title: 'surf-go Non-Provider CLI Parity: Detailed Architecture and Implementation Guide'
Ticket: SURF-20260408-R4
Status: active
Topics:
    - go
    - glazed
    - chromium
    - cli
    - native-messaging
    - architecture
    - onboarding
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: |-
        Current surf-go command registration surface and grouping model
        Current Go CLI registration surface
    - Path: go/internal/cli/commands/chatgpt.go
      Note: Example of a richer provider-specific first-class command in Go
    - Path: go/internal/cli/commands/navigate.go
      Note: Example of a typed first-class Go command instead of args-json wrapper
    - Path: go/internal/cli/commands/tool_raw.go
      Note: Escape hatch for raw tool dispatch while first-class commands are missing
    - Path: go/internal/cli/commands/tool_simple.go
      Note: |-
        Current generic Glazed wrapper used by most Go commands
        Current generic Glazed wrapper pattern
    - Path: go/internal/cli/transport/client.go
      Note: Socket transport used by all surf-go commands
    - Path: go/internal/host/router/toolmap.go
      Note: |-
        Go host mapping layer that already knows how to route many missing commands
        Go router parity evidence and existing mappings
    - Path: native/cli.cjs
      Note: |-
        Node CLI command catalog, UX, and alias behavior used as parity baseline
        Node CLI baseline and user-facing command catalog
    - Path: native/host-helpers.cjs
      Note: |-
        Canonical Node request mapping for non-provider tools
        Canonical Node request mapping for missing non-provider commands
    - Path: sources/01-node-vs-go-non-provider-command-gap.json
      Note: Generated inventory of Node-vs-Go non-provider command gaps
    - Path: src/service-worker/index.ts
      Note: |-
        Browser-side implementation of the extension message types the host emits
        Browser-side implementations for zoom
    - Path: ttmp/2026/04/08/SURF-20260408-R4--surf-go-non-provider-cli-parity-architecture-and-implementation-guide/sources/01-node-vs-go-non-provider-command-gap.json
      Note: Generated missing-command inventory
ExternalSources: []
Summary: Detailed intern-oriented architecture and implementation guide for closing non-provider CLI parity gaps between the Node surf CLI and surf-go.
LastUpdated: 2026-04-08T22:05:00-04:00
WhatFor: Onboard a new engineer to the Surf command stack and provide a concrete plan for implementing missing non-provider surf-go commands.
WhenToUse: Use when implementing or reviewing surf-go command parity work outside the provider-specific AI integrations.
---


# surf-go Non-Provider CLI Parity: Detailed Architecture and Implementation Guide

## Executive Summary

This document explains how the Surf system is put together, where the Node CLI and `surf-go` differ today, and how to close the non-provider command parity gap without breaking the architectural boundaries that already exist. It is written for an engineer who has not worked in this repository before and needs both architectural orientation and a concrete implementation plan.

The most important finding is that the Go host and extension already support a substantial portion of the missing non-provider browser functionality. In many cases the missing piece is not host-side browser support, but simply the absence of a first-class `surf-go` command and typed Glazed argument schema. The practical work therefore splits into two categories. The first category is command-surface work: expose missing tools in `surf-go`, wire their flags, and shape output predictably. The second category is true feature work: a smaller set of commands may still require browser-side or host-side implementation details, but most of the high-value missing verbs are already routable in `go/internal/host/router/toolmap.go`.

The recommended approach is not to port the Node CLI one command at a time in arbitrary order. Instead, treat the work as a layered CLI product effort. Start with the highest-value missing primitives that unlock more workflows (`js`, `upload`, `form.fill`, `locate.*`, `wait.load`), then fill in adjacent inspection and utility commands (`network.curl`, `network.path`, `perf.*`, `tab.group`, `tab.unname`), and only then add lower-value convenience or bookkeeping commands (`bookmark.*`, `history.*`, `zoom`, `resize`, `batch`). This keeps the intern's work tractable and makes each phase independently shippable.

For this ticket, every new public `surf-go` verb should be authored as a Glazed command. That does not mean every command needs a large bespoke implementation file, but it does mean the public command surface should be described through Glazed command descriptions, Glazed sections, typed fields, and predictable output shaping. Ad hoc Cobra-only shortcuts or raw `--args-json` exposure are acceptable for diagnostics through `tool-raw`, but they are not the target end state for parity work in this ticket.

## Problem Statement

The repository currently contains two major CLI surfaces for browser automation:

1. The Node-based `surf` CLI, implemented primarily in `native/cli.cjs`, which offers the widest end-user command surface and the most polished user experience.
2. The Go-based `surf-go` CLI, implemented in `go/cmd/surf-go` and `go/internal/cli/commands`, which already works against the Go native host but exposes fewer commands and generally relies on more generic `--args-json` based wrappers.

This creates several concrete problems.

- A user who knows the Node CLI command vocabulary cannot reliably transfer that knowledge to `surf-go`.
- The Go host is forced to carry routing support for tools that the Go CLI does not expose cleanly.
- The implementation burden for future features grows because there is no single clear pattern for deciding when a command should be a typed Glazed command versus a generic wrapper.
- New engineers have to reverse engineer four layers at once: Node CLI surface, Go CLI surface, host router mapping, and service worker implementation.

The specific scope for this ticket is the non-provider command surface. We explicitly exclude AI-provider verbs such as `gemini`, `perplexity`, `grok`, and `aistudio`, because those have their own orchestration complexity and their own ticket history. The focus here is browser primitives, inspection, navigation, form operations, history/bookmarks, and related tooling.

## Scope

### In Scope

- Analysis of Node non-provider CLI surface.
- Analysis of Go CLI command-registration and command-shaping architecture.
- Inventory of missing non-provider commands.
- Implementation guidance for exposing missing tools in `surf-go`.
- Recommendations for which commands should be typed Glazed commands versus generic wrappers.
- Phased backlog and test plan.

### Out of Scope

- Provider-specific AI commands: `gemini`, `perplexity`, `grok`, `aistudio`, `aistudio.build`.
- Redesign of the extension message protocol.
- Replacing the Node CLI entirely.
- Large compatibility layers whose only purpose is to preserve accidental Node UX quirks.

## Current System Overview

There are four layers that matter.

1. CLI surface layer.
   - Node: `native/cli.cjs`
   - Go: `go/cmd/surf-go/main.go`, `go/internal/cli/commands/*`
2. Host request mapping layer.
   - Node: `native/host-helpers.cjs`
   - Go: `go/internal/host/router/toolmap.go`
3. Native transport layer.
   - Go CLI socket client: `go/internal/cli/transport/client.go`
   - Go native host runtime: `go/cmd/surf-host-go/main.go`
   - Extension native-connection manager: `src/native/port-manager.ts`
4. Browser execution layer.
   - Service worker handlers: `src/service-worker/index.ts`

### High-Level Request Flow

```text
+-----------------+        unix socket / NDJSON       +-------------------+
| surf-go command | --------------------------------> | surf-host-go      |
| (Glazed/Cobra)  |                                    | runtime + router   |
+-----------------+                                    +-------------------+
          |                                                      |
          |                                                      | Native Messaging
          v                                                      v
+-----------------+                                    +-------------------+
| command schema  |                                    | extension service |
| + row shaping   |                                    | worker            |
+-----------------+                                    +-------------------+
                                                                   |
                                                                   v
                                                        +-------------------+
                                                        | Chrome APIs / DOM |
                                                        | / CDP / tab state |
                                                        +-------------------+
```

### Why This Split Exists

The Node CLI was built first and therefore contains both UX concerns and protocol knowledge in a single file tree. The Go version was intentionally designed more cleanly, with transport, routing, and command description separated. That separation is good, but it means parity work has to be done consciously at the CLI layer rather than assuming “host support implies command support.”

## Evidence-Based Architecture Walkthrough

### 1. Node CLI Surface Is the Current User-Facing Baseline

The Node CLI declares human-facing commands, groups, examples, aliases, and option semantics inside `native/cli.cjs`. The relevant command catalog for the missing work spans `native/cli.cjs:630-1210`.

Important examples:

- Semantic location verbs live at `native/cli.cjs:630-672`.
- `wait.load` is declared at `native/cli.cjs:705-731`.
- `js` is declared at `native/cli.cjs:785-799`.
- `network.curl` and `network.path` are declared at `native/cli.cjs:815-912`.
- `form.fill`, `perf.*`, and `upload` are declared at `native/cli.cjs:987-1011`.
- `batch`, `zoom`, `resize`, `bookmark.*`, `history.*` are declared at `native/cli.cjs:1091-1199`.

This file is not only a list of commands. It is also the repository's most complete description of user expectations:

- canonical verb names,
- grouped help output,
- typed flags,
- examples,
- aliases,
- some argument conventions.

A new intern should treat `native/cli.cjs` as the product-level parity baseline, not as code to mechanically translate.

### 2. Node Host Helpers Show the Canonical Mapping Contract

`native/host-helpers.cjs` is the clearest statement of “when the user runs command X, which extension message shape do we send?” The relevant mapping block is `native/host-helpers.cjs:650-1035`.

Examples:

- `js` maps to `EXECUTE_JAVASCRIPT` at `native/host-helpers.cjs:771-772`.
- `wait.load` maps to `WAIT_FOR_LOAD` at `native/host-helpers.cjs:789-790`.
- `form.fill` maps to `FORM_FILL` at `native/host-helpers.cjs:842-847`.
- `upload` maps to `UPLOAD_FILE` at `native/host-helpers.cjs:854-856`.
- `locate.role`, `locate.text`, `locate.label` map to `LOCATE_*` messages at `native/host-helpers.cjs:873-902`.
- `batch` maps to `BATCH_EXECUTE` at `native/host-helpers.cjs:980-1008`.
- `zoom`, `resize`, `bookmark.*`, `history.*` map at `native/host-helpers.cjs:1015-1031`.

This layer is valuable because it strips away CLI polish and shows the actual host contract. If a Go command is missing but the Go router already maps the same tool, then the work is CLI-surface work rather than feature invention.

### 3. The Go Router Already Knows Many of the Missing Verbs

The Go equivalent lives in `go/internal/host/router/toolmap.go`. The relevant mapping range is `go/internal/host/router/toolmap.go:133-425`.

The important architectural fact is that the Go router already covers many of the non-provider commands still missing from `surf-go`.

Examples already mapped in Go:

- `js` at `go/internal/host/router/toolmap.go:133-135`
- `wait.load` at `go/internal/host/router/toolmap.go:170-171`
- `network.curl` at `go/internal/host/router/toolmap.go:194-195`
- `network.path` at `go/internal/host/router/toolmap.go:204-205`
- `tab.unname` at `go/internal/host/router/toolmap.go:214-215`
- `scroll.top` / `scroll.bottom` / `scroll.to` at `go/internal/host/router/toolmap.go:237-242`
- `form.fill` at `go/internal/host/router/toolmap.go:304-313`
- `zoom` at `go/internal/host/router/toolmap.go:365-372`
- `resize` at `go/internal/host/router/toolmap.go:373-374`
- `window.resize` at `go/internal/host/router/toolmap.go:404-425`

Also notable: some lower-priority verbs are already represented in the unsupported/tool inventory at the top of `toolmap.go`, such as `batch`, `perf.*`, `bookmark.*`, and `history.*`.

This tells us two things.

- The Go host is ahead of the Go CLI surface in several places.
- The intern's first tasks should focus on surfacing host support, not re-implementing the protocol.

### 4. The Service Worker Already Implements Several Browser-Side Backends

The browser implementation for several missing commands already exists in `src/service-worker/index.ts`.

Examples from `src/service-worker/index.ts:2501-2615`:

- `TAB_RELOAD` at `2501-2505`
- `ZOOM_GET`, `ZOOM_SET`, `ZOOM_RESET` at `2507-2525`
- `BOOKMARK_ADD`, `BOOKMARK_REMOVE`, `BOOKMARK_LIST` at `2527-2579`
- `HISTORY_LIST`, `HISTORY_SEARCH` at `2581-2614`

This is critical for scoping. For `zoom`, `bookmark.*`, and `history.*`, the browser-side logic is already done. The gap is mostly command exposure and output shaping.

### 5. The Current Go CLI Surface Is Smaller and More Generic

`go/cmd/surf-go/main.go:20-267` registers the current Go CLI commands. Today it includes:

- root-level `chatgpt`, `tool-raw`, `navigate`
- grouped commands for `page`, `wait`, `tab`, `window`, `frame`, `dialog`, `network`, `console`, `cookie`, `emulate`
- a set of direct input actions (`click`, `type`, `key`, `scroll`, `hover`, `drag`, `select`, `screenshot`, `back`, `forward`, `reload`)

What it does not register are many Node verbs that the host and service worker already understand.

### 6. The Current Generic Go Command Pattern Uses `--args-json`

Most Go commands are instances of `NewSimpleToolCommand` in `go/internal/cli/commands/tool_simple.go:34-114`.

That command shape has three advantages.

- It is quick to add.
- It keeps transport behavior consistent.
- It avoids writing too much command-specific code.

It also has major user-experience drawbacks.

- Users need to know JSON keys rather than normal CLI flags.
- Help output is generic rather than task-specific.
- Required-argument validation is weaker.
- It is easy to expose the same tool with inconsistent semantics across commands.

This means parity work should not be “only add more `NewSimpleToolCommand` calls.” Some commands deserve proper typed wrappers.

## Generated Command-Gap Inventory

A generated inventory is stored at `sources/01-node-vs-go-non-provider-command-gap.json`.

At the time of this analysis:

- Node socket tools: `105`
- current `surf-go` registered tools: `58`
- excluded from this ticket: providers plus `ai`, `health`, `smoke`

The remaining missing non-provider tools are:

- `form_input`
- `find_and_type`
- `autocomplete`
- `set_value`
- `smart_type`
- `scroll_to_position`
- `get_scroll_info`
- `close_dialogs`
- `page_state`
- `javascript_tool`
- `click_type`
- `click_type_submit`
- `type_submit`
- `scroll_to`
- `left_click_drag`
- `wait`
- `computer`
- `locate.role`
- `locate.text`
- `locate.label`
- `tab.unname`
- `tab.group`
- `tab.ungroup`
- `tab.groups`
- `scroll.top`
- `scroll.bottom`
- `scroll.to`
- `scroll.info`
- `wait.load`
- `js`
- `network.curl`
- `network.path`
- `form.fill`
- `perf.start`
- `perf.stop`
- `perf.metrics`
- `upload`
- `batch`
- `zoom`
- `resize`
- `bookmark.add`
- `bookmark.remove`
- `bookmark.list`
- `history.list`
- `history.search`

Not all of these deserve equal treatment.

## Command Categories and Recommended Handling

The easiest mistake is to treat every missing command as the same kind of task. They are not.

### Category A: High-Value Typed Commands

These should get dedicated Go command structs with first-class flags, validation, and good help.

- `js`
- `upload`
- `form.fill`
- `locate.role`
- `locate.text`
- `locate.label`
- `wait.load`
- `network.curl`
- `network.path`
- `zoom`
- `resize`
- `history.search`

Why these deserve typed commands:

- Their flags are semantically meaningful.
- Users will run them directly rather than only through scripts.
- Good validation matters.
- Node UX is already specific and easy to mirror.

### Category B: Medium-Value Glazed Wrappers

These should still ship as Glazed commands, but they can use thinner implementations and shared helpers as long as the public interface is explicit and documented.

- `tab.unname`
- `tab.group`
- `tab.ungroup`
- `tab.groups`
- `scroll.top`
- `scroll.bottom`
- `scroll.to`
- `scroll.info`
- `perf.start`
- `perf.stop`
- `perf.metrics`
- `bookmark.add`
- `bookmark.remove`
- `bookmark.list`
- `history.list`

These are still useful, but the first ship can focus on a thinner Glazed command surface rather than the richer bespoke ergonomics reserved for the highest-value verbs.

### Category C: Low-Value Alias or Composite Commands

These should not be implemented first.

- `page_state`
- `javascript_tool`
- `scroll_to_position`
- `scroll_to`
- `get_scroll_info`
- `close_dialogs`
- `form_input`
- `find_and_type`
- `autocomplete`
- `set_value`
- `smart_type`
- `click_type`
- `click_type_submit`
- `type_submit`
- `left_click_drag`

These are either aliases, convenience composites, or UX sugar on top of lower-level primitives. They should come only after the primitives are first-class in Go.

### Category D: Special Case / Requires Design Choice

- `wait`
- `batch`
- `computer`

These are special because they are not just one more browser primitive.

- `wait` is local host-side timing behavior rather than a browser message.
- `batch` is workflow semantics and deserves thought about execution guarantees and output format.
- `computer` appears to be a higher-level orchestration surface rather than a narrow browser primitive.

Recommendation: keep these out of the first parity milestone.

## Recommended Implementation Strategy

The recommended rule set is simple.

1. If the Go router and service worker already support a verb, implement the `surf-go` command first before touching browser-side code.
2. Implement every new public `surf-go` verb as a Glazed command.
3. Reserve `tool-raw` for diagnostics, parity validation, and temporary escape-hatch workflows rather than as the user-facing solution.
4. Do not introduce backward-compatibility shims unless a real caller requires them.
5. Prefer the Node command semantics, but not necessarily every Node alias.
6. Use shared Glazed helpers internally when that reduces duplication, but keep each public command explicit in registration, help, fields, and examples.

### Why Not Just Mirror Node 1:1?

A literal 1:1 mirror would be fast in the short term but messy in the medium term.

- Node has legacy aliases that are product conveniences, not architectural necessities.
- `surf-go` already has a cleaner command architecture with Glazed and Cobra.
- The goal should be semantic parity and ergonomic parity, not line-by-line CLI duplication.

## Detailed Architecture for a New Intern

### The Most Important Mental Model

Think of the system as a pipeline with one decision point per layer.

```text
User CLI text
  -> Go command object decides how to parse flags and arguments
  -> transport client sends a tool_request over the unix socket
  -> Go host router maps logical tool name to extension message type
  -> extension service worker executes Chrome API / DOM / CDP work
  -> response flows back through host to CLI row formatter
```

When something is missing, ask these questions in order.

1. Does the Node CLI expose the user-facing command already?
2. Does the Node host mapping already define the underlying message shape?
3. Does the Go router already map that tool name?
4. Does the service worker already implement the message type?
5. If yes to 2-4, why does `surf-go` not expose it yet?

This diagnostic order avoids unnecessary implementation.

### The Three Most Important Files for Parity Work

#### `native/cli.cjs`

Use this to learn:

- the intended command name,
- expected arguments and flags,
- examples,
- whether a command is really first-class or just an alias.

#### `go/internal/host/router/toolmap.go`

Use this to learn:

- whether Go host support already exists,
- the exact message sent to the extension,
- validation expectations,
- when a tool is local-only versus extension-routed.

#### `src/service-worker/index.ts`

Use this to learn:

- whether the extension actually implements the command,
- which Chrome APIs are touched,
- which response payload shape comes back.

## API Reference: Request and Response Shapes

### Socket Request Envelope from `surf-go`

The Go CLI builds a socket request with `BuildToolRequest` in `go/internal/cli/commands/base.go`.

Conceptually:

```json
{
  "type": "tool_request",
  "method": "execute_tool",
  "params": {
    "tool": "network.curl",
    "args": {
      "id": "r_001"
    }
  },
  "id": "go-...",
  "tabId": 123,
  "windowId": 456
}
```

### Host Router Mapping Contract

`MapToolToMessage` in `go/internal/host/router/toolmap.go` converts the logical tool to an extension message.

Example for `network.curl`:

```go
case "network.curl":
    return base(map[string]any{
        "type": "GET_NETWORK_ENTRY",
        "requestId": firstNonNil(a["id"], a["0"]),
        "formatAsCurl": true,
    }), nil
```

### Service Worker Handler Contract

The service worker receives one message type such as `GET_NETWORK_ENTRY` or `ZOOM_SET` and returns a plain object payload.

Example conceptually:

```ts
case "ZOOM_SET": {
  await chrome.tabs.setZoom(tabId, level);
  return { success: true, zoom: level };
}
```

### CLI Row Shaping

The Go CLI row formatter is deliberately separate from command execution. This matters because some commands should output a single object row, others a list, and others a textual response.

If a new command returns a structured payload, prefer preserving structure rather than collapsing to a text blob.

## Design Decisions

### Decision 1: Use Glazed Commands for All New Public Verbs

Rationale:

- better help output,
- stronger validation,
- easier onboarding,
- easier future extension,
- avoids `--args-json` becoming the permanent public interface,
- keeps the Go CLI architecture coherent instead of mixing Glazed and one-off command styles.

### Decision 2: Keep `tool_raw` as the Escape Hatch

Rationale:

- allows incremental rollout,
- enables manual testing before first-class commands exist,
- reduces pressure to implement every edge case immediately.

### Decision 3: Prefer Existing Go Router Semantics Over Re-Importing Node Logic

Rationale:

- the Go router is already the Go host's canonical mapping layer,
- duplicating Node helper logic in the Go CLI would create a second mapping source of truth.

### Decision 4: Implement in Vertical Slices, Not by Group Name Alone

Rationale:

Users benefit most from a few fully usable command slices rather than dozens of half-finished wrappers.

Recommended first slice:

- `js`
- `upload`
- `form.fill`
- `locate.role`
- `locate.text`
- `locate.label`
- `wait.load`

This slice unlocks many actual automation workflows.

### Decision 5: Use a Shared Glazed Command Authoring Contract

Rationale:

- new interns need repeatable implementation structure,
- command reviews become easier when all commands follow the same template,
- it reduces drift in defaults, sections, output, and validation behavior.

## Proposed Command Architecture in Go

### Command Authoring Contract

Every new public command in this ticket should satisfy the same structural contract.

- Define a real Glazed command constructor in `go/internal/cli/commands`.
- Use `cmds.CommandDescription` to declare the command name, arguments, flags, examples, and sections.
- Expose transport settings through Glazed sections instead of ad hoc environment-variable reads.
- Use a typed settings struct with `glazed.parameter` tags where practical.
- Implement `RunIntoGlazeProcessor` so the command participates in the same output pipeline as the rest of `surf-go`.
- Keep tool-name mapping in one place inside the command implementation rather than pushing mapping logic into `main.go`.
- Add command-level tests covering required inputs, normalization, and request shaping.
- Register the command in `go/cmd/surf-go/main.go` through the same builder path used by existing first-class commands.

Conceptually:

```text
New command request
  -> create command file in go/internal/cli/commands
  -> define settings struct and command description
  -> implement RunIntoGlazeProcessor
  -> add registration in go/cmd/surf-go/main.go
  -> add focused tests
  -> validate with tool-raw and then real command invocation
```

### Pattern A: Typed Single-Tool Command

Use this for `js`, `upload`, `zoom`, `wait.load`, `history.search`.

Pseudocode:

```go
func NewJSCommand() (*JSCommand, error) {
    desc := cmds.NewCommandDescription(
        "js",
        cmds.WithFlags(
            fields.New("file", fields.TypeString, ...),
            fields.New("socket-path", fields.TypeString, ...),
            fields.New("tab-id", fields.TypeInteger, ...),
            fields.New("window-id", fields.TypeInteger, ...),
        ),
        cmds.WithArguments(
            fields.New("code", fields.TypeString, ...),
        ),
    )
    return &JSCommand{CommandDescription: desc}, nil
}

func (c *JSCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    settings := &JSSettings{}
    decode(vals, settings)
    code := settings.Code
    if settings.File != "" {
        code = readFile(settings.File)
    }
    resp, err := ExecuteTool(ctx, client, "js", map[string]any{"code": code}, tabID, windowID)
    if err != nil { return err }
    return emitRows(gp, resp)
}
```

### Pattern B: Typed Multi-Flag Tool Family

Use this for `locate.role`, `locate.text`, `locate.label`, because the verb family is cohesive and the options are meaningful.

Possible shape:

```text
surf-go locate role button --name "Submit" --action click
surf-go locate text "Accept" --exact --action click
surf-go locate label "Username" --action fill --value "john"
```

This is better than a flat command list because it teaches the conceptual grouping explicitly.

### Pattern C: Thin Shared Glazed Wrapper

Use a shared Glazed helper when:

- the tool has no complex custom validation,
- the tool takes only one or two simple flags,
- the output is already well-shaped,
- a bespoke command file would mostly duplicate other tiny commands.

Even in this case, the public command should still have:

- an explicit Glazed command constructor,
- declared fields and sections,
- command-specific examples,
- tests proving the wrapped tool name and args.

Candidates:

- `bookmark.remove`
- `perf.stop`
- `tab.groups`

## Recommended CLI Shape Additions

### New Top-Level or Grouped Commands

```text
surf-go js
surf-go locate role
surf-go locate text
surf-go locate label
surf-go upload
surf-go form fill
surf-go perf start
surf-go perf stop
surf-go perf metrics
surf-go history list
surf-go history search
surf-go bookmark add
surf-go bookmark remove
surf-go bookmark list
surf-go zoom
surf-go resize
```

### Existing Groups to Expand

- `wait`
  - add `load`
- `tab`
  - add `unname`, `group`, `ungroup`, `groups`
- `network`
  - add `curl`, `path`
- `scroll`
  - consider creating a real group rather than relying only on root-level `scroll`

## Proposed Directory and File-Level Work Plan

### Phase 0: Freeze the Gap Inventory

Files:

- `sources/01-node-vs-go-non-provider-command-gap.json`
- `design-doc/...` (this document)
- `tasks.md`

Goal:

- make the missing-command inventory explicit and durable.

### Phase 1: Add Typed Foundations

Files likely to add:

- `go/internal/cli/commands/js.go`
- `go/internal/cli/commands/upload.go`
- `go/internal/cli/commands/form_fill.go`
- `go/internal/cli/commands/locate.go`
- `go/internal/cli/commands/wait_load.go`
- shared helper additions under `go/internal/cli/commands` if needed for Glazed field reuse

Files to update:

- `go/cmd/surf-go/main.go`

### Phase 2: Expose Existing Routed Tools

Files likely to update:

- `go/cmd/surf-go/main.go`
- `go/internal/cli/commands/network_curl.go`
- `go/internal/cli/commands/network_path.go`
- `go/internal/cli/commands/tab_group.go`
- `go/internal/cli/commands/tab_unname.go`
- `go/internal/cli/commands/scroll_group.go`

Commands:

- `network.curl`
- `network.path`
- `tab.unname`
- `tab.group`
- `tab.ungroup`
- `tab.groups`
- `scroll.top`
- `scroll.bottom`
- `scroll.to`
- `scroll.info`

### Phase 3: Add Browser Utility Surface

Files likely to add:

- `go/internal/cli/commands/perf.go`
- `go/internal/cli/commands/zoom.go`
- `go/internal/cli/commands/resize.go`
- `go/internal/cli/commands/bookmark.go`
- `go/internal/cli/commands/history.go`

Commands:

- `zoom`
- `resize`
- `bookmark.add`
- `bookmark.remove`
- `bookmark.list`
- `history.list`
- `history.search`
- `perf.start`
- `perf.stop`
- `perf.metrics`

### Phase 4: Decide on Composite UX Commands

Commands:

- `batch`
- convenience aliases/composites

This phase should only happen after the underlying primitives are solid.

## Example Implementation Slice

Here is the recommended first week for a new intern.

### Slice 1: `js`

Reason:

- easy to validate,
- very useful,
- already routed,
- teaches typed command pattern.

Checklist:

- add `go/internal/cli/commands/js.go`
- support positional `code`
- support `--file`
- wire into root command
- add tests

### Slice 2: `upload`

Reason:

- high user value,
- already mapped in Node,
- requires slightly more validation and argument normalization.

Checklist:

- typed flags: `--ref`, `--files`
- normalize comma-separated file list
- validate at least one file
- execute tool `upload`

### Slice 3: `locate.*`

Reason:

- unlocks semantic automation,
- establishes grouped typed subcommands.

Checklist:

- `locate role <role> [--name --action --value --all]`
- `locate text <text> [--exact --action --value]`
- `locate label <label> [--action --value]`

### Slice 4: `form.fill` and `wait.load`

Reason:

- common follow-ups after location and interaction,
- low conceptual overhead.

## Testing Strategy

### Unit Tests

Add command-focused tests for:

- required argument validation,
- file loading behavior (`js --file`),
- normalization (`upload --files`),
- request envelope fields,
- row shaping of structured outputs.

Likely locations:

- `go/internal/cli/commands/*_test.go`
- `go/cmd/surf-go/integration_test.go`

### Router-Level Safety Check

Do not modify router mappings unless a command truly requires host-side changes. If you do touch `toolmap.go`, add or update tests there too.

### Manual Validation Against Real Browser Session

Use `tool-raw` first to confirm the host/service-worker path before adding the command.

Examples:

```bash
cd go
SURF_SOCKET_PATH="$HOME/snap/chromium/common/surf-cli/surf.sock" \
  go run ./cmd/surf-go tool-raw --tool js --args-json '{"code":"return document.title"}'

SURF_SOCKET_PATH="$HOME/snap/chromium/common/surf-cli/surf.sock" \
  go run ./cmd/surf-go tool-raw --tool locate.role --args-json '{"role":"button","all":true}'
```

Then validate the first-class command:

```bash
go run ./cmd/surf-go js "return document.title"
go run ./cmd/surf-go locate role button --all
```

### Recommended Review Loop

```text
1. Confirm tool exists in Node CLI catalog.
2. Confirm Go router maps it.
3. Confirm service worker implements backing message type.
4. Add Go command.
5. Add tests.
6. Validate with tool-raw.
7. Validate with first-class command.
8. Update docs/help if semantics differ.
```

## Detailed Implementation Notes for "Everything Is Glazed"

The phrase "make them all Glazed" needs to be interpreted precisely so an intern does not over-engineer or under-deliver.

It does not require every command to have a completely unique internal execution engine. Shared execution helpers are good. Shared field builders are good. Shared output-shaping utilities are good. What it does require is that every public command be visible to the CLI as a real Glazed command with first-class documentation and field metadata.

Practical implications:

- `go/cmd/surf-go/main.go` should register Glazed commands, not anonymous Cobra handlers that manually parse arguments.
- Each command should participate in Glazed help and output formatting.
- Shared helper functions should live below the command boundary, not replace the boundary.
- If a command remains only reachable through `tool-raw`, it is not done for this ticket.

Recommended internal layering:

```text
command file
  -> settings struct
  -> command description
  -> shared helper for request construction
  -> base ExecuteTool call
  -> shared emitRows/output shaping helper
```

Recommended review checklist for each command:

- Does the command appear in `surf-go --help` in the right group?
- Are all user-facing parameters declared as Glazed fields or arguments?
- Are transport settings wired through Glazed sections?
- Does the command avoid direct environment-variable configuration?
- Does help text mirror Node semantics where appropriate?
- Does the output preserve structured data rather than stringify it?
- Is there a test that proves the tool name and arg map sent to the host?

## Risks and Sharp Edges

### Risk 1: Implementing Aliases Before Primitives

This creates noise and little value. Avoid it.

### Risk 2: Overusing `--args-json`

This hides missing design decisions behind generic plumbing. It is useful as a stopgap, not as a final UX strategy.

### Risk 3: Duplicating Mapping Logic in the CLI Layer

The CLI should not remap message types itself. It should call `ExecuteTool` with the canonical tool name and let `toolmap.go` stay authoritative.

### Risk 4: Output Shape Drift

Some commands in Node print user-oriented text while Go currently leans toward structured rows. Do not accidentally destroy useful structure for the sake of text parity.

Recommended policy:

- preserve structured data where it exists,
- add user-friendly formatting only where it clearly helps,
- be consistent within command families.

### Risk 5: Group Naming and Command Naming Drift

If you introduce a Go-only naming scheme, users will have to learn two CLIs. Prefer Node-compatible command names unless there is a strong reason not to.

## Alternatives Considered

### Alternative A: Expose Everything Through `tool-raw`

Rejected.

Reason:

- poor UX,
- weak validation,
- bad discoverability,
- not acceptable as the main CLI surface.

### Alternative B: Auto-Generate Go Commands From the Node Catalog

Rejected for now.

Reason:

- attractive in theory, but the Node catalog mixes product help, aliases, and behavior conventions.
- generation would still require manual decisions for argument typing, grouping, and structured output.

### Alternative C: Port All Missing Verbs as Generic Wrappers First

Partially accepted only for low-priority commands.

Reason:

- viable for low-risk exposure,
- insufficient for core user-facing verbs.

## Concrete Implementation Plan

### Milestone 1: Workflow Unlockers

Implement first-class commands for:

- `js`
- `upload`
- `form.fill`
- `locate.role`
- `locate.text`
- `locate.label`
- `wait.load`

Definition of done:

- commands wired,
- tests added,
- verified through real browser session,
- help text is explicit and copy-paste friendly.

### Milestone 2: Inspection and Utility Parity

Implement:

- `network.curl`
- `network.path`
- `tab.unname`
- `tab.group`
- `tab.ungroup`
- `tab.groups`
- `scroll.top`
- `scroll.bottom`
- `scroll.to`
- `scroll.info`

### Milestone 3: Browser Bookkeeping and Metrics

Implement:

- `perf.start`
- `perf.stop`
- `perf.metrics`
- `zoom`
- `resize`
- `bookmark.add`
- `bookmark.remove`
- `bookmark.list`
- `history.list`
- `history.search`

### Milestone 4: Optional Composite Commands

Only after the previous milestones feel solid:

- `batch`
- selected convenience aliases/composites

## Onboarding Guide for the Intern

If you are new to the repository, do not start by coding.

Read in this order:

1. `native/cli.cjs:630-1210`
2. `native/host-helpers.cjs:650-1035`
3. `go/internal/host/router/toolmap.go:133-425`
4. `go/cmd/surf-go/main.go:20-267`
5. `go/internal/cli/commands/tool_simple.go:34-114`
6. `go/internal/cli/commands/navigate.go`
7. `go/internal/cli/commands/chatgpt.go`
8. `src/service-worker/index.ts:2501-2615`

Then answer these questions for the command you are implementing.

- Is it already in Node CLI help?
- Is it already in Go router mapping?
- Is it already in the service worker?
- Is it worth a typed command?
- What should its output look like in Go?

If you cannot answer all five questions, do not start implementing yet.

## Open Questions

1. Should `scroll.*` become a dedicated grouped command family in `surf-go`, or remain root-level verbs for parity with current simple action placement?
2. Should `batch` remain deferred until there is a broader workflow story for Go, or is a minimal `BATCH_EXECUTE` wrapper sufficient?
3. Should Node-style aliases such as `read` and `find` be added to `surf-go`, or intentionally omitted to keep the Go CLI smaller?
4. Should some user-facing Node text formatting be reproduced in Go, or should Go remain primarily structured-output first?

## References

### Key Source Files

- `native/cli.cjs:630-1210`
- `native/host-helpers.cjs:650-1035`
- `go/internal/host/router/toolmap.go:133-425`
- `src/service-worker/index.ts:2501-2615`
- `go/cmd/surf-go/main.go:20-267`
- `go/internal/cli/commands/tool_simple.go:34-114`
- `go/internal/cli/commands/navigate.go`
- `go/internal/cli/commands/chatgpt.go`
- `go/internal/cli/transport/client.go`
- `sources/01-node-vs-go-non-provider-command-gap.json`

### Useful Validation Commands

```bash
cd go && go test ./internal/cli/commands ./cmd/surf-go ./cmd/surf-host-go
cd go && go run ./cmd/surf-go tool-raw --tool js --args-json '{"code":"return document.title"}'
cd go && go run ./cmd/surf-go tool-raw --tool locate.role --args-json '{"role":"button","all":true}'
```
