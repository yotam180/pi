# Built-in Dev Tool Automations

## Type
feature

## Status
done

## Priority
low

## Project
03-built-in-library-and-pi-setup

## Description
Implement built-in automations for developer tooling setup: Cursor extension installation and git hook installation. These are common setup tasks that teams currently handle with custom scripts.

### Automations to implement:
1. **`pi:cursor/install-extensions`** — reads a list of extension IDs (from an `extensions` input or a file) and installs any that are missing via `cursor --install-extension`
2. **`pi:git/install-hooks`** — copies hook scripts from a source directory to `.git/hooks/`, making them executable

## Acceptance Criteria
- [x] `builtins/.pi/cursor/install-extensions/automation.yaml` exists and accepts an `extensions` input (comma-separated list or newline-separated)
- [x] `builtins/.pi/git/install-hooks.yaml` exists and accepts a `source` input (directory path relative to repo root)
- [x] Each automation is idempotent
- [x] Each has a clear `name:` and `description:`
- [x] Unit tests verify YAML structure

## Implementation Notes

### cursor/install-extensions
- Located at `internal/builtins/embed_pi/cursor/install-extensions/automation.yaml`
- Accepts `extensions` input (required, string): comma or newline-separated extension IDs
- Normalizes input: replaces commas with newlines, trims whitespace, removes empty lines
- Checks existing extensions via `cursor --list-extensions`, installs missing ones via `cursor --install-extension --force`
- Prints `[already installed]` when all present, `[installed]` with counts otherwise

### git/install-hooks
- Located at `internal/builtins/embed_pi/git/install-hooks.yaml`
- Accepts `source` input (required, string): directory path relative to repo root
- Validates source directory exists and `.git/hooks` is present
- Uses `cmp -s` for idempotent file comparison (skips unchanged hooks)
- Copies files, `chmod +x` on each hook
- Handles edge cases: no hook files found, all up to date, partial updates

### Tests
- 16 new unit tests in `builtins_test.go` covering: existence, resolvability, bash step type, input specs, env var usage, idempotency patterns, cursor CLI usage, chmod, .git/hooks reference
- 6 new integration tests in `examples_test.go` covering: list presence, [built-in] marker, info details, info inputs, list INPUTS column

## Subtasks
- [x] Create `builtins/.pi/cursor/install-extensions/automation.yaml`
- [x] Create `builtins/.pi/git/install-hooks.yaml`
- [x] Write unit tests
- [x] Write integration tests

## Blocked By
23-builtin-automation-infrastructure
