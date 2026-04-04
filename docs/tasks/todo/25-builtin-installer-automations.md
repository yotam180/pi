# Built-in Installer Automations

## Type
feature

## Status
todo

## Priority
medium

## Project
03-built-in-library-and-pi-setup

## Description
Implement the core "check-then-install" built-in automations for common developer tools. These are the automations that make `pi setup` useful out of the box — any project can reference them with the `pi:` prefix.

Each installer automation should:
- Check if the tool is already installed (and optionally at the right version)
- Install it if missing, using the most portable method available
- Accept a `version` input where applicable
- Be idempotent — running twice produces no changes on the second run
- Print clear status: `[already installed]` or `[installed]`

### Automations to implement:
1. **`pi:install-homebrew`** — macOS only, installs Homebrew via the official install script
2. **`pi:install-python`** — installs Python at a given version; uses `mise` if available, otherwise attempts `brew install python@X.Y` on macOS
3. **`pi:install-node`** — installs Node.js at a given version; uses `mise` if available, otherwise attempts `brew install node@X` on macOS
4. **`pi:install-uv`** — installs `uv` via `curl -LsSf https://astral.sh/uv/install.sh | sh`
5. **`pi:install-tsx`** — installs `tsx` globally via `npm install -g tsx`

## Acceptance Criteria
- [ ] All 5 automations exist in `builtins/.pi/` and are discoverable with `pi list`
- [ ] Each accepts appropriate inputs (e.g., `version` for python/node)
- [ ] Each is idempotent — skips install when tool is already present at the right version
- [ ] `pi run pi:install-uv` on a machine without `uv` installs it; on a machine with `uv`, prints a skip message
- [ ] macOS-only automations (`pi:install-homebrew`) use `if: os.macos` at the automation level
- [ ] Unit tests verify YAML structure and input specs for each
- [ ] At least one integration test runs `pi:install-tsx` (since npm/tsx are in CI)

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `builtins/.pi/install-homebrew.yaml`
- [ ] Create `builtins/.pi/install-python.yaml`
- [ ] Create `builtins/.pi/install-node.yaml`
- [ ] Create `builtins/.pi/install-uv.yaml`
- [ ] Create `builtins/.pi/install-tsx.yaml`
- [ ] Write unit tests for YAML structure and input validation
- [ ] Write integration test for at least one installer

## Blocked By
23-builtin-automation-infrastructure
