# `first:` Step Block

## Type
feature

## Status
todo

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
- [ ] `first:` block is valid anywhere a step appears (`steps:`, `install.run:`, `install.test:`)
- [ ] Exactly the first sub-step whose `if:` condition passes is executed; all others skipped
- [ ] A sub-step with no `if:` always matches and acts as a fallback
- [ ] If no sub-step matches (all have `if:` and all fail), the block is silently skipped
- [ ] `first:` accepts an optional `description:` for `pi info` output
- [ ] Sub-steps inside `first:` support all normal step fields: `env:`, `dir:`, `timeout:`, `silent:`, `pipe:`
- [ ] `pipe:` on the last sub-step of a `first:` block correctly pipes output to the next step in the parent list
- [ ] `pi info` renders `first:` blocks sensibly (shows description or sub-step summary)
- [ ] Parse error if `first:` contains zero sub-steps
- [ ] Tests cover: first matches, middle matches, fallback matches, nothing matches, `pipe:` through `first:`

## Implementation Notes

## Subtasks
- [ ] Define `FirstBlock` type in the automation schema
- [ ] Update parser to recognize `first:` as a step-level key
- [ ] Implement executor logic: evaluate `if:` conditions in order, run first match
- [ ] Handle `pipe:` correctly when `first:` is a piped step
- [ ] Add `pi info` rendering
- [ ] Add tests
- [ ] Add example automation demonstrating `first:` (e.g. update an installer example)

## Blocked By
