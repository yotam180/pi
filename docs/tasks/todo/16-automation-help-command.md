# Automation Help Command

## Type
feature

## Status
todo

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
- [ ] `pi info <name>` (or equivalent) prints the automation's name, description, and input docs
- [ ] Required inputs are clearly distinguished from optional ones
- [ ] Default values are shown
- [ ] Works for automations with and without inputs
- [ ] Error message for unknown automation name

## Implementation Notes

## Subtasks
- [ ] Add `newInfoCmd()` or extend run command with `--info` flag
- [ ] Format and print automation details
- [ ] Write unit and integration tests

## Blocked By
