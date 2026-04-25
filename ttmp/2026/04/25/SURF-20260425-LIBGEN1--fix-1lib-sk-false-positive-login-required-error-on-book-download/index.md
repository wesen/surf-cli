---
Title: Fix 1lib.sk false-positive login required error on book download
Ticket: SURF-20260425-LIBGEN1
Status: active
Topics:
    - browser-automation
    - cli
    - surf-go
    - libgen
    - 1lib
    - bugfix
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go/internal/cli/commands/libgen_download.go
      Note: Contains buggy buildLibgenDownloadCheckCode() that needs visibility check
ExternalSources: []
Summary: "Fixed false-positive 'login required' error when downloading books from 1lib.sk. The login check now verifies element visibility before treating a[data-mode='singlelogin'] as a login-required condition."
LastUpdated: 2026-04-25T08:25:40.172899842-04:00
WhatFor: ""
WhenToUse: ""
---


# Fix 1lib.sk false-positive login required error on book download

## Overview

When downloading a book from 1lib.sk via `surf-go libgen download --save-to`, the CLI falsely reported "login required" even when the user was properly logged in and the browser would have successfully downloaded the file. The root cause was `buildLibgenDownloadCheckCode()` in `libgen_download.go` using `document.querySelector('a[data-mode="singlelogin"]')` without checking whether the matched element was actually visible. When logged in, 1lib.sk keeps hidden login links in the DOM (e.g., `display: none`), which the code incorrectly treated as active login requirements.

## Fix

Added an `isVisible()` helper inside the JavaScript check that verifies `display`, `visibility`, `opacity`, and element dimensions before flagging `login_required`.

## Documents

- [Bug Analysis](./analysis/01-bug-analysis-false-positive-login-required-on-1lib-sk-download.md)
- [Implementation Guide](./implementation-guide/01-implementation-guide-visibility-aware-login-check.md)
- [Investigation Diary](./reference/01-investigation-diary.md)

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- browser-automation
- cli
- surf-go
- libgen
- 1lib
- bugfix

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
