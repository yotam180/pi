# Clone and Examine fzf Workflows

## Type
research

## Status
in_progress

## Priority
high

## Project
07-fzf-adoption-test

## Description
Clone junegunn/fzf into `~/projects/fzf` and examine all developer workflows. Document every build command, test command, lint/format command, Docker operation, setup step, and release command. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

## Acceptance Criteria
- [x] fzf cloned to `~/projects/fzf`
- [x] Every Makefile target documented
- [x] Every script documented
- [x] Every CI workflow documented
- [x] Required tools/runtimes listed
- [x] PI feature gap analysis completed
- [x] Findings recorded in Implementation Notes

## Implementation Notes

### Repository Overview

fzf (v0.71.0) is a Go CLI tool (~15k LOC) with:
- 10 Makefile targets
- 2 install/uninstall scripts (bash)
- 8 shell integration files (bash/zsh/fish completion + key bindings)
- 8 CI workflows
- GoReleaser config with macOS notarization
- Ruby-based integration tests (via tmux)
- `.tool-versions` specifying: golang 1.23, ruby 3.4, shfmt 3.12

---

### Makefile Targets

| Target | Command | Description |
|--------|---------|-------------|
| `all` (default) | `go build` with ldflags | Build fzf binary for current platform into `target/` |
| `test` | `go test -v` on 4 packages | Run Go unit tests (algo, src, tui, util) |
| `itest` | `ruby test/runner.rb` | Run Ruby integration tests (requires tmux) |
| `bench` | `go test -bench=.` | Run benchmarks in `src/` |
| `lint` | gofmt + rubocop + shfmt | Lint Go, Ruby, and bash code |
| `fmt` | gofmt + shfmt | Format Go and bash code |
| `install` | build + copy to `bin/fzf` | Build and install locally |
| `build` | `goreleaser build --snapshot` | Cross-compile for all platforms via goreleaser |
| `release` | tests + goreleaser release | Full release flow (tests, version check, GH release) |
| `clean` | `rm -r dist target` | Clean build artifacts |
| `generate` | `go generate ./...` | Run code generators |
| `update` | `go get -u && go mod tidy` | Update Go dependencies |
| `docker` | `docker build && docker run` | Build Docker image and run with tmux |
| `docker-test` | `docker build && docker run` | Build Docker image and run tests |

**Build flags**: `-a -ldflags "-s -w -X main.version=$(VERSION) -X main.revision=$(REVISION)" -tags "$(TAGS)" -trimpath`

Version/revision are derived from git tags and commit hashes. Can be overridden via `FZF_VERSION` and `FZF_REVISION` env vars.

Multi-arch targets: 386, amd64, arm5/6/7/8, ppc64le, riscv64, loong64, s390x.

---

### Scripts

#### `install` (bash, 453 lines)
- Downloads pre-built fzf binary from GitHub Releases
- Falls back to `go install` if no prebuilt binary available
- Optionally configures shell integration (completion + key bindings) for bash, zsh, fish
- Interactive prompts: auto-completion, key bindings, shell config update
- Flags: `--all`, `--bin`, `--xdg`, `--[no-]key-bindings`, `--[no-]completion`, `--[no-]update-rc`, `--no-bash`, `--no-zsh`, `--no-fish`
- Generates `~/.fzf.bash` and `~/.fzf.zsh` source files
- Appends source lines to `.bashrc`/`.zshrc`
- Handles fish separately via `fish_user_paths` and `fish_user_key_bindings`

#### `uninstall` (bash, 121 lines)
- Removes generated config files (`~/.fzf.bash`, `~/.fzf.zsh`)
- Removes source lines from `.bashrc`/`.zshrc`
- Removes fish integration files
- Flag: `--xdg`

#### `shell/update.sh` (bash, 68 lines)
- Applies `common.sh` template into completion and key-binding scripts
- Runs `shfmt` to format bash scripts
- `--check` mode: diff-only, exits non-zero if formatting needed

#### `bin/fzf-preview.sh` (bash)
- Preview helper script used by fzf

#### `shell/common.sh` (bash)
- Shared shell functions included in completion and key-binding scripts

---

### Shell Integration Files

| File | Purpose |
|------|---------|
| `shell/completion.bash` | Bash fuzzy completion |
| `shell/completion.zsh` | Zsh fuzzy completion |
| `shell/completion.fish` | Fish fuzzy completion |
| `shell/key-bindings.bash` | Bash key bindings (CTRL-T, CTRL-R, ALT-C) |
| `shell/key-bindings.zsh` | Zsh key bindings |
| `shell/key-bindings.fish` | Fish key bindings |

---

### CI Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `linux.yml` | push/PR to master/devel | Full test suite: lint (rubocop + gofmt + shfmt), unit test, fuzz test, integration test (tmux) |
| `macos.yml` | push/PR to master | Unit tests + integration tests on macOS |
| `typos.yml` | PR | Spell check via typos |
| `codeql-analysis.yml` | push/PR to master/devel | CodeQL security analysis for Go |
| `labeler.yml` | PR | Auto-label PRs |
| `sponsors.yml` | schedule/manual | Generate sponsors section in README |
| `depsreview.yaml` | PR | Dependency review |
| `winget.yml` | release | Publish to Windows Package Manager |

**Key CI details from linux.yml**:
- Go 1.23 via `actions/setup-go@v6`
- Ruby 3.4.6 via `ruby/setup-ruby@v1`
- Packages: zsh, fish, tmux, shfmt (apt)
- Ruby gems via `bundle install`
- Steps: rubocop → unit test → fuzz test → integration test
- Fuzz test: `go test ./src/algo/ -fuzz=FuzzIndexByteTwo -fuzztime=5s` (x2)
- Integration test: `make install && ./install --all && tmux new-session -d && ruby test/runner.rb --verbose`

