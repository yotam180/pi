# Built-in Docker Automations

## Type
feature

## Status
todo

## Priority
medium

## Project
03-built-in-library-and-pi-setup

## Description
Implement the standard Docker Compose built-in automations: `pi:docker/up`, `pi:docker/down`, and `pi:docker/logs`. These are thin wrappers around `docker-compose` commands that any project can reference without writing their own automation YAML.

Each automation should:
- Check that `docker` and `docker-compose` (or `docker compose`) are available
- Use `docker compose` (v2 plugin) as the default, falling back to `docker-compose` (standalone)
- Forward any extra CLI args to the underlying command
- Work from the repo root (where `docker-compose.yaml` is expected to be)

## Acceptance Criteria
- [ ] `builtins/.pi/docker/up.yaml` runs `docker compose up -d` (or `docker-compose up -d`)
- [ ] `builtins/.pi/docker/down.yaml` runs `docker compose down`
- [ ] `builtins/.pi/docker/logs.yaml` runs `docker compose logs -f --tail 200`
- [ ] Extra args are forwarded: `pi run pi:docker/up -- --build` passes `--build` to docker compose
- [ ] Each automation has a clear `name:` and `description:`
- [ ] Unit/integration tests verify the automations are discoverable and have correct structure

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `builtins/.pi/docker/up.yaml`
- [ ] Create `builtins/.pi/docker/down.yaml`
- [ ] Create `builtins/.pi/docker/logs.yaml`
- [ ] Write tests verifying discoverability and YAML validity

## Blocked By
23-builtin-automation-infrastructure
