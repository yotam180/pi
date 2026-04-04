# Polyglot Example Workspace

## Type
infra

## Status
todo

## Priority
medium

## Project
02-polyglot-runner-and-shell-integration

## Description
Create an `examples/polyglot/` workspace that demonstrates mixing bash, Python, and TypeScript steps with pipe support. This workspace serves as the integration test bed for Project 2 features and as documentation for users.

## Acceptance Criteria
- [ ] `examples/polyglot/` workspace exists with `pi.yaml`, `.pi/` folder, and `README.md`
- [ ] Contains at least 3 automations demonstrating:
  - A Python step (inline and file)
  - A TypeScript step (inline and file)
  - A multi-step automation with `pipe_to: next` piping bash output through Python
- [ ] All automations work end-to-end with `pi run`
- [ ] Integration tests added to `tests/integration/` covering these automations
- [ ] No references to Vyper or internal tooling

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `examples/polyglot/` workspace structure
- [ ] Write automations with Python, TypeScript, and pipe examples
- [ ] Add integration tests
- [ ] Write README

## Blocked By
09-pipe-support
