# `if:` Field on Setup Entries

## Type
feature

## Status
todo

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
- [ ] `if:` field is parsed from `pi.yaml` setup entries
- [ ] Setup entries with `if:` evaluating to false are skipped with a message
- [ ] Setup entries without `if:` always run (backward compatible)
- [ ] `pi setup` still fails fast on step failure (non-zero exit)
- [ ] Unit tests for config parsing
- [ ] Integration test with example pi.yaml using `if:` on setup entries

## Implementation Notes

## Subtasks
- [ ] Add `If` field to `SetupEntry` in config package
- [ ] Add condition evaluation in setup command
- [ ] Write config parsing tests
- [ ] Write integration test

## Blocked By
19-if-on-steps
20-if-on-automations
