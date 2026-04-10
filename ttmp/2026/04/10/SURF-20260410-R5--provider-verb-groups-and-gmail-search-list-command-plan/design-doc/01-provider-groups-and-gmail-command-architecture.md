---
Title: Provider groups and Gmail command architecture
Ticket: SURF-20260410-R5
Status: active
Topics:
    - surf-go
    - glazed
    - cli
    - browser-automation
    - gmail
    - chatgpt
    - kagi
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: Root command registration is where provider groups will be introduced
    - Path: go/internal/cli/commands/chatgpt.go
      Note: Current ChatGPT ask command implementation that will move under the chatgpt provider group
    - Path: go/internal/cli/commands/chatgpt_transcript.go
      Note: Current ChatGPT transcript implementation that will move under the chatgpt provider group
    - Path: go/internal/cli/commands/kagi_assistant.go
      Note: Current Kagi assistant implementation that will move under the kagi provider group
    - Path: go/internal/cli/commands/kagi_search.go
      Note: Current Kagi search implementation that will move under the kagi provider group
    - Path: go/pkg/doc/tutorials/01-building-browser-side-verbs.md
      Note: Existing browser-side command playbook that the Gmail workflow tutorial will build on
ExternalSources: []
Summary: Group top-level provider verbs under provider namespaces and add the first Gmail command family with search and inbox listing.
LastUpdated: 2026-04-10T10:00:00-04:00
WhatFor: Plan the surf-go CLI refactor from flat provider verbs to grouped provider namespaces and define the first Gmail browser-side commands.
WhenToUse: Use when implementing or reviewing surf-go provider grouping, Gmail command design, command registration changes, and related help/test updates.
---


# Provider groups and Gmail command architecture

## Executive Summary

The current `surf-go` CLI has accumulated provider-specific verbs directly at the root: `chatgpt`, `chatgpt-transcript`, `kagi-search`, and `kagi-assistant`. That surface is workable for a small number of commands, but it does not scale. The next step is to introduce provider groups so the CLI reads as `surf-go chatgpt ...`, `surf-go kagi ...`, and `surf-go gmail ...`. The immediate functional work under that structure is to add the first Gmail family: `surf-go gmail search` and `surf-go gmail list --inbox`.

This ticket deliberately combines the namespace refactor and Gmail design because Gmail should not be added as another top-level verb. The grouped command structure needs to be in place first so the new family lands in the correct long-term shape.

The recommended implementation is:

- create true Cobra provider groups at the root, not aliases or hidden adapters
- move ChatGPT and Kagi commands underneath those groups
- keep existing browser-side command implementation patterns: embedded JS, shared fetch helpers, dual-mode Glazed output where appropriate, explicit tab ownership/cleanup
- add a new Gmail help/tutorial document focused on stateful browser workflows, because Gmail is more stateful than Kagi search or transcript export

## Problem Statement

The flat root command surface has three concrete problems.

First, discoverability degrades as provider-specific verbs accumulate. `surf-go kagi-search` and `surf-go kagi-assistant` already form an obvious conceptual family, but the current shape does not reflect that family in help output or shell completion.

Second, command naming becomes inconsistent once a provider grows more than one feature. `chatgpt` is an action verb, while `chatgpt-transcript` is a noun-ish export verb. `kagi-search` and `kagi-assistant` are provider-prefixed but still root-level. That inconsistency makes future commands harder to name and harder to predict.

Third, Gmail is not a good fit for the current flat surface. Gmail is likely to grow a family of related verbs: listing inbox threads, searching threads, exporting messages, downloading attachments, and potentially more stateful workflows later. It should start under `surf-go gmail` from the beginning.

## Goals

The goals for this ticket are:

- replace flat ChatGPT and Kagi root verbs with provider groups
- define a stable grouped command tree for at least `chatgpt`, `kagi`, and `gmail`
- implement `surf-go gmail search`
- implement `surf-go gmail list --inbox`
- keep the commands as real Glazed commands under grouped Cobra parents
- preserve the existing browser-side command authoring pattern: research with `surf-go js`, embed production JS with `go:embed`, parse once in Go, expose Markdown and/or structured rows as appropriate
- update help documentation so grouped command authoring and stateful workflow design are discoverable

