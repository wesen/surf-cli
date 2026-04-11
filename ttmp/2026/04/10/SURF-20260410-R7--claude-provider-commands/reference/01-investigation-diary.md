# Investigation Diary

## 2026-04-10

- Started Claude provider investigation from the existing grouped surf-go architecture.
- Confirmed there was no existing `claude` or Anthropic browser provider in the Go CLI.
- Opened `https://claude.ai/new` through Surf and verified a live logged-in page session.
- Found that mixing a manual DOM write with Surf's `type` tool polluted the user turn text; future probes should use exactly one input path per tab.
- Found that the Claude composer DOM node exposes a Tiptap-style `editor` object with `editorView`, `editorState`, `commandManager`, and `pmViewDesc`. This is likely the correct integration surface for prompt insertion.
- Confirmed that using the Tiptap editor API (`focus().clearContent().insertContent(...)`) preserves the full prompt string on a fresh `/new` page. The remaining problem is only the submit-button detection after insertion.
- The Tiptap-native insertion plus send-button click appears to submit successfully. A follow-up probe failed only because the editor node was replaced after navigation, so post-submit probes must not hold stale DOM references.
- Found stable transcript selectors for Claude: user turns sit under `div[data-testid="user-message"]`, while assistant output is nested under a `div.font-claude-response` container with body text often in `p.font-claude-response-body`.
- Top-level Claude model menu currently exposes Opus 4.6, Sonnet 4.6, Haiku 4.5, Extended thinking, and More models. Need to inspect whether More models opens a deeper selectable submenu or just a separate picker.

## 2026-04-10 - model menu and thinking mode findings

- Verified that the Claude model selector must be opened with the native `.click()` path. Synthetic `MouseEvent` dispatches were insufficient and did not reliably mount the menu.
- Verified live menu structure on `https://claude.ai/new`:
  - top level current/default model item
  - `Extended thinking` toggle
  - `More models` submenu
- Verified current submenu entries exposed in the live session:
  - `Sonnet 4.6`
  - `Haiku 4.5`
  - `Opus 4.5`
  - `Opus 3`
  - `Sonnet 4.5`
- Verified that submenu content mounts incrementally. Returning after the first non-default item was too early; the command must wait for the submenu model count to settle.
- Updated `claude ask` to:
  - use `.click()` for menu interactions
  - expand `More models`
  - expose `--thinking-mode default|standard|extended`
  - list all currently available models with `--list-models`
- Live `claude ask --list-models` now reports the full discovered model set plus current thinking mode.
- Remaining separate issue: live prompt submission can still time out in the response wait loop even after the message is sent. That is a response-polling problem, not a model-selection problem.

## 2026-04-11 - citation links and searched-the-web export

- Probed the live Claude chat `4991bdb0-f662-4300-8448-02569afaf2ca` on 2026-04-11.
- Verified that assistant responses expose direct citation links as normal `a[href]` elements inside the outermost `div.font-claude-response` node.
- Verified that the `Searched the web` control is a collapsible button inside the assistant subtree.
- Verified that clicking the button expands a nested results panel containing:
  - the search query header
  - result count
  - per-result links and host labels
- Updated `claude transcript` to expand `Searched the web` sections before extraction and attach a structured `searchWeb` object per assistant turn.
- Updated `claude ask` to return `citations` and `searchWeb` metadata for the completed response.
- Verified that Claude completion should not rely on the send button becoming visible again. On completed responses, the reliable marker is the assistant action bar (`Copy`, `Retry`, `Edit`) within the response node.

## 2026-04-11 - long Claude ask completion fixed

- Reproduced the long-response hang with a web-backed Claude answer in a tmux session.
- Confirmed the response text and citation links were present in the tab while the CLI still waited.
- Found the root cause: the completion check looked for action buttons inside the inner `div.font-claude-response` node, but Claude mounts the completed-response action bar (`Copy`, `Retry`, `Edit`) on an ancestor wrapper.
- Updated the completion check to search the assistant wrapper chain instead of only the inner response node.
- Re-ran the long web-backed prompt in tmux and confirmed that `claude ask` now exits successfully with:
  - `response`
  - `citations`
  - `searchWeb`
  - `waitedMs`
- Observed successful completion on chat `33a970e5-e68e-4711-bc03-6d1b220d0c61` with `waitedMs: 14002`.

## Step 5: Claude Artifact Generation, Widget Iframe Discovery, and Current Handoff State

