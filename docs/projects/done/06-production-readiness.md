# Production Readiness

## Status
done

## Priority
high

## Description
Set up the CI/CD pipeline, release automation, and Homebrew distribution for `pi`. Every push is tested automatically; every git tag produces signed, cross-compiled binaries attached to a GitHub release; and users can install with `brew install yotam180/pi/pi`.

## Goals
- `pi --version` reports the exact git tag or commit it was built from
- Every push and PR runs the full test suite on Linux and macOS via GitHub Actions
- Pushing a `vX.Y.Z` tag triggers GoReleaser: builds binaries for all target platforms, creates a GitHub Release with checksums, and updates the Homebrew cask automatically
- Users can install and upgrade `pi` with a single `brew` command (once repo is public)
- The release process is fully automated — no manual steps after tagging

## Background & Context
The repo is at `github.com/yotam180/pi`. The Makefile already injects version via `git describe --tags` ldflags. GitHub CLI access is configured as `yotam180`.

GoReleaser is the standard tool for Go release automation — it handles cross-compilation, archive packaging, checksum generation, GitHub Release creation, and Homebrew cask updates in a single config file. It's free for open source.

## Scope

### In scope
- `pi --version` (or `pi version`) command that prints the injected version string
- GitHub Actions CI workflow: `go test ./... -race` + `go vet ./...` on push and PRs, matrix over ubuntu-latest and macos-latest
- GoReleaser config (`.goreleaser.yaml`) with cross-compilation targets:
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- GitHub Actions release workflow: triggered on `v*` tag push, runs GoReleaser
- `homebrew-pi` tap repo at `github.com/yotam180/homebrew-pi`
- GoReleaser Homebrew tap integration: auto-commits updated cask on release

### Out of scope
- Code signing / notarization (macOS Gatekeeper — deferred)
- Windows builds and Scoop/Winget packaging
- Docker image publishing
- Snap / APT / RPM packaging
- Automated version bumping or changelog generation

## Success Criteria
- [x] `pi --version` prints a version string (e.g. `pi v0.1.0` or `pi dev-abc1234`)
- [x] Pushing to `main` or opening a PR triggers CI and runs all tests on Linux and macOS
- [x] Pushing tag `v0.1.0` produces a GitHub Release with 4 binary archives + checksums
- [x] GoReleaser auto-publishes Homebrew cask to `yotam180/homebrew-pi` on release
- [ ] `brew install yotam180/pi/pi` installs a working `pi` binary — blocked by repo being private (infrastructure is complete)
- [ ] `brew upgrade pi` picks up the next release automatically — blocked by repo being private
- [x] No manual steps required after `git tag vX.Y.Z && git push --tags`

## Notes
- Uses GoReleaser v2 with `homebrew_casks:` (the modern replacement for deprecated `brews:`)
- The `homebrew-pi` repo is public; the `pi` repo is currently private
- `HOMEBREW_TAP_TOKEN` secret is configured in the `pi` repo
- Release workflow includes Node.js + tsx setup to run full test suite
- Cask includes postflight hook to remove quarantine bit on macOS (since binaries aren't code-signed)
- `skip_upload: auto` prevents pre-release tags from updating the tap
- First release: v0.1.0