## Non-Goals

The following are explicitly out of scope for this ticket:

- implementing Gmail attachment download
- implementing Gmail message export bodies beyond the search/list surface
- adding backwards-compatibility adapters for the old flat verb names unless explicitly requested later
- implementing provider groups for every future provider in one pass
- changing the underlying host protocol unless Gmail proves the `js` path insufficient

The no-backwards-compatibility point matters. The repository guidance explicitly says not to add adapters unless asked. The correct default is to move the CLI to the grouped shape directly and update tests/help accordingly.

## Current State

Current root-level provider commands are registered directly in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go):

- `chatgpt`
- `chatgpt-transcript`
- `kagi-search`
- `kagi-assistant`

These commands are implemented as Glazed commands in `go/internal/cli/commands` and built into Cobra commands in the root builder. The browser-side playbook now exists in [01-building-browser-side-verbs.md](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/pkg/doc/tutorials/01-building-browser-side-verbs.md), and it reflects the current patterns for `chatgpt-transcript`, `kagi-search`, and `kagi-assistant`.

That means the codebase already has the right low-level primitives for grouped provider commands. The missing piece is CLI structure and the Gmail-specific browser workflow research.

## Proposed CLI Structure

The proposed provider tree is:

```text
surf-go
  chatgpt
    ask
    transcript
    models
  kagi
    search
    assistant
  gmail
    list
    search
```

Notes:

- `chatgpt ask` is the grouped replacement for the current `chatgpt` command.
- `chatgpt transcript` is the grouped replacement for `chatgpt-transcript`.
- `chatgpt models` is optional in the first pass; if implemented, it should map cleanly to model-listing behavior rather than forcing users to remember `--list-models` on `ask`.
- `kagi search` replaces `kagi-search`.
- `kagi assistant` replaces `kagi-assistant`.
- `gmail list --inbox` is the first list-oriented read path.
- `gmail search --query <...>` is the first explicit search path.

This tree keeps provider nouns at level one and task verbs at level two. That separation is more readable and scales better.

## Design Decisions

### Decision 1: Use true provider groups, not prefixed command names

Provider groups should be real Cobra parent commands. The grouped structure should not be simulated with dashed names such as `kagi-search` or `chatgpt-transcript`.

Rationale:

- help output and shell completion become predictable
- additional provider verbs stop polluting the root
- the structure matches how users mentally model the features

### Decision 2: Do not add backward-compatibility aliases by default

The repo guidance says not to add adapters unless requested. The ticket should therefore assume a clean move to grouped verbs.

Rationale:

- avoids duplicate help entries and maintenance burden
- avoids test duplication
- forces the codebase into one stable naming convention

If the user later asks for deprecated aliases, that should be a separate explicit decision.

### Decision 3: Gmail commands should start read-only

The first Gmail commands should avoid destructive or persistent mutations.

`gmail list --inbox` should list inbox-visible threads.
`gmail search --query ...` should perform Gmail search and return thread/message summary rows.

Rationale:

- Gmail is a high-value, high-risk surface
- read-only commands let us validate selectors, page states, and scaling concerns before adding export/download or attachment workflows

### Decision 4: Gmail should use the same browser-side command pattern first

The first Gmail pass should use the same overall implementation model as the other complex verbs:

- research with `surf-go js`
- production JS embedded with `go:embed`
- shared fetch helper in Go
- dual-mode output where appropriate
- mock-host integration tests plus live browser validation

Rationale:

- keeps architecture consistent
- reduces one-off protocols
- Gmail list/search should be feasible with DOM-based extraction first

If Gmail later requires a hybrid DOM/network strategy, that can be layered on after the first verbs exist.

## Gmail Command Design

### `surf-go gmail list --inbox`

Intent:

