# Shorten Input Variable Prefix: `PI_INPUT_*` → `PI_IN_*`

## Type
improvement

## Status
todo

## Priority
medium

## Project
12-yaml-ergonomics

## Description
Input variables injected into step environments currently use the prefix `PI_INPUT_`. For an input named `version`, the variable is `$PI_INPUT_VERSION`. This prefix is verbose, especially in inline bash/python scripts with multiple inputs.

Shorten the prefix to `PI_IN_`. So `version` → `$PI_IN_VERSION`, `greeting` → `$PI_IN_GREETING`.

Both prefixes should be injected for every input so that existing automations using `$PI_INPUT_*` continue to work. `PI_INPUT_*` should be documented as deprecated — it will be removed in a future major version.

## Acceptance Criteria
- [ ] `$PI_IN_<NAME>` is available in all step environments when the automation defines inputs
- [ ] `$PI_INPUT_<NAME>` continues to work (backward compatibility)
- [ ] `pi info` shows the new `$PI_IN_*` form in input documentation
- [ ] All built-in and example automations that reference inputs updated to use `$PI_IN_*`
- [ ] Tests cover: `$PI_IN_*` resolves correctly, `$PI_INPUT_*` still resolves (with same value)

## Implementation Notes

## Subtasks
- [ ] Update step env injection to set both `PI_IN_*` and `PI_INPUT_*`
- [ ] Update `pi info` output
- [ ] Update built-in automations (install-python, etc.)
- [ ] Update example automations
- [ ] Add/update tests

## Blocked By
