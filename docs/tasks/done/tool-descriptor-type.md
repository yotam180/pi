# Create ToolDescriptor and Derive setupAddKnownTools from It

## Type
improvement

## Status
done

## Priority
medium

## Project
18-tool-registry-centralization

## Description

Create an `internal/tools` package with a `ToolDescriptor` type and a `Registry` slice. Migrate `setupAddKnownTools` in `cli/setup_add.go` to be generated from this registry. Fix the `pi:ruby` gap and any other missing or inconsistent entries discovered during migration.

## Acceptance Criteria
- [x] `internal/tools` package exists with `ToolDescriptor` and `Registry`
- [x] `setupAddKnownTools` in `cli/setup_add.go` is gone — replaced by `tools.BuildShortNameMap()`
- [x] All existing short-form and `pi:` prefix aliases are preserved (regression test)
- [x] `pi:ruby` gap fixed — `ruby` moved to command-only (no builtin exists yet)
- [x] A test asserts coverage: every `pi:install-*` builtin has an entry in `tools.Registry`
- [x] A reverse test asserts: every Registry entry with a BuiltinName references an actual builtin
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Architecture

Created `internal/tools/tools.go` with:
- `ToolDescriptor` struct: `BuiltinName`, `ShortNames []string`, `InstallHint string`
- `Registry` — authoritative slice of all tool descriptors
- `BuildShortNameMap()` — generates the short-name-to-builtin map (used by `cli/setup_add.go`)
- `InstallHintFor(name)` — lookup hint by tool name (used by `reqcheck/reqcheck.go`)
- `ToolResolutionHelp()` — generates help text for `pi setup add --help`

### Changes

**`internal/tools/tools.go`** — new package, single source of truth for tool metadata.

**`internal/cli/setup_add.go`**:
- Deleted the 45-entry `setupAddKnownTools` map variable
- Replaced with `setupAddKnownTools()` function that delegates to `tools.BuildShortNameMap()`
- Deleted `setupAddToolResolutionHelp()` function, replaced with `tools.ToolResolutionHelp()`
- Removed `sort` import (no longer needed)
- Added `tools` import

**`internal/cli/setup_add_test.go`**:
- Updated test calls from `setupAddToolResolutionHelp()` to `tools.ToolResolutionHelp()`
- Added `tools` import

**`internal/reqcheck/reqcheck.go`**:
- Deleted the 17-entry `installHints` map
- `InstallHintFor()` now delegates to `tools.InstallHintFor()`
- Added `tools` import

**`internal/tools/tools_test.go`** — comprehensive test suite:
- `TestBuildShortNameMap_AllEntriesPresent` — all short names resolve correctly
- `TestBuildShortNameMap_PiPrefixVariants` — all `pi:` prefix forms work
- `TestBuildShortNameMap_NoCommandOnlyEntries` — command-only tools excluded from short name map
- `TestInstallHintFor_KnownTools` — all tools with hints return them
- `TestToolResolutionHelp_*` — help text tests (mirrored from old CLI tests)
- `TestRegistryCoversAllInstallBuiltins` — every `pi:install-*` builtin has a Registry entry
- `TestRegistryBuiltinNamesExist` — every Registry entry with a BuiltinName references an actual builtin

### pi:ruby gap resolution

`ruby` was in `setupAddKnownTools` mapping to `pi:install-ruby`, but no `install-ruby.yaml` builtin exists in `internal/builtins/embed_pi/`. The `TestRegistryBuiltinNamesExist` test caught this. Fixed by removing `BuiltinName` from the ruby entry (making it command-only for install hints). A future task should create the `pi:install-ruby` builtin.

### Dependency direction

`cli → tools` and `reqcheck → tools` — no import cycles. `tools` imports only `builtins` (in tests only, via test file).

## Subtasks
- [x] Create `internal/tools/tools.go` with `ToolDescriptor` and `Registry`
- [x] Add `BuildShortNameMap()` function
- [x] Wire into `cli/setup_add.go`
- [x] Delete `setupAddKnownTools` map
- [x] Fix `pi:ruby` gap (moved to command-only)
- [x] Write coverage test
- [x] Run `go test ./...` and verify no regressions

## Blocked By
