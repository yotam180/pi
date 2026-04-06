# Executor Interpolation and IO Test Coverage

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The executor package is at 90.3% coverage overall, but several functions have significant gaps:

1. **`interpolateWith` — 0% coverage**: Wraps `interpolateWithCtx` with nil inputEnv. No test exercises it.
2. **`stdout()` / `stderr()` — 66.7% coverage each**: The nil-fallback branches (defaulting to os.Stdout/os.Stderr) are never tested.
3. **`printer()` — 66.7% coverage**: The lazy-creation path (when `e.Printer` is nil) is tested only indirectly.
4. **`evaluateCondition` — 76.9% coverage**: The initial Predicates parse-error branch is not tested.
5. **`AppendToParentEval` — 71.4% coverage**: The write-error path is not tested.
6. **`registry()` — 80% coverage**: The `e.Runners != nil` path isn't directly tested.
7. **`execFirstBlock` — 81% coverage**: The no-match-with-capturePipe path (empty lastPipeBuffer) needs a test.
8. **`resolveScriptPath` — 66.7%**: The absolute path branch isn't tested.

These gaps are straightforward to close with unit tests that exercise the specific branches.

## Acceptance Criteria
- [x] `interpolateWith` has test coverage
- [x] `stdout()` / `stderr()` nil-fallback branches tested
- [x] `printer()` lazy creation tested
- [x] `evaluateCondition` parse error branch tested
- [x] `AppendToParentEval` error paths tested
- [x] `registry()` custom Runners path tested
- [x] `execFirstBlock` no-match + capturePipe tested
- [x] `resolveScriptPath` absolute path branch tested
- [x] All existing tests pass
- [x] `go build ./...` and `go test ./...` pass
- [x] Coverage increases from 90.3%

## Implementation Notes

### Approach
Created a single `coverage_gaps_test.go` file with 53 focused unit tests targeting the specific uncovered branches. Tests follow the established patterns from the existing test files (table-driven where appropriate, helper functions from `test_helpers_test.go`).

### Coverage improvement
- **Before: 90.3% → After: 93.1%** (+2.8 percentage points)
- Functions brought to 100%: `interpolateWith`, `interpolateWithCtx`, `stdout()`, `stderr()`, `stdin()`, `printer()`, `execStep`, `registry()`, `resolveScriptPath`, `isCommandNotFound`, `formatRequirementLabel`, `runtimeCommand`, `installHint`, `FormatValidationError`, `extractVersion`, `compareVersions`
- Functions significantly improved: `evaluateCondition` (76.9% → 92.3%), `execFirstBlock` (81% → 95.2%), `AppendToParentEval` (71.4% → 85.7%)

### What remains uncovered
- `detectVersionExec` (0%) — runs actual commands, untestable without external binaries
- `ValidationError.Error()` (0%) — trivial one-liner, low value
- `captureVersion` (83.3%) — requires install flow to trigger remaining branch
- `writeTempScript` error paths (63.6%) — would require simulating disk-full conditions

### Tests added (by category)
- **interpolation**: 5 tests (empty/nil, empty map, outputs.last, literal, multiple keys)
- **IO accessors**: 5 tests (stdout/stderr/stdin nil fallback, explicit writer)
- **printer**: 2 tests (explicit, lazy creation)
- **evaluateCondition**: 4 tests (parse error, unknown predicate, automation-level, step-level)
- **AppendToParentEval**: 2 tests (invalid path, read-only file)
- **registry**: 3 tests (custom runners, cached default, custom takes precedence)
- **execFirstBlock**: 3 tests (no-match pipe capture, parent shell sub-step, condition error)
- **resolveScriptPath**: 2 tests (absolute, relative)
- **resolveStepDir**: 3 tests (empty, absolute, not-a-directory)
- **misc helpers**: 13 tests (isCommandNotFound, writeTempScript, resolvePythonBin ± venv, unimplemented step type)
- **validation formatting**: 9 tests (FormatValidationError, formatRequirementLabel, runtimeCommand, InstallHintFor)
- **version detection**: 9 tests (detectVersion mocked, extractVersion patterns, compareVersions)
- **env construction**: 2 tests (BuildStepEnv empty, with runtime paths)
- **stepExecCtx**: 1 test (roundtrip)

## Subtasks
- [x] Create task file
- [x] Write tests for each gap
- [x] Run tests and verify coverage
- [x] Update architecture.md
- [x] Commit

## Blocked By
