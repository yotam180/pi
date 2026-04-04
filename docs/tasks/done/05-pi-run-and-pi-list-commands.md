# `pi run` and `pi list` Commands

## Type
feature

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Wire the `pi run` and `pi list` CLI commands to the execution engine and automation discovery layer. After this task, PI is end-to-end functional for bash and `run:` steps.

## Acceptance Criteria
- [x] `pi run <name> [args...]` resolves the automation, runs its steps, and exits with the correct code
- [x] `pi run` with an unknown name prints a clear error and lists available automations
- [x] `pi run` without arguments prints usage
- [x] `pi list` prints a formatted table of all discovered automations: name and description
- [x] `pi list` with no automations found prints a friendly message (not an error)
- [x] Both commands walk up the directory tree to find `pi.yaml` (so they work from any subdirectory of the project, like `git` does)
- [x] `--help` on both commands is informative

## Implementation Notes

### Decisions
- **New package `internal/project`**: Contains `FindRoot()` which walks up from a start directory to find `pi.yaml`, similar to how git finds `.git/`. This keeps the root-finding logic reusable and testable in isolation.
- **Extracted functions**: `runAutomation()` and `listAutomations()` are extracted from the Cobra command handlers for testability. The Cobra handler just gets `os.Getwd()` and delegates.
- **Exit code propagation**: `runAutomation()` returns `*executor.ExitError` for non-zero step exit codes. The top-level `Execute()` in `root.go` checks for this type and calls `os.Exit(code)`.
- **Tabwriter output**: `pi list` uses `text/tabwriter` for cleanly aligned columns with NAME and DESCRIPTION headers. Automations with no description show "-" as placeholder.
- **Stdout/Stderr handling**: `runAutomation()` accepts `io.Writer` but checks if they're `*os.File` to pass to the executor (which needs `*os.File` for subprocess stdout/stderr). This keeps tests functional while preserving real terminal behavior.

### File structure
```
internal/project/root.go       — FindRoot() walks up to find pi.yaml
internal/project/root_test.go  — 4 tests
internal/cli/run.go            — newRunCmd() + runAutomation()
internal/cli/run_test.go       — 8 tests
internal/cli/list.go           — newListCmd() + listAutomations()
internal/cli/list_test.go      — 6 tests
internal/cli/root.go           — Execute() now handles ExitError for exit code propagation
internal/cli/root_test.go      — 6 tests (updated from stubs to real behavior)
```

### Test coverage (24 new/updated tests)
- **project/root**: find in project dir, find from subdirectory, not found error, picks closest pi.yaml
- **cli/run**: success, nested name (docker/up), not found error with available list, exit code propagation (42), args forwarding, run: step chaining, from subdirectory, no pi.yaml error
- **cli/list**: success with header/names/descriptions, no-description placeholder, empty (no automations), from subdirectory, no pi.yaml error, sorted order
- **cli/root**: help, run help, run requires arg, list help, setup stub, version

### Smoke test results
- `pi list` — aligned table with NAME/DESCRIPTION columns
- `pi run greet` — executes bash step correctly
- `pi run greet hello world` — args forwarded as $1 $2
- `pi run docker/up` — nested automation works
- `pi run chain` — run: step chaining works
- `pi run nonexistent` — error with available automations listed
- `pi run` (no args) — arg validation error
- `pi list` from subdirectory — walks up to find pi.yaml
- `pi list` with no .pi/ — friendly message
- Exit code propagation — `exit 42` in step → `pi run` exits 42

## Subtasks
- [x] Implement `pi.yaml` root finder (walk up from CWD)
- [x] Wire `pi run` cobra command to executor
- [x] Wire `pi list` cobra command to discovery
- [x] Format `pi list` output cleanly (align columns)
- [x] Manual smoke test against a temp workspace

## Blocked By
04-step-executor (done)
