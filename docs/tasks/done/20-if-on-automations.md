# `if:` Field on Automations

## Type
feature

## Status
done

## Priority
medium

## Project
04-conditional-step-execution

## Description
Add an `if:` field at the automation level. When an automation's top-level `if:` evaluates to false, the entire automation is skipped — no steps execute. Unlike step-level skipping (which is silent), skipping an entire automation prints a one-line notice: `[skipped] <name> (condition: <expr>)`.

## Acceptance Criteria
- [x] `if:` field is parsed from automation YAML at the top level
- [x] Automation with `if:` evaluating to false is fully skipped (no steps run)
- [x] Skip message printed to stderr: `[skipped] <name> (condition: <expr>)`
- [x] Automation without `if:` always runs (backward compatible)
- [x] `run:` step calling a skipped automation succeeds without error
- [x] Unit tests for parsing and execution
- [x] Integration test

## Implementation Notes
- Added `If string` field to `Automation` struct in `internal/automation/automation.go`
- Added `If` parsing in `UnmarshalYAML` via the raw struct
- Added validation in `validate()` — invalid `if:` expressions are caught at YAML load time
- Added automation-level condition check at the start of `RunWithInputs()` in `internal/executor/executor.go`, before `pushCall()` — this means skipped automations don't consume call stack slots, which naturally prevents false circular-dependency errors
- Skip message is printed to stderr, and the method returns nil (no error) — this means `run:` steps calling a skipped automation succeed silently

### Tests added
- **automation_test.go**: 7 new tests for parsing `if:` on automations (with/without, complex, invalid, func-call, combined with step-level if)
- **executor_test.go**: 7 new tests for execution behavior (true executes, false skips with message, no-if always runs, run-step calls skipped, run-step calls executed, complex condition, skip-prevents-circular-dep)
- **Integration tests**: 4 new tests (list discovers new automations, impossible automation skipped, macOS-only OS-aware, run-step calling skipped automation)
- **Example automations**: 3 new files in `examples/conditional/.pi/` (macos-only, impossible, call-conditional)

## Subtasks
- [x] Add `If` field to `Automation` struct
- [x] Add automation-level condition check in executor
- [x] Handle `run:` step calling a conditionally-skipped automation
- [x] Write tests

## Blocked By
19-if-on-steps
