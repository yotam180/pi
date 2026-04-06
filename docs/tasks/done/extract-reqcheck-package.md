# Extract Runtime Requirement Checking into internal/reqcheck/

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `executor/validate.go` file contained all runtime requirement checking logic — version detection, version comparison, install hints, requirement satisfaction checks, and formatting. This logic is conceptually independent from step execution but lived in the `executor` package, creating unnecessary coupling:

1. **`cli/doctor.go`** imported `executor` solely for `CheckRequirementForDoctor`, `InstallHintFor`, and `ExitError` — three exports that have nothing to do with step execution.
2. **`executor.ValidateRequirements()`** is an Executor method that used requirement-checking helpers that don't need Executor state.
3. Future consumers (library API, IDE plugins, CI health checks) would need to import `executor` just to check requirements.

**Goal:** Extract requirement checking into `internal/reqcheck/` with clean, focused API.

## Acceptance Criteria
- [x] `internal/reqcheck/` package exists with all requirement-checking types and functions
- [x] `executor/validate.go` delegates to `reqcheck` for checking logic
- [x] `cli/doctor.go` imports `reqcheck` instead of `executor` for requirement checking
- [x] `ExitError` remains in executor (it's used broadly for step exit codes)
- [x] All existing tests pass unchanged
- [x] New unit tests in `reqcheck` package (42 tests, 95.9% coverage)
- [x] Existing executor tests that exercise requirement checking still pass
- [x] Executor coverage actually improved (92.7% → 94.2%)
- [x] `go build ./...`, `go test ./...`, `go vet ./...` all pass
- [x] Architecture docs updated

## Implementation Notes

### What moved to reqcheck

- `CheckResult` struct
- `ValidationError` struct
- `CheckRequirement()`, `CheckRequirementForDoctor()`, `checkImpl()` (shared logic)
- `DetectVersion()`, `detectVersionExec()`, `ExtractVersion()`, `CompareVersions()`, `parseVersionParts()`
- `RuntimeCommand()`
- `FormatRequirementLabel()`
- `FormatValidationError()`
- `installHints` map, `InstallHintFor()`
- `versionRegex`

### What stayed in executor

- `ExitError` — used broadly for step exit codes, not specific to requirement checking
- `ValidateRequirements()` — the Executor method that orchestrates checking + provisioning
- `tryProvision()` — depends on Executor's Provisioner field
- `BuildStepEnv()`, `prependPathInEnv()` — step execution utilities unrelated to requirement checking

### Backward compatibility approach

The executor package uses Go type aliases for seamless backward compatibility:
```go
type CheckResult = reqcheck.CheckResult
type ValidationError = reqcheck.ValidationError
var FormatValidationError = reqcheck.FormatValidationError
var InstallHintFor = reqcheck.InstallHintFor
var CheckRequirementForDoctor = reqcheck.CheckRequirementForDoctor
```

Plus unexported thin wrappers (`checkRequirement`, `extractVersion`, etc.) so existing internal tests continue compiling without changes.

### Coverage results
- `internal/reqcheck`: 95.9% (42 tests)
- `internal/executor`: 94.2% (up from 92.7% — extraction reduced the denominator)
- All 1467 tests pass across 18 packages

## Subtasks
- [x] Create task file
- [x] Create internal/reqcheck/ package with types
- [x] Move checking logic from executor/validate.go
- [x] Add type aliases and delegations in executor for backward compatibility
- [x] Update cli/doctor.go to import reqcheck
- [x] Write tests for reqcheck (42 tests)
- [x] Run full test suite (all pass)
- [x] Run go vet (clean)
- [x] Update architecture.md
- [x] Commit

## Blocked By
