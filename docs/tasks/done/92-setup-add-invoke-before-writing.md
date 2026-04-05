# setup add: invoke the automation before writing to pi.yaml

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
`pi setup add` currently writes the entry to `pi.yaml` immediately without ever running the automation. This means a misconfigured call (missing required inputs, typo in flags, etc.) silently produces a broken `pi.yaml`.

The correct behaviour:
1. `pi setup add pi:install-node` — runs the automation first. If it needs `version` and none is provided, it fails with a clear message (sourced from the automation's input validation / error handling). `pi.yaml` is **not** modified.
2. `pi setup add pi:install-node --version 22` — runs the automation. If it succeeds, adds the entry to `pi.yaml`. If it fails, surfaces the actual error output and does **not** modify `pi.yaml`.
3. `pi setup add pi:install-node --version 22 --only-add` — skips execution, writes to `pi.yaml` directly (original current behaviour). Useful for CI bootstrapping or when the tool is already set up.

This also means `pi setup add` becomes the recommended way to both *install a tool* and *register it* in one step.

## Acceptance Criteria
- [x] `pi setup add pi:install-node` (missing version) fails before touching `pi.yaml`, with an error message that explains what's missing.
- [x] `pi setup add pi:install-node --version 22` runs node installation; on success writes to `pi.yaml`; on failure shows the full automation output/error and exits non-zero without touching `pi.yaml`.
- [x] `pi setup add pi:install-node --version 22 --only-add` skips execution and writes to `pi.yaml` immediately (previous behaviour).
- [x] `--only-add` flag is documented in `pi setup add --help`.
- [x] Automation stdout/stderr is streamed live to the terminal during the run (not buffered).
- [x] `go build ./...` and `go test ./...` pass.

## Implementation Notes

### Approach
Reused the existing `ProjectContext` / `Discover()` / `NewExecutor()` pipeline from `setup.go` to invoke a single automation inside `runSetupAdd`, before the `config.AddSetupEntry()` call.

### Key changes
- `internal/cli/setup_add.go`:
  - Added `--only-add` bool flag to `newSetupAddCmd()`
  - Added `onlyAdd` parameter to `runSetupAdd()`
  - New `invokeSetupAutomation()` function handles discovery, condition checking, and execution for a single setup entry
  - When `!onlyAdd`, the automation runs first; on failure, `pi.yaml` is untouched
  - `--only-add` preserves the previous write-only behavior
- Updated `--help` long description and examples to document `--only-add`
- `invokeSetupAutomation()` uses `project.FindRoot()` to locate the project root from the working directory, loads config, discovers automations (local + packages + builtins), checks `if:` conditions, and runs via `exec.RunWithInputs()`

### Test updates
- All existing unit tests (`setup_add_test.go`): updated to pass `onlyAdd: true` since they test YAML-writing behavior in isolation (no `.pi/` directory or real automations)
- 4 new unit tests:
  - `TestRunSetupAdd_InvokesBeforeWriting` — automation runs and output appears before pi.yaml write
  - `TestRunSetupAdd_InvokeFailure_NoPiYamlModification` — failing automation doesn't touch pi.yaml
  - `TestRunSetupAdd_OnlyAddSkipsExecution` — `--only-add` writes nonexistent automation without error
  - `TestRunSetupAdd_InvokeNotFound_NoPiYamlModification` — not-found automation doesn't touch pi.yaml
- All existing integration tests: updated to use `--only-add` (they add builtins like `pi:install-uv` which would actually try to install things)
- 4 new integration tests:
  - `TestSetupAdd_InvokesBeforeWriting` — creates local `.pi/greet.yaml`, verifies output + pi.yaml write
  - `TestSetupAdd_InvokeFailure_DoesNotModifyPiYaml` — creates failing automation, verifies pi.yaml untouched
  - `TestSetupAdd_OnlyAddSkipsExecution` — `--only-add` with nonexistent automation succeeds
  - `TestSetupAdd_NotFoundWithoutOnlyAdd` — nonexistent automation without `--only-add` fails

### Helper added
- `writeTestPiDir()` test helper in CLI unit tests for creating `.pi/*.yaml` files in temp dirs

## Subtasks
- [x] Add `--only-add` bool flag to `newSetupAddCmd`.
- [x] Wire single-automation invocation in `runSetupAdd` (before the write step).
- [x] Ensure automation stderr/stdout is forwarded to the caller's stderr/stdout.
- [x] Update / add integration tests in `tests/integration/setup_add_test.go`.
- [x] Update `--help` text and examples.

## Blocked By
