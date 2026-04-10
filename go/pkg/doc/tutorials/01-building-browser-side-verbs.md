---
Title: Build Browser-Side Verbs in surf-go
Slug: build-browser-side-verbs
Short: Practical runbook for turning complex browser automation into a tested surf-go verb.
Topics:
- tutorial
- commands
- browser-automation
- surf-go
- glazed
Commands:
- js
- navigate
- tab.new
- tab.close
- chatgpt transcript
- kagi search
- kagi assistant
Flags:
- with-glaze-output
- debug-socket
- tab-id
- window-id
- keep-tab-open
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial is the practical process for building a non-trivial browser-side verb in `surf-go`. It is based on three real implementations in this repository:

- `surf-go chatgpt transcript` in `go/internal/cli/commands/chatgpt_transcript.go`
- `surf-go kagi search` in `go/internal/cli/commands/kagi_search.go`
- `surf-go kagi assistant` in `go/internal/cli/commands/kagi_assistant.go`

The goal is not just to explain the shape of the final code. The goal is to explain the workflow that gets you there without guessing: how to prototype the page logic, how to decide whether to reuse an existing tab or create one, how to structure the embedded JavaScript, how to expose both Markdown and structured rows, how to clean up browser state after the command finishes, and how to test the result at the right layers.

If someone follows this document carefully, they should be able to add the next complex browser verb without having to rediscover the same pitfalls.

For stateful mailbox-style workflows, also read:

- `surf-go help build-stateful-gmail-verbs`

## What This Pattern Is For

Use this pattern when the command's real logic lives inside a browser page rather than inside the Go process.

Typical examples:

- scrape a conversation or search results page
- interact with page controls before extracting data
- wait for a rich web app to settle
- open or close page subviews and attach their content to the main output

Do not use this pattern for purely transport-level commands such as:

- listing tabs
- switching windows
- basic navigation wrappers with no page scraping

Those commands can usually remain thin wrappers over an existing host tool.

## The Actual Architecture

Before writing code, understand the execution path. A browser-side verb in `surf-go` typically flows through these layers:

1. Cobra + Glazed command
2. Go command logic in `go/internal/cli/commands`
3. socket transport in `go/internal/cli/transport/client.go`
4. host tool request envelope from `go/internal/cli/commands/base.go`
5. native host router in `go/internal/host/router/toolmap.go`
6. extension service worker in `src/service-worker/index.ts`
7. page DOM and page JavaScript

For the commands in this repository, the most common tool path is:

```text
surf-go command
  -> ExecuteTool(...)
  -> tool_request over unix socket
  -> service worker receives mapped host message
  -> page-side JS runs
  -> response comes back as structured text or plain text
  -> Go parses once
  -> Go renders rows or Markdown
```

That last split matters. The browser script should be responsible for page interaction and extraction. The Go code should be responsible for presentation and CLI ergonomics.

## The Reference Implementations

Use these as templates.

### `surf-go chatgpt transcript`

Relevant files:

- `go/internal/cli/commands/chatgpt_transcript.go`
- `go/internal/cli/commands/scripts/chatgpt_transcript.js`
- `go/internal/cli/commands/chatgpt_transcript_test.go`
- `go/cmd/surf-go/integration_test.go`

What it demonstrates:

- extracting multiple structured records from a complex page
- handling follow-up UI state such as Activity sidebars
- dual-mode Glazed output
- optional export artifacts
- embedding production page logic with `go:embed`

### `surf-go kagi search`

Relevant files:

- `go/internal/cli/commands/kagi_search.go`
- `go/internal/cli/commands/scripts/kagi_search.js`
- `go/internal/cli/commands/kagi_search_test.go`
- `go/cmd/surf-go/integration_test.go`

What it demonstrates:

- navigate-or-create-tab strategy
- search-result scraping with deduplication
- suppressing noisy page blocks when the page does not expose useful data
- the same dual-mode output structure with a simpler page model

