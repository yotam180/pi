# `if:` Field on Steps

## Type
feature

## Status
todo

## Priority
high

## Project
04-conditional-step-execution

## Description
Add an `if:` field to the `Step` struct in the automation YAML schema, and wire it into the executor so that steps whose condition evaluates to false are silently skipped.

Changes needed:
1. **Schema**: Add `If string` field to `Step` and `stepRaw` structs in `internal/automation/automation.go`
2. **Executor**: Before executing a step, if `step.If != ""`:
   - Extract predicates from the expression via `conditions.Predicates()`
   - Resolve predicates via the predicate resolver
   - Evaluate via `conditions.Eval()`
   - If false: skip silently (no output, no error), continue to next step
   - If true: execute normally
3. **Pipe handling**: If a step with `pipe_to: next` is skipped, the piped input should pass through to the step after (or be discarded if the next step is also skipped)

## Acceptance Criteria
- [ ] `if:` field is parsed from automation YAML on steps
- [ ] Steps with `if:` evaluating to false are silently skipped
- [ ] Steps with `if:` evaluating to true execute normally
- [ ] Steps without `if:` always execute (backward compatible)
- [ ] Invalid `if:` expressions produce a clear parse error
- [ ] `pipe_to: next` on a skipped step doesn't break the pipeline
- [ ] Unit tests in automation package for parsing
- [ ] Unit tests in executor package for conditional execution
- [ ] Integration test with example automation using `if:`

## Implementation Notes

## Subtasks
- [ ] Add `If` field to `Step` and `stepRaw` in automation package
- [ ] Add condition evaluation to executor before step dispatch
- [ ] Handle pipe_to edge case for skipped steps
- [ ] Write automation parsing tests
- [ ] Write executor unit tests
- [ ] Create example automation using `if:` and write integration test

## Blocked By
17-condition-expression-parser
18-predicate-resolution
