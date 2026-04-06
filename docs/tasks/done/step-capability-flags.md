# Add Step Capability Flags and Fix parent_shell Validation

## Type
improvement

## Status
done

## Priority
high

## Project
16-step-type-plugin-architecture

## Description

`parent_shell: true` is currently validated at YAML parse time by comparing the step type to the literal string `"bash"`. This is fragile — it's a name comparison, not a capability check, and would block any new shell-type language from ever supporting the feature. This task adds a `SupportsParentShell bool` flag to `SubprocessConfig` and moves the validation to execution time (or to `pi validate` via the step walker), where the actual runner is available.

## Acceptance Criteria
- [x] `SubprocessConfig` has a `SupportsParentShell bool` field
- [x] `NewBashRunner()` sets it to `true`; other runners leave it `false`
- [x] Parse-time `StepTypeBash` check is removed from `automation/step.go`
- [x] Validation of `parent_shell` against runner capability is preserved (either at validate time or execution time)
- [x] Error message is capability-driven, not type-name-driven
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

**Approach chosen: Both Option A and Option B (belt and suspenders).**

1. **`SubprocessConfig.SupportsParentShell`** — Added to `internal/executor/runners.go`. Only `NewBashRunner()` sets it to `true`.

2. **`Registry.StepTypeSupportsParentShell()`** — Added to `internal/executor/runner_iface.go`. Queries the runner's capability flag without exposing internals.

3. **Parse-time check removed** — The `s.t != StepTypeBash` guard in `automation/step.go` was deleted. The `parent_shell + pipe` structural check remains (that's about step semantics, not capability).

4. **Validation check (Option B)** — New `checkParentShellCapability` check registered in `validate.DefaultRunner()`. Uses the step walker + `executor.NewDefaultRegistry().StepTypeSupportsParentShell()` to catch invalid `parent_shell` usage at `pi validate` time.

5. **Runtime guard (Option A)** — `execParentShell` in `executor.go` checks `e.registry().StepTypeSupportsParentShell(step.Type)` before proceeding. This catches cases that bypass validation (e.g., programmatic automation construction).

6. **Error message** — `"step type %q does not support parent_shell"` — capability-driven, not name-driven.

7. **Tests updated:**
   - `step_test.go`: `TestLoad_ParentShellOnNonBashStep_Error` → `TestLoad_ParentShellOnNonBashStep_ParsesSuccessfully` (parse now succeeds)
   - `parent_shell_test.go`: Added `TestParentShell_PythonStep_Rejected`, `TestParentShell_TypeScriptStep_Rejected`, `TestRegistry_SupportsParentShell`
   - `validate_test.go`: Added 5 tests for `checkParentShellCapability` (bash ok, python error, typescript error, no parent_shell ok, inside first block). Updated `DefaultRunner` check count from 11 to 12.

## Subtasks
- [x] Add `SupportsParentShell` to `SubprocessConfig`
- [x] Update `NewBashRunner()`
- [x] Remove name-check from `step.go`
- [x] Add capability check in chosen location (validate or executor)
- [x] Update error message
- [x] Tests

## Blocked By
