# Deduplicate Step Validation and Improve Automation Package Coverage

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description
The `automation` package had three functions that performed nearly identical step validation logic:

1. **`validateSteps()`** — iterates steps, checks empty value, validates step type, validates `if:` expression
2. **`validateFirstBlock()`** — iterates first: sub-steps, checks empty value, validates step type, validates `if:` expression, plus checks for nested first: blocks
3. **`validateInstallPhase()`** — iterates install phase steps, checks empty value, validates step type, validates `if:` expression, plus handles first: blocks

This duplication meant any new step-level validation would need to be added in three places. The fix was to extract a shared `validateSingleStep()` helper that validates one step's invariants, then have all three call it.

Additionally, the automation package was at 89.5% coverage with several functions well below 90%.

## Acceptance Criteria
- [x] Shared `validateSingleStep()` helper extracted; `validateSteps()`, `validateFirstBlock()`, and `validateInstallPhase()` use it
- [x] No behavioral change — all existing tests pass unchanged
- [x] Coverage for `validateFirstBlock` ≥ 90% (achieved: 100%)
- [x] Coverage for `validateInstall` ≥ 90% (achieved: 93.8%)
- [x] Coverage for `toFirstStep` ≥ 90% (achieved: 92.3%)
- [x] Coverage for `validateInstallPhase` ≥ 90% (achieved: 100%)
- [x] Coverage for `WalkStepsUntil` ≥ 85% (achieved: 100%)
- [x] Coverage for `walkInstallPhaseUntil` ≥ 85% (achieved: 100%)
- [x] Overall automation package coverage ≥ 92% (achieved: 94.2%)
- [x] `go build ./...` and `go test ./...` pass
- [x] `architecture.md` updated

## Implementation Notes

### Approach
1. Extracted `validateSingleStep(prefix, step)` — validates non-empty value, valid step type, and parseable if: expression. Takes a prefix string for error messages so callers control the error context.
2. Rewrote `validateSteps()`, `validateFirstBlock()`, and `validateInstallPhase()` to delegate to `validateSingleStep()` instead of inlining the same checks.
3. Added 18 new step_test.go tests covering first: block error paths (timeout/silent/parent_shell/with on block, sub-step invalid if/empty value, outer invalid if), install phase error paths (empty step-list test/run, verify invalid if, run empty step, run first-block invalid if), and validateSingleStep unit tests.
4. Added 8 new walker_test.go tests covering WalkStepsUntil stops in run/verify/test phases, step-list phase with first: block, completes all install phases, no install.

### Coverage improvements
| Function | Before | After |
|---|---|---|
| `validateSingleStep` (new) | N/A | 100.0% |
| `validateSteps` | 92.3% | 100.0% |
| `validateFirstBlock` | 71.4% | 100.0% |
| `validateInstall` | 75.0% | 93.8% |
| `validateInstallPhase` | 81.2% | 100.0% |
| `toFirstStep` | 76.9% | 92.3% |
| `WalkStepsUntil` | 60.0% | 100.0% |
| `walkInstallPhaseUntil` | 66.7% | 100.0% |
| **Overall package** | **89.5%** | **94.2%** |

### Remaining gaps
- `validateInstall` 93.8%: the remaining uncovered line is the verify phase validation error path when verify is nil (unreachable — the nil check guards it)
- `toFirstStep` 92.3%: the remaining uncovered line is the pipe resolution error path (only triggers with both `pipe: true` and `pipe_to: next` on a first: block, which is already tested via `toStep`)

## Subtasks
- [x] Extract `validateSingleStep()` helper
- [x] Rewrite `validateSteps()`, `validateFirstBlock()`, `validateInstallPhase()` to use it
- [x] Add tests for uncovered error paths
- [x] Add tests for walker coverage gaps
- [x] Verify build + tests + coverage
- [x] Update architecture.md

## Blocked By
