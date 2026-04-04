# Conditional Execution Integration Testing & QA

## Type
improvement

## Status
todo

## Priority
medium

## Project
04-conditional-step-execution

## Description
Create a comprehensive example workspace and integration test suite that exercises all aspects of conditional step execution end-to-end. This is the final QA pass for Project 04.

Create `examples/conditional/` workspace with automations that use `if:` at step, automation, and setup levels, covering all predicate types and boolean operators. Write integration tests that run the binary against this workspace and verify correct skip/execute behavior.

Also update `pi info` to show `if:` conditions on steps and automations, and update `pi list` if needed.

## Acceptance Criteria
- [ ] `examples/conditional/` workspace exists with representative automations
- [ ] Integration tests cover: step-level if, automation-level if, setup entry if, skipped steps, various predicates
- [ ] `pi info <name>` shows the `if:` condition when present (on automation and on steps)
- [ ] All existing tests still pass (no regressions)
- [ ] Architecture doc updated with conditional execution details
- [ ] README updated if needed

## Implementation Notes

## Subtasks
- [ ] Create `examples/conditional/` workspace
- [ ] Write integration tests
- [ ] Update `pi info` to display `if:` conditions
- [ ] Update docs (architecture.md, README.md)
- [ ] Full regression run

## Blocked By
19-if-on-steps
20-if-on-automations
21-if-on-setup-entries
