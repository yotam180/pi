# Automation Discovery

## Type
feature

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Implement the `.pi/` folder scanner that discovers all automations in a project. Given a directory, find every automation — whether defined as `.pi/docker/up.yaml` or `.pi/docker/up/automation.yaml` — and produce a resolved map of name → automation. Also implement the lookup function used by `pi run`.

## Acceptance Criteria
- [x] `Discover(piDir string)` returns a map of automation name → loaded automation for all automations in the `.pi/` folder
- [x] Name resolution: `.pi/docker/up.yaml` → `"docker/up"`, `.pi/setup/cursor/automation.yaml` → `"setup/cursor"`
- [x] Lookup function: `Find(name string)` returns the automation or a clear "not found" error listing available names
- [x] Both resolution forms are handled (flat `.yaml` and directory `automation.yaml`)
- [x] Names are normalized: no leading/trailing slashes, lowercase
- [x] Unit tests: directory with mixed flat and directory automations, name collision detection (two files resolving to same name = error), empty `.pi/` dir

## Implementation Notes

### Decisions
- **Package**: `internal/discovery` — clean separation from `automation` (which handles parsing a single file) and `config` (which handles `pi.yaml`).
- **Result struct**: Instead of returning a bare `map[string]*automation.Automation`, wrapped in a `Result` struct that co-locates the sorted names list and the `Find()` method. Keeps the API clean and avoids requiring callers to sort names themselves.
- **`Discover()` returns empty result for missing dir**: If `.pi/` doesn't exist, that's not an error — the project just has no automations. This matches how `git status` gracefully handles missing directories.
- **Name derivation**: Two-path logic — `automation.yaml` files use their parent directory as the name; all other `.yaml` files strip the extension. `automation.yaml` at the root of `.pi/` is explicitly skipped (would resolve to `.` which isn't meaningful).
- **Normalization**: Lowercase + trim slashes + forward-slash separators. Applied both at discovery time and lookup time so names are always consistent.
- **Collision detection**: If two paths resolve to the same name (e.g., `docker/up.yaml` and `docker/up/automation.yaml`), it's a hard error with both paths mentioned for easy debugging.
- **Find() error messages**: Two variants — empty ("no automations discovered") and non-empty (lists all available with descriptions in formatted columns). Makes CLI output helpful.

### File structure
```
internal/discovery/discovery.go      — Discover(), Result, Find(), deriveName(), normalizeName()
internal/discovery/discovery_test.go — 18 tests
```

### Test coverage (18 tests)
- **Discover**: empty dir, non-existent dir, flat YAML, directory automation, mixed formats, name collision, skips non-YAML, name normalization (case), deeply nested, invalid YAML, root-level automation, automation.yaml at root (skipped)
- **Find**: existing automation, case-insensitive lookup, not found (with available list), not found (empty), trims slashes
- **Names**: returns copy (mutation safety)

## Subtasks
- [x] Implement recursive `.pi/` walker
- [x] Implement name derivation logic
- [x] Implement `Find` with helpful error output
- [x] Unit tests with temp directory fixtures (18 tests)

## Blocked By
02-config-and-automation-schema (done)
