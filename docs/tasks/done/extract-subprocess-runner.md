# Extract SubprocessRunner from Bash/Python/TypeScript Runners

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `BashRunner`, `PythonRunner`, and `TypeScriptRunner` in `internal/executor/runners.go` share a nearly identical code structure:

1. Resolve file step via `resolveFileStep()`
2. Build command args (file path vs inline script — with language-specific differences)
3. Call `runStepCommand()`
4. Handle errors: check `ExitError` passthrough, check `isCommandNotFound()`, wrap with context

The only differences between them are:
- **Binary name** — `bash`, `python3`/venv, `tsx`
- **Inline args format** — bash uses `-c <script> --`, python uses `-c <script>`, typescript creates a temp file
- **Missing-binary error message** — each has a specific message
- **Binary resolution** — python has venv detection, others don't

This duplication makes it harder to add new language runners (e.g., Ruby, Lua, Deno) and violates DRY. Extract a shared `SubprocessRunner` struct that encapsulates the common pattern, with a `SubprocessConfig` that captures the per-language differences.

### Design

```go
type SubprocessConfig struct {
    Binary      string              // command to run (or "" to use BinaryFunc)
    BinaryFunc  func() string       // dynamic binary resolution (e.g. python venv)
    InlineArgs  func(script string) []string  // how to pass inline scripts
    TempFileExt string              // if non-empty, write inline to temp file with this ext
    NotFoundMsg string              // user-friendly error message when binary missing
    Language    string              // "bash", "python", "typescript" — for error messages
}
```

The `SubprocessRunner` handles the full lifecycle:
1. `resolveFileStep()` for file-path values
2. `InlineArgs()` or temp-file creation for inline scripts
3. `runStepCommand()` with the resolved binary
4. Error handling with `ExitError` passthrough and `NotFoundMsg`

The existing `BashRunner`, `PythonRunner`, `TypeScriptRunner` types are replaced by `SubprocessRunner` instances in `NewDefaultRegistry()`. The `StepRunner` interface is unchanged.

### Benefits
- Adding a new language runner becomes a ~10-line config struct
- Common error handling is in one place
- Test coverage of the shared path applies to all languages
- Expands the platform toward user-configurable step types in the future

## Acceptance Criteria
- [x] `SubprocessRunner` struct with `SubprocessConfig` exists in `internal/executor/runners.go`
- [x] `BashRunner`, `PythonRunner`, `TypeScriptRunner` are replaced with `SubprocessRunner` instances
- [x] `NewDefaultRegistry()` uses the new config-driven approach
- [x] All existing tests pass unchanged (behavior preservation)
- [x] `go build ./...` and `go test ./...` pass
- [x] Architecture docs updated

## Implementation Notes

Starting implementation. The key decisions:
- Keep `RunStepRunner` separate — it doesn't follow the subprocess pattern
- `SubprocessConfig` uses function fields for flexibility (BinaryFunc, InlineArgs)
- TypeScript's temp-file pattern becomes `TempFileExt: ".ts"` in config
- `resolvePythonBin()` moves to a `BinaryFunc` closure
- Error wrapping preserves exact same messages for backward compatibility

## Subtasks
- [x] Design SubprocessConfig struct
- [x] Implement SubprocessRunner.Run()
- [x] Replace BashRunner/PythonRunner/TypeScriptRunner
- [x] Update NewDefaultRegistry()
- [x] Verify all tests pass
- [x] Update architecture doc

## Blocked By
