# `packages:` Declaration in `pi.yaml`

## Type
feature

## Status
done

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
- [x] `packages:` block is parsed correctly in `pi.yaml`
- [x] Both simple string form and object form (`source:` / `as:`) are supported in the same list
- [x] `pi setup` fetches all GitHub packages before running setup automations
- [x] `file:` entries are verified to exist; missing file source prints warning but doesn't halt setup
- [x] `as:` aliases are registered and usable in `run:` steps across all automations in the project
- [x] Duplicate alias names in `pi.yaml` are a parse error
- [x] Alias collision with a local `.pi/` path emits a warning; local wins
- [x] `pi.yaml` without `packages:` block continues to work as before
- [x] Tests cover: simple GitHub entry, file entry with alias, file entry without alias, duplicate alias error, missing file source warning

## Implementation Notes

### Approach
- Added `PackageEntry` struct to `internal/config` with polymorphic YAML parsing (string or object form)
- Added `packages:` field to `ProjectConfig` with validation: empty source, duplicate aliases, slash in aliases
- `PackageEntry` has helper methods: `IsFileSource()`, `FilePath()`, `PackageAliases()`
- Extended `discovery.Result` with package tracking: `packageSet`, `packages`, `aliasMap`, `packageAuto` maps
- `MergePackage()` discovers automations from a package dir's `.pi/` and merges; local automations shadow package ones with a warning
- `FindWithAliases()` now auto-populates known aliases from the Result's own alias map when nil is passed
- Added `findAlias()` and `findInPackage()` resolution methods for alias and GitHub/file refs
- `discoverAllWithConfig()` in CLI layer handles the full pipeline: discover local → fetch/verify packages → merge packages → merge builtins
- `resolveFilePackage()` resolves relative paths against project root
- `resolveGitHubPackage()` uses the existing `cache.Cache` for fetch/verify
- `pi setup` shows "Fetching packages..." header and per-package status lines before running setup automations
- All CLI commands (run, list, info, validate, doctor) now use `discoverAllWithConfig` to see package automations
- `display.PackageFetch()` renders status lines with icons: ↓ fetching, ✓ cached/found, ✗ failed, ⚠ not found

### Design Decisions
- Relative `file:` paths are resolved relative to project root (not CWD) — consistent with how `dir:` on steps works
- Package automations that don't collide with local names are added to the main `Automations` map so they work with `Find()` for plain name lookups (e.g., `run: docker/up` works without needing `mytools/docker/up`)
- Alias resolution works via the `refparser` already: first path segment matching a known alias triggers `RefAlias` type
- The `file:` source with `as:` alias is the primary developer workflow — iterate locally without push/clone
- GitHub packages are fetched via the existing `cache.Cache` which handles SSH/token/HTTPS auth fallback
- Package merge order: local > file: sources > GitHub packages > builtins (implemented via call order in `discoverAllWithConfig`)

### Tests Added
- 10 config tests: simple string, object form, mixed, duplicate alias, empty source, slash in alias, backward compat, aliases helper, IsFileSource, FilePath
- 8 discovery tests: basic merge, local shadows, alias resolution, find local falls through, package source tracking, empty package dir, unknown alias
- 4 display tests: cached, with detail, fetching, failed
- 9 integration tests: list, run local, run package automation, run via alias, run utils, info, validate, setup fetches packages, local shadows package

## Subtasks
- [x] Update `pi.yaml` schema/parser for `packages:`
- [x] Implement alias registry
- [x] Integrate package fetching into `pi setup` pre-flight
- [x] Add tests

## Blocked By
68-automation-reference-parser, 69-github-package-cache
