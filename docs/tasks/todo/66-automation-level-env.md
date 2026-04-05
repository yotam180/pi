# Automation-Level `env:` Block

## Type
feature

## Status
todo

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
- [ ] Top-level `env:` in an automation file applies to all steps
- [ ] Step-level `env:` overrides automation-level `env:` for the same key (step wins)
- [ ] Automation-level env does not propagate into `run:` sub-automations
- [ ] Works correctly with `install:` block steps (test, run, verify) as well as `steps:`
- [ ] `pi info` shows automation-level env vars
- [ ] The single-step shorthand (task 61) also supports automation-level `env:`
- [ ] Tests cover: basic inheritance, step override, no bleed into sub-automations

## Implementation Notes

## Subtasks
- [ ] Update automation schema to include top-level `env:` field
- [ ] Update executor to merge automation env into each step's env (with step-level override)
- [ ] Update `pi info` output
- [ ] Add tests
- [ ] Add example automation using automation-level env

## Blocked By
