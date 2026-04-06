# Install Lifecycle Hardening

## Status
todo

## Priority
medium

## Description

Three concrete correctness problems in the install lifecycle (`install:` block execution): version capture always runs through Bash regardless of step type, the scalar shorthand silently assumes Bash, and `first:` blocks inside install phases don't record their output to `stepOutputs`, causing stale `outputs.last` values. These are independent bugs that can be fixed without the larger architectural work.

## Goals

- `captureVersion` dispatches to the runner registered for the step type specified, not hardcoded Bash.
- The scalar shorthand for install phases is clearly typed (or documented) as Bash.
- `execInstallFirstBlock` records the matched sub-step's stdout to `e.stepOutputs`, consistent with all other step execution paths.

## Background & Context

**Bug 1 — `captureVersion` hardcodes Bash:**
```go
func (e *Executor) captureVersion(...) string {
    step := automation.Step{Type: automation.StepTypeBash, Value: versionCmd}
    runner := e.registry().Get(automation.StepTypeBash) // always Bash
```
If someone writes `version: node --version` in their installer, that's fine because it's shell-compatible. But if they want a Python script to emit the version string, the field literally can't be used that way. The fix: parse `versionCmd` as a step type + value pair (similar to how install phase scalar works), or at minimum consult the runner registry using the automation's primary step type.

**Bug 2 — Scalar `InstallPhase` hardcodes Bash:**
```go
if phase.IsScalar {
    steps = []automation.Step{{Type: automation.StepTypeBash, Value: phase.Scalar}}
}
```
Every scalar like `test: command -v brew` is executed as Bash. This is undocumented and implicit. The fix is to at minimum document it clearly in the YAML schema docs, and ideally extend the scalar form to accept an explicit type prefix (`test: {bash: "command -v brew"}`) for future-proofing.

**Bug 3 — `execInstallFirstBlock` drops stdout:**
```go
ctx := e.newRunContext(a, sub, nil, io.Discard, nil, inputEnv, e.RepoRoot)
// ^ stdout is io.Discard — never recorded
return runner.Run(ctx)
```
Compare with `execInstallPhaseWithStderr` which always calls `e.recordOutput(outputCapture.String())`. A `first:` block inside an install phase leaves `stepOutputs` stale.

## Scope

### In scope
- Fix `captureVersion` to use the runner registry correctly (bash default is fine if the field is a plain shell command, but must not hardcode the type)
- Fix `execInstallFirstBlock` to capture stdout and call `e.recordOutput()`
- Document (or add a parse-time error) for scalar install phase being Bash-only

### Out of scope
- Full polyglot install phase scalar (that requires schema changes, is a future feature)
- Changes to how `install:` blocks are parsed beyond clarifying documentation

## Success Criteria

- `captureVersion` passes `ctx.Step.Type` to the registry lookup, falling back to Bash only when the type is unset
- `execInstallFirstBlock` captures and records the matched sub-step's stdout
- A test confirms that `outputs.last` after a `first:` block in an install phase reflects the block's output
- `go build ./...` and `go test ./...` pass

## Notes

These are surgical fixes. Don't over-engineer. The captureVersion fix in particular should be minimal — the `version:` YAML field is a string today; making it a typed step is a schema change. For now: run it through Bash by default, but use `e.registry()` to get the bash runner (rather than re-creating it), so the dependency is correct.

## Progress

- [x] **Bug 1 — `captureVersion` hardcodes Bash:** Fixed. Now uses a `const versionStepType` resolved via `e.registry()`. Comment documents the limitation and future path. Task: `capture-version-polyglot.md` (done).
- [x] **Bug 3 — `execInstallFirstBlock` drops stdout:** Fixed. Now captures stdout into `bytes.Buffer` and calls `e.recordOutput()`. Task: `install-first-block-output-recording.md` (done).
- [x] **Bug 2 — Scalar `InstallPhase` hardcodes Bash:** Documented explicitly in README.md. Scalar install phases (`test:`, `run:`, `verify:` as strings) now clearly state they execute as bash commands. The step-list form is documented as the path for non-bash logic. A `pi validate --warnings` check was considered but rejected as too noisy (every installer uses scalar form). Task: `document-scalar-install-phase-bash-only.md` (done).
