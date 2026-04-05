# Rename `pipe_to: next` to `pipe: true`

## Type
improvement

## Status
todo

## Priority
medium

## Project
12-yaml-ergonomics

## Description
The current `pipe_to: next` syntax is misleading — it implies the field could take values other than `next`, but `next` is the only valid value. The intent is simply "pipe this step's stdout to the next step's stdin." The cleaner form is `pipe: true`.

Both forms should work. `pipe_to: next` should emit a deprecation warning at parse time pointing authors toward `pipe: true`. `pipe_to: <anything-other-than-next>` remains a parse error.

## Acceptance Criteria
- [ ] `pipe: true` is the canonical form and works identically to `pipe_to: next`
- [ ] `pipe_to: next` continues to work but emits a deprecation warning
- [ ] `pipe_to: <other-value>` remains a parse error
- [ ] `pi validate` flags `pipe_to: next` as deprecated style
- [ ] All example and built-in automations updated to use `pipe: true`
- [ ] Tests cover: `pipe: true`, `pipe_to: next` (with warning), invalid `pipe_to` value

## Implementation Notes

## Subtasks
- [ ] Update schema/parser to accept `pipe: true`
- [ ] Add deprecation warning for `pipe_to: next`
- [ ] Update `pi validate` to flag old style
- [ ] Update all example automations
- [ ] Add/update tests

## Blocked By
