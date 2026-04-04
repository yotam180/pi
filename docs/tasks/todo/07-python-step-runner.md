# Python Step Runner

## Type
feature

## Status
todo

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Add support for the `python` step type in the executor. Python steps can be either inline scripts or references to `.py` files (resolved relative to the automation YAML). The step should detect and use an active virtualenv (`VIRTUAL_ENV` env var) if present, otherwise fall back to `python3`. Args are passed via `sys.argv`. If `python3` is not found, emit a clear error.

## Acceptance Criteria
- [ ] `python` step (inline): runs the string as a Python script via `python3 -c "<script>"`, args available as `sys.argv[1:]`
- [ ] `python` step (file path): runs the `.py` file, path resolved relative to the automation file's directory
- [ ] If `VIRTUAL_ENV` is set, uses `$VIRTUAL_ENV/bin/python` instead of system `python3`
- [ ] Exit code propagation works identically to bash steps
- [ ] Clear error if `python3` is not found in PATH (and no venv active)
- [ ] `isFilePath` detection updated to also handle `.py` files
- [ ] Unit tests: inline success, inline failure, file step, args forwarded, venv detection, file not found
- [ ] Mark `python` as implemented in `automation.go`'s `implementedStepTypes`

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Update `isFilePath()` to detect `.py` extension
- [ ] Implement `execPython()` in executor
- [ ] Add python case to `execStep()` switch
- [ ] Mark StepTypePython as implemented
- [ ] Write unit tests (at least 8)

## Blocked By
<!-- None — builds on the existing executor -->
