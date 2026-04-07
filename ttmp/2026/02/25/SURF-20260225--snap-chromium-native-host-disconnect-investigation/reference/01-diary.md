---
Title: Diary
Ticket: SURF-20260225
Status: active
Topics:
    - linux
    - chromium
    - snap
    - native-messaging
    - debugging
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: native/cli.cjs
      Note: CLI client socket path must match host path across namespaces
    - Path: native/host.cjs
      Note: Native host socket path and lifecycle are central to Snap namespace failures
    - Path: scripts/install-native-host.cjs
      Note: Installer writes Linux Chromium native messaging manifests and wrapper paths
    - Path: scripts/uninstall-native-host.cjs
      Note: Uninstaller paths need to mirror any Snap-aware install behavior
    - Path: src/native/port-manager.ts
      Note: Extension connect/disconnect logs surface native host exit symptoms
ExternalSources: []
Summary: Detailed debugging log for Surf CLI disconnects on Snap Chromium, including local reproduction, confinement evidence, and source-backed handling options.
LastUpdated: 2026-02-25T15:58:42.434350155-05:00
WhatFor: Chronological investigation diary for Chromium Snap native host failures
WhenToUse: Use when debugging native messaging disconnects on Linux Chromium installs, especially Snap-based Chromium
---


# Diary

## Goal

Capture a complete, command-level investigation of why Surf fails to stay connected on Linux Chromium (Snap), including reproducible local evidence, failed hypotheses, and best-practice handling options validated against primary sources.

## Step 1: Reproduce and map Surf’s Linux Chromium integration

I started by confirming the exact installation and connection paths used by Surf, then checking how those compare to the Chromium package actually installed on this machine. The goal was to avoid guessing and verify path assumptions from code and runtime state.

I focused first on installer behavior and native messaging manifest destinations, because the symptom looked like either Chromium not finding the host, or finding it but failing to run it.

### Prompt Context

**User prompt (verbatim):** "I have trouble running this on linux with chromium, I don't think the chromium is able to connect to the local launcher even though it seems installed properly. Help me debug."

**Assistant interpretation:** Investigate Surf CLI + Chromium integration on Linux and identify why native host connection fails.

**Inferred user intent:** Obtain a concrete root cause and practical fix (or patch) for Chromium/Surf connectivity.

**Commit (code):** N/A (no repo code changed yet)

### What I did

- Inspected Surf native host installer, uninstaller, CLI socket logic, and native messaging bridge:
  - `scripts/install-native-host.cjs`
  - `scripts/uninstall-native-host.cjs`
  - `native/cli.cjs`
  - `native/host.cjs`
  - `src/native/port-manager.ts`
- Verified local browser/runtime state:
  - `chromium --version` => `Chromium 145.0.7632.109 snap`
  - `snap list chromium` confirmed Snap package.
- Ran `surf install jgbccjmbpfinlegbjkomjeoiklaijmmm --browser chromium` and captured output.
- Verified installed manifest path and content:
  - `~/.config/chromium/NativeMessagingHosts/surf.browser.host.json`

### Why

- The native messaging host chain is path-sensitive and packaging-sensitive.
- Snap Chromium does not always use the same host lookup paths as non-Snap Chromium.

### What worked

- Reproduction of the CLI symptom:
  - `surf tab.list` => `Error: Socket not found. Is Chrome running with the extension?`
- Confirmed Surf installer writes only to:
  - `~/.config/chromium/NativeMessagingHosts`

### What didn't work

- Initial assumption that default Linux Chromium path would be sufficient.
- No connection recovered after standard install + restart.

### What I learned

- This machine uses Snap Chromium, and Snap-specific path/runtime behavior must be handled explicitly.

### What was tricky to build

- The tricky part was distinguishing three similar failure classes:
  1) manifest not found,
  2) host executable not runnable,
  3) host runnable but socket unreachable.
- Symptom text (`Native host has exited`) can appear for more than one underlying cause.

### What warrants a second pair of eyes

- Whether Surf should support Snap Chromium as first-class behavior or declare unsupported with explicit diagnostics.

### What should be done in the future

- Add explicit Snap detection and tailored installation/runtime logic in Surf installer + host/CLI socket handling.

### Code review instructions

