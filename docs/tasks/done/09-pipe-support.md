# Pipe Support Between Steps

## Type
feature

## Status
done

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Implement `pipe_to: next` support on steps. When a step has `pipe_to: next`, its stdout is captured and piped as stdin to the next step. This is the core mechanism for PI's polyglot value proposition — e.g., piping `docker-compose logs` through a Python log formatter.

## Acceptance Criteria
- [x] A step with `pipe_to: next` has its stdout captured (not printed to terminal)
- [x] The captured stdout becomes stdin of the next step
- [x] The last step's output goes to the terminal as normal (even if it has `pipe_to: next`, that's a no-op on the last step)
- [x] Works for all step types: bash, python, typescript, run
- [x] Steps without `pipe_to` print to terminal normally (current behavior)
- [x] Stderr always goes to the terminal (not captured)
- [x] Exit code propagation still works (if piping step fails, pipeline stops)
- [x] Unit tests: bash→bash pipe, bash→python pipe, pipe chain (3 steps), pipe with failure in middle, stderr passes through, no pipe (default behavior unchanged)

## Implementation Notes

### Key design decisions

1. **Executor.Stdout/Stderr changed from `*os.File` to `io.Writer`**: This was necessary because piping requires writing stdout to a `bytes.Buffer` instead of a file descriptor. The refactor also simplified the CLI→executor interface (removed type assertions in `run.go`).

2. **Added `Executor.Stdin io.Reader`**: Steps not receiving piped input use the executor's Stdin (defaulting to `os.Stdin`). Piped steps receive the previous step's captured buffer.

3. **Added `lastPipeBuffer *bytes.Buffer` field on Executor**: Stores the captured stdout from a pipe source step, passed as stdin to the next step.

4. **Pipe logic in `Run()` loop**: Each step checks `isPipeSrc = step.PipeTo == "next" && i < len(steps)-1`. When true, stdout goes to a buffer. The buffer becomes the next step's stdin. When `pipe_to: next` on the last step, it's a no-op (output goes to terminal).

5. **`execRun` with piping**: For `run:` steps that receive piped input, the executor temporarily overrides its `Stdout`/`Stdin` fields, then restores them after the nested automation completes.

6. **Each `exec*` method now takes `stdout io.Writer` and `stdin io.Reader` parameters** instead of reading from the executor struct directly. This makes the piping explicit and composable.

### Test coverage

- 10 unit tests: bash→bash, bash→python, python→bash, 3-step chain, failure in middle, stderr passthrough, no-pipe default, last-step no-op, through run: step, multiline data
- 3 integration tests: `examples/pipe/` workspace with `upper` (bash→bash pipe) and `count-lines` (bash→python→bash 3-step chain)

## Subtasks
- [x] Modify executor to detect `pipe_to: next` and capture stdout
- [x] Pass captured output as stdin to next step
- [x] Ensure stderr is not affected
- [x] Write unit tests (at least 6) — wrote 10
- [x] Write integration tests — wrote 3
- [x] Create `examples/pipe/` workspace

## Blocked By
07-python-step-runner (for testing cross-language pipes) — resolved, already done
