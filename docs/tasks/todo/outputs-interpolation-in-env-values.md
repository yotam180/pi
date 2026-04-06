# Allow outputs.last / outputs.N in Step env: Values

## Type
feature

## Status
todo

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
- [ ] `env:` values in steps support `outputs.last`, `outputs.<N>`, and `inputs.<name>` references
- [ ] Automation-level `env:` values support the same references (resolved at automation start)
- [ ] A test verifies env interpolation with `outputs.last`
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Docs updated (README and website if applicable)

## Implementation Notes

## Subtasks
- [ ] Apply `interpolateValue` to step `env:` before building RunContext
- [ ] Apply `interpolateValue` to automation-level `env:` at automation start
- [ ] Write tests
- [ ] Update docs

## Blocked By
