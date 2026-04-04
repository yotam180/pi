# Automation Discovery

## Type
feature

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Implement the `.pi/` folder scanner that discovers all automations in a project. Given a directory, find every automation — whether defined as `.pi/docker/up.yaml` or `.pi/docker/up/automation.yaml` — and produce a resolved map of name → automation. Also implement the lookup function used by `pi run`.

## Acceptance Criteria
- [ ] `Discover(piDir string)` returns a map of automation name → loaded automation for all automations in the `.pi/` folder
- [ ] Name resolution: `.pi/docker/up.yaml` → `"docker/up"`, `.pi/setup/cursor/automation.yaml` → `"setup/cursor"`
- [ ] Lookup function: `Find(name string)` returns the automation or a clear "not found" error listing available names
- [ ] Both resolution forms are handled (flat `.yaml` and directory `automation.yaml`)
- [ ] Names are normalized: no leading/trailing slashes, lowercase
- [ ] Unit tests: directory with mixed flat and directory automations, name collision detection (two files resolving to same name = error), empty `.pi/` dir

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Implement recursive `.pi/` walker
- [ ] Implement name derivation logic
- [ ] Implement `Find` with helpful error output
- [ ] Unit tests with temp directory fixtures

## Blocked By
02-config-and-automation-schema
