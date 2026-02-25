---
Title: Investigation Diary
Ticket: SURF-20260225-R1
Status: active
Topics:
    - linux
    - chromium
    - snap
    - native-messaging
    - debugging
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Diary includes updated snap socket guidance and diagnostics usage
    - Path: native/cli.cjs
      Note: Diary captures CLI error outcomes from socket connection failures
    - Path: native/host.cjs
      Note: Diary captures socket and host lifecycle evidence from this runtime
    - Path: scripts/install-native-host.cjs
      Note: Commands and outcomes depend on installer path/wrapper logic
    - Path: scripts/uninstall-native-host.cjs
      Note: Diary includes snap-aware uninstall implementation and validation
ExternalSources: []
Summary: Chronological command log and reasoning trail for full Snap Chromium compatibility research and report delivery.
LastUpdated: 2026-02-25T16:04:46.979814202-05:00
WhatFor: Preserve reproducible investigation history for future implementation
WhenToUse: Use to reproduce findings and verify assumptions during implementation
---



# Investigation Diary

## Goal

Produce a full research report for Snap Chromium native host compatibility in Surf CLI, backed by local reproduction and external primary sources, then upload the deliverable bundle to reMarkable.

## Prompt Context

**User prompt (verbatim):** "write a full research report and upload to remarkable. $ticket-research-docmgr-remarkable"

**Assistant interpretation:** Create a new docmgr ticket with a full design-grade research report + detailed diary, run quality checks, and upload a bundled PDF to reMarkable.

**Inferred user intent:** Obtain a durable, shareable engineering document suitable for implementation planning and review.

**Commit (code):** N/A (documentation/research only)

## Phase 1: Ticket setup and doc scaffolding

### Commands run

```bash
docmgr ticket create-ticket --ticket SURF-20260225-R1 --title "Full research report: Snap Chromium native host compatibility for Surf CLI" --topics linux,chromium,snap,native-messaging,debugging,architecture
docmgr doc add --ticket SURF-20260225-R1 --doc-type design-doc --title "Snap Chromium Native Messaging Compatibility Research Report"
docmgr doc add --ticket SURF-20260225-R1 --doc-type reference --title "Investigation Diary"
docmgr doc list --ticket SURF-20260225-R1
```

### Findings

1. Ticket workspace and required docs were created successfully.
2. Paths created:
   - `.../design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md`
   - `.../reference/01-investigation-diary.md`

### What worked

- `docmgr` ticket and document creation workflow completed without errors.

### What didn't work

- N/A.

## Phase 2: Repository evidence collection (line-anchored)

### Commands run

```bash
nl -ba scripts/install-native-host.cjs | sed -n '1,260p'
nl -ba scripts/install-native-host.cjs | sed -n '260,380p'
nl -ba scripts/uninstall-native-host.cjs | sed -n '1,220p'
nl -ba native/host.cjs | sed -n '1,120p'
nl -ba native/host.cjs | sed -n '1520,1668p'
nl -ba native/cli.cjs | sed -n '1,60p'
nl -ba native/cli.cjs | sed -n '2920,2960p'
nl -ba src/native/port-manager.ts | sed -n '1,160p'
nl -ba README.md | sed -n '648,670p'
```

### Findings

1. Linux install path assumptions:
   - manifest: `.config/chromium/NativeMessagingHosts`
   - wrapper dir: `.local/share/surf-cli`
2. Host and CLI use fixed Linux socket path: `/tmp/surf.sock`.
3. Extension uses `connectNative("surf.browser.host")` and logs disconnect reasons.
4. README marks Linux support as experimental.

### What worked

- Evidence clearly supports architecture and gap analysis sections.

### What didn't work

- N/A.

## Phase 3: Local runtime reproduction and failure isolation

### Commands run

```bash
chromium --version && snap list chromium
surf install jgbccjmbpfinlegbjkomjeoiklaijmmm --browser chromium
surf tab.list || true
snap run --shell chromium -c 'echo HOME=$HOME; echo CHROME_CONFIG_HOME=$CHROME_CONFIG_HOME; echo SNAP_USER_COMMON=$SNAP_USER_COMMON; echo XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR'
snap run --shell chromium -c '/home/manuel/.local/share/surf-cli/host-wrapper.sh </dev/null >/tmp/snap-wrap.out 2>/tmp/snap-wrap.err; echo EXIT:$?; cat /tmp/snap-wrap.err'
snap run --shell chromium -c '/home/manuel/.nvm/versions/node/v22.21.0/bin/node --version; echo EXIT:$?'
stat -c 'host %n inode=%i mtime=%y' /tmp/surf.sock 2>/dev/null || echo 'host /tmp/surf.sock missing'
snap run --shell chromium -c 'stat -c "snap %n inode=%i mtime=%y" /tmp/surf.sock 2>/dev/null || echo "snap /tmp/surf.sock missing"'
```

