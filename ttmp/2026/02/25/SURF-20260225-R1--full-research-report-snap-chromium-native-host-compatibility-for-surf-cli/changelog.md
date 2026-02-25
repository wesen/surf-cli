# Changelog

## 2026-02-25

- Initial workspace created
- Collected code-level evidence for installer, host/CLI socket wiring, and extension native messaging flow.
- Reproduced Snap Chromium runtime constraints and captured command-level outputs (`Permission denied` for wrapper/node in hidden paths, socket inode mismatch across snap/host `/tmp`).
- Researched external primary sources (Chrome native messaging, Chromium config directory behavior, Snap confinement/home interface, Launchpad Chromium snap issue history, Firefox portal design reference).
- Authored full design report with phased implementation plan and test strategy.
- Authored detailed chronological investigation diary.

## 2026-02-25

Completed full research report and detailed diary with source-backed Snap compatibility recommendations; pending reMarkable upload verification.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md — Primary report deliverable
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/reference/01-investigation-diary.md — Chronological command/evidence diary


## 2026-02-25

Uploaded bundled full research report to reMarkable at /ai/2026/02/25/SURF-20260225-R1 and verified remote listing.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/design-doc/01-snap-chromium-native-messaging-compatibility-research-report.md — Uploaded as part of final bundle
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/ttmp/2026/02/25/SURF-20260225-R1--full-research-report-snap-chromium-native-host-compatibility-for-surf-cli/reference/01-investigation-diary.md — Uploaded as part of final bundle

## 2026-02-25

Completed implementation Task 1: introduced shared socket-path resolution with `SURF_SOCKET_PATH` override across host/CLI/MCP/workflow executor and added unit tests. (commit `fbedaed`)

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/socket-path.cjs — Shared socket-path resolver
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/host.cjs — Host now reads socket path from shared resolver
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/cli.cjs — CLI now reads socket path from shared resolver
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/mcp-server.cjs — MCP server now reads socket path from shared resolver
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/do-executor.cjs — Workflow executor now reads socket path from shared resolver
- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/test/unit/socket-path.test.ts — Added tests for override/default behavior