This step covered the first real Claude artifact-generation run, the resulting failure mode in `claude ask`, and the follow-up browser investigation needed to understand where the artifact controls actually live. The important outcome is that artifact creation itself is working in the live browser session, but the current completion logic is still too eager for artifact-producing answers and the download/copy/save controls are not on the outer Claude chat page.

The investigation also established a clean handoff point. We now know the artifact preview is rendered in a separate iframe-backed widget surface, we know the raw `a.claude.ai/isolated-segment.html` page is not the full widget shell, and we have a concrete next target: the `claudemcpcontent.com/mcp_apps?...` iframe URL that appears to host the actual widget runtime.

### Prompt Context

**User prompt (verbatim):** "now ask it to generate a preview html, which will craete an artifact. use tmux again, it will take a while probably"

**Assistant interpretation:** Run a longer Claude prompt that produces an HTML preview artifact, keep the browser tab around, and inspect the resulting artifact surface rather than only the text reply.

**Inferred user intent:** Validate that the new Claude verbs can handle richer Claude outputs, especially generated artifacts, and then figure out how to extract or control the artifact surface in a way that can later become part of `surf-go`.

**Commit (code):** `906b853` — `fix(claude): export citations and search-web details`

### What I did
- Ran the artifact-oriented Claude ask flow in `tmux` so the command could run to completion without being interrupted by the interactive shell.
- Preserved the generated chat tab by using `--keep-tab-open` semantics so the completed browser state could be inspected after the CLI returned.
- Captured the CLI output in `/tmp/surf-claude-artifact.out` for later comparison against the actual final browser state.
- Confirmed the resulting conversation URL and title:
  - `https://claude.ai/chat/291e04b6-370f-40cf-b541-5a13004cea80`
  - title: `Art of insight landing page - Claude`
- Probed the finished tab and identified a real artifact iframe mounted in the chat page.
- Added ordered probe scripts under the ticket `scripts/` directory to inspect the artifact surface and the generation lifecycle:
  - `37-claude-artifact-probe.js`
  - `38-claude-artifact-surface.js`
  - `39-claude-artifact-controls.js`
  - `40-claude-artifact-iframe-context.js`
  - `41-claude-generation-controls.js`
  - `42-claude-artifact-floating-button.js`
  - `43-claude-artifact-floating-open.js`
  - `44-claude-body-children.js`
  - `45-artifact-tab-selector-check.js`
- Compared three distinct surfaces:
  - the outer Claude chat page
  - the raw `a.claude.ai/isolated-segment.html?...` artifact URL
  - the user-provided iframe source URL hosted on `claudemcpcontent.com/mcp_apps?...`
- Opened the user-provided `claudemcpcontent.com` iframe source in a dedicated new tab so it can be inspected directly instead of only through the outer chat document.

### Why
- Artifact generation is a materially different Claude output mode than plain text answers. It needs its own completion criteria and extraction logic.
- The current `claude ask` implementation was designed around assistant text, citations, and searched-the-web panels. That is insufficient once Claude starts emitting iframe-backed widgets.
- The user specifically wants access to artifact actions such as download / copy / save. That requires understanding which DOM context actually hosts those controls.

### What worked
- Claude successfully generated a landing-page artifact in the live browser session.
- The generated artifact conversation is stable and reproducible:
  - chat URL: `https://claude.ai/chat/291e04b6-370f-40cf-b541-5a13004cea80`
- The artifact run clearly demonstrated the current bug with strong evidence from the preserved CLI output:
  - `/tmp/surf-claude-artifact.out` shows
    - `conversationTitle: Art of insight landing page`
    - `currentModel: Sonnet 4.6`
    - `response: Sketching the cover...`
    - `waitedMs: 9849`
  - while the final browser tab later showed the finished landing-page response and a rendered artifact preview.
- The outer chat page does contain a real iframe for the artifact preview.
- The user-provided selector evidence corrected an earlier wrong assumption: the `#more-btn` control is not on the outer page at all.
- The user-provided iframe snippet identified a richer widget host than the earlier raw `a.claude.ai/isolated-segment.html` URL:
  - `https://6ae93f73233bdffea492b6b8571cc222.claudemcpcontent.com/mcp_apps?...&dev=true`
- A dedicated browser tab was opened for that widget URL so it can be inspected directly in the next step.

### What didn't work
- `claude ask` returned too early for the artifact case. The observed returned text was only:
  - `Sketching the cover...`