### Key outputs captured

1. `Chromium 145.0.7632.109 snap`
2. Installer output:
   - wrapper: `/home/manuel/.local/share/surf-cli/host-wrapper.sh`
   - manifest: `/home/manuel/.config/chromium/NativeMessagingHosts/surf.browser.host.json`
3. CLI failure:
   - `Error: Connection refused. Native host not running.`
4. Snap env:
   - `HOME=/home/manuel/snap/chromium/3369`
   - `CHROME_CONFIG_HOME=/home/manuel/snap/chromium/common`
   - `XDG_RUNTIME_DIR=/run/user/1000/snap.chromium`
5. Snap execution failures:
   - wrapper: `EXIT:126 ... Permission denied`
   - node in `.nvm`: `EXIT:126 ... Permission denied`
6. Socket inode mismatch:
   - host `/tmp/surf.sock` inode `8995165`
   - snap `/tmp/surf.sock` inode `12069922`

### What worked

- Reproduction yielded concrete, repeatable symptoms and failure classes.

### What didn't work

- No direct functional connection path from Snap Chromium to current Surf host installation.

### What was tricky to build

- There are multiple independent blockers. Fixing only manifest path does not solve host executable access or socket namespace mismatch.

## Phase 4: External research and source validation

### Commands run

Used `web.search_query`, `web.open`, and `web.find` to collect primary-source evidence from:

1. Chrome Extensions native messaging docs.
2. Chromium source docs (`user_data_dir.md`).
3. Snap docs (`snap-confinement`, `home-interface`).
4. Launchpad Chromium Snap native-messaging bug discussion.
5. Firefox portal design doc (as sandboxed native messaging reference model).

### Findings

1. Native host launch is browser-driven; host path rules are strict (absolute path on Linux/macOS).
2. Chromium config roots can vary via env (`CHROME_CONFIG_HOME`, `XDG_CONFIG_HOME`).
3. Snap strict confinement and `home` interface rules explain hidden-path access issues.
4. Historical Chromium Snap bug discussion confirms long-running native messaging friction.
5. Portal mediation exists in Firefox design docs, but no equivalent proven Chromium path in this investigation.

### What worked

- Sources aligned with local repro and reduced speculative guidance.

### What didn't work

- No single upstream Chromium Snap “drop-in recipe” that solves all Surf-specific needs without app changes.

## Phase 5: Report authoring

### What I wrote

1. Full design doc with:
   - executive summary,
   - scope,
   - evidence-backed current state,
   - gap analysis,
   - proposed architecture + pseudocode,
   - phased implementation plan,
   - test strategy,
   - risks/alternatives/open questions,
   - internal + external references.
2. This diary with chronological commands and results.

### Why

- User requested a full research report suitable for implementation and distribution.

## Phase 6: Bookkeeping and quality checks

### Commands run

```bash
docmgr doc relate --doc <design-doc> --file-note "..."
docmgr doc relate --doc <diary-doc> --file-note "..."
docmgr changelog update --ticket SURF-20260225-R1 --entry "Completed full research report and diary with evidence-backed recommendations; ready for remarkable upload." --file-note "..."
# vocabulary cleanup if needed
docmgr vocab add --category topics --slug architecture --description "Architecture design and system-structure analysis"
docmgr doctor --ticket SURF-20260225-R1 --stale-after 30
```

### Findings

- `docmgr doctor` pass required adding `architecture` topic vocabulary.

### What worked

- Doctor checks passed after vocabulary alignment.

### What didn't work

- Initial doctor run would warn if vocabulary topic slugs are missing.

## Phase 7: reMarkable delivery

### Planned commands

```bash
remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run <index> <design-doc> <diary> <changelog> <tasks> <readme> \
  --name "SURF-20260225-R1 Snap Chromium Native Messaging Full Research Report" \
  --remote-dir "/ai/2026/02/25/SURF-20260225-R1" \
  --toc-depth 2
remarquee upload bundle <same files> ...
remarquee cloud ls /ai/2026/02/25/SURF-20260225-R1 --long --non-interactive
```

### Delivery status

- Pending at this diary stage. (Upload executed immediately after final validation.)

