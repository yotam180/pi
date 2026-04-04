# `if:` Field on Automations

## Type
feature

## Status
todo

## Priority
medium

## Project
04-conditional-step-execution

## Description
Add an `if:` field at the automation level. When an automation's top-level `if:` evaluates to false, the entire automation is skipped — no steps execute. Unlike step-level skipping (which is silent), skipping an entire automation should print a one-line notice: `[skipped] <name> (condition: <expr>)`.

This is important for built-in automations like `pi:install-homebrew` which should only run on macOS — the condition is at the automation level, not repeated on every step.

Changes needed:
1. **Schema**: Add `If string` field to `Automation` struct in `internal/automation/automation.go`, parsed from YAML
2. **Executor**: Before running any step, check the automation's `if:` condition. If false, print skip notice to stderr, return nil (no error)
3. **`pi run`**: If a directly-invoked automation is skipped, the skip message should be visible
4. **`run:` steps**: If a called automation is skipped via `if:`, the parent step succeeds silently

## Acceptance Criteria
- [ ] `if:` field is parsed from automation YAML at the top level
- [ ] Automation with `if:` evaluating to false is fully skipped (no steps run)
- [ ] Skip message printed to stderr: `[skipped] <name> (condition: <expr>)`
- [ ] Automation without `if:` always runs (backward compatible)
- [ ] `run:` step calling a skipped automation succeeds without error
- [ ] Unit tests for parsing and execution
- [ ] Integration test

## Implementation Notes

## Subtasks
- [ ] Add `If` field to `Automation` struct
- [ ] Add automation-level condition check in executor
- [ ] Handle `run:` step calling a conditionally-skipped automation
- [ ] Write tests

## Blocked By
19-if-on-steps
