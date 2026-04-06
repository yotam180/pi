# Derive installHints from Tool Registry

## Type
improvement

## Status
done

## Priority
medium

## Project
18-tool-registry-centralization

## Description

`installHints` in `reqcheck/reqcheck.go` is a manually-maintained map of ~15 entries for `pi doctor` output. It overlaps with the `ToolDescriptor.InstallHint` field from the tool registry, but is kept separately and drifts independently. This task migrates `InstallHintFor` to derive hints from `tools.Registry`, and removes the standalone `installHints` map.

## Acceptance Criteria
- [x] `installHints` map in `reqcheck` is removed
- [x] `InstallHintFor` derives its result from `tools.Registry` (and/or a command hints supplement)
- [x] All hints that existed before are still present and return the same strings
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

Completed as part of the `tool-descriptor-type` task. The `installHints` map was removed from `reqcheck/reqcheck.go` and `InstallHintFor()` now delegates to `tools.InstallHintFor()`. All non-builtin commands (docker, jq, git, curl, wget, rustc, cargo, rustup, make, mise) were added to `tools.Registry` as command-only entries (no `BuiltinName`). Tests verify all existing hints are preserved.

## Subtasks
- [x] Add entries for non-builtin commands (docker, jq, git, curl, mise, etc.) to `tools.Registry`
- [x] Update `InstallHintFor` to consult registry
- [x] Delete `installHints` map
- [x] Verify pi doctor test coverage for hints

## Blocked By
- ~~`tool-descriptor-type`~~ (completed)
