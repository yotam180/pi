# pi doctor Command

## Type
feature

## Status
todo

## Priority
medium

## Project
05-environment-robustness

## Description
Add a `pi doctor` command that scans all automations in the project, collects their `requires:` entries, and prints a per-automation health table showing which requirements are satisfied and which are missing.

Output format:
```
pi doctor

  docker/logs-formatted
    ✓ python >= 3.11     (3.13.0)
    ✓ command: docker    (24.0.5)

  setup/build-image
    ✓ command: docker    (24.0.5)
    ✗ command: kubectl   not found → brew install kubectl
```

The command should:
- Discover all automations (local + built-in)
- Skip automations with no `requires:` block
- For each automation with requirements, check each one and display status
- Exit with code 0 if all requirements are met, 1 if any are missing
- Be fast — only PATH lookups and `--version` calls, no network requests

## Acceptance Criteria
- [ ] `pi doctor` cobra command added
- [ ] Discovers all automations and collects requirements
- [ ] Prints per-automation health table with ✓/✗ icons
- [ ] Shows detected version in parentheses for satisfied requirements
- [ ] Shows install hint for missing requirements
- [ ] Exit code 0 when all satisfied, 1 when any missing
- [ ] Automations without `requires:` are silently skipped
- [ ] Integration test with example workspace
- [ ] `go test ./...` passes

## Implementation Notes

## Subtasks
- [ ] Add `doctor.go` to `internal/cli/`
- [ ] Reuse requirement checking logic from task 30
- [ ] Build formatted output
- [ ] Add integration test
- [ ] Update CLI reference in docs

## Blocked By
30-pre-execution-validation
