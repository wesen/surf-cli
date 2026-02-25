---
Title: Snap Chromium Native Messaging Compatibility Research Report
Ticket: SURF-20260225-R1
Status: active
Topics:
    - linux
    - chromium
    - snap
    - native-messaging
    - debugging
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Declared Linux support status and install guidance
    - Path: native/cli.cjs
      Note: CLI socket connection path and user-facing ENOENT/ECONNREFUSED behavior
    - Path: native/host.cjs
      Note: Native host socket binding
    - Path: scripts/install-native-host.cjs
      Note: Linux install path assumptions and wrapper generation behavior
    - Path: scripts/uninstall-native-host.cjs
      Note: Mirror cleanup behavior and current Linux path assumptions
    - Path: src/native/port-manager.ts
      Note: Extension connectNative/disconnect flow and reconnect behavior
ExternalSources: []
Summary: Evidence-based research report on why Surf CLI native messaging fails with Snap Chromium and how to implement robust compatibility.
LastUpdated: 2026-02-25T16:04:46.803033125-05:00
WhatFor: Guide implementation of Snap Chromium support for Surf native messaging
WhenToUse: Use when implementing or reviewing Linux Chromium/Snap native messaging support
---


# Snap Chromium Native Messaging Compatibility Research Report

## 1. Executive Summary

This report investigates why `surf-cli` fails to maintain native host connectivity with Chromium installed as a Snap package on Linux, and proposes an implementation plan to make the integration reliable.

The core findings are:

1. Surf installs Chromium manifests and wrappers in paths that work for traditional Linux Chromium installs but are incompatible with Snap Chromium confinement assumptions.
2. Snap Chromium runs with a rewritten runtime environment (`HOME`, `CHROME_CONFIG_HOME`, `XDG_RUNTIME_DIR`) and strict confinement constraints that affect executable access and host-manifest resolution.
3. Surf hardcodes a Unix socket path (`/tmp/surf.sock`) in both host and CLI. In Snap scenarios this can resolve to different namespaces, producing host/CLI split-brain.
4. A robust solution requires Snap-aware manifest placement, Snap-accessible host runtime placement, and a shared socket path strategy with explicit configurability.

## 2. Problem Statement and Scope

### 2.1 Problem

Users running Chromium as a Snap on Linux report the extension repeatedly disconnecting from the native host, with logs similar to:

- `Connecting to native host...`
- `Native host disconnected: Native host has exited.`

CLI-side symptoms include:

- `Error: Socket not found. Is Chrome running with the extension?`
- `Error: Connection refused. Native host not running.`

### 2.2 Scope

In scope:

1. Surf native host install/runtime pathing on Linux.
2. Chromium extension ↔ native host ↔ Surf CLI communication chain.
3. Snap-specific behavior affecting filesystem/process/socket access.
4. Concrete implementation plan for Surf repository.

Out of scope:

1. Generic extension correctness unrelated to native messaging transport.
2. Re-architecting Surf protocol beyond compatibility needs.
3. Non-Linux platform behavior (macOS/Windows).

## 3. Current-State Architecture (Evidence-Based)

### 3.1 Surf installer writes Linux Chromium manifests to `~/.config/chromium/NativeMessagingHosts`

Evidence:

- Linux Chromium manifest path in installer: `scripts/install-native-host.cjs:16-20`
- Manifest install directory uses `path.join(os.homedir(), browserConfig[platform])`: `scripts/install-native-host.cjs:154-167`
- Wrapper dir on Linux is fixed to `~/.local/share/surf-cli`: `scripts/install-native-host.cjs:91-104`

Interpretation:

- Installer assumes a conventional Linux home/config model and a single per-user manifest location.

### 3.2 Uninstaller mirrors the same assumptions

Evidence:

- Linux manifest removal path is `~/.config/chromium/NativeMessagingHosts`: `scripts/uninstall-native-host.cjs:16-20, 75-79`
- Wrapper dir removal targets `~/.local/share/surf-cli`: `scripts/uninstall-native-host.cjs:48-56, 101-107`