### `surf-go kagi assistant`

Relevant files:

- `go/internal/cli/commands/kagi_assistant.go`
- `go/internal/cli/commands/scripts/kagi_assistant.js`
- `go/internal/cli/commands/kagi_assistant_test.go`
- `go/cmd/surf-go/integration_test.go`

What it demonstrates:

- stateful UI interaction rather than simple page scraping
- selecting assistants, models, lenses, and toggles in a live app
- applying optional tagging state before submission
- tracking ownership of a command-created tab and closing it by default

## Step 0 - Write Down the Exact User Contract

Start by writing one sentence that states what the command does. This is not optional. It prevents the implementation from drifting into “browser stuff” without a stable target.

Good examples:

- “Export the current ChatGPT conversation as structured turns and optionally attach Activity sidebar text.”
- “Run a Kagi search for a query and return the main result rows as Markdown or structured rows.”

Then write a short contract with four parts:

1. entry condition
2. page action
3. extraction target
4. output modes

For `kagi-search`, the contract is:

- entry condition: the Surf native host is running
- page action: open a tab at the Kagi search URL, or reuse an explicitly targeted tab
- extraction target: Kagi result rows and optionally Quick Answer if substantive
- output modes: Markdown by default, structured rows behind `--with-glaze-output`
- cleanup: close a tab the command created itself unless `--keep-tab-open` is set

For `kagi-assistant`, the contract is:

- entry condition: the Surf native host is running and the browser session is already logged into Kagi
- page action: open or reuse the assistant page, select assistant/model/lens/options, submit a prompt, and optionally apply tags
- extraction target: response text, thinking/details text, metadata, and tag-selection state
- output modes: Markdown by default, structured rows behind `--with-glaze-output`
- cleanup: close a tab the command created itself unless `--keep-tab-open` is set

If you cannot write this down cleanly, the command is not ready to implement.

## Step 1 - Prototype in the Real Browser with `surf-go js`

Do not start by writing Go command code. Start by proving the browser logic in a real page.

This is the recommended loop:

1. open the target page in the browser
2. run tiny JavaScript probes using `surf-go js`
3. inspect the returned data
4. refine selectors and timing
5. save the useful probes in the ticket `scripts/` folder

Example:

```bash
cd go
go run ./cmd/surf-go js 'return { href: location.href, title: document.title, ready: document.readyState }'
```

Then move to a probe that identifies the exact structures you care about. For example, on Kagi we validated:

- result containers
- title links
- snippet blocks
- whether Quick Answer actually contained useful answer text

Store those probes in the ticket workspace while researching. Use ordered file names so the investigation can be replayed later instead of reverse-engineered from timestamps.

Recommended pattern:

- `01-page-shape-probe.js`
- `02-dialog-open-probe.js`
- `03-option-selection-probe.js`
- `04-submit-and-extract-probe.js`

For Kagi Assistant, that ordered trail lives under:

- `ttmp/2026/04/10/KAGI-ASSISTANT/scripts/`

Research scripts belong in the ticket. Production scripts belong in the Go package.

## Step 2 - Stabilize the Page Algorithm Before You Write the Command

The first script that returns plausible data is not enough. At this stage, your job is to remove ambiguity.

Things to check deliberately:

- does the page render duplicate logical items?
- do you need to dedupe by URL, message ID, or some other stable key?
- do some containers look like results but actually contain menus or sidebars?
- does the page expose useful data immediately, or do you need to wait for content to settle?
- do some visible UI blocks contain only labels and no useful payload?

This is where most browser verbs succeed or fail.

For `chatgpt-transcript`, stabilization required:

- iterating conversation sections in order
- gathering message candidates from each section
- choosing the longest non-empty candidate per message
- retrying Activity flyout opening and matching by duration text

For `kagi-search`, stabilization required:

- waiting for either `main div._0_SRI.search-result` or `main div.__srgi`
- deduping by final URL
- ignoring Quick Answer when it contained only `Quick Answer`, `References`, and `Continue in Assistant`

