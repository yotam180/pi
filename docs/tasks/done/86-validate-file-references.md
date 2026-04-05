# Validate File References in pi validate

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
`pi validate` currently checks YAML schema validity and cross-references (shortcut→automation, setup→automation, run:→automation), but it does not validate that file-path references in steps actually exist on disk. When a step uses `bash: my-script.sh` or `python: transform.py`, the script file must exist relative to the automation YAML file's directory — but we only discover this at runtime.

Add file reference validation to `pi validate` so that broken script references are caught at static analysis time. This applies to all subprocess step types (bash, python, typescript) where the step value looks like a file path (uses `isFilePath()` from executor/helpers.go).

The validation should:
1. Walk all automation steps (including first: block sub-steps and install phase steps)
2. For each step with a file-path value, resolve the path relative to the automation's directory
3. Check that the resolved file exists on disk
4. Report all broken file references (not just the first)

This improves the user experience significantly — users can run `pi validate` in CI to catch broken file references before they fail at runtime.

## Acceptance Criteria
- [x] `pi validate` checks file-path step values exist on disk for bash, python, typescript steps — DONE
- [x] File paths are resolved relative to the automation's YAML file directory (using automation.Dir()) — DONE
- [x] All broken file references are reported (not short-circuiting on first error) — DONE
- [x] Install phase step file references are also validated — DONE (scalar phases + step list phases)
- [x] first: block sub-step file references are also validated — DONE
- [x] New unit tests in cli/validate_test.go cover the new validation — 6 new tests added
- [x] New integration test in tests/integration/validate_integ_test.go covers end-to-end — 3 new tests added
- [x] Built-in automations are excluded from file reference checks (they use inline scripts only) — DONE
- [x] Existing tests continue to pass — all 1151 tests pass

## Implementation Notes
### Approach
- Extract `isFilePath()` from `internal/executor/helpers.go` to a shared location (or the automation package) so both executor and validate can use it without import cycles
- Actually, `isFilePath` is simple enough that we can duplicate it or just use it directly from executor — since cli already imports executor for ExitError, we can access it if we make it exported. But that breaks package boundaries. Better: move the helper to the automation package since it's about step value classification.
- Decision: Export `IsFilePath` from `executor/helpers.go` — the cli package already imports executor, so no new dependency needed. The function is a pure string check with no side effects.
- Add `validateFileReferences()` in `cli/validate.go` that walks all automations and checks file-path steps
- File reference errors use the same ValidationResult.Errors pattern as existing checks
- Built-in automations (detected via `result.IsBuiltin(name)`) are skipped since they use inline scripts and have no real filesystem directory

### Tech decisions
- Reuse `executor.IsFilePath()` for consistent path detection (same rules as runtime)
- Reuse `resolveScriptPath()` logic — but since it's in executor, just do `filepath.Join(a.Dir(), value)` directly in validate.go (it's one line)
- Report format: `<automation-name>: step[N] <type>: file not found: <path> (resolved to <absolute-path>)`

## Subtasks
- [x] Export IsFilePath from executor/helpers.go
- [x] Add validateFileReferences() to cli/validate.go
- [x] Handle install phase steps and first: blocks
- [x] Skip built-in automations
- [x] Add unit tests
- [x] Add integration test with example workspace
- [x] Update architecture.md

## Blocked By
