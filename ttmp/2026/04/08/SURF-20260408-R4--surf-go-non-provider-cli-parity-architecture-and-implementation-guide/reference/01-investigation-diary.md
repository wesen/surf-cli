---
Title: Investigation Diary
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
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-go/main.go
      Note: |-
        Current Go CLI registration surface used to compute gaps
        Evidence for current registration gaps
    - Path: go/internal/host/router/toolmap.go
      Note: |-
        Go router evidence showing host support already exists for many missing verbs
        Evidence that many non-provider tools already route in Go
    - Path: native/cli.cjs
      Note: |-
        Baseline Node CLI command catalog used to identify missing non-provider verbs
        Evidence for missing command inventory
    - Path: native/host-helpers.cjs
      Note: |-
        Node mapping layer used to verify browser message contracts
        Evidence for extension message contracts
    - Path: sources/01-node-vs-go-non-provider-command-gap.json
      Note: Generated gap inventory for this ticket
    - Path: src/service-worker/index.ts
      Note: |-
        Browser-side implementation evidence for zoom, bookmarks, history, and related commands
        Evidence for existing browser-side handlers
ExternalSources: []
Summary: Chronological evidence log for the non-provider surf-go parity analysis ticket.
LastUpdated: 2026-04-08T22:05:00-04:00
WhatFor: Preserve the exact evidence and command trace used to write the implementation guide.
WhenToUse: Use when resuming implementation or reviewing why specific phases and priorities were chosen.
---


# Investigation Diary

## Goal

Create a new ticket with a detailed, intern-oriented analysis and implementation guide for closing the non-provider command gap between the Node `surf` CLI and `surf-go`, then prepare that ticket for reMarkable delivery.

## Step 1: Establish the ticket and scope

The request followed a command-gap review where provider-specific verbs had already been excluded by direction. That narrowed the scope substantially. The point of this ticket is not “complete Node feature parity,” but “complete non-provider CLI parity with a sane implementation plan that a new engineer can follow.”

### Prompt Context

**User prompt (verbatim):** "alright, create a new ticket and a detailed analysis and implementation guide. 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a fresh docmgr ticket focused on the missing non-provider `surf-go` command surface and write a detailed architecture/design/implementation guide suitable for a new intern.

**Inferred user intent:** Produce a durable onboarding-quality artifact, not just an action list or a short parity inventory.

### What I did

- Reviewed the relevant ticket/doc workflow skills:
  - `/home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md`
  - `/home/manuel/.codex/skills/remarkable-upload/SKILL.md`
  - `/home/manuel/.codex/skills/docmgr/SKILL.md`
- Created a new ticket:
  - `docmgr ticket create-ticket --ticket SURF-20260408-R4 --title "surf-go Non-Provider CLI Parity Architecture and Implementation Guide" --topics go,glazed,chromium,cli,native-messaging,architecture,onboarding`
- Added primary docs:
  - design doc
  - investigation diary

### Why

- A dedicated ticket keeps this work independent from the earlier Go-host rollout and provider-integration tickets.
- The scope is architectural/documentary at this stage, so it deserves its own artifact set rather than being buried inside a previous implementation diary.

### What worked

- `docmgr` created a clean workspace under:
  - `ttmp/2026/04/08/SURF-20260408-R4--surf-go-non-provider-cli-parity-architecture-and-implementation-guide`
- Design-doc and diary scaffolding were created without issue.

### What didn't work

- `docmgr ticket list --ticket SURF-2026` returned no results because the filter was too broad and did not match the exact ticket ID format being used.
- Resolution: ran `docmgr ticket list` without the partial filter.

### What I learned

- The existing ticket history is clustered around 2026-02-25, so creating a new date-scoped ticket made the continuation boundary much cleaner.

### What was tricky to build

- The tricky part was not the ticket creation itself. It was defining a scope that is narrow enough to be implementable and broad enough to onboard a new engineer properly.

### What warrants a second pair of eyes

- The final scope boundary, especially whether `batch` and alias/composite commands belong in this ticket or in a follow-up.

### What should be done in the future

1. Keep provider-specific parity work in the older provider ticket.
2. Keep the new ticket focused on browser/tooling parity and command architecture.

