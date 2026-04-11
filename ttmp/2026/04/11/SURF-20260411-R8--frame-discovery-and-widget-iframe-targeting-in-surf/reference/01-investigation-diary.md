---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: src/cdp/controller.ts
      Note: Diary references CDP frame enumeration and isolated-world evaluation from this file
    - Path: src/content/accessibility-tree.ts
      Note: Diary references current iframe metadata returned by GET_FRAME_BY_SELECTOR
    - Path: src/service-worker/index.ts
      Note: Diary references frame switching and content-script failure behavior from this file
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---

# Investigation Diary

## Goal

Capture the frame-discovery and widget-iframe targeting work needed to make Surf able to diagnose and eventually target embedded cross-origin widget frames such as Claude artifacts.

## Step 1: Ticket Initialization and Evidence Capture

This step created the dedicated ticket for the frame-discovery problem and anchored the design work to the actual Surf frame plumbing rather than to the higher-level Claude provider code. The important outcome is that the ticket now clearly frames the problem as a platform-level instrumentation issue, not a selector bug or a Claude-only workaround task.

This step also captured the critical implementation anchors from the existing codebase: the service worker path that uses extension frame ids, the CDP path that uses CDP frame ids, and the content script path that currently returns only minimal iframe metadata. Those are the core pieces a new engineer needs to understand before touching the implementation.

### Prompt Context

**User prompt (verbatim):** "ok, let's create a new ticket for the framediscovery, and include a detailed and technical analysis and ipmlementation guide.

reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

Then add tasks, and work on them one by one. Remember to save intermediate / exploratory scripts with numerical prefix in the scripts/ directory of the ticket."

**Assistant interpretation:** Create a new frame-discovery ticket, write a detailed implementation guide and tasks there, upload the bundle to reMarkable, and then begin executing the plan step by step while preserving exploratory artifacts in the ticket script directory.

**Inferred user intent:** Turn the ad hoc Claude artifact iframe debugging into a properly scoped platform ticket with enough structure that another engineer can understand and continue the work.

### What I did
- Created the new ticket workspace:
  - `SURF-20260411-R8`
- Added:
  - a primary design doc
  - a diary doc
- Collected the key implementation anchors from the current codebase:
  - service-worker frame-switch logic
  - CDP frame enumeration/evaluation
  - Go router mappings
  - Go CLI frame command registration
  - content-script frame selector helpers
- Wrote the design doc and task breakdown around those concrete files.

### Why
- The Claude artifact investigation had reached the point where the correct next step was platform work, not more provider-specific probing.
- Without a dedicated ticket, the reasoning and file anchors would remain spread across transient conversation state and the earlier Claude ticket.

### What worked
- `docmgr` is available and the ticket workspace was created successfully.
- `remarquee` is available and authenticated for the later upload step.
- The current codebase clearly exposes the split frame model:
  - extension-frame id path in the service worker
  - CDP frame id path in the controller

### What didn't work
- N/A

### What I learned
- The frame-discovery problem is already sharply defined by the current code structure; the missing piece is tooling and explicit modeling, not lack of evidence.

### What was tricky to build
- The tricky part at this stage was keeping the new ticket focused on the platform issue rather than repeating Claude-specific narrative. The design document therefore treats Claude artifacts as the motivating case, not the solution scope.

### What warrants a second pair of eyes
- The planned resolved-frame identity model should be reviewed before implementation because it will affect multiple command paths.

### What should be done in the future
- Implement the diagnostic path first before changing frame switching behavior.

### Code review instructions
- Start with the new design doc in this ticket.
- Then inspect these files in order:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/service-worker/index.ts`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/cdp/controller.ts`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/content/accessibility-tree.ts`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap.go`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go`

### Technical details
- Ticket path:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/11/SURF-20260411-R8--frame-discovery-and-widget-iframe-targeting-in-surf`
- Core evidence anchors:
  - `src/service-worker/index.ts:1700-1799`
  - `src/cdp/controller.ts:903-950`
  - `src/content/accessibility-tree.ts:1484-1505`
  - `go/internal/host/router/toolmap.go:243-254`
  - `go/cmd/surf-go/main.go:290-299`

## Step 2: First Diagnostic Script - Main Page Iframe Inventory

