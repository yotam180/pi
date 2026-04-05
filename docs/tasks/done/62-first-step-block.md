# `first:` Step Block

## Type
feature

## Status
done

## Priority
high

## Project
12-yaml-ergonomics

## Description
Add a `first:` step block — a group of sub-steps where only the first sub-step whose `if:` condition passes is executed. All remaining sub-steps are skipped.

This replaces the common and painful pattern of compounding negation chains used in installer files and other mutually exclusive branch scenarios:

**Before (compounding negations):**
```yaml
run:
  - bash: mise install go@1.23 && mise use go@1.23
    if: command.mise
  - bash: brew install go
    if: command.brew and not command.mise
  - bash: |
      echo "no suitable installer found" >&2
      exit 1
    if: not command.mise and not command.brew
```

**After (`first:` block):**
```yaml
run:
  - first:
      - bash: mise install go@1.23 && mise use go@1.23
        if: command.mise
      - bash: brew install go
        if: command.brew
      - bash: |
          echo "no suitable installer found" >&2
          exit 1
```

Each sub-step only needs its own positive condition. The last sub-step with no `if:` acts as a natural fallback. If all sub-steps have `if:` conditions and none match, the `first:` block is silently skipped (consistent with how a skipped `if:` step behaves).

`first:` is a generic step-level construct usable anywhere a step appears: in `steps:`, in `install.run:`, in `install.test:`. It is not specific to installers.

`first:` itself can optionally have a `description:` field (shown by `pi info`) but does not support `env:`, `dir:`, `timeout:`, or `pipe:` at the block level — those belong on the individual sub-steps.

## Acceptance Criteria
- [x] `first:` block is valid anywhere a step appears (`steps:`, `install.run:`, `install.test:`)
- [x] Exactly the first sub-step whose `if:` condition passes is executed; all others skipped
- [x] A sub-step with no `if:` always matches and acts as a fallback
- [x] If no sub-step matches (all have `if:` and all fail), the block is silently skipped
- [x] `first:` accepts an optional `description:` for `pi info` output
- [x] Sub-steps inside `first:` support all normal step fields: `env:`, `dir:`, `timeout:`, `silent:`, `pipe:`
- [x] `pipe_to: next` on a `first:` block correctly pipes output to the next step in the parent list
- [x] `pi info` renders `first:` blocks sensibly (shows description or sub-step summary)
- [x] Parse error if `first:` contains zero sub-steps
- [x] Tests cover: first matches, middle matches, fallback matches, nothing matches, `pipe:` through `first:`

## Implementation Notes

### Design decisions
- `first:` is NOT a new step type — it's a container field on `Step`. A step with `First != nil` is a first-match block. This avoids adding a new `StepType` constant and keeps the step type system clean.
- `Step.IsFirst()` returns true when `First` is non-nil.
- Nested `first:` blocks are explicitly disallowed (parse-time error) — complexity with no clear use case.
- Block-level modifiers (`env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`, `with:`) are rejected at parse time with clear messages pointing to sub-steps. Only `description:`, `if:`, and `pipe_to:` are valid at the block level.
- When a piped `first:` block has no matching sub-step, an empty pipe buffer is set so downstream steps get empty stdin (not stale data from a prior step).
- `pi validate` traverses into `first:` blocks to check `run:` step references.

### Files changed
- `internal/automation/step.go` — `Step.First` field, `Step.IsFirst()`, `stepRaw.First`, `toFirstStep()`, `validateFirstBlock()`, updated `validateSteps()` and `validateInstallPhase()`
- `internal/executor/executor.go` — `execFirstBlock()`, integration into main step loop, empty pipe buffer on no-match
- `internal/executor/install.go` — `execInstallFirstBlock()` for install phase handling
- `internal/cli/info.go` — `printFirstBlockDetail()` with lettered sub-step rendering (a, b, c)
- `internal/cli/validate.go` — traverse into `first:` blocks for `run:` step reference validation
- `internal/automation/step_test.go` — 13 new parsing tests
- `internal/executor/first_block_test.go` — 14 new executor tests
- `tests/integration/first_block_integ_test.go` — 8 new integration tests
- `examples/first-block/` — example workspace with 5 automation files

### Test counts
- 13 parsing unit tests (basic, description, pipe_to, outer if, empty error, type conflict, env/dir/timeout/silent/parent_shell on block error, sub-step env, install phase, nested error, mixed steps)
- 14 executor unit tests (first/middle/fallback matches, none match, mixed steps, pipe to next, pipe no match, outer if skip, run sub-step, exit error, install phase, silent, loud override)
- 8 integration tests (list, pick-platform, no-match, with-pipe, mixed, info, validate, installer)

## Subtasks
- [x] Define `FirstBlock` type in the automation schema
- [x] Update parser to recognize `first:` as a step-level key
- [x] Implement executor logic: evaluate `if:` conditions in order, run first match
- [x] Handle `pipe:` correctly when `first:` is a piped step
- [x] Add `pi info` rendering
- [x] Add tests
- [x] Add example automation demonstrating `first:`

## Blocked By
