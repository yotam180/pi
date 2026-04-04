# Pre-Execution Requirement Validation

## Type
feature

## Status
done

## Priority
high

## Project
05-environment-robustness

## Description
Before running any step, the executor should validate all `requires:` entries on the automation. If any requirement is not satisfied, execution should fail immediately with a clear, formatted error table listing every missing requirement and a concrete install hint for each.

The validation should:
1. For runtime requirements (`python`, `node`): check if the command exists in PATH, then check `--version` output against the minimum version constraint
2. For command requirements: check if the command exists in PATH, optionally check version
3. Produce a formatted error table:

```
✗ pi run format-logs

  Missing requirements:
    python >= 3.11   not found (python3 --version returned 3.9.7)
                     → install: brew install python@3.13   or  pi setup (if configured)
    command: jq      not found
                     → install: brew install jq
```

Validation should only check requirements for steps that will actually execute (respect `if:` conditions). For installer automations, check requirements at the automation level.

## Acceptance Criteria
- [x] `validateRequirements()` function in executor checks all requirements before step execution
- [x] Runtime version checking: runs `<cmd> --version`, parses semver output, compares against constraint
- [x] Command existence checking: uses `exec.LookPath()`
- [x] Formatted error output with missing requirements table and install hints
- [x] Validation respects `if:` conditions — skipped steps don't need their requirements checked
- [x] Install hints are generated for common tools (python, node, docker, jq, etc.)
- [x] Unit tests for validation logic (missing runtime, wrong version, missing command, all satisfied)
- [x] Integration test: automation with `requires: python >= 99.0` fails with clear error
- [x] `go test ./...` passes

## Implementation Notes

### Architecture
- New file `internal/executor/validate.go` contains all validation logic
- `CheckResult` struct holds per-requirement check results (satisfied, detected version, error)
- `ValidationError` type wraps multiple failed results for the formatted output
- `ValidateRequirements()` method on `Executor` checks all requirements before step execution
- Validation is called in `RunWithInputs()` after input resolution, before step/install execution

### Version detection
- `detectVersion()` runs `<cmd> --version` and captures combined stdout+stderr
- `extractVersion()` uses regex `(\d+(?:\.\d+)+)` to extract the first semver-like string
- Handles all common formats: `Python 3.13.0`, `v20.11.0`, `jq-1.7.1`, `docker version 24.0.5`, `kubectl v1.28.3`

### Version comparison
- `compareVersions()` splits on `.`, compares each numeric component
- Handles unequal component counts (e.g. `3.11` vs `3.11.0` → equal, `1.28` vs `1.28.3` → -1)

### Testability
- Added `ExecOutput` field to `RuntimeEnv` — optional mock for `<cmd> --version` calls
- When nil, real command execution is used; in tests, mock returns controlled output

### Install hints
- Map of common tool names → install commands (python, node, docker, jq, kubectl, helm, tsx, git, curl, wget, make, mise, uv)
- Unknown tools get no hint (empty string)

### Error output format
```
✗ pi run <automation-name>

  Missing requirements:
    <label>               <error>
                           → install: <hint>
```

## Subtasks
- [x] Implement `checkRequirement()` function for a single requirement
- [x] Implement version parsing from `--version` output (various formats: `Python 3.13.0`, `v20.11.0`, `jq-1.7.1`)
- [x] Implement semver comparison for `>=` constraints
- [x] Build install hint table for common tools
- [x] Integrate validation into `RunWithInputs()` before step execution
- [x] Write formatted error output
- [x] Write unit tests (24 tests in validate_test.go)
- [x] Write integration test (7 tests in examples_test.go + requires-validation example workspace)

## Blocked By
29-requires-schema-parsing (done)