## Code review instructions

1. Start in the design doc:
   - `.../design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md`
2. Validate key repository references:
   - `scripts/install-native-host.cjs`
   - `scripts/uninstall-native-host.cjs`
   - `native/host.cjs`
   - `native/cli.cjs`
   - `src/native/port-manager.ts`
3. Re-run reproduction commands from Phase 3 to confirm behaviors.

## What warrants a second pair of eyes

1. Final socket-path default strategy in mixed Chrome + Snap Chromium environments.
2. Update/upgrade semantics for copied Snap runtime assets.
3. Long-term feasibility of portal-based approach for Chromium if ecosystem changes.

## What should be done in the future

1. Implement Phase 1 (shared socket path override) first as low-risk foundation.
2. Add Snap-aware installer/uninstaller pathing.
3. Add explicit `surf doctor` diagnostics for native messaging on Linux/Snap.
4. Add integration test checklist in CI docs for Snap-specific manual verification.

## Related

- Design doc: `../design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md`
- Ticket index: `../index.md`
- Changelog: `../changelog.md`

## Phase 8: Implementation kickoff and Task 1 completion

I moved from research into implementation mode by creating a concrete task backlog in the same ticket and then executing Task 1 end-to-end with tests and a code commit. The first implementation goal was to remove hardcoded socket coupling by introducing a shared resolver and env override.

I intentionally completed Task 1 as an isolated change to keep risk low and make review straightforward before touching installer/runtime behavior for Snap in later tasks.

### Prompt Context

**User prompt (verbatim):** "ok, create tasks in the ticket for implementation, then work task by task, committing at regular intervals, keeping a detailed diary as you go. When you need input from me running in the real browser (in case you can't run tests from the CLI), stop and let me know."

**Assistant interpretation:** Create a task-by-task implementation backlog, execute sequentially with regular commits, and keep this diary updated after each step.

**Inferred user intent:** Controlled, auditable implementation progress with clear checkpoints and explicit request for real-browser help only when required.

**Commit (code):** `fbedaed` — "feat: add configurable surf socket path"

### What I did

1. Replaced research-complete checklist with implementation backlog (`T1`..`T6`) in `tasks.md`.
2. Added `native/socket-path.cjs` with `getSocketPath()` and `SURF_SOCKET_PATH` override support.
3. Updated socket usage to read from shared resolver in:
   - `native/host.cjs`
   - `native/cli.cjs`
   - `native/mcp-server.cjs`
   - `native/do-executor.cjs`
4. Updated docs:
   - `README.md` environment variables now include `SURF_SOCKET_PATH`.
5. Added targeted unit test:
   - `test/unit/socket-path.test.ts`
6. Ran targeted tests:
   - `npm run test -- test/unit/socket-path.test.ts`
   - `npm run test -- test/unit/do-executor.test.ts`

### Why

- Snap compatibility needs a configurable transport path to avoid hardcoded `/tmp/surf.sock` assumptions.
- A shared resolver prevents path drift across host/CLI/MCP/workflow components.

### What worked

- New tests passed (`3/3` for socket-path test file).
- Existing `do-executor` tests remained green (`21/21`), confirming no behavior regression in workflow execution path.
- Commit captured the task cleanly.

### What didn't work

- N/A for this phase.

### What I learned

- Centralizing socket-path resolution is low-risk and immediately useful beyond Snap support (custom deployments, isolated test environments).

### What was tricky to build

- The only notable risk was missing one socket consumer; search-and-update across all native entrypoints prevented partial adoption.

### What warrants a second pair of eyes

- Whether we should also route any future socket users through the same module by lint/guardrail to avoid reintroducing hardcoded paths.

### What should be done in the future

- Proceed with Task 2 and Task 3 (Snap-aware installer/uninstaller and runtime placement).

### Code review instructions

1. Review shared resolver:
   - `native/socket-path.cjs`
2. Review integration points:
   - `native/host.cjs`
   - `native/cli.cjs`
   - `native/mcp-server.cjs`
   - `native/do-executor.cjs`
3. Validate tests:
   - `npm run test -- test/unit/socket-path.test.ts`
   - `npm run test -- test/unit/do-executor.test.ts`

## Phase 9: Tasks 2-4 implementation (Snap install/uninstall + diagnostics)

I completed the next three tasks as separate commits to keep review and rollback easy: installer support, uninstaller support, and user-facing diagnostics/docs. Each change was validated locally with script checks and runtime probes.