- The early-return behavior means the current completion detector still treats an intermediate streaming state as final when artifacts are involved.
- Inspecting the raw `a.claude.ai/isolated-segment.html?v=2f582ab1fd` URL did not reveal the expected artifact action controls.
- Searching the outer Claude chat DOM for `#more-btn` failed because that button is not mounted in the outer document.
- Earlier assumptions that the artifact controls might be on the top-level chat page or on the isolated-segment page were therefore incorrect.

### What I learned
- There are at least three distinct Claude artifact surfaces that matter here:
  - the outer Claude chat page, which embeds and frames the artifact
  - the raw isolated segment page, which is not the complete widget shell
  - the widget runtime inside the iframe, apparently hosted on `claudemcpcontent.com/mcp_apps?...`
- The presence of an artifact iframe is not enough to conclude where the user-visible controls live.
- Artifact-producing Claude answers need a more conservative completion rule than plain text answers. Seeing partial assistant text is not sufficient; the command likely needs to observe either:
  - the disappearance of the input stop control, or
  - a stable completed-response action area on the final assistant turn, or
  - a stable artifact/widget state in addition to completed text.
- The user's report about `#more-btn` being inside the iframe is consistent with the evidence and should now be treated as the correct model.

### What was tricky to build
- The tricky part was not locating the final chat, but correctly identifying which DOM context owned the artifact controls. The same logical Claude answer exposed different fragments across multiple surfaces, and earlier probes were looking at the wrong one.
- Artifact answers also blur the definition of “done”. A textual answer may continue to evolve while the artifact is still being generated or mounted. That means the completion condition must be tied to Claude’s real UI lifecycle, not just text extraction.
- Another sharp edge is that opening the iframe `src` directly can still land on a page that is not the same runtime shell the user interacts with in the embedded widget. That is why the `claudemcpcontent.com/mcp_apps?...` URL matters more than the earlier isolated-segment URL.

### What warrants a second pair of eyes
- The artifact-aware completion rule in `claude ask`. This is now the highest-risk correctness issue in the Claude provider.
- The exact browser strategy for interacting with the embedded artifact widget. It may require browser-level interaction primitives rather than simple outer-page `js` execution if the widget remains isolated from the host page.
- The eventual export/download implementation. It should be reviewed carefully before being turned into a user-facing verb because artifacts add cross-origin and lifecycle complexity that plain transcripts do not have.

### What should be done in the future
- Update `claude ask` so artifact-producing responses do not return at the first partial text fragment.
- Probe the `claudemcpcontent.com/mcp_apps?...` widget tab directly for:
  - `#more-btn`
  - its menu items
  - any download / copy / save-as-artifact actions
- If the widget controls are only exposed after hover/focus, add a focused browser-interaction probe sequence rather than relying solely on page-level DOM reads.
- Once the artifact lifecycle is understood, extend `claude transcript` or a dedicated Claude artifact verb to export:
  - artifact metadata
  - widget URL
  - any downloadable payload or action endpoints

### Code review instructions
- Start with the current Claude command and transcript implementation:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/claude.go`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/claude_transcript.go`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/claude_ask.js`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/claude_transcript.js`
- Then read the artifact probe scripts in order starting at `37-claude-artifact-probe.js` through `45-artifact-tab-selector-check.js`.
- Compare `/tmp/surf-claude-artifact.out` against the actual final browser state for chat `291e04b6-370f-40cf-b541-5a13004cea80`.
- Reproduce by opening the artifact chat and confirming that the CLI-captured text is earlier than the final visible answer.

### Technical details
- Artifact run output preserved at:
  - `/tmp/surf-claude-artifact.out`
- Current known artifact conversation:
  - `https://claude.ai/chat/291e04b6-370f-40cf-b541-5a13004cea80`
- Earlier raw iframe URL that was insufficient for control inspection:
  - `https://a.claude.ai/isolated-segment.html?v=2f582ab1fd`
- User-provided richer widget iframe source:
  - `https://6ae93f73233bdffea492b6b8571cc222.claudemcpcontent.com/mcp_apps?connect-src=https%3A%2F%2Fesm.sh+https%3A%2F%2Fcdnjs.cloudflare.com+https%3A%2F%2Fcdn.jsdelivr.net+https%3A%2F%2Funpkg.com&resource-src=https%3A%2F%2Fesm.sh+https%3A%2F%2Fcdnjs.cloudflare.com+https%3A%2F%2Fcdn.jsdelivr.net+https%3A%2F%2Funpkg.com+https%3A%2F%2Fassets.claude.ai&dev=true`
