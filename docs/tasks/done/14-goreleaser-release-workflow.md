# GoReleaser & Release Workflow

## Type
infra

## Status
done

## Priority
high

## Project
06-production-readiness

## Description
Set up GoReleaser to produce cross-compiled binaries and a GitHub Release on every `vX.Y.Z` tag push. GoReleaser handles compilation, archiving, checksum generation, and GitHub Release creation automatically.

## Acceptance Criteria
- [x] `.goreleaser.yaml` is committed to the repo root
- [x] GoReleaser builds binaries for: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- [x] Each binary is packaged as a `.tar.gz` archive with the binary + `README.md` (no LICENSE file exists yet)
- [x] A `checksums.txt` file is included in the release
- [x] `.github/workflows/release.yml` triggers on `push` to tags matching `v*`
- [x] The release workflow uses `GITHUB_TOKEN` (the default Actions token is sufficient for same-repo releases)
- [x] Version is injected via GoReleaser ldflags (same variable as Makefile: `cli.version`)
- [x] Pushing tag `v0.1.0` will produce a GitHub Release named `v0.1.0` with all artifacts attached
- [x] `goreleaser check` passes (validates config locally)

## Implementation Notes

### GoReleaser config (`.goreleaser.yaml`)
- Uses GoReleaser v2 config format (`version: 2`)
- Builds 4 targets: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- CGO disabled for clean cross-compilation
- Binaries stripped with `-s -w` ldflags for smaller size
- Version injected via `-X github.com/vyper-tooling/pi/internal/cli.version={{ .Version }}`
- Archives use `tar.gz` format (using `formats:` not deprecated `format:`)
- Build IDs use `ids:` not deprecated `builds:`
- Changelog auto-generated, excludes docs/test/chore commits
- Release targets `yotam180/pi` GitHub repo

### Release workflow (`.github/workflows/release.yml`)
- Triggers on `v*` tag push
- Uses `goreleaser/goreleaser-action@v6` with `version: "~> v2"`
- Runs tests before release as safety check
- `fetch-depth: 0` for full git history (needed for changelog)
- Uses default `GITHUB_TOKEN` — no extra secrets needed

### Other changes
- Added `dist/` to `.gitignore`
- Added `snapshot` target to Makefile for local testing
- Added comment to `build` target noting production releases use GoReleaser
- Added `dist/` to `clean` target

### Verification
- `goreleaser check` passes clean
- `goreleaser build --snapshot --clean` succeeds for all 4 targets
- `goreleaser release --snapshot --clean` produces archives and checksums
- Version injection confirmed: built binary reports `0.0.0-SNAPSHOT-<hash>`
- Archive contents verified: `pi` binary + `README.md`
- All 205 existing tests pass

## Subtasks
- [x] Install `goreleaser` locally for config validation (`brew install goreleaser`)
- [x] Write `.goreleaser.yaml` with builds, archives, checksum, and GitHub release sections
- [x] Configure ldflags in GoReleaser to inject version (same variable as Makefile)
- [x] Create `.github/workflows/release.yml` with GoReleaser action
- [x] `GITHUB_TOKEN` — default `secrets.GITHUB_TOKEN` is sufficient, no extra config needed
- [ ] Tag `v0.1.0` and verify the release end-to-end (deferred — requires push to remote)
- [x] Update `Makefile` `build` target to note that production releases use GoReleaser

## Blocked By
12-version-command
