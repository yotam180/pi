# Executor Quality Improvements and Test Gap Coverage

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

Targeted quality improvements to the executor and display packages:

1. **Fix `registry()` lazy init** — `Executor.registry()` creates a new `NewDefaultRegistry()` on every call when `Runners` is nil. This means each step execution allocates a fresh registry (4+ allocations per installer lifecycle). Cache the default once.

2. **Improve step failure error context** — When a subprocess step fails (exit code non-zero), the error message doesn't always include the automation name and file path. Add structured context to step errors so developers can immediately identify which automation and step failed, especially in multi-automation `pi setup` runs.

3. **Display package test gaps** — `Warn()` has 0% test coverage and `PackageFetch` with `⚠` icon is untested. Add tests to close these gaps.

## Acceptance Criteria
- [x] `registry()` caches the default registry on first call
- [x] Step execution errors include automation name context where missing
- [x] `Warn()` method has test coverage
- [x] `PackageFetch` with warning icon has test coverage
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### registry() caching
The `Executor.registry()` method was creating a new `NewDefaultRegistry()` on every call when `e.Runners` was nil. This happened 4+ times per installer lifecycle (test, run, verify, version). Added a `cachedRegistry` field that stores the lazily-created default once.

### Error context
`execStep` was wrapping errors with just `"step[%d]"` for dir errors and unimplemented step types, and passing runner errors raw without automation name context. Updated to consistently include `automation %q step[%d]` in error messages, matching the pattern used by `execParentShell` and condition evaluation. ExitError is still returned raw (intentional — the CLI exits silently with the code, since the step's stderr already communicated the failure).

### Display test gaps
- Added `TestWarn_NoColor` and `TestWarn_WithColor` — Warn was at 0% coverage, now at 100%
- Added `TestPackageFetch_Warning_NoColor` and `TestPackageFetch_Warning_WithColor` — the ⚠ icon path was untested
- Display package coverage improved from 85.7% to 89.8%

## Subtasks
- [x] Fix registry() caching
- [x] Audit and improve step error context
- [x] Add display.Warn test
- [x] Add display.PackageFetch warning icon test
- [x] Run full test suite
- [x] Update architecture docs

## Blocked By
