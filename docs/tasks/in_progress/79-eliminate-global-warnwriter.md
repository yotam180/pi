# Eliminate Global automation.WarnWriter

## Type
improvement

## Status
in_progress

## Priority
high

## Project
standalone

## Description
The `automation` package has a global mutable variable `WarnWriter io.Writer` that is set/unset via `defer` in multiple CLI commands (`run.go`, `setup.go`, `validate.go`). This pattern is:
- Not thread-safe
- Hard to reason about (action-at-a-distance)
- Prevents concurrent automation loading
- Inconsistent with the rest of the codebase (e.g. `discovery.Discover()` takes `warnWriter` as a parameter)

The fix is to thread the warn writer through as a parameter to `Load()` and `LoadFromBytes()`, and remove the global variable entirely.

## Acceptance Criteria
- [x] `automation.WarnWriter` global variable is removed
- [x] `automation.Load()` and `automation.LoadFromBytes()` accept `warnWriter io.Writer` parameter
- [x] `discovery.Discover()` threads `warnWriter` through to automation loading
- [x] `builtins.Discover()` threads `warnWriter` through to automation loading
- [x] CLI commands no longer set/unset `automation.WarnWriter`
- [x] All existing tests pass
- [x] No regressions in deprecation warning output for `pipe_to: next`

## Implementation Notes
- The `WarnWriter` is only used in one place in the automation package: `step.go` line ~298, inside `resolvePipe()`, to emit a deprecation warning for `pipe_to: next`
- The approach: add `warnWriter io.Writer` to the `Load`/`LoadFromBytes` functions, thread it through the YAML unmarshalling chain via the step parsing
- Since `UnmarshalYAML` can't easily accept parameters, we'll pass the warn writer through the automation struct before unmarshalling, or use a different approach
- Decision: Add a `WarnWriter` field to the `Automation` struct (unexported: `warnWriter`), set it before unmarshal, and have step parsing use it via a package-level function that reads from a context. Actually, better approach: since YAML unmarshal calls `stepRaw.toStep()` → `resolvePipe()`, we can pass warnWriter into the parsing chain by having `UnmarshalYAML` receive it indirectly.
- Cleanest approach: Since go's `yaml.Unmarshal` calls `UnmarshalYAML` without extra context, we'll use a thin wrapper. We'll create an internal `loadContext` that holds the warn writer, and have the step parsing consult it. Given that YAML parsing is inherently single-threaded per call, a function-scoped approach works.
- Final decision: Pass `warnWriter` to `Load`/`LoadFromBytes`. These functions parse steps via `yaml.Unmarshal` → `Automation.UnmarshalYAML` → `stepRaw.toStep()` → `resolvePipe()`. The simplest correct approach: set a field on the Automation struct before/after unmarshal that `resolvePipe` can access. Since `Automation.UnmarshalYAML` is the entry point, we can set the warnWriter on the automation in `Load`/`LoadFromBytes` before calling unmarshal, and have `resolvePipe` accept it as a parameter threaded from `toStep`.

## Subtasks
- [x] Add `warnWriter` parameter to `resolvePipe()` and `stepRaw.toStep()`
- [x] Add `warnWriter` parameter to `Load()` and `LoadFromBytes()`
- [x] Thread `warnWriter` through `Automation.UnmarshalYAML` (not possible directly — use post-unmarshal approach or pre-set field)
- [x] Update `discovery.Discover()` to pass `warnWriter` through
- [x] Update `builtins.Discover()` to pass `warnWriter` through
- [x] Remove `WarnWriter` global and CLI set/unset patterns
- [x] Update all tests

## Blocked By
