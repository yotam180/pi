# Real-World Adoption Test: fzf

## Status
done

## Priority
medium

## Description
Test PI by adopting it in junegunn/fzf — a popular Go CLI tool (68k+ stars) with a Makefile, install scripts, shell integration, multi-arch builds, and Ruby-based integration tests. This validates PI's feature set against a real project and surfaces missing capabilities.

## Goals
- Identify feature gaps by examining fzf's developer workflow
- Create tasks for any missing PI features discovered
- Transform fzf to use PI for its developer workflows
- Validate that `pi setup`, `pi run`, and `pi shell` work correctly for a real project

## Background & Context
PI's feature set has been built based on hypothetical workflows. Testing against a real, well-known project ensures that the automation model works in practice. fzf is a good candidate because:
- Written in Go (matching PI's primary audience)
- Has a Makefile with build, test, lint, format commands
- Has install/uninstall shell scripts
- Has shell integration (bash, zsh, fish completion and key bindings)
- Has multi-arch cross-compilation
- Has Ruby-based integration tests (exercises polyglot support)
- Has goreleaser for releases
- Is not too large (~15k LOC Go)

## Scope

### In scope
- Cloning fzf and examining its workflows
- Creating PI automations for all developer commands
- Creating setup automations for the dev environment
- Defining shell shortcuts for common operations
- Identifying and documenting missing PI features
- Creating tasks for missing features

### Out of scope
- Actually submitting a PR to fzf to adopt PI
- Modifying fzf's source code
- Windows support

## Success Criteria
- [x] fzf cloned and workflows documented
- [x] Missing PI features identified and tasks created
- [x] `.pi/` folder created with automations for all fzf developer workflows
- [x] `pi.yaml` with shortcuts and setup entries
- [x] `pi setup` installs all required tools
- [x] `pi shell` shortcuts work for common operations (build, test, lint)
- [x] All automations produce identical results to the original commands

## Results

### Features Added to PI During This Project
1. **`env:` on steps** — environment variable injection per step, scoped and isolated
2. **`pi:install-go`** — built-in installer automation for Go
3. **`go version` fallback** — version detection now tries `<cmd> version` as fallback when `--version` fails

### Key Finding
PI can model 100% of fzf's developer workflows. 18 automations were created covering build, test, lint, format, Docker, setup, and utility workflows. All major automations tested and produce identical results to the original Makefile commands.

## Notes
- Follow `docs/playbooks/real-world-adoption-test.md` for step-by-step process
- Use the development version of `pi` built from this repo
- Document findings about PI's strengths and weaknesses in the task files
