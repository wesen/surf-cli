# Tasks

## Completed

- [x] Create ticket workspace `SURF-20260225-R3`.
- [x] Create primary design doc and reference docs.
- [x] Collect line-anchored evidence from Node host mappings and provider orchestration.
- [x] Collect line-anchored evidence from service worker provider primitives.
- [x] Collect line-anchored evidence from Go host/router restrictions and transport.
- [x] Collect installer/runtime profile evidence for `node-full` and `core-go` (including Snap target behavior).
- [x] Create reproducible provider inventory script under ticket `scripts/`.
- [x] Generate provider inventory source artifact in `sources/`.
- [x] Document provider-specific logic and third-party package usage.
- [x] Write exhaustive architecture and migration research report.
- [x] Write chronological investigation diary with command history and findings.
- [x] Write provider compatibility matrix and contract reference.
- [x] Update ticket index summary and links.
- [x] Relate key repository files to ticket docs via `docmgr doc relate`.
- [x] Update changelog with research completion entry.
- [x] Run `docmgr doctor` and resolve/confirm warnings.
- [x] Upload bundled research docs to reMarkable (dry-run + real upload + verify listing).

## Follow-up Implementation Backlog (not executed in this research ticket)

- [ ] Remove provider blocklist gate in `go/internal/host/router/toolmap.go` behind feature flag or phased rollout.
- [ ] Add Go provider dispatcher package and host runtime integration.
- [ ] Implement provider handlers one-by-one (ChatGPT -> Perplexity -> Gemini -> Grok -> AI Studio -> AI Studio Build).
- [ ] Add Node-vs-Go output parity fixture tests for provider commands.
- [ ] Add runtime self-check for profile/socket alignment (especially Snap Chromium).

## ChatGPT Integration Implementation Plan (current execution)

- [ ] Task 1: Add host-side provider bridge primitives for internal extension request/response roundtrips (timeout-aware, ID-correlated).
- [ ] Task 2: Implement `go/internal/host/providers/chatgpt.go` with ChatGPT orchestration flow (cookies, tab create/close, CDP eval/command, prompt submit, response wait).
- [ ] Task 3: Integrate ChatGPT provider dispatch into `go/cmd/surf-host-go/main.go` before router mapping.
- [ ] Task 4: Add focused unit tests for ChatGPT provider flow with mocked bridge behavior.
- [ ] Task 5: Add/adjust host and router tests so `chatgpt` is handled by provider path while other providers remain blocked.
- [ ] Task 6: Run `go test ./...` in `go/` and fix regressions.
- [ ] Task 7: Validate local CLI behavior against running extension (if environment available), otherwise capture exact manual test handoff for browser-side verification.
- [ ] Task 8: Update diary with detailed command log/results and commit each completed task incrementally.
