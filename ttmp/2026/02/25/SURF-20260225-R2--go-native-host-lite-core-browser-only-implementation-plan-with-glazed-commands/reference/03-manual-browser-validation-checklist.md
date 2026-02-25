---
Title: Manual Browser Validation Checklist
Ticket: SURF-20260225-R2
Status: active
Topics:
    - go
    - chromium
    - native-messaging
    - validation
DocType: reference
Intent: short-term
Owners: []
RelatedFiles:
    - Path: go/cmd/surf-host-go/main.go
      Note: Go host runtime behavior under real browser connection
    - Path: go/cmd/surf-go/main.go
      Note: Go CLI command wiring used during manual checks
    - Path: scripts/install-native-host.cjs
      Note: Installer/runtime profile setup for core-go and node-full
ExternalSources: []
Summary: Manual real-browser checklist to validate core-go profile behavior in Chromium/Snap before rollout.
LastUpdated: 2026-02-25T18:25:00-05:00
WhatFor: Capture reproducible human-in-the-loop validation steps and expected results
WhenToUse: Use when executing T6.8 before enabling go profile by default
---

# Manual Browser Validation Checklist

## Preconditions

1. Install native host for Chromium and restart browser:
   - `surf install <extension-id> --browser chromium`
2. Set profile and socket for your runtime:
   - `export SURF_HOST_PROFILE=core-go`
   - For Snap Chromium: `export SURF_SOCKET_PATH=~/snap/chromium/common/surf-cli/surf.sock`
3. Confirm extension service worker is running and no immediate disconnect loop appears.

## Core command smoke

1. `surf-go tab list --output json`
   - Expect tab list response (no `extension_disconnected`).
2. `surf-go page read --args-json '{}' --output json`
   - Expect readable page snapshot result.
3. `surf-go click --args-json '{"ref":"e1"}' --output json`
   - Expect tool response (success or meaningful page-specific error, not host crash).
4. `surf-go screenshot --args-json '{"output":"/tmp/surf-go-check.png"}' --output json`
   - Expect screenshot result and file exists.

## Group command coverage

1. `surf-go window list --output json`
2. `surf-go frame list --output json`
3. `surf-go dialog info --output json`
4. `surf-go network list --output json`
5. `surf-go console read --output json`
6. `surf-go cookie list --output json`

All commands should return normalized `tool_response` envelopes and not terminate the host unexpectedly.

## Stream behavior

1. `surf-go console stream --duration-sec 5 --output json`
2. `surf-go network stream --duration-sec 5 --output json`

Expected:
1. stream starts successfully,
2. events emit as rows,
3. stream exits cleanly after timeout without orphaned socket errors.

## Disconnect handling

1. While a command is running, reload the extension in Chromium.
2. Re-run one command (for example, `surf-go page read --output json`).

Expected:
1. existing clients receive `extension_disconnected` style failure,
2. rerun succeeds after extension is back,
3. host log (`/tmp/surf-host-go.log`) shows clean restart path.

## Fallback profile check

1. `export SURF_HOST_PROFILE=node-full`
2. Re-run a few commands above (`tab list`, `page read`, `network list`).

Expected:
1. node fallback still works,
2. behavior remains compatible with prior installs.

## Record results

For each step, record:
1. pass/fail,
2. command output (or key line),
3. host log snippet if failed,
4. whether failure is `core-go` only or both profiles.
