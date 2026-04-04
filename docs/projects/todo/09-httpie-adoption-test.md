# Real-World Adoption Test: httpie

## Status
todo

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
- [ ] httpie cloned and workflows documented
- [ ] Missing PI features identified and tasks created
- [ ] `.pi/` folder created with automations for all httpie developer workflows
- [ ] `pi.yaml` with shortcuts and setup entries
- [ ] `pi setup` installs all required tools
- [ ] `pi shell` shortcuts work for common operations (build, test, lint)
- [ ] All automations produce identical results to the original commands

## Notes
- Follow `docs/playbooks/real-world-adoption-test.md` for step-by-step process
- Use the development version of `pi` built from this repo
- Compare findings to the fzf and bat adoption tests documented in projects 07 and 08
