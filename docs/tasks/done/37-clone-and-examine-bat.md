# Clone and Examine bat Workflows

## Type
research

## Status
done

## Priority
high

## Project
08-bat-adoption-test

## Description
Clone sharkdp/bat into `~/projects/bat` and examine all developer workflows. Document every build command, test command, lint/format command, CI workflow, and release process. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/sharkdp/bat.git ~/projects/bat`
2. Read `Cargo.toml` thoroughly — document build config and dependencies
3. Read any Makefile or build scripts
4. Read CI workflows (`.github/workflows/`)
5. Read any release configuration (goreleaser, cargo-dist, etc.)
6. List all tools/runtimes required (Rust, cargo, clippy, rustfmt, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [x] bat cloned to `~/projects/bat`
- [x] Every build/test/lint command documented
- [x] Every CI workflow documented
- [x] Required tools/runtimes listed
- [x] PI feature gap analysis completed
- [x] Findings recorded in Implementation Notes

## Implementation Notes

### Repository Overview

bat (v0.26.1) is a Rust CLI tool (~16.7k LOC) providing a `cat` replacement with syntax highlighting and Git integration. Key characteristics:

- **Build system**: Cargo (no Makefile at all — pure Cargo-based project)
- **Build script**: `build/main.rs` — generates man pages, shell completions, and static syntax mappings at compile time
- **MSRV**: Rust 1.88 (declared in `Cargo.toml` via `rust-version` field)
- **Features**: Multiple Cargo feature flags (`application`, `git`, `paging`, `lessopen`, `build-assets`, `regex-onig`, `regex-fancy`, `minimal-application`)
- **Assets**: Syntax definitions and themes stored as git submodules, compiled into binary assets via `assets/create.sh`
- **Nix**: Has `flake.nix` + `.envrc` for direnv/nix-based dev shells
- **No Makefile**: Unlike fzf, bat has zero Makefiles — all workflows are Cargo commands or shell scripts

---

### Build Commands

| Command | Description |
|---------|-------------|
| `cargo build --bins` | Build debug binary |
| `cargo build --locked --release` | Build release binary (with LTO, stripped, codegen-units=1) |
| `cargo build --locked --release --target=<triple>` | Cross-compile release build for specific target |
| `cargo install --path . --locked` | Install bat binary to `~/.cargo/bin/` |
| `cargo install --path . --locked --force` | Force reinstall (e.g. after asset rebuild) |

**Release profile** (from Cargo.toml):
- LTO enabled
- Stripped binaries
- `codegen-units = 1` (maximum optimization)

**Cross-compilation**: CI uses `cross` (a Cargo wrapper) for non-native targets. 13 target triples are built in CI across Linux (glibc + musl), macOS (x86_64 + aarch64), and Windows.

---

### Test Commands

| Command | Description |
|---------|-------------|
| `cargo test` | Run all unit and integration tests |
| `cargo test --locked` | Run tests with locked dependencies |
| `cargo test --locked --release` | Run tests in release mode (used with updated assets) |
| `cargo test --locked --release --test assets -- --ignored` | Run ignored asset tests |
| `cargo test --locked --test system_wide_config -- --ignored` | Run system-wide config tests (requires `BAT_SYSTEM_CONFIG_PREFIX` env var) |
| `cargo test --locked --no-default-features --features minimal-application,bugreport,build-assets` | MSRV test with reduced feature set |
| `cargo check --locked --verbose --lib --no-default-features --features <combo>` | Feature combination checks (5 different combos in CI) |

**Feature check combinations tested in CI**:
1. `regex-onig`
2. `regex-onig,git`
3. `regex-onig,paging`
4. `regex-onig,git,paging`
5. `minimal-application`

**Test files** (Rust):
- `tests/integration_tests.rs` — 4134 lines of integration tests
- `tests/assets.rs` — asset validation
- `tests/no_duplicate_extensions.rs` — extension uniqueness
- `tests/snapshot_tests.rs` — snapshot testing
- `tests/system_wide_config.rs` — system config tests
- `tests/test_pretty_printer.rs` — library API tests
- `tests/github-actions.rs` — CI-specific tests

---

### Scripts

#### `assets/create.sh` (bash)
- Rebuilds binary assets (syntaxes + themes) from git submodules
- Initializes submodules interactively if needed
- Applies and reverses patches from `assets/patches/`
- Runs `bat cache --build --blank --acknowledgements`
- Must be run before `cargo install` when modifying syntax/theme assets

#### `tests/syntax-tests/regression_test.sh` (bash)
- Runs syntax highlighting regression tests
- Calls Python scripts: `create_highlighted_versions.py` and `compare_highlighted_versions.py`
- Requires `bat` in PATH

#### `tests/syntax-tests/update.sh` (bash)
- Updates highlighted test output files
- Calls `python3 create_highlighted_versions.py -O highlighted`
- Requires Python 3 and `bat` in PATH

#### `tests/syntax-tests/test_custom_assets.sh` (bash)
- Tests custom asset loading functionality
- Exercises `bat cache --build` and `bat --no-custom-assets`

#### `tests/scripts/license-checks.sh` (bash)
- Checks that no GPL-licensed code is included
- Uses `git grep --recurse-submodules` on submodules

#### `tests/benchmarks/run-benchmarks.sh` (bash)
- Performance benchmarks using `hyperfine`
- Requires: `hyperfine`, `jq`, `python3`
- Tests startup time, syntax highlighting speed, many-file performance

#### `diagnostics/info.sh` (bash)
- Collects system/bat diagnostic info for bug reports
- Interactive script with consent prompts

---

### CI Workflows

#### `CICD.yml` — Main CI/CD Pipeline
Triggers: push to `master`, PRs, workflow_dispatch, tag pushes

**Jobs:**

| Job | OS | What it does |
|-----|-----|-------------|
| `crate_metadata` | ubuntu | Extracts crate name, version, maintainer, homepage, MSRV from Cargo.toml |
| `lint` | ubuntu | `cargo fmt -- --check` + `cargo clippy --locked --all-targets --all-features -- -D warnings` |
| `min_version` | ubuntu | Tests with MSRV (rust-version from Cargo.toml) using reduced feature set |
| `license_checks` | ubuntu | Runs `tests/scripts/license-checks.sh` with submodules |
| `test_with_new_syntaxes_and_themes` | ubuntu | Full asset rebuild + test + regression test + custom assets test |
| `test_with_system_config` | ubuntu | Tests with `BAT_SYSTEM_CONFIG_PREFIX` env var set |
| `documentation` | ubuntu | `cargo doc` with `-D warnings` via `RUSTDOCFLAGS` |
| `cargo-audit` | ubuntu | Security audit via `cargo audit` |
| `build` | matrix (13 targets) | Cross-compile, test, feature checks, package creation (tar.gz + .deb) |
| `all-jobs` | ubuntu | Aggregator gate — fails if any upstream job fails |
| `winget` | ubuntu | Publish to Windows Package Manager (on tag push only) |

**Key CI notes**:
- Uses `dtolnay/rust-toolchain@stable` for Rust installation (not `actions/setup-rust`)
- Cross-compilation uses `cross` (installed from Git)
- Build matrix covers 13 target triples across Linux/macOS/Windows
- Debian packages (.deb) are built for all Linux targets
- Man pages and shell completions are extracted from build artifacts
- Release is triggered by tag push (`v*`) via `softprops/action-gh-release@v2`
- No GoReleaser — pure Cargo + cross + custom packaging scripts

#### `require-changelog-for-PRs.yml` — Changelog Check
Triggers: PRs

Checks that every PR includes a CHANGELOG.md entry with the PR number and author username. Skips dependabot PRs.

---

### Required Tools/Runtimes

| Tool | Version | Purpose | Needed for |
|------|---------|---------|------------|
| Rust (rustc) | >= 1.88 (MSRV) | Compile bat | Build, test |
| Cargo | (comes with Rust) | Build system, dependency management | Build, test |
| rustfmt | (comes with Rust toolchain) | Code formatting | Lint |
| clippy | (comes with Rust toolchain) | Linting | Lint |
| cargo-audit | latest | Security vulnerability scanning | Audit |
| cross | latest | Cross-compilation for non-native targets | CI cross-compile |
| Python 3 | any | Syntax regression test scripts, benchmarks | Test |
| git | any | Submodule management for assets | Asset rebuild |
| hyperfine | latest | Benchmarking | Benchmarks |
| jq | latest | Benchmark result parsing | Benchmarks |
| bat | (self) | Syntax regression tests use bat binary | Test |

**Not required for basic dev workflow**: Nix/direnv (optional), Docker (not used), Ruby (not used unlike fzf).

---

### PI Feature Coverage Gap Analysis

#### Can PI model today ✓

| Workflow | PI automation | Notes |
|----------|---------------|-------|
| `cargo build --bins` | `bash: cargo build --bins` | Single bash step |
| `cargo build --release` | `bash: cargo build --locked --release` | Single bash step |
| `cargo test` | `bash: cargo test --locked` | Single bash step |
| `cargo test --release` | `bash: cargo test --locked --release` | Single bash step |
| `cargo test -- --ignored` | `bash: cargo test ... -- --ignored` | Single bash step |
| System config test | `bash: cargo test ...` with `env: BAT_SYSTEM_CONFIG_PREFIX=...` | Uses step-level `env:` |
| MSRV test | `bash: cargo test --locked <msrv-features>` | Single bash step |
| `cargo fmt -- --check` | `bash: cargo fmt -- --check` | Single bash step |
| `cargo fmt` | `bash: cargo fmt` | Single bash step |
| `cargo clippy ...` | `bash: cargo clippy ...` | Single bash step |
| `cargo doc` | `bash: cargo doc ...` with `env: RUSTDOCFLAGS=-Dwarnings` | Uses step-level `env:` |
| `cargo audit` | `bash: cargo audit` | Single bash step |
| Asset rebuild | `bash: bash assets/create.sh` | Single bash step |
| License checks | `bash: tests/scripts/license-checks.sh` | Single bash step |
| Syntax regression test | `bash: tests/syntax-tests/regression_test.sh` | Single bash step |
| Custom assets test | `bash: tests/syntax-tests/test_custom_assets.sh` | Single bash step |
| Benchmarks | `bash: tests/benchmarks/run-benchmarks.sh --release` | Single bash step |
| Feature checks | Multiple bash steps | 5 `cargo check` steps with different feature combos |
| Install locally | `bash: cargo install --path . --locked` | Single bash step |
| Setup: install Rust | `install:` automation | Can be local `.pi/` automation using `rustup` |
| Setup: install cargo-audit | `install:` automation | Local install automation |
| Setup: install hyperfine | `install:` automation | Local install automation |
| Shell shortcuts | `pi.yaml → shortcuts:` | Works for build/test/lint/etc |

#### PI feature gaps identified ✗

| Gap | Severity | Description |
|-----|----------|-------------|
| **No `pi:install-rust` built-in** | Medium | PI has `install-python`, `install-node`, `install-go` but not Rust. Rust is a major language and bat (a 50k+ star project) uses it. Can be modeled as a local `.pi/` install automation using `rustup`, but Rust is common enough to warrant a built-in. |
| **No `pi:install-ruby` built-in** | Very Low | Unlike fzf, bat doesn't need Ruby at all. Still no built-in, but no urgency from this project. |

#### Key Differences from fzf Adoption Test

| Aspect | fzf | bat |
|--------|-----|-----|
| Language | Go | Rust |
| Build system | Makefile (10 targets) | Cargo only (no Makefile) |
| Build tool | `go build` with ldflags | `cargo build` / `cargo install` |
| Cross-compile | Makefile multi-arch targets | `cross` tool in CI |
| Test runner | Go tests + Ruby integration | Cargo tests only (Rust) |
| Lint | gofmt + rubocop + shfmt | rustfmt + clippy |
| CI complexity | 8 workflows | 2 workflows (1 main + 1 changelog check) |
| Release | GoReleaser | Manual packaging + `softprops/action-gh-release` |
| Shell integration | Complex (completion + key bindings) | None (bat has no shell integration) |
| Feature flags | Build tags via env var | Cargo features (native to Cargo) |
| Asset management | None | Git submodules + build script |
| Required runtimes | Go, Ruby, shfmt | Rust only (Python for some tests) |
| Extra tools | goreleaser, tmux | cargo-audit, cross, hyperfine |

#### Summary

**PI can model 100% of bat's developer workflows today** using bash steps. The project is actually simpler to model than fzf because:

1. **No Makefile** — all workflows are direct Cargo commands, which map cleanly to single bash steps
2. **No Ruby/tmux dependency** — simpler toolchain
3. **Fewer CI workflows** — 2 vs 8 for fzf
4. **Feature flags are native to Cargo** — no need for env var injection to control features (though `env:` is useful for `RUSTDOCFLAGS`)

The only real gap is the missing `pi:install-rust` built-in, which is medium priority since Rust is a major ecosystem. All other gaps identified in the fzf test (`env:` on steps, `pi:install-go`) have already been implemented.

**Conclusion**: PI's feature set is mature enough to handle both Go and Rust projects. The `env:` step support and existing built-in installer framework make it straightforward to model Cargo-based workflows. A `pi:install-rust` built-in would complete the major-language coverage (Python, Node, Go, Rust).

## Subtasks
- [x] Clone repo
- [x] Document Cargo/build config
- [x] Document CI workflows
- [x] List required tools
- [x] Assess PI feature coverage
- [x] Document gaps

## Blocked By
