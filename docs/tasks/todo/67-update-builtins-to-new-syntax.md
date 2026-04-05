# Update Built-in and Example Automations to New Concise Syntax

## Type
chore

## Status
todo

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
- [ ] All files in `internal/builtins/embed_pi/` use new syntax (no `name:`, `PI_IN_*`, `first:`, `pipe: true`)
- [ ] All files in `examples/` use new syntax
- [ ] All `pi.yaml` files in examples use bare setup paths
- [ ] `docs/README.md` examples updated to canonical new style
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] `pi validate` on all example workspaces passes with no warnings

## Implementation Notes

## Subtasks
- [ ] Update `internal/builtins/embed_pi/*.yaml` files
- [ ] Update `examples/**/*.yaml` and `examples/**/pi.yaml` files
- [ ] Update `docs/README.md`
- [ ] Run full test suite

## Blocked By
60-drop-name-field, 61-single-step-shorthand, 62-first-step-block, 63-pipe-true-syntax, 64-shorter-input-prefix, 65-setup-bare-paths, 66-automation-level-env
