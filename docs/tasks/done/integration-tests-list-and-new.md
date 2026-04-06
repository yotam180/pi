# Integration Tests for pi list and pi new Commands

## Type
infra

## Status
done

## Priority
high

## Project
standalone

## Description
The `pi list` and `pi new` commands are critical user-facing commands that currently only have unit tests in `internal/cli/`. There are no integration tests that exercise the full compiled binary for these commands. Integration tests catch regressions that unit tests miss (flag parsing, output formatting, exit codes, file system interactions).

Add integration tests in `tests/integration/` covering:

**pi list:**
- Basic list output with table headers and automation names
- SOURCE column shows [workspace] for local automations
- INPUTS column shows input summaries
- --builtins flag includes pi:* automations
- Empty project (no automations) shows guidance message

**pi new:**
- Basic scaffold creates .pi/<name>.yaml
- --bash flag pre-fills with bash command
- --python flag pre-fills with python script
- --description flag sets description
- Nested paths create subdirectories
- Already-exists error
- No project error (suggests pi init)
- Strip .yaml extension

## Acceptance Criteria
- [x] Integration tests for `pi list` pass (13 tests)
- [x] Integration tests for `pi new` pass (14 tests)
- [x] All existing tests still pass
- [x] `go build ./...` succeeds
- [x] Architecture docs updated with new test counts

## Implementation Notes
- `pi list --builtins` shows builtins without the `pi:` prefix in the NAME column; the `[built-in]` indicator is in the SOURCE column. Tests check for `[built-in]` and known builtin names rather than `pi:` prefix.
- The `TestNew_CreatedFileIsRunnable` test verifies the full lifecycle: `pi new` creates a file, then `pi run` executes it successfully.
- Used `t.TempDir()` for all `pi new` tests to avoid polluting the examples directory.
- `pi list` on the packages example correctly shows the alias in the SOURCE column.

## Subtasks
- [x] Write list_test.go in tests/integration/
- [x] Write new_test.go in tests/integration/
- [x] Verify all tests pass
- [x] Update docs

## Blocked By
