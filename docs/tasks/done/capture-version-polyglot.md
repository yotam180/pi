# Fix captureVersion to Use the Runner Registry

## Type
bug

## Status
done

## Priority
medium

## Project
17-install-lifecycle-hardening

## Description

`captureVersion` in `executor/install.go` always runs the `version:` command through the Bash runner, regardless of what the automation uses. This contradicts the polyglot story — if you write a Python-based installer and want to emit the version from a Python script, you can't. The fix is to run the version command through the Bash runner by default (which is fine for shell commands like `brew --version | head -1`), but acquire that runner via `e.registry().Get(automation.StepTypeBash)` rather than by constructing an anonymous step with a hardcoded type.

The deeper fix — allowing `version:` to be a typed step (e.g. `version: {python: get_version.py}`) — requires a YAML schema change and is out of scope here. The minimum fix just ensures the code is registry-aware and doesn't hardcode a type constant.

**Current:**
```go
func (e *Executor) captureVersion(a *automation.Automation, versionCmd string, inputEnv []string) string {
    step := automation.Step{Type: automation.StepTypeBash, Value: versionCmd} // hardcoded
    runner := e.registry().Get(automation.StepTypeBash)                       // hardcoded
```

**Target (minimum fix):**
```go
// Use the registry to get the default runner for version capture.
// The version: field is always a shell command today, so Bash is correct,
// but we go through the registry to avoid hardcoding type knowledge here.
const versionStepType = automation.StepTypeBash
runner := e.registry().Get(versionStepType)
if runner == nil {
    return ""
}
step := automation.Step{Type: versionStepType, Value: versionCmd}
```

This is a one-line change in substance but documents the intent and keeps the dependency on the registry, not on the type constant.

## Acceptance Criteria
- [x] `captureVersion` uses `e.registry()` to resolve the runner
- [x] No new hardcoded step type constants in `captureVersion` (comment documents the current limitation)
- [x] `go build ./...` passes
- [x] `go test ./...` passes

## Implementation Notes

**Fix applied:** Refactored `captureVersion` to define a `const versionStepType` and resolve the runner via `e.registry().Get(versionStepType)` before constructing the step. The runner lookup now happens before the step is created (fail-fast), and the intent is documented with a comment explaining the current limitation and future path.

The registry lookup is now the single dependency — no hardcoded step type constant is scattered across multiple locations. The semantic change is minor (registry.Get was already called with the same type), but the code now explicitly uses a named constant and resolves the runner first, making the dependency on the registry clear and future changes (e.g., supporting typed version steps) localized to one constant.

**Tests added:**
- `TestCaptureVersion_UsesRegistry` — verifies version capture returns the correct trimmed output
- `TestCaptureVersion_EmptyCommand` — verifies empty version command returns empty string
- `TestCaptureVersion_CommandFails` — verifies failed version command returns empty string

## Subtasks
- [x] Update `captureVersion` to use the registry
- [x] Add comment explaining the current limitation and future path

## Blocked By
