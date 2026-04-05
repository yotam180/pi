# Migrate Error Type Assertions to errors.As/errors.Is

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
The codebase uses direct type assertions (`err.(*SomeError)`) for error type checking throughout production and test code. This is fragile — if any error gets wrapped with `fmt.Errorf("...: %w", err)`, the type assertion silently fails and the wrapped error type is lost.

Go's `errors.As()` and `errors.Is()` (introduced in Go 1.13) handle unwrapping automatically and are the idiomatic way to check error types. Migrating to them makes error handling robust against future error wrapping and follows Go best practices.

### Affected error types:
- `*executor.ExitError` — checked in root.go, runners.go, executor.go, and many test files
- `*executor.ValidationError` — checked in executor.go
- `*config.DuplicatePackageError` — checked in add.go
- `*config.DuplicateSetupEntryError` — checked in setup_add.go
- `*config.ReplacedSetupEntryError` — checked in setup_add.go
- `*exec.ExitError` (stdlib) — checked in runners.go, integration test helpers

### Scope:
- Production code: ~10 instances across cli/, executor/, config/
- Test code: ~18 instances across executor tests, cli tests, integration tests

## Acceptance Criteria
- [x] All `err.(*SomeType)` assertions in production code replaced with `errors.As()`
- [x] All `err.(*SomeType)` assertions in test code replaced with `errors.As()`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./...` passes
- [x] No regressions in integration tests
- [x] architecture.md updated to mention errors.As/errors.Is convention

## Implementation Notes
- Pattern: `if exitErr, ok := err.(*ExitError); ok {` → `var exitErr *ExitError; if errors.As(err, &exitErr) {`
- Pattern: `if _, ok := err.(*ExitError); ok {` → `if errors.As(err, &exitErr) {` (or `var target *ExitError; errors.As(err, &target)`)
- Add `"errors"` import where needed
- For `*exec.ExitError` in integration tests: same pattern applies
- stdlib's `exec.ExitError` already supports `errors.As` via `Unwrap()`

## Subtasks
- [x] Migrate internal/executor/runners.go (4 instances)
- [x] Migrate internal/executor/executor.go (1 instance)
- [x] Migrate internal/cli/root.go (1 instance)
- [x] Migrate internal/cli/add.go (1 instance)
- [x] Migrate internal/cli/setup_add.go (2 instances)
- [x] Migrate test files (~18 instances)

## Blocked By
