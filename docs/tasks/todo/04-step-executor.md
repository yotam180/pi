# Step Executor

## Type
feature

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Implement the execution engine that runs an automation's steps in order. Covers bash steps (inline and file) and `run:` steps (call another automation). Handles argument passing, working directory, exit code propagation, and clear failure output.

## Acceptance Criteria
- [ ] `bash` step (inline): runs the string as a bash script, inherits stdout/stderr, `$@` receives args passed to `pi run`
- [ ] `bash` step (file path): runs the `.sh` file, path resolved relative to the automation file's directory
- [ ] `run:` step: looks up the named automation and executes it recursively; args are forwarded
- [ ] Circular `run:` dependency detection with a clear error (A → B → A)
- [ ] If any step exits non-zero, execution stops and `pi run` exits with the same code
- [ ] Working directory for all steps is the repo root (the directory containing `pi.yaml`), not the `.pi/` subdirectory
- [ ] Unit tests: bash inline success, bash inline failure (exit 1), file step, run: step chaining, circular dep detection

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Implement bash step runner (inline)
- [ ] Implement bash step runner (file path)
- [ ] Implement run: step runner with recursion guard
- [ ] Wire exit code propagation
- [ ] Unit tests

## Blocked By
03-automation-discovery
