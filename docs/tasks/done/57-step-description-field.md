# Step Description Field

## Type
feature

## Status
in_progress

## Priority
medium

## Project
standalone

## Description
Add an optional `description:` field to steps in automation YAML files. This allows automation authors to document what each step does in human-readable terms, improving the output of `pi info` and making automations more self-documenting.

Currently, `pi info` shows step details with truncated command text, which is often cryptic:
```
  1. bash: docker-compose logs -f --tail 200  [silent]
  2. python: transform.py
```

With `description:`, authors can add context:
```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    description: Stream container logs
    silent: true
  - python: transform.py
    description: Format and filter log output
```

And `pi info` would show:
```
  1. bash: docker-compose logs -f --tail 200  [silent]
     Stream container logs
  2. python: transform.py
     Format and filter log output
```

### Scope
- Add `description` field to `Step` struct and `stepRaw` unmarshalling
- Support `description:` on both regular steps and install phase steps
- Display descriptions in `pi info` step detail output
- Backward compatible: steps without `description:` behave exactly as before
- Works with all step types: bash, python, typescript, run

### Out of scope
- Showing descriptions in step trace lines (those should stay concise)
- Description on automations (already have `description:` at top level)

## Acceptance Criteria
- [x] `description:` field parsed from step YAML
- [x] `pi info` shows step descriptions when present
- [x] Unit tests cover parsing with and without description
- [x] Unit tests cover `pi info` display with descriptions
- [x] Integration test verifies end-to-end behavior
- [x] All existing tests pass unchanged
- [x] Documentation updated

## Implementation Notes
- `description` is a simple string field on `Step` — no complex parsing needed
- Added to `stepRaw` as `yaml:"description"` and mapped in `toStep()`
- Display in `pi info` as an indented line below the step detail line
- Compatible with all other step fields (if:, env:, dir:, timeout:, silent:, parent_shell:)

## Subtasks
- [x] Add `Description` field to `Step` struct in automation.go
- [x] Add `Description` field to `stepRaw` struct
- [x] Map description in `toStep()`
- [x] Update `printStepsDetail()` in info.go to show descriptions
- [x] Add unit tests for parsing
- [x] Add unit tests for pi info display
- [x] Add integration test with example workspace
- [x] Update architecture.md and README.md

## Blocked By
