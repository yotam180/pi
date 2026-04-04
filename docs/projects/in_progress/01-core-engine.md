# Core Engine

## Status
in_progress

## Priority
high

## Description
Build the foundational `pi` CLI binary. This project establishes the config schema (`pi.yaml` + `.pi/` folder), the automation resolution and execution model, and the `pi run` command with bash step support. Everything else in PI builds on top of this.

## Goals
- A developer can drop `pi.yaml` and a `.pi/` folder into any repo and run automations with `pi run <name>`
- Automations can be defined as single `.yaml` files or as `<name>/automation.yaml` directories (for automations that bundle assets/scripts alongside them)
- Automations can call other local automations via `run:` steps
- Bash steps work (inline and file references)
- The binary is a single Go binary with no runtime dependencies

## Background & Context
See `docs/README.md` for the full product definition.

The `.pi/` folder model is inspired by GitHub Actions — each automation is a self-contained unit that can bundle its own scripts and assets. The key difference is it's local-first and language-agnostic.

## Scope

### In scope
- `pi.yaml` config parsing (shortcuts + setup sections)
- `.pi/` folder scanning and automation loading
- Automation resolution: local `.pi/<name>.yaml` and `.pi/<name>/automation.yaml`
- `pi run <name> [args]` command
- `pi list` command (list all discovered automations with descriptions)
- Step type: `bash` (inline string and `.sh` file path, relative to automation file)
- Step type: `run` (call another local automation by name)
- Error handling: clear messages when automation not found or step fails
- Exit codes propagated correctly
- `examples/` folder with realistic sample workspaces used as integration tests

### Out of scope
- Python / TypeScript step types (Project 2)
- Shell shortcut injection (Project 2)
- Built-in automation library (Project 3)
- Marketplace / remote automations
- `pipe_to` between steps
- Inter-step communication / named outputs

## Success Criteria
- [ ] `pi run docker/up` works in the `examples/basic-docker/` workspace
- [ ] `pi run build-and-deploy` works when that automation chains `run:` steps
- [ ] `pi list` prints all automations in a workspace with names and descriptions
- [ ] `examples/` contains at least 2 sample workspaces covering different usage patterns
- [ ] Single binary, `go build` with no CGO dependencies
- [ ] `go test ./...` passes

## Notes
- Automation names are their path relative to `.pi/`, without the `.yaml` extension. Example: `.pi/docker/up.yaml` → name `docker/up`.
- When looking up `docker/up`, PI tries: `.pi/docker/up.yaml` first, then `.pi/docker/up/automation.yaml`.
- Arguments passed to `pi run <name> [args]` are available to bash steps as `$@`.
- Keep the config schema strict with good validation errors — the agent and developers will depend on clear feedback.
- The `examples/` folder is the integration test bed. Each workspace should be a realistic scenario, not a hello-world toy.
