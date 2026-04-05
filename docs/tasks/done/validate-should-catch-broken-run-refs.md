# pi validate should catch broken run: references in steps

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
`pi validate` does not catch `run:` steps that reference non-existent automations. A broken reference only surfaces at runtime.

### Steps to reproduce
1. Create `.pi/check.yaml`:
```yaml
description: Run all checks
steps:
  - run: fmt-check
  - run: lint
  - run: test
```
2. Ensure `fmt-check` does NOT exist as an automation.
3. Run `pi validate`.

### Expected
```
✗ check: step[0] run: references unknown automation "fmt-check"
Validation failed: 1 error(s)
```

### Actual
```
✓ Validated 30 automation(s), 6 shortcut(s), 1 setup entry(ies)
```

Validation passes silently. The error is only discovered at runtime when `pi run check` fails.

### Notes
This check did work in an earlier build — step[0] `run: "cargo build"` was correctly flagged as referencing an unknown automation. The regression may be related to the `pi new` changes or the `--` forwarding work, where `run:` step resolution behavior may have been altered.

## Acceptance Criteria
- [x] `pi validate` reports an error for any `run:` step that references a non-existent automation
- [x] Both workspace automations and builtins are considered valid targets
- [x] Add a test case covering this scenario

## Implementation Notes
Verified that the bug was already fixed by the current codebase. The `validateRunStepRefs` function in `internal/cli/validate.go` correctly walks all steps via `automation.WalkSteps` and calls `disc.Find(step.Value)` for every `run:` step type. Both unit tests (`TestValidate_BrokenRunStep`) and integration tests (`TestValidate_InvalidProject`) cover this scenario. The `validate-invalid` fixture includes `ghost-automation` as a broken run: ref and all tests pass.

## Subtasks
- [x] Verify `validateRunStepRefs` catches broken references (confirmed working)
- [x] Verify test coverage exists (unit + integration tests present and passing)

## Blocked By
