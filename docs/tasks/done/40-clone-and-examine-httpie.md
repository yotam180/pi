# Clone and Examine httpie Workflows

## Type
research

## Status
done

## Priority
high

## Project
09-httpie-adoption-test

## Description
Clone httpie/cli into `~/projects/httpie` and examine all developer workflows. Document every build command, test command, lint/format command, CI workflow, and release process. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/httpie/cli.git ~/projects/httpie`
2. Read `setup.py`/`pyproject.toml` — document build config and dependencies
3. Read any Makefile or build scripts
4. Read CI workflows (`.github/workflows/`)
5. Read any release configuration
6. List all tools/runtimes required (Python, pip, pytest, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [x] httpie cloned to `~/projects/httpie`
- [x] Every build/test/lint command documented
- [x] Every CI workflow documented
- [x] Required tools/runtimes listed
- [x] PI feature gap analysis completed
- [x] Findings recorded in Implementation Notes

## Implementation Notes

### Repository Overview

httpie (latest master) is a Python CLI HTTP client (~35k stars) providing a user-friendly alternative to `curl`. Key characteristics:

- **Language**: Python 3.7+
- **Build system**: `setup.cfg` + `setup.py` (declarative metadata, setuptools-based)
- **Package management**: pip with extras (`[dev]`, `[test]`)
- **Virtual environment**: `venv` (stdlib), managed by Makefile
- **Test runner**: pytest with coverage, doctest support, and marker-based skipping
- **Linter**: flake8 with plugins (comprehensions, deprecated, mutable, tuple)
- **Doc checker**: mdl (Ruby-based markdown linter)
- **Benchmarks**: Python `pyperf` via custom scripts in `extras/profiling/`
- **Release**: PyPI via twine, Homebrew, Snap, Chocolatey, Linux standalone binary
- **15 CI workflows** covering testing, linting, coverage, release to 5+ package managers

---

### Makefile Targets

httpie's Makefile is the primary developer interface. It manages a local virtualenv and all workflows.

| Target | Command | Description |
|--------|---------|-------------|
| `list-tasks` (default) | `make` | Lists all available make targets |
| `all` | `make all` | Uninstall + install + test (full rebuild) |
| `install` | `make install` | Create venv + install all dependencies + editable install |
| `install-reqs` | `make install-reqs` | pip install pip/wheel/build, `.[dev]`, `.[test]`, and editable `.` |
| `venv` | `make venv` | Create Python venv in `./venv` with `--prompt httpie` |
| `clean` | `make clean` | Remove venv, caches, build artifacts, `__pycache__` dirs |
| `test` | `make test` | Run `python -m pytest` from venv |
| `test-cover` | `make test-cover` | Run pytest with `--cov=httpie --cov=tests` |
| `test-all` | `make test-all` | clean → install → test → test-dist → codestyle |
| `test-dist` | `make test-dist` | Test sdist and wheel builds |
| `test-sdist` | `make test-sdist` | Build sdist, install, verify `http --version` |
| `test-bdist-wheel` | `make test-bdist-wheel` | Build wheel, install, verify `http --version` |
| `codestyle` | `make codestyle` | Run flake8 on httpie/, tests/, extras/, *.py |
| `build` | `make build` | Build sdist + wheel for PyPI (sets build channel) |
| `publish` | `make publish` | test-all + twine upload to PyPI |
| `publish-no-test` | `make publish-no-test` | Version check + build + twine check + twine upload |
| `uninstall-httpie` | `make uninstall-httpie` | pip uninstall httpie |
| `brew-deps` | `make brew-deps` | Generate Homebrew dependency list via Python script |
| `brew-test` | `make brew-test` | Build and test Homebrew formula locally |
| `content` | `make content` | Regenerate man pages + installation docs |
| `man` | `make man` | Generate man pages via Python script |
| `installation-docs` | `make installation-docs` | Generate installation instructions via Python script |
| `doc-check` | `make doc-check` | Markdown linting via `mdl` (Ruby) |
| `codecov-upload` | `make codecov-upload` | Upload coverage report to Codecov |
| `twine-check` | `make twine-check` | Validate distribution packages with twine |

**Key Makefile design**:
- Uses `SYSTEM_PYTHON=python3` to bootstrap the venv
- Sets `export PATH := $(VENV_BIN):$(PATH)` so all commands use venv Python
- `VENV_ROOT=venv`, `VENV_BIN=$(VENV_ROOT)/bin`, `VENV_PYTHON=$(VENV_BIN)/python`

---

### Package Configuration (`setup.cfg`)

**Dependencies** (install_requires):
- `pip`, `charset_normalizer>=2.0.0`, `defusedxml>=0.6.0`, `requests[socks]>=2.22.0`, `Pygments>=2.5.2`, `requests-toolbelt>=0.9.1`, `multidict>=4.7.0`, `setuptools`, `rich>=9.10.0`
- Platform-specific: `colorama>=0.2.4` on Windows, `importlib-metadata>=1.4.0` on Python<3.8

**Dev extras** (`[dev]`):
- pytest, pytest-httpbin, responses, pytest-mock, werkzeug, flake8, flake8-comprehensions, flake8-deprecated, flake8-mutable, flake8-tuple, pyopenssl, pytest-cov, pyyaml, twine, wheel, Jinja2

**Test extras** (`[test]`):
- pytest, pytest-httpbin, responses, pytest-mock, werkzeug

**Entry points**: `http`, `https`, `httpie` CLI commands

**Flake8 config**: ignore E501 (line too long), W503 (line break before binary operator)

---

### CI Workflows (15 total)

#### Core Development Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `tests.yml` | push to master, PRs | Full test suite on 3 OS × 6 Python versions × 2 pyopenssl modes |
| `code-style.yml` | PRs | flake8 linting on Python 3.9 |
| `coverage.yml` | PRs | pytest-cov → codecov upload + dist tests |
| `benchmark.yml` | PR labeled `benchmark` | Performance benchmarks via pyperf |
| `content.yml` | push to master | Regenerate man pages + docs, auto-PR |
| `docs-check-markdown.yml` | PRs (*.md changes) | Markdown linting via `mdl` (Ruby 2.7) |
| `docs-deploy.yml` | push to master, releases | Trigger doc rebuild on Vercel |
| `stale.yml` | manual dispatch | Mark stale PRs after 30 days |

#### Release Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `release-pypi.yml` | manual dispatch | Build + publish to PyPI via twine |
| `release-brew.yml` | manual dispatch | Update Homebrew formula |
| `release-snap.yml` | manual dispatch | Build + publish Snap package (edge → stable) |
| `release-linux-standalone.yml` | manual dispatch, releases | Build standalone Linux binary + .deb + .rpm |
| `release-choco.yml` | manual dispatch | Build + publish Chocolatey package |

#### Package Testing Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `test-package-mac-brew.yml` | PRs (brew formula changes) | Test Homebrew formula on macOS |
| `test-package-linux-snap.yml` | PRs (snapcraft changes) | Test Snap package on Linux |

**Test matrix** (`tests.yml`):
- OS: ubuntu-latest, macos-13, windows-latest
- Python: 3.7, 3.8, 3.9, 3.10, 3.11, 3.12
- pyopenssl: 0, 1
- Windows runs pip install + pytest directly
- Linux/Mac uses `make install && make test`

---

### Required Tools/Runtimes

| Tool | Version | Purpose | Needed for |
|------|---------|---------|------------|
| Python 3 | >= 3.7 (CI tests 3.7–3.12) | Runtime, build, test | Everything |
| pip | (comes with Python) | Package management | Install, build |
| venv | (stdlib, comes with Python) | Virtual environment | Development setup |
| pytest | latest | Test runner | Testing |
| flake8 | latest | Code linting | Code style |
| twine | latest | Package publishing | Release to PyPI |
| wheel | latest | Build wheels | Building distributions |
| build | latest | PEP 517 build tool | Building distributions |
| Ruby | 2.7 | mdl markdown linter | Doc checking (optional) |
| mdl | latest | Markdown linting | Doc checking (optional) |
| pyperf | >= 2.3.0 | Performance benchmarks | Benchmarking (optional) |
| snapcraft | latest | Snap packaging | Release (CI only) |
| brew | latest | Homebrew formula testing | Release (optional) |
| git | any | Version control | Development |

**Not required for basic dev workflow**: Ruby (only for doc-check), pyperf (only for benchmarks), snapcraft/brew (only for packaging), choco (only for Windows release).

---

### PI Feature Coverage Gap Analysis

#### Can PI model today ✓

| Workflow | PI automation | Notes |
|----------|---------------|-------|
| `make venv` | `bash: python3 -m venv --prompt httpie venv` | Single bash step |
| `make install-reqs` | Multi-step bash | pip upgrade + `.[dev]` + editable install |
| `make install` | `run:` step chaining venv + install-reqs | Two sub-automations via `run:` |
| `make test` | `bash: venv/bin/python -m pytest` | Single bash step |
| `make test-cover` | `bash: venv/bin/python -m pytest --cov=httpie --cov=tests` | Single bash step |
| `make codestyle` | `bash: venv/bin/flake8 httpie/ tests/ extras/ *.py` | Single bash step (install flake8 if missing first) |
| `make clean` | `bash: rm -rf venv *.egg dist build ...` | Single bash step |
| `make build` | Multi-step bash | Clean + set build channel + build + restore |
| `make test-sdist` | Multi-step bash | clean + venv + build sdist + install + verify |
| `make test-bdist-wheel` | Multi-step bash | clean + venv + build wheel + install + verify |
| `make test-all` | `run:` step chaining clean + install + test + test-dist + codestyle | Chain of sub-automations |
| `make publish` | Multi-step bash | Version check + build + twine upload |
| `make content` | `run:` step chaining man + installation-docs | Two sub-automations |
| `make man` | `bash: venv/bin/python extras/scripts/generate_man_pages.py` | Single bash step |
| `make installation-docs` | `bash: venv/bin/python docs/installation/generate.py` | Single bash step |
| `make brew-test` | Multi-step bash | brew uninstall + install HEAD + verify + audit |
| `make doc-check` | `bash: mdl --git-recurse --style docs/markdownlint.rb .` | Single bash step (requires Ruby + mdl) |
| `make uninstall-httpie` | `bash: venv/bin/pip3 uninstall --yes httpie` | Single bash step |
| `make codecov-upload` | `bash: venv/bin/codecov` | Single bash step |
| Setup: install Python | `pi:install-python` | Built-in installer |
| Shell shortcuts | `pi.yaml → shortcuts:` | Works for test/lint/build/etc |

#### PI feature gaps identified ✗

| Gap | Severity | Description |
|-----|----------|-------------|
| None | — | PI can model 100% of httpie's workflows today |

#### Key Observations

1. **All workflows reduce to bash steps**: httpie's Makefile targets are all bash commands that can be directly modeled as PI `bash:` steps. No new step types needed.

2. **Virtualenv is implicit**: The Makefile uses `venv/bin/python` and `venv/bin/pip3` explicitly. PI doesn't need to "know about" virtualenvs — the bash steps just reference the venv-local binaries directly. This is clean and doesn't require any new PI feature.

3. **`pi:install-python` already exists**: PI's built-in Python installer handles the primary runtime dependency. The rest (pip, venv) come with Python.

4. **No new built-ins needed**: Unlike fzf (which surfaced `pi:install-go` need) and bat (which surfaced `pi:install-rust` need), httpie doesn't require any new built-in automations. Python is already covered.

5. **Ruby dependency is optional**: `mdl` is only used for markdown linting and is completely optional for development. Not worth creating a built-in for.

6. **Release workflows are CI-only**: The 5+ release workflows (PyPI, Homebrew, Snap, Chocolatey, Linux standalone) are all workflow_dispatch or release-triggered. They don't need to be modeled as PI automations — they're GitHub Actions concerns.

#### Key Differences from fzf and bat Adoption Tests

| Aspect | fzf | bat | httpie |
|--------|-----|-----|--------|
| Language | Go | Rust | Python |
| Build system | Makefile (10 targets) | Cargo (no Makefile) | Makefile (22 targets) + setup.cfg |
| Package manager | go modules | Cargo | pip + setuptools |
| Virtual env | N/A | N/A | venv (stdlib) |
| Test runner | `go test` + Ruby integration | `cargo test` | pytest |
| Lint | gofmt + rubocop + shfmt | rustfmt + clippy | flake8 |
| Release targets | GoReleaser | Custom packaging | PyPI + Homebrew + Snap + Chocolatey + Linux |
| CI complexity | 8 workflows | 2 workflows | 15 workflows |
| PI gaps found | 3 (env:, install-go, install-ruby) | 1 (install-rust) | 0 |
| Required tools | Go, Ruby, shfmt, goreleaser | Rust, cargo-audit, cross | Python, pip (optional: Ruby for mdl) |

#### Summary

**PI can model 100% of httpie's developer workflows today** with zero feature gaps. This is the cleanest adoption test so far:

1. **Python is PI's sweet spot**: `pi:install-python` is already a built-in. Python's `venv` is stdlib and trivially managed via bash steps.
2. **Setup.cfg + pip workflow is simple**: Install deps via `pip install '.[dev]'` — no complex toolchain management.
3. **Everything is bash**: Even the most complex workflow (publish to PyPI) is a sequence of bash commands.
4. **The env: and install: features added after the fzf test** make httpie modeling even simpler (e.g., `HTTPIE_TEST_WITH_PYOPENSSL` env var for conditional testing).

This validates that PI is mature for Python projects. The previous adoption tests drove all the features needed (step-level `env:`, installer lifecycle, built-in automations for major languages).

## Subtasks
- [x] Clone repo
- [x] Document build/package config
- [x] Document CI workflows
- [x] List required tools
- [x] Assess PI feature coverage
- [x] Document gaps

## Blocked By