- User-reported widget control selector:
  - `#more-btn > svg`
- User-reported widget button HTML:
  - `<button id="more-btn" aria-label="Widget actions">...</button>`
- Practical next command to resume this thread:
  - inspect the `claudemcpcontent.com` widget tab directly and verify whether `document.querySelector('#more-btn')` resolves there.
- Relevant tab IDs observed during this investigation:
  - `441390959` — Claude chat tab with the finished artifact answer
  - `441390961` — raw `a.claude.ai/isolated-segment.html?...` tab
  - `441390962` — dedicated `claudemcpcontent.com/mcp_apps?...` widget tab opened for direct inspection
- Current worktree state relevant to this step:
  - scripts `37-claude-artifact-probe.js` through `45-artifact-tab-selector-check.js` exist under the R7 ticket `scripts/` folder and were created specifically for this investigation
  - `.codex` and `.envrc` are unrelated untracked paths and should be ignored during code review

## Step 6: Normalize Temporary Claude Artifact Probes into the Ticket Script Sequence

This step cleaned up the research trail rather than changing product code. The goal was to ensure that every ad hoc Claude artifact probe used during the iframe and widget investigation is now preserved under the ticket `scripts/` directory with a stable numeric prefix, so another developer can replay the exact path rather than reconstructing it from shell history or `/tmp` leftovers.

This step also captured one additional artifact-menu finding. Clicking the outer-page `More options for Art of insight landing page` control did not reveal a download-oriented menu; it behaved more like a page- or chat-level control than a direct artifact-action surface. That makes the iframe/widget runtime the more likely location for the true download/copy/save controls the user described.

### Prompt Context

**User prompt (verbatim):** "save all scripts used for research even temporary into the scripts/ folder of the ticket, retroactively too, with numerical prefix so that we can trace your steps"

**Assistant interpretation:** Copy the temporary research scripts created in `/tmp` into the R7 ticket script directory, continue the numbering sequence, and make sure the diary reflects the new ordered inventory.

**Inferred user intent:** Preserve the Claude artifact investigation in a way that is durable, reviewable, and reproducible by someone picking up the work later.

### What I did
- Enumerated the temporary `/tmp` probe files that were created during the Claude artifact iframe investigation.
- Copied them into the R7 ticket `scripts/` directory with the next available numeric prefixes:
  - `46-claude-mcp-proxy-find-more-btn.js`
  - `47-claude-mcp-proxy-inner-iframe.js`
  - `48-claude-isolated-frame-more-btn.js`
  - `49-claude-artifact-outer-iframe-inventory.js`
  - `50-claude-artifact-outer-controls.js`
  - `51-claude-artifact-more-options-click.js`
  - `52-claude-artifact-iframe-src.txt`
- Preserved the earlier already-created artifact probes `37` through `45` as part of the same ordered sequence.
- Polled the still-running `click-claude-more-options.js` probe and captured its result before normalizing the file into the ticket.
- Recorded the current behavioral result of clicking the outer-page `More options for Art of insight landing page` button.

### Why
- Research scripts that only exist in `/tmp` are effectively undocumented. They are easy to lose and hard to correlate with diary notes.
- The user explicitly asked for a complete, numerically ordered script trail so the investigation can be retraced step by step.
- Preserving the latest probe result matters because it narrows the artifact-action surface: the visible outer-page more-options button did not expose the expected artifact download/copy/save menu.

### What worked
- All current temporary Claude artifact probe files were copied into the ticket `scripts/` folder successfully.
- The ordered sequence now extends cleanly from `37` through `52`.
- The polled result from `51-claude-artifact-more-options-click.js` showed:
  - the clicked control had `aria-label` `More options for Art of insight landing page`
  - no meaningful menu container was detected after the click
  - the filtered items that remained visible looked like general page/chat controls rather than a dedicated artifact action menu
- This strengthens the working hypothesis that the actual artifact actions the user saw are inside the artifact/widget runtime, not on the outer Claude chat shell.

### What didn't work
- Clicking the outer-page more-options button did not reveal explicit `download`, `save`, or `copy artifact` menu items.
- The resulting visible items still looked like global or chat-level controls, including `Share`, generic `Copy`, and navigation links.

### What I learned
- The artifact investigation now has a durable script timeline instead of a mix of ticket scripts and ephemeral `/tmp` files.
- The outer page definitely knows about the artifact and exposes some related controls, but the specific artifact actions the user described are still not visible through the current outer-page click path.
- That makes it more likely that the user’s `#more-btn` evidence refers to the richer widget runtime inside the artifact iframe rather than the outer Claude shell.

