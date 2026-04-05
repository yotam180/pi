# pi validate command

## Type
feature

## Status
in_progress

## Priority
high

## Project
standalone

## Description
Implement a `pi validate` CLI command that statically validates all automation YAML files and `pi.yaml` without executing anything. This catches schema errors, broken references, and configuration mistakes early — ideal for CI pipelines.

The command should:
1. Find project root (walk up to `pi.yaml`)
2. Parse and validate `pi.yaml` (config validation)
3. Discover and parse all `.pi/` automations (schema validation)
4. Cross-validate references: check that shortcuts, setup entries, and `run:` steps reference automations that exist
5. Report all errors (not just the first one) with file paths and clear messages
6. Exit 0 on success, 1 on validation errors
7. Print a summary: "Validated N automations, M shortcuts, K setup entries"

## Acceptance Criteria
- [x] `pi validate` command exists and works
- [x] Validates pi.yaml config (project name, shortcuts, setup entries)
- [x] Validates all automation YAML files in .pi/
- [x] Cross-validates that shortcut run: targets exist in discovered automations
- [x] Cross-validates that setup entry run: targets exist in discovered automations  
- [x] Cross-validates that run: steps reference existing automations
- [x] Reports all errors, not just the first
- [x] Exit code 0 on success, 1 on errors
- [x] Prints success summary with counts
- [x] Unit tests for the validate logic
- [x] Integration tests with example workspaces (valid + invalid)
- [x] Documentation updated (architecture.md, README.md)

## Implementation Notes
Implementing as a new `internal/cli/validate.go` file with a `newValidateCmd()` function.

Validation layers:
1. Config validation — already handled by `config.Load()`, just surface errors
2. Discovery validation — already handled by `discovery.Discover()`, surface errors  
3. Cross-reference validation — NEW logic: check that all `run:` references (shortcuts, setup, steps) resolve to known automations
4. Summary reporting — count and report results

Cross-reference checking is the novel part. The existing `Find()` method handles resolution but returns on first error. We need to check all references and collect errors.

Decision: Keep the validate logic in cli/validate.go since it's a CLI command that wires together existing packages. The cross-reference checking doesn't need its own package — it's thin glue code.

Decision: Include built-in automations in the validation target set, since `run: pi:something` and `setup: pi:install-python` are valid references.

## Subtasks
- [x] Create `internal/cli/validate.go` with command implementation
- [x] Add cross-reference validation for shortcuts → automations
- [x] Add cross-reference validation for setup entries → automations
- [x] Add cross-reference validation for run: steps → automations
- [x] Add unit tests in `internal/cli/validate_test.go`
- [x] Create `examples/validate-valid/` workspace for integration tests
- [x] Create `examples/validate-invalid/` workspace for integration tests
- [x] Add integration tests in `tests/integration/examples_test.go`
- [x] Register command in `root.go`
- [x] Update docs

## Blocked By
