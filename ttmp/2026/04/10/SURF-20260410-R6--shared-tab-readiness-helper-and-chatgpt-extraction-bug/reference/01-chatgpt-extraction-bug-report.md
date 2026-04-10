---
Title: ChatGPT extraction bug report
Ticket: SURF-20260410-R6
Status: active
Topics:
  - surf-go
  - glazed
  - cli
  - browser-automation
  - chatgpt
  - debugging
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: `surf chatgpt` can return a citation/source fragment instead of the full assistant body even when `chatgpt-transcript` sees the correct answer on the same conversation.
LastUpdated: 2026-04-10T10:33:48-04:00
WhatFor: Record the concrete ChatGPT extraction failure, evidence, likely root cause, and fix direction.
WhenToUse: Use when debugging ChatGPT response extraction or reviewing the provider polling logic.
---

# ChatGPT extraction bug report

## Goal

Document the mismatch between the interactive `chatgpt` command and `chatgpt-transcript` on the same live ChatGPT conversation.

## Context

The user ran a long ChatGPT prompt through the CLI and received an obviously garbled result consisting mostly of repeated source labels such as `MIT OpenCourseWare`, `MIT Press`, and `Mathematical Association of America`, rather than the actual body of the answer.

The resulting conversation remained open at:

- `https://chatgpt.com/c/69d905f3-181c-8333-bd54-b1ff21a3572d`

A direct transcript export of that exact conversation shows that the full assistant response is present in the DOM and can be extracted correctly.

## Quick Reference

### User-visible failure

Observed `surf chatgpt` output shape:

- repeated source labels
- missing the actual answer body
- total runtime about `214734 ms`

### Contradictory evidence from `chatgpt-transcript`

Using `chatgpt-transcript` on the same tab returns:

- the original user prompt as turn 1
- a full assistant body of about `18933` characters as turn 2
- assistant model `gpt-5-4-thinking`
- coherent prose beginning with:
  - `Sanjoy Mahajan’s The Art of Insight in Science and Engineering is best understood...`

### Concrete comparison

`chatgpt-transcript` succeeded on tab `441390650` for conversation:

- `https://chatgpt.com/c/69d905f3-181c-8333-bd54-b1ff21a3572d`

That confirms the answer exists in the page DOM and that the first bug is in the interactive provider extraction path, not in the underlying ChatGPT page content.

## Likely Root Cause

The interactive provider in [chatgpt.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/providers/chatgpt.go) used a polling heuristic built around the last assistant node found globally in the document.

That logic is weaker than the transcript extractor because it can lock onto the wrong assistant-associated subtree when ChatGPT renders:

- citations
- source chips
- auxiliary assistant subnodes
- nested assistant-marked blocks inside the same turn

By contrast, [chatgpt_transcript.js](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/scripts/chatgpt_transcript.js) extracts by:

- iterating conversation-turn sections in order
- collecting assistant candidates within a turn
- deduplicating by message id
- selecting the longest non-empty candidate for that logical message

That turn-based selection is robust against smaller citation-only fragments.

## Implemented Fix

The provider polling logic now mirrors the transcript extractor more closely.

Implemented behavior:

- identify the last assistant conversation turn, not just the last assistant-tagged node globally
- gather assistant candidates within that turn
- deduplicate by `data-message-id` where available
- prefer the longest non-empty candidate text for the turn
- keep the existing `stopVisible` and `finished` completion checks

After rebuilding the installed snap-side `surf-host-go` binary, the live host logs changed from:

- `turnCount=0` with citation-fragment extraction

to:

- `turnCount=2`
- `foundAssistant=true`
- stable non-zero assistant text growth

That is strong evidence that the extraction-selection bug is fixed.

## Remaining Issue

The long-running interactive validation exposed a second behavior:

- the provider now tracks the correct assistant turn
- but on research-heavy responses the command may remain in polling because `stopVisible` stays `true` for a long time even after the extracted text stabilizes

This is a different issue from the original garbled-output bug. The extraction-selection bug and the completion-gate behavior should be tracked separately.

## Validation Status

Validated:

- `chatgpt-transcript` returns the correct full body on the problematic conversation
- the installed host now logs the correct `turnCount` and assistant selection behavior during live polling
- provider tests were updated and pass

Still pending:

- full end-to-end proof that the interactive `chatgpt` command returns successfully on this exact long-form prompt with the updated host

## Validation Plan

1. Keep the turn-based extraction patch in place.
2. Re-run the long-form ChatGPT prompt flow against the installed snap-side host.
3. Decide whether the completion gate should continue to wait for `stopVisible == false` in long-running browsing/research responses.
4. Confirm `surf chatgpt` returns the same substantive body that `chatgpt-transcript` sees on the final conversation page.
