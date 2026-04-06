# Outputs Interpolation Integration Tests

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The outputs interpolation feature (`outputs.last`, `outputs.<N>`, `inputs.<name>` in `with:` and `env:` values) has extensive unit test coverage (30 tests in `executor/outputs_test.go`) but **zero integration tests** via the actual binary. There's also no example project demonstrating this feature.

Integration tests are essential here because the feature involves wiring between:
1. The CLI layer (argument parsing, executor construction)
2. The executor's interpolation engine (`interpolateValue`, `interpolateEnv`, `interpolateWithCtx`)
3. The step runner layer (`RunContext.InterpolateWith`, `ResolvedAutomationEnv`, `ResolvedStepEnv`)

If any of this wiring breaks (e.g., during a refactor of the executor or CLI), only integration tests would catch it.

**What to build:**
1. An `examples/outputs/` project with automations exercising all interpolation variants
2. Integration tests in `tests/integration/outputs_test.go` that run the pi binary and verify correct behavior

**Scenarios to cover:**
- `outputs.last` in `with:` values (run step passes previous step output to sub-automation)
- `outputs.last` in `env:` values (step-level env references previous step output)
- `outputs.<N>` indexed access in `with:` and `env:`
- `inputs.<name>` in `with:` and `env:` values
- `outputs.last` in automation-level `env:`
- Mixed interpolation (literal + outputs + inputs in same automation)
- Pipe + output capture interaction (pipe: true still records outputs)

## Acceptance Criteria
- [x] `examples/outputs/` project exists with working automations
- [x] `tests/integration/outputs_test.go` covers all scenarios listed above
- [x] All tests pass (`go test ./...`)
- [x] `go build ./...` passes
- [x] Architecture doc updated with new test file description

## Implementation Notes

### Approach
Created 10 automation YAML files in `examples/outputs/.pi/` covering all interpolation variants, plus 13 integration tests in `tests/integration/outputs_test.go`.

### Example automations created
- `chain-with.yaml` — outputs.last passed via `with:` to a sub-automation
- `chain-env.yaml` — outputs.last referenced in step-level `env:`
- `indexed.yaml` — outputs.0 and outputs.1 indexed access in `env:`
- `inputs-in-with.yaml` — inputs.msg forwarded via `with:` to sub-automation
- `inputs-in-env.yaml` — inputs.tag referenced in step-level `env:`
- `auto-env-inputs.yaml` — inputs.version in automation-level `env:`
- `mixed.yaml` — combines outputs, inputs, and literal strings
- `pipe-and-capture.yaml` — pipe: true + indexed output access
- `three-step-chain.yaml` — chained outputs.last through 3 steps
- `echo-input.yaml` — helper automation used by other automations

### Test strategy
Used `runPiStdout` (stdout-only capture) for output assertions since trace lines go to stderr. This avoids false positives from trace line matching. Used `runPi` (combined) for list/validate/info assertions.

### Key findings
All interpolation paths work correctly end-to-end. The executor trace lines show interpolated env values (e.g., `→ bash: echo "tag=v1.0"` instead of raw `outputs.last`), confirming that `expandTraceVars` properly resolves both regular and interpolated env values for display.

## Subtasks
- [x] Create `examples/outputs/` project structure
- [x] Write automation YAML files for each scenario
- [x] Write integration tests
- [x] Run tests, iterate
- [x] Update architecture.md

## Blocked By
