# Step-Level Timeout Support

## Type
feature

## Status
done

## Priority
high

## Project
standalone

## Description
Add a `timeout:` field on steps that sets a maximum execution duration. When a step exceeds its timeout, PI kills the process and returns a clear error. This prevents automations from hanging forever — a common pain point with scripts that wait for input, slow network calls, or infinite loops.

The `timeout:` field accepts a Go-style duration string (e.g., `30s`, `5m`, `1h30m`). It works with all step types (bash, python, typescript) and is compatible with all existing step features (env, dir, silent, if, pipe_to). `run:` steps do not support timeout directly — the child automation's own step timeouts apply.

`parent_shell` steps do not support `timeout:` since they don't execute as subprocesses.

## Acceptance Criteria
- [x] `timeout:` field parsed from step YAML (duration string like `30s`, `5m`)
- [x] Invalid timeout values produce clear parse-time error
- [x] Steps exceeding their timeout are killed (process + group) with a clear error message
- [x] Timeout error is returned as an `ExitError` with a distinct code (124, matching `timeout` command convention)
- [x] `timeout:` + `pipe_to: next` works correctly (timeout during capture)
- [x] `timeout:` + `silent: true` works correctly
- [x] `timeout:` + `if: false` skips the step (no timeout applied)
- [x] `timeout:` is invalid on `parent_shell: true` steps (parse-time error)
- [x] `timeout:` is invalid on `run:` steps (parse-time error)
- [x] `pi info` shows `[timeout: <value>]` annotation on steps with timeout
- [x] Unit tests for parsing (valid/invalid durations)
- [x] Unit tests for execution (timeout triggers, no timeout doesn't interfere)
- [x] Integration tests with example workspace
- [x] Architecture docs updated
- [x] All existing tests pass

## Implementation Notes

### Design Decisions
- Use Go `time.Duration` for the parsed value, stored as `time.Duration` on the `Step` struct
- YAML field is a string, parsed via `time.ParseDuration()` during YAML unmarshalling
- Timeout enforcement uses `context.WithTimeout` wrapping `exec.CommandContext`
- On timeout, the process is killed via the context cancellation mechanism
- Exit code 124 matches the GNU `timeout` command convention for consistency
- `timeout:` is invalid on `run:` steps because run steps are recursive executor calls, not direct process executions — the child automation's steps have their own timeouts
- `timeout:` is invalid on `parent_shell:` steps because they don't execute as subprocesses

### Implementation approach
- `Step` struct: add `Timeout time.Duration` field
- `stepRaw`: add `Timeout string` yaml field, parse with `time.ParseDuration` in `toStep()`
- `runStepCommand()` in `runners.go`: when timeout is set, use `exec.CommandContext` with a deadline
- `pi info`: add timeout annotation in `printStepsDetail()`
- Validation: negative/zero timeouts rejected at parse time

## Subtasks
- [x] Add Timeout field to Step and stepRaw, parse in toStep()
- [x] Validate: reject on run: and parent_shell: steps, reject non-positive durations
- [x] Enforce timeout in runStepCommand() using exec.CommandContext
- [x] Add pi info annotation
- [x] Add unit tests (parsing + execution)
- [x] Add integration test workspace + tests
- [x] Update architecture.md and docs/README.md
- [x] QA pass

## Blocked By
