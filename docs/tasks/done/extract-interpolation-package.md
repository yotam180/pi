# Extract Interpolation Package

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The `Executor` struct in `internal/executor/executor.go` has interpolation logic (`interpolateValue`, `interpolateEnv`, `interpolateWithCtx`, `interpolateWith`) and output tracking (`stepOutputs`, `lastOutput`, `recordOutput`) mixed in with step orchestration. These are self-contained concerns that can be extracted into a dedicated `internal/interpolation` package.

**Current state:**
- `interpolateValue()` resolves `outputs.last`, `outputs.<N>`, and `inputs.<name>` references in string values
- `interpolateEnv()` applies interpolation to env map values
- `interpolateWithCtx()` / `interpolateWith()` apply interpolation to `with:` map values
- `stepOutputs` / `lastOutput()` / `recordOutput()` manage the per-automation output buffer
- All of this lives on the `Executor` struct, coupling interpolation to execution

**Benefits of extraction:**
1. Interpolation logic becomes independently testable without needing an Executor
2. Reduces the Executor's responsibilities (currently ~573 lines)
3. Makes it easier to extend interpolation in the future (e.g., new reference types, templating)
4. Improves modularity for future customization features
5. OutputTracker can be reused by dry-run, install phases, or other contexts

**Proposed design:**
- New package: `internal/interpolation`
- `OutputTracker` struct: manages `stepOutputs []string`, exposes `Record(output)`, `Last()`, `Get(index)`, `Save()`/`Restore()` for scoped resets
- `ResolveValue(v string, outputs *OutputTracker, inputEnv []string)` — replaces `interpolateValue`
- `ResolveEnv(env map[string]string, outputs *OutputTracker, inputEnv []string)` — replaces `interpolateEnv`
- `ResolveWith(with map[string]string, outputs *OutputTracker, inputEnv []string)` — replaces `interpolateWithCtx`/`interpolateWith`

## Acceptance Criteria
- [x] `internal/interpolation` package exists with `OutputTracker` and `Resolve*` functions
- [x] All interpolation logic removed from `Executor` struct methods
- [x] Executor uses `interpolation.OutputTracker` and `interpolation.Resolve*` functions
- [x] Comprehensive unit tests in `internal/interpolation` with 100% coverage
- [x] All existing tests pass (`go test ./...`)
- [x] `go build ./...` passes
- [x] Architecture doc updated

## Implementation Notes

### Design decisions

1. **OutputTracker is a value type** — embedded directly in `Executor` as `Outputs interpolation.OutputTracker` (not a pointer). This avoids nil-check boilerplate and makes the zero value useful.

2. **Thin delegation wrappers retained** — `executor.go` keeps `interpolateValue()`, `interpolateEnv()`, `interpolateWithCtx()`, `interpolateWith()` as one-line wrappers that delegate to `interpolation.Resolve*`. This preserves backward compatibility with existing tests that call these methods. The wrappers are straightforward enough to not warrant removal at this stage.

3. **Snapshot/Restore pattern** — replaces the old `savedOutputs := e.stepOutputs; e.stepOutputs = nil; defer func() { e.stepOutputs = savedOutputs }()` pattern with cleaner `Snapshot()`/`Reset()`/`Restore()` calls. Used in both `RunWithInputs` and `execInstallPhaseWithStderr`.

4. **ResolveValue handles nil tracker** — When `tracker` is nil, `outputs.last` returns `""` and indexed outputs return the passthrough value. This matches the previous behavior where an empty `stepOutputs` slice produced the same results.

5. **Coverage** — interpolation package has 100% coverage (36 tests). Executor coverage shifted from 90.1% to 89.4% because the core logic moved to interpolation; combined coverage across both packages is better than before.

## Subtasks
- [x] Create `internal/interpolation/interpolation.go` with OutputTracker and Resolve* functions
- [x] Create `internal/interpolation/interpolation_test.go` with comprehensive tests
- [x] Update `executor.go` to use new package
- [x] Update `install.go` to use OutputTracker Snapshot/Reset/Restore
- [x] Update `dry_run.go` to use interpolation.ResolveEnv/ResolveWith
- [x] Update test files (outputs_test.go, coverage_gaps_test.go, install_test.go) for new API
- [x] Verify all tests pass
- [x] Update `docs/architecture.md`

## Blocked By
