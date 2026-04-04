# Step Executor

## Type
feature

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Implement the execution engine that runs an automation's steps in order. Covers bash steps (inline and file) and `run:` steps (call another automation). Handles argument passing, working directory, exit code propagation, and clear failure output.

## Acceptance Criteria
- [x] `bash` step (inline): runs the string as a bash script, inherits stdout/stderr, `$@` receives args passed to `pi run`
- [x] `bash` step (file path): runs the `.sh` file, path resolved relative to the automation file's directory
- [x] `run:` step: looks up the named automation and executes it recursively; args are forwarded
- [x] Circular `run:` dependency detection with a clear error (A → B → A)
- [x] If any step exits non-zero, execution stops and `pi run` exits with the same code
- [x] Working directory for all steps is the repo root (the directory containing `pi.yaml`), not the `.pi/` subdirectory
- [x] Unit tests: bash inline success, bash inline failure (exit 1), file step, run: step chaining, circular dep detection

## Implementation Notes

### Decisions
- **Package**: `internal/executor` — clean separation from `automation` (parsing), `discovery` (resolution), and `cli` (commands).
- **Executor struct**: Holds `RepoRoot`, `Discovery` (for run: lookups), `Stdout`/`Stderr` (for testability), and a `callStack` for circular dependency detection.
- **ExitError type**: Wraps non-zero exit codes so callers can extract the code and propagate it to the process exit.
- **File path detection**: A bash step value is treated as a file path if it ends in `.sh`, contains no newlines, and contains no spaces. This distinguishes `scripts/deploy.sh` from `echo hello`.
- **File vs inline execution**: File steps use `bash <resolved_path> [args...]`; inline steps use `bash -c "<script>" -- [args...]`. This ensures `$1`, `$2`, etc. work correctly in both cases.
- **Circular dependency detection**: Uses a call stack (slice of automation names). On each `Run()`, the name is pushed; on return, it's popped. If a name appears twice, the full chain is reported (e.g., `a → b → c → a`).
- **Added `FilePath` field to Automation**: The `Automation` struct now stores its source file path (set by `Load()`), with a `Dir()` helper. This is needed for resolving relative `.sh` file references.
- **Working directory**: All steps run with `cmd.Dir = RepoRoot` (the directory containing `pi.yaml`).

### File structure
```
internal/executor/executor.go      — Executor, ExitError, Run(), execStep(), execBash(), execRun(), isFilePath(), resolveScriptPath()
internal/executor/executor_test.go — 20 tests
```

### Changes to existing packages
- `internal/automation/automation.go`: Added `FilePath` field (set by `Load()`), added `Dir()` method, added `path/filepath` import.

### Test coverage (20 tests)
- **Bash inline**: success, with args ($1 $2), failure (exit 42, correct exit code), multiline script
- **Bash file**: success, not found error, with args forwarded
- **Run step**: chaining (outer→inner), deep chaining (a→b→c), args forwarded through run, not found error
- **Circular dependency**: direct (a→a), indirect (a→b→c→a) with chain in error message
- **Multiple steps**: stops on first failure (step3 doesn't run), all succeed sequentially
- **Working directory**: confirms pwd matches repo root
- **Mixed steps**: bash and run: interleaved, correct execution order
- **ExitError message**: contains exit code
- **isFilePath**: table test with 6 cases
- **Call stack isolation**: call stack empty after sequential runs

## Subtasks
- [x] Implement bash step runner (inline)
- [x] Implement bash step runner (file path)
- [x] Implement run: step runner with recursion guard
- [x] Wire exit code propagation
- [x] Unit tests (20 tests)

## Blocked By
03-automation-discovery (done)
