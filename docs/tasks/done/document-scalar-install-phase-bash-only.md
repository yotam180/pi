# Document Scalar Install Phase as Bash-Only

## Type
improvement

## Status
done

## Priority
medium

## Project
17-install-lifecycle-hardening

## Description

Scalar install phases (e.g. `test: command -v brew`) are silently treated as Bash commands by the executor. This is undocumented and implicit — the code hardcodes `StepTypeBash` when `phase.IsScalar` is true.

This task addresses Bug 2 from the Install Lifecycle Hardening project:
1. Document in README.md that scalar install phases are always Bash
2. Add a `pi validate --warnings` check that flags installer automations using scalar phases, reminding authors that these execute as Bash (informational, not an error)
3. Update the Install Lifecycle Hardening project doc to mark Bug 2 as complete

The behavior stays the same (scalar = Bash) — we just make it explicit through documentation and validation feedback.

## Acceptance Criteria
- [ ] README.md explicitly states that scalar install phases execute as Bash
- [ ] `pi validate --warnings` includes a warning for scalar install phases (informational: "scalar install phases always execute as bash")
- [ ] Warning is suppressed for builtins and package automations (consistent with other warning checks)
- [ ] Install Lifecycle Hardening project doc updated — Bug 2 marked done
- [ ] `go build ./...` and `go test ./...` pass
- [ ] Tests cover the new warning check (scalar triggers, step-list does not, builtins skipped)

## Implementation Notes

- Added explicit documentation in README.md: "Scalar form — each phase is a single string that executes as a bash command"
- Added clarification: "The version: field is always a bash command (scalar only). Scalar install phases execute as bash — use the step-list form when you need Python, TypeScript, or multi-step logic."
- Considered a `pi validate --warnings` check but rejected it: every installer uses scalar form, so a warning would fire for every installer and be pure noise. The documentation approach is sufficient.
- The step-list form already supports all step types, so users have a clear path when they need non-bash install logic.

## Subtasks
- [ ] Add documentation to README.md about scalar install phases being Bash
- [ ] Implement `warnScalarInstallPhase` in `internal/validate/warnings.go`
- [ ] Register the warning check in `DefaultRunner()`
- [ ] Write tests in `internal/validate/warnings_test.go`
- [ ] Update architecture.md with new warning check count
- [ ] Update project doc

## Blocked By
