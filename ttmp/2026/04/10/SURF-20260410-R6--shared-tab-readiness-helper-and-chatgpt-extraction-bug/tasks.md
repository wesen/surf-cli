# Tasks

## TODO

- [x] Write a design doc for a shared fresh-tab readiness helper.
- [x] Write a detailed bug report for the ChatGPT extraction mismatch.
- [x] Add a shared tab-readiness helper in `go/internal/cli/commands`.
- [x] Make the helper enforce exact `tabId` ownership plus a JS readiness probe.
- [x] Update `kagi-search` to use the shared helper instead of local sleep/retry logic.
- [x] Update `kagi-assistant` to use the shared helper instead of local sleep/retry logic.
- [x] Update command integration tests to expect `tab.new` -> readiness probe -> main JS -> `tab.close`.
- [x] Review the interactive ChatGPT provider extraction logic against the transcript extractor.
- [x] Patch the ChatGPT provider to use turn-based assistant extraction rather than the weaker last-node heuristic.
- [x] Update provider tests for the new extraction path.
- [x] Live-validate `kagi-search` after the shared helper is in place.
- [ ] Finish live-validation of the interactive ChatGPT path through to final command return on a long response.
- [ ] Decide whether the long-running ChatGPT completion gate should continue to require `stopVisible == false` for research-heavy responses.
- [ ] Relate the final code changes to this ticket with `docmgr doc relate`.
- [ ] Run `go test ./internal/cli/commands ./internal/host/providers ./cmd/surf-go` from the `go/` module root.
- [ ] Run `docmgr doctor --ticket SURF-20260410-R6 --stale-after 30`.