### 3.3 Extension attempts long-lived native connection via `connectNative`

Evidence:

- Native connect: `src/native/port-manager.ts:56`
- Disconnect logging: `src/native/port-manager.ts:101-107`
- Reconnect attempts: `src/native/port-manager.ts:109-116`

### 3.4 Host and CLI both hardcode `/tmp/surf.sock` on Linux

Evidence:

- Host socket constant: `native/host.cjs:17-20`
- Host listener: `native/host.cjs:1637-1641`
- CLI socket constant: `native/cli.cjs:14-17`
- CLI error mapping (`ENOENT`, `ECONNREFUSED`): `native/cli.cjs:2933-2941`

### 3.5 Repository docs already label Linux support as experimental

Evidence:

- `README.md:648-651` states Linux support is experimental.

## 4. Reproduction and Runtime Evidence

Environment snapshot (2026-02-25):

- `chromium --version` => `Chromium 145.0.7632.109 snap`
- `snap list chromium` confirms Snap packaging.

Install behavior observed:

```bash
surf install jgbccjmbpfinlegbjkomjeoiklaijmmm --browser chromium
```

Reported install path:

- `~/.config/chromium/NativeMessagingHosts/surf.browser.host.json`

Failure symptoms observed:

```bash
surf tab.list
# Error: Connection refused. Native host not running.
```

Snap runtime env observed:

```bash
snap run --shell chromium -c 'echo HOME=$HOME; echo CHROME_CONFIG_HOME=$CHROME_CONFIG_HOME; echo SNAP_USER_COMMON=$SNAP_USER_COMMON; echo XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR'
```

Output includes:

- `HOME=/home/manuel/snap/chromium/3369`
- `CHROME_CONFIG_HOME=/home/manuel/snap/chromium/common`
- `SNAP_USER_COMMON=/home/manuel/snap/chromium/common`

Execution access failures under Snap context:

```bash
snap run --shell chromium -c '/home/manuel/.local/share/surf-cli/host-wrapper.sh ...'
# EXIT:126
# Permission denied

snap run --shell chromium -c '/home/manuel/.nvm/.../node --version'
# EXIT:126
# Permission denied
```

Socket namespace mismatch evidence:

- Host shell: `/tmp/surf.sock` inode `8995165`
- Snap shell: `/tmp/surf.sock` inode `12069922`

Interpretation:

- Even when both sides use the same literal path string, they can observe different socket objects in Snap vs non-Snap runtime contexts.

## 5. External Research Findings

### 5.1 Chrome native messaging contract

Source: Chrome Extensions native messaging documentation.

Key points:

1. Browser launches native host as a separate process.
2. On Linux/macOS, host `path` must be absolute.
3. For per-user install, Linux manifest location is under browser-specific `NativeMessagingHosts` paths.
4. “Native host has exited” and “host not found” are standard failure modes when host startup/pathing fails.

Reference:

- https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging

### 5.2 Chromium config root may differ from `~/.config`

Source: Chromium `user_data_dir.md`.

Key points:

1. Chromium default Linux config root is `~/.config/chromium`.
2. The `~/.config` prefix can be overridden by `$CHROME_CONFIG_HOME` (and `$XDG_CONFIG_HOME`).

Reference:

- https://chromium.googlesource.com/chromium/src/+/HEAD/docs/user_data_dir.md

### 5.3 Snap strict confinement and interfaces

Sources: Snap documentation (confinement + home interface).

Key points:

1. Strict confinement isolates access and requires explicit interfaces for resources.
2. `home` interface grants non-hidden home files by default (hidden-dot paths are special/restricted).

References:

- https://snapcraft.io/docs/explanation/security/snap-confinement/
- https://snapcraft.io/docs/home-interface

### 5.4 Known history: Chromium Snap native-messaging friction

Source: Launchpad bug 1741074.

Relevant historical notes include:

1. native host placement caveats under Snap profile paths,
2. practical failures in multiple extension integrations,
3. design discussion around portal mediation for confined browsers.

Reference:

- https://bugs.launchpad.net/bugs/1741074