A good rule is this:

- if the extraction logic depends on “the page looked right once”, it is not stable enough
- if the logic names explicit selectors, retries, and dedupe rules, it is probably ready

## Step 3 - Decide the Tab Strategy Explicitly

This is the first implementation decision that trips people up.

Your command must decide how it acquires a page context. There are three common modes:

1. current tab only
2. reuse explicit `--tab-id` or `--window-id`
3. create a new tab if none is supplied

Do not leave this implicit.

Why this matters:

- `navigate` still depends on a resolved active tab unless you give it a specific target context
- some commands should be side-effect-free on the user's current page
- some commands should be self-contained and open their own target page

`chatgpt-transcript` uses the current page context. That is correct because it exports the current open conversation.

`kagi-search` should not require the user to manually create a tab. That is why its fetch path now does:

```text
if tab-id and window-id are both absent:
  create a new tab at the Kagi URL
  capture returned tabId
  run JS against that tab
else:
  navigate the targeted tab/window
  run JS against that target
```

That logic lives in `fetchKagiSearch(...)` in `go/internal/cli/commands/kagi_search.go`.

When adding a new verb, decide this up front and encode it in the fetch helper. Do not let it emerge accidentally from whatever `navigate` happens to do.

`kagi-assistant` uses the same acquisition pattern, but it also mutates more browser state:

```text
if tab-id and window-id are both absent:
  create a new tab at the assistant URL
  remember that the command owns that tab
  run assistant interaction JS against that tab
  close the owned tab unless --keep-tab-open was set
else:
  navigate the targeted tab/window
  run assistant interaction JS against that target
  never close a user-supplied target
```

## Step 3.5 - Define Browser-State Ownership and Cleanup

This step should be explicit in the design, not left as an implementation detail.

If the command changes browser state, decide which of these it owns:

- created tabs
- opened dialogs or flyouts
- form values
- toggles or dropdown selections
- created remote objects such as tags or saved items

Use these rules:

1. If the command creates a tab because the user did not supply one, the command owns that tab.
2. If the user supplied `--tab-id` or `--window-id`, the command does not own that tab and must not close it.
3. If cleanup would destroy or mutate user data, it must be behind an explicit opt-in flag.
4. If cleanup only removes command-owned temporary browser state, make it the default.

The current Kagi commands follow this contract:

- default: close an owned tab when finished
- opt-out: `--keep-tab-open`
- never close a user-targeted tab

## Step 4 - Move Production Page Logic into an Embedded Script

Once the browser algorithm is stable, move it into `go/internal/cli/commands/scripts/` and embed it.

Example:

```go
//go:embed scripts/kagi_search.js
var kagiSearchScript string
```

Why this is the right pattern:

- the production page logic is easy to locate with code search
- command and browser behavior stay together
- the script ships with the binary
- future contributors can modify the page logic without hunting through ticket artifacts

Do not read production scripts from `ttmp/...` at runtime.

Keep the embedded script focused on page work:

- wait
- query
- click if needed
- extract
- return structured data

Do not format Markdown or CLI tables inside the browser script.

## Step 5 - Pass Options Through a Small Prelude

The embedded script should not hardcode every runtime choice. Instead, pass a small options object from Go into the script.

Pattern:

```go
options := map[string]any{
    "maxResults": s.MaxResults,
}
b, _ := json.Marshal(options)
code := fmt.Sprintf("const SURF_OPTIONS = %s;\n%s", string(b), embeddedScript)
```

The JavaScript then starts with:

```js
const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
```

Use this pattern for things like:

- maximum result count
- whether optional subviews should be opened
- retry counts
- feature toggles for extraction modes

Do not interpolate dozens of ad hoc strings into the script body.

## Step 6 - Keep the Browser Script Strictly About Extraction

Your browser script should return a plain structured object. Think of it as a tiny page-specific extractor.

Good returned fields:

