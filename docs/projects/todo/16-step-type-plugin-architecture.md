# Step Type Plugin Architecture

## Status
todo

## Priority
high

## Description

Adding a new step type (language runner) to PI currently requires 8+ coordinated changes across multiple files and packages — a strong signal that step types are not data-driven. This project completes the `SubprocessConfig`/`Registry` pattern that already exists by moving all per-type knowledge (file extensions, capability flags) into the config struct, and making the rest of the system derive from it. The goal: adding a new language runner is a single-file change.

## Goals

- `SubprocessConfig` carries the file extension for script files of that language.
- The central `IsFilePath()` function is replaced by per-runner dispatch using the extension in `SubprocessConfig`.
- `parent_shell: true` validity is checked against a runner capability flag, not against the hardcoded string `"bash"`.
- Positional arguments (`$1`, `$2`, `PI_ARG_1`, `PI_ARG_2`) are available in all subprocess step types, not just bash.
- Registering a new step type in `NewDefaultRegistry()` and defining its `SubprocessConfig` is sufficient — no other files need updating.

## Background & Context

The runner `Registry` and `SubprocessConfig` struct (introduced in task 51) were a big step toward extensibility, but they didn't go far enough. Three problems remain:

**1. `IsFilePath` in `executor/helpers.go` owns file extension knowledge:**
```go
func IsFilePath(value string) bool {
    return strings.HasSuffix(value, ".sh") ||
        strings.HasSuffix(value, ".py") ||
        strings.HasSuffix(value, ".ts")
}
```
This list must be manually updated when adding a new runner.

**2. `parent_shell` validation checks the type name:**
```go
if s.t != StepTypeBash {
    return Step{}, fmt.Errorf("step[%d]: 'parent_shell' is only valid on 'bash' steps", index)
}
```
The constraint should be "this runner supports producing shell eval-able commands," not "the step type is named bash."

**3. Positional args are bash-only:**
`ctx.Args` is passed as subprocess argv only by `SubprocessRunner`. Python and TypeScript get `$PI_ARGS` but not `$1`/`$2`. A clean fix: inject `PI_ARG_1`, `PI_ARG_2`... as env vars in `BuildStepEnv` whenever `ctx.Args` is non-empty, for all step types.

**Proposed `SubprocessConfig` additions:**
```go
type SubprocessConfig struct {
    // ... existing fields unchanged ...

    // FileExt is the script file extension for this language (e.g. ".sh", ".py", ".ts").
    // Used to identify file references vs. inline scripts. Required when the language
    // supports referencing external script files.
    FileExt string

    // SupportsParentShell indicates that steps of this type may use parent_shell: true.
    // Only set for languages whose output can be eval-ed by the calling shell (i.e., bash).
    SupportsParentShell bool
}
```

The `automation` package cannot import `executor`. To validate `parent_shell` at parse time without a circular import, the solution is to move the validation to execution time in the executor, where the runner is available. Alternatively: expose a `StepTypeCapabilities` registry in a shared package.

## Scope

### In scope
- Add `FileExt` to `SubprocessConfig`; update all three runner constructors
- Replace the global `IsFilePath` extension list with per-runner config lookup
- Add `SupportsParentShell bool` to `SubprocessConfig`; move `parent_shell` validation from parse time to execution time
- Add `PI_ARG_1`, `PI_ARG_2`... env var injection in `BuildStepEnv` for all step types
- Update `NewBashRunner()`, `NewPythonRunner()`, `NewTypeScriptRunner()` with correct flag values
- All tests pass; no new scanner of type names

### Out of scope
- Adding new step type languages (this project enables it, not does it)
- Changing the YAML schema for step types
- Runtime provisioning or install phase changes

## Success Criteria

- `grep -r "StepTypeBash" executor/` returns zero results outside of runner construction (no type-name comparisons for behavior)
- `grep -r '\.sh"\|\.py"\|\.ts"' executor/helpers.go` returns zero results (extensions live in configs)
- Python and TypeScript steps correctly receive `PI_ARG_1`, `PI_ARG_2`... when extra args are passed
- Adding a hypothetical `ruby:` step type requires: one new `SubprocessConfig` struct, one `Register()` call, and adding the constant to the `automation` package — nothing else
- `go build ./...` and `go test ./...` pass

## Notes

The `parent_shell` validation currently happens at YAML parse time in `automation/step.go`. Moving it to execution time means invalid automations are caught later (at run, not at validate). Consider adding the check to `pi validate` via the step walker instead — that preserves early detection without the circular import. The step walker already visits all steps; adding a capability check there is clean.
