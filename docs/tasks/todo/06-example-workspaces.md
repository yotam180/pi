# Example Workspaces

## Type
infra

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Create the `examples/` folder with realistic sample workspaces that demonstrate PI usage and serve as the project's integration tests. These should not reference Vyper internals — they must be self-contained scenarios any developer can understand and run.

## Acceptance Criteria
- [ ] `examples/` folder exists at repo root with at least 2 workspaces
- [ ] Each workspace has a `pi.yaml`, a `.pi/` folder with automations, and a `README.md` explaining the scenario
- [ ] `examples/basic/` — demonstrates: inline bash step, a `.sh` file step, and `run:` step chaining
- [ ] `examples/docker-project/` — demonstrates: a realistic multi-step Docker Compose workflow (up, down, logs, build-then-up) using only bash steps
- [ ] Both workspaces can be run end-to-end with `pi run <automation>` from within the workspace directory
- [ ] An integration test script (or Go test using `exec.Command`) runs `pi run` against each workspace and asserts correct exit codes
- [ ] No references to Vyper, vyper-platform, or internal tooling anywhere in examples/

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `examples/basic/` workspace
- [ ] Create `examples/docker-project/` workspace
- [ ] Write README for each workspace
- [ ] Write integration test that runs `pi run` against each workspace

## Blocked By
05-pi-run-and-pi-list-commands
