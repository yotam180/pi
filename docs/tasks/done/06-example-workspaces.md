# Example Workspaces

## Type
infra

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Create the `examples/` folder with realistic sample workspaces that demonstrate PI usage and serve as the project's integration tests. These should not reference Vyper internals — they must be self-contained scenarios any developer can understand and run.

## Acceptance Criteria
- [x] `examples/` folder exists at repo root with at least 2 workspaces
- [x] Each workspace has a `pi.yaml`, a `.pi/` folder with automations, and a `README.md` explaining the scenario
- [x] `examples/basic/` — demonstrates: inline bash step, a `.sh` file step, and `run:` step chaining
- [x] `examples/docker-project/` — demonstrates: a realistic multi-step Docker Compose workflow (up, down, logs, build-then-up) using only bash steps
- [x] Both workspaces can be run end-to-end with `pi run <automation>` from within the workspace directory
- [x] An integration test script (or Go test using `exec.Command`) runs `pi run` against each workspace and asserts correct exit codes
- [x] No references to Vyper, vyper-platform, or internal tooling anywhere in examples/

## Implementation Notes

### Decisions
- **Integration tests use `exec.Command`**: Go tests in `tests/integration/` build the `pi` binary once in `TestMain`, then use it to run automations against the example workspaces. Each test asserts exit codes and output content.
- **Simulated Docker**: The docker-project example simulates Docker operations with `echo` commands — no real Docker required. This keeps tests fast and self-contained.
- **Script file example**: `examples/basic/.pi/build/compile.sh` sits next to `compile.yaml` to demonstrate the "scripts beside automations" pattern.

### File structure
```
examples/
  basic/
    pi.yaml                         — project config with shortcuts
    README.md                       — usage guide
    .pi/
      greet.yaml                    — inline bash, argument passing
      deploy.yaml                   — run: step chaining (calls build/compile)
      build/
        compile.yaml                — .sh file step
        compile.sh                  — script file beside automation
  docker-project/
    pi.yaml                         — project config with shortcuts
    README.md                       — usage guide
    .pi/docker/
      up.yaml                       — simulated docker-compose up
      down.yaml                     — simulated docker-compose down
      logs.yaml                     — simulated logs with arg filtering
      build.yaml                    — simulated image build
      build-and-up.yaml             — run: chaining (build → up)
tests/integration/
  examples_test.go                  — 13 integration tests
```

### Test coverage (13 integration tests)
- **basic**: list (headers + names), greet, greet with arg, build/compile, deploy (run: chain), not found error, from subdirectory
- **docker-project**: list (all 5 automations), up, down, logs, logs with arg, build-and-up (run: chain order verified)

## Subtasks
- [x] Create `examples/basic/` workspace
- [x] Create `examples/docker-project/` workspace
- [x] Write README for each workspace
- [x] Write integration test that runs `pi run` against each workspace

## Blocked By
05-pi-run-and-pi-list-commands (done)
