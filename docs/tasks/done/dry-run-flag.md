# --dry-run Flag for pi run

## Type
feature

## Status
done

## Priority
high

## Project
standalone

## Description

Add a `--dry-run` flag to `pi run` that shows what steps would be executed without actually running them. This is a debugging and trust-building feature that aligns with the "Explain the magic, don't hide it" philosophy.

When `--dry-run` is set:
- Steps are printed with their trace line but NOT executed
- `if:` conditions are evaluated and skip/run decisions are shown
- `run:` steps are resolved and recursed into (showing the target's steps too)
- Installer automations show their lifecycle (test → run → verify)
- `first:` blocks show all sub-steps with conditions, marking which would match
- Input resolution and environment variable injection are computed and displayed
- No side effects occur (no commands run, no files written, no parent_shell eval)

Output format: reuse the existing step trace format but with a `[dry-run]` prefix or similar clear indicator. Show the full step expansion including resolved variables where possible.

## Acceptance Criteria
- [ ] `pi run --dry-run <name>` prints all steps that would be executed
- [ ] Conditional steps show their condition and whether they'd be skipped
- [ ] `run:` steps recurse into the target automation
- [ ] Installer automations show their lifecycle phases
- [ ] No commands are actually executed
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Unit tests cover dry-run mode
- [ ] Integration test confirms dry-run output

## Implementation Notes

### Architecture
- Added `DryRun bool` field to `Executor` struct
- Created `dry_run.go` in the executor package — clean separation from execution logic
- When `DryRun` is true, `RunWithInputs` delegates to `dryRunAutomation()` after input resolution and condition evaluation (these are side-effect-free and needed for accurate output)
- Dry-run output goes to stderr (same as trace lines), stdout remains empty

### Output Design
- Uses indentation to show nesting (run: step recursion, installer phases)
- Shows annotations in `[brackets]`: conditions, dir, timeout, silent, pipe, parent_shell
- `first:` blocks show all sub-steps with `← match`, `skipped`, or `not reached` indicators
- Installers show full lifecycle: test → run → verify → version
- Circular dependencies are caught gracefully with a message instead of crashing

### Testing
- 14 unit tests in `dry_run_test.go` covering all execution paths
- 2 CLI-level tests in `run_test.go` (dry-run shows output, does not execute)
- 5 integration tests in `tests/integration/dry_run_test.go` (end-to-end binary)
- Manual QA: `pi run --dry-run --with version=3.13 pi:install-python` shows rich output

## Subtasks
- [ ] Add DryRun field to Executor
- [ ] Add --dry-run flag to `pi run` CLI command
- [ ] Implement dry-run logic in RunWithInputs
- [ ] Handle installer dry-run
- [ ] Handle first: block dry-run
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Update docs

## Blocked By
