# Validate Unknown YAML Fields in Automation Files

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

When a user writes a typo in an automation YAML file — for example `descrption:` instead of `description:`, or `step:` instead of `steps:` — the field is silently ignored by `yaml.v3`'s default unmarshalling. This is a common source of confusion: the user thinks they've set a description or defined steps, but PI never sees the data.

This violates two core philosophy principles:
- **Principle #2 (Correct the action, not the developer):** PI can detect the typo and tell the user.
- **Principle #6 (Explain the magic, don't hide it):** Silently ignoring fields hides what PI is (not) doing.

**Goal:** Add a 10th validation check to `pi validate` that detects unknown YAML keys in automation files and reports them with helpful suggestions (e.g., "unknown field 'descrption' in docker/up — did you mean 'description'?").

**Approach:** Use `yaml.v3`'s `yaml.Node` API to walk the raw YAML tree and compare top-level keys against the set of known automation fields. Unknown keys trigger a validation warning with Levenshtein-based "did you mean?" suggestions. The check also validates step-level keys for the same reason.

**Known automation-level fields:** `name`, `description`, `steps`, `install`, `inputs`, `if`, `requires`, `bash`, `python`, `typescript`, `run`, `env`, `dir`, `timeout`, `silent`, `parent_shell`, `pipe_to`, `pipe`, `with`

**Known step-level fields:** `bash`, `python`, `typescript`, `run`, `if`, `env`, `dir`, `timeout`, `silent`, `parent_shell`, `pipe_to`, `pipe`, `description`, `first`, `with`

## Acceptance Criteria
- [x] New validation check `unknown-fields` registered in `DefaultRunner()`
- [x] Detects unknown top-level keys in automation YAML files
- [x] Detects unknown step-level keys in step entries
- [x] Reports "did you mean?" suggestions using Levenshtein distance
- [x] Only checks local automations (not builtins or packages)
- [x] All existing tests pass unchanged
- [x] New tests cover: valid files pass, unknown top-level field, unknown step field, suggestion accuracy, multiple unknowns, installer fields, shorthand fields
- [x] `go build ./...` and `go test ./...` pass
- [x] Architecture docs updated

## Implementation Notes

### Design decisions

The check needs raw YAML access (node tree), which the current `validate.Context` doesn't have. Two options:

1. **Re-read and re-parse YAML files in the check** — each automation has `FilePath`, so the check reads the file, parses into `yaml.Node`, and inspects keys. This keeps the existing validate API clean.

2. **Extend Context to include raw YAML nodes** — this changes the discovery/loading pipeline.

Going with option 1: the check reads files independently. This is consistent with the `checkFileReferences` pattern which also does I/O (os.Stat calls). The file read is cheap (automation files are tiny) and isolated.

For builtins (embedded), we skip the check since they're authored by PI itself and tested at build time. For packages, we also skip — package authors own their YAML quality.

### Levenshtein suggestions

The `discovery` package already has `levenshtein()` and `suggestNames()` in `suggest.go`. To reuse this in `validate`, we have options:
- Duplicate the small function (it's 15 lines)
- Extract to a shared `internal/stringutil` or similar

Going with: putting a small `suggestField()` helper directly in the validate check, reusing the same algorithm. It's small enough to not warrant a shared package, and the discovery one is purpose-built for automation name suggestions.

### What was done

1. **New file: `internal/validate/unknown_fields.go`** — contains all unknown-field detection logic:
   - `checkUnknownFields()` — the registered check function, iterates local (non-builtin, non-package) automations
   - `checkFileUnknownFields()` — reads YAML file, parses as `yaml.Node`, checks top-level keys + descends into `steps:` and `install:`
   - `checkStepNodeUnknownFields()` — validates step-level keys, descends into `first:` sub-steps
   - `checkFirstSubStepUnknownFields()` — validates keys in first: block sub-steps
   - `checkInstallNodeUnknownFields()` — validates install: block keys (test/run/verify/version)
   - `suggestField()` — Levenshtein-based closest-match suggestion (max distance = len/2, min 2)
   - `levenshtein()` — Wagner-Fischer edit distance (duplicated from discovery/suggest.go, 25 lines)
   - Three known-field sets: `knownAutomationKeys` (20 keys), `knownStepKeys` (15 keys), `knownInstallKeys` (4 keys)

2. **Registration**: Added as 10th check in `DefaultRunner()` — runs after all other checks

3. **Design decisions**:
   - Re-reads YAML files independently (doesn't extend Context with raw nodes) — keeps API clean, consistent with `checkFileReferences` pattern
   - Skips builtins and packages — they're authored/owned externally
   - Returns no errors for missing/unparseable files — other checks handle those
   - Suggestion threshold: max Levenshtein distance = len(key)/2 with floor of 2 — catches common typos without false positives

4. **Test coverage**:
   - 29 unit tests in `unknown_fields_test.go` covering all code paths
   - 3 CLI integration tests in `cli/validate_test.go` (end-to-end via `runValidate`)
   - Total test count: 1783 (up from 1748)

## Subtasks
- [x] Create task file
- [x] Implement the check in validate/unknown_fields.go
- [x] Register as check #10 in DefaultRunner()
- [x] Write unit tests (29 in unknown_fields_test.go + 3 CLI integration tests)
- [x] Run full test suite (1783 tests, all pass)
- [x] Update architecture.md
- [ ] Commit

## Blocked By
