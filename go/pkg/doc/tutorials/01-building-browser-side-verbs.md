---
Title: Build Browser-Side Verbs in surf-go
Slug: build-browser-side-verbs
Short: Step-by-step process for turning complex browser automation into a tested surf-go verb.
Topics:
- tutorial
- commands
- browser-automation
- surf-go
Commands:
- js
- navigate
- chatgpt-transcript
- kagi-search
Flags:
- with-glaze-output
- debug-socket
- tab-id
- window-id
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Browser-side verbs in `surf-go` are commands whose real work happens inside a live web page. They are different from simple transport wrappers because the hard part is usually DOM discovery, timing, retries, and deciding which data shape is stable enough to expose to the CLI. The right process is to prove the browser behavior first, then package it as a Go command once the page-side algorithm is stable.

This tutorial explains the process we used for `chatgpt-transcript`, which is now the reference implementation for a complex browser-side verb in this codebase. It also shows the same flow applied to `kagi-search`.

## What You Will Build

By the end of this tutorial, you will know how to build a `surf-go` command that:

- runs browser logic inside the active tab,
- survives real page timing and DOM changes,
- exposes structured Glazed rows for machine use,
- optionally exposes human-readable text output with dual-mode Glazed support,
- and keeps the browser-side logic embedded with `go:embed` so it is easy to find and modify.

## Prerequisites

- A working `surf-go` checkout under `go/`
- The Surf browser extension and native host running
- A command you can use to prototype browser-side JavaScript, currently `surf-go js`
- Familiarity with Glazed command construction in `go/internal/cli/commands`

## Step 1 - Start with a concrete browser problem

A good browser-side verb starts from a page task, not from a command skeleton. You need to know exactly what the user wants from the page before you choose flags, output schemas, or scraping structure.

For `chatgpt-transcript`, the target problem was: extract the current `chatgpt.com` conversation as structured turns and optionally include the Activity sidebar content behind `Thought for ...` buttons.

For `kagi-search`, the target problem is: navigate to a Kagi results page for a query, wait until the results settle, and return structured result rows with titles, URLs, snippets, and result ordering.

Before writing command code, reduce the browser job to one sentence like this:

- “Open or reuse the target page.”
- “Wait until the stable content appears.”
- “Extract the minimal structured payload we actually want.”

That sentence is the contract that the later Go command should preserve.

## Step 2 - Prototype the page logic with `surf-go js`

The fastest way to discover selectors, timing, and failure modes is to run JavaScript directly in the real browser session. This lets you answer page questions before committing to a Go API shape.

The working loop is:

1. write a small page script,
2. run it through `surf-go js`,
3. inspect the returned payload,
4. repeat until the browser algorithm is stable.

Example:

```bash
cd go
go run ./cmd/surf-go js 'return { href: location.href, title: document.title }'
```

When the page task is non-trivial, keep the scripts in a ticket `scripts/` folder while you investigate. This preserves the chronology of what worked and failed. Once the logic is stable enough to become a real command, move the production version into the Go package and embed it.

## Step 3 - Harden the browser algorithm before writing the verb

The first script that “works once” is usually not good enough. Browser-side verbs fail in practice because of page timing, duplicate DOM nodes, stale overlays, and layout changes. You need to harden the algorithm while it is still cheap to change.

For `chatgpt-transcript`, hardening meant:

- walking conversation-turn sections in DOM order,
- deduplicating message candidates by ID,
- choosing the longest non-empty assistant payload,
- opening Activity sidebars with retries,
- and matching the opened flyout to the expected `Thought for ...` duration.

For Kagi, hardening will likely mean:

- waiting for the search results container rather than assuming the page is ready immediately,
- handling ads or sidebars that should not count as primary results,
- and choosing selectors that survive minor class changes.

A practical rule: if you cannot describe the retry and matching logic in prose, the script is not ready to become a command.

## Step 4 - Move production page logic into the Go package

Once the browser algorithm is stable, move it out of ad hoc ticket scripts and into the Go command package. The command-owned script should live next to the command and be embedded with `go:embed`.

Example pattern:

```go
//go:embed scripts/chatgpt_transcript.js
var chatGPTTranscriptScript string
```

This matters for maintenance:

- the command and its page logic stay together,
- code search finds the exact browser behavior behind the verb,
- and future edits do not require hunting through ticket artifacts.

Ticket scripts still matter for research, but the production command should not depend on `ttmp/...` paths.

## Step 5 - Keep the browser script focused on extraction, not presentation

The embedded browser script should return structured data, not final CLI formatting. It should do the minimum page work needed to produce a stable machine representation.

For `chatgpt-transcript`, the browser script returns a payload like:

- conversation URL,
- title,
- turn count,
- transcript items,
- optional activity fields.

The Go layer then decides whether to:

