# setup add help text should list all builtin tools

## Type
improvement

## Status
done

## Priority
low

## Project
standalone

## Description
The `pi setup add --help` text showed only python, node, and go in the tool name resolution examples, hardcoded. Rust and other tools were missing even though they resolve correctly.

### Fix Applied
The help text is now auto-generated from the `setupAddKnownTools` map via the `setupAddToolResolutionHelp()` function. This means the help text always stays in sync as new builtins are added.

The function:
- Filters out `pi:` prefixed aliases (shown only once per tool)
- Picks the canonical short name for each tool (the one matching the `pi:install-<name>` suffix pattern)
- Sorts alphabetically for consistent output

## Acceptance Criteria
- [x] `pi setup add --help` lists rust (and all other tools) in the resolution examples
- [x] The list is auto-generated from the known tools map
- [x] Test ensures the list contains all expected builtins
- [x] Test verifies deterministic output (no map-iteration randomness)
- [x] Test verifies canonical names are preferred over aliases

## Implementation Notes

### Changes
- `internal/cli/setup_add.go`: Replaced hardcoded help text with `setupAddToolResolutionHelp()` that generates from `setupAddKnownTools`
- `internal/cli/setup_add_test.go`: Added 3 tests: `TestSetupAddToolResolutionHelp_ContainsAllBuiltins`, `TestSetupAddToolResolutionHelp_Deterministic`, `TestSetupAddToolResolutionHelp_PrefersCanonicalName`

### Name selection logic
When multiple short-form names resolve to the same target:
1. Prefer the name matching the `pi:install-<name>` suffix (e.g., "homebrew" over "brew")
2. If neither or both match, prefer shorter name
3. If same length, prefer alphabetically earlier

## Subtasks
- [x] Auto-generate help text from setupAddKnownTools
- [x] Fix non-deterministic map iteration
- [x] Add tests

## Blocked By