### What was tricky to build
- The main subtlety here was not the copying itself, but ensuring that the script numbering remains chronological and that the diary preserves the context of why each probe exists. Without that, the script directory becomes a pile of similarly named experiments instead of a replayable investigation record.
- Another subtlety was that one probe was still running when the user asked for normalization. I had to capture its result first so the file copy would represent a meaningful step in the investigation rather than an orphaned script with no recorded outcome.

### What warrants a second pair of eyes
- The naming and ordering of the Claude artifact probe sequence. Another reviewer should confirm that the script names are specific enough for future reuse.
- The interpretation of the outer-page more-options result. It is likely correct, but it should be revalidated once we can inspect the widget runtime directly.

### What should be done in the future
- Keep all future Claude artifact probes in the ticket `scripts/` folder from the start instead of creating new `/tmp`-only files.
- Continue from `52` onward for any new Claude artifact/widget experiments.
- Probe the actual widget runtime next, because the outer-page more-options path did not expose the desired artifact actions.

### Code review instructions
- Inspect the ordered Claude artifact probe sequence in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/10/SURF-20260410-R7--claude-provider-commands/scripts`
- Start at `37-claude-artifact-probe.js` and continue through `52-claude-artifact-iframe-src.txt`.
- Compare the later probes (`49` to `51`) against the diary notes to see how the investigation moved from iframe inventory to outer-page control probing.

### Technical details
- Newly normalized files:
  - `46-claude-mcp-proxy-find-more-btn.js`
  - `47-claude-mcp-proxy-inner-iframe.js`
  - `48-claude-isolated-frame-more-btn.js`
  - `49-claude-artifact-outer-iframe-inventory.js`
  - `50-claude-artifact-outer-controls.js`
  - `51-claude-artifact-more-options-click.js`
  - `52-claude-artifact-iframe-src.txt`
- Key result from `51-claude-artifact-more-options-click.js`:
  - clicked `More options for Art of insight landing page`
  - no dedicated artifact download/save menu was surfaced in the outer page

## Step 7: Attempt Direct Access to the Claude Artifact Widget Runtime

This step tried the two direct technical paths that should have allowed Surf to reach the richer artifact widget surface: CDP frame evaluation and extension frame-context commands. Both paths failed for concrete, different reasons. That is enough evidence to stop guessing and document the actual blocker instead of trying a third speculative workaround.

The result is useful even though it did not unlock the download/copy/save controls. We now know that the widget runtime is present in the outer DOM, but Surf’s current frame-access mechanisms do not surface it in a way that makes its DOM inspectable. That is a product limitation worth addressing directly rather than working around piecemeal.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Continue the Claude artifact investigation after preserving the scripts, focusing on whether the widget runtime can be reached through the existing Surf frame and JS tooling.

**Inferred user intent:** Determine whether the current Surf browser primitives are enough to inspect or automate Claude artifact widget controls, and if not, establish a clear technical blocker.

### What I did
- Added two new ordered probe scripts to the R7 ticket sequence:
  - `53-claude-mcp-shell-from-outer-page.js`
  - `54-claude-aboutblank-widget-frame-probe.js`
- Reconfirmed the outer-page iframe inventory on the artifact chat tab `441390959`.
- Verified that the outer page now visibly contains two relevant iframes at once:
  - `https://6ae93f73233bdffea492b6b8571cc222.claudemcpcontent.com/mcp_apps?...&dev=true`
  - `https://a.claude.ai/isolated-segment.html?v=2f582ab1fd`
- Used page JS on the outer Claude chat page to attempt direct access to the first `mcp_apps` iframe.
- Observed a cross-origin browser error when trying to read `frame.contentDocument` from the outer page.
- Re-ran `frame list` and confirmed that Surf’s CDP frame tree still only exposes:
  - the main Claude chat frame
  - an `about:blank` child frame
  - the `a.claude.ai/isolated-segment.html?...` frame
  - the Intercom frame
- Evaluated directly inside the CDP frame for `a.claude.ai/isolated-segment.html?...` and confirmed that it only contains a trivial `Isolated Segment` page, not the richer widget controls.
- Evaluated directly inside the `about:blank` child frame and confirmed it is empty.
- Tried `frame.switch --index 0` followed by a frame-context read path using `page read`.
- Observed `Content script not loaded. Try refreshing the page.` for the selected child frame.

