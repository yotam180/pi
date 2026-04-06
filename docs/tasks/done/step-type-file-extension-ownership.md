# Move File Extension Ownership into SubprocessConfig

## Type
improvement

## Status
done

## Priority
high

## Project
16-step-type-plugin-architecture

## Description

`IsFilePath` in `executor/helpers.go` contains a hardcoded list of file extensions (`.sh`, `.py`, `.ts`) that must be manually updated whenever a new language runner is added. This task moves the extension into `SubprocessConfig.FileExt` on each runner, and updates `SubprocessRunner.Run` to use it — so the global `IsFilePath` either disappears or becomes a thin generic helper with no hardcoded extensions.

**Current state:**
```go
// helpers.go
func IsFilePath(value string) bool {
    hasKnownExt := strings.HasSuffix(value, ".sh") ||
        strings.HasSuffix(value, ".py") ||
        strings.HasSuffix(value, ".ts")
    return hasKnownExt && !strings.Contains(value, "\n") && !strings.Contains(value, " ")
}
```

**Target state:**
```go
// SubprocessConfig — add:
FileExt string // e.g. ".sh", ".py", ".ts"

// helpers.go — make generic:
func IsFilePath(value, ext string) bool {
    return ext != "" &&
        strings.HasSuffix(value, ext) &&
        !strings.Contains(value, "\n") &&
        !strings.Contains(value, " ")
}

// SubprocessRunner.Run — use config:
resolved, isFile, err := resolveFileStep(ctx.Automation.Dir(), ctx.Step.Value, r.Config.Language, r.Config.FileExt)
```

**Runner updates:**
- `NewBashRunner()` → `FileExt: ".sh"`
- `NewPythonRunner()` → `FileExt: ".py"`
- `NewTypeScriptRunner()` → `FileExt: ".ts"` (already uses `TempFileExt: ".ts"` for temp files, but `FileExt` is for detection)

Note: `TypeScript` already uses `TempFileExt: ".ts"` for temp file creation. `FileExt` is for file-reference detection (separate concern). Both should be set to `.ts`.

## Acceptance Criteria
- [x] `SubprocessConfig` has a `FileExt string` field
- [x] All three runner constructors set `FileExt`
- [x] `IsFilePath` in `helpers.go` takes an extension parameter (or is removed)
- [x] `resolveFileStep` or `SubprocessRunner.Run` uses `r.Config.FileExt` for detection
- [x] No hardcoded `.sh`, `.py`, `.ts` strings in extension-checking logic outside of runner constructors
- [x] Adding `ruby:` with `FileExt: ".rb"` requires no changes to `helpers.go`
- [x] All existing tests pass

## Implementation Notes

### Approach
- `IsFilePath(value, ext string) bool` — now takes an extension parameter from the caller; no hardcoded extensions
- `resolveFileStep(automationDir, value, lang, ext string)` — also takes the extension parameter
- `SubprocessRunner.Run` passes `r.Config.FileExt` to `resolveFileStep`
- For `validate.go` (static file reference checking): added `Registry.FileExtForStepType()` and `DefaultFileExtensions()` which returns a `map[StepType]string` from the default registry. `checkFileReferences` now looks up the extension for each step's type.
- Tests updated to pass extension parameter. Added new test cases: `.rb` with `.rb` ext (true), `.sh` with `.py` ext (false), any value with empty ext (false).

### Key decisions
- Kept `IsFilePath` as an exported function (not inlined) since `validate.go` uses it independently of the runner
- Added `FileExtForStepType()` on `Registry` to type-assert to `*SubprocessRunner` and extract the extension — keeps the `StepRunner` interface clean
- `DefaultFileExtensions()` creates a throwaway default registry to build the map — acceptable since validation is not a hot path

## Subtasks
- [x] Add `FileExt` to `SubprocessConfig`
- [x] Update all three runner constructors
- [x] Update `IsFilePath` signature to take extension parameter
- [x] Update `resolveFileStep` call site
- [x] Update `SubprocessRunner.Run` to pass `r.Config.FileExt`
- [x] Add `FileExtForStepType()` and `DefaultFileExtensions()` for validate package
- [x] Update `validate.go` `checkFileReferences` to use per-type extensions
- [x] Update tests

## Blocked By
