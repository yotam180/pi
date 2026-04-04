# pi doctor Command

## Type
feature

## Status
done

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
- [x] `pi doctor` cobra command added
- [x] Discovers all automations and collects requirements
- [x] Prints per-automation health table with ✓/✗ icons
- [x] Shows detected version in parentheses for satisfied requirements
- [x] Shows install hint for missing requirements
- [x] Exit code 0 when all satisfied, 1 when any missing
- [x] Automations without `requires:` are silently skipped
- [x] Integration test with example workspace
- [x] `go test ./...` passes

## Implementation Notes

### Architecture
- `internal/cli/doctor.go` — Cobra command and output formatting
- Reuses `discoverAll()` from `internal/cli/discover.go` for automation discovery
- Uses new `CheckRequirementForDoctor()` from `internal/executor/validate.go` — always detects version even without constraints
- Uses new `InstallHintFor()` from `internal/executor/validate.go` — exported install hint lookup
- Output format: per-automation section header, then indented `✓`/`✗` lines with label, version, and install hint

### Key decisions
- Built-in automations are included in the scan (they currently have no `requires:` so they're silently skipped)
- The `CheckRequirementForDoctor()` function always runs version detection (unlike `checkRequirement()` which skips it when no constraint is set) — this is needed to show the detected version for display purposes
- Exit code 1 on any missing requirement, 0 when all satisfied — matches the task spec

### Test coverage
- 9 unit tests in `internal/cli/doctor_test.go` (help, root help, no-automations, no-requirements, satisfied, missing, exit-code, mixed, skips-no-requires)
- 5 unit tests in `internal/executor/validate_test.go` (CheckRequirementForDoctor with/without constraint, not-found, InstallHintFor known/unknown)
- 7 integration tests in `tests/integration/examples_test.go` (all-satisfied, missing-requirements, version-mismatch, skips-no-requires, detected-version, install-hint, healthy-workspace)

## Subtasks
- [x] Add `doctor.go` to `internal/cli/`
- [x] Reuse requirement checking logic from task 30
- [x] Build formatted output
- [x] Add integration test
- [x] Update CLI reference in docs

## Blocked By
30-pre-execution-validation (done)
