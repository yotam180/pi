# Built-in Automation Library & `pi setup`

## Status
todo

## Priority
medium

## Description
Ship PI with a standard collection of automations for common developer environment tasks (install Python, uv, Node, Homebrew, Cursor extensions, git hooks, etc.) and implement the `pi setup` command that runs idempotent setup steps. Built-in automations are defined in the PI repo's own `.pi/` folder and available to every project with the `pi:` prefix — no marketplace, no download, no config needed.

## Goals
- Any project can do `run: pi:install-python` without hosting their own automation
- `pi setup` runs all automations listed under `setup:` in `pi.yaml` sequentially and idempotently
- Built-in automations cover the most common "check if installed, install if not" patterns for the tools PI's target users (engineering teams) actually use
- The built-in library itself is a showcase of how to write good PI automations — good documentation, clean steps

## Background & Context
Today, `setup_environment.sh` in Vyper Platform contains functions like `ensure_uv_is_installed`, `ensure_required_cursor_packages`, `install_precommit_hook`, and `install_dependencies`. Each is a check-then-act pattern written in bash. The built-in library lets any project replace these with declarative references.

Bundling automations with the binary avoids the chicken-and-egg problem of needing a network connection or a package manager to set up a machine.

## Scope

### In scope
- `pi setup` command — runs `setup:` list from `pi.yaml` in order, stops on first failure
- Built-in automations (prefixed `pi:`, defined in PI's own `.pi/` folder):
  - `pi:install-homebrew` — check/install Homebrew (macOS)
  - `pi:install-python` — check/install Python at a given `version` (via `mise` or `pyenv`)
  - `pi:install-node` — check/install Node.js at a given `version`
  - `pi:install-uv` — check/install `uv`
  - `pi:install-tsx` — check/install `tsx` globally via npm
  - `pi:cursor/install-extensions` — install Cursor extensions from a list file or inline list
  - `pi:git/install-hooks` — copy hooks from a source directory to `.git/hooks/`
  - `pi:docker/up`, `pi:docker/down`, `pi:docker/logs` — basic Docker Compose ops
- `with:` parameter passing to built-in automations (design and schema in task `07-automation-inputs-schema`)
- Setup output: clear per-step status (`[+] installed`, `[*] already installed`, `[-] failed`)

### Out of scope
- Marketplace / remote automations (separate project)
- Windows support
- Rollback / uninstall for setup steps
- Parallel setup steps

## Success Criteria
- [ ] `pi setup` runs all setup steps and reports status for each
- [ ] `pi setup` is idempotent — running it twice on a configured machine produces no changes and no errors
- [ ] All built-in automations listed above are implemented and work on macOS
- [ ] Replace `setup_environment.sh` in the PI repo itself with `pi.yaml` + `pi setup` as the dogfood integration test
- [ ] `with:` parameters work: `run: pi:install-python` with `version: "3.13"` installs Python 3.13
- [ ] `go test ./...` passes; integration test runs `pi setup` in a clean temp environment

## Dependencies
- Task `07-automation-inputs-schema` must be completed before any built-in automation that uses `with:` is implemented. The `inputs:` schema is intentionally uniform across local, built-in, and marketplace automations — do not implement `with:` as a built-in-only shortcut.

## Notes
- Built-in automations are embedded into the binary at build time (Go `//go:embed`). They should be readable in source at `.pi/` in the PI repo itself — this is the dogfood and the documentation at the same time.
- Idempotency is the responsibility of each automation's check logic, not the framework. Document this clearly in the built-in examples.
- For `pi:install-python`, prefer `mise` as the version manager if available; fall back to a direct download. Do not assume `pyenv`.
- Setup steps should not require sudo unless absolutely unavoidable; document any that do.
