# TypeScript Step Runner

## Type
feature

## Status
todo

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Add support for the `typescript` step type in the executor. TypeScript steps run via `tsx` (which must be installed). Steps can be inline scripts or references to `.ts` files (resolved relative to the automation YAML). Args are passed via `process.argv`. If `tsx` is not found, emit a clear error with an install hint (`npm install -g tsx`).

## Acceptance Criteria
- [ ] `typescript` step (inline): runs the string as a TypeScript script via `tsx` with a temp file, args available via `process.argv.slice(2)`
- [ ] `typescript` step (file path): runs the `.ts` file via `tsx <file>`, path resolved relative to the automation file's directory
- [ ] Exit code propagation works identically to bash steps
- [ ] Clear error if `tsx` is not found, including install hint
- [ ] `isFilePath` detection updated to handle `.ts` files
- [ ] Unit tests: inline success, inline failure, file step, args forwarded, tsx not found (if testable), file not found
- [ ] Mark `typescript` as implemented in `automation.go`'s `implementedStepTypes`

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Update `isFilePath()` to detect `.ts` extension
- [ ] Implement `execTypeScript()` in executor
- [ ] Add typescript case to `execStep()` switch
- [ ] Mark StepTypeTypeScript as implemented
- [ ] Write unit tests (at least 6)

## Blocked By
<!-- None — builds on the existing executor -->