- page URL
- page title
- counts
- extracted items
- optional metadata that helps explain extraction behavior

Examples from the current commands:

- `chatgpt-transcript` returns `href`, `title`, `turnCount`, `withActivity`, `activityExported`, `transcript`
- `kagi-search` returns `query`, `href`, `title`, `waitedMs`, `resultCount`, `results`, `quickAnswer`

Bad responsibilities for the browser script:

- writing Markdown headings
- deciding CLI output mode
- flattening everything into a single text blob just because that is easier

The Go layer is where representation choices belong.

## Step 7 - Write One Shared Fetch Function in Go

Every serious browser-side verb should have one shared fetch helper that:

1. validates settings
2. constructs any URL or target
3. prepares the script prelude
4. creates the transport client
5. acquires or resolves the tab
6. executes the relevant host tools
7. parses the result once

This is the center of the command.

Examples:

- `fetchChatGPTTranscript(...)`
- `fetchKagiSearch(...)`

Why this matters:

- both structured and writer output need the same underlying data
- you want one place to change transport behavior
- tests become simpler

Pseudocode:

```text
fetchThing(ctx, settings):
  validate settings
  build embedded JS with SURF_OPTIONS
  create transport client
  resolve tab strategy
  maybe navigate or create tab
  remember whether the command owns the tab
  execute js tool
  parse structured response
  close owned tab unless the command was told to keep it open
  return typed data wrapper
```

If your command has separate fetch logic for Markdown and row output, stop and refactor.

## Step 8 - Parse Once, Then Fan Out

After the host returns, parse the response once and then feed the result into either rows or Markdown.

Current repository helpers:

- `parseResult(...)` in `go/internal/cli/commands/format.go`
- `extractErrorText(...)` in `go/internal/cli/commands/format.go`

Typical command-local functions:

- `parseKagiSearchResponse(...)`
- `kagiSearchDataToRows(...)`
- `renderKagiSearchMarkdown(...)`

The correct layering is:

```text
raw host response
  -> parse structured payload
  -> internal typed wrapper or map
  -> rows or Markdown
```

Do not re-parse the same text separately in `RunIntoGlazeProcessor(...)` and `RunIntoWriter(...)`.

## Step 9 - Make It a Real Glazed Command

In this repository, new public verbs should be real Glazed commands.

That means:

1. command description
2. typed settings struct with `glazed` tags
3. flags in `cmds.WithFlags(...)`
4. command and Glazed sections
5. `RunIntoGlazeProcessor(...)`
6. optionally `RunIntoWriter(...)`

Typical skeleton:

```go
type MyCommand struct {
    *cmds.CommandDescription
}

type MySettings struct {
    Query string `glazed:"query"`
}

var _ cmds.GlazeCommand = (*MyCommand)(nil)
var _ cmds.WriterCommand = (*MyCommand)(nil)
```

If the command should support both human-readable output and structured rows, use dual mode.

## Step 10 - Use Dual Mode for Human + Machine Output

For complex verbs, Markdown is often the best default for humans, while rows are better for automation.

The current repository pattern is:

- default writer mode
- `--with-glaze-output` for rows

Registration pattern in `go/cmd/surf-go/main.go`:

```go
cobraCmd, err := cli.BuildCobraCommand(cmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
    cli.WithParserConfig(cli.CobraParserConfig{
        ShortHelpSections: []string{schema.DefaultSlug},
        MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
    }),
)
```

Why this is necessary:

- this repository's Glazed version does not automatically do the right thing if a command implements both writer and glaze interfaces without explicit dual-mode registration

Use this mode when:

- the result is naturally read as a report
- the same data also benefits from machine-readable rows

Current examples:

- `chatgpt-transcript`
- `kagi-search`

## Step 11 - Wire the Command into the Root and Help System

Implementation is not complete until the command is registered and documented.

Relevant root file:

- `go/cmd/surf-go/main.go`

Relevant embedded help files:

