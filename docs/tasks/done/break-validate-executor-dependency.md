# Break validate → executor Dependency

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `internal/validate` package imports `internal/executor` for three things:
1. `executor.DefaultFileExtensions()` — maps step types to file extensions
2. `executor.IsFilePath()` — pure string check (value ends with ext, no newlines, no spaces)
3. `executor.NewDefaultRegistry()` + `StepTypeSupportsParentShell()` — checks parent_shell capability per step type

This creates a dependency from validation → execution, which is architecturally backwards. Validation should be independent of the execution engine. The three items above are all static metadata about step types, not execution logic.

**The fix:** Move `IsFilePath` to the `automation` package (where `StepType` is defined) and add static metadata maps for default file extensions and parent-shell capability. The `validate` package then imports only `automation` (which it already does) instead of `executor`.

The executor's `SubprocessConfig`/`Registry` remain the runtime source of truth. The `automation` package provides the static compile-time metadata that validation needs.

**Why this matters:**
- Cleaner dependency graph: validate depends on automation (data model), not executor (runtime)
- Enables future scenarios where validation runs without executor being initialized
- Reduces coupling surface — changes to executor don't ripple into validate
- Follows the established pattern of extracting metadata closer to its type definitions

## Acceptance Criteria
- [x] `IsFilePath()` lives in `automation` package
- [x] Default file extensions per step type available from `automation` package
- [x] Parent-shell capability per step type available from `automation` package
- [x] `validate` package no longer imports `executor`
- [x] Executor's `IsFilePath` and `DefaultFileExtensions` removed (no external callers)
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` pass
- [x] No behavior change

## Implementation Notes

### Approach
- Created `internal/automation/step_metadata.go` with three functions:
  - `IsFilePath(value, ext string) bool` — identical logic to the former `executor.IsFilePath`
  - `DefaultFileExtensions() map[StepType]string` — returns a copy of the static extension map (bash→.sh, python→.py, typescript→.ts)
  - `StepTypeSupportsParentShell(stepType StepType) bool` — static lookup (only bash supports parent_shell)
- Updated `internal/validate/validate.go` to import from `automation` instead of `executor`:
  - `checkFileReferences` uses `automation.DefaultFileExtensions()` and `automation.IsFilePath()`
  - `checkParentShellCapability` uses `automation.StepTypeSupportsParentShell()` (no longer creates a full executor Registry)
- Updated `internal/executor/helpers.go`:
  - Removed the exported `IsFilePath` function (no external callers)
  - `resolveFileStep()` now calls `automation.IsFilePath()` directly
- Removed `executor.DefaultFileExtensions()` from `runner_iface.go` (no callers after validate migration)
- Removed `TestIsFilePath` from `executor_test.go` (equivalent tests now in `automation/step_metadata_test.go`)

### Coverage impact
- `automation`: 89.3% → 89.5% (new tests)
- `executor`: 89.9% → 90.8% (removed dead code)
- `validate`: 97.2% → 97.2% (unchanged)

### What was NOT changed
- Executor's `validate.go` thin wrappers (checkRequirement, runtimeCommand, etc.) — these are used by 39+ executor tests. Migrating them to import `reqcheck` directly is a separate task.
- Executor's `Registry.FileExtForStepType()` and `Registry.StepTypeSupportsParentShell()` — these remain as the runtime capability query mechanism for the executor. The automation package provides the static compile-time equivalent for validation.

### Testing
- 4 new tests in `step_metadata_test.go`: IsFilePath, DefaultFileExtensions (with copy safety), StepTypeSupportsParentShell, DefaultFileExtensionsConsistency
- All 1000+ existing tests pass unchanged
- `go vet ./...` clean

## Subtasks
- [x] Add `IsFilePath()` to `automation` package
- [x] Add `DefaultFileExtensions()` to `automation` package
- [x] Add `StepTypeSupportsParentShell()` to `automation` package
- [x] Update `validate` to import from `automation` instead of `executor`
- [x] Remove `executor.IsFilePath` and `executor.DefaultFileExtensions` (no external callers)
- [x] Verify all tests pass
- [x] Update architecture.md

## Blocked By
