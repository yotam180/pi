# Drop Required `name:` Field from Automation Files

## Type
improvement

## Status
done

## Priority
high

## Project
12-yaml-ergonomics

## Description
The `name:` field at the top of every automation YAML is redundant — it always mirrors the file path. An automation at `.pi/build/default.yaml` always has `name: build/default`. PI already knows the file path when loading the automation, so it can derive the name without the author spelling it out.

Make `name:` optional. When absent, PI derives the name from the file path: strip the `.pi/` prefix, strip `.yaml` suffix, and collapse `automation.yaml` to its parent folder name (e.g. `.pi/setup/install-cursor-extensions/automation.yaml` → `setup/install-cursor-extensions`).

When `name:` is present and matches the derived path, accept silently. When present and *mismatches*, emit a parse-time warning (but don't fail). This lets existing files keep working while the field gradually gets dropped.

## Acceptance Criteria
- [x] `name:` is optional in all automation YAML files
- [x] When absent, PI derives the automation name from the file path using the rules above
- [x] When present and matching, no warning emitted
- [x] When present and mismatching the derived name, a warning is emitted at parse time
- [x] All existing automation files with `name:` still load correctly
- [x] `pi info`, `pi list`, `pi run` all use the derived name correctly when `name:` is absent
- [x] Tests cover: absent name, present-matching name, present-mismatching name
- [x] At least one example automation updated to omit `name:` (to serve as a canonical example)

## Implementation Notes

### Approach
The change touches three layers:

1. **`internal/automation/automation.go`**: Removed the `name == ""` validation from `validate()`. The `name:` field in YAML is now optional — when absent, `Name` is left as the zero value `""`. The caller (discovery or builtins) is responsible for setting it.

2. **`internal/discovery/discovery.go`**: Added `reconcileAutomationName()` called after `Load()`. When `Name` is empty, it's set to the path-derived name. When present but mismatching, a warning is printed to the `warnWriter` (stderr in production, `bytes.Buffer` or `nil` in tests). The `Discover()` function signature now takes an `io.Writer` for warnings.

3. **`internal/builtins/builtins.go`**: After `LoadFromBytes()`, if `Name` is empty, it's set from the derived path name. No mismatch warning for builtins (they're authored by us).

### Key decisions
- **Warning, not error**: Mismatching `name:` emits a warning but doesn't fail — backward compatible.
- **Name preserved on mismatch**: When `name:` is present (even mismatching), the declared value is kept as `a.Name`. The automation is keyed by path-derived name in the discovery map.
- **`warnWriter` parameter**: Rather than using global stderr, `Discover()` accepts an `io.Writer` for testability.

### Files changed
- `internal/automation/automation.go` — removed name validation
- `internal/automation/automation_test.go` — updated `TestLoad_MissingName` → `TestLoad_MissingName_Allowed`
- `internal/discovery/discovery.go` — `Discover()` takes `warnWriter`, added `reconcileAutomationName()`
- `internal/discovery/discovery_test.go` — 5 new tests, updated all `Discover()` calls
- `internal/builtins/builtins.go` — set name from path when absent
- `internal/cli/discover.go` — pass `os.Stderr` to `Discover()`
- `internal/builtins/embed_pi/hello.yaml` — removed `name:` (canonical example)
- `examples/basic/.pi/greet.yaml` — removed `name:` (canonical example)

### Test coverage
- `TestDiscover_NameAbsent_DerivedFromPath` — flat file, name derived
- `TestDiscover_NameAbsent_NestedPath` — nested file (docker/up), name derived
- `TestDiscover_NameAbsent_AutomationYAML` — automation.yaml form, name derived
- `TestDiscover_NamePresent_Matching_NoWarning` — matching name, no warning
- `TestDiscover_NamePresent_Mismatching_Warning` — mismatching name, warning emitted

## Subtasks
- [x] Update automation struct and parser to make `name` optional
- [x] Implement path-to-name derivation logic
- [x] Add mismatch warning
- [x] Update at least one example to omit `name:`
- [x] Add tests

## Blocked By
