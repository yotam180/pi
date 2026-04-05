# Update Built-in and Example Automations to New Concise Syntax

## Type
chore

## Status
done

## Priority
medium

## Project
12-yaml-ergonomics

## Description
Once the ergonomics tasks (60–66) are implemented, update all built-in automations (`internal/builtins/embed_pi/`) and all example automations (`examples/`) to use the new PI-flavored concise style. This serves as both a correctness check (if the new syntax works in the real files, it works) and a reference for users reading the built-ins.

Apply all applicable improvements:
- Remove `name:` fields (task 60)
- Use single-step shorthand where the automation has only one step (task 61)
- Replace `pipe_to: next` with `pipe: true` (task 63)
- Replace `PI_INPUT_*` with `PI_IN_*` in step bodies (task 64)
- Use `first:` blocks in installer `run:` sections to replace the compounding negation cascade (task 62)
- Apply automation-level `env:` where multiple steps share the same vars (task 66)
- Use bare paths in all `pi.yaml` setup blocks (task 65)

Also update `docs/README.md` to reflect the new canonical syntax throughout all examples.

## Acceptance Criteria
- [x] All files in `internal/builtins/embed_pi/` use new syntax (no `name:`, `PI_IN_*`, `first:`, `pipe: true`)
- [x] All files in `examples/` use new syntax
- [x] All `pi.yaml` files in examples use bare setup paths
- [x] `docs/README.md` examples updated to canonical new style
- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] `pi validate` on all example workspaces passes with no warnings

## Implementation Notes

### Changes applied to builtins (`internal/builtins/embed_pi/`)
- Removed `name:` from all 13 YAML files
- Converted `hello.yaml`, all docker automations, `cursor/install-extensions/automation.yaml`, and `git/install-hooks.yaml` to single-step shorthand
- Converted `install-python`, `install-node`, `install-go` `run:` phases from cascading `if:` conditions to `first:` blocks
- No `pipe_to:` or `PI_INPUT_*` changes needed (already used new syntax)

### Changes applied to examples (`examples/`)
- Removed `name:` from 72 automation YAML files
- Converted 39 single-step automations to shorthand
- No `pipe_to:`, `PI_INPUT_*`, or setup bare path changes needed (already used new syntax)

### Test updates
- Updated `builtins_test.go` to handle `first:` blocks in install-python/node/go run phases — added `collectAllSteps()` helper that flattens `first:` sub-steps
- Added `parent_shell` and `with` support to shorthand parsing in `automation.go` (needed for `cd-tmp.yaml` and `caller.yaml`)
- Added corresponding unit tests: `TestLoad_ShorthandWithParentShell`, `TestLoad_ShorthandRunWithWith`
- All 826+ tests pass

### `docs/README.md` updates
- Removed `name:` from installer examples (install-homebrew, install-python)
- Converted install-python `run:` to `first:` block style
- Converted install-cursor-extensions and install-uv examples to shorthand
- Added fallback step to install-python `first:` block for completeness

### Notable decisions
- Docker automations keep bash-level branching (can't express `docker compose version` check via PI `if:` predicates)
- `first:` blocks cannot be used as single-step shorthand (only bash/python/typescript/run are shorthand step types) — `pick-platform.yaml` in examples/first-block keeps its `steps:` wrapper
- Rust installer keeps bash-level `if command -v rustup` (single-step, not a multi-step cascade)

## Subtasks
- [x] Update `internal/builtins/embed_pi/*.yaml` files
- [x] Update `examples/**/*.yaml` and `examples/**/pi.yaml` files
- [x] Update `docs/README.md`
- [x] Run full test suite

## Blocked By
60-drop-name-field, 61-single-step-shorthand, 62-first-step-block, 63-pipe-true-syntax, 64-shorter-input-prefix, 65-setup-bare-paths, 66-automation-level-env