- Start in:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/scripts/install-native-host.cjs`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/host.cjs`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/cli.cjs`
- Validate with:
  - `chromium --version`
  - `surf install <extension-id> --browser chromium`
  - `surf tab.list`

### Technical details

- Key observed installer output:
  - `Installed for: Chromium: /home/manuel/.config/chromium/NativeMessagingHosts/surf.browser.host.json`
- This alone does not guarantee Snap Chromium can execute/read the target host path.

## Step 2: Verify whether wrapper/host is ever executed

I then tested the wrapper directly outside browser control and under Snap confinement. This isolated native messaging lifecycle issues from extension logic and proved whether host startup itself was viable.

I used short-lived execution with stdin closed to simulate browser disconnect behavior and checked binary stdout framing plus host logs.

### Prompt Context

**User prompt (verbatim):** "I get [Surf] Service worker loaded
debug.ts:10 [Surf] Connecting to native host...
debug.ts:28 [Surf] Native host disconnected: Native host has exited.
debug.ts:28 [Surf] Extension installed/updated: update 

now. 

maybe we can test if the wrapper gets executed at all?"

**Assistant interpretation:** Confirm host wrapper execution path and determine if disconnection occurs before/after host startup.

**Inferred user intent:** Rule in/out wrapper launch failure quickly.

**Commit (code):** N/A (no repo code changed yet)

### What I did

- Inspected wrapper and manifests:
  - `/home/manuel/.local/share/surf-cli/host-wrapper.sh`
  - `~/.config/chromium/NativeMessagingHosts/surf.browser.host.json`
  - `~/snap/chromium/common/chromium/NativeMessagingHosts/surf.browser.host.json`
- Directly executed wrapper outside Snap:
  - `timeout 3s /home/manuel/.local/share/surf-cli/host-wrapper.sh`
- Captured outputs/logs:
  - stdout bytes: `b'\x15\x00\x00\x00{"type":"HOST_READY"}'`
  - `/tmp/surf-host.log` showed host start + socket listen + stdin end.

### Why

- If wrapper cannot produce `HOST_READY`, browser disconnect is expected.
- If wrapper can produce `HOST_READY`, failure is likely path/confinement/socket boundary related.

### What worked

- Outside Snap, wrapper executed correctly and host reached `HOST_READY`.

### What didn't work

- Same wrapper path under Snap shell failed:
  - `/home/manuel/.local/share/surf-cli/host-wrapper.sh: Permission denied`
  - Exit code `126`.

### What I learned

- Snap confinement blocks execution from Surf’s default wrapper location (`~/.local/share/surf-cli`) in this setup.

### What was tricky to build

- It was easy to misread this as a Node path issue only, but the first hard failure was wrapper execution permission inside Snap context.

### What warrants a second pair of eyes

- Confirm whether confinement behavior varies across distro/snap revisions and whether there is an officially supported host executable location policy for Chromium Snap.

### What should be done in the future

- Installer should detect Snap Chromium and avoid writing executable wrapper into blocked locations.

### Code review instructions

- Validate wrapper behavior in both contexts:
  - Host shell: `/home/manuel/.local/share/surf-cli/host-wrapper.sh </dev/null`
  - Snap shell: `snap run --shell chromium -c '/home/manuel/.local/share/surf-cli/host-wrapper.sh </dev/null'`

### Technical details

- Critical failure evidence:
  - `EXIT:126`
  - `/bin/bash: line 3: /home/manuel/.local/share/surf-cli/host-wrapper.sh: Permission denied`

## Step 3: Prove Snap confinement + namespace impacts beyond wrapper execution

After confirming wrapper path execution failure, I tested whether a Snap-accessible wrapper + Node runtime could run, and then checked whether the host socket was visible to the external Surf CLI process.

This step exposed a second independent blocker: socket path namespace mismatch (`/tmp` inside snap vs host namespace).

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue root-cause analysis after wrapper tests, looking for additional disconnect causes.

**Inferred user intent:** Identify all blockers, not just the first one.

**Commit (code):** N/A (no repo code changed yet)

### What I did

- Verified Snap shell environment:
  - `HOME=/home/manuel/snap/chromium/3369`
  - `CHROME_CONFIG_HOME=/home/manuel/snap/chromium/common`
  - `XDG_RUNTIME_DIR=/run/user/1000/snap.chromium`
- Confirmed Node execution constraints:
  - `/home/manuel/.nvm/.../node --version` => `Permission denied` (exit `126`).
  - `/usr/bin/node` unavailable inside Snap shell.
- Built a Snap-accessible runtime bundle under:
  - `/home/manuel/snap/chromium/common/surf-cli`
  - copied node binary + surf-cli package + wrapper.
- Verified copied runtime can execute in Snap shell.
- Compared socket/log inode identities across namespaces:
  - Host namespace `/tmp/surf.sock` inode differed from Snap namespace `/tmp/surf.sock` inode.

### Why

- Even if host starts, Surf CLI cannot connect unless both sides point to the same filesystem namespace path.
- Hardcoded `/tmp/surf.sock` is unsafe when host runs inside Snap namespace.

### What worked

- Snap-local wrapper and copied node could execute successfully in Snap shell (`EXIT:0`, binary HOST_READY output).

### What didn't work

- Host still not reachable from external Surf CLI due separate `/tmp` namespace.
- Attempt to create files under `/run/user/1000` from Snap shell failed with permission denied in this environment.

### What I learned

- There are at least two blockers for Snap Chromium:
  1) executable/location access restrictions,
  2) socket path namespace mismatch.

### What was tricky to build

- Root cause required iterative narrowing because each fix exposed a deeper layer:
  - Fix wrapper path => hit host source access (`EACCES` in `.nvm`) => fix runtime location => hit socket namespace split.

### What warrants a second pair of eyes

- Best cross-namespace socket strategy for Snap Chromium that is reliable across distributions and Snap policy changes.

### What should be done in the future

- Add configurable socket path (env var) to both CLI and host, plus Snap-aware defaults.
- Add explicit diagnostics when Chromium is Snap and host/socket are in incompatible locations.

### Code review instructions

- Inspect socket constants and hardcoded paths:
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/host.cjs`
  - `/home/manuel/code/others/llms/pi/nicobailon/surf-cli/native/cli.cjs`
