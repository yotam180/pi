# Real-World Adoption Test: zx

## Status
todo

## Priority
medium

## Description
Test PI by adopting it in google/zx — a popular TypeScript tool (~43k stars) for writing shell scripts in JavaScript/TypeScript. This is the fourth real-world adoption test, targeting a TypeScript/Node.js project to validate PI's support for the JavaScript ecosystem (npm, TypeScript, tsx, testing frameworks).

zx is a good candidate because:
- Written in TypeScript (new ecosystem — tests PI's TypeScript step support and `pi:install-node`/`pi:install-tsx` built-ins)
- Uses npm for package management
- Has its own build system and test suite
- Has CI workflows
- Moderate size (~5k LOC)
- Thematically aligned — zx is about shell scripting, PI is about developer automation

## Goals
- Validate PI works for TypeScript/Node.js projects
- Identify JavaScript/TypeScript-specific feature gaps
- Exercise PI's TypeScript step type in a real context
- Create tasks for any missing PI features
- Transform zx to use PI for its developer workflows
- Compare experience to the Go (fzf), Rust (bat), and Python (httpie) adoption tests

## Background & Context
Three adoption tests have been completed:
- fzf (Go) — 100% coverage, surfaced `env:` on steps and `pi:install-go`
- bat (Rust) — 100% coverage, surfaced `pi:install-rust`
- httpie (Python) — 100% coverage, zero new gaps

This fourth test targets TypeScript/Node.js to complete coverage of PI's four supported step types (bash, python, typescript, run). It will validate that `pi:install-node`, `pi:install-tsx`, npm workflows, and TypeScript testing all work well.

## Scope

### In scope
- Cloning zx and examining its workflows
- Creating PI automations for all developer commands
- Creating setup automations for the dev environment
- Defining shell shortcuts for common operations
- Identifying and documenting missing PI features
- Creating tasks for missing features

### Out of scope
- Submitting a PR to zx to adopt PI
- Modifying zx's source code
- Windows support

## Success Criteria
- [ ] zx cloned and workflows documented
- [ ] Missing PI features identified and tasks created
- [ ] `.pi/` folder created with automations for all zx developer workflows
- [ ] `pi.yaml` with shortcuts and setup entries
- [ ] `pi setup` installs all required tools
- [ ] `pi shell` shortcuts work for common operations (build, test, lint)
- [ ] All automations produce identical results to the original commands

## Notes
- Follow `docs/playbooks/real-world-adoption-test.md` for step-by-step process
- Use the development version of `pi` built from this repo
- Compare findings to the fzf, bat, and httpie adoption tests documented in projects 07, 08, and 09
