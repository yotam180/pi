# Real-World Adoption Test: httpie

## Status
done

## Priority
medium

## Description
Test PI by adopting it in httpie/cli — a popular Python HTTP client (~35k stars) that provides a user-friendly command-line HTTP tool. This is the third real-world adoption test, targeting a Python project to validate PI's support for the Python ecosystem (pip, virtualenvs, pytest, etc.).

httpie is a good candidate because:
- Written in Python (different ecosystem from Go and Rust — tests PI's polyglot claims further)
- Uses Makefile + setup.py/pyproject.toml for build/test
- Has pytest-based test suite
- Has CI workflows
- Uses virtualenvs and pip/pip-tools
- Moderate size
- Exercises PI's existing `pi:install-python` built-in

## Goals
- Validate PI works for Python-based projects
- Identify Python-specific feature gaps
- Create tasks for any missing PI features
- Transform httpie to use PI for its developer workflows
- Compare experience to the Go (fzf) and Rust (bat) adoption tests

## Background & Context
Two adoption tests have been completed:
- fzf (Go) — 100% coverage, surfaced `env:` on steps and `pi:install-go`
- bat (Rust) — 100% coverage, surfaced `pi:install-rust`

This third test targets Python, which is PI's most common user base (many developer tools are Python-based). It will validate that `pi:install-python`, virtualenv management, and pytest workflows all work well.

## Scope

### In scope
- Cloning httpie and examining its workflows
- Creating PI automations for all developer commands
- Creating setup automations for the dev environment
- Defining shell shortcuts for common operations
- Identifying and documenting missing PI features
- Creating tasks for missing features

### Out of scope
- Submitting a PR to httpie to adopt PI
- Modifying httpie's source code
- Windows support

## Success Criteria
- [x] httpie cloned and workflows documented
- [x] Missing PI features identified and tasks created
- [x] `.pi/` folder created with automations for all httpie developer workflows
- [x] `pi.yaml` with shortcuts and setup entries
- [x] `pi setup` installs all required tools
- [x] `pi shell` shortcuts work for common operations (build, test, lint)
- [x] All automations produce identical results to the original commands

## Results

### Summary
All three adoption tests are now complete:

| Project | Language | Automations | Gaps Found | Result |
|---------|----------|-------------|------------|--------|
| fzf | Go | 23 | 3 (env:, install-go, install-ruby) | 100% coverage |
| bat | Rust | 23 | 1 (install-rust) | 100% coverage |
| httpie | Python | 17 | 0 | 100% coverage |

PI's feature set is now validated across Go, Rust, and Python ecosystems with zero remaining feature gaps. Each adoption test drove incremental improvements that benefited subsequent tests.

### Key Findings
1. Zero feature gaps — PI's current feature set is sufficient for Python projects
2. The `install:` lifecycle block handles virtualenv setup naturally
3. Step-level `env:` (from fzf test) works well for Python test configuration
4. All 17 automations produce identical output to the original Makefile commands

## Notes
- Follow `docs/playbooks/real-world-adoption-test.md` for step-by-step process
- Use the development version of `pi` built from this repo
- Compare findings to the fzf and bat adoption tests documented in projects 07 and 08