This step created the first numbered script in the new ticket and used it to capture the visible iframe inventory directly from the main page DOM of the Claude artifact chat. The goal was to freeze one simple source of truth before changing any Surf runtime behavior. That source of truth is the DOM the user can actually see, not the frame list that Surf currently derives through CDP or `webNavigation`.

The result immediately reproduced the core mismatch cleanly. The main page sees four iframe elements, including the visible `claudemcpcontent.com/mcp_apps?...` artifact widget and the hidden `a.claude.ai/isolated-segment.html?...` iframe. That is the concrete evidence the later diagnostic command needs to correlate against extension and CDP frame inventories.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Begin the implementation work in the new ticket by creating and saving the first diagnostic script in the ticket `scripts/` directory, then run it against the known Claude artifact page.

**Inferred user intent:** Ensure the ticket starts with concrete reproducible evidence, not just design prose.

### What I did
- Created the first numbered script in the ticket:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/11/SURF-20260411-R8--frame-discovery-and-widget-iframe-targeting-in-surf/scripts/01-frame-dom-iframe-inventory.js`
- Ran it against the live Claude artifact tab `441390959` using `surf-go js`.
- Stored the resulting output in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/11/SURF-20260411-R8--frame-discovery-and-widget-iframe-targeting-in-surf/sources/01-claude-artifact-dom-iframe-inventory.yaml`
- Marked task 2 complete in the ticket task list.

### Why
- The first implementation task should establish a stable baseline that is independent of Surf's current frame plumbing.
- A direct main-page iframe inventory is the cleanest baseline because it captures what the user actually sees, including cross-origin iframe `src`, `title`, and `sandbox` metadata.

### What worked
- The script returned four iframe elements on the Claude artifact page.
- It captured the visible widget iframe as DOM index `0` with:
  - `src: https://6ae93f73233bdffea492b6b8571cc222.claudemcpcontent.com/mcp_apps?...`
  - `title: Art of insight landing page`
  - `sandbox: allow-scripts allow-same-origin allow-forms`
  - `allow: fullscreen *; clipboard-write *`
- It also captured the hidden isolated frame as DOM index `1` with:
  - `src: https://a.claude.ai/isolated-segment.html?v=2f582ab1fd`
  - `title: Marketing analytics`
- The output now serves as a durable baseline for the later `FRAME_DIAGNOSE` command.

### What didn't work
- N/A

### What I learned
- The main page DOM inventory is richer than either of Surf's current frame-discovery paths.
- This confirms that the future diagnostic command must include DOM iframe inventory explicitly rather than relying only on CDP or `webNavigation`.

### What was tricky to build
- The script itself was straightforward. The main subtlety was keeping it deliberately narrow: inventory visible iframe elements only, do not attempt cross-origin access yet. That keeps the baseline trustworthy and avoids mixing discovery with speculative probing.

### What warrants a second pair of eyes
- The exact choice of fields in the iframe inventory output. It is already useful, but a reviewer may want to add or remove some fields before this shape becomes the stable internal diagnostic format.

### What should be done in the future
- Implement the service-worker `FRAME_DIAGNOSE` message next using this script's output shape as the DOM-iframe input baseline.

### Code review instructions
- Start with the script:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/11/SURF-20260411-R8--frame-discovery-and-widget-iframe-targeting-in-surf/scripts/01-frame-dom-iframe-inventory.js`
- Then inspect the captured output:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/11/SURF-20260411-R8--frame-discovery-and-widget-iframe-targeting-in-surf/sources/01-claude-artifact-dom-iframe-inventory.yaml`
- Compare that output with the current `frame.list` behavior on the same tab.

### Technical details
- Claude artifact tab used for the run:
  - `441390959`
- Output summary:
  - `iframeCount: 4`
  - DOM index `0`: visible `claudemcpcontent.com/mcp_apps?...` widget iframe
  - DOM index `1`: hidden `a.claude.ai/isolated-segment.html?...` iframe
  - DOM index `2`: blank 1x1 iframe
  - DOM index `3`: Intercom iframe

## Step 3: Implement `FRAME_DIAGNOSE`, Reachability Probing, Router Mapping, and CLI Command

