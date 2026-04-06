# Unify Step Dispatch in Executor

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The executor package had duplicated step dispatch logic in two places:

1. **`executor.go` → `RunWithInputs()`** — the main step loop
2. **`install.go` → `execInstallPhaseWithStderr()`** — the install phase loop

Both duplicated: runner lookup, RunContext construction, output capture, and recording. The install phase also had a separate `execInstallFirstBlock()` that duplicated the runner dispatch from the main `execFirstBlock()`.

Additionally, `execStepSuppressed()` temporarily mutated `e.Stdout`/`e.Stderr` on the Executor struct — a fragile pattern that could leave state corrupted on unexpected returns.

## Acceptance Criteria
- [x] `execStepSuppressed()` eliminated — no more Executor state mutation for silent steps
- [x] `execStep()` takes explicit `displayStdout` and `stderrW` parameters
- [x] `newRunContext()` takes explicit `stderrW` parameter instead of reading `e.stderr()`
- [x] Install phase steps dispatch through `execStep()` instead of inline runner dispatch
- [x] `execInstallFirstBlock()` delegates to `execStep()` for the matched sub-step
- [x] All existing tests pass (`go test ./...`) — 0 failures
- [x] All integration tests pass
- [x] No new linter warnings from `go vet ./...`
- [x] Architecture docs updated
- [x] Coverage maintained (89.4% → 89.5%)

## Implementation Notes

### Approach taken

Refactored `execStep()` from reading I/O destinations from Executor fields to accepting them as explicit parameters:

```go
// Before:
func (e *Executor) execStep(ctx, step, index, stdinOverride, capturePipe)
// After:
func (e *Executor) execStep(ctx, step, index, stdinOverride, capturePipe, displayStdout, stderrW)
```

**Key changes:**

1. **`execStep` explicit I/O**: `displayStdout` is where visible output goes (tee'd with capture buffer for `outputs.last`). `stderrW` is where stderr goes. Callers pass `io.Discard` for suppression.

2. **`execStepSuppressed` eliminated**: The main loop and `execFirstBlock` now compute `stdoutW, stderrW` based on the silent/loud flags and pass them directly. No more `e.Stdout`/`e.Stderr` swapping.

3. **`newRunContext` explicit stderr**: Now takes a `stderrW` parameter instead of always reading `e.stderr()`. This lets install phases route stderr to their controlled writer.

4. **Install phase uses `execStep`**: `execInstallPhaseWithStderr` now creates a `stepExecCtx` and calls `execStep(ctx, step, i, nil, false, io.Discard, stderrWriter)`. This shares runner lookup, dir resolution, output capture, and `outputs.last` recording.

5. **Install first-block uses `execStep`**: `execInstallFirstBlock` now takes a `stepExecCtx` and delegates the matched sub-step to `execStep()` instead of building its own RunContext.

### What was NOT changed
- `RunAutomation` closure still temporarily swaps `e.Stdout`/`e.Stdin` for `run:` steps — this is a deeper concern that affects the recursive execution model and is best addressed separately.
- `execFirstBlock` and `execInstallFirstBlock` remain separate functions — they have different orchestration needs (pipe support, parent_shell, trace output vs. pure suppressed execution).
- `captureVersion` still builds its own RunContext — it's a standalone utility that doesn't benefit from `execStep`'s output recording.

## Subtasks
- [x] Analyze the exact overlap between the two loops
- [x] Design explicit I/O parameter approach
- [x] Add `displayStdout` and `stderrW` params to `execStep`
- [x] Add `stderrW` param to `newRunContext`
- [x] Eliminate `execStepSuppressed`
- [x] Refactor `execInstallPhaseWithStderr` to use `execStep`
- [x] Refactor `execInstallFirstBlock` to use `execStep`
- [x] Update test (`execInstallFirstBlock` signature change)
- [x] Run tests, vet, verify coverage
- [x] Update architecture docs

## Blocked By
