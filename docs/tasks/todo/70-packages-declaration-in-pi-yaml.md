# `packages:` Declaration in `pi.yaml`

## Type
feature

## Status
todo

## Priority
high

## Project
13-external-packages

## Description
Add a `packages:` block to `pi.yaml` where teams declare their external automation dependencies. `pi setup` fetches all declared packages before running setup automations.

**Formats supported in `packages:`:**

```yaml
packages:
  # Simple GitHub package — most common case
  - yotam180/pi-common@v1.2

  # GitHub package with explicit source key (same as above, verbose form)
  - source: yotam180/pi-common@v1.2

  # Local folder source — for developing automations without push/clone cycle
  - source: file:~/my-automations
    as: mytools           # optional alias; enables `run: mytools/docker/up`

  # Local folder without alias — reference with full file: prefix
  - source: file:~/shared-automations
```

The `as:` alias lets you write `run: mytools/docker/up` instead of `run: file:~/my-automations/docker/up`. Aliases must be unique within a `pi.yaml`. An alias that collides with a local `.pi/` automation path emits a warning (local wins per resolution order).

**`pi setup` integration:**
Before running any setup automations, `pi setup` fetches all GitHub packages that aren't already cached. `file:` entries are verified to exist on disk (print a warning if not, but don't fail). The user sees:

```
  ↓  fetching yotam180/pi-common@v1.2...
  ✓  yotam180/pi-common@v1.2  cached
  ✓  file:~/my-automations     found  (alias: mytools)
```

## Acceptance Criteria
- [ ] `packages:` block is parsed correctly in `pi.yaml`
- [ ] Both simple string form and object form (`source:` / `as:`) are supported in the same list
- [ ] `pi setup` fetches all GitHub packages before running setup automations
- [ ] `file:` entries are verified to exist; missing file source prints warning but doesn't halt setup
- [ ] `as:` aliases are registered and usable in `run:` steps across all automations in the project
- [ ] Duplicate alias names in `pi.yaml` are a parse error
- [ ] Alias collision with a local `.pi/` path emits a warning; local wins
- [ ] `pi.yaml` without `packages:` block continues to work as before
- [ ] Tests cover: simple GitHub entry, file entry with alias, file entry without alias, duplicate alias error, missing file source warning

## Implementation Notes

## Subtasks
- [ ] Update `pi.yaml` schema/parser for `packages:`
- [ ] Implement alias registry
- [ ] Integrate package fetching into `pi setup` pre-flight
- [ ] Add tests

## Blocked By
68-automation-reference-parser, 69-github-package-cache
