# Step execution visibility and silent/loud flags

## Type
feature

## Status
done

## Priority
medium

## Project
11-improve-display-ux

## Description
When `pi setup` runs an automation like `setup/install-rust` that contains multiple steps (e.g. the installer step followed by a `bash: rustup component add rustfmt clippy`), only the installer line is printed. The subsequent bash step executes silently, giving no indication that it ran at all. This is confusing — the user cannot tell whether the extra setup work happened.

This task makes step execution explicit by default and adds controls to suppress or force output:

1. **Default trace line per step**: Before executing any step (bash, python, typescript, run), pi prints a short trace line:
   ```
     → bash: rustup component add rustfmt clippy
     → run:  setup/install-cargo-audit
   ```
   The command is truncated to ~80 chars if long. Installer steps already have their own output format and are exempt.

2. **`silent: true` step flag**: A step may declare `silent: true` to suppress its trace line and its stdout/stderr output. Useful for noisy housekeeping commands that the user doesn't need to see.

3. **`--loud` CLI flag**: Passing `--loud` to `pi run` or `pi setup` forces all steps (including `silent: true` steps) to print their trace line and output. This is the escape hatch for debugging.

## Acceptance Criteria
- [x] Every non-installer step prints a `  → <type>: <truncated-command>` line before executing (by default)
- [x] A step with `silent: true` prints no trace and suppresses its stdout/stderr
- [x] `pi run --loud <name>` overrides `silent: true` and prints trace + output for all steps
- [x] `pi setup --loud` same as above for the full setup sequence
- [x] `silent: true` is valid on `bash:`, `python:`, `typescript:`, and `run:` steps
- [x] The YAML schema docs / README are updated to document `silent: true` and `--loud`
- [x] `go test ./...` passes
- [ ] Manual smoke test: `pi setup` on bat project shows the `rustup component add` trace line

## Implementation Notes

### Architecture
- `Silent` bool field added to `automation.Step` struct, parsed from YAML via `stepRaw.Silent`
- `Loud` bool field added to `executor.Executor` struct
- Trace lines are emitted via `display.Printer.StepTrace()` which uses dim styling and truncates at 80 chars
- `truncateTrace()` collapses multiline commands (shows first line + `...`) and truncates long values
- `execStepSuppressed()` wraps `execStep()` with stdout/stderr redirected to `io.Discard`
- When a silent step uses `pipe_to: next`, pipe capture still works — only non-pipe stdout is discarded
- `pi info` shows `[silent]` annotation on steps with `silent: true`
- Trace lines go to stderr (via Printer → Stderr writer), keeping stdout clean for piped output

### Test changes
- Integration tests that compare exact stdout output now use `runPiStdout()` helper (captures stdout only)
- Added `runPiSplit()` helper for tests needing separate stdout/stderr
- Executor test for stderr passthrough uses `strings.Contains` instead of exact match (trace lines added to stderr)

### Test coverage
- 2 automation schema tests (silent parsing, explicit false)
- 7 executor unit tests (default trace, run step trace, silent suppression, loud override, silent still executes, silent pipe capture)
- 7 integration tests (default trace lines, silent suppression, loud override, all-silent, all-silent+loud, info annotation)
- 10 display unit tests (StepTrace no color, with color, truncation, multiline collapse, truncateTrace table tests)

## Subtasks
- [x] Add `silent` field to step schema structs
- [x] Update YAML parsing to read `silent: true`
- [x] Add trace-line printing before each step execution in the runner
- [x] Honour `silent: true` by suppressing trace + output
- [x] Add `--loud` flag to `pi run` and `pi setup` CLI commands
- [x] Thread `loud` bool through execution context so it can override `silent`
- [x] Update README / docs with new flag and step option
- [x] Write unit tests for the runner (silent, loud, default)

## Blocked By