### Code review instructions

- Review ticket scaffold contents under:
  - `ttmp/2026/04/08/SURF-20260408-R4--surf-go-non-provider-cli-parity-architecture-and-implementation-guide/`

### Technical details

```bash
docmgr ticket create-ticket --ticket SURF-20260408-R4 --title "surf-go Non-Provider CLI Parity Architecture and Implementation Guide" --topics go,glazed,chromium,cli,native-messaging,architecture,onboarding
docmgr doc add --ticket SURF-20260408-R4 --doc-type design-doc --title "surf-go Non-Provider CLI Parity: Detailed Architecture and Implementation Guide"
docmgr doc add --ticket SURF-20260408-R4 --doc-type reference --title "Investigation Diary"
```

## Step 2: Gather architecture evidence and generate the gap inventory

This step established the evidence that drives the guide. The core question was simple: which commands are missing from `surf-go`, and for how many of them is the host/service-worker support already present?

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build an evidence-backed architectural model of the command system before writing any recommendations.

**Inferred user intent:** Avoid speculative design. Anchor the guide to actual file-level behavior.

### What I did

- Inspected the Node CLI help/catalog layer:
  - `native/cli.cjs:620-1210`
- Inspected the Node host mapping layer:
  - `native/host-helpers.cjs:650-1035`
- Inspected the Go router mapping layer:
  - `go/internal/host/router/toolmap.go:120-430`
- Inspected service-worker implementations for representative commands:
  - `src/service-worker/index.ts:2480-2615`
- Inspected current `surf-go` command registration:
  - `go/cmd/surf-go/main.go:1-280`
- Inspected the generic Go command pattern:
  - `go/internal/cli/commands/tool_simple.go:1-114`
- Generated a machine-readable source artifact:
  - `sources/01-node-vs-go-non-provider-command-gap.json`

### Why

- The command system spans at least four files/layers. Without explicitly walking all four, it is too easy to misclassify a command as “missing from the system” when it is only missing from the Go CLI surface.

### What worked

- The evidence clearly showed that many missing verbs are already mapped in Go router code and, in several cases, already implemented in the service worker.
- The generated JSON artifact made the gap inventory durable and reviewable.

### What didn't work

- N/A.

### What I learned

- The biggest parity gap is command exposure and UX, not host capability.
- `toolmap.go` is ahead of `surf-go/main.go` for a meaningful number of commands.

### What was tricky to build

- The Node CLI catalog includes aliases, conveniences, and real first-class commands in one place. The guide therefore had to separate “commands worth implementing first” from “commands that merely exist in Node.”

### What warrants a second pair of eyes

- The classification of commands into high-value typed commands versus lower-priority wrappers.

### What should be done in the future

1. Keep regenerating the gap inventory if the command surface changes during implementation.
2. Consider adding a repo script for the inventory instead of the one-off inline generator used here.

### Code review instructions

- Review the generated inventory:
  - `sources/01-node-vs-go-non-provider-command-gap.json`
- Cross-check against:
  - `native/cli.cjs`
  - `go/cmd/surf-go/main.go`
  - `go/internal/host/router/toolmap.go`

### Technical details

```bash
python - <<'PY'
# compared native/cli.cjs ALL_SOCKET_TOOLS against go/cmd/surf-go/main.go registrations
# excluded providers and a small set of auxiliary tools
PY
```

## Step 3: Write the implementation guide and phase plan

The design document was written as an onboarding artifact first and an implementation plan second. That ordering mattered. A new intern needs to understand the request pipeline and source-of-truth boundaries before they start adding commands.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Produce a long-form document with architecture narrative, diagrams, pseudocode, API references, file references, and a phased implementation plan.

**Inferred user intent:** Make the ticket self-contained enough that someone unfamiliar with the repo can start productive work from the document alone.

### What I did

- Wrote the design doc with sections covering:
  - executive summary
  - scope
  - architecture walkthrough
  - request-flow diagram
  - current-state evidence
  - command-gap analysis
  - implementation categories
  - API envelope examples
  - typed-command design patterns
  - phased plan
  - onboarding checklist
- Wrote a task backlog in `tasks.md` organized by phases and priority.
- Updated `index.md` and `changelog.md` to reflect the ticket’s actual purpose and current contents.

