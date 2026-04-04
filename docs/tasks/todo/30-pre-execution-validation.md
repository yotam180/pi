# Pre-Execution Requirement Validation

## Type
feature

## Status
todo

## Priority
high

## Project
05-environment-robustness

## Description
Before running any step, the executor should validate all `requires:` entries on the automation. If any requirement is not satisfied, execution should fail immediately with a clear, formatted error table listing every missing requirement and a concrete install hint for each.

The validation should:
1. For runtime requirements (`python`, `node`): check if the command exists in PATH, then check `--version` output against the minimum version constraint
2. For command requirements: check if the command exists in PATH, optionally check version
3. Produce a formatted error table:

```
✗ pi run format-logs

  Missing requirements:
    python >= 3.11   not found (python3 --version returned 3.9.7)
                     → install: brew install python@3.13   or  pi setup (if configured)
    command: jq      not found
                     → install: brew install jq
```

Validation should only check requirements for steps that will actually execute (respect `if:` conditions). For installer automations, check requirements at the automation level.

## Acceptance Criteria
- [ ] `validateRequirements()` function in executor checks all requirements before step execution
- [ ] Runtime version checking: runs `<cmd> --version`, parses semver output, compares against constraint
- [ ] Command existence checking: uses `exec.LookPath()`
- [ ] Formatted error output with missing requirements table and install hints
- [ ] Validation respects `if:` conditions — skipped steps don't need their requirements checked
- [ ] Install hints are generated for common tools (python, node, docker, jq, etc.)
- [ ] Unit tests for validation logic (missing runtime, wrong version, missing command, all satisfied)
- [ ] Integration test: automation with `requires: python >= 99.0` fails with clear error
- [ ] `go test ./...` passes

## Implementation Notes

## Subtasks
- [ ] Implement `checkRequirement()` function for a single requirement
- [ ] Implement version parsing from `--version` output (various formats: `Python 3.13.0`, `v20.11.0`, `jq-1.7.1`)
- [ ] Implement semver comparison for `>=` constraints
- [ ] Build install hint table for common tools
- [ ] Integrate validation into `RunWithInputs()` before step execution
- [ ] Write formatted error output
- [ ] Write unit tests
- [ ] Write integration test

## Blocked By
29-requires-schema-parsing
