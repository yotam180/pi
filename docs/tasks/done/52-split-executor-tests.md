# Split Executor Test File

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
The `internal/executor/executor_test.go` file is 2633 lines containing 109 tests that cover many different concerns: basic execution, conditional steps, automation-level if, installer lifecycle, step env, step trace/silent/loud, and parent shell. This monolithic file makes it hard to find relevant tests, add new tests in the right place, and review changes.

Split the file into focused test files, each covering one concern area. Shared test helpers remain in a common file. The split is purely organizational — no test logic changes, no new tests.

## Acceptance Criteria
- [x] `executor_test.go` is split into multiple files by concern area
- [x] Shared test helpers (newAutomation, newExecutor, bashStep, etc.) live in a dedicated helpers file
- [x] All 105 executor test functions pass after the split (215 runs including subtests)
- [x] Full test suite (`go test ./...`) passes — all 622 tests
- [x] Architecture docs updated with new test file breakdown

## Implementation Notes
Target files:
- `test_helpers_test.go` — shared helpers: newAutomation, newExecutor, bashStep, etc.
- `executor_test.go` — basic execution: bash inline/file, run step, circular deps, multi-step, working dir, mixed steps, exit error, call stack isolation
- `python_runner_test.go` — python inline/file, venv detection, mixed bash+python
- `typescript_runner_test.go` — typescript inline/file, tsx not found, mixed bash+typescript
- `pipe_test.go` — all pipe_to:next tests
- `inputs_test.go` — RunWithInputs tests
- `conditional_step_test.go` — step-level if: tests
- `conditional_automation_test.go` — automation-level if: tests
- `install_test.go` — installer lifecycle tests
- `step_env_test.go` — step-level env: tests, buildEnv tests
- `step_trace_test.go` — step trace, silent, loud tests
- `parent_shell_test.go` — parent shell tests

## Subtasks
- [x] Create task file
- [x] Split files (12 test files created)
- [x] Verify tests pass (105 functions, 215 runs including subtests)
- [x] Update docs (architecture.md updated)

## Blocked By