This phase intentionally focused on deterministic file/runtime path handling and guidance; it does not yet include full real-browser end-to-end verification, which is tracked in Task 5.

### Prompt Context

**User prompt (verbatim):** (same as Phase 8)

**Assistant interpretation:** Continue sequential task execution with regular commits and detailed diary updates.

**Inferred user intent:** Deliver implementation progress in small, auditable increments.

**Commit (code):**
- `cf9cba6` — "feat: add snap-aware chromium install targets"
- `f5f170d` — "feat: add snap-aware chromium uninstall cleanup"
- `b7ff64a` — "docs: add snap socket diagnostics and guidance"

### What I did

1. Task 2 (`cf9cba6`) in `scripts/install-native-host.cjs`:
   - Added snap-root detection for Chromium (`~/snap/chromium/common`).
   - Added dual-target install for Linux Chromium:
     - standard manifest target (`~/.config/chromium/...`)
     - snap manifest target (`~/snap/chromium/common/chromium/...`)
   - Added snap runtime preparation:
     - copies surf package runtime into `~/snap/chromium/common/surf-cli/runtime/surf-cli`
     - copies node binary to `~/snap/chromium/common/surf-cli/node`
     - writes snap wrapper exporting `SURF_SOCKET_PATH=~/snap/chromium/common/surf-cli/surf.sock`
   - Added installer hint instructing CLI socket env export for non-snap shell usage.

2. Task 3 (`f5f170d`) in `scripts/uninstall-native-host.cjs`:
   - Added multi-path manifest removal for Linux Chromium:
     - standard + snap manifest paths.
   - Added snap wrapper directory cleanup (`~/snap/chromium/common/surf-cli`) when `--all`.

3. Task 4 (`b7ff64a`) in `native/cli.cjs` and `README.md`:
   - CLI now prints effective socket path on connection errors.
   - Added snap-specific hint if snap Chromium manifest is present and socket path is default.
   - README updated with:
     - `SURF_SOCKET_PATH` use case guidance,
     - Linux snap note in install section,
     - socket override guidance for snap Chromium users.

### Why

- Installer/uninstaller needed first-class snap path awareness to avoid manual manifest/runtime copying.
- Error diagnostics needed to surface the hidden socket mismatch root cause directly to users.

### What worked

1. Script syntax checks:
   - `node --check scripts/install-native-host.cjs`
   - `node --check scripts/uninstall-native-host.cjs`
2. Installer validation:
   - `node scripts/install-native-host.cjs <extension-id> --browser chromium`
   - confirmed creation of both manifests (standard + snap).
   - confirmed snap wrapper runtime paths and socket export.
3. Snap wrapper runtime probe:
   - `snap run --shell chromium -c '/home/manuel/snap/chromium/common/surf-cli/host-wrapper.sh ...'`
   - returned `EXIT:0` with HOST_READY payload bytes.
4. Uninstaller validation:
   - `node scripts/uninstall-native-host.cjs --browser chromium`
   - removed both standard + snap manifests.
5. Targeted tests still pass:
   - `npm run test -- test/unit/socket-path.test.ts`
   - `npm run test -- test/unit/do-executor.test.ts`
6. CLI diagnostics probe:
   - `node native/cli.cjs tab.list`
   - now outputs socket path + snap hint.

### What didn't work

- During early install-script iteration, snap runtime node copy hit:
  - `ETXTBSY: text file is busy, copyfile '/usr/bin/node' -> '.../snap/chromium/common/surf-cli/node'`
- I fixed it by treating `ETXTBSY` as non-fatal when target node already exists.

### What I learned

- In active sessions, replacing a currently executing snap-side node binary can fail; installer must tolerate that while preserving functional runtime.

### What was tricky to build

- The install pipeline had to support both “regular Linux Chromium” and “Snap Chromium” simultaneously without breaking existing behavior.
- The most fragile point was runtime staging and wrapper generation order (copy runtime first, then wrapper, then manifest).

### What warrants a second pair of eyes

- Whether snap runtime staging should eventually avoid full package copy (size/perf tradeoff) in favor of slimmer runtime artifacts.
- Whether `--all` semantics in uninstaller should continue removing all browser targets when combined with `--browser`.

### What should be done in the future

- Execute Task 5 real-browser verification with user assistance to confirm extension-side connect/disconnect behavior and command execution.

### Code review instructions

1. Installer changes:
   - `scripts/install-native-host.cjs`
2. Uninstaller changes:
   - `scripts/uninstall-native-host.cjs`
