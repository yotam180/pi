# Dogfood: PI Setup for the PI Repo Itself

## Type
chore

## Status
todo

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
- [ ] `pi.yaml` exists at the repo root with setup entries using `pi:` built-in automations
- [ ] `.pi/` folder contains repo-specific automations (build, test, etc.)
- [ ] `pi setup` in a fresh clone of the PI repo installs all required dev tools
- [ ] `pi list` shows both local and built-in automations
- [ ] Developer workflow documented in README: clone → `pi setup` → ready to develop

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `pi.yaml` with setup entries
- [ ] Create `.pi/build.yaml` and `.pi/test.yaml` local automations
- [ ] Test `pi setup` end-to-end
- [ ] Update README with `pi setup` instructions

## Blocked By
25-builtin-installer-automations
