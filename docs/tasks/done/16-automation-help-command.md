# Automation Help Command

## Type
feature

## Status
done

## Priority
low

## Project
standalone

## Description
Implement a way to view an automation's description and input documentation from the CLI. Either `pi run --help <name>` or `pi info <name>` should print:
- The automation's name and description
- All declared inputs with their type, required/optional status, default value, and description

This was deferred from task 07-automation-inputs-schema.

## Acceptance Criteria
- [x] `pi info <name>` (or equivalent) prints the automation's name, description, and input docs
- [x] Required inputs are clearly distinguished from optional ones
- [x] Default values are shown
- [x] Works for automations with and without inputs
- [x] Error message for unknown automation name

## Implementation Notes

Chose `pi info <name>` as a dedicated subcommand (vs `--info` flag on run) because:
- Cleaner UX — info is a read-only inspect operation, run is an execution
- Cobra `ExactArgs(1)` provides automatic validation
- Follows the same discover-find pattern as `run` and `list`

Files added/changed:
- `internal/cli/info.go` — `newInfoCmd()` + `showAutomationInfo()` (testable helper) + `printAutomationInfo()` + `printInputSpec()`
- `internal/cli/info_test.go` — 9 unit tests covering: simple, no-description, with-inputs, required vs optional, multi-step, not-found, no-pi-yaml, subdirectory, default values
- `internal/cli/root.go` — wired `newInfoCmd()` into the root command
- `tests/integration/examples_test.go` — 4 integration tests: basic automation, with-inputs, not-found, no-args

Output format:
```
Name:         greet
Description:  Greet someone
Steps:        1

Inputs:
  name (string, required)
      Who to greet
  greeting (string, optional, default: "hello")
      The greeting word
```

## Subtasks
- [x] Add `newInfoCmd()` subcommand
- [x] Format and print automation details (name, description, steps, inputs)
- [x] Write unit tests (9 tests)
- [x] Write integration tests (4 tests)

## Blocked By