- show inbox thread rows from the current Gmail session

Entry condition:

- the browser session is already logged into Gmail
- Gmail web UI is reachable

Page action:

- open or reuse Gmail
- ensure the inbox view is active when `--inbox` is set
- wait for the thread list to stabilize

Extraction target:

- thread row summaries, not full message bodies

Suggested row fields:

- `index`
- `threadId` if available from DOM attributes or row links
- `subject`
- `participants`
- `snippet`
- `timestampText`
- `unread`
- `starred`
- `hasAttachment`
- `href`
- `mailbox`

Flags to design initially:

- `--inbox` as an explicit mode switch for the initial implementation
- `--max-results`
- `--tab-id`
- `--window-id`
- `--keep-tab-open`
- `--debug-socket`

Default output:

- structured rows are likely the primary useful output here, but a short Markdown inbox summary may still be worth supporting via dual-mode if the row fields are readable enough.

### `surf-go gmail search`

Intent:

- run a Gmail search query and extract matching thread/message summary rows

Entry condition:

- the browser session is already logged into Gmail

Page action:

- open or reuse Gmail
- focus the Gmail search box
- submit the user query
- wait for results view stabilization

Extraction target:

- Gmail search result rows, again at thread/message summary level rather than full thread export

Suggested row fields:

- `index`
- `query`
- `threadId`
- `subject`
- `participants`
- `snippet`
- `timestampText`
- `labels`
- `unread`
- `hasAttachment`
- `href`

Flags to design initially:

- `--query`
- `--max-results`
- `--tab-id`
- `--window-id`
- `--keep-tab-open`
- `--debug-socket`

### Gmail state and risk notes

Gmail is more stateful than Kagi search or transcript export. The design must explicitly account for:

- logged-out pages
- account chooser/interstitial pages
- keyboard-focus differences in the Gmail search box
- split-pane or preview-pane UI variants
- pagination or lazy loading in the inbox list
- row selectors that differ between inbox and search states

These are exactly why a separate Gmail/stateful help document should be added as part of this work.

## Implementation Strategy

### Phase 1: Group existing provider verbs

Create Cobra parent groups under the root:

- `chatgpt`
- `kagi`
- `gmail`

Under `chatgpt`, move:

- current `chatgpt` implementation to `ask`
- current `chatgpt-transcript` implementation to `transcript`

Under `kagi`, move:

- current `kagi-search` implementation to `search`
- current `kagi-assistant` implementation to `assistant`

This phase is primarily root-command and help-surface work. The underlying Glazed command implementations can usually stay intact; the main change is how they are registered into Cobra.

### Phase 2: Research Gmail list/search DOM and workflow

Before writing command code:

- create a ticket-local ordered script trail under `scripts/`
- probe Gmail inbox row selectors
- probe Gmail search box selectors and submission behavior
- probe differences between inbox rows and search result rows
- validate whether stable thread links or IDs are accessible in DOM

This phase should leave behind:

- a diary of working and rejected selectors
- research scripts with ordered numeric prefixes
- a decision on whether list and search can share one embedded extractor with mode flags or need separate scripts

### Phase 3: Implement `gmail list`

Create a new Glazed command in `go/internal/cli/commands`.

The fetch path should:

- acquire or create a Gmail tab
- navigate to inbox if needed
- remember tab ownership
- run embedded JS
- parse once
- close owned tab unless `--keep-tab-open`

The browser-side JS should:

- verify Gmail is loaded and authenticated
- locate inbox rows
- normalize row fields
- return a structured object with counts and rows

### Phase 4: Implement `gmail search`

Create a second Glazed command or a shared Gmail command family if the list/search behavior can be parameterized cleanly.

The browser-side logic should:

- interact with the Gmail search box safely
- submit the query
- wait for results stabilization
- extract result rows

### Phase 5: Documentation and playbooks

Update help and add a second tutorial focused on more stateful browser workflows, using Gmail as the motivating example.

That tutorial should cover:

