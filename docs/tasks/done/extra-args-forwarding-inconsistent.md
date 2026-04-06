# Extra args via -- are consistently forwarded via PI_ARGS

## Type
bug

## Status
done

## Priority
medium

## Project
standalone

## Description
When passing extra arguments after `--` to `pi run`, the behavior was inconsistent between automations that have `inputs:` defined and those that don't.

### Root cause
- Automations WITHOUT `inputs:`: args were passed as bash positional params (`$1`, `$@`) but not as an env var. If the script didn't reference `$@`, the args were silently unused.
- Automations WITH `inputs:`: args were consumed by `ResolveInputs()` and mapped to input env vars. Excess positional args already error (from prior fix).

### Fix
PI now sets `PI_ARGS` environment variable containing all extra args (space-joined) for automations without inputs. This gives a clean, documented, and discoverable way to access forwarded args alongside the existing `$@`/`$1` mechanism.

## Acceptance Criteria
- [x] `pi run <automation> -- <args>` behaves consistently whether or not the automation has `inputs:`
- [x] When args are forwarded, input defaults are still applied (not dropped)
- [x] Behavior is documented in `pi run --help`

## Implementation Notes

### Approach: PI_ARGS environment variable

Added `PI_ARGS` env var injection in `RunWithInputs()` (executor.go). When args are passed to an automation without `inputs:`, `PI_ARGS` is set to the space-joined args string. This is injected alongside other input env vars so it's available in all step types.

For automations WITH `inputs:`, args are consumed by `ResolveInputs()` (which sets `args = nil`), so `PI_ARGS` is not set — the args are already available as `PI_IN_*` env vars.

### What changed
- `internal/executor/executor.go` — added `PI_ARGS` env var injection after input resolution
- `internal/cli/run.go` — updated `pi run --help` to document args forwarding behavior
- `internal/executor/inputs_test.go` — 4 new tests: PI_ARGS set, not set when no args, not set when inputs consume args, single arg
- `internal/cli/run_test.go` — 1 new CLI-level test
- `tests/integration/inputs_test.go` — 3 new integration tests: forwarded, single arg, empty when no args

### Design rationale
- `PI_ARGS` is more discoverable than `$@` — it follows the `PI_*` convention already used by `PI_IN_*`
- Both mechanisms coexist: `$@`/`$1` for positional bash patterns, `PI_ARGS` for "append all extra args"
- The env var approach works across all step types (bash, python, typescript), not just bash

## Subtasks
- [x] Investigate current arg handling behavior
- [x] Add PI_ARGS env var injection in executor
- [x] Update pi run --help text
- [x] Write executor unit tests (4 tests)
- [x] Write CLI unit test (1 test)
- [x] Write integration tests (3 tests)
- [x] Full test suite passes
- [x] Manual QA

## Blocked By
