# Rename `pipe_to: next` to `pipe: true`

## Type
improvement

## Status
done

## Priority
medium

## Project
12-yaml-ergonomics

## Description
The current `pipe_to: next` syntax is misleading — it implies the field could take values other than `next`, but `next` is the only valid value. The intent is simply "pipe this step's stdout to the next step's stdin." The cleaner form is `pipe: true`.

Both forms should work. `pipe_to: next` should emit a deprecation warning at parse time pointing authors toward `pipe: true`. `pipe_to: <anything-other-than-next>` remains a parse error.

## Acceptance Criteria
- [x] `pipe: true` is the canonical form and works identically to `pipe_to: next`
- [x] `pipe_to: next` continues to work but emits a deprecation warning
- [x] `pipe_to: <other-value>` remains a parse error
- [x] `pi validate` flags `pipe_to: next` as deprecated style
- [x] All example and built-in automations updated to use `pipe: true`
- [x] Tests cover: `pipe: true`, `pipe_to: next` (with warning), invalid `pipe_to` value

## Implementation Notes

### Approach
- Replaced `Step.PipeTo string` with `Step.Pipe bool` — the executor checks `step.Pipe` instead of `step.PipeTo == "next"`
- Added `Pipe *bool` field to `stepRaw` (yaml tag `pipe`) alongside the existing `PipeTo string` (yaml tag `pipe_to`)
- Added `resolvePipe()` method on `stepRaw` that normalizes both forms into a single `bool`:
  - `pipe: true` → `Pipe = true`
  - `pipe_to: next` → `Pipe = true` + deprecation warning
  - Both specified → parse error
  - `pipe_to: <non-next>` → parse error (same as before)
- `WarnWriter` package-level variable in `automation` package controls deprecation warning output
- CLI commands (`run`, `setup`, `validate`) set `WarnWriter = stderr` before loading automations
- `pi info` shows `[pipe]` annotation on steps with `pipe: true`
- All 8 example files and the polyglot README updated to use `pipe: true`
- No built-in automations used `pipe_to`, so no changes needed there

### Test coverage
- New unit tests in `step_test.go`: `TestLoad_PipeTrue`, `TestLoad_PipeFalseExplicit`, `TestLoad_PipeToNextDeprecationWarning`, `TestLoad_PipeAndPipeTo_Error`, `TestLoad_PipeToInvalidValue_Error`, `TestLoad_ParentShellWithPipeTrue_Error`, `TestLoad_FirstBlock_WithPipeTrue`
- New unit test in `automation_test.go`: `TestLoad_ShorthandWithPipeTrue`
- Existing backward-compat tests updated to check `step.Pipe` instead of `step.PipeTo`
- All existing integration tests pass unchanged

## Subtasks
- [x] Update schema/parser to accept `pipe: true`
- [x] Add deprecation warning for `pipe_to: next`
- [x] Update `pi validate` to flag old style
- [x] Update all example automations
- [x] Add/update tests

## Blocked By
