# Allow outputs.last / outputs.N in Step env: Values

## Type
feature

## Status
done

## Priority
medium

## Project
standalone

## Description

`outputs.last` and `outputs.<N>` interpolation currently only works inside `with:` values on `run:` steps. It cannot be used in a step's `env:` map or in automation-level `env:`. This is inconsistent — the interpolation engine exists and works, it's just not wired into env evaluation. This task extends the interpolation surface so all `env:` values support the same references.

**Current (only works):**
```yaml
steps:
  - bash: cat version.txt
  - run: pi:version-satisfies
    with:
      version: outputs.last   # ← only here
```

**After this task (should also work):**
```yaml
steps:
  - bash: cat version.txt
  - bash: echo "Checking version $MY_VERSION"
    env:
      MY_VERSION: outputs.last   # ← new
```

**Implementation:** `interpolateValue` already handles the interpolation logic. The change is to also call it when building step environment variables in `BuildStepEnv`. The interpolation needs access to the executor's current `stepOutputs` and `inputEnv` — so it needs to happen inside the executor, not inside `BuildStepEnv` directly.

Best approach: before constructing the `RunContext`, apply `interpolateValue` to each `env:` map entry if it contains `outputs.` or `inputs.` prefixes. This is analogous to how `with:` values are interpolated in `RunStepRunner.Run`.

Also consider extending interpolation to automation-level `env:` values (evaluated once when the automation starts).

## Acceptance Criteria
- [x] `env:` values in steps support `outputs.last`, `outputs.<N>`, and `inputs.<name>` references
- [x] Automation-level `env:` values support the same references (resolved at automation start)
- [x] A test verifies env interpolation with `outputs.last`
- [x] `go build ./...` and `go test ./...` pass
- [ ] Docs updated (README and website if applicable)

## Implementation Notes

### Approach taken

Added `interpolateEnv` method on `*Executor` that applies `interpolateValue` to each value in an env map. Returns the original map unchanged if no values contain interpolation references (optimization to avoid unnecessary allocations).

**Step-level env:** In `execStep`, after building the RunContext, the step's `Env` map is interpolated using `e.interpolateEnv(step.Env, ctx.inputEnv)` and set on `RunContext.ResolvedStepEnv`. The `runStepCommand` function uses this override when building the process environment via `BuildStepEnv`.

**Automation-level env:** In `RunWithInputs`, the automation's `Env` map is interpolated once at automation start and stored on `stepExecCtx.automationEnv`. This is used throughout the automation's lifecycle instead of `a.Env`. At automation start, `outputs.*` references resolve to empty (since no steps have run yet), but `inputs.*` references work correctly.

**Key design decisions:**
- Added `ResolvedAutomationEnv` and `ResolvedStepEnv` fields to `RunContext` to avoid mutating the original `Automation` or `Step` structs
- `runStepCommand` checks for resolved env overrides and falls back to the struct fields when not set
- Install lifecycle (`execInstall`) is not affected — it doesn't use `stepExecCtx` and doesn't need env interpolation (install phases have a different lifecycle)

### Test coverage
Added 14 tests:
- 7 unit tests for `interpolateEnv` (nil, empty, no interpolation, outputs.last, outputs indexed, inputs, mixed)
- 7 integration tests (outputs.last in step env, outputs indexed in step env, inputs in step env, inputs in automation env, outputs.last in automation env at start, literal passthrough, mixed outputs and literals)

## Subtasks
- [x] Apply `interpolateValue` to step `env:` before building RunContext
- [x] Apply `interpolateValue` to automation-level `env:` at automation start
- [x] Write tests
- [ ] Update docs

## Blocked By
