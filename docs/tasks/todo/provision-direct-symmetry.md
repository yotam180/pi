# Fix provisionDirect Asymmetry

## Type
improvement

## Status
todo

## Priority
medium

## Project
15-runtime-provider-registry

## Description

`provisionDirect` silently supports node and python while returning a "not supported — install mise" error for go and rust. All four are full members of `KnownRuntimes`, so users expect symmetric behavior. This task makes the asymmetry explicit and consistent, ideally by implementing direct download for go and rust, or at minimum by making the fallback behavior deterministic and clearly documented.

**Current behavior:**
- `node` → downloads from nodejs.org tarball ✓
- `python` → downloads from astral-sh/python-build-standalone ✓
- `go` → returns error: "direct provisioning for go is not supported — install mise first" ✗
- `rust` → returns error: "direct provisioning for rust is not supported — install mise first" ✗

**Expected behavior (option A — implement direct download):**
- `go` → download from `https://go.dev/dl/` and extract to install dir
- `rust` → run `curl https://sh.rustup.rs | sh` with `RUSTUP_HOME`/`CARGO_HOME` pointed at the PI install dir

**Expected behavior (option B — symmetric "needs mise" behavior):**
- All four runtimes have a clear `DirectDownload: bool` flag in `RuntimeDescriptor`
- `provisionDirect` checks this flag and returns a consistent error for any runtime that doesn't support direct download
- The error message is the same format for all unsupported runtimes

Option A is better for users. Option B is a minimum viable fix. Use `RuntimeDescriptor.DirectDownload` (from the descriptor task) as the gate.

## Acceptance Criteria
- [ ] `RuntimeDescriptor.DirectDownload` flag is set correctly for each runtime
- [ ] `provisionDirect` uses the flag to decide behavior — no runtime-name switch statement
- [ ] Either: go and rust direct downloads are implemented and tested; OR: all unsupported runtimes return the same error format
- [ ] `go build ./...` and `go test ./...` pass

## Implementation Notes

For Go direct download:
- Download URL: `https://dl.google.com/go/go{version}.{os}-{arch}.tar.gz`
- Extract into `~/.pi/runtimes/go/{version}/`
- The `bin/` dir is inside the tarball root

For Rust direct download:
- `rustup` is the standard installer — it manages versions, but installs to `$RUSTUP_HOME`
- Set `RUSTUP_HOME=$HOME/.pi/runtimes/rust/{version}` and `CARGO_HOME=$HOME/.pi/runtimes/rust/{version}`
- Run: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | RUSTUP_HOME=... CARGO_HOME=... sh -s -- -y --no-modify-path --default-toolchain {version}`

## Subtasks
- [ ] Add `DirectDownload bool` to `RuntimeDescriptor` for relevant runtimes
- [ ] Refactor `provisionDirect` to use descriptor flag
- [ ] (If implementing) Add `provisionGoDirect`
- [ ] (If implementing) Add `provisionRustDirect`
- [ ] Tests

## Blocked By
- ~~`runtime-descriptor-type`~~ (completed)