### 5.5 Portal-based model (Firefox documented design)

Source: Firefox native messaging portal design doc.

Key points:

1. In strict confinement, browser cannot directly locate/launch host reliably.
2. WebExtensions XDG portal can mediate host discovery/launch outside sandbox.
3. Documentation explicitly frames this as confinement workaround; currently known in Firefox.

Reference:

- https://firefox-source-docs.mozilla.org/toolkit/components/extensions/webextensions/native-messaging-portal-design.html

## 6. Gap Analysis

Against desired outcome (“Surf works on Linux Chromium including Snap”), the current implementation has these gaps:

1. Manifest-location gap:
   - Installer writes only one Linux Chromium path based on `os.homedir()` and static suffix.
   - No handling of `CHROME_CONFIG_HOME`/Snap profile roots.

2. Executable-access gap:
   - Linux wrapper in hidden user path (`~/.local/...`) and host assets under hidden toolchain paths (`~/.nvm/...`) can be inaccessible in strict confinement contexts.

3. Socket-coordination gap:
   - Hardcoded `/tmp/surf.sock` assumes shared namespace between host (browser-launched) and CLI (user shell). This assumption breaks in Snap contexts.

4. Diagnostics gap:
   - User-facing errors indicate “socket not found/refused,” but do not guide Snap-specific remediation.

## 7. Proposed Solution

## 7.1 Design goals

1. Preserve existing non-Snap behavior.
2. Add explicit Snap-aware compatibility path.
3. Keep manual overrides possible via env vars.
4. Fail fast with actionable diagnostics when unsupported constraints remain.

### 7.2 API and config changes

Add a shared socket path override:

- `SURF_SOCKET_PATH` (read by both `native/host.cjs` and `native/cli.cjs`)

Optional (installer/runtime) additions:

- `SURF_SNAP_CHROMIUM=1` force Snap mode
- `SURF_SNAP_RUNTIME_DIR` custom runtime bundle root

### 7.3 Runtime resolution model

1. Resolve browser runtime kind (`snap` vs `regular`) for `--browser chromium` installs.
2. In Snap mode:
   - Install manifest into Snap Chromium-visible config root (`$SNAP_USER_COMMON/chromium/NativeMessagingHosts` equivalent host path under real home).
   - Place wrapper and required runtime (host + node + deps) in Snap-accessible directory under real-home snap data (`~/snap/chromium/common/surf-cli/...`).
3. Set/resolve a socket path visible from both contexts, e.g.:
   - `~/snap/chromium/common/surf-cli/surf.sock`
4. Use same path in CLI by default in Snap mode, or via `SURF_SOCKET_PATH`.

### 7.4 Pseudocode sketch

```js
// shared-socket.js
function resolveSocketPath() {
  if (process.env.SURF_SOCKET_PATH) return process.env.SURF_SOCKET_PATH;
  if (isWin()) return "//./pipe/surf";

  if (isLikelySnapChromiumRuntime()) {
    // shared regular FS path, not /tmp
    return path.join(realHome(), "snap/chromium/common/surf-cli/surf.sock");
  }

  return "/tmp/surf.sock";
}
```

```js
// installer chromium branch (linux)
if (browser === "chromium" && isSnapChromiumInstalled()) {
  const snapRoot = path.join(realHome(), "snap/chromium/common");
  const runtimeDir = path.join(snapRoot, "surf-cli");
  const manifestDir = path.join(snapRoot, "chromium/NativeMessagingHosts");

  copyHostRuntimeTo(runtimeDir); // host.cjs + deps + node wrapper
  writeWrapper(runtimeDir, { socketPath: path.join(runtimeDir, "surf.sock") });
  writeManifest(manifestDir, { path: path.join(runtimeDir, "host-wrapper.sh") });
}
```

### 7.5 Diagnostics improvements

When `ENOENT/ECONNREFUSED` occurs and Snap Chromium is detected:

1. Print effective socket path.
2. Print effective manifest path(s).
3. Suggest one-line debug command:

```bash
snap run --shell chromium -c '<wrapper> </dev/null >/tmp/surf.out 2>/tmp/surf.err; echo $?; cat /tmp/surf.err'
```

