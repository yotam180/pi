# Deduplicate install phase execution paths in executor/install.go

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

`install.go` contains four functions with heavily duplicated logic:

1. `execInstallPhaseLive()` — runs install phase steps with stdout suppressed but stderr live
2. `execInstallPhaseCapture()` — runs install phase steps with both stdout and stderr suppressed (stderr optionally captured)
3. `execInstallFirstBlockLive()` — runs first: blocks in install phases with stderr live
4. `execInstallFirstBlock()` — runs first: blocks in install phases with stderr suppressed

Functions 1 and 2 differ only in how stderr is wired: live always sends stderr to `e.stderr()`, while capture sends it to either a capture buffer or `io.Discard`. Both share the same step iteration, condition evaluation, first-block dispatch, registry lookup, output capture, and output recording logic.

Functions 3 and 4 differ only in how stderr is wired, identical to the live/capture distinction above.

The fix is to unify each pair into a single function parameterized by an I/O config (where to send stdout and stderr), eliminating ~80 lines of duplication while maintaining the exact same behavior.

## Acceptance Criteria
- [x] `execInstallPhaseLive()` and `execInstallPhaseCapture()` consolidated into a single function
- [x] `execInstallFirstBlockLive()` and `execInstallFirstBlock()` consolidated into a single function  
- [x] All existing tests pass without modification (`go test ./...`)
- [x] `go build ./...` succeeds
- [x] No behavioral changes — same output, same error messages, same I/O routing
- [x] Architecture docs updated

## Implementation Notes

### Analysis

Comparing the two phase execution functions:

**execInstallPhaseLive** (lines 58-101):
- Scalar→step conversion: identical
- stepOutputs save/restore: identical
- Step iteration with if: eval: identical
- first: block dispatch: calls `execInstallFirstBlockLive` (live stderr)
- Regular step: stdout → `&outputCapture`, stderr → `e.stderr()` (live)
- Output recording: identical

**execInstallPhaseCapture** (lines 105-153):
- Scalar→step conversion: identical
- stepOutputs save/restore: identical
- Step iteration with if: eval: identical
- first: block dispatch: calls `execInstallFirstBlock` (captured stderr)
- Regular step: stdout → `&outputCapture`, stderr → `stderrCapture` or `io.Discard`
- Output recording: identical

The difference is purely the stderr writer. Both can be unified with a single `stderrWriter io.Writer` parameter.

Same for the first-block pair — they only differ in whether stderr goes to `e.stderr()` or the capture buffer.

### Approach (plan)

1. Create `execInstallPhaseWith(a, phase, inputEnv, stderrWriter)` that takes an explicit stderr target
2. `execInstallPhaseLive` becomes: `execInstallPhaseWith(a, phase, inputEnv, e.stderr())`
3. `execInstallPhaseCapture` becomes: `execInstallPhaseWith(a, phase, inputEnv, stderrTarget)` where stderrTarget is the capture buffer or `io.Discard`
4. Same approach for the first-block functions
5. Keep the caller-facing API names for readability (thin wrappers or inline at call sites)

### What was done

1. **Unified first-block functions**: `execInstallFirstBlockLive` and `execInstallFirstBlock` merged into a single `execInstallFirstBlock(a, step, index, inputEnv, stderrWriter io.Writer)`. The callers pass `e.stderr()` for live or a buffer/`io.Discard` for suppressed.

2. **Unified phase functions**: `execInstallPhaseLive` and `execInstallPhaseCapture` merged into `execInstallPhaseWithStderr(a, phase, inputEnv, stderrWriter io.Writer)`. `execInstallPhase()` and `execInstallPhaseLive()` kept as thin one-line wrappers for API clarity. `execInstallPhaseCapture()` removed — the one caller (verify phase in `execInstall`) now uses `execInstallPhase()` directly.

3. **Net result**: Removed ~80 lines of duplicated logic. `install.go` went from 241 lines to 165 lines. The `stderrCapture *bytes.Buffer` parameter on `execInstallPhaseCapture` was only ever called with `nil`, so it was effectively always `io.Discard` — simplified accordingly.

4. **No test changes**: All 1355 tests pass without modification, confirming behavioral equivalence.

## Subtasks
- [x] Unify `execInstallFirstBlockLive` + `execInstallFirstBlock`
- [x] Unify `execInstallPhaseLive` + `execInstallPhaseCapture`
- [x] Remove dead code
- [x] Run full test suite
- [x] Update docs

## Blocked By
