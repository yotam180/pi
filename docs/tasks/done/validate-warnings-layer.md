# Add Warnings Layer to pi validate

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `pi validate` command currently only reports errors (things that are broken). Add a warnings layer that surfaces non-fatal issues — things that aren't broken but could be improved. This is a key extensibility improvement because it prepares the infrastructure for user-configurable linting rules and helps teams maintain higher-quality automation files.

Changes:
1. Add a `Warnings` field to `validate.Result` alongside `Errors`
2. Add `WarnCheck` interface and `WarnCheckFunc` adapter for warning-producing checks
3. Update the `Runner` with `RegisterWarn()` and `RunWithOpts(ctx, includeWarnings)` methods
4. Add three warning checks to `DefaultRunner()`
5. Update the CLI validate command with `--warnings` / `-w` flag

Warning checks implemented:
- **missing-description**: Automations without a `description:` field. Descriptions power `pi list`, `pi info`, and AI-assisted workflows.
- **unused-automations**: Local automations not referenced by any shortcut, setup entry, or `run:` step from any other automation. Detects dead automations.
- **shortcut-shadowing**: Shortcuts whose names shadow shell builtins or common commands (reuses `shell.CheckShadowedNames()`). Already checked at `pi shell` install time, but validates earlier.

## Acceptance Criteria
- [x] `validate.Result` has a `Warnings []string` field
- [x] `WarnCheck` interface and `WarnCheckFunc` adapter added
- [x] Warning checks run via `Runner.RunWithOpts(ctx, true)` — `Run(ctx)` backwards-compatible (no warnings)
- [x] `pi validate` shows warnings when `--warnings` / `-w` flag is passed
- [x] Default `pi validate` (no flag) only shows errors (backwards compatible)
- [x] Success message includes warning count when warnings exist
- [x] Three warning checks implemented: missing-description, unused-automations, shortcut-shadowing
- [x] All new checks have comprehensive unit tests (32 warning tests + 6 CLI tests)
- [x] `go build ./...` passes
- [x] `go test ./...` passes (all 17 packages)
- [x] Architecture docs updated
- [x] README CLI reference updated

## Implementation Notes

### Design decisions
- **`WarnCheck` as separate interface from `Check`**: Warnings are fundamentally different from errors — they don't affect exit codes, they're opt-in, and they require different display formatting. Using the same interface would require every check to declare its severity, which adds complexity to the 11 existing checks.
- **`RunWithOpts` instead of modifying `Run`**: The `Run(ctx)` method stays backwards-compatible for all existing callers. `RunWithOpts(ctx, includeWarnings)` is the new entry point for the CLI.
- **Warnings hidden by default**: Following the principle of least surprise. `pi validate` with no flags should produce the same output as before. Teams that want linting opt in with `--warnings`.
- **Warnings not shown when errors exist**: If the project has broken references, showing "missing description" warnings would be noise. Fix errors first, then lint.
- **Replaced `empty-steps` with `shortcut-shadowing`**: Initially planned an `empty-steps` check, but the automation parser rejects files with no steps/install/step-key at parse time. Replaced with `shortcut-shadowing` which catches a real-world issue (already checked by `pi shell` but not by `pi validate`).

### Coverage
- validate package: 96.4% → 97.1%
- CLI package: 84.2% → 84.3%

### Test counts
- 32 new tests in `warnings_test.go`
- 6 new tests in `validate_test.go` (CLI-level --warnings flag tests)

## Subtasks
- [x] Add Warnings to Result struct and create WarnCheck interface
- [x] Implement missing-description check
- [x] Implement unused-automations check
- [x] Implement shortcut-shadowing check
- [x] Update CLI validate command with --warnings flag
- [x] Write tests (32 unit + 6 integration)
- [x] QA and docs

## Blocked By
