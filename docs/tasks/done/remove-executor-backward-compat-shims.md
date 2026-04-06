# Remove executor backward-compatibility shims

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description
The executor package contains two files of backward-compatibility shims that are no longer used by any external code:

1. **`predicates.go`** — 5 exported functions (`DefaultRuntimeEnv`, `ResolvePredicates`, `ResolvePredicatesWithEnv`, `ValidatePredicateName`, `ValidateConditionExpr`) that delegate 1:1 to the `conditions` package. Plus a type alias `RuntimeEnv = conditions.RuntimeEnv`. Only used by `predicates_test.go` within the same package.

2. **`validate.go`** (lower half) — 7 unexported wrapper functions (`checkRequirement`, `runtimeCommand`, `detectVersion`, `extractVersion`, `compareVersions`, `formatRequirementLabel`, `installHint`) that delegate 1:1 to the `reqcheck` package. Plus 3 exported type aliases (`CheckResult`, `ValidationError`, `FormatValidationError`) and 2 exported var aliases (`InstallHintFor`, `CheckRequirementForDoctor`) — none used by any Go code outside the executor package.

These shims were kept during earlier extraction refactors for backward compatibility, but no external consumers remain. Removing them:
- Reduces dead code surface area
- Makes the canonical package ownership clearer (conditions owns predicates, reqcheck owns requirement checking)
- Removes confusing re-exports that could mislead future developers

## Acceptance Criteria
- [x] `predicates.go` deleted entirely
- [x] `predicates_test.go` updated to import and call `conditions` package directly
- [x] Exported type aliases and var aliases removed from `validate.go`
- [x] Unexported wrapper functions removed from `validate.go`
- [x] `validate_test.go` and `coverage_gaps_test.go` updated to import and call `reqcheck`/`conditions` directly
- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] No coverage regression (executor package stays ≥ 95%)
- [x] `architecture.md` updated to reflect the cleanup

## Implementation Notes
### Analysis
- `predicates.go` exports: `RuntimeEnv` (type alias), `DefaultRuntimeEnv()`, `ResolvePredicates()`, `ResolvePredicatesWithEnv()`, `ValidatePredicateName()`, `ValidateConditionExpr()` — all delegate to `conditions` package. No Go code outside executor uses any of these.
- `validate.go` exports: `CheckResult` (type alias), `ValidationError` (type alias), `FormatValidationError` (var), `InstallHintFor` (var), `CheckRequirementForDoctor` (var) — no Go code outside executor uses any of these.
- `validate.go` unexported: `checkRequirement`, `runtimeCommand`, `detectVersion`, `extractVersion`, `compareVersions`, `formatRequirementLabel`, `installHint` — only used by `validate_test.go` and `coverage_gaps_test.go`.
- The `RuntimeEnv` type alias is used by `Executor.RuntimeEnv` field and `ValidateRequirements` method — these need to reference `conditions.RuntimeEnv` directly after cleanup.
- `ExitError` is defined in `executor.go`, not in `validate.go` — no conflict.

### Approach
1. Delete `predicates.go` entirely
2. Update `predicates_test.go` to import `conditions` and call functions directly
3. Strip `validate.go` down to just `ValidateRequirements()` and `tryProvision()` (the actual executor logic)
4. Update `validate_test.go` and `coverage_gaps_test.go` to call `reqcheck.*` directly
5. Update `executor.go` to reference `conditions.RuntimeEnv` directly (the type alias removal)

## Subtasks
- [x] Delete `predicates.go`
- [x] Rewrite `predicates_test.go` → import conditions
- [x] Strip `validate.go` shims
- [x] Update `validate_test.go` → import reqcheck
- [x] Update `coverage_gaps_test.go` → import reqcheck
- [x] Fix `executor.go` RuntimeEnv references
- [x] Build + test + verify coverage
- [x] Update architecture.md

## Blocked By
