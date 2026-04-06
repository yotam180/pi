# Extract Testable Command Execution in Runtimes Package

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `internal/runtimes` package has the lowest test coverage in the project (71.5%). The root cause is that `provisionWithMise` (14.3% coverage), `provisionNodeDirect` (73.5%), and `provisionPythonDirect` (76.9%) all call `exec.Command` directly, making them impossible to unit test without real binaries and network access.

Extract command execution into a pluggable `CmdRunner` interface on the `Provisioner` struct (similar to the existing `LookPath` override pattern). This enables comprehensive unit testing of all provisioning logic — command construction, argument assembly, error handling, directory creation, symlink management — without requiring mise, curl, or network.

## Acceptance Criteria
- [x] `Provisioner` struct has a `CmdRunner` field (defaults to real exec at runtime)
- [x] All exec.Command calls in runtimes.go go through CmdRunner
- [x] `provisionWithMise` has tests covering: mise install success, mise where, symlink creation, mise install failure, mise where failure
- [x] `provisionNodeDirect` and `provisionPythonDirect` have tests covering error paths (unsupported OS/arch)
- [x] Coverage of `internal/runtimes` package is ≥ 85%
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### CmdRunner Interface Design
- Two methods: `Run(bin, args, stdout, stderr)` and `Output(bin, args, stderr) (string, error)`
- `Run` is for commands where we only care about success/failure (mise install, bash scripts)
- `Output` is for commands where we need to capture stdout (mise where)
- Default implementation `execCmdRunner` uses `os/exec` — zero overhead for production
- Follows the same pattern as the existing `LookPath` override on Provisioner

### Refactoring
- All three provisioning methods (`provisionWithMise`, `provisionNodeDirect`, `provisionPythonDirect`) refactored to use `p.runner()` helper
- `runner()` returns `p.Runner` if set, otherwise returns `&execCmdRunner{}`
- The `cp -a` call in `provisionNodeDirect` was also routed through the runner for consistency

### Test Coverage
- Before: 71.5% overall, `provisionWithMise` at 14.3%
- After: 89.2% overall, `provisionWithMise` at 86.7%, `Provision` at 100%
- 14 new tests added via `mockCmdRunner` that records calls and supports configurable error/output functions
- Tests verify: command arguments, binary paths, symlink creation, symlink overwrite, error propagation, empty bin dirs, nonexistent dirs, default version fallback

### Remaining uncovered lines (89.2% → 100% gap)
- `homeDir()` error path (UserHomeDir failure — hard to trigger)
- Platform-specific OS/arch branches in direct provisioners (only the current platform's branches are exercised)
- These are acceptable — they're environment-dependent code paths

## Subtasks
- [x] Design CmdRunner interface
- [x] Refactor provisionWithMise to use CmdRunner
- [x] Refactor provisionNodeDirect and provisionPythonDirect to use CmdRunner
- [x] Write unit tests for provisionWithMise
- [x] Write unit tests for direct provisioning error paths
- [x] Verify coverage improvement

## Blocked By
