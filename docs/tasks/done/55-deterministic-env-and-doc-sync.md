# Deterministic Env Ordering and Architecture Doc Sync

## Type
improvement

## Status
in_progress

## Priority
medium

## Project
standalone

## Description
Task #54 fixed non-deterministic env key ordering in `pi info` output. However, the same class of issue exists in two other places:

1. **`buildEnv()` in `helpers.go`**: Iterates `stepEnv` (`map[string]string`) without sorting. While env var order doesn't affect lookup, it makes the environment list non-deterministic, which can make tests brittle and debugging harder.

2. **`InputEnvVars()` in `automation.go`**: Iterates `resolved` (`map[string]string`) without sorting, producing `PI_INPUT_*` env vars in random order.

Additionally, `architecture.md` has stale references from task #54's changes:
- References to `appendInputEnv()` (removed)
- References to `supportedStepTypes`/`implementedStepTypes` (consolidated to `validStepTypes`)
- Reference to `IsImplemented()` (renamed to `IsValid()`)

## Acceptance Criteria
- [x] `buildEnv()` iterates step env keys in sorted order
- [x] `InputEnvVars()` iterates input keys in sorted order
- [x] Architecture doc updated: no stale references to removed/renamed symbols
- [x] Tests verify deterministic ordering where applicable
- [x] All existing tests pass

## Implementation Notes

### `InputEnvVars()` fix
- Added `sort.Strings()` on the key list before iterating to produce `PI_INPUT_*` vars.
- New test `TestInputEnvVars_DeterministicOrder` runs 20 iterations with 4 keys to verify stable alphabetical ordering.

### `buildEnv()` fix
- When step env has entries, collect keys into a slice, sort it, then iterate in order.
- New test `TestBuildEnv_StepEnvDeterministicOrder` runs 20 iterations with 3 keys, extracts the step env entries from the full environment, and asserts they appear in alphabetical order.

### Architecture doc sync
- Removed all references to `appendInputEnv()` (function removed in task #54)
- Updated `supportedStepTypes`/`implementedStepTypes` → `validStepTypes` in the "Adding a new step type" guide
- Updated stale description of `appendInputEnv()` in the Inputs section to describe current `InputEnvVars()` and `buildEnv()` behavior
- Fixed drifted test counts: automation 87→88, executor 161→162, display 35→30, runtimes 17→16, validate 40→34, predicates 11→12, step_env 8→9, total 636→638

## Subtasks
- [x] Sort step env keys in `buildEnv()` (helpers.go)
- [x] Sort input keys in `InputEnvVars()` (automation.go)
- [x] Update architecture.md stale references
- [x] Add/update tests for deterministic ordering
- [x] Full QA pass

## Blocked By
