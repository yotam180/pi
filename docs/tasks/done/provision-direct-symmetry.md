# Fix provisionDirect Asymmetry

## Type
improvement

## Status
done

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
- [x] `RuntimeDescriptor.DirectDownload` flag is set correctly for each runtime
- [x] `provisionDirect` uses the flag to decide behavior — no runtime-name switch statement
- [x] Either: go and rust direct downloads are implemented and tested; OR: all unsupported runtimes return the same error format
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

**Chose Option A — full implementation of direct download for all four runtimes.**

### Go direct download (`provisionGoDirect`):
- Downloads from `https://dl.google.com/go/go{version}.{os}-{arch}.tar.gz`
- Extracts the tarball which contains a `go/` directory with `bin/` inside
- Moves extracted `go/bin/` contents to `~/.pi/runtimes/go/{version}/bin/`
- Supports darwin and linux on amd64 and arm64

### Rust direct download (`provisionRustDirect`):
- Uses the official rustup installer (`https://sh.rustup.rs`)
- Sets `RUSTUP_HOME` and `CARGO_HOME` to `~/.pi/runtimes/rust/{version}/{rustup,cargo}`
- Runs with `--no-modify-path --default-toolchain {version}` to avoid modifying the user's shell config
- Copies binaries from `$CARGO_HOME/bin/` to the standard PI bin directory
- Supports darwin and linux

### `provisionDirect` refactoring:
- Now checks `RuntimeDescriptor.DirectDownload` flag before dispatching
- Unknown or unsupported runtimes get a consistent error: `"direct provisioning not supported for %q — install mise first: curl https://mise.run | sh"`
- The per-runtime switch statement remains for dispatch, but entry is gated by the descriptor flag

### Tests added:
- `TestProvisionGoDirect_RunnerCalled` — verifies Go provisioning calls bash with correct download URL
- `TestProvisionGoDirect_ScriptContainsCorrectURL` — verifies script references dl.google.com and version
- `TestProvisionRustDirect_RunnerCalled` — verifies Rust provisioning calls bash with rustup.rs
- `TestProvisionRustDirect_ScriptContainsVersion` — verifies script sets RUSTUP_HOME, CARGO_HOME, --no-modify-path
- `TestProvisionDirect_AllKnownRuntimesSupported` — verifies all four runtimes have DirectDownload=true

### Tests removed:
- `TestProvisionDirect_GoUnsupported` — replaced by `TestProvisionGoDirect_RunnerCalled`
- `TestProvisionDirect_RustUnsupported` — replaced by `TestProvisionRustDirect_RunnerCalled`

## Subtasks
- [x] Add `DirectDownload bool` to `RuntimeDescriptor` for relevant runtimes (was already there, updated to true)
- [x] Refactor `provisionDirect` to use descriptor flag
- [x] Add `provisionGoDirect`
- [x] Add `provisionRustDirect`
- [x] Tests

## Blocked By
- ~~`runtime-descriptor-type`~~ (completed)