- `go/pkg/doc/doc.go`
- `go/pkg/doc/tutorials/01-building-browser-side-verbs.md`

Root-level requirements:

- build or register the command
- if dual mode, use `cli.BuildCobraCommand(...)` with `cli.WithDualMode(true)`
- load embedded docs into the help system

The help-system wiring now looks like:

```go
if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
    return nil, err
}
help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
```

If a new command has no help surface and no discoverability, the task is only partially done.

## Step 12 - Test at Three Layers

This is the minimum testing standard for a complex browser verb.

### Layer 1: unit tests

Use unit tests for:

- script prelude generation
- URL construction
- response parsing
- row shaping
- Markdown rendering
- small helpers such as extracting `tabId` from a tab-creation response

Examples:

- `go/internal/cli/commands/chatgpt_transcript_test.go`
- `go/internal/cli/commands/kagi_search_test.go`

### Layer 2: mock-host integration tests

Use mock unix-socket tests in `go/cmd/surf-go/integration_test.go` to verify:

- the command sends the expected tool sequence
- the arguments are correct
- the embedded script prelude contains the expected options
- cleanup requests such as `tab.close` happen only when they should

This is especially important for commands like `kagi-search` whose request sequence matters. After the active-tab bug, the integration test now confirms:

1. `tab.new` is sent first when no explicit target exists
2. `js` is sent second with the expected `SURF_OPTIONS`
3. `tab.close` is sent third when the command created the tab itself

For `kagi-assistant`, the integration test also confirms:

1. `tab.new` is sent first when no explicit target exists
2. `js` contains the expected assistant/model/tag options
3. `tab.close` is sent only for a command-owned tab

### Layer 3: real-browser validation

You still need one real-browser pass. Mock tests do not validate selectors.

Use either:

- `surf-go js` against the real browser session
- or Playwright for DOM inspection when the extension socket is unavailable

For real-browser validation, confirm at least:

- the page selectors still match
- the scraper returns clean data
- retries behave correctly
- the command does not depend on accidental browser state

## Step 13 - Keep a Research Trail

Every complex browser verb should leave behind a useful trail in the ticket.

For the current work, the ticket artifacts include:

- ChatGPT transcript research diary
- Kagi search research diary
- ticket-side probe scripts under `ttmp/.../scripts`

Why this matters:

- future contributors can see what failed
- site-specific quirks are documented
- temporary probes do not contaminate production code

A good diary records:

- working selectors
- rejected selectors
- timing behavior
- dedupe rules
- ownership and cleanup rules
- what was validated live versus only in tests

## A Concrete Walkthrough: `kagi-search`

This section compresses the actual implementation into a sequence you can reuse.

### 1. define the contract

We wanted:

- query in
- Kagi results out
- Markdown by default
- structured rows optionally
- no requirement that the user manually open a tab first

### 2. validate the DOM

We used page probes to confirm:

- `main div._0_SRI.search-result`
- `main div.__srgi`
- `h3 a[href^="http"]`
- `._0_DESC.__sri-desc`
- `.__sri-desc`

### 3. identify the noisy page blocks

We found that Quick Answer on the test page exposed only:

- `Quick Answer`
- `References`
- `Continue in Assistant`

That meant:

- include Quick Answer only when it contains substantive text
- otherwise return `null`

### 4. choose tab behavior

We first used `navigate` directly and discovered the active-tab dependency. The fix was:

- `tab.new` when no explicit tab or window target is supplied
- `navigate` only when a target is explicitly given
- `tab.close` when the command created the tab itself and `--keep-tab-open` is false

That is the kind of issue you should expect and design around.

### 5. embed the production extractor

The final script lives in:

- `go/internal/cli/commands/scripts/kagi_search.js`

and is embedded with:

```go
//go:embed scripts/kagi_search.js
var kagiSearchScript string
```

### 6. implement shared fetch and dual-mode output

Go fetch path:

- build search URL
- create client
- create or navigate tab
- run JS
- parse once

