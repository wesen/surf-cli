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
