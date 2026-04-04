# Built-in Docker Automations

## Type
feature

## Status
done

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
- [x] `builtins/.pi/docker/up.yaml` runs `docker compose up -d` (or `docker-compose up -d`)
- [x] `builtins/.pi/docker/down.yaml` runs `docker compose down`
- [x] `builtins/.pi/docker/logs.yaml` runs `docker compose logs -f --tail 200`
- [x] Extra args are forwarded: `pi run pi:docker/up -- --build` passes `--build` to docker compose
- [x] Each automation has a clear `name:` and `description:`
- [x] Unit/integration tests verify the automations are discoverable and have correct structure

## Implementation Notes

### File locations
All three automations live in `internal/builtins/embed_pi/docker/` (embedded into the binary at build time):
- `up.yaml` — `docker compose up -d "$@"` with v1 fallback
- `down.yaml` — `docker compose down "$@"` with v1 fallback
- `logs.yaml` — `docker compose logs -f --tail 200 "$@"` with v1 fallback

### Docker Compose detection strategy
Each script checks `docker compose version` (v2 plugin) first, then falls back to `docker-compose` (standalone v1). If neither is found, it prints an error to stderr and exits 1.

### Arg forwarding
All scripts use `"$@"` to forward extra CLI args. The executor's bash runner already passes args via `bash -c "<script>" -- <args>`, so `$@` correctly receives them.

### Test coverage
- **Unit tests** (7 new tests in `internal/builtins/builtins_test.go`):
  - All three automations are discoverable and have correct names/descriptions
  - All are resolvable via `Find()`
  - All use single bash steps
  - Scripts reference both `docker compose` (v2) and `docker-compose` (v1)
  - Each script forwards args via `"$@"` and uses the correct subcommand
- **Integration tests** (5 new tests in `tests/integration/examples_test.go`):
  - `pi list` shows all three docker automations
  - All have `[built-in]` marker in list output
  - `pi info` shows details for each
  - Docker-builtins example workspace resolves `pi:docker/up` in list and info

### Example workspace
Created `examples/docker-builtins/` with a `call-docker-up` automation that references `pi:docker/up` via a `run:` step, used by integration tests.

## Subtasks
- [x] Create `builtins/.pi/docker/up.yaml`
- [x] Create `builtins/.pi/docker/down.yaml`
- [x] Create `builtins/.pi/docker/logs.yaml`
- [x] Write tests verifying discoverability and YAML validity

## Blocked By
23-builtin-automation-infrastructure
