# Implement PI Features from bat Gap Analysis

## Type
feature

## Status
done

## Priority
medium

## Project
08-bat-adoption-test

## Description
Based on the findings from task 37 (clone-and-examine-bat), implement any missing PI features or built-in automations needed to support bat's developer workflows.

Expected gaps may include:
- `pi:install-rust` built-in for Rust/Cargo setup
- Other Rust-specific tooling automations
- Any new step types or execution features

If task 37 finds no feature gaps, this task should be marked done with a note.

## Acceptance Criteria
- [x] All feature gaps from task 37 have corresponding tasks created
- [x] All created tasks are completed or documented as out-of-scope
- [x] `go test ./...` passes after all changes
- [x] Documentation updated for any new features

## Implementation Notes

### Gap Analysis Results (from task 37)

Only one real feature gap was identified:

| Gap | Severity | Resolution |
|-----|----------|------------|
| No `pi:install-rust` built-in | Medium | **Implemented** — new `install-rust.yaml` built-in |

All other workflows (build, test, lint, format, audit, etc.) are already fully supported by PI's existing bash step and `env:` features.

### What was implemented

#### `pi:install-rust` built-in automation

**File**: `internal/builtins/embed_pi/install-rust.yaml`

Follows the existing installer pattern but uses `rustup` (the standard Rust installer) instead of mise/brew:

- **test**: Checks if `rustc --version` outputs a version matching the requested major.minor
- **run**: Uses `rustup install` + `rustup default` if rustup is already available; otherwise downloads and runs the official `rustup.rs` installer script
- **version**: Extracts version from `rustc --version` output

Key design decision: Unlike `install-python`/`install-node`/`install-go` which use mise → brew fallback, Rust has its own standard installer (`rustup`) that is universally accepted. Using rustup is the right choice because:
1. It's the official, recommended way to install Rust
2. It handles toolchain management (multiple versions, components like clippy/rustfmt)
3. Almost every Rust project assumes rustup is available
4. mise also uses rustup under the hood for Rust

#### Install hints

Added install hints for `rustc`, `cargo`, and `rustup` — all pointing to the official `rustup.rs` installer.

### Tests added

- `TestDiscover_InstallRustAcceptsVersionInput` — validates version input spec, required flag, and PI_INPUT_VERSION reference
- `TestDiscover_InstallRustUsesRustup` — validates run phase references rustup
- Integration tests updated: `install-rust` added to pi list marker test, pi info details test, and pi info inputs test

### QA

- `go build ./...` passes
- `go test ./...` passes (all 550+ tests)
- Manual QA: `pi info pi:install-rust` shows correct name, description, lifecycle, and inputs
- Manual QA: `pi list` shows `install-rust` with `[built-in]` marker and `version` input
- Manual QA: `pi run pi:install-rust --with version=1.94` correctly detects already-installed Rust 1.94.1

## Subtasks
- [x] Implement `pi:install-rust` built-in
- [x] Add unit tests
- [x] Add integration tests
- [x] Add install hints for rustc/cargo/rustup
- [x] Update documentation
- [x] QA verification

## Blocked By
37-clone-and-examine-bat
