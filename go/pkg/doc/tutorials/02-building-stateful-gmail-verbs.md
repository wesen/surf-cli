---
Title: Build Stateful Gmail Verbs in surf-go
Slug: build-stateful-gmail-verbs
Short: Practical runbook for Gmail list/search style commands that navigate app state before extraction.
Topics:
- tutorial
- gmail
- browser-automation
- surf-go
- glazed
Commands:
- gmail list
- gmail search
- js
- tab.new
- tab.close
Flags:
- inbox
- max-results
- keep-tab-open
- with-glaze-output
- tab-id
- window-id
- debug-socket
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial documents the first Gmail command family in `surf-go`:

- `surf-go gmail list --inbox`
- `surf-go gmail search "<query>"`

These commands are more stateful than `surf-go kagi search` or `surf-go chatgpt transcript` because Gmail is a single-page application with:

- a long-lived shell
- delayed thread-list rendering
- route changes in the hash fragment
- DOM reuse between inbox and search results

## What Gmail Changes

For Gmail-style verbs, it is not enough to prove that JavaScript runs in the tab. You also need to prove that the page state you care about is active.

Examples:

- `gmail list` must be on the inbox route and must wait for real `tr.zA` rows
- `gmail search` must not just fill the search box; it must wait for the Gmail search route and for the thread snapshot to change from the inbox baseline

This is why Gmail verbs build on the shared fresh-tab helper but still keep page-state waits inside the embedded script.

## Recommended Workflow

1. Open a real Gmail session and confirm login state with a tiny `js` probe.
2. Create ordered ticket scripts under `ttmp/.../scripts`:
   - page markers
   - row inventory
   - search box inventory
   - semantic field probe
   - search submission probe
3. Record accepted selectors in the ticket diary before writing production code.
4. Move the final production logic into embedded JS under `go/internal/cli/commands/scripts`.
5. Keep the Go layer responsible for:
   - tab ownership
   - cleanup
   - response parsing
   - row shaping
   - Markdown rendering

## Gmail Selectors That Worked Here

From the live Gmail session used in this repo:

- thread rows: `tr.zA`
- unread marker: row class `zE`
- participant: `.yP, .yW span[email], .yW`
- subject: `.bog, .y6 span[id]`
- snippet: `.y2`
- timestamp: `.xW span, .xW .xS`
- thread ids: descendant nodes with `data-thread-id` and `data-legacy-thread-id`
- search input: `input[name="q"], input[aria-label*="Search mail"]`
- search button: `button[aria-label="Search mail"]`

These selectors are not guaranteed universal forever, but they are concrete enough to build and test against. Do not skip this probe step.

## Gmail Search-Specific Lesson

The first implementation mistake was returning rows as soon as any `tr.zA` rows existed. That was wrong because inbox rows already exist before search results load.

The corrected search script waits for:

- the Gmail search route in `location.href` or `document.title`
- and a changed row snapshot compared to the initial inbox rows

That distinction matters for any stateful browser verb where the page shell and the data view update on different schedules.

## Owned Tabs

Gmail commands should still follow the normal ownership rules:

1. if the command created the tab, close it by default
2. keep it open only with `--keep-tab-open`
3. never close user-targeted tabs supplied with `--tab-id` or `--window-id`

## Validation Checklist

Use all of these:

```bash
cd go
go test ./internal/cli/commands ./cmd/surf-go
go run ./cmd/surf-go gmail list --inbox --max-results 5
go run ./cmd/surf-go gmail search "from:noreply@github.com newer_than:30d" --max-results 5
go run ./cmd/surf-go gmail list --inbox --with-glaze-output --output yaml
go run ./cmd/surf-go gmail search "from:noreply@github.com newer_than:30d" --with-glaze-output --output yaml
```

If a Gmail command returns inbox rows for a search query, the bug is almost always in the page-state wait, not in the Go transport.
