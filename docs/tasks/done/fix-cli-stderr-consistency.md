# Fix CLI stderr parameter consistency in discover.go

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

`discoverAllWithConfig()` accepts a `stderr io.Writer` parameter for routing diagnostic output, but three internal call sites ignore it and hardcode `os.Stderr`:

1. `discovery.Discover(piDir, os.Stderr)` on line 30 — should use `stderr` (or a fallback to `os.Stderr` only when `stderr` is nil, consistent with how the caller controls output)
2. `newOnDemandFetcher(os.Stderr)` on line 48 — should use `stderr`
3. `result.MergePackage(pkg.Source, pkg.As, pkgDir, os.Stderr)` in `mergePackages()` on line 147 — should use the `stderr` parameter that `mergePackages` already receives

This breaks the abstraction: callers pass `nil` to suppress output, or a custom writer for testing, but these three paths still write to the real terminal. This makes the function untestable (tests can't capture diagnostic output) and creates inconsistent behavior where some status lines go to the caller's writer while others go to the terminal.

## Acceptance Criteria
- [x] All `os.Stderr` references in `discover.go` replaced with the `stderr` parameter
- [x] Where `stderr` may be nil and the callee doesn't accept nil, use a fallback pattern (nil → io.Discard)
- [x] New tests verify that discovery diagnostic output routes through the provided writer
- [x] Existing tests pass unchanged
- [x] `go build ./...` passes
- [x] `go test ./...` passes

## Implementation Notes
The fix is straightforward: replace `os.Stderr` with `stderr` in the three locations. The callee `discovery.Discover()` accepts a nil warnWriter (it checks for nil before writing), so nil is safe there. `MergePackage()` also accepts nil. `newOnDemandFetcher` passes it to `display.NewForWriter` and `fmt.Fprintf`, which handles nil safely (display returns a Printer that no-ops on nil writer). So the fix is a direct substitution with no nil-guard needed.

After fixing, add a test that calls `discoverAllWithConfig` with a custom writer and verifies that discovery warnings flow through it, not to os.Stderr.

## Subtasks
- [x] Fix discover.go — replace 3 os.Stderr references with stderr parameter
- [x] Add test for stderr routing in discoverAllWithConfig
- [x] Add test for stderr routing in mergePackages
- [x] Verify all tests pass
- [x] Update architecture.md

## Blocked By
