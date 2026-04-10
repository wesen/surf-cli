---
Title: Shared tab readiness helper implementation guide
Ticket: SURF-20260410-R6
Status: active
Topics:
  - surf-go
  - glazed
  - cli
  - browser-automation
  - chatgpt
  - kagi
  - gmail
  - debugging
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Step-by-step implementation guide for adding exact-tab readiness to fresh-tab browser verbs and for validating DOM extraction fixes against the installed host.
LastUpdated: 2026-04-10T10:33:48-04:00
WhatFor: Give the next developer a concrete sequence to follow when moving commands from ad hoc tab sleeps to exact tab readiness and when validating provider-side extraction changes.
WhenToUse: Use when implementing new fresh-tab commands or when debugging a mismatch between a browser-page extractor and the installed host behavior.
---

# Shared tab readiness helper implementation guide

## Purpose

This guide describes the concrete implementation sequence used to stabilize `kagi-search` and `kagi-assistant`, and the validation sequence used to compare the interactive ChatGPT provider against `chatgpt-transcript`.

The key lesson is that there are two separate execution environments:

- the local `surf-go` client process that runs your Glazed command
- the installed `surf-host-go` binary that the browser extension actually launches for provider-backed commands such as `chatgpt`

If you patch a provider and only run unit tests, you have not yet validated the installed browser path.

## Step 1: Normalize fresh-tab ownership

Any command that creates a tab and then immediately runs JS should use the shared helper in:

- [tab_ready.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/tab_ready.go)

The sequence is:

1. `tab.new`
2. capture the exact returned `tabId`
3. run a small `js` probe against that exact `tabId`
4. wait until:
   - `document.readyState == "complete"`
   - `location.href` is not `about:blank`
   - optional URL exact or prefix match succeeds
5. run the real extractor or interaction script
6. close the owned tab unless `--keep-tab-open` is set

Do not:

- match tabs by title
- match tabs by “most recent tab”
- rely on arbitrary sleeps before the first JS call

## Step 2: Keep readiness generic and extraction specific

The readiness probe only verifies that the tab is executable and on the intended page. It does not wait for result rows, assistant responses, or Gmail table contents.

Those page-specific readiness checks belong in the embedded browser script for that command.

This separation matters because:

- generic readiness is reusable across Kagi and Gmail
- page-specific readiness needs site-specific selectors and timing

## Step 3: Validate fresh-tab commands against the real browser

Use the real socket path:

```bash
export SURF_SOCKET_PATH=/home/manuel/snap/chromium/common/surf-cli/surf.sock
```

Then validate the command from the `go/` module root with a fresh tab:

```bash
go run ./cmd/surf-go kagi-search --query "hello" --keep-tab-open
```

The important validation target is not just “no error”. The important target is:

- a real fresh tab
- a real browser load
- real extracted rows

For Kagi, the regression case was:

- result containers existed
- but the earlier code returned before extractable result rows were hydrated

The shared helper plus page-specific DOM waiting fixed that.

## Step 4: Distinguish client-side and provider-side fixes

`kagi-search` and `kagi-assistant` are client-side commands. Running `go run ./cmd/surf-go ...` exercises your current checkout immediately.

`chatgpt` is different. The browser extension launches the installed host binary:

- `/home/manuel/snap/chromium/common/surf-cli/surf-host-go`

That means provider changes in:

- [chatgpt.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/providers/chatgpt.go)

do not take effect until you rebuild the installed binary:

```bash
cd go
go build -o /home/manuel/snap/chromium/common/surf-cli/surf-host-go ./cmd/surf-host-go
```

If the browser is still holding the old host process, restart or reload the extension so it launches the updated binary.

## Step 5: Compare provider extraction against a page-ground-truth command

When a provider-backed command returns the wrong content, compare it to a page-ground-truth command that only reads the final page DOM.

For ChatGPT, that ground truth is:

- [chatgpt_transcript.js](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/chatgpt_transcript.js)

Concrete validation sequence:

1. find the completed ChatGPT conversation tab
2. run:

```bash
go run ./cmd/surf-go chatgpt-transcript --tab-id <TAB_ID> --with-glaze-output --output yaml
```

3. record:
   - `messageId`
   - `model`
   - `textLength`
   - leading content preview

4. compare that to what the provider returned

This establishes whether the bug is:

- a page-content problem
- a host/provider extraction problem
- or a completion-state problem

## Step 6: Treat extraction and completion as separate concerns

In the ChatGPT case, two distinct issues appeared:

1. extraction-selection bug
   - provider picked a citation/source fragment
   - transcript extractor found the full assistant body

2. completion-state behavior
   - after patching extraction, the provider tracked the correct assistant turn
   - but the command could still remain in polling while `stopVisible` stayed `true`

Do not collapse those into one vague “ChatGPT is broken” report. Track them separately.

## Step 7: Use host logs to verify the installed path

For snap Chromium, inspect the installed host log with:

```bash
snap run --shell chromium -c 'tail -n 120 /tmp/surf-host-go.log'
```

Useful signals:

- `opened tab <id>`
- `turnCount=<n>`
- `foundAssistant=<bool>`
- `assistantCount=<n>`
- `len=<n>`

When the provider was still on the old logic, the log showed `turnCount=0` while returning citation fragments.

After installing the patched binary, the log showed:

- `turnCount=2`
- `foundAssistant=true`
- non-zero text growth

That was the critical confirmation that the selection logic changed in the running host.

## Step 8: Update tests to reflect the normalized transport sequence

When a command adopts the shared helper, the integration tests should reflect the new order:

1. `tab.new`
2. readiness-probe `js`
3. main page `js`
4. `tab.close`

Do not leave old integration tests asserting a direct `tab.new` -> main `js` flow once the helper is in use.

## Files to inspect

Core code:

- [tab_ready.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/tab_ready.go)
- [kagi_search.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_search.go)
- [kagi_assistant.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_assistant.go)
- [chatgpt.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/providers/chatgpt.go)
- [chatgpt_transcript.js](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/chatgpt_transcript.js)

Tests:

- [integration_test.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/integration_test.go)
- [chatgpt_test.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/providers/chatgpt_test.go)

## Minimum validation checklist

1. `go test ./internal/cli/commands ./internal/host/providers ./cmd/surf-go`
2. `go run ./cmd/surf-go kagi-search --query "hello" --keep-tab-open`
3. `go run ./cmd/surf-go chatgpt-transcript --tab-id <TAB_ID> --with-glaze-output --output yaml`
4. rebuild installed host:

```bash
go build -o /home/manuel/snap/chromium/common/surf-cli/surf-host-go ./cmd/surf-host-go
```

5. inspect the running host log:

```bash
snap run --shell chromium -c 'tail -n 120 /tmp/surf-host-go.log'
```
