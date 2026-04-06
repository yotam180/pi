# Staticcheck Cleanup

## Type
chore

## Status
in_progress

## Priority
medium

## Project
standalone

## Description
Run staticcheck and fix all reported issues across the codebase. Issues found:
1. ST1005 in `internal/cli/new.go:101` — error string ends with punctuation
2. S1009 in `internal/config/config_test.go:81` — unnecessary nil check before len()
3. S1009 in `internal/discovery/discovery.go:344` — unnecessary nil check before len()
4. ST1005 in `internal/discovery/discovery.go:366` — error string ends with newline (from `b.String()`)
5. U1000 in `internal/executor/validate.go:84` — unused func `checkRequirementImpl`
6. U1000 in `tests/integration/helpers_test.go:15` — unused func `requireNode`
7. U1000 in `tests/integration/on_demand_test.go:179` — unused func `createOnDemandCacheDir`

## Acceptance Criteria
- [x] All 7 staticcheck issues resolved
- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] `staticcheck ./...` clean

## Implementation Notes
- ST1005: Go convention — error strings should not be capitalized or end with punctuation. The `new.go` error wraps with `%w\n\nRun 'pi init'...` ending with `.` — changed to no trailing period.
- S1009: `len()` on a nil map is 0, so `if m != nil && len(m) > 0` simplifies to `if len(m) > 0`.
- U1000: `checkRequirementImpl` was a legacy thin wrapper that became unused after reqcheck extraction. Removed. `requireNode` and `createOnDemandCacheDir` are dead test helpers — removed.
- The ST1005 on discovery.go:366 is from `fmt.Errorf("%s", b.String())` — the builder content ends with `\n`. Fixed by trimming trailing newline from the builder output.

## Subtasks
- [x] Fix ST1005 in cli/new.go
- [x] Fix S1009 in config_test.go
- [x] Fix S1009 and ST1005 in discovery.go
- [x] Remove unused checkRequirementImpl in executor/validate.go
- [x] Remove unused requireNode in integration helpers
- [x] Remove unused createOnDemandCacheDir in on_demand_test.go
- [x] Verify clean build + tests + staticcheck

## Blocked By
