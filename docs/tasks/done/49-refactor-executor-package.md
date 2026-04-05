# Refactor executor package for modularity

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description
The `internal/executor/executor.go` file has grown to 682 lines and mixes several concerns: step runner implementations (bash, python, typescript, run), installer lifecycle management, pipe support, suppression/silent logic, environment building, and helper utilities. This makes it harder to understand, extend (e.g. adding new step types), and maintain.

Refactor by extracting cohesive groups of functions into separate files within the same package:
1. **`runners.go`** — Step runner implementations: `execBash()`, `execPython()`, `execTypeScript()`, `execRun()`, `resolvePythonBin()`, `isCommandNotFound()`
2. **`install.go`** — Installer lifecycle: `execInstall()`, `execInstallPhase()`, `execInstallPhaseCapture()`, `execBashSuppressed()`, `execScriptSuppressed()`, `captureVersion()`, `printInstallStatus()`, `printIndentedStderr()`
3. **`helpers.go`** — Shared utilities: `isFilePath()`, `resolveScriptPath()`, `appendInputEnv()`, `buildEnv()`, `prependPathInEnv()`

The `executor.go` file retains: `Executor` struct, `ExitError`, `Run()`, `RunWithInputs()`, `execStep()`, `execStepSuppressed()`, `evaluateCondition()`, `pushCall()`/`popCall()`, and `stdout()`/`stderr()`/`stdin()` helpers.

No logic changes — purely structural. All tests must continue to pass unchanged.

## Acceptance Criteria
- [x] `runners.go` contains all step runner functions
- [x] `install.go` contains installer lifecycle functions
- [x] `helpers.go` contains shared utilities
- [x] `executor.go` is reduced to core orchestration (~200 lines)
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes with no changes to test files
- [x] `go vet ./...` passes
- [x] Architecture docs updated to reflect new file structure

## Implementation Notes
Pure move refactor — no logic changes, no API changes, no test changes. Functions are grouped by cohesion: runners handle specific language execution, install handles the structured lifecycle, helpers are stateless utilities.

## Subtasks
- [x] Extract step runners into `runners.go`
- [x] Extract installer lifecycle into `install.go`
- [x] Extract helpers into `helpers.go`
- [x] Verify all tests pass
- [x] Update `docs/architecture.md`

## Blocked By
