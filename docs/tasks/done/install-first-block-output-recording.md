# Fix execInstallFirstBlock to Record Step Output

## Type
bug

## Status
done

## Priority
medium

## Project
17-install-lifecycle-hardening

## Description

`execInstallFirstBlock` runs the matched sub-step's stdout to `io.Discard` and never calls `e.recordOutput()`. This means `outputs.last` returns stale data after a `first:` block inside an install phase. Every other execution path in the executor records output after each step; this is the one gap.

**Current (`executor/install.go`):**
```go
func (e *Executor) execInstallFirstBlock(a *automation.Automation, step automation.Step, index int, inputEnv []string, stderrWriter io.Writer) error {
    for j, sub := range step.First {
        // ...condition check...
        runner := e.registry().Get(sub.Type)
        ctx := e.newRunContext(a, sub, nil, io.Discard, nil, inputEnv, e.RepoRoot) // stdout discarded
        ctx.Stderr = stderrWriter
        return runner.Run(ctx)  // output never recorded
    }
    return nil
}
```

**Fix:** Capture stdout into a `bytes.Buffer` and call `e.recordOutput()`, matching the pattern in `execInstallPhaseWithStderr`:
```go
var outputCapture bytes.Buffer
ctx := e.newRunContext(a, sub, nil, &outputCapture, nil, inputEnv, e.RepoRoot)
ctx.Stderr = stderrWriter
if err := runner.Run(ctx); err != nil {
    return err
}
e.recordOutput(outputCapture.String())
return nil
```

## Acceptance Criteria
- [x] `execInstallFirstBlock` captures sub-step stdout and calls `e.recordOutput()`
- [x] A test verifies that `outputs.last` contains the first-block's output after execution
- [x] Existing install tests still pass
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

**Fix applied:** Changed `execInstallFirstBlock` to capture stdout into a `bytes.Buffer` instead of discarding it to `io.Discard`, then call `e.recordOutput()` after the runner completes — matching the pattern used in `execInstallPhaseWithStderr`.

**Before:**
```go
ctx := e.newRunContext(a, sub, nil, io.Discard, nil, inputEnv, e.RepoRoot)
return runner.Run(ctx)
```

**After:**
```go
var outputCapture bytes.Buffer
ctx := e.newRunContext(a, sub, nil, &outputCapture, nil, inputEnv, e.RepoRoot)
if err := runner.Run(ctx); err != nil { return err }
e.recordOutput(outputCapture.String())
return nil
```

**Tests added:**
- `TestExecInstall_FirstBlockRecordsOutput` — end-to-end install with first: block, verifies outer stepOutputs are correctly restored
- `TestExecInstallFirstBlock_OutputCapturedInPhase` — unit test calling `execInstallFirstBlock` directly, verifies `stepOutputs[0]` contains the first: block's stdout

## Subtasks
- [x] Fix `execInstallFirstBlock` to capture and record output
- [x] Write test for outputs.last after first: block in install phase

## Blocked By
