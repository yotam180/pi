# Support timeout on run: Steps

## Type
research

## Status
done

## Priority
low

## Project
standalone

## Description

`timeout: 30s` is currently forbidden on `run:` steps with the reasoning that "set timeouts on the target automation's steps instead." This advice is inadequate when the target automation is a third-party builtin or package automation whose steps you don't control. The caller has no way to bound a sub-automation's execution time.

This task is initially research: understand the right design and update the docs/error message before implementing anything.

**The problem:**
```yaml
steps:
  - run: pi:install-python     # Can't timeout this — but it could hang
    timeout: 5m                # Currently a parse error
```

**Options to evaluate:**

**Option A — Implement timeout on run: steps:**
A `run:` step can declare `timeout:`, which the executor honors by running the entire recursive `RunWithInputs` call inside a `context.WithTimeout`. If the sub-automation exceeds the timeout, the current step in the sub-automation receives a kill signal and the `run:` step returns `ExitError{Code: 124}`.
- Pros: Clean user model, same timeout semantics as subprocess steps
- Cons: Killing sub-automation mid-flight may leave state (partial installs etc.)

**Option B — Top-level automation timeout:**
`pi.yaml` setup entries and `pi run` could accept `--timeout 5m`. This bounds the whole automation, not individual steps.
- Pros: Simpler implementation, no need to plumb context through `RunWithInputs`
- Cons: Less granular; doesn't help inside multi-step automations

**Option C — Keep current restriction, improve error message:**
The restriction stays, but the error message is improved to explain *why* and suggest using `--timeout` at the CLI level (once Option B exists), or wrapping the `run:` in a bash timeout:
```yaml
- bash: timeout 300 pi run pi:install-python
```
- Pros: Zero implementation cost
- Cons: Bash workaround is ugly; doesn't compose

**Recommendation:** Implement Option A. The context plumbing for `RunWithInputs` is already partially in place (the executor has a `context.Background()` in `runStepCommand`). Promoting it to a shared cancellable context is achievable. The risk of partial state on timeout exists but is acceptable — it's the same risk as ctrl+C.

## Acceptance Criteria
- [x] Research document (this file updated with decision) is complete
- [x] If implementing: `timeout:` on `run:` steps works end-to-end with the same 124 exit code semantics
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

**Decision: Option A implemented** in task `context-propagation-and-run-timeout`.

The implementation followed the plumbing path outlined below, with one key difference: instead of modifying the public `RunWithInputs()` signature, an internal `runWithContext()` method was added. This minimized the blast radius while achieving full context propagation.

Implementation details:
1. `executor/executor.go` — added `runWithContext(ctx, a, args, withArgs, stdout, stdin)` as internal entry point; `RunWithInputs()` delegates to it
2. `executor/runners.go` — `RunStepRunner.Run()` wraps context with `context.WithTimeout` when step has `timeout:`; `runStepCommand()` uses `RunContext.Ctx` as base context
3. `executor/runner_iface.go` — `RunContext` gained `Ctx context.Context` field; `RunAutomation` callback accepts `context.Context`
4. `automation/step.go` — removed parse-time restriction on `timeout:` for `run:` steps

## Subtasks
- [x] Evaluate options, update implementation notes with decision
- [x] Implement chosen option
- [x] Write tests
- [x] Update docs

## Blocked By
