# Refactor install phase scalar execution to use step runner registry

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `install.go` file contains `execBashLive()` and `execBashSuppressed()` functions that bypass the step runner registry (`Registry` / `StepRunner` interface) and directly call `exec.Command("bash", ...)`. This creates several problems:

1. **Automation-level env not applied**: Scalar install phases call `e.buildEnv(inputEnv, nil, nil)`, passing `nil` for both `automationEnv` and `stepEnv`. This means automation-level `env:` declarations are silently ignored during scalar install phases. Step-list install phases already go through the registry and properly apply env.

2. **No timeout support**: Scalar install phases don't use `runStepCommand()`, so they can't benefit from timeout support.

3. **Duplicated logic**: The bash command setup pattern (inline vs file detection, arg construction, exec.Command creation, error wrapping) is duplicated in `execBashLive`/`execBashSuppressed` vs what `BashRunner` does.

4. **`captureVersion()` also bypasses**: The version-capture function similarly creates its own `exec.Command` without going through the runner.

The fix is to refactor `execBashLive`, `execBashSuppressed`, and `captureVersion` to construct a `RunContext` and dispatch through the runner registry, just like `execInstallPhaseLive` and `execInstallPhaseCapture` already do for step-list phases.

## Acceptance Criteria
- [x] `execBashLive()` removed — scalar install run phases dispatched through the registry
- [x] `execBashSuppressed()` removed — scalar install test/verify phases dispatched through the registry
- [x] `captureVersion()` uses the runner registry (or at minimum `buildEnv` with automation env)
- [x] Automation-level `env:` is correctly applied during all install phases (scalar and step-list)
- [x] All existing tests pass without modification
- [x] New test confirms automation-level env is applied during scalar install phases
- [x] `go build ./...` and `go test ./...` succeed

## Implementation Notes

### Analysis

Looking at the existing code flow:

- `execInstallPhase` → `execInstallPhaseCapture` → for scalars, calls `execBashSuppressed`; for step lists, iterates steps and dispatches through registry
- `execInstallPhaseLive` → for scalars, calls `execBashLive`; for step lists, iterates steps and dispatches through registry
- `captureVersion` → directly `exec.Command("bash", "-c", versionCmd)`

The step-list paths already construct `RunContext` properly via `newRunContext`, which calls `buildEnv` through the `BuildEnv` callback. The scalar paths skip all of this.

### Approach

Refactor the scalar paths to construct a synthetic `automation.Step` of type bash with the scalar value, then dispatch through the same runner path as step lists. This means:
1. Creating a temporary `automation.Step{Type: StepTypeBash, Value: phase.Scalar}`
2. Creating a `RunContext` via `newRunContext` (which sets up BuildEnv properly)
3. Overriding Stdout/Stderr as needed for the phase (live vs suppressed)
4. Calling `runner.Run(ctx)`

This preserves the existing behavior (stdout suppressed, stderr control) while gaining:
- Proper env layering (input + automation + step)
- File path detection for scalar scripts
- Timeout support infrastructure

## Subtasks
- [x] Refactor `execInstallPhaseCapture` to handle scalars through the registry
- [x] Refactor `execInstallPhaseLive` to handle scalars through the registry
- [x] Refactor `captureVersion` to pass automation env
- [x] Remove `execBashLive` and `execBashSuppressed`
- [x] Add test for automation-level env in scalar install phases
- [x] Run full test suite

## Blocked By
