# Decouple RunContext from Executor State

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The `RunContext` struct (used by all step runners) currently depends on the `Executor` in two ways that hurt modularity:

1. **`BuildEnv` callback**: `RunContext.BuildEnv` is set to `e.buildEnv`, which is a method on `Executor`. But `buildEnv` only reads `e.runtimePaths` — which is already available in `RunContext.RuntimePaths`. This means `buildEnv` can be extracted to a standalone function that takes `runtimePaths` as a parameter, making `RunContext` fully self-contained and step runners independently testable without an `Executor` instance.

2. **Double `resolveStepDir` call**: `execStep()` calls `resolveStepDir()` to validate the directory exists, then `newRunContext()` calls `resolveStepDir()` again to set `WorkDir`. This means the directory is stat'd twice for every step with `dir:`. The fix: pass the resolved dir into `newRunContext` or resolve+validate in one place.

These changes improve:
- **Modularity**: `RunContext` becomes a self-contained value — no closures over Executor state for env building
- **Testability**: Step runners can be tested with a fully constructed `RunContext` without needing `Executor`
- **Expandability**: Future custom step types (user-registered runners) won't need to understand `Executor` internals
- **Performance**: Minor — avoids double stat for `dir:` steps

## Acceptance Criteria
- [x] `buildEnv` is a standalone function (not an Executor method)
- [x] `RunContext.BuildEnv` callback removed; replaced with a direct function or the field computed at construction
- [x] `resolveStepDir` is called once per step, not twice
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` and `go vet ./...` pass
- [x] Step runners remain unmodified (interface compatibility)

## Implementation Notes

### Approach

**buildEnv extraction:**
Changed `buildEnv` from an Executor method to a standalone package function `BuildStepEnv(runtimePaths, inputEnv, automationEnv, stepEnv)`. The `RunContext` no longer carries a `BuildEnv` callback — instead it carries `RuntimePaths` (already present) and callers use the standalone function. Since `runStepCommand` is the only consumer of `BuildEnv`, it can call `BuildStepEnv` directly using `ctx.RuntimePaths`.

**resolveStepDir fix:**
Moved the resolve+validate logic out of `execStep` and into the step dispatch site. `newRunContext` now takes an optional `resolvedDir` parameter so the caller can pass the already-validated path. When `resolvedDir` is empty, `newRunContext` still falls back to `RepoRoot`.

## Subtasks
- [x] Extract `buildEnv` to standalone function `BuildStepEnv`
- [x] Remove `BuildEnv` callback from `RunContext`
- [x] Update `runStepCommand` to call `BuildStepEnv` directly
- [x] Fix double `resolveStepDir` call
- [x] Update install.go to use new function
- [x] Run tests
- [x] Update architecture.md
- [x] Update task file

## Blocked By
