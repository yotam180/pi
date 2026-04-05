# Single-Step Automation Shorthand

## Type
improvement

## Status
done

## Priority
high

## Project
12-yaml-ergonomics

## Description
A large percentage of automations are a single step. Currently they require the full `steps:` wrapper:

```yaml
description: Run the pytest test suite
steps:
  - bash: venv/bin/python -m pytest
```

This should be writable as a top-level step key:

```yaml
description: Run the pytest test suite
bash: venv/bin/python -m pytest
```

Support all subprocess step types as top-level shorthands: `bash:`, `python:`, `typescript:`, and `run:`. All step-level fields (`env:`, `dir:`, `timeout:`, `if:`, `silent:`, `pipe_to:`) should be usable alongside the shorthand key at the top level.

The shorthand is syntactic sugar — internally PI expands it to a single-step `steps:` list before execution. Having both `steps:` and a top-level step key in the same file is a parse error.

## Acceptance Criteria
- [x] `bash:` as a top-level key executes as a single bash step
- [x] `python:` as a top-level key executes as a single python step
- [x] `typescript:` as a top-level key executes as a single typescript step
- [x] `run:` as a top-level key executes as a single run step
- [x] All step-level modifier fields work at the top level alongside the step key: `env:`, `dir:`, `timeout:`, `if:`, `silent:`, `pipe_to:`
- [x] Having both a top-level step key and `steps:` in the same file is a parse error with a clear message
- [x] `description:` at the top level remains the automation description (not the step description)
- [x] `pi info` on a shorthand automation shows it correctly
- [x] Tests cover each step type as shorthand, and the conflict error case

## Implementation Notes

### Approach
The shorthand is implemented entirely in `UnmarshalYAML` on the `Automation` struct. The raw struct was extended with optional top-level step keys (`bash`, `python`, `typescript`, `run`) plus step modifier fields (`env`, `dir`, `timeout`, `silent`, `pipe_to`).

When a top-level step key is detected, `buildShorthandStep()` constructs a `stepRaw`, applies the modifier fields, converts to a `Step` via `toStep(0)`, and sets `a.Steps` to a single-element slice. The rest of the system (executor, info, validate, discovery) sees a normal single-step automation — no changes needed downstream.

### Error cases handled
- Multiple top-level step keys → clear error
- Top-level step key + `steps:` → clear error
- Top-level step key + `install:` → clear error
- `validate()` error message updated to mention the shorthand option

### Top-level `if:` semantics
The `if:` at top level maps to the automation-level condition (same as before). For a single-step shorthand, this is semantically equivalent to a step-level `if:` since there's only one step. No ambiguity.

### Test coverage
- 12 unit tests in `automation_test.go`: bash/python/typescript/run shorthands, modifiers (env/dir/timeout/silent), if condition, conflict with steps, conflict with install, multiple step keys, pipe_to, multiline, inputs
- 8 integration tests in `shorthand_integ_test.go`: list, run bash, run with env, run step delegation, run with input, info, info with modifiers, validate

## Subtasks
- [x] Update YAML parser to detect top-level step keys
- [x] Expand shorthand to single-step list internally
- [x] Validate no coexistence with `steps:`
- [x] Add tests
- [x] Update example files

## Blocked By
