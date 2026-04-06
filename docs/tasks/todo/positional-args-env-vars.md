# Expose Positional Args as PI_ARG_N Env Vars for All Step Types

## Type
improvement

## Status
todo

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
- Bash: `$1`, `$2`, `$@` + `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`...
- Python: `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`... (use `os.environ["PI_ARG_1"]`)
- TypeScript: `$PI_ARGS` + `PI_ARG_1`, `PI_ARG_2`... (use `process.env.PI_ARG_1`)

**Implementation:** In `BuildStepEnv` (or in the executor before calling `BuildStepEnv`), when `ctx.Args` is non-empty, inject:
```go
for i, arg := range ctx.Args {
    env = append(env, fmt.Sprintf("PI_ARG_%d=%s", i+1, arg))
}
```

This is additive — existing behavior is unchanged.

Also update `PI_ARG_COUNT` (or similar) if needed, or just document the convention.

## Acceptance Criteria
- [ ] `PI_ARG_1`, `PI_ARG_2`... are injected into the step environment for all subprocess step types when extra args are passed
- [ ] Existing `$PI_ARGS` and bash positional args behavior is unchanged
- [ ] A Python step test confirms `PI_ARG_1` etc. are accessible
- [ ] `go build ./...` and `go test ./...` pass
- [ ] README/docs updated to mention `PI_ARG_N` for multi-language positional arg access

## Implementation Notes

## Subtasks
- [ ] Add `PI_ARG_N` injection to `BuildStepEnv` or executor
- [ ] Write test for Python step accessing `PI_ARG_1`
- [ ] Write test for TypeScript step accessing `PI_ARG_1`
- [ ] Update docs

## Blocked By
