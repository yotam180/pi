# Sandboxed Runtime Provisioning

## Type
feature

## Status
todo

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
- [ ] `runtimes:` block parsed from `pi.yaml` into config
- [ ] `provision: never` (default) — produces error with install hint
- [ ] `provision: auto` — provisions runtime into `~/.pi/runtimes/` and runs step
- [ ] `provision: ask` — prompts user (or skips in non-interactive/CI mode)
- [ ] `manager: mise` — calls mise to install runtime
- [ ] `manager: direct` — downloads from official CDN (node, python)
- [ ] Provisioned runtimes isolated to `~/.pi/runtimes/`, not on system PATH
- [ ] Executor prepends provisioned runtime bin to PATH for step execution only
- [ ] Unit tests for provisioning logic
- [ ] Integration test: provision python on a fresh system, run a python step
- [ ] `go test ./...` passes

## Implementation Notes

## Subtasks
- [ ] Parse `runtimes:` block in config
- [ ] Implement runtime provisioning manager interface
- [ ] Implement mise backend
- [ ] Implement direct download backend (node, python CDN URLs)
- [ ] Integrate provisioning into requirement validation flow
- [ ] PATH scoping for provisioned runtimes during step execution
- [ ] Write unit and integration tests

## Blocked By
30-pre-execution-validation
