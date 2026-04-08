# Changelog

## 2026-02-25

- Initial workspace created.
- Gathered exhaustive evidence for Node provider orchestration, service worker primitives, Go host/router constraints, and installer runtime profile logic.
- Added reproducible compatibility inventory script and generated source artifact.
- Authored full design research report, detailed diary, and provider contract matrix.
- Prepared ticket for validation and reMarkable delivery.

## 2026-02-25 - Exhaustive provider compatibility research completed

Completed evidence-backed architecture analysis, provider contract matrix, diary, and reproducibility artifacts; prepared migration plan for Go-host provider parity and documented runtime constraints for Chromium Snap.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/host/router/toolmap.go — Current provider blocklist
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/host.cjs — Provider orchestration baseline
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R3--go-host-provider-compatibility-research-chatgpt-gemini-perplexity-grok-ai-studio/sources/01-provider-compat-inventory.json — Generated compatibility inventory

## 2026-02-26 - ChatGPT file upload implementation (Node + Go parity)

- Added selector-capable `UPLOAD_FILE` handling in service worker (supports `selector` or legacy `ref`).
- Implemented Node ChatGPT upload path in `native/chatgpt-client.cjs` and wired callback in `native/host.cjs`.
- Implemented Go ChatGPT upload path in `go/internal/host/providers/chatgpt.go`.
- Added tests for Node (`test/unit/chatgpt-client.test.ts`) and Go (`go/internal/host/providers/chatgpt_test.go`) upload behavior.
- Captured test runs and outcomes in diary.

## 2026-04-07 - Completed Snap Go-host ChatGPT runtime debugging and polling parity work

- Added service-worker lifecycle logging for native-host request handling and ChatGPT evaluate results.
- Added runtime/socket identification to `HOST_READY` so the extension can distinguish `go-host` from `node-host`.
- Diagnosed Snap wrapper fallback behavior and confirmed Go-host activation once `~/snap/chromium/common/surf-cli/surf-host-go` was installed.
- Fixed Go ChatGPT polling to prefer direct assistant nodes and restored model-list parsing to canonical IDs, including portal-rendered legacy submenu items.
- Captured snap-private log access path (`snap run --shell chromium -c 'tail -n 120 /tmp/surf-host-go.log'`) and final working evidence in the investigation diary.
