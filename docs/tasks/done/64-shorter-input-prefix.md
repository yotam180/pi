# Shorten Input Variable Prefix: `PI_INPUT_*` → `PI_IN_*`

## Type
improvement

## Status
done

## Priority
medium

## Project
12-yaml-ergonomics

## Description
Input variables injected into step environments currently use the prefix `PI_INPUT_`. For an input named `version`, the variable is `$PI_INPUT_VERSION`. This prefix is verbose, especially in inline bash/python scripts with multiple inputs.

Shorten the prefix to `PI_IN_`. So `version` → `$PI_IN_VERSION`, `greeting` → `$PI_IN_GREETING`.

Both prefixes should be injected for every input so that existing automations using `$PI_INPUT_*` continue to work. `PI_INPUT_*` should be documented as deprecated — it will be removed in a future major version.

## Acceptance Criteria
- [x] `$PI_IN_<NAME>` is available in all step environments when the automation defines inputs
- [x] `$PI_INPUT_<NAME>` continues to work (backward compatibility)
- [x] `pi info` shows the new `$PI_IN_*` form in input documentation
- [x] All built-in and example automations that reference inputs updated to use `$PI_IN_*`
- [x] Tests cover: `$PI_IN_*` resolves correctly, `$PI_INPUT_*` still resolves (with same value)

## Implementation Notes

### Approach
- `InputEnvVars()` in `internal/automation/inputs.go` is the single point where input env vars are constructed. Changed it to emit both `PI_IN_<NAME>=<value>` and `PI_INPUT_<NAME>=<value>` for each input, in that order per key (sorted alphabetically).
- `pi info` now shows `→ $PI_IN_<NAME>` next to each input spec line, making the env var reference immediately visible.
- All built-in automations (install-python, install-node, install-go, install-rust, git/install-hooks, cursor/install-extensions) updated to use `$PI_IN_*`.
- All example automations (inputs/greet, docker-project/docker/logs, installer-schema/install-marker, shorthand/greet-input) updated to use `$PI_IN_*`.
- Existing tests that use `$PI_INPUT_*` in bash scripts continue to pass because both prefixes are set.

### Files Changed
- `internal/automation/inputs.go` — core change: emit both prefixes
- `internal/cli/info.go` — display `→ $PI_IN_<NAME>` in printInputSpec
- `internal/automation/inputs_test.go` — updated to expect both prefixes
- `internal/cli/info_test.go` — updated to expect `→ $PI_IN_*` format
- `internal/executor/inputs_test.go` — added 2 new tests for short prefix and both-prefixes
- `internal/builtins/embed_pi/*.yaml` — all 6 files updated to `$PI_IN_*`
- `internal/builtins/builtins_test.go` — updated string assertions
- `examples/**/*.yaml` — 4 files updated to `$PI_IN_*`
- `tests/integration/inputs_test.go` — added info env var prefix test
- `docs/README.md` — examples updated
- `docs/architecture.md` — descriptions updated

## Subtasks
- [x] Update step env injection to set both `PI_IN_*` and `PI_INPUT_*`
- [x] Update `pi info` output
- [x] Update built-in automations (install-python, etc.)
- [x] Update example automations
- [x] Add/update tests

## Blocked By
