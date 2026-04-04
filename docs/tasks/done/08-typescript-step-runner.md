# TypeScript Step Runner

## Type
feature

## Status
done

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Add support for the `typescript` step type in the executor. TypeScript steps run via `tsx` (which must be installed). Steps can be inline scripts or references to `.ts` files (resolved relative to the automation YAML). Args are passed via `process.argv`. If `tsx` is not found, emit a clear error with an install hint (`npm install -g tsx`).

## Acceptance Criteria
- [x] `typescript` step (inline): runs the string as a TypeScript script via `tsx` with a temp file, args available via `process.argv.slice(2)`
- [x] `typescript` step (file path): runs the `.ts` file via `tsx <file>`, path resolved relative to the automation file's directory
- [x] Exit code propagation works identically to bash steps
- [x] Clear error if `tsx` is not found, including install hint
- [x] `isFilePath` detection updated to handle `.ts` files
- [x] Unit tests: inline success, inline failure, file step, args forwarded, tsx not found (if testable), file not found
- [x] Mark `typescript` as implemented in `automation.go`'s `implementedStepTypes`

## Implementation Notes

### Approach
- Inline TypeScript steps write the script to a temp file (`pi-ts-*.ts`) and run `tsx <tmpfile> [args...]`. This is necessary because `tsx` doesn't have a `-c` flag like Python.
- File steps run `tsx <resolved-path> [args...]` directly.
- Temp files are cleaned up via `defer os.Remove()`.
- The `isFilePath()` function already detected `.ts` extensions from before — no changes needed there.
- Error handling follows the same pattern as bash/python: exit code propagation via `ExitError`, clear "tsx not found" message with install hint.

### Tests added (8 total)
1. `TestTypeScriptInline_Success` — inline script writes to file
2. `TestTypeScriptInline_Failure` — exit code 42 propagated
3. `TestTypeScriptInline_WithArgs` — args available via `process.argv.slice(2)`
4. `TestTypeScriptFile_Success` — `.ts` file resolved relative to automation dir
5. `TestTypeScriptFile_WithArgs` — file step with args
6. `TestTypeScriptFile_NotFound` — clear error for missing `.ts` file
7. `TestTypeScriptTsxNotFound` — clear error with install hint when tsx is absent (tested by overriding PATH)
8. `TestMixedSteps_BashAndTypeScript` — bash→ts→bash multi-step automation

### Updated existing tests
- `TestLoad_UnsupportedStepType_TypeScript` → renamed to `TestLoad_TypeScriptStep_Accepted` (now verifies TS steps are accepted)
- `TestStepType_IsImplemented` → updated to expect `true` for TypeScript

## Subtasks
- [x] Update `isFilePath()` to detect `.ts` extension (already done)
- [x] Implement `execTypeScript()` in executor
- [x] Add typescript case to `execStep()` switch
- [x] Mark StepTypeTypeScript as implemented
- [x] Write unit tests (8 written, exceeds minimum of 6)

## Blocked By
<!-- None — builds on the existing executor -->
