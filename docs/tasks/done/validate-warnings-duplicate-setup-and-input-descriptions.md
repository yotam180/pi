# Add validation warnings for duplicate setup entries and missing input descriptions

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description
Add two new warning checks to `pi validate --warnings`:

1. **`warnDuplicateSetupEntries`** — warns when the same `run:` target appears multiple times in `setup:`. Duplicate setup entries are almost always a copy-paste mistake. The check compares resolved `run:` targets (not raw strings) so that `setup/install-go` and `setup/install-go` are caught, but entries with different `if:` conditions or different `with:` values are not flagged (they're intentional variants).

2. **`warnMissingInputDescription`** — warns when a local automation declares `inputs:` but one or more input specs lack a `description:` field. Input descriptions power `pi info` output and help AI assistants understand how to use automations (philosophy principle 9: "AI is a first-class user"). Builtins and package automations are excluded from this check.

Both checks follow the existing `WarnCheck` pattern in `internal/validate/warnings.go`.

## Acceptance Criteria
- [x] `warnDuplicateSetupEntries` detects exact `run:` duplicates in `setup:`
- [x] `warnDuplicateSetupEntries` does NOT flag entries with different `with:` or `if:` values
- [x] `warnMissingInputDescription` flags local automations with inputs missing `description:`
- [x] `warnMissingInputDescription` skips builtins and package automations
- [x] Both checks are registered in `DefaultRunner()` as warn checks
- [x] Both checks have comprehensive unit tests
- [x] `pi validate --warnings` shows the new warnings on applicable projects
- [x] `go build ./...` and `go test ./...` pass
- [x] architecture.md updated with new check descriptions

## Implementation Notes
- `warnDuplicateSetupEntries` uses a composite key of `(run, if, with)` to identify duplicates. The `with:` map is serialized as sorted `key=value` pairs joined by commas for deterministic comparison. Entries with different `if:` or `with:` values are intentional variants (e.g. installing Python 3.12 and 3.13) and are not flagged.
- `warnMissingInputDescription` iterates `InputKeys` (preserving declaration order) and checks each `InputSpec.Description`. Results are sorted alphabetically for deterministic output.
- Both checks follow the existing `WarnCheckFunc` pattern and are registered in `DefaultRunner()`.
- Total warn checks increased from 3 to 5. Test count in `warnings_test.go` increased from 32 to 41.
- Coverage for `internal/validate` improved from 97.1% to 97.3%.

## Subtasks
- [x] Implement `warnDuplicateSetupEntries` in `warnings.go`
- [x] Implement `warnMissingInputDescription` in `warnings.go`
- [x] Add tests in `warnings_test.go`
- [x] Register in `DefaultRunner()`
- [x] Update architecture.md
- [x] Update DefaultRunner check count in validate_test.go

## Blocked By
