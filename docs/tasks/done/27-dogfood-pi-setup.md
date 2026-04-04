# Dogfood: PI Setup for the PI Repo Itself

## Type
chore

## Status
done

## Priority
low

## Project
03-built-in-library-and-pi-setup

## Description
Replace any manual setup instructions in the PI repo with a working `pi.yaml` that uses `pi setup` with built-in automations. This is the ultimate integration test — PI sets up its own development environment.

The PI repo's `pi.yaml` should define setup entries for:
- Installing Go (if applicable — or just documenting it as a prereq)
- Installing tsx (via `pi:install-tsx`)
- Any other dev dependencies

And shortcuts for common development tasks like:
- `pibuild` → `pi run build`
- `pitest` → `pi run test`

## Acceptance Criteria
- [x] `pi.yaml` exists at the repo root with setup entries using `pi:` built-in automations
- [x] `.pi/` folder contains repo-specific automations (build, test, etc.)
- [x] `pi setup` in a fresh clone of the PI repo installs all required dev tools
- [x] `pi list` shows both local and built-in automations
- [x] Developer workflow documented in README: clone → `pi setup` → ready to develop

## Implementation Notes

### pi.yaml
- Setup: `pi:install-tsx` (only external dev dependency besides Go itself)
- Go is treated as a prerequisite — documented in README rather than auto-installed
- Shortcuts: `pibuild`, `pitest`, `picheck`, `pilint`, `piinstall`, `piclean`

### .pi/ automations
Created 7 local automations:
- `build` — builds bin/pi with version ldflags from `git describe`
- `test` — `go test ./... -race -count=1`
- `lint` — `go vet ./...`
- `check` — lint + test (full CI check, uses `run:` steps)
- `install` — builds then copies to /usr/local/bin (uses `run:` to call build first)
- `clean` — removes bin/ and dist/
- `snapshot` — goreleaser snapshot build

### Makefile removal
- Deleted Makefile entirely — all targets replaced by `.pi/` automations
- Updated README to show `pi run` commands instead of `make`
- Updated architecture.md references from Makefile to PI automations

### Verified
- `pi list` shows all 7 local + all built-in automations
- `pi run build` compiles with correct version tag
- `pi run lint` runs go vet
- `pi run clean` removes artifacts
- `pi setup --no-shell` installs tsx idempotently
- `go test ./...` passes (390 tests)

## Subtasks
- [x] Create `pi.yaml` with setup entries
- [x] Create `.pi/build.yaml` and `.pi/test.yaml` local automations
- [x] Create additional automations: lint, check, install, clean, snapshot
- [x] Test `pi setup` end-to-end
- [x] Update README with `pi setup` instructions
- [x] Get rid of makefile

## Blocked By
25-builtin-installer-automations
