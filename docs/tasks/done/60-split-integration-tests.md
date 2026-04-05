# Split Integration Test File (examples_test.go)

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description
The `tests/integration/examples_test.go` file has grown to 2450 lines with 130+ test functions covering many distinct feature areas (basic, docker, pipe, version, inputs, info, conditionals, builtins, installers, requires, doctor, runtime provisioning, step-env, step-visibility, parent-shell, step-dir, step-timeout, step-description, validate). This makes it hard to navigate, adds cognitive overhead, and makes parallel development on different feature areas error-prone.

Split `examples_test.go` into domain-specific test files following the pattern already established by `polyglot_test.go` and `shell_test.go` in the same package.

## Acceptance Criteria
- [x] `examples_test.go` is split into ~10-15 focused test files by feature domain
- [x] No test logic is changed — only file boundaries move
- [x] Shared infrastructure (`TestMain`, `runPi`, `runPiStdout`, `runPiSplit`, etc.) stays in a shared file
- [x] `go test ./tests/integration/...` passes with the same results before and after
- [x] Each new file has a clear name matching its domain (e.g., `conditionals_test.go`, `builtins_test.go`)
- [x] Architecture docs updated with new file structure

## Implementation Notes
### Approach
Renamed `examples_test.go` to `main_test.go` (contains TestMain and shared helpers only), then created 19 domain-specific test files. Used `_integ_test.go` suffix for files whose names would collide with unit test files in the executor package (step_env, step_dir, etc.). No test logic was changed — only file boundaries moved.

### Test groupings identified:
- **basic_test.go**: TestBasic_* (7 tests)
- **docker_test.go**: TestDocker_* (6 tests)
- **pipe_test.go**: TestPipe_* (3 tests)
- **version_test.go**: TestVersion_* (3 tests)
- **inputs_test.go**: TestInputs_* (8 tests)
- **info_test.go**: TestInfo_* (4 tests)
- **conditionals_test.go**: TestConditional_* (~20 tests)
- **builtins_test.go**: TestBuiltins_* (~25 tests)
- **installer_schema_test.go**: TestInstallerSchema_* (11 tests)
- **requires_test.go**: TestRequiresValidation_* (7 tests)
- **doctor_test.go**: TestDoctor_* (7 tests)
- **runtime_provisioning_test.go**: TestRuntimeProvisioning_* (6 tests)
- **step_env_test.go**: TestStepEnv_* (5 tests)
- **step_visibility_test.go**: TestStepVisibility_* (6 tests)
- **parent_shell_test.go**: TestParentShell_* (7 tests)
- **step_dir_test.go**: TestStepDir_* (6 tests)
- **step_timeout_test.go**: TestStepTimeout_* (5 tests)
- **step_description_test.go**: TestStepDescription_* (5 tests)
- **validate_test.go**: TestValidate_* (5 tests)

Shared helpers (`TestMain`, `findRepoRoot`, `examplesDir`, `runPi`, `runPiStdout`, `runPiSplit`) move to existing `helpers_test.go`.

## Subtasks
- [x] Move shared helpers to helpers_test.go
- [x] Create each domain test file
- [x] Remove examples_test.go
- [x] Verify all tests pass
- [x] Update docs

## Blocked By
None
