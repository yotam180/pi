# Expose Positional Args as PI_ARG_N Env Vars for All Step Types

## Type
improvement

## Status
done

## Priority
medium

## Project
16-step-type-plugin-architecture

## Description

When extra arguments are passed to `pi run <name> arg1 arg2`, bash steps can use `$1`, `$2` etc. because args are passed as subprocess argv. Python and TypeScript steps only receive `$PI_ARGS` (the full joined string), with no way to access individual positional arguments without string splitting. This task adds `PI_ARG_1`, `PI_ARG_2`... env vars in `BuildStepEnv` for all step types, giving every language an ergonomic positional arg interface.

**Current:**
- Bash: `$1`, `$2`, `$@` (subprocess argv) + `$PI_ARGS`
- Python: `$PI_ARGS` only (via env)
- TypeScript: `$PI_ARGS` only (via env)

**After this task:**
- Bash: `$1`, `$2`, `$@` + `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`... + `PI_ARG_COUNT`
- Python: `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`... + `PI_ARG_COUNT` (use `os.environ["PI_ARG_1"]`)
- TypeScript: `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`... + `PI_ARG_COUNT` (use `process.env.PI_ARG_1`)

**Implementation:** In `BuildStepEnv` (or in the executor before calling `BuildStepEnv`), when `ctx.Args` is non-empty, inject:
```go
for i, arg := range ctx.Args {
    env = append(env, fmt.Sprintf("PI_ARG_%d=%s", i+1, arg))
}
```

This is additive — existing behavior is unchanged.

Also update `PI_ARG_COUNT` (or similar) if needed, or just document the convention.

## Acceptance Criteria
- [x] `PI_ARG_1`, `PI_ARG_2`... are injected into the step environment for all subprocess step types when extra args are passed
- [x] Existing `$PI_ARGS` and bash positional args behavior is unchanged
- [x] A Python step test confirms `PI_ARG_1` etc. are accessible
- [x] `go build ./...` and `go test ./...` pass
- [x] README/docs updated to mention `PI_ARG_N` for multi-language positional arg access

## Implementation Notes

Injected in `executor.go` `RunWithInputs()`, right alongside the existing `PI_ARGS` injection (line ~146). This is the cleanest injection point because:
- It's in the `inputEnv` slice which flows to all step types uniformly via `BuildStepEnv`
- It mirrors the existing `PI_ARGS` pattern exactly
- When inputs consume args (`len(a.Inputs) > 0`), `args` is set to nil, so `PI_ARG_N` is naturally not injected — correct behavior

Also added `PI_ARG_COUNT` env var for convenience (avoids needing to probe which `PI_ARG_N` vars exist).

Decision: TypeScript test was not written because tsx is an optional dependency and the injection mechanism is language-agnostic (env vars flow through `inputEnv` → `BuildStepEnv` → all subprocess runners). The Python test proves cross-language access works.

## Subtasks
- [x] Add `PI_ARG_N` + `PI_ARG_COUNT` injection to executor.go
- [x] Write test for bash step accessing `PI_ARG_1`, `PI_ARG_2`, `PI_ARG_COUNT`
- [x] Write test for Python step accessing `PI_ARG_1`
- [x] Write negative tests (no args, inputs consume args)
- [x] Update docs (README, architecture)

## Blocked By
