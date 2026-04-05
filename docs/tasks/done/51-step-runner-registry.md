# Step Runner Registry Pattern

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
Refactor the executor's step dispatch to use a registry-based `StepRunner` interface instead of the hardcoded switch statement in `execStep()`. Currently, adding a new step type requires modifying the switch in `executor.go`, adding a method on the `Executor` struct in `runners.go`, and updating the install.go dispatch. This violates the open/closed principle and makes the system harder to extend.

The goal is to define a `StepRunner` interface, implement it for each existing step type (bash, python, typescript, run), and use a registry to resolve runners by step type. New step types can then be added by registering a runner without touching the executor core.

This is a pure refactor — no behavior changes, all existing tests must continue to pass.

## Acceptance Criteria
- [x] `StepRunner` interface defined with a clean contract
- [x] `Registry` type that maps `StepType` → `StepRunner`
- [x] `BashRunner`, `PythonRunner`, `TypeScriptRunner`, `RunStepRunner` implement `StepRunner`
- [x] `Executor.execStep()` dispatches through the registry instead of a switch
- [x] Install phase step dispatch uses the same registry pattern
- [x] All 622+ existing tests pass unchanged
- [x] `architecture.md` updated with the new pattern
- [x] Adding a new step type is now: implement interface + register — no executor changes needed

## Implementation Notes
### Design decisions:
- The `StepRunner` interface takes an `ExecutionContext` struct rather than having runners depend on `*Executor`. This keeps runners decoupled from executor internals.
- `ExecutionContext` bundles: automation, step, args, stdout, stdin, inputEnv, repoRoot, stderr, runtimePaths, discovery, buildEnv func.
- The registry is a simple `map[StepType]StepRunner` — no fancy plugin system. KISS.
- Runners are created in a `NewDefaultRegistry()` factory, keeping registration centralized but open.
- The `RunStepRunner` needs access to the executor for recursive `RunWithInputs()` calls, so it receives a callback function rather than a direct reference.

### Key files modified:
- `internal/executor/runner_iface.go` — new file: interface, context, registry
- `internal/executor/runners.go` — refactored: methods → struct implementations
- `internal/executor/executor.go` — updated: uses registry
- `internal/executor/install.go` — updated: uses registry for non-bash steps
- `internal/executor/helpers.go` — unchanged: still provides shared utilities

## Subtasks
- [x] Define StepRunner interface and ExecutionContext
- [x] Create Registry type with NewDefaultRegistry()
- [x] Refactor bash, python, typescript, run runners to structs
- [x] Wire registry into Executor
- [x] Wire registry into install.go dispatch
- [x] Run full test suite
- [x] Update architecture.md

## Blocked By
