# Pipe Support Between Steps

## Type
feature

## Status
todo

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Implement `pipe_to: next` support on steps. When a step has `pipe_to: next`, its stdout is captured and piped as stdin to the next step. This is the core mechanism for PI's polyglot value proposition — e.g., piping `docker-compose logs` through a Python log formatter.

## Acceptance Criteria
- [ ] A step with `pipe_to: next` has its stdout captured (not printed to terminal)
- [ ] The captured stdout becomes stdin of the next step
- [ ] The last step's output goes to the terminal as normal (even if it has `pipe_to: next`, that's a no-op on the last step)
- [ ] Works for all step types: bash, python, typescript, run
- [ ] Steps without `pipe_to` print to terminal normally (current behavior)
- [ ] Stderr always goes to the terminal (not captured)
- [ ] Exit code propagation still works (if piping step fails, pipeline stops)
- [ ] Unit tests: bash→bash pipe, bash→python pipe, pipe chain (3 steps), pipe with failure in middle, stderr passes through, no pipe (default behavior unchanged)

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Modify executor to detect `pipe_to: next` and capture stdout
- [ ] Pass captured output as stdin to next step
- [ ] Ensure stderr is not affected
- [ ] Write unit tests (at least 6)

## Blocked By
07-python-step-runner (for testing cross-language pipes)
