# Changelog

## 2026-04-25

- Initial workspace created


## 2026-04-25

Fixed false-positive 'login required' error in 1lib.sk book download. The login check in buildLibgenDownloadCheckCode() now verifies element visibility (display, visibility, opacity, offset dimensions) before reporting login_required. Previously, hidden a[data-mode='singlelogin'] elements in the DOM would trigger the error even when the user was properly logged in.

### Related Files

- /home/manuel/code/others/llms/pi/nicobailon/surf-cli/go/internal/cli/commands/libgen_download.go — Added isVisible helper and visibility check to login link detection

