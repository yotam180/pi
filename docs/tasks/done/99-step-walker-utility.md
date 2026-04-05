# Extract Step Walker Utility for Automation Traversal

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

Multiple consumers in the codebase walk the same automation structure — iterating over steps, descending into `first:` blocks, and visiting all install phases (test, run, verify). Currently this traversal logic is duplicated in:

1. `cli/validate.go` — `validateRunStepRefs` walks steps + first: + install phases checking run: references
2. `cli/validate.go` — `validateFileReferences` walks the same structure checking file-path references
3. Future validation checks (input references, condition expression validation, etc.) would duplicate again

The fix: extract a `StepWalker` utility in `internal/automation` that provides a visitor-pattern traversal of all steps in an automation (including first: sub-steps and install phase steps). Consumers provide a callback and the walker handles the traversal.

This improves:
- **Code quality**: eliminates duplicate traversal code
- **Modularity**: traversal lives in one tested place
- **Expandability**: adding new analysis/validation passes becomes trivial — just provide a visitor function

## Acceptance Criteria
- [x] `internal/automation` exports a `WalkSteps` function that visits all steps in an automation
- [x] Each visited step includes context: its location (step index, first: sub-index, install phase name)
- [x] WalkSteps covers: regular steps, first: block sub-steps, install phases (test, run, verify — including scalar phases)
- [x] `cli/validate.go` is refactored to use `WalkSteps` for both run-step-ref and file-ref validation
- [x] All existing tests pass
- [x] New unit tests for WalkSteps in `internal/automation/walker_test.go`
- [x] `docs/architecture.md` updated

## Implementation Notes

### Design decisions

**Function signature**: `WalkSteps(a *Automation, fn StepVisitor)` where `StepVisitor = func(step Step, loc StepLocation)`. The walker returns no error — the visitor captures errors externally if needed (keeps the walker pure). If the visitor needs to signal "stop walking", we use a `WalkStepsUntil` variant that takes `func(...) bool` (return true to stop).

**StepLocation struct**: encodes where in the automation the step was found:
- `Phase` — "" for regular steps, "test"/"run"/"verify" for install phases
- `Index` — step index (0-based)
- `FirstIndex` — -1 for regular steps, sub-step index for first: blocks
- `IsScalar` — true for scalar install phases (the generated step)

This lets callers like `validate.go` format precise error locations without reimplementing the traversal.

**Scalar install phases**: The walker generates a synthetic `Step{Type: StepTypeBash, Value: phase.Scalar}` for scalar phases and passes it through with `IsScalar: true`. This means consumers don't need to handle scalar phases separately.

## Subtasks
- [x] Design StepLocation and WalkSteps API
- [x] Implement WalkSteps in internal/automation/walker.go
- [x] Write unit tests in internal/automation/walker_test.go (17 tests)
- [x] Refactor cli/validate.go to use WalkSteps
- [x] Run full test suite (all 1405 tests pass)
- [x] Update docs/architecture.md

## Blocked By
