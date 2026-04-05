# Automation-Level `env:` Block

## Type
feature

## Status
done

## Priority
medium

## Project
12-yaml-ergonomics

## Description
Currently `env:` is only supported at the step level. If multiple steps in an automation need the same environment variables, each step must repeat the `env:` block independently. This is tedious and error-prone.

Add support for a top-level `env:` block in automation files. Variables declared there apply to all steps in the automation. Step-level `env:` vars override automation-level vars with the same name (step wins over automation).

```yaml
description: Cross-compile fzf for Linux amd64

env:
  GOOS: linux
  GOARCH: amd64
  CGO_ENABLED: "0"

steps:
  - bash: go build -o target/fzf-linux ./...
  - bash: sha256sum target/fzf-linux > target/fzf-linux.sha256
  - bash: echo "Done. Binary at target/fzf-linux"
```

Automation-level env vars are scoped to the automation — they do not bleed into sub-automations called via `run:` steps (those inherit the process environment, not the parent automation's declared env).

## Acceptance Criteria
- [x] Top-level `env:` in an automation file applies to all steps
- [x] Step-level `env:` overrides automation-level `env:` for the same key (step wins)
- [x] Automation-level env does not propagate into `run:` sub-automations
- [x] Works correctly with `install:` block steps (test, run, verify) as well as `steps:`
- [x] `pi info` shows automation-level env vars
- [x] The single-step shorthand (task 61) also supports automation-level `env:`
- [x] Tests cover: basic inheritance, step override, no bleed into sub-automations

## Implementation Notes

### Approach
- Added `Env map[string]string` field to `Automation` struct
- In `UnmarshalYAML`, top-level `env:` is always stored on `a.Env` (automation-level) — both for multi-step and shorthand automations. Previously, shorthand mapped `env:` to the step's env; now it maps to automation env for consistency.
- `buildEnv()` in `helpers.go` now takes three env layers: `inputEnv`, `automationEnv`, `stepEnv` — merged in that order. Step env is appended last so it overrides automation env for the same key (last-writer-wins in exec).
- `BuildEnv` function pointer on `RunContext` updated to 3-param signature
- `runStepCommand()` calls `ctx.BuildEnv(ctx.InputEnv, ctx.Automation.Env, ctx.Step.Env)` — the `Automation.Env` comes from the automation struct on the RunContext, naturally scoped per-automation. When a `run:` step invokes another automation, that automation's own `Env` is used (or nil if none declared).
- `pi info` shows `Env:` line with sorted key names when automation-level env is present
- Install phases (`install.go`) pass `nil` for automationEnv in `buildEnv()` calls (install scalar commands don't have access to the Automation struct directly; install phase step dispatch goes through `newRunContext` which does pass `Automation.Env`)

### Files changed
- `internal/automation/automation.go` — added `Env` field, changed shorthand to store env on automation instead of step
- `internal/executor/helpers.go` — `buildEnv()` now accepts `automationEnv` parameter
- `internal/executor/runner_iface.go` — `BuildEnv` function signature updated
- `internal/executor/runners.go` — `runStepCommand()` passes `ctx.Automation.Env`
- `internal/executor/install.go` — updated `buildEnv()` calls with extra nil parameter
- `internal/cli/info.go` — added `Env:` display for automation-level env
- `internal/automation/automation_test.go` — 4 new tests, 1 updated test
- `internal/executor/step_env_test.go` — 5 new tests, updated existing buildEnv tests
- `internal/cli/info_test.go` — 1 new test
- `tests/integration/step_env_integ_test.go` — 6 new integration tests
- `tests/integration/shorthand_integ_test.go` — updated to match new behavior
- `examples/step-env/.pi/` — 3 new example automations
- `docs/README.md` — new section documenting automation-level env
- `docs/architecture.md` — updated with new design decision section and test counts

## Subtasks
- [x] Update automation schema to include top-level `env:` field
- [x] Update executor to merge automation env into each step's env (with step-level override)
- [x] Update `pi info` output
- [x] Add tests
- [x] Add example automation using automation-level env

## Blocked By
