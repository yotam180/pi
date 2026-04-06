# Context Propagation and Timeout on Run Steps

## Type
improvement

## Status
in_progress

## Priority
high

## Project
standalone

## Description

Thread `context.Context` through the executor's execution paths and use it to implement timeout on `run:` steps — a feature that was previously forbidden. This also fixes the `RunAutomation` closure in `newRunContext` that temporarily mutates `e.Stdout`/`e.Stdin` (a thread-safety and correctness concern), and lays the groundwork for cancellation, graceful shutdown, and future parallel step execution.

**Current problems:**
1. `runStepCommand()` creates its own `context.Background()` — there's no way for a caller to cancel or set a deadline on running steps
2. `RunAutomation` closure in `newRunContext` mutates `e.Stdout` and `e.Stdin` during recursive calls, then restores them — fragile and not concurrency-safe
3. `timeout:` on `run:` steps is forbidden with the rationale "set timeouts on the target automation's steps instead" — but this doesn't work for third-party builtins/packages

**Implementation plan:**
1. Add `context.Context` parameter to `RunWithInputs()` and `Run()` (or wrap via an internal method)
2. Pass context through `execStep()`, `execFirstBlock()`, `execInstall()`, and into `RunContext`
3. `runStepCommand()` uses the passed context instead of `context.Background()`
4. When a `run:` step has `timeout:`, the executor wraps the context with `context.WithTimeout` before recursive `RunWithInputs`
5. Replace `RunAutomation` closure's Stdout/Stdin mutation with explicit parameters passed through context or RunContext
6. Remove the validation/parse restriction that prevents `timeout:` on `run:` steps
7. Update docs

**This resolves the `timeout-on-run-steps` research task (Option A).**

## Acceptance Criteria
- [x] `context.Context` flows through all executor execution paths
- [x] `RunAutomation` closure no longer mutates `e.Stdout` / `e.Stdin`
- [x] `timeout:` on `run:` steps works with exit code 124 (same semantics as subprocess timeouts)
- [x] Existing timeout tests still pass
- [x] No behavioral regressions — `go test ./...` passes
- [x] Architecture docs updated

## Implementation Notes

### Approach: Internal context threading

Instead of changing the public `Run()`/`RunWithInputs()` signatures (which would have a large blast radius), added an internal `runWithContext()` method that carries `context.Context` and explicit `stdout`/`stdin`. The public methods delegate to it with `context.Background()`.

### Key changes:

1. **`executor.go`**: Added `runWithContext(ctx, a, args, withArgs, stdout, stdin)` as the internal execution entry point. `RunWithInputs()` delegates to it. The `RunAutomation` closure in `newRunContext` now calls `runWithContext()` directly, passing the context and explicit I/O — no more `e.Stdout`/`e.Stdin` mutation.

2. **`runner_iface.go`**: Added `Ctx context.Context` field to `RunContext`. Updated `RunAutomation` callback signature to accept `context.Context` as first parameter.

3. **`runners.go`**: `RunStepRunner.Run()` wraps the context with `context.WithTimeout` when `Step.Timeout > 0` on run: steps. `runStepCommand()` uses `RunContext.Ctx` as the base context instead of `context.Background()`, enabling cancellation from parent run: step timeouts.

4. **`step.go`**: Removed the parse-time restriction that prevented `timeout:` on `run:` steps.

5. **`install.go`**: All install lifecycle methods now thread `context.Context` through.

### Naming convention
Used `goCtx` for `context.Context` parameters and `sCtx` for `*stepExecCtx` to avoid confusion with the overloaded "ctx" term.

### Run step timeout semantics
When a `run:` step declares `timeout:`, `RunStepRunner` creates a deadline context. The target automation runs within this context — all subprocess commands inherit the deadline via `exec.CommandContext`. If the deadline expires:
- The currently running subprocess is killed
- `runStepCommand` returns an error
- `RunStepRunner` detects `context.DeadlineExceeded` and returns `ExitError{Code: 124}`

This matches the same exit code 124 semantics as subprocess step timeouts.

## Subtasks
- [x] Thread context.Context through RunWithInputs and execution paths
- [x] Eliminate Stdout/Stdin mutation in RunAutomation closure
- [x] Allow timeout on run: steps in parser validation
- [x] Implement context deadline for run: step timeouts
- [x] Write tests (5 new unit tests, 3 new integration tests)
- [x] Update docs (README.md, architecture.md)

## Blocked By
