# Code Quality Cleanup: Dead Code and Non-Deterministic Output

## Type
improvement

## Status
in_progress

## Priority
medium

## Project
standalone

## Description
Address three code quality issues discovered during codebase review:

1. **Non-deterministic env key ordering in `pi info`**: `printStepsDetail()` in `info.go` iterates `map[string]string` for env keys, producing unstable output. This is a subtle bug — running `pi info` twice on the same automation can show env keys in different orders.

2. **Dead code: `supportedStepTypes` and `implementedStepTypes` duplication**: These two maps in `automation.go` are identical (all step types are now implemented). They were originally meant to distinguish "planned" from "implemented" types, but that distinction no longer exists. Consolidate into one.

3. **Dead code: `appendInputEnv()` function**: This function in `helpers.go` was superseded by `buildEnv()` during the runtime provisioning work (task #32). It's no longer called anywhere but still exists as exported dead code.

## Acceptance Criteria
- [x] `pi info` env keys are sorted alphabetically for deterministic output
- [x] `supportedStepTypes` and `implementedStepTypes` consolidated into a single `validStepTypes` map
- [x] `appendInputEnv()` removed from helpers.go
- [x] New unit test for env key ordering in `pi info`
- [x] All existing tests pass (`go build ./...` and `go test ./...`)
- [x] Architecture docs updated

## Implementation Notes

### Non-deterministic env key ordering
- `info.go:138` iterates `s.Env` (a `map[string]string`) directly. Maps in Go have random iteration order.
- Fix: sort the env keys before display.
- Added a dedicated test `TestShowAutomationInfo_StepEnvSorted` that creates an automation with multiple env keys and asserts they appear in sorted order.

### supportedStepTypes / implementedStepTypes consolidation
- Both maps are identical: `{bash: true, run: true, python: true, typescript: true}`.
- `supportedStepTypes` is used in `stepRaw.toStep()` (line 395) — checked during YAML parse.
- `implementedStepTypes` is used in `validate()` (lines 485, 537) and `IsImplemented()` (line 636).
- Since all supported types are implemented, merged into a single `validStepTypes` map used everywhere.
- The `IsImplemented()` method renamed to `IsValid()` for clarity.

### appendInputEnv removal
- Function exists at helpers.go:29 but is never called — `buildEnv()` replaced all usages.
- Cleanly remove the function and its comment.

## Subtasks
- [x] Fix env key ordering in info.go
- [x] Add test for sorted env keys
- [x] Consolidate step type maps in automation.go
- [x] Remove appendInputEnv from helpers.go
- [x] Update architecture.md
- [x] Full QA pass

## Blocked By