### Why

- A ticket bundle is more useful when the index, tasks, changelog, and primary design doc all point at the same intended workflow.

### What worked

- The guide now explains the system from CLI surface down to Chrome API execution.
- The guide explicitly distinguishes typed-command work from generic-wrapper work.
- The task list is structured so an intern can execute the work in vertical slices.

### What didn't work

- N/A.

### What I learned

- The onboarding value comes from explicit architectural boundaries, not just detailed command inventories.

### What was tricky to build

- The guide needed to stay detailed without collapsing into an unstructured dump of file excerpts. The solution was to organize by layers and decision points rather than by file order.

### What warrants a second pair of eyes

- The milestone ordering, especially whether utility commands like `zoom` and `history.*` should come before some workflow commands.

### What should be done in the future

1. Add implementation diary entries as code lands under this ticket.
2. Update the gap inventory if new commands land in `surf-go` while this ticket is still active.

### Code review instructions

- Start with the design doc:
  - `design-doc/01-surf-go-non-provider-cli-parity-detailed-architecture-and-implementation-guide.md`
- Then review:
  - `tasks.md`
  - `index.md`
  - `changelog.md`

### Technical details

- The guide references these code anchors heavily:
  - `native/cli.cjs:630-1210`
  - `native/host-helpers.cjs:650-1035`
  - `go/internal/host/router/toolmap.go:133-425`
  - `src/service-worker/index.ts:2501-2615`

## Step 4: Tighten the plan so every new public verb is explicitly Glazed

After the initial ticket was written, the scope was tightened further: the goal is not merely to add the missing commands, but to ensure every new public `surf-go` verb introduced by this effort is authored as a Glazed command. That required revisiting both the design document and the task backlog.

### Prompt Context

**User prompt (verbatim):** "ok, we want to make them all glazed btw. Create a detailed list of tasks, update the doc if needed."

**Assistant interpretation:** Rewrite the ticket so the implementation plan no longer treats generic wrappers as an acceptable end state for medium-priority commands.

**Inferred user intent:** Make the architectural standard explicit now, so implementation work does not drift into a mixed command model later.

### What I did

- Updated the design doc to state that every new public `surf-go` verb in this ticket should be a Glazed command.
- Added a command authoring contract section describing the minimum structure expected of each command.
- Reframed the medium-priority category from “simple wrappers” to “thin shared Glazed wrappers.”
- Replaced the coarse backlog in `tasks.md` with a much more detailed, command-by-command Glazed implementation checklist.
- Updated `changelog.md` to record the scope tightening.

### Why

- Without an explicit constraint, an intern could reasonably keep adding generic wrappers and accidentally lock the Go CLI into an inconsistent long-term architecture.
- The ticket should encode not only what to build, but how to build it.

### What worked

- The updated design doc now explains the difference between shared helper reuse and a generic public command surface.
- The task list is now granular enough to use as an execution checklist rather than only a milestone sketch.

### What didn't work

- The first attempt to patch the design doc in one large step missed some anchor text and had to be redone in smaller, more targeted edits.

### What I learned

- “Make them all Glazed” needs explicit interpretation in writing, otherwise different engineers can satisfy it in incompatible ways.

### What was tricky to build

- The subtle part was preserving room for shared implementation helpers without weakening the requirement that the public command surface itself be Glazed.

### What warrants a second pair of eyes

- The exact command grouping for `scroll`, `zoom`, and `resize`.
- Whether any medium-priority commands should still be postponed despite now having a clearer Glazed path.

### What should be done in the future

1. Keep the ticket tasks updated as command implementations land, not only the changelog.
2. Re-upload the ticket bundle to reMarkable after the next major documentation refresh or after the first implementation slice lands.

### Code review instructions

- Review the revised design sections:
  - `design-doc/01-surf-go-non-provider-cli-parity-detailed-architecture-and-implementation-guide.md`
- Review the new execution checklist:
  - `tasks.md`

### Technical details

```bash
sed -n '320,760p' design-doc/01-surf-go-non-provider-cli-parity-detailed-architecture-and-implementation-guide.md
sed -n '1,260p' tasks.md
```
