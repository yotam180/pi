# Deduplicate Printer Construction and Info Annotation Building

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The CLI package has two patterns that are duplicated across multiple files, increasing maintenance burden and making it harder to add new commands correctly:

### 1. Printer construction (4 occurrences)

The following 3-line pattern for creating a `display.Printer` that auto-detects TTY appears in `add.go`, `discover.go` (x2):

```go
printer := display.NewWithColor(stderr, false)
if f, ok := stderr.(*os.File); ok {
    printer = display.New(f)
}
```

Extract this into a `display.NewForWriter(w io.Writer) *Printer` helper that encapsulates the TTY detection logic. The existing `New()` already does TTY detection but requires `*os.File` — `NewForWriter()` accepts `io.Writer` and gracefully degrades to no-color when the underlying writer isn't a file-backed terminal.

### 2. Info annotation building (duplicated in printStepsDetail and printFirstBlockDetail)

The annotation-building logic in `info.go` is duplicated between regular steps and first-block sub-steps. Both functions build `[]string` annotation slices checking the same fields (if, pipe, silent, parent_shell, dir, timeout, env) with the same formatting. Extract a shared `stepAnnotations(step) []string` function.

## Acceptance Criteria
- [x] `display.NewForWriter()` exists and encapsulates TTY detection for arbitrary `io.Writer`
- [x] All 3 occurrences of the manual printer pattern in `cli/` use `NewForWriter()`
- [x] `stepAnnotations()` helper extracted in `info.go`
- [x] Both `printStepsDetail` and `printFirstBlockDetail` use the shared helper
- [x] All existing tests pass (`go test ./...`)
- [x] New tests for `NewForWriter()` added to `display_test.go`
- [x] Documentation updated

## Implementation Notes

### Approach
- `NewForWriter()` is semantically identical to the existing inline pattern: try to cast to `*os.File`, use `New()` (TTY-aware) if it is, fall back to `NewWithColor(w, false)` otherwise.
- The annotation helper returns `[]string` so callers can still join and format as they wish (regular steps vs sub-steps have different label formats).
- This is a pure refactor — no behavior change. All existing tests serve as regression tests.

### Decisions
- Named it `NewForWriter` not `NewAutoDetect` because the former reads naturally at call sites: `display.NewForWriter(stderr)`.

## Subtasks
- [x] Add `NewForWriter()` to `display.go`
- [x] Add unit tests for `NewForWriter()` in `display_test.go`
- [x] Update `cli/add.go` to use `NewForWriter()`
- [x] Update `cli/discover.go` (2 occurrences) to use `NewForWriter()`
- [x] Extract `stepAnnotations()` in `cli/info.go`
- [x] Update `printStepsDetail()` and `printFirstBlockDetail()` to use it
- [x] Run `go test ./...` — all pass
- [x] Update docs

## Blocked By
