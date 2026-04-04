# `pi run` and `pi list` Commands

## Type
feature

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Wire the `pi run` and `pi list` CLI commands to the execution engine and automation discovery layer. After this task, PI is end-to-end functional for bash and `run:` steps.

## Acceptance Criteria
- [ ] `pi run <name> [args...]` resolves the automation, runs its steps, and exits with the correct code
- [ ] `pi run` with an unknown name prints a clear error and lists available automations
- [ ] `pi run` without arguments prints usage
- [ ] `pi list` prints a formatted table of all discovered automations: name and description
- [ ] `pi list` with no automations found prints a friendly message (not an error)
- [ ] Both commands walk up the directory tree to find `pi.yaml` (so they work from any subdirectory of the project, like `git` does)
- [ ] `--help` on both commands is informative

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Implement `pi.yaml` root finder (walk up from CWD)
- [ ] Wire `pi run` cobra command to executor
- [ ] Wire `pi list` cobra command to discovery
- [ ] Format `pi list` output cleanly (align columns)
- [ ] Manual smoke test against an `examples/` workspace

## Blocked By
04-step-executor
