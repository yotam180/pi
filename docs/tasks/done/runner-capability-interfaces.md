# Runner Capability Interfaces

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The `Registry` in `internal/executor/runner_iface.go` uses `runner.(*SubprocessRunner)` type assertions to query runner capabilities (`FileExtForStepType` and `StepTypeSupportsParentShell`). This breaks the `StepRunner` abstraction — any custom runner that isn't a `SubprocessRunner` silently loses these capabilities even if it logically supports them.

**The fix:** Introduce optional capability interfaces following Go's composition idiom. Runners that support file extensions implement `FileExtProvider`. Runners that support parent_shell implement `ParentShellCapable`. The Registry queries these interfaces instead of asserting concrete types.

This is a non-breaking refactor: `SubprocessRunner` implements the new interfaces (satisfying them via methods that read from `SubprocessConfig`), and the public API of `Registry` remains identical. External consumers see no behavior change — the improvement is purely in extensibility.

**Current code (type assertion):**
```go
func (r *Registry) FileExtForStepType(stepType automation.StepType) string {
    runner := r.runners[stepType]
    if sr, ok := runner.(*SubprocessRunner); ok {
        return sr.Config.FileExt
    }
    return ""
}
```

**Target code (interface):**
```go
type FileExtProvider interface {
    FileExt() string
}

func (r *Registry) FileExtForStepType(stepType automation.StepType) string {
    runner := r.runners[stepType]
    if p, ok := runner.(FileExtProvider); ok {
        return p.FileExt()
    }
    return ""
}
```

## Acceptance Criteria
- [x] `FileExtProvider` interface defined in `runner_iface.go`
- [x] `ParentShellCapable` interface defined in `runner_iface.go`
- [x] `SubprocessRunner` implements both interfaces
- [x] `Registry.FileExtForStepType()` uses `FileExtProvider` instead of type assertion
- [x] `Registry.StepTypeSupportsParentShell()` uses `ParentShellCapable` instead of type assertion
- [x] No `*SubprocessRunner` type assertions remain in `runner_iface.go`
- [x] All existing tests pass with zero changes
- [x] New test verifying a custom (non-SubprocessRunner) runner can declare capabilities
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach
- Defined two small optional interfaces (`FileExtProvider`, `ParentShellCapable`) in `runner_iface.go`
- Added `FileExt()` and `SupportsParentShell()` methods to `SubprocessRunner` in `runners.go` — these read from `SubprocessConfig` fields that already exist
- Updated `Registry.FileExtForStepType()` and `Registry.StepTypeSupportsParentShell()` to query the interfaces via Go interface assertion instead of concrete type assertion
- No behavior change for existing code: `SubprocessRunner` satisfies both interfaces, so all existing callers see identical results

### Tests added
4 new tests in `coverage_gaps_test.go`:
- `TestRegistry_CapabilityInterfaces_CustomRunner` — custom non-SubprocessRunner implementing both interfaces
- `TestRegistry_CapabilityInterfaces_PlainRunner` — runner implementing neither interface returns zero values
- `TestRegistry_CapabilityInterfaces_PartialImplementation` — runner with FileExt but SupportsParentShell=false
- `TestRegistry_CapabilityInterfaces_UnregisteredType` — unregistered step type returns zero values

### Why this matters
Before this change, adding a custom step runner (e.g. for a new language) required either extending `SubprocessRunner`'s config or accepting that `FileExtForStepType` and `StepTypeSupportsParentShell` would silently return wrong defaults. Now, any `StepRunner` implementation can declare these capabilities by implementing the interfaces.

## Subtasks
- [ ] Define `FileExtProvider` interface
- [ ] Define `ParentShellCapable` interface
- [ ] Add `FileExt()` and `SupportsParentShell()` methods to `SubprocessRunner`
- [ ] Update `Registry.FileExtForStepType` to use interface
- [ ] Update `Registry.StepTypeSupportsParentShell` to use interface
- [ ] Add test with a custom runner implementing the capability interfaces
- [ ] Verify all existing tests pass unchanged
- [ ] Update architecture.md

## Blocked By