### Why
- At this stage the question is no longer whether the artifact exists. It does.
- The actual question is whether the current Surf primitives can reach the widget runtime that seems to contain controls like `#more-btn`.
- The clean engineering decision depends on whether this is a command-level problem or a lower-level Surf frame-instrumentation limitation.

### What worked
- The outer page clearly exposes the `mcp_apps` iframe URL and title in DOM inspection.
- The direct outer-page probe produced a decisive browser error instead of ambiguous empty data:
  - `SecurityError: Failed to read a named property 'document' from 'Window': Blocked a frame with origin "https://claude.ai" from accessing a cross-origin frame.`
- The CDP probe of the `a.claude.ai/isolated-segment.html?...` frame was conclusive: it is not the richer widget runtime and does not contain `#more-btn`.
- The `about:blank` child frame probe was also conclusive: it is empty.

### What didn't work
- Direct page-JS inspection of the `mcp_apps` iframe from the outer Claude page failed due to cross-origin restrictions.
- CDP `frame.list` did not expose a frame corresponding to the visible `mcp_apps` URL.
- CDP `frame.js` therefore could not target the richer widget runtime directly.
- The extension frame-context route (`frame.switch` followed by a content-script-backed command) also failed because the selected child frame reported:
  - `Content script not loaded. Try refreshing the page.`

### What I learned
- The widget runtime is real and visible in the outer DOM, but it is not currently reachable through Surf’s existing frame inspection paths.
- The frame-instrumentation mismatch is now the main blocker, not selector quality.
- Surf’s current frame machinery sees only:
  - the trivial isolated-segment CDP frame
  - an empty `about:blank` child frame
  - but not the richer `mcp_apps` widget runtime in an inspectable way.
- This likely means the missing capability is at the Surf platform layer: either frame discovery, frame targeting, or content-script injection into the widget frame.

### What was tricky to build
- The subtle part here is that the outer page and the frame APIs disagree about what frames “exist.” The visible DOM shows the `mcp_apps` iframe, while CDP/frame enumeration collapses the accessible child frames down to an empty `about:blank` frame plus the trivial isolated-segment frame.
- That disagreement makes it easy to chase selectors forever when the real problem is instrumentation coverage. The correct conclusion is not “the selector is wrong,” but “the frame runtime is not exposed through the current toolchain.”

### What warrants a second pair of eyes
- The Surf frame/injection model for cross-origin sandboxed iframes like Claude artifacts.
- Whether `chrome.webNavigation.getAllFrames` and CDP `Page.getFrameTree` are both insufficient for this widget runtime, or whether we are missing a frame-targeting detail.
- Whether content scripts should be injected into the `mcp_apps` widget frame explicitly.

### What should be done in the future
- Treat this as a lower-level Surf capability gap rather than a Claude-command bug.
- Investigate whether the extension can inject content scripts into the `claudemcpcontent.com` widget iframe.
- Investigate whether CDP target/frame enumeration can be extended so the `mcp_apps` runtime is addressable directly.
- Once the widget frame is actually reachable, resume probing for `#more-btn` and the artifact action menu.

### Code review instructions
- Review the new probes first:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/10/SURF-20260410-R7--claude-provider-commands/scripts/53-claude-mcp-shell-from-outer-page.js`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/04/10/SURF-20260410-R7--claude-provider-commands/scripts/54-claude-aboutblank-widget-frame-probe.js`
- Then inspect the frame plumbing in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/service-worker/index.ts`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/src/cdp/controller.ts`
- Focus on the mismatch between:
  - outer DOM iframe inventory
  - `frame.list`
  - `frame.switch`
  - content-script availability in the selected child frame

### Technical details
- Outer-page iframe inventory on chat `291e04b6-370f-40cf-b541-5a13004cea80` now shows both:
  - `claudemcpcontent.com/mcp_apps?...`
  - `a.claude.ai/isolated-segment.html?...`
- Direct outer-page access error for the first iframe:
  - `SecurityError: Failed to read a named property 'document' from 'Window': Blocked a frame with origin "https://claude.ai" from accessing a cross-origin frame.`
- `frame.list` still exposes only:
  - main Claude frame
  - empty `about:blank` child frame
  - `a.claude.ai/isolated-segment.html?...`
  - Intercom
- `frame.js` against the isolated-segment frame returned only a trivial document with `h1` text `Isolated Segment`.
- `frame.js` against the `about:blank` child frame returned an empty document.
- `page read` after `frame.switch --index 0` returned:
  - `Content script not loaded. Try refreshing the page.`
