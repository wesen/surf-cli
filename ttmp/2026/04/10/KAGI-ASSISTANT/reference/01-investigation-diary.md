# Kagi Assistant Investigation Diary

## 2026-04-10

Started live-session probing against the real Chromium profile via surf-go and the native host socket. Goal is to map Kagi Assistant UI controls before designing the command surface.

- Confirmed live Kagi Assistant page is loaded in a dedicated tab. Initial probe found a textarea with placeholder `Ask Assistant`, a `Web access` checkbox, and a large number of unlabeled sidebar/action buttons. Need deeper control mapping for model and assistant selectors.

- DOM probing showed the composer area itself is hard to isolate from static body content, but the page includes a large inline script with Kagi-specific globals and translations. Next step is to inspect window globals directly for model and assistant inventories.

- Window globals exposed `window.modelListUltimate` and model-selection translations. A live button on the page also exposes the current model/custom-assistant state, which is likely the real selector entry point.

- Identified `#profile-select` as the actual model/assistant chooser. It sits in `.prompt-options` beside the web-search toggle and lens selector. Next step is to open it and inventory the rendered listbox options.

- Opened the chooser successfully. The popup contains three built-in assistants (Quick, Research, Research (Experimental)), a Custom Assistant section (Code, Study), and provider-grouped raw models with stable `data-model` identifiers in SVG icons.

- Prompt submission on a live assistant tab navigates to a conversation URL under `/assistant/<uuid>`. The result DOM is not ChatGPT-style; assistant replies render as `.chat_bubble` blocks and expose reasoning in a native `<details><summary>Thinking</summary>...</details>` subtree. The page body also includes per-answer metadata in `.message-info` (`Model`, `Version`, `Tokens`, `Cost / Total`, `End to end time`, `Submitted`).

- The post-submit conversation view is read-only and prompts the user to click to start a new conversation. That means the command should treat a writable `/assistant` page as the execution context and should not attempt to reuse an existing read-only thread for fresh prompt submission.

- Normalized all Kagi Assistant probe artifacts into ordered scripts under `ttmp/2026/04/10/KAGI-ASSISTANT/scripts/` so the investigation can be replayed step by step. The ordered sequence is:
  - `01-kagi-assistant-dom-probe.js`
  - `02-kagi-assistant-controls-probe.js`
  - `03-kagi-assistant-composer-probe.js`
  - `04-kagi-assistant-globals-probe.js`
  - `05-kagi-assistant-global-text-scan.js`
  - `06-kagi-assistant-detail-probe.js`
  - `07-kagi-assistant-target-button-probe.js`
  - `08-kagi-assistant-prompt-probe.js`
  - `09-kagi-assistant-open-selectors-probe.js`
  - `10-kagi-assistant-open-profile-select.js`
  - `11-kagi-assistant-profile-dialog-probe.js`
  - `12-kagi-assistant-select-profile-probe.js`
  - `13-kagi-assistant-lens-probe.js`
  - `14-kagi-assistant-lens-real-probe.js`
  - `15-kagi-assistant-web-toggle-probe.js`
  - `16-kagi-assistant-submit-probe.js`
  - `17-kagi-assistant-select-model-gpt5mini.js`
  - `18-kagi-assistant-select-assistant-quick.js`

- The later `realClick`-style probes (`14`, `15`, `17`, `18`) are the reliable interaction variants. Earlier click-only probes are still preserved because they show the failed assumptions and intermediate selector discoveries.

- Investigated conversation tagging controls and preserved the probes as `19` through `26` in the ticket script sequence. Key result: tagging is not a separate modal. Clicking `#tags-add` expands an inline `dialog.promptOptionsSelector` inside `#tags` with a search field (`Search Tags`), existing tag checkboxes (`Temporary`, `Public`, `3d-printing`, `engineering`, `photo`), and a `Create <name>` button that becomes enabled when a missing tag name is typed.

- Verified that existing tag application works directly on a finished assistant thread: toggling the `Temporary` checkbox immediately changed the visible thread tags from `Untagged` to `Untagged, Temporary`. Verified that typing a non-existent name enables the `Create <name>` button without mutating the account until it is clicked.

- Extended `kagi-assistant` with tag flags:
  - `--tags` accepts a comma-separated list of tag names to apply after the assistant response is created.
  - `--create-tags` allows missing tags to be created through the inline tag editor.
  - `--list-tags` lists the available conversation tags as structured Glazed rows.
  - `--list-all-options` now includes tags in addition to assistants, models, and lenses.

- Live validation after the implementation:
  - `surf-go kagi-assistant --list-tags --with-glaze-output --output yaml` returned the current tag inventory correctly.
  - `surf-go kagi-assistant "Reply with exactly: KAGI_TAGGED_2" --assistant Quick --tags "Temporary,photo" --with-glaze-output --output yaml` completed successfully and returned a `tagSelection` payload showing both tags as applied, with `visibleTags` equal to `Untagged`, `Temporary`, and `photo`.