Go output path:

- rows via `kagiSearchDataToRows(...)`
- Markdown via `renderKagiSearchMarkdown(...)`

### 7. lock behavior with tests

Tests verify:

- URL building
- script prelude content
- row shaping
- Markdown rendering
- tab-creation fallback sequence

That is the full lifecycle for a serious browser-side verb.

## Practical Checklist

Before starting:

- decide the user contract
- decide whether the command owns the tab or reuses it
- decide what browser state the command must clean up
- decide whether the result should default to Markdown or rows

During browser research:

- validate selectors in a real page
- look for duplicates
- look for page blocks that look useful but are just chrome
- save probes in the ticket

Before writing Go:

- ensure the page algorithm can be explained step by step
- move production JS into `go/internal/cli/commands/scripts`
- embed it with `go:embed`

Before calling it done:

- add unit tests
- add mock-host integration tests
- run a real-browser validation pass
- wire the command into the root
- verify cleanup behavior for command-owned tabs and user-targeted tabs
- make sure `surf-go help <command>` and `surf-go help build-browser-side-verbs` work

## Common Failure Modes

These are the failures you should expect.

### “The script worked in a probe but fails when embedded”

Cause:

- the wrapper or prelude changed the JavaScript shape

Fix:

- compare the embedded script and the probe directly
- confirm `SURF_OPTIONS` is valid JSON
- re-run the exact script body through `surf-go js`

### “The command only works if a tab is already open”

Cause:

- the fetch path implicitly relied on active-tab resolution

Fix:

- decide whether the command should create its own tab
- if yes, do `tab.new` explicitly and capture the resulting `tabId`

### “The command leaves junk tabs behind”

Cause:

- the command created a tab but never modeled ownership and cleanup explicitly

Fix:

- track whether the command created the tab
- close that tab by default after successful completion
- add an escape hatch such as `--keep-tab-open`
- never close a tab supplied by `--tab-id` or `--window-id`

### “The page returns something that looks structured but is actually useless”

Cause:

- a visible UI block exposed only labels or shell text

Fix:

- inspect the raw block contents
- suppress the block unless it contains meaningful payload

### “Rows work but Markdown does not”

Cause:

- the command only implements the glaze path

Fix:

- implement `RunIntoWriter(...)`
- register the command in dual mode

### “Markdown works but structured rows do not”

Cause:

- the command was registered only as a writer path

Fix:

- confirm `RunIntoGlazeProcessor(...)` exists
- confirm `cli.WithDualMode(true)` and `cli.WithGlazeToggleFlag("with-glaze-output")` are used

## Minimal Template

This is the smallest useful template for a new complex verb.

```text
research:
  use surf-go js to prove selectors and timing
  save probes in ticket scripts/

production:
  create go/internal/cli/commands/scripts/my_verb.js
  embed it with go:embed
  write buildMyVerbCode(settings)
  write fetchMyVerb(ctx, settings)
  write parseMyVerbResponse(resp)
  write myVerbDataToRows(data)
  write renderMyVerbMarkdown(data) if dual mode
  register command in main.go

validation:
  unit tests
  mock-host integration test
  real-browser validation
  owned-tab cleanup validation
  help page and discoverability check
```

That template is intentionally boring. Complex browser verbs become maintainable when the structure stays boring and the page-specific complexity stays isolated in the embedded script.

## See Also

- [kagi_search.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/kagi_search.go) — compact example of navigate-or-create-tab plus dual-mode output
- [chatgpt_transcript.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/chatgpt_transcript.go) — more complex example with subview scraping and export behavior
- [integration_test.go](/home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/cmd/surf-go/integration_test.go) — mock-host patterns for validating command tool sequences
- [writing-help-entries](glaze help writing-help-entries) — Glazed help frontmatter and embedding conventions
- [how-to-write-good-documentation-pages](glaze help how-to-write-good-documentation-pages) — style guidance for help pages
