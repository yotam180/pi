# Step trace lines show raw variable names instead of resolved values

## Type
bug

## Status
done

## Priority
medium

## Project
standalone

## Description
When running an automation with inputs, the step trace line shows the raw environment variable reference instead of the resolved value.

### Steps to Reproduce
1. Create an automation with an input:
```yaml
name: build
inputs:
  profile:
    type: string
    default: dev
steps:
  - bash: cargo build --profile $PI_IN_PROFILE
```
2. Run: `pi run build --with profile=release`

### Expected
```
  → bash: cargo build --profile release
```

### Actual (before fix)
```
  → bash: cargo build --profile $PI_IN_PROFILE
```

The command DOES execute correctly (release profile is used), but the trace output shows the unexpanded template.

## Acceptance Criteria
- [x] Step trace lines show resolved variable values, not raw `$PI_IN_*` references
- [x] Complex commands with multiple variables all show resolved values
- [x] Non-input env vars (automation-level and step-level) are also expanded in trace

## Implementation Notes

### Approach
Added `expandTraceVars()` function in `internal/executor/helpers.go` that uses `os.Expand()` to resolve environment variable references (`$VAR` and `${VAR}`) in step values before they are printed as trace lines. The expansion uses a combined map built from:
1. `inputEnv` (`PI_IN_*` and `PI_INPUT_*` variables)
2. `automationEnv` (automation-level `env:` block)
3. `stepEnv` (step-level `env:` block)

Step-level env overrides automation-level env for the same key, matching execution semantics. Unknown variables (like `$HOME`) pass through unchanged — they are resolved by the shell at execution time, not by PI.

### Changes
- `internal/executor/helpers.go`: Added `expandTraceVars(value, inputEnv, automationEnv, stepEnv)` using `os.Expand()` with a lookup function that falls back to `$KEY` for unknown vars.
- `internal/executor/executor.go`: Updated all three `StepTrace` call sites (regular steps, first: sub-steps, parent_shell steps) to call `expandTraceVars()` before printing. Updated `execParentShell()` signature to receive `*automation.Automation` and `inputEnv` for env context.

### Tests added
- 11 unit tests for `expandTraceVars()` covering: no variables, single/multiple input vars, braced syntax, automation env, step env, step-overrides-automation priority, unknown vars passthrough, mixed known/unknown, deprecated PI_INPUT_ prefix.
- 5 integration-style executor tests: input var expansion in trace, automation env expansion, step env expansion, multiple input vars, first: block var expansion.

### Key decisions
- Used `os.Expand()` from the standard library for reliable `$VAR` / `${VAR}` parsing rather than regex.
- Expansion is trace-only — the actual command execution is unchanged and still uses real shell env vars.
- Unknown variables produce `$KEY` in the output (not empty string), so traces show what the shell will resolve at runtime.

## Subtasks
- [x] Add expandTraceVars helper function
- [x] Update StepTrace call sites in executor
- [x] Add unit tests for expandTraceVars
- [x] Add integration tests for trace expansion
- [x] Run full test suite

## Blocked By