- emit one Glazed row per turn,
- render Markdown to stdout,
- or write a JSON or Markdown artifact.

This separation keeps page logic stable even when output requirements change.

## Step 6 - Build the Glazed command around shared fetch logic

A browser-side verb should have one shared fetch path that all output modes use. In practice that means:

1. decode Glazed settings,
2. build the embedded script prelude and options,
3. execute the browser script through the existing host primitive,
4. parse the returned payload once,
5. fan out into row output or text output.

For `chatgpt-transcript`, this became a shared helper that fetches the transcript data once and then supports both:

- `RunIntoGlazeProcessor(...)` for structured rows,
- `RunIntoWriter(...)` for Markdown output.

That pattern avoids duplicated transport, duplicated parsing, and drifting logic between text and structured modes.

## Step 7 - Use Glazed dual mode when the command needs both human and machine output

Some browser verbs should default to human-readable output but still support structured rows. In `surf-go`, the right pattern is Glazed dual mode.

`chatgpt-transcript` now behaves like this:

- default: Markdown to stdout,
- `--with-glaze-output`: structured rows.

This is implemented by:

- making the command implement both `cmds.WriterCommand` and `cmds.GlazeCommand`,
- and registering it with:

```go
cli.BuildCobraCommand(cmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpSections: []string{schema.DefaultSlug},
        MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
    }),
)
```

In this repository's Glazed version, this explicit dual-mode setup is important. A command that simply implements both interfaces without dual-mode registration will not behave the way you expect.

## Step 8 - Test in three layers

A browser-side verb is not done until it is tested at multiple layers. Each layer catches a different class of failure.

### Unit tests

Use unit tests for:

- script prelude generation,
- response parsing,
- row shaping,
- Markdown rendering,
- and export file rendering.

These tests should run fast and not depend on the browser.

### Mock-host integration tests

Use mock socket tests to verify:

- the CLI sends the expected host tool,
- the embedded script prelude contains the right options,
- and the command wiring stays stable.

### Real-browser validation

Use real-browser validation to confirm:

- selectors still match,
- timing is correct,
- retries behave the way you think they do,
- and the actual site returns the expected shape.

Do not skip this layer for commands that manipulate rich web apps.

## Step 9 - Write down what failed, not just what worked

Complex browser verbs always collect edge cases during development. If you do not write them down, the next person will rediscover them the hard way.

For `chatgpt-transcript`, the investigation diary captured details such as:

- why backend replay was not the first implementation path,
- why Activity flyout scraping required matching by duration,
- and why `EXECUTE_JAVASCRIPT` had to stop mangling template literals.

A good diary should record:

- failed selectors,
- timing quirks,
- protocol surprises,
- and exactly which behaviors were validated only in the real browser.

## Applying This Process to `kagi-search`

`kagi-search` now follows the same sequence in code.

### Concrete implementation shape

1. Navigate to a Kagi search URL for a fixed query.
2. Wait for the primary results container.
3. Return a small structured payload with:
   - title,
   - URL,
   - snippet,
   - result index.

### Current command shape

Structured rows:

```bash
surf-go kagi-search --query "llm transcript attribution" --with-glaze-output --output yaml
```

Human-readable Markdown:

```bash
surf-go kagi-search --query "llm transcript attribution"
```

The implementation uses the same dual-mode pattern as `chatgpt-transcript`: one shared fetch path, Markdown by default, and structured rows behind `--with-glaze-output`.

### What to validate before calling it done

- Result selectors survive a reload.
- Search results are stable enough to scrape after normal navigation.
- Non-result UI blocks do not leak into the main output.
- The page logic behaves predictably when Kagi returns no results, redirects, or partial results.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| The script works in a ticket file but fails once embedded | The worker wrapper or command prelude changed the JavaScript shape | Re-run the embedded script through `surf-go js` and compare the exact wrapped code path |
| A command returns rows but not the expected Markdown | The command only implements `GlazeCommand` | Add `RunIntoWriter(...)` and register with `cli.WithDualMode(true)` |
| The Markdown path works but structured rows do not | The command is running in writer mode only | Use `--with-glaze-output` and verify the command was registered in dual mode |
| Browser automation hangs only in one environment | The wrapper environment is misleading the diagnosis | Re-run the exact browser step from a normal user shell before changing command logic |
| A page scraper becomes brittle after a site update | Selectors depend on incidental layout details | Rework the browser algorithm around durable anchors and explicit matching logic |

## See Also

- [build-first-command](glaze help build-first-command) — Baseline Glazed command construction patterns.
- [writing-help-entries](glaze help writing-help-entries) — Frontmatter and embedding conventions for help pages.
- [how-to-write-good-documentation-pages](glaze help how-to-write-good-documentation-pages) — Documentation style and structure rules.
- [chatgpt_transcript.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt_transcript.go) — Current reference implementation of a complex browser-side verb.
