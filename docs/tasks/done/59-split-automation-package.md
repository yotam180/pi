# Split automation package into focused files

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
The `internal/automation/automation.go` file is 661 lines and mixes several distinct concerns:
- Requirements parsing (`Requirement`, `requirementRaw`, `parseNameVersion`, `validateVersionString`) — ~145 lines
- Step types and install spec (`StepType`, `Step`, `stepRaw`, `InstallPhase`, `InstallSpec`, `toStep()`) — ~230 lines
- Input handling (`InputSpec`, `inputsRaw`, `ResolveInputs`, `InputEnvVars`) — ~100 lines
- Core automation struct, YAML unmarshalling, `Load`/`LoadFromBytes`, validation — ~190 lines

Similarly, `automation_test.go` is 2352 lines covering all these areas in a single file.

Split both into focused files that mirror the source structure:
- `requirements.go` / `requirements_test.go` — requirement types and parsing
- `step.go` / `step_test.go` — step types, install spec, step parsing
- `inputs.go` / `inputs_test.go` — input spec, resolution, env vars
- `automation.go` / `automation_test.go` — core automation struct, load, validation (reduced)

This improves readability, makes it easier to extend any single area without touching the others, and aligns with Go best practices for package organization.

## Acceptance Criteria
- [x] `requirements.go` contains all requirement-related types, parsing, and validation
- [x] `step.go` contains StepType, Step, stepRaw, InstallPhase, InstallSpec, toStep()
- [x] `inputs.go` contains InputSpec, inputsRaw, ResolveInputs, InputEnvVars
- [x] `automation.go` retains core Automation struct, Load, LoadFromBytes, validate
- [x] Test files mirror the source file split
- [x] `go build ./...` passes
- [x] `go test ./...` passes with same test count
- [x] `go vet ./...` passes
- [x] Architecture docs updated

## Implementation Notes
- All files stay in `package automation` — this is purely a file-level reorganization
- No API changes, no renames — existing imports remain valid
- All types are in the same package so no circular dependency concerns
- The shared `writeFile` test helper and `boolPtr` helper stay in `automation_test.go` since they're used across test files
- Validation functions (`validateSteps`, `validateInstall`, `validateInstallPhase`) moved to `step.go` alongside the types they validate
- Source files: automation.go (163) + step.go (267) + inputs.go (118) + requirements.go (139) = 687 total (was 661 in one file, slightly larger due to separate import blocks)
- Test files: automation_test.go (14) + step_test.go (53) + inputs_test.go (16) + requirements_test.go (20) = 103 tests (unchanged)
- All 690 project tests pass after the split

## Subtasks
- [x] Create `requirements.go` with requirement types and parsing
- [x] Create `step.go` with step types, install spec
- [x] Create `inputs.go` with input handling
- [x] Trim `automation.go` to core struct
- [x] Split test file to match
- [x] Verify all tests pass
- [x] Update architecture docs

## Blocked By
