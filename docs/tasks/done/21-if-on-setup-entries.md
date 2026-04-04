# `if:` Field on Setup Entries

## Type
feature

## Status
done

## Priority
medium

## Project
04-conditional-step-execution

## Description
Add an `if:` field to `pi.yaml` setup entries so that `pi setup` can conditionally skip certain setup steps based on the environment.

```yaml
setup:
  - run: setup/install-brew
    if: os.macos
  - run: setup/install-uv
    if: not command.uv
  - run: setup/install-node
```

Changes needed:
1. **Config schema**: Add `If string` field to `SetupEntry` struct in `internal/config/config.go`
2. **Setup command**: In `pi setup`, before running each entry, evaluate its `if:` condition. If false, print a skip notice and continue to the next entry.

## Acceptance Criteria
- [x] `if:` field is parsed from `pi.yaml` setup entries
- [x] Setup entries with `if:` evaluating to false are skipped with a message
- [x] Setup entries without `if:` always run (backward compatible)
- [x] `pi setup` still fails fast on step failure (non-zero exit)
- [x] Unit tests for config parsing
- [x] Integration test with example pi.yaml using `if:` on setup entries

## Implementation Notes

### Changes made
1. **Config schema** (`internal/config/config.go`): Added `If string` field to `SetupEntry` struct with `yaml:"if"` tag. Added validation in `validate()` that calls `conditions.Predicates()` on `if:` expressions at load time, catching invalid expressions early (same pattern as automation/step validation).

2. **Setup command** (`internal/cli/setup.go`): Added `evaluateSetupCondition()` function that uses `conditions.Predicates()` + `executor.ResolvePredicates()` + `conditions.Eval()` — the same three-step evaluation used by the executor's `evaluateCondition()`. In the setup loop, before running each entry, checks the `if:` condition. Skipped entries show: `==> setup[N]: <name> [skipped] (condition: <expr>)`.

3. **Tests**:
   - Config parsing: 2 new tests (valid `if:` parsing, invalid `if:` expression error) → 11 total config tests
   - Setup unit tests: 2 new tests (skip with false condition, run with true condition) → 6 total setup tests  
   - Integration tests: 2 new tests (conditional entries end-to-end, skip shows condition) → 58 total integration tests

4. **Example workspace**: Added 3 setup automations to `examples/conditional/` workspace (setup-always, setup-never, setup-platform) and updated its `pi.yaml` with conditional setup entries.

5. **Docs**: Updated architecture.md with new design decision section, updated README.md with `if:` in pi.yaml example and Environment Setup section.

## Subtasks
- [x] Add `If` field to `SetupEntry` in config package
- [x] Add condition evaluation in setup command
- [x] Write config parsing tests
- [x] Write integration test

## Blocked By
19-if-on-steps
20-if-on-automations
