# Built-in Installer Automations

## Type
feature

## Status
done

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
- [x] All 5 automations exist in `internal/builtins/embed_pi/` and are discoverable with `pi list`
- [x] Each accepts appropriate inputs (e.g., `version` for python/node)
- [x] Each is idempotent — skips install when tool is already present at the right version
- [x] `pi run pi:install-uv` on a machine without `uv` installs it; on a machine with `uv`, prints a skip message
- [x] macOS-only automations (`pi:install-homebrew`) use `if: os.macos` at the automation level
- [x] Unit tests verify YAML structure and input specs for each
- [x] At least one integration test runs `pi:install-tsx` (since npm/tsx are in CI)

## Implementation Notes

### File locations
All five automations live in `internal/builtins/embed_pi/` (embedded into the binary at build time):
- `install-homebrew.yaml` — macOS only (`if: os.macos`), installs via official Homebrew install script
- `install-python.yaml` — accepts `version` input; tries `mise` first, falls back to `brew install python@X.Y`
- `install-node.yaml` — accepts `version` input (major version); tries `mise` first, falls back to `brew install node@X`
- `install-uv.yaml` — installs via official `astral.sh/uv/install.sh` script
- `install-tsx.yaml` — installs via `npm install -g tsx`

### Design decisions
- All scripts are idempotent: they check `command -v` first and print `[already installed]` with version info when the tool is present
- `install-python` and `install-node` accept a `version` input via the `inputs:` schema, which gets injected as `PI_INPUT_VERSION` env var
- `install-homebrew` uses automation-level `if: os.macos` so it's skipped on non-macOS platforms
- `install-python` checks both `pythonX.Y` and `python3 --version` for version match
- `install-node` extracts major version from `node --version` for comparison against the requested major version

### Test coverage
- **Unit tests** (12 new tests in `internal/builtins/builtins_test.go`):
  - All 5 automations exist with correct names, descriptions, single bash steps
  - All are resolvable via `Find()`
  - All scripts are idempotent (check `command -v`, print `[already installed]`/`[installed]`)
  - `install-homebrew` has `if: os.macos` condition
  - `install-python` and `install-node` accept `version` input (required, with description, uses `PI_INPUT_VERSION`)
  - `install-uv` uses official astral.sh installer
  - `install-tsx` uses `npm install -g tsx`
  - `install-python` and `install-node` try mise then brew
  - `install-homebrew`, `install-uv`, `install-tsx` have no inputs
- **Integration tests** (6 new tests in `tests/integration/examples_test.go`):
  - All 5 installer automations appear in `pi list` with `[built-in]` marker
  - `pi info` shows details for each (name, description substring)
  - `pi info` shows `version` input for `install-python` and `install-node`
  - `pi info` shows `os.macos` condition for `install-homebrew`
  - `pi run pi:install-tsx` runs successfully and outputs `[already installed]` or `[installed]`
  - `pi list` shows `version` input in INPUTS column for `install-python`

## Subtasks
- [x] Create `internal/builtins/embed_pi/install-homebrew.yaml`
- [x] Create `internal/builtins/embed_pi/install-python.yaml`
- [x] Create `internal/builtins/embed_pi/install-node.yaml`
- [x] Create `internal/builtins/embed_pi/install-uv.yaml`
- [x] Create `internal/builtins/embed_pi/install-tsx.yaml`
- [x] Write unit tests for YAML structure and input validation (12 tests)
- [x] Write integration tests (6 tests including pi:install-tsx execution)

## Blocked By
23-builtin-automation-infrastructure