3. Diagnostics/docs:
   - `native/cli.cjs`
   - `README.md`
4. Validation commands:
   - `node scripts/install-native-host.cjs <extension-id> --browser chromium`
   - `node scripts/uninstall-native-host.cjs --browser chromium`
   - `node native/cli.cjs tab.list`

## Phase 10: Task 5 local validation complete, awaiting real-browser verification

I ran local validation checks for syntax, targeted tests, installer/uninstaller behavior, and ticket doc health. At this point, the remaining verification requires an actual Chromium extension session, which cannot be fully simulated via CLI-only checks.

### Prompt Context

**User prompt (verbatim):** (same as Phase 8)

**Assistant interpretation:** Continue until a real-browser-dependent checkpoint requires user input, then pause and request it.

**Inferred user intent:** Keep momentum with autonomous implementation while preserving correctness at browser integration boundaries.

**Commit (code):** N/A (validation checkpoint)

### What I did

1. Verified script syntax:
   - `node --check scripts/install-native-host.cjs`
   - `node --check scripts/uninstall-native-host.cjs`
2. Verified installer/uninstaller behavior locally:
   - install creates both standard + snap Chromium manifests
   - uninstall removes both targets
3. Verified targeted tests:
   - `npm run test -- test/unit/socket-path.test.ts`
   - `npm run test -- test/unit/do-executor.test.ts`
4. Verified ticket quality:
   - `docmgr doctor --ticket SURF-20260225-R1 --stale-after 30` => all checks passed.

### Why

- Ensures implementation tasks are internally consistent before asking for user-assisted browser verification.

### What worked

- All local checks passed.

### What didn't work

- Full extension/native-host handshake verification from this environment alone (requires user-driven browser session).

### What should be done in the future

- Run user-assisted browser verification flow and then complete Task 5 and Task 6.

## Phase 11: User-provided post-fix runtime evidence (extension still disconnects)

After the user re-tested in a real browser session, they still observed repeated service-worker disconnect messages.

### Prompt Context

**User prompt (verbatim):**

> I still get 
> debug.ts:28 [Surf] Connecting to native host...
> debug.ts:28 [Surf] Native host disconnected: Native host has exited.
> debug.ts:28 [Surf] Connecting to native host...
> debug.ts:28 [Surf] Native host disconnected: Native host has exited. in the service worker, after reloading the extension.

**Additional user evidence (verbatim excerpt):**

> 2026-02-25T21:44:18.507Z Host starting...
> 2026-02-25T21:44:18.511Z Host initialization complete, waiting for connections...
> 2026-02-25T21:44:18.511Z Socket server listening on /tmp/surf.sock
> 2026-02-25T21:44:18.512Z Sent HOST_READY to extension
> 2026-02-25T21:44:18.513Z stdin ended (extension disconnected), notifying clients

### Interpretation

1. Host process does start and emit `HOST_READY`.
2. Host then sees stdin close almost immediately, which means extension side disconnects after launch.
3. This is not consistent with a pure “host failed to start” scenario.

### Why this mattered

It narrowed the problem from startup failure to lifecycle/disconnect behavior and environment mismatch.

## Phase 12: Timestamp/time-zone correlation check

The user asked whether timestamp differences might be time-zone related.

### Prompt Context

**User prompt (verbatim):**

> for the record 2026-02-25T21:44:18.507Z ...
> date
> Wed Feb 25 04:51:15 PM EST 2026
> which might be some time zone stuff?

### Interpretation

1. `Z` timestamps in host log are UTC.
2. User terminal `date` is EST (`UTC-05:00`).
3. `21:44:18Z` corresponds to `16:44:18 EST` on 2026-02-25.

### Conclusion

Timestamp offset is expected and not itself evidence of the disconnect root cause.

## Phase 13: Node runtime scope + Go migration feasibility investigation

User asked whether installer copies Node into extension and whether a Go version could avoid Node install complexity.

### Prompt Context

**User prompt (verbatim):**

> so the installer script copies node into the extension folder?
>
> ok, this node script, how much does it do and can we maybe make a go version so we don't have to deal with installing a whole node?

### Commands run

