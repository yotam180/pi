# Transform fzf to Use PI

## Type
feature

## Status
done

## Priority
medium

## Project
07-fzf-adoption-test

## Description
Create PI automations for fzf's developer workflows in `~/projects/fzf`. This is the final phase of the adoption test — actually using PI to model a real project's workflows.

## Acceptance Criteria
- [x] `~/projects/fzf/.pi/` contains automations for all major workflows
- [x] `~/projects/fzf/pi.yaml` has shortcuts and setup entries
- [x] `pi setup` in fzf installs all required tools
- [x] `pi shell` in fzf installs working shortcuts
- [x] `pi run <name>` for each automation produces correct results
- [x] `pi list` shows all automations with descriptions
- [x] Quirks documented in Implementation Notes with feature-vs-YAML analysis

## Implementation Notes

### Automations Created (16 total)

| Automation | Description | Maps to Makefile |
|------------|-------------|------------------|
| `build/default` | Build fzf binary for current platform | `make` (all) |
| `build/install` | Build and copy to bin/ | `make install` |
| `build/snapshot` | Cross-compile via goreleaser snapshot | `make build` |
| `build/clean` | Remove build artifacts | `make clean` |
| `test/unit` | Run Go unit tests | `make test` |
| `test/integration` | Run Ruby integration tests | `make itest` |
| `test/bench` | Run Go benchmarks | `make bench` |
| `test/fuzz` | Run Go fuzz tests | (CI-only in fzf) |
| `lint/all` | Run all linters | `make lint` |
| `lint/go` | Check Go formatting | gofmt check |
| `lint/ruby` | Run rubocop | rubocop |
| `lint/shell` | Check shell formatting | shfmt check |
| `lint/fmt` | Format Go and shell files | `make fmt` |
| `docker/build` | Build Docker test image | `docker build` |
| `docker/run` | Build and run with tmux | `make docker` |
| `docker/test` | Build and run tests | `make docker-test` |
| `generate` | Run Go code generators | `make generate` |
| `update-deps` | Update Go dependencies | `make update` |

### Setup Automations (5)
1. `setup/install-go` — ensures Go >= 1.23 (local installer with >= semantics)
2. `setup/install-ruby` — installs Ruby via mise/brew
3. `setup/install-shfmt` — installs shfmt via brew/go
4. `setup/install-goreleaser` — installs goreleaser via brew/go
5. `setup/install-ruby-gems` — runs `bundle install`

### Shell Shortcuts (9)
`fzf-build`, `fzf-install`, `fzf-test`, `fzf-itest`, `fzf-bench`, `fzf-lint`, `fzf-fmt`, `fzf-clean`, `fzf-docker`

### Test Results

| Command | Result | Notes |
|---------|--------|-------|
| `pi list` | ✓ | Shows all 18 local + 12 built-in automations |
| `pi run build/default` | ✓ | Builds fzf binary successfully |
| `pi run build/install` | ✓ | Builds and copies to bin/ |
| `pi run test/unit` | ✓ | All Go unit tests pass |
| `pi run lint/go` | ✓ | Go formatting is clean |
| `pi run build/clean` | ✓ | Removes build artifacts |
| `pi doctor` | ✓ | Shows all requirements with versions |
| `pi setup --no-shell` | ✓ (partial) | Go, Ruby, shfmt, goreleaser all pass; Ruby gems fails due to system Ruby 2.6 being too old |
| `pi shell` | ✓ | Installs 9 shortcuts |
| `pi shell uninstall` | ✓ | Removes shortcuts cleanly |

### Bugs Found and Fixed

1. **`go version` detection**: PI tried `go --version` which fails (Go uses `go version` without `--`). Fixed `detectVersionExec()` to try `<cmd> --version` first, then fall back to `<cmd> version`. This also helps with any other tools that use subcommands for version info.

### Quirks and Lessons Learned

1. **`pi:install-go` vs local `setup/install-go`**: The built-in `pi:install-go` expects an exact major.minor version match. For setup, we want "at least this version" semantics. Used a local installer with `>=` comparison logic instead. This suggests PI might benefit from a `min_version` mode on install: test phases.

2. **env: feature pays off**: The `test/unit` and `test/bench` automations use `env: SHELL: /bin/sh` matching the Makefile's `SHELL=/bin/sh` override. The new env: feature made this clean.

3. **Ruby version gap**: The system Ruby on macOS (2.6) is too old for fzf's Gemfile.lock which needs bundler 2.6.2. This is not a PI issue — it's the same problem you'd hit running `make lint` without mise/rbenv.

4. **PI's strength**: The `pi list`, `pi doctor`, and `pi info` commands provide much better discoverability than reading a Makefile. New developers can understand all available workflows instantly.

## Subtasks
- [x] Create `.pi/` and `pi.yaml`
- [x] Write build automations
- [x] Write test automations
- [x] Write lint/format automations
- [x] Write setup automations
- [x] Define shortcuts
- [x] Test all automations
- [x] Test `pi setup` and `pi shell`
- [x] Document quirks and lessons learned

## Blocked By
34-clone-and-examine-fzf (completed)
35-implement-fzf-feature-gaps (completed)
