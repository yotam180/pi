# `pi shell` Command

## Type
feature

## Status
done

## Priority
high

## Project
02-polyglot-runner-and-shell-integration

## Description
Implement the `pi shell` command that reads shortcuts from `pi.yaml` and writes shell functions into the developer's shell config. Each shortcut function either `cd`s to the repo root and runs `pi run <automation>`, or runs without `cd` when `anywhere: true`. Shortcuts are written to a PI-managed file (`~/.pi/shell/<repo-slug>.sh`) and a single source line is added to `.zshrc`/`.bashrc`.

## Acceptance Criteria
- [x] `pi shell` reads `pi.yaml → shortcuts` and writes shell functions to `~/.pi/shell/<project-name>.sh`
- [x] Each function `cd`s to repo root and runs `pi run <automation> "$@"` by default
- [x] Functions with `anywhere: true` use `pi run --repo <root> <automation> "$@"` without `cd`
- [x] A `source ~/.pi/shell/*.sh` line is added to `~/.zshrc` (and `~/.bashrc` if it exists), only once
- [x] `pi shell uninstall` removes the file for the current repo and cleans the source line if no repos remain
- [x] `pi shell list` shows all currently installed shortcuts across all repos
- [x] Running `pi shell` again overwrites the file (idempotent)
- [x] `pi run --repo <path>` flag is added so anywhere-shortcuts can specify the repo without `cd`
- [x] `pi setup` automatically runs `pi shell` as its final step — a full setup leaves the developer with both environment and shortcuts configured
- [x] `pi setup` skips the `pi shell` step when `CI=true` (or any common CI env var is set) — shell config is irrelevant in CI
- [x] `pi setup --no-shell` flag skips the shell step explicitly
- [x] Unit tests for function generation, source line injection, uninstall
- [x] Integration test: run `pi shell` in example workspace and verify the generated file content

## Implementation Notes

### Architecture
- Created `internal/shell/` package for all shell shortcut logic (generation, install, uninstall, list)
- Shell functions are written to `~/.pi/shell/<project>.sh`
- Source line uses a glob pattern (`for f in ~/.pi/shell/*.sh; do source "$f"; done`) so new projects auto-activate
- Source line injection is idempotent — it checks for existing `# Added by PI` marker before appending
- `pi shell` command has `uninstall` and `list` subcommands (not flags) for clean UX

### `--repo` flag
- Added `--repo <path>` flag to `pi run` so "anywhere" shortcuts can specify the project root without cd
- The `--repo` value is used as the starting directory for `FindRoot()`, so it works even if pointing to a subdirectory

### `pi setup` redesign
- Replaced the stub with a real implementation that runs setup entries sequentially
- After setup entries complete, runs `pi shell` automatically unless `--no-shell` or CI is detected
- CI detection checks: `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, `CIRCLECI`, `JENKINS_URL`, `BUILDKITE`, `TRAVIS`, `CODEBUILD_BUILD_ID`, `TF_BUILD`

### Testing
- 11 unit tests in `internal/shell/` covering generation, install, uninstall, list, idempotency, multi-project scenarios
- 7 CLI tests in `internal/cli/` for shell and setup commands
- 7 integration tests using the built binary: install, idempotent, uninstall, list, --repo, setup integration, --no-shell
- Total test count: 152 (up from 127)

## Subtasks
- [x] Create `internal/shell/` package for shortcut generation
- [x] Implement shell function generation (default + anywhere variants)
- [x] Implement source line injection into `.zshrc`/`.bashrc`
- [x] Wire `pi shell` cobra command
- [x] Add `--repo` flag to `pi run`
- [x] Implement `uninstall` and `list` subcommands
- [x] Add CI env var detection to `pi setup` (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.)
- [x] Add `--no-shell` flag to `pi setup`
- [x] Write unit tests (11 shell + 7 CLI)
- [x] Write integration tests (7 tests)

## Blocked By
<!-- None — can start as soon as pi run works -->
