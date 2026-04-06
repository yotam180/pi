# Extract Condition Evaluation into Standalone Evaluator

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The condition evaluation logic (parse predicates → resolve runtime values → evaluate boolean expression) is currently split across two locations:

1. `executor.evaluateCondition()` — a method on Executor that resolves predicates, evaluates conditions, and returns a "skip" boolean
2. `cli/setup.go:evaluateSetupCondition()` — a standalone function that duplicates the same 7-line pattern

Additionally, the `RuntimeEnv` struct and predicate resolution functions (`ResolvePredicates`, `ResolvePredicatesWithEnv`, `ValidatePredicateName`, `ValidateConditionExpr`) live in the `executor` package despite being conceptually part of the conditions system. This creates unnecessary coupling:

- The `conditions` package handles parsing and AST evaluation
- The `executor` package handles predicate resolution (which is really a conditions concern)
- The `executor` package defines `RuntimeEnv` (which is about condition predicates, not step execution)

**Goal:** Create a `conditions.Evaluator` type that encapsulates the full condition lifecycle: parse → resolve predicates → evaluate. Move `RuntimeEnv` and predicate resolution from executor to conditions. Eliminate the duplicated evaluation pattern.

## Acceptance Criteria
- [x] `RuntimeEnv` struct moved from executor to conditions package
- [x] `ResolvePredicates`, `ResolvePredicatesWithEnv` moved to conditions
- [x] `ValidatePredicateName`, `ValidateConditionExpr` moved to conditions
- [x] New `conditions.Evaluator` type that wraps the parse→resolve→eval lifecycle
- [x] `Executor.evaluateCondition()` delegates to `conditions.Evaluator`
- [x] `cli/setup.go:evaluateSetupCondition()` replaced with `conditions.Evaluator`
- [x] All existing tests pass
- [x] Predicate resolution tests moved to conditions package
- [x] No circular imports
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach

Create a new file `internal/conditions/evaluator.go` containing:
- `RuntimeEnv` (moved from executor/predicates.go)
- `DefaultRuntimeEnv()` (moved from executor/predicates.go)
- `Evaluator` struct holding a `*RuntimeEnv` and `repoRoot`
- `NewEvaluator(repoRoot string, env *RuntimeEnv)` constructor
- `(ev *Evaluator) ShouldSkip(expr string) (bool, error)` — the full lifecycle
- Predicate resolution functions (moved from executor/predicates.go)
- Predicate validation functions (moved from executor/predicates.go)

The executor package keeps thin type aliases/re-exports during the transition to avoid breaking external references.

### Design decisions
- The `Evaluator` is a lightweight struct (repoRoot + RuntimeEnv pointer), not a heavy service
- `ShouldSkip` name matches the usage pattern: "should this step/automation be skipped?"
- Keep `Eval` and `Predicates` on the conditions package as they are (public API)
- `RuntimeEnv` belongs in conditions because it's about resolving condition predicates

## Subtasks
- [x] Create task file
- [x] Move RuntimeEnv + predicate resolution to conditions package
- [x] Create Evaluator type
- [x] Update executor to use Evaluator
- [x] Update cli/setup.go to use Evaluator
- [x] Update cli/validate.go and doctor.go to use conditions directly
- [x] Move tests
- [x] Verify all tests pass (16/16 packages)
- [x] Update architecture.md

## Blocked By
