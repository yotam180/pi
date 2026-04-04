# Real-World Adoption Test: bat

## Status
todo

## Priority
medium

## Description
Test PI by adopting it in sharkdp/bat — a popular Rust CLI tool (~50k stars) that provides a `cat` replacement with syntax highlighting. This is the second real-world adoption test, targeting a Rust project to surface different feature gaps than the Go-focused fzf test.

bat is a good candidate because:
- Written in Rust (different ecosystem from Go — tests PI's polyglot claims)
- Uses Cargo for build/test (not Make)
- Has CI workflows with cross-platform builds
- Has integration/system tests
- Moderate size (~15k LOC Rust)
- No existing `pi:install-rust` built-in — will likely need one

## Goals
- Validate PI works for Rust/Cargo-based projects
- Identify Rust-specific feature gaps
- Create tasks for any missing PI features (e.g., `pi:install-rust`)
- Transform bat to use PI for its developer workflows
- Compare experience to the fzf adoption test

## Background & Context
The fzf adoption test (project 07) validated PI against a Go project and found that PI can model 100% of its workflows. That test surfaced the need for `env:` on steps and `pi:install-go`. This follow-up tests a fundamentally different ecosystem (Rust/Cargo) to ensure PI's automation model generalizes beyond Go.

## Scope

### In scope
- Cloning bat and examining its workflows
- Creating PI automations for all developer commands
- Creating setup automations for the dev environment
- Defining shell shortcuts for common operations
- Identifying and documenting missing PI features
- Creating tasks for missing features

### Out of scope
- Submitting a PR to bat to adopt PI
- Modifying bat's source code
- Windows support

## Success Criteria
- [ ] bat cloned and workflows documented
- [ ] Missing PI features identified and tasks created
- [ ] `.pi/` folder created with automations for all bat developer workflows
- [ ] `pi.yaml` with shortcuts and setup entries
- [ ] `pi setup` installs all required tools
- [ ] `pi shell` shortcuts work for common operations (build, test, lint)
- [ ] All automations produce identical results to the original commands

## Notes
- Follow `docs/playbooks/real-world-adoption-test.md` for step-by-step process
- Use the development version of `pi` built from this repo
- Compare findings to the fzf adoption test documented in project 07