4. Suggest `SURF_SOCKET_PATH` override if mismatch is suspected.

## 8. Implementation Plan (Phased)

### Phase 1: Socket path configurability

Files:

1. `native/host.cjs`
2. `native/cli.cjs`

Changes:

1. Read `SURF_SOCKET_PATH` env override.
2. Keep existing defaults for non-Snap scenarios.
3. Add unit tests for path resolution utility (extract helper module for testability).

### Phase 2: Snap-aware installer/uninstaller

Files:

1. `scripts/install-native-host.cjs`
2. `scripts/uninstall-native-host.cjs`

Changes:

1. Detect Snap Chromium presence.
2. Install/remove manifests in Snap-aware location(s).
3. Build/remove Snap runtime wrapper directory.
4. Keep legacy Linux paths for non-Snap.

### Phase 3: Runtime bundling for Snap compatibility

Files:

1. `scripts/install-native-host.cjs`
2. `README.md`

Changes:

1. Copy required host runtime to Snap-accessible directory.
2. Validate runtime files post-install.
3. Document behavior and limitations.

### Phase 4: Diagnostics and docs

Files:

1. `native/cli.cjs`
2. `README.md`

Changes:

1. Improve error hints for Snap detection.
2. Add Snap troubleshooting section with explicit commands.

## 9. Testing and Validation Strategy

### 9.1 Unit tests

1. Socket path resolver tests:
   - default Linux path
   - `SURF_SOCKET_PATH` override
   - snap-detected path
2. Installer path resolution tests:
   - regular Linux Chromium
   - snap Chromium mode

### 9.2 Manual integration matrix

1. Chrome `.deb` + Surf global install.
2. Chromium Snap + Surf global install.
3. Chromium Snap + non-standard Node install.

Validation commands:

```bash
surf install <extension-id> --browser chromium
surf tab.list
surf screenshot --no-save
```

Snap-specific checks:

```bash
snap run --shell chromium -c 'env | grep -E "HOME|CHROME_CONFIG_HOME|SNAP_USER_COMMON"'
snap run --shell chromium -c '<resolved-wrapper> </dev/null >/tmp/out 2>/tmp/err; echo $?; cat /tmp/err'
```

## 10. Risks, Alternatives, Open Questions

### 10.1 Risks

1. Snap policies may vary across channels/releases and affect path assumptions.
2. Copying runtime artifacts into Snap-common directory can drift with Surf upgrades unless lifecycle is maintained.
3. Mixed-browser environments (Chrome + Snap Chromium) need deterministic behavior to avoid surprising overrides.

### 10.2 Alternatives

1. Declare Snap Chromium unsupported (lowest implementation cost, poor user outcome).
2. Maintain separate “snap helper” package (higher ops overhead).
3. Portal-based architecture (conceptually robust, currently no Chromium-equivalent integration path documented for Surf use).

### 10.3 Open questions

1. Should installer write manifests to both regular and Snap locations when Snap Chromium is installed?
2. Should socket default switch automatically in Snap mode or remain explicit via env var?
3. Is a dedicated `surf doctor` command warranted now or after Phase 2?

## 11. References

### 11.1 Repository evidence

1. `scripts/install-native-host.cjs`
2. `scripts/uninstall-native-host.cjs`
3. `native/host.cjs`
4. `native/cli.cjs`
5. `src/native/port-manager.ts`
6. `README.md`

### 11.2 External sources

1. Chrome native messaging: https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging
2. Chromium user data dir: https://chromium.googlesource.com/chromium/src/+/HEAD/docs/user_data_dir.md
3. Snap confinement: https://snapcraft.io/docs/explanation/security/snap-confinement/
4. Snap home interface: https://snapcraft.io/docs/home-interface
5. Launchpad Chromium Snap bug: https://bugs.launchpad.net/bugs/1741074
6. Firefox portal design for confined native messaging: https://firefox-source-docs.mozilla.org/toolkit/components/extensions/webextensions/native-messaging-portal-design.html
