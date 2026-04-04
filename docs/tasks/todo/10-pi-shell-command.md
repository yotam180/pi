# `pi shell` Command

## Type
feature

## Status
todo

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Implement the `pi shell` command that reads shortcuts from `pi.yaml` and writes shell functions into the developer's shell config. Each shortcut function either `cd`s to the repo root and runs `pi run <automation>`, or runs without `cd` when `anywhere: true`. Shortcuts are written to a PI-managed file (`~/.pi/shell/<repo-slug>.sh`) and a single source line is added to `.zshrc`/`.bashrc`.

## Acceptance Criteria
- [ ] `pi shell` reads `pi.yaml → shortcuts` and writes shell functions to `~/.pi/shell/<project-name>.sh`
- [ ] Each function `cd`s to repo root and runs `pi run <automation> "$@"` by default
- [ ] Functions with `anywhere: true` use `pi run --repo <root> <automation> "$@"` without `cd`
- [ ] A `source ~/.pi/shell/*.sh` line is added to `~/.zshrc` (and `~/.bashrc` if it exists), only once
- [ ] `pi shell --uninstall` removes the file for the current repo and cleans the source line if no repos remain
- [ ] `pi shell --list` shows all currently installed shortcuts across all repos
- [ ] Running `pi shell` again overwrites the file (idempotent)
- [ ] `pi run --repo <path>` flag is added so anywhere-shortcuts can specify the repo without `cd`
- [ ] `pi setup` automatically runs `pi shell` as its final step — a full setup leaves the developer with both environment and shortcuts configured
- [ ] `pi setup` skips the `pi shell` step when `CI=true` (or any common CI env var is set) — shell config is irrelevant in CI
- [ ] `pi setup --no-shell` flag skips the shell step explicitly
- [ ] Unit tests for function generation, source line injection, uninstall
- [ ] Integration test: run `pi shell` in example workspace and verify the generated file content

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `internal/shell/` package for shortcut generation
- [ ] Implement shell function generation (default + anywhere variants)
- [ ] Implement source line injection into `.zshrc`/`.bashrc`
- [ ] Wire `pi shell` cobra command
- [ ] Add `--repo` flag to `pi run`
- [ ] Implement `--uninstall` and `--list`
- [ ] Add CI env var detection to `pi setup` (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.)
- [ ] Add `--no-shell` flag to `pi setup`
- [ ] Write tests

## Blocked By
<!-- None — can start as soon as pi run works -->
