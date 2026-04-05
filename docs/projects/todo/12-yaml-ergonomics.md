# YAML Ergonomics

## Status
todo

## Priority
high

## Description
Make PI automation YAML files more concise, readable, and developer-friendly. The current format has several sources of unnecessary ceremony that were identified by reviewing real-world adoption tests (fzf, bat, httpie, zx). The fixes are additive — existing YAML continues to work, new shorthands and primitives make writing automations faster and cleaner.

## Goals
- Remove the redundant `name:` field — PI derives it from the file path
- Allow single-step automations to skip the `steps:` wrapper
- Add `first:` step block for mutually exclusive conditional branches (replaces compounding negation chains)
- Rename `pipe_to: next` → `pipe: true` (cleaner, matches the actual semantics)
- Shorten input variable prefix from `PI_INPUT_*` to `PI_IN_*`
- Allow `setup:` entries in `pi.yaml` to be bare paths instead of `run:` objects
- Support automation-level `env:` that applies to all steps
- Update all built-in automations to use the new concise style

## Background & Context
The adoption tests (fzf, bat, httpie, zx) produced real `.pi/` folders with real automation YAML files. Reviewing those files surfaced consistent pain points:

- Every file starts with a `name:` line that duplicates the file path
- Single-command automations require 3-4 lines of scaffolding for 1 line of content
- Installer files contain a repetitive 8-line cascade of `if: command.X` / `if: command.Y and not command.X` / `if: not command.X and not command.Y` that buries the actual install commands
- `pipe_to: next` — `next` is the only valid value, making the field name misleading
- `PI_INPUT_VERSION` in bash strings is verbose, especially in multi-input scripts
- `setup:` in `pi.yaml` forces `- run: path` even though every entry is always a run

All fixes are backward-compatible with a deprecation path: old syntax continues to work, new syntax is preferred and linted.

## Scope

### In scope
- Parser/schema changes for all new shorthands
- `first:` step block implementation (executor + parser)
- Deprecation warnings for `name:` (if present but derivable), `pipe_to: next`, and `PI_INPUT_*`
- Updating all built-in automations (`.pi/` and `internal/builtins/embed_pi/`) to new style
- Updating all example automations to new style
- Updating `docs/README.md` to document new syntax

### Out of scope
- Removing old syntax support (deprecate, not break)
- Marketplace or CLI changes unrelated to YAML format

## Success Criteria
- [x] `name:` field is optional; PI derives it from file path when absent
- [x] A single-step automation can be written as a top-level `bash:` / `python:` / `typescript:` / `run:` key
- [ ] `first:` block works in any step list: runs the first sub-step whose `if:` passes, skips rest
- [ ] `pipe: true` is the canonical form; `pipe_to: next` still works with deprecation warning
- [ ] `$PI_IN_*` resolves inputs; `$PI_INPUT_*` still works with deprecation warning
- [ ] `setup:` entries in `pi.yaml` accept bare strings (`- setup/install-go`)
- [ ] Automation-level `env:` applies to all steps in the automation
- [ ] All built-in and example YAML files use the new concise style
- [ ] All existing tests pass; new tests cover each new feature

## Notes
- Maintain backward compatibility throughout — real-world users already have `.pi/` folders
- The `first:` block is the highest-complexity item; implement and test it independently
- Consider a `pi validate` warning mode that flags old-style syntax
