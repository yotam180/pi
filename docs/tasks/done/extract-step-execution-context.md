# Extract Step Execution Context Struct

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The executor's internal step execution methods (`execStep`, `execStepSuppressed`, `execFirstBlock`) all pass the same 7 parameters through their call chains:

```go
func (e *Executor) execStep(a *automation.Automation, step automation.Step, args []string, index int, stdinOverride io.Reader, capturePipe bool, inputEnv []string) error
```

This long parameter list is repeated identically in `execStepSuppressed`, `execFirstBlock`, and partially in `RunWithInputs`. It makes the code harder to read, harder to extend (adding a new execution concern requires changing every method signature), and creates subtle bugs when parameters are passed in the wrong order.

**Refactoring:** Extract these parameters into a `stepExecCtx` struct that bundles per-step execution state. Methods receive a single `*stepExecCtx` instead of 7 positional parameters.

```go
type stepExecCtx struct {
    automation    *automation.Automation
    args          []string
    inputEnv      []string
    stdinOverride io.Reader
    capturePipe   bool
}
```

The step itself and its index are not included because they vary per invocation within the same automation. The `stepExecCtx` represents the automation-level execution state that's constant across all steps in a single `RunWithInputs` call.

## Acceptance Criteria
- [x] `stepExecCtx` struct defined in executor package
- [x] `execStep` uses `*stepExecCtx` + step + index
- [x] `execStepSuppressed` uses `*stepExecCtx` + step + index
- [x] `execFirstBlock` uses `*stepExecCtx` + step + index
- [x] `RunWithInputs` constructs `stepExecCtx` and passes it through the step loop
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` pass
- [x] Code is measurably simpler (fewer parameters per call)

## Implementation Notes

### Approach
Introduced `stepExecCtx` in `executor.go` as an unexported struct that bundles the three parameters constant across all steps within a single `RunWithInputs` call:
- `automation *automation.Automation`
- `args []string`
- `inputEnv []string`

The remaining per-step parameters (`step`, `index`, `stdinOverride`, `capturePipe`) stay as direct arguments because they change for each step in the loop.

### Method signature changes
Before:
```go
func (e *Executor) execStep(a *automation.Automation, step automation.Step, args []string, index int, stdinOverride io.Reader, capturePipe bool, inputEnv []string) error
```

After:
```go
func (e *Executor) execStep(ctx *stepExecCtx, step automation.Step, index int, stdinOverride io.Reader, capturePipe bool) error
```

This reduces parameter count from 7 to 5 on `execStep`, `execStepSuppressed`, `execFirstBlock`, and from 3 to 2 on `execParentShell`. More importantly, it creates a single struct to extend when new per-automation execution state is needed.

### What's NOT included
The `install.go` methods (`execInstallPhase`, `execInstallFirstBlock`) were not refactored — they have a different parameter set (`phase`, `stderrWriter`) and don't benefit from the same struct. Forcing them would be artificial coupling.

### Test impact
Zero test changes required — `stepExecCtx` is internal to the executor package and invisible to tests. All 16 packages pass unchanged.

## Subtasks
- [x] Define `stepExecCtx` struct
- [x] Refactor `execStep` to accept `*stepExecCtx`
- [x] Refactor `execStepSuppressed` to accept `*stepExecCtx`
- [x] Refactor `execFirstBlock` to accept `*stepExecCtx`
- [x] Refactor `execParentShell` to accept `*stepExecCtx`
- [x] Refactor step loop in `RunWithInputs` to construct and pass `stepExecCtx`
- [x] Run tests
- [x] Update architecture.md

## Blocked By