- Reproduce inode mismatch check:
  - host shell `stat /tmp/surf.sock`
  - snap shell `snap run --shell chromium -c 'stat /tmp/surf.sock'`

### Technical details

- Representative hard failure while trying to run host from `.nvm` inside Snap:

```text
Error: EACCES: permission denied, open '/home/manuel/.nvm/versions/node/v22.21.0/lib/node_modules/surf-cli/native/host.cjs'
```

## Step 4: Web research and recommended handling options

I validated local findings against primary documentation/sources for Chrome native messaging behavior, Chromium user-data path rules, and Snap confinement caveats. I focused on source-backed guidance that can be translated into installer/runtime logic.

The evidence supports that Snap Chromium requires different host placement conventions and that strict confinement can block ordinary host executable paths.

### Prompt Context

**User prompt (verbatim):** "maybe search the web to see how to best handle that as well?

Keep a detailed diary of your investigations in a new ticket $docmgr $diary"

**Assistant interpretation:** Research best-practice handling online and capture investigation in docmgr diary format.

**Inferred user intent:** Move from raw debugging to an actionable, documented implementation direction.

**Commit (code):** N/A (no repo code changed yet)

### What I did

- Researched and read source material:
  - Chrome native messaging docs (host registration and Linux locations).
  - Chromium user-data/config path override docs (`CHROME_CONFIG_HOME`).
  - Launchpad/Chromium Snap bug discussion with native messaging caveats.
  - Snap confinement docs for strict confinement behavior.
- Consolidated recommended handling options and mapped them to Surf code areas.

### Why

- Needed authoritative references before proposing installer/runtime behavior changes for Snap.

### What worked

- Sources aligned with local reproduction: Snap-specific path/runtime behavior is materially different and cannot be treated like regular Linux Chromium.

### What didn't work

- No single authoritative “one size fits all” recipe was found that solves both executable confinement and cross-namespace IPC without Snap-aware app logic.

### What I learned

- Best practical direction for Surf:
  - Implement explicit Snap detection.
  - Use Snap-appropriate manifest/host placement.
  - Introduce explicit host/CLI socket path configurability to avoid `/tmp` namespace assumptions.
  - If Snap support remains partial, fail fast with clear guidance.

### What was tricky to build

- The web sources provide pieces of behavior but not Surf-specific architecture guidance; integration decisions require combining docs with local experiments.

### What warrants a second pair of eyes

- Final selection of canonical socket location and installer strategy for Snap Chromium across both deb/snap mixed setups.

### What should be done in the future

- Create a dedicated design doc in this ticket for implementation strategy and tests.
- Add an automated self-check command (or install-time doctor) for Snap constraints.

### Code review instructions

- Review assumptions in:
  - installer Linux Chromium paths
  - wrapper/runtime placement
  - socket path constants
- Validate against both:
  - non-Snap Chrome/Chromium
  - Snap Chromium

### Technical details

- Sources:
  - Chrome Native Messaging: https://developer.chrome.com/docs/apps/nativeMessaging
  - Chromium profile/config env overrides (`CHROME_CONFIG_HOME`): https://chromium.googlesource.com/chromium/src/+/HEAD/docs/user_data_dir.md
  - Chromium Snap native messaging caveat discussion: https://bugs.launchpad.net/ubuntu/+source/chromium-browser/+bug/1741074
  - Snap strict confinement overview: https://snapcraft.io/docs/snap-confinement

## Related

- Ticket index: `ttmp/2026/02/25/SURF-20260225--snap-chromium-native-host-disconnect-investigation/index.md`
- Ticket changelog: `ttmp/2026/02/25/SURF-20260225--snap-chromium-native-host-disconnect-investigation/changelog.md`
