# Refactor Step Runners & Deduplicate Validation Logic

## Type
improvement

## Status
in_progress

## Priority
high

## Project
standalone

## Description

The executor package has two areas of significant code duplication that hurt maintainability and make the codebase harder to extend:

### 1. Step Runner Duplication (`runners.go`, `install.go`)
`execBash()`, `execPython()`, and `execTypeScript()` all follow the same pattern:
- Check if value is a file path → resolve it → verify it exists
- Build `exec.Command` with the right binary and args
- Set Dir, Stdout, Stderr, Stdin, Env
- Run → handle `ExitError` → handle command-not-found
- Return error

This means adding a new step type (like `go run`, `ruby`, etc.) requires copy-pasting ~30 lines and knowing all the right places to touch. The `install.go` file has its own suppressed variants (`execBashSuppressed`, `execScriptSuppressed`) that duplicate parts of this further.

### 2. Validation Duplication (`validate.go`)
`checkRequirement()` and `CheckRequirementForDoctor()` are nearly identical — the only difference is that the doctor variant always detects version even without a min constraint. This is ~50 lines of duplicated code.

### Plan
1. Extract common `runCommand()` helper that handles file-path resolution, command construction, environment setup, and error wrapping — eliminating the boilerplate from each runner.
2. Merge `checkRequirement()` and `CheckRequirementForDoctor()` by adding an `alwaysDetectVersion` parameter to a single function.

### Design Constraints
- Zero behavior change — all 622 tests must pass unchanged.
- No new packages — this is an internal refactor within `executor`.
- Public API (`CheckRequirementForDoctor`, `ExitError`, etc.) stays the same.

## Acceptance Criteria
- [x] Common `runCommand()` helper extracted, used by all three runners
- [x] `checkRequirement` and `CheckRequirementForDoctor` deduplicated
- [x] All existing tests pass with no modifications
- [x] `go build ./...`, `go vet ./...`, `go test ./...` all clean
- [x] Architecture docs updated

## Implementation Notes
### Approach: `runCommand()` helper
Rather than a full interface (which would add indirection for minimal gain), extract a `runCommand()` method on `Executor` that takes:
- binary name (e.g. "bash", "python3", "tsx")
- command args (already resolved)
- stdout, stdin writers
- inputEnv + stepEnv

This eliminates the Dir/Env/Stdout/Stderr/Stdin/error-handling boilerplate.

Each runner keeps its own file-resolution and arg-building logic (which is genuinely different per language), but delegates the actual exec to `runCommand()`.

### Approach: validation dedup
Merge into a single `checkRequirementImpl(req, env, alwaysDetectVersion)` and have both `checkRequirement()` and `CheckRequirementForDoctor()` call it.

## Subtasks
- [x] Add `runCommand()` to helpers.go
- [x] Refactor `execBash()` to use it
- [x] Refactor `execPython()` to use it
- [x] Refactor `execTypeScript()` to use it
- [x] Deduplicate `checkRequirement` / `CheckRequirementForDoctor`
- [x] Run full test suite
- [x] Update docs

## Blocked By
