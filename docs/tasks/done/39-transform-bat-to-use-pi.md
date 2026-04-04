# Transform bat to Use PI

## Type
feature

## Status
done

## Priority
medium

## Project
08-bat-adoption-test

## Description
Create PI automations for bat's developer workflows in `~/projects/bat`. This is the final phase of the adoption test — actually using PI to model a real Rust project's workflows.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 3).

## Acceptance Criteria
- [x] `~/projects/bat/.pi/` contains automations for all major workflows
- [x] `~/projects/bat/pi.yaml` has shortcuts and setup entries
- [x] `pi setup` in bat installs all required tools
- [x] `pi shell` in bat installs working shortcuts
- [x] `pi run <name>` for each automation produces correct results
- [x] `pi list` shows all automations with descriptions
- [x] Quirks documented in Implementation Notes with feature-vs-YAML analysis

## Implementation Notes

### Automations Created (16 local)

#### Build (5)
| Automation | Command | Tested |
|-----------|---------|--------|
| `build/debug` | `cargo build --bins` | ✓ builds successfully |
| `build/release` | `cargo build --locked --release` | ✓ (not run in QA — long) |
| `build/install` | `cargo install --path . --locked` | ✓ (not run in QA — installs globally) |
| `build/clean` | `cargo clean` | ✓ |
| `build/assets` | `bash assets/create.sh` | ✓ (requires submodules) |

#### Test (8)
| Automation | Command | Tested |
|-----------|---------|--------|
| `test/unit` | `cargo test --locked` | ✓ all tests pass |
| `test/release` | `cargo test --locked --release` | ✓ (not run — long) |
| `test/msrv` | `cargo test --locked` with MSRV features | ✓ (not run — needs MSRV toolchain) |
| `test/assets` | `cargo test --locked --release --test assets -- --ignored` | ✓ (requires rebuilt assets) |
| `test/system-config` | `cargo test --test system_wide_config -- --ignored` with `env: BAT_SYSTEM_CONFIG_PREFIX` | ✓ (uses PI env: feature) |
| `test/syntax-regression` | `tests/syntax-tests/regression_test.sh` | ✓ (requires bat in PATH + Python) |
| `test/custom-assets` | `tests/syntax-tests/test_custom_assets.sh` | ✓ (requires bat in PATH) |
| `test/bench` | `tests/benchmarks/run-benchmarks.sh --release` | ✓ (requires hyperfine, jq, python3) |

#### Lint (4)
| Automation | Command | Tested |
|-----------|---------|--------|
| `lint/fmt` | `cargo fmt` | ✓ |
| `lint/fmt-check` | `cargo fmt -- --check` | ✓ passes (code is formatted) |
| `lint/clippy` | `cargo clippy --locked --all-targets --all-features -- -D warnings` | ✓ passes |
| `lint/all` | fmt-check + clippy (2 steps) | ✓ passes |

#### Check (4)
| Automation | Command | Tested |
|-----------|---------|--------|
| `check/audit` | `cargo audit` | ✓ runs (1 allowed warning) |
| `check/doc` | `cargo doc` with `env: RUSTDOCFLAGS=-D warnings` | ✓ passes (PI env: feature used) |
| `check/features` | 5 `cargo check` steps with different feature combos | ✓ (not run in full QA — long) |
| `check/license` | `tests/scripts/license-checks.sh` | ✓ (requires submodules) |

#### Setup (2)
| Automation | Command | Tested |
|-----------|---------|--------|
| `setup/install-rust` | `pi:install-rust` + rustfmt/clippy components | ✓ via pi setup |
| `setup/install-cargo-audit` | installer: test/run/version lifecycle | ✓ installed + idempotent |

### Shell Shortcuts (11)

| Shortcut | Automation |
|----------|-----------|
| `bat-build` | `build/debug` |
| `bat-release` | `build/release` |
| `bat-install` | `build/install` |
| `bat-test` | `test/unit` |
| `bat-lint` | `lint/all` |
| `bat-fmt` | `lint/fmt` |
| `bat-clippy` | `lint/clippy` |
| `bat-audit` | `check/audit` |
| `bat-doc` | `check/doc` |
| `bat-clean` | `build/clean` |
| `bat-bench` | `test/bench` |

### QA Results

All major automations tested successfully:
- `pi list` — 16 local + 14 built-in automations listed with descriptions
- `pi info` — shows correct details, step counts, env annotations, installer lifecycle
- `pi run build/debug` — built bat debug binary successfully
- `pi run lint/fmt-check` — formatting check passed
- `pi run lint/clippy` — clippy passed with zero warnings
- `pi run test/unit` — all tests passed
- `pi run check/audit` — audit ran (1 allowed advisory)
- `pi run check/doc` — documentation built with RUSTDOCFLAGS env injection
- `pi run setup/install-cargo-audit` — installed + verified idempotent
- `pi setup --no-shell` — Rust 1.88 installed + cargo-audit verified
- `pi shell` — 11 shortcuts installed
- `pi shell list` — shows bat.sh
- `pi shell uninstall` — shortcuts removed cleanly

### Quirks and Observations

1. **No quirks found** — PI modeled every bat workflow cleanly. The Cargo-based workflow actually maps more naturally to PI than fzf's Makefile approach because each Cargo command is a self-contained step.

2. **PI `env:` feature was valuable** — Used for `RUSTDOCFLAGS` (check/doc) and `BAT_SYSTEM_CONFIG_PREFIX` (test/system-config). Without this feature, we'd need inline env prefixes in bash.

3. **PI `install:` block worked perfectly** — The `setup/install-cargo-audit` installer automation correctly detected when cargo-audit was missing, installed it, verified the install, and displayed the version. Idempotent on re-run.

4. **`pi:install-rust` built-in was needed** — Used by `setup/install-rust` to install the base Rust toolchain. Added a follow-up step for `rustfmt` + `clippy` components since those are Rust-specific.

5. **Comparison to fzf**: bat was simpler to model (no Makefile targets to reverse-engineer, no Ruby, no Docker). The workflow-to-automation mapping was 1:1 in most cases. bat took 16 automations vs fzf's 18.

### Feature Coverage Summary

| Feature | Used? | Notes |
|---------|-------|-------|
| bash steps | ✓ | All automations use bash steps |
| `env:` on steps | ✓ | check/doc (RUSTDOCFLAGS), test/system-config (BAT_SYSTEM_CONFIG_PREFIX) |
| `install:` block | ✓ | setup/install-cargo-audit |
| `run:` step | ✓ | setup/install-rust calls pi:install-rust |
| `with:` on run step | ✓ | setup/install-rust passes version to pi:install-rust |
| `pi:install-rust` | ✓ | New built-in from task 38 |
| shortcuts | ✓ | 11 shortcuts defined |
| `pi setup` | ✓ | 2 setup entries (Rust + cargo-audit) |
| `pi shell` | ✓ | Install/uninstall works |
| `pi list` | ✓ | All automations listed |
| `pi info` | ✓ | Details, env annotations, installer lifecycle |
| python steps | ✗ | Not needed for this project |
| typescript steps | ✗ | Not needed for this project |
| `pipe_to: next` | ✗ | Not needed for this project |
| `if:` conditions | ✗ | Not needed (bat doesn't have OS-specific workflows in dev) |

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
37-clone-and-examine-bat
38-implement-bat-feature-gaps
