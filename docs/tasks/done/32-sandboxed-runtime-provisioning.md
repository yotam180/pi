# Sandboxed Runtime Provisioning

## Type
feature

## Status
done

## Priority
medium

## Project
05-environment-robustness

## Description
When enabled in `pi.yaml`, PI provisions missing runtimes into `~/.pi/runtimes/<runtime>/<version>/` instead of erroring. This makes automations work on fresh machines without requiring a prior `pi setup`.

### Configuration in `pi.yaml`
```yaml
runtimes:
  provision: auto     # auto | ask | never (default: never)
  manager: mise       # mise | direct (default: mise if available, else direct)
```

### Provisioning modes
1. `provision: never` (default) — error with install hint, do nothing
2. `provision: ask` — prompt the developer, provision if confirmed
3. `provision: auto` — silently provision, print one line: `[provisioned] node 20.11.0 → ~/.pi/runtimes/node/20.11.0/`

### Provisioning backends
- `manager: mise` — calls `mise install node@20` and prepends the mise shim path
- `manager: direct` — downloads pre-built binary from official release CDN for the current OS/arch (node and python only)

Provisioned runtimes are used only for PI step execution — they are never added to the system PATH permanently. The executor prepends the provisioned runtime's bin directory to PATH only for the duration of the step execution.

### Out of scope
- Provisioning for `command:` requirements (only runtimes: node, python)
- Rollback if a provisioned runtime fails mid-install
- Network-isolated / air-gapped provisioning

## Acceptance Criteria
- [x] `runtimes:` block parsed from `pi.yaml` into config
- [x] `provision: never` (default) — produces error with install hint
- [x] `provision: auto` — provisions runtime into `~/.pi/runtimes/` and runs step
- [x] `provision: ask` — prompts user (or skips in non-interactive/CI mode)
- [x] `manager: mise` — calls mise to install runtime
- [x] `manager: direct` — downloads from official CDN (node, python)
- [x] Provisioned runtimes isolated to `~/.pi/runtimes/`, not on system PATH
- [x] Executor prepends provisioned runtime bin to PATH for step execution only
- [x] Unit tests for provisioning logic
- [x] Integration test: provision python on a fresh system, run a python step
- [x] `go test ./...` passes

## Implementation Notes

### Architecture
- New package `internal/runtimes/` contains the provisioning logic
- `Provisioner` struct with `Mode`, `Manager`, `BaseDir` fields
- `config.ProjectConfig` extended with `Runtimes *RuntimesConfig` field
- `RuntimesConfig` has `Provision` (never/ask/auto) and `Manager` (mise/direct) fields
- `EffectiveProvisionMode()` and `EffectiveRuntimeManager()` provide defaults
- `Executor` extended with `Provisioner` and `runtimePaths` fields
- `ValidateRequirements()` enhanced: failed runtime requirements trigger `tryProvision()` when provisioner is configured
- All step execution methods now use `buildEnv()` instead of `appendInputEnv()` to inject provisioned PATH
- `prependPathInEnv()` modifies the PATH entry in env slice to prepend provisioned bin dirs
- CLI layer (`run.go`, `setup.go`) loads config and creates provisioner when mode != never

### Mise backend
- Calls `mise install <runtime>@<version>`, then `mise where <spec>` to find install path
- Symlinks binaries from mise install dir into `~/.pi/runtimes/<name>/<version>/bin/`
- Falls back to direct download when mise is not installed

### Direct backend
- Node: downloads from `https://nodejs.org/dist/v<version>/node-v<version>-<platform>-<arch>.tar.gz`
- Python: uses python-build-standalone releases from astral-sh/python-build-standalone GitHub repo
- Extracts into temp dir, moves bin/ to managed location

### Testing
- 16 unit tests in `internal/runtimes/runtimes_test.go` covering all modes and edge cases
- 11 new tests in `internal/executor/validate_test.go` for provisioning integration and PATH construction
- 6 new config tests for RuntimesConfig parsing and validation
- 7 integration tests in `tests/integration/examples_test.go` with two example workspaces

## Subtasks
- [x] Parse `runtimes:` block in config
- [x] Implement runtime provisioning manager interface
- [x] Implement mise backend
- [x] Implement direct download backend (node, python CDN URLs)
- [x] Integrate provisioning into requirement validation flow
- [x] PATH scoping for provisioned runtimes during step execution
- [x] Write unit and integration tests

## Blocked By
30-pre-execution-validation (done)
