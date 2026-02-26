# Surf Go Runtime Guide

This guide documents the Go runtime path for Surf:

- `surf-host-go`: Go native messaging host (used by browser extension through wrapper profile switching)
- `surf-go`: Go CLI client that talks to the local Surf socket

This complements `README.md` and focuses on install/verification/troubleshooting for `core-go`.

## Current Status (as of February 26, 2026)

- Runtime profile switching is supported via `SURF_HOST_PROFILE=node-full|core-go`.
- `core-go` supports core browser commands.
- ChatGPT provider flow is implemented in Go host.
- Other provider flows (Gemini/Perplexity/Grok/AI Studio) are still blocked in `go-core` profile.

## 1) Install Go Host (Recommended Path)

### Prerequisites

- Go installed (`go version` works)
- Node installed (`node -v` works)
- Surf extension loaded in browser and extension ID available

### Install

```bash
npm install -g surf-cli
surf install <extension-id> --browser chromium
```

The installer will:

- install native-host manifest(s)
- create host wrapper script
- build `surf-host-go` when Go + source are available
- enable profile switching in wrapper (`SURF_HOST_PROFILE`)

For Chrome, replace `--browser chromium` with `--browser chrome`.

## 2) Enable Go Runtime

Set runtime profile before launching browser:

```bash
export SURF_HOST_PROFILE=core-go
```

For Snap Chromium, also set CLI socket path:

```bash
export SURF_SOCKET_PATH=~/snap/chromium/common/surf-cli/surf.sock
```

Then launch browser and reload extension.

## 3) Verify Installation

Run installer output and confirm you see a hint like:

- `core-go profile available. Set SURF_HOST_PROFILE=core-go ...`

Check expected files on Linux:

```bash
ls -l ~/.local/share/surf-cli/surf-host-go
ls -l ~/.local/share/surf-cli/host-wrapper.sh
```

For Snap Chromium:

```bash
ls -l ~/snap/chromium/common/surf-cli/surf-host-go
ls -l ~/snap/chromium/common/surf-cli/host-wrapper.sh
```

Inspect manifest path (Chromium):

```bash
cat ~/.config/chromium/NativeMessagingHosts/surf.browser.host.json
cat ~/snap/chromium/common/chromium/NativeMessagingHosts/surf.browser.host.json
```

## 4) Quick Runtime Checks

Tail Go host log (default path):

```bash
tail -n 120 /tmp/surf-host-go.log
```

Custom log path:

```bash
export SURF_HOST_LOG=/tmp/my-surf-go.log
```

Test with Go CLI (talks to host socket):

```bash
cd go
go run ./cmd/surf-go tab list
go run ./cmd/surf-go page read
```

Test ChatGPT through raw tool path:

```bash
cd go
go run ./cmd/surf-go tool-raw --tool chatgpt --args-json '{"query":"say ping"}'
```

## 5) Build From Source (Repo Checkout)

If you are working from this repository:

```bash
npm install
npm run build
npm run build:go-host
```

Manual Go builds:

```bash
cd go
go build -o /tmp/surf-host-go ./cmd/surf-host-go
go build -o /tmp/surf-go ./cmd/surf-go
```

Run tests:

```bash
cd go
go test ./...
```

## 6) Common Errors and Fixes

### `Go host source not found in package; core-go profile unavailable.`

Cause:

- install is running from a package layout without `go/cmd/surf-host-go/main.go`

Fix:

- run install from this repository checkout, or reinstall package version that includes `go/`

```bash
cd /path/to/surf-cli-repo
npm install -g .
node scripts/install-native-host.cjs <extension-id> --browser chromium
```

### `Native host disconnected: Native host has exited.`

Check:

- host logs (`/tmp/surf-host-go.log` and `/tmp/surf-host.log`)
- extension reload/install events in service worker console
- socket alignment (`SURF_SOCKET_PATH`) when using Snap

### CLI cannot connect to socket

Set explicit socket path:

```bash
export SURF_SOCKET_PATH=/tmp/surf.sock
```

For Snap Chromium:

```bash
export SURF_SOCKET_PATH=~/snap/chromium/common/surf-cli/surf.sock
```

## 7) Runtime Selection Model

The wrapper decides runtime at launch time:

- `SURF_HOST_PROFILE=core-go` and `surf-host-go` exists -> run Go host
- otherwise -> fallback to Node host (`native/host.cjs`)

This means one install supports both runtimes without reinstalling.

## 8) Environment Variables Relevant to Go Runtime

- `SURF_HOST_PROFILE`: `node-full` (default) or `core-go`
- `SURF_SOCKET_PATH`: override socket path for CLI/host communication
- `SURF_HOST_LOG`: Go host log file path (default `/tmp/surf-host-go.log`)
- `SURF_GO_PATH`: Go binary used by installer when building `surf-host-go`
- `SURF_NODE_PATH`: Node binary path used by wrapper fallback
- `SURF_HOST_PATH`: Node host script path (`native/host.cjs`)

## 9) Known Scope Differences

- `surf-go` command surface is currently core browser-focused.
- `chatgpt` is available via host provider handling (use `tool-raw` from `surf-go` if needed).
- Gemini/Perplexity/Grok/AI Studio remain blocked in `go-core` profile until their Go handlers are implemented.

## 10) Reinstall/Reset Procedure

If runtime behavior is unclear, do a clean reinstall:

```bash
surf uninstall --all
surf install <extension-id> --browser chromium
export SURF_HOST_PROFILE=core-go
export SURF_SOCKET_PATH=~/snap/chromium/common/surf-cli/surf.sock   # snap only
```

Then restart browser and reload extension.
