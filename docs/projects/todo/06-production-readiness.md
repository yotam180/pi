# Production Readiness

## Status
todo

## Priority
high

## Description
Set up the CI/CD pipeline, release automation, and Homebrew distribution for `pi`. Every push is tested automatically; every git tag produces signed, cross-compiled binaries attached to a GitHub release; and users can install with `brew install yotam180/pi/pi`.

## Goals
- `pi --version` reports the exact git tag or commit it was built from
- Every push and PR runs the full test suite on Linux and macOS via GitHub Actions
- Pushing a `vX.Y.Z` tag triggers GoReleaser: builds binaries for all target platforms, creates a GitHub Release with checksums, and updates the Homebrew formula automatically
- Users can install and upgrade `pi` with a single `brew` command
- The release process is fully automated — no manual steps after tagging

## Background & Context
The repo is at `github.com/yotam180/pi`. The Makefile already injects version via `git describe --tags` ldflags. GitHub CLI access is configured as `yotam180`.

GoReleaser is the standard tool for Go release automation — it handles cross-compilation, archive packaging, checksum generation, GitHub Release creation, and Homebrew formula updates in a single config file. It's free for open source.

A Homebrew formula requires:
1. A GitHub Release with a source tarball (or pre-built binaries)
2. A `homebrew-pi` tap repo at `github.com/yotam180/homebrew-pi`
3. A formula file that GoReleaser can auto-update on each release

## Scope

### In scope
- `pi --version` (or `pi version`) command that prints the injected version string
- GitHub Actions CI workflow: `go test ./... -race` + `go vet ./...` on push and PRs, matrix over ubuntu-latest and macos-latest
- GoReleaser config (`.goreleaser.yaml`) with cross-compilation targets:
  - `darwin/amd64`, `darwin/arm64`
  - `linux/amd64`, `linux/arm64`
- GitHub Actions release workflow: triggered on `v*` tag push, runs GoReleaser
- `homebrew-pi` tap repo at `github.com/yotam180/homebrew-pi`
- GoReleaser Homebrew tap integration: auto-commits updated formula on release

### Out of scope
- Code signing / notarization (macOS Gatekeeper — deferred)
- Windows builds and Scoop/Winget packaging
- Docker image publishing
- Snap / APT / RPM packaging
- Automated version bumping or changelog generation

## Success Criteria
- [ ] `pi --version` prints a version string (e.g. `pi v0.1.0` or `pi dev-abc1234`)
- [ ] Pushing to `main` or opening a PR triggers CI and runs all tests on Linux and macOS
- [ ] Pushing tag `v0.1.0` produces a GitHub Release with 4 binary archives + checksums
- [ ] `brew install yotam180/pi/pi` installs a working `pi` binary on macOS (both Intel and Apple Silicon)
- [ ] `brew upgrade pi` picks up the next release automatically
- [ ] No manual steps required after `git tag vX.Y.Z && git push --tags`

## Notes
- Use GoReleaser v2 (current major version as of 2026).
- The `homebrew-pi` repo must exist before GoReleaser can push to it. Create it as a public repo with a `README.md` before the first release.
- GoReleaser needs a `GITHUB_TOKEN` with write access to both `pi` and `homebrew-pi` repos. Use a fine-grained PAT stored as a repository secret.
- The ldflags in the Makefile (`-X github.com/vyper-tooling/pi/internal/cli.version=$(VERSION)`) already wire version injection — the `pi --version` task just needs to expose that variable via the CLI.
