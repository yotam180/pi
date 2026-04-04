# Built-in Dev Tool Automations

## Type
feature

## Status
todo

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
- [ ] `builtins/.pi/cursor/install-extensions/automation.yaml` exists and accepts an `extensions` input (comma-separated list or newline-separated)
- [ ] `builtins/.pi/git/install-hooks.yaml` exists and accepts a `source` input (directory path relative to repo root)
- [ ] Each automation is idempotent
- [ ] Each has a clear `name:` and `description:`
- [ ] Unit tests verify YAML structure

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `builtins/.pi/cursor/install-extensions/automation.yaml`
- [ ] Create `builtins/.pi/git/install-hooks.yaml`
- [ ] Write unit tests

## Blocked By
23-builtin-automation-infrastructure