```bash
rg --files native scripts | sort
wc -l native/*.cjs scripts/install-native-host.cjs scripts/uninstall-native-host.cjs
sed -n '1,260p' native/host.cjs
sed -n '260,620p' native/host.cjs
sed -n '620,1040p' native/host.cjs
sed -n '1040,1668p' native/host.cjs
sed -n '1,260p' native/cli.cjs
cat package.json
rg -n "function writeMessage|function processInput|HOST_READY|process.stdin.on\(\"end\"|const server = net.createServer|function handleToolRequest" native/host.cjs
rg -n "if \(extensionMsg\.type === \"(CHATGPT_QUERY|PERPLEXITY_QUERY|GEMINI_QUERY|GROK_QUERY|GROK_VALIDATE|AISTUDIO_QUERY|AISTUDIO_BUILD|AI_ANALYZE)\"\)" native/host.cjs
nl -ba scripts/install-native-host.cjs | sed -n '1,220p'
nl -ba scripts/install-native-host.cjs | sed -n '220,420p'
nl -ba native/socket-path.cjs
nl -ba native/protocol.cjs
git status --short
git diff -- scripts/install-native-host.cjs
```

### Findings

1. The installer does **not** copy Node into extension files.
2. For Linux Chromium Snap target, installer copies Node and runtime package under `~/snap/chromium/common/surf-cli/...` and writes a snap-side wrapper/manifest.
3. The host runtime (`native/host.cjs`) is large and multi-responsibility, including provider automation clients and transport orchestration.

### Key evidence

1. Snap runtime copy behavior: `scripts/install-native-host.cjs:134-171`, `:281-297`.
2. Node host responsibilities and size: `native/host.cjs` (~1668 LOC), provider handler branches at `:399`, `:454`, `:555`, `:649`, `:722`, `:842`, `:919`, `:1034`.
3. Extension disconnect handling: `native/host.cjs:1506-1520`.

### Outcome

Prepared recommendation for phased migration:

1. short-term packaging/runtime hardening,
2. medium-term Go host lite for transport + core actions,
3. long-term provider parity or sidecar model.

## Phase 14: Ticket deliverable work for Node-vs-Go research + upload request

The user requested a dedicated research doc in the ticket, updated diary, and reMarkable upload.

### Prompt Context

**User prompt (verbatim):**

> Create a research doc in the ticket and upload to remarkable. Also keep diary entries for all the stuff we did so far.

### What I did

1. Loaded skill workflow and writing/checklist references:
   - `/home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md`
   - `references/writing-style.md`
   - `references/deliverable-checklist.md`
2. Added design doc:
   - `docmgr doc add --ticket SURF-20260225-R1 --doc-type design-doc --title "Node Native Host Scope and Go Migration Feasibility"`
3. Drafted full evidence-based design report in:
   - `design-doc/02-node-native-host-scope-and-go-migration-feasibility.md`
4. Appended this diary with Phases 11-14.

### Pending after this phase

1. Update ticket index/tasks/changelog and relate files.
2. Run `docmgr doctor --ticket SURF-20260225-R1 --stale-after 30`.
3. Run reMarkable dry-run bundle upload, then real upload, then verify cloud listing.

## Phase 15: Validation + reMarkable delivery completion

Completed the publication workflow requested by the user.

### Commands run

```bash
docmgr doctor --ticket SURF-20260225-R1 --stale-after 30
remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/02-node-native-host-scope-and-go-migration-feasibility.md \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/reference/01-investigation-diary.md \
  --name "SURF-20260225-R1-research-update-node-go-feasibility" \
  --remote-dir "/ai/2026/02/25/SURF-20260225-R1" \
  --toc-depth 2
remarquee upload bundle \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/02-node-native-host-scope-and-go-migration-feasibility.md \
  ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/reference/01-investigation-diary.md \
  --name "SURF-20260225-R1-research-update-node-go-feasibility" \
  --remote-dir "/ai/2026/02/25/SURF-20260225-R1" \
  --toc-depth 2
remarquee cloud ls /ai/2026/02/25/SURF-20260225-R1 --long --non-interactive
```

### Results

1. `docmgr doctor` passed (`All checks passed`).
2. Dry-run bundle upload succeeded.
3. Real bundle upload succeeded:
   - `OK: uploaded SURF-20260225-R1-research-update-node-go-feasibility.pdf -> /ai/2026/02/25/SURF-20260225-R1`
4. Cloud listing confirms presence of both report bundles:
   - `SURF-20260225-R1 Snap Chromium Native Messaging Full Research Report`
   - `SURF-20260225-R1-research-update-node-go-feasibility`

### Bookkeeping updates completed

1. Added new design doc link in ticket `index.md`.
2. Updated tasks (`T6`, `T7`) to complete.
3. Added changelog entries for research update and upload verification.

### Remaining open work

1. `T5` remains open pending additional real-browser validation to close the disconnect loop.