This step implemented the first real vertical slice of the frame-diagnostics feature. Instead of changing frame matching logic immediately, the code now exposes a new `FRAME_DIAGNOSE` service-worker path, a new `frame.diagnose` router mapping, and a real `surf-go frame diagnose` dual-mode command. That is the correct first milestone because it gives us one place to see the mismatch before we attempt any behavioral fix.

The implementation is intentionally raw-first. It reports the three inventories we already know are disagreeing: DOM iframe inventory, extension frame inventory, and CDP frame inventory. It also adds content-script reachability probing using a lightweight `PING` message, because the prior `Content script not loaded` failures are part of the same problem and need to be visible in diagnostics instead of only surfacing as generic command errors later.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Start executing the first substantive implementation slice from the new ticket by adding diagnostics rather than changing frame-switch behavior first.

**Inferred user intent:** Build the platform tooling needed to reason about widget frames before attempting a riskier behavioral fix.

### What I did
- Added `PING` handling to the content script in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/content/accessibility-tree.ts`
- Added service-worker helpers to collect:
  - main-page DOM iframe inventory
  - extension-frame diagnostics via `chrome.webNavigation.getAllFrames`
  - content-script reachability per extension frame
- Added a new service-worker message path:
  - `FRAME_DIAGNOSE`
- Added Go router support for:
  - `frame.diagnose`
- Added a new dual-mode Glazed command:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/frame_diagnose.go`
- Registered the new command under:
  - `surf-go frame diagnose`
- Added a focused integration test for the new CLI surface in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/integration_test.go`
- Ran validation:
  - `cd go && go test ./internal/cli/commands ./cmd/surf-go`
  - `npm run build`

### Why
- Diagnostics must come first. Without them, frame-switch changes would still be guesswork.
- `PING` is the cheapest possible probe for child-frame content-script reachability and directly explains the current `Content script not loaded` failures.
- Wiring the CLI early makes the diagnostic path usable by humans immediately instead of leaving it buried as a raw host message.

### What worked
- The new Go command compiles and the new integration test passes.
- The extension build passes after the service-worker and content-script changes.
- The new vertical slice covers all layers needed for the first usable diagnostic flow:
  - content script
  - service worker
  - router
  - CLI command

### What didn't work
- Live browser validation has not been completed yet because the unpacked extension needs to be reloaded before the new service worker code can run in Chromium.
- That means this step is code-complete and build-validated, but not yet browser-validated against the Claude artifact page.

### What I learned
- The first usable frame-diagnostics command does not need a perfect correlation model. Raw inventories plus reachability status are already enough to make the platform mismatch visible.
- Adding `PING` at the content-script layer is a clean way to distinguish discovery failure from injection/reachability failure.

### What was tricky to build
- The main subtlety was keeping the first implementation slice focused. It would have been easy to also redesign `frameContexts` or selector matching in the same patch, but that would have mixed diagnosis with remediation and made failures harder to interpret.
- Another subtlety is that `frame.diagnose` is best treated as a real user-facing diagnostic command, not as an internal-only raw tool. That is why it was implemented as a Glazed dual-mode command instead of just another simple passthrough.

### What warrants a second pair of eyes
- The raw diagnostic output shape. It is intentionally practical rather than elegant, but it should be reviewed before we commit to it as a long-term external command contract.
- The warning heuristics. The current count-mismatch warnings are useful, but probably not the final correlation story.

### What should be done in the future
- Reload the unpacked extension and run `surf-go frame diagnose` against the Claude artifact tab.
- Once the live output is confirmed, move to selector metadata enrichment and better frame matching.

### Code review instructions
- Start with the service-worker additions in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/service-worker/index.ts`
- Then inspect the new content-script `PING` path in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/content/accessibility-tree.ts`
- Then inspect the new command in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/frame_diagnose.go`
- Then inspect registration and router wiring in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap.go`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/main.go`
- Finally review the integration test added in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/integration_test.go`

### Technical details
- New command:
  - `surf-go frame diagnose`
- Expected output categories:
  - `mainPage`
  - `domIframes`
  - `extensionFrames`
  - `cdpFrames`
  - `warnings`
- Validation run results:
  - `go test` passed for `./internal/cli/commands` and `./cmd/surf-go`
  - `npm run build` passed for the extension bundle
