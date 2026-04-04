# GoReleaser & Release Workflow

## Type
infra

## Status
todo

## Priority
high

## Project
06-production-readiness

## Description
Set up GoReleaser to produce cross-compiled binaries and a GitHub Release on every `vX.Y.Z` tag push. GoReleaser handles compilation, archiving, checksum generation, and GitHub Release creation automatically.

## Acceptance Criteria
- [ ] `.goreleaser.yaml` is committed to the repo root
- [ ] GoReleaser builds binaries for: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- [ ] Each binary is packaged as a `.tar.gz` archive with the binary + `README.md` + `LICENSE`
- [ ] A `checksums.txt` file is included in the release
- [ ] `.github/workflows/release.yml` triggers on `push` to tags matching `v*`
- [ ] The release workflow uses `GITHUB_TOKEN` (the default Actions token is sufficient for same-repo releases)
- [ ] Version is injected via GoReleaser ldflags (replaces / complements the Makefile approach)
- [ ] Pushing tag `v0.1.0` produces a GitHub Release named `v0.1.0` with all artifacts attached
- [ ] `goreleaser check` passes (validates config locally with `goreleaser` CLI)

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Install `goreleaser` locally for config validation (`brew install goreleaser`)
- [ ] Write `.goreleaser.yaml` with builds, archives, checksum, and GitHub release sections
- [ ] Configure ldflags in GoReleaser to inject version (same variable as Makefile)
- [ ] Create `.github/workflows/release.yml` with GoReleaser action
- [ ] Store `GITHUB_TOKEN` — default `secrets.GITHUB_TOKEN` is sufficient, confirm permissions
- [ ] Tag `v0.1.0` and verify the release end-to-end
- [ ] Update `Makefile` `build` target to note that production releases use GoReleaser

## Blocked By
12-version-command