---

### GoReleaser Config

- Version 2
- Builds for: darwin, linux, windows, freebsd, openbsd, android
- Architectures: amd64, arm, arm64, loong64, ppc64le, s390x, riscv64
- ARM variants: v5, v6, v7
- macOS binary signing + notarization via Apple certificates
- Archives: tar.gz (unix), zip (windows)
- Snapshot versioning: `{{ .Version }}-devel`

---

### Dockerfile

```dockerfile
FROM rubylang/ruby:3.4.1-noble
RUN apt-get update -y && apt install -y git make golang zsh fish tmux
RUN gem install --no-document -v 5.22.3 minitest
COPY . /fzf
RUN cd /fzf && make install && ./install --all
CMD ["bash", "-ic", "tmux new 'set -o pipefail; ruby /fzf/test/runner.rb | tee out && touch ok' && cat out && [ -e ok ]"]
```

Used for running integration tests in an isolated environment with all required shells.

---

### Required Tools/Runtimes

| Tool | Version | Purpose | Available on dev machines |
|------|---------|---------|--------------------------|
| Go | >= 1.23 | Build, test | Yes |
| Ruby | >= 3.4 | Integration tests, rubocop | Usually no |
| shfmt | >= 3.12 | Shell script formatting | Usually no |
| goreleaser | latest | Cross-compilation, releases | Optional |
| tmux | latest | Integration test runner | Usually yes |
| zsh | latest | Shell integration tests | Usually yes (macOS) |
| fish | latest | Shell integration tests | Usually no |
| bash | latest | Build scripts, shell integration | Always |
| Docker | latest | Integration test environment | Optional |
| bundle (Bundler) | latest | Ruby dependency management | Comes with Ruby |

**From `.tool-versions`**: golang 1.23, ruby 3.4, shfmt 3.12

---

### PI Feature Coverage Gap Analysis

#### Can PI model today ✓

| Workflow | PI automation | Notes |
|----------|---------------|-------|
| `make` (default build) | `bash: go build ...` step | Straightforward single bash step |
| `make test` | `bash: go test ...` step | Single bash command |
| `make itest` | `bash: ruby test/runner.rb` step | Single bash command (requires ruby) |
| `make bench` | `bash: go test -bench=...` step | Single bash command |
| `make clean` | `bash: rm -rf dist target` | Single bash command |
| `make install` | Multi-step: build then copy | Two bash steps |
| `make generate` | `bash: go generate ./...` | Single bash command |
| `make update` | `bash: go get -u && go mod tidy` | Single bash command |
| `make docker` | `bash: docker build && docker run` | Two bash steps or single inline |
| `make docker-test` | `bash: docker build && docker run` | Similar to above |
| `make fmt` | Multi-step: gofmt + shfmt | Two bash steps |
| `make lint` | Multi-step: gofmt + rubocop + shfmt | Three bash steps |
| `make build` (goreleaser) | `bash: goreleaser build --snapshot` | Single bash command |
| `make release` | Multi-step bash | Complex but all bash |
| Setup: install Go | `pi:install-python` pattern | PI already has install-node/python; Go would need a new built-in |
| Setup: install Ruby | `install:` automation | Not built-in, but doable as local `.pi/` automation |
| Setup: install shfmt | `install:` automation | Not built-in, doable as local automation |
| Setup: install goreleaser | `install:` automation | Not built-in, doable as local automation |
| Setup: install tmux | `install:` automation | Not built-in, doable as local automation |
| Shell shortcuts | `pi.yaml → shortcuts:` | Works perfectly for build/test/lint/etc |

#### PI feature gaps identified ✗

| Gap | Severity | Description |
|-----|----------|-------------|
| **No `pi:install-go` built-in** | Low | PI has `install-python` and `install-node` but not Go. fzf needs Go >= 1.23. Can be modeled as a local `.pi/` install automation, but Go is common enough to warrant a built-in. |
| **No `pi:install-ruby` built-in** | Low | Ruby is needed for fzf integration tests. Same as Go — can be local but might be useful built-in. |
| **No `pi:install-shfmt` built-in** | Very Low | shfmt is niche. Better as a local automation. |
| **No env var injection for steps** | Medium | The Makefile uses `TAGS=pprof` to control build tags, `SHELL=/bin/sh GOOS=` to override env before `go test`. PI steps can't declare environment variables. Workaround: inline them in the bash command (`TAGS=pprof go build ...`). This works but is less clean. An `env:` field on steps would be better. |
| **No `fuzz` test type** | Very Low | `go test -fuzz=...` is just a bash command; no special support needed. |
| **No cross-compile / matrix support** | Low | The Makefile has targets for each arch. PI has no concept of matrix builds or parameterized parallel execution. Not blocking — goreleaser handles this for fzf. |
| **No `make` step type** | Very Low | PI doesn't have a `make` step type. Not needed — bash works fine. |

#### Summary

**PI can model 100% of fzf's developer workflows today** using bash steps. The gaps are all quality-of-life improvements, not blockers:

1. **Most impactful gap**: No `env:` field on steps for cleanly setting environment variables. Workaround exists (inline in bash) but is inelegant.
2. **Nice-to-have**: `pi:install-go` built-in would benefit many Go projects since PI itself is a Go tool.
3. **Not blocking**: All other gaps are low priority.

The main conclusion is that PI's current feature set is sufficient to model a real-world Go project's developer workflows. The automation model (YAML with bash/python/typescript steps) handles the variety of tools well.

## Subtasks
- [x] Clone repo
- [x] Document Makefile targets
- [x] Document scripts
- [x] Document CI workflows
- [x] List required tools
- [x] Assess PI feature coverage
- [x] Document gaps

## Blocked By
