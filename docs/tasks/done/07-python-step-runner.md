# Python Step Runner

## Type
feature

## Status
done

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Add support for the `python` step type in the executor. Python steps can be either inline scripts or references to `.py` files (resolved relative to the automation YAML). The step should detect and use an active virtualenv (`VIRTUAL_ENV` env var) if present, otherwise fall back to `python3`. Args are passed via `sys.argv`. If `python3` is not found, emit a clear error.

## Acceptance Criteria
- [x] `python` step (inline): runs the string as a Python script via `python3 -c "<script>"`, args available as `sys.argv[1:]`
- [x] `python` step (file path): runs the `.py` file, path resolved relative to the automation file's directory
- [x] If `VIRTUAL_ENV` is set, uses `$VIRTUAL_ENV/bin/python` instead of system `python3`
- [x] Exit code propagation works identically to bash steps
- [x] Clear error if `python3` is not found in PATH (and no venv active)
- [x] `isFilePath` detection updated to also handle `.py` files (and `.ts` for future TypeScript task)
- [x] Unit tests: inline success, inline failure, file step, args forwarded, venv detection, file not found, multiline, mixed bash+python
- [x] Mark `python` as implemented in `automation.go`'s `implementedStepTypes`

## Implementation Notes

### Args passing
- For inline Python: `python3 -c "script" arg1 arg2` — args land in `sys.argv[1:]` (no `--` separator, unlike bash)
- For file Python: `python3 script.py arg1 arg2` — args in `sys.argv[1:]`

### Virtualenv detection
- `resolvePythonBin()` checks `$VIRTUAL_ENV` env var. If set, uses `$VIRTUAL_ENV/bin/python`; otherwise `python3`
- Tested with a fake venv containing a bash wrapper that writes a marker file

### Error handling
- File not found: explicit stat check before exec, clear error with resolved path
- python3 not found: `isCommandNotFound()` helper detects exec errors for missing binaries

### `isFilePath()` changes
- Extended to detect `.py` and `.ts` extensions (both needed; `.ts` pre-emptively for task 08)
- Same rules: no newlines, no spaces

### Tests added (9 new)
- TestPythonInline_Success
- TestPythonInline_Failure (exit code 42)
- TestPythonInline_WithArgs
- TestPythonFile_Success
- TestPythonFile_WithArgs
- TestPythonFile_NotFound
- TestPythonInline_Multiline
- TestPythonVenvDetection (fake venv with bash wrapper)
- TestMixedSteps_BashAndPython (interleaved bash+python)
- Also updated isFilePath tests for .py and .ts cases

### Existing tests updated
- TestLoad_UnsupportedStepType_Python → TestLoad_PythonStep_Accepted (now expects success)
- TestStepType_IsImplemented: Python now returns true

## Subtasks
- [x] Update `isFilePath()` to detect `.py` extension
- [x] Implement `execPython()` in executor
- [x] Add python case to `execStep()` switch
- [x] Mark StepTypePython as implemented
- [x] Write unit tests (9 tests, exceeding the 8 minimum)

## Blocked By
<!-- None — builds on the existing executor -->
