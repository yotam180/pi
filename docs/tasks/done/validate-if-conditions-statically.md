# Static Validation of if: Conditions in pi validate

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

`pi validate` performs extensive static analysis — checking automation references, file paths, input mismatches, and circular dependencies — but it does **not** validate `if:` conditions on automation steps or automation-level `if:` fields. Currently, invalid `if:` expressions (syntax errors or unknown predicate names like `os.macoss`) are only caught at runtime when the automation is actually executed.

Since `pi validate` already walks all automations via `discovery.Result` and `automation.WalkSteps`, it should statically check that:
1. All `if:` expressions parse correctly (syntax validation via `conditions.Predicates()`)
2. All predicate names in `if:` expressions are recognized (pattern validation — `os.macos`, `command.<x>`, `env.<X>`, etc.)

This catches typos and syntax errors at validate time instead of at runtime.

**Locations of if: fields to validate:**
- Automation-level `if:` (on the automation itself)
- Step-level `if:` (on each step in steps: lists)
- Sub-step `if:` within `first:` blocks
- Install phase step `if:` conditions

**What "valid predicate name" means (static check, no resolution):**
- Exact known names: `os.macos`, `os.linux`, `os.windows`, `os.arch.arm64`, `os.arch.amd64`, `shell.zsh`, `shell.bash`
- Known prefixed patterns: `env.<NAME>` (non-empty name), `command.<name>` (non-empty name)
- Known function-call patterns: `file.exists("<path>")`, `dir.exists("<path>")`

## Acceptance Criteria
- [x] `pi validate` checks all `if:` expressions on automations, steps, first: sub-steps, and install phase steps
- [x] Syntax errors in `if:` expressions are reported (by automation loader at parse time)
- [x] Unknown predicate names (e.g., `os.macoss`) are reported
- [x] Valid predicates with dynamic suffixes (e.g., `command.docker`, `env.MY_VAR`) pass validation
- [x] Setup entry `if:` expressions already validated by config.Load() are not double-checked
- [x] Tests cover: valid conditions, syntax errors, unknown predicates, dynamic predicates, all step locations
- [x] All existing tests pass
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach

Added two layers of validation:

1. **`ValidatePredicateName(name string) error`** in `executor/predicates.go` — checks that a predicate name matches one of the known exact names (`os.macos`, `shell.zsh`, etc.) or known prefix patterns (`env.<NAME>`, `command.<name>`, `file.exists(...)`, `dir.exists(...)`). Returns an error with a helpful message listing valid predicates if the name is unrecognized.

2. **`ValidateConditionExpr(expr string) error`** in `executor/predicates.go` — parses an if: expression using `conditions.Predicates()` (which handles syntax validation) and then validates each extracted predicate name via `ValidatePredicateName`. This combines syntax + semantic validation in one call.

3. **`validateConditions(disc, result)`** in `cli/validate.go` — walks all discovered automations checking both automation-level `if:` and all step-level `if:` (including first: sub-steps and install phases) via `automation.WalkSteps`.

### Key design decisions

- **No duplication with automation.validate()**: The automation YAML loader already validates `if:` **syntax** at parse time via `conditions.Predicates()`. Our new validation adds **semantic** checking (unknown predicate names) which the parser cannot do.
- **No import cycle risk**: `ValidateConditionExpr` lives in `executor` and imports `conditions`. The `cli` package imports both, maintaining the existing dependency direction.
- **Static-only**: `ValidatePredicateName` checks structural validity without resolving runtime values. `command.docker` passes validation regardless of whether docker is installed — that's a runtime concern.
- **knownExactPredicates**: Extracted to a package-level map to keep the validation function clean and make it easy to add new predicates.

### Tests added

**`executor/predicates_test.go`**:
- `TestValidatePredicateName` — 25 subtests (17 valid, 8 invalid with specific error messages)
- `TestValidateConditionExpr` — 8 subtests (5 valid, 3 invalid)

**`cli/validate_test.go`** (9 new tests):
- `TestValidate_ValidConditions` — valid automation-level and step-level conditions pass
- `TestValidate_ConditionSyntaxErrorCaughtAtLoadTime` — confirms syntax errors caught by automation loader
- `TestValidate_UnknownPredicateInCondition` — step-level unknown predicate detected
- `TestValidate_UnknownPredicateOnAutomation` — automation-level unknown predicate detected
- `TestValidate_ConditionInFirstBlock` — unknown predicate in first: sub-step detected
- `TestValidate_DynamicPredicatesPass` — command.*, env.*, file.exists(), dir.exists() pass
- `TestValidate_MultipleConditionErrors` — all errors collected and counted
- `TestValidate_ConditionInInstallPhase` — install phase step conditions validated
- `TestValidate_ConditionWithOtherErrors` — condition errors reported alongside other validation errors

### Coverage
- `executor`: 93.1% → 93.3%
- `cli`: 80.6% → 81.1%

## Subtasks
- [x] Add `ValidatePredicateName()` function for static predicate name checking
- [x] Add `ValidateConditionExpr()` that combines parsing + predicate validation
- [x] Add `validateConditions()` to `cli/validate.go`
- [x] Wire into `validateProject()`
- [x] Write unit tests (predicates_test.go)
- [x] Write integration tests (validate_test.go)
- [x] Update architecture.md

## Blocked By