- page states and transitions
- safe handling of account/interstitial states
- read-only vs mutating workflows
- pagination and batching concerns
- cleanup and ownership rules

## Alternatives Considered

### Alternative A: Keep flat verbs and only add Gmail group

Rejected because it would leave the CLI inconsistent. Gmail would be grouped while ChatGPT and Kagi remained flat, which is the worst of both worlds.

### Alternative B: Add hidden aliases for all old names

Rejected for now because the repo guidance says not to add compatibility adapters unless explicitly requested.

### Alternative C: Implement Gmail with a brand-new host protocol instead of `js`

Rejected for the first pass. That would add architectural complexity before we know whether DOM extraction is insufficient.

## Detailed Implementation Plan

1. Refactor root command registration in [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go).
   - create provider group parents
   - move ChatGPT and Kagi children under the correct parent
   - keep dual-mode command registration intact

2. Rename user-facing command paths without unnecessary rewrites.
   - `chatgpt` command implementation becomes the subcommand `ask`
   - `chatgpt-transcript` becomes `chatgpt transcript`
   - `kagi-search` becomes `kagi search`
   - `kagi-assistant` becomes `kagi assistant`

3. Update help docs and examples.
   - root help should show provider groups instead of the current flat commands
   - the browser-side playbook should be linked from the new Gmail/stateful tutorial

4. Research Gmail DOM and workflows using `surf-go js`.
   - start with inbox listing probes
   - then search submission probes
   - store all scripts under the new ticket `scripts/` folder with numeric prefixes

5. Implement `gmail list` as a Glazed command.
   - embedded JS with `go:embed`
   - shared fetch helper with explicit tab ownership
   - row shaping and optional writer output

6. Implement `gmail search` as a Glazed command.
   - decide whether it shares most code with `gmail list` or deserves a separate command path
   - ensure query submission is robust in the real Gmail UI

7. Add tests.
   - unit tests for script prelude generation, parsing, and row shaping
   - mock-host integration tests for `tab.new` -> `js` -> `tab.close` sequences
   - live browser validation against real Gmail session states

8. Add documentation.
   - create the new Gmail/stateful help document
   - update command examples and grouped help references

## API and File Reference Map

Primary files likely to change:

- [main.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go)
- [chatgpt.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt.go)
- [chatgpt_transcript.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt_transcript.go)
- [kagi_search.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_search.go)
- [kagi_assistant.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_assistant.go)
- future Gmail command files under `go/internal/cli/commands/`
- future Gmail scripts under `go/internal/cli/commands/scripts/`
- [01-building-browser-side-verbs.md](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/pkg/doc/tutorials/01-building-browser-side-verbs.md)

Likely new files:

- `go/internal/cli/commands/gmail_list.go`
- `go/internal/cli/commands/gmail_search.go`
- `go/internal/cli/commands/scripts/gmail_list.js`
- `go/internal/cli/commands/scripts/gmail_search.js`
- `go/internal/cli/commands/gmail_list_test.go`
- `go/internal/cli/commands/gmail_search_test.go`
- new help/tutorial document under `go/pkg/doc/tutorials/`

## Open Questions

1. Should `gmail list` default to inbox if no mode flag is given, or should `--inbox` be required in the first implementation?
2. Should `chatgpt models` be introduced immediately as part of the grouping cleanup, or can model listing remain a flag-only behavior under `chatgpt ask` for now?
3. For Gmail, is dual-mode output worthwhile for both commands, or should one of them be structured-first if Markdown adds little value?
4. Are Gmail row selectors stable enough across inbox density/view settings, or do we need a capability check and a clearer failure mode in v1?

## Recommended Execution Order

1. Group ChatGPT and Kagi verbs under provider parents.
2. Update help and tests for the new CLI tree.
3. Create the Gmail research scripts and diary.
4. Implement `gmail list --inbox` first.
5. Implement `gmail search` second.
6. Add the Gmail/stateful workflow tutorial.

That order minimizes risk. It lands the namespace refactor first, then Gmail enters the tree in the correct place.
