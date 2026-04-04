# `if:` Field on Steps

## Type
feature

## Status
done

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
- [x] `if:` field is parsed from automation YAML on steps
- [x] Steps with `if:` evaluating to false are silently skipped
- [x] Steps with `if:` evaluating to true execute normally
- [x] Steps without `if:` always execute (backward compatible)
- [x] Invalid `if:` expressions produce a clear parse error
- [x] `pipe_to: next` on a skipped step doesn't break the pipeline
- [x] Unit tests in automation package for parsing
- [x] Unit tests in executor package for conditional execution
- [x] Integration test with example automation using `if:`

## Implementation Notes

### Schema changes
- Added `If string` field to both `Step` and `stepRaw` structs in `internal/automation/automation.go`
- `stepRaw.If` is parsed from YAML `if:` tag, then copied to `Step.If` in `toStep()`
- Validation at load time: `conditions.Predicates()` is called to parse the expression, producing a clear error on invalid syntax

### Executor changes
- Added `evaluateCondition(expr string) (bool, error)` method to `Executor`
- Returns true when the step should be skipped (condition evaluates to false)
- Uses `ResolvePredicatesWithEnv()` with the executor's `RuntimeEnv` field (or `DefaultRuntimeEnv()` if nil)
- Added `RuntimeEnv *RuntimeEnv` field to `Executor` for testability — tests inject a fake env

### Pipe passthrough for skipped steps
- When a skipped step has `pipe_to: next`, any existing piped input from a prior step is passed through to the next step
- If the skipped step is the first in a pipe chain (no prior piped input), `pipedInput` stays nil
- Multiple consecutive skipped pipe steps correctly pass data through

### Tests added
- **Automation parsing (6 tests)**: basic if, no if, complex expressions, if+pipe_to, invalid if, function-call if
- **Executor unit (13 tests)**: true/false conditions, no-if backward compat, all-skipped, not operator, complex and/or expressions, mixed conditional/unconditional, pipe passthrough (single skip, no prior pipe, multiple skips), file.exists predicate
- **Integration (4 tests)**: list, platform-info (OS-aware), skip-all, pipe-conditional

### Example workspace
- Created `examples/conditional/` with 3 automations: `platform-info`, `skip-all`, `pipe-conditional`

## Subtasks
- [x] Add `If` field to `Step` and `stepRaw` in automation package
- [x] Add condition evaluation to executor before step dispatch
- [x] Handle pipe_to edge case for skipped steps
- [x] Write automation parsing tests
- [x] Write executor unit tests
- [x] Create example automation using `if:` and write integration test

## Blocked By
17-condition-expression-parser
18-predicate-resolution
