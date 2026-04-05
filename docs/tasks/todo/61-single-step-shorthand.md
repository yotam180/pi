# Single-Step Automation Shorthand

## Type
improvement

## Status
todo

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

Support all subprocess step types as top-level shorthands: `bash:`, `python:`, `typescript:`, and `run:`. All step-level fields (`env:`, `dir:`, `timeout:`, `if:`, `silent:`, `pipe:`) should be usable alongside the shorthand key at the top level.

The shorthand is syntactic sugar — internally PI expands it to a single-step `steps:` list before execution. Having both `steps:` and a top-level step key in the same file is a parse error.

## Acceptance Criteria
- [ ] `bash:` as a top-level key executes as a single bash step
- [ ] `python:` as a top-level key executes as a single python step
- [ ] `typescript:` as a top-level key executes as a single typescript step
- [ ] `run:` as a top-level key executes as a single run step
- [ ] All step-level modifier fields work at the top level alongside the step key: `env:`, `dir:`, `timeout:`, `if:`, `silent:`, `pipe:`
- [ ] Having both a top-level step key and `steps:` in the same file is a parse error with a clear message
- [ ] `description:` at the top level remains the automation description (not the step description)
- [ ] `pi info` on a shorthand automation shows it correctly
- [ ] Tests cover each step type as shorthand, and the conflict error case

## Implementation Notes

## Subtasks
- [ ] Update YAML parser to detect top-level step keys
- [ ] Expand shorthand to single-step list internally
- [ ] Validate no coexistence with `steps:`
- [ ] Add tests
- [ ] Update example files

## Blocked By
