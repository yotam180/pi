# Deduplicate config/writer.go YAML Block Manipulation

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

`internal/config/writer.go` contains significant duplication between the packages-writer and setup-writer code. Both subsystems perform the same core operations:

1. **Find a named YAML block** (`setup:` or `packages:`) in the raw file content
2. **Append a new entry** to an existing block, or create the block if it doesn't exist
3. **Walk back over trailing blank lines** within the block before inserting

The duplicated functions:
- `appendNewPackagesBlock` / `appendNewSetupBlock` — identical pattern, only the block name differs
- `appendToExistingPackagesBlock` / `appendToExistingSetupBlock` — nearly identical logic
- Block-finding loops (scanning for `packages:` or `setup:` line) — repeated inline in multiple functions

This duplication makes it harder to add new programmatic YAML mutations (e.g., if we need to modify `shortcuts:` or add a new top-level block in the future). Extracting shared helpers reduces the surface area for bugs and makes the writer easily extensible.

**Proposed approach:**
- Extract `findBlockIndex(lines, blockName)` — returns the line index of a named YAML block
- Extract `appendToBlock(lines, blockIdx, entryYAML)` — appends an entry to an existing YAML list block
- Extract `appendNewBlock(content, blockName, entryYAML)` — creates a new block at the end of the file
- Refactor both packages and setup writers to use these shared helpers
- The `replaceSetupEntry` function (setup-specific, handles in-place replacement) stays as-is since it has no packages counterpart

## Acceptance Criteria
- [x] Shared helpers extracted: `findBlockIndex`, `appendToBlock`, `appendNewBlock`
- [x] `insertPackageEntry` and `insertSetupEntry` use the shared helpers
- [x] All existing tests pass unchanged (`go test ./internal/config/...`)
- [x] No change in behavior — pure refactoring
- [x] `go build ./...` and `go test ./...` pass
- [x] `replaceSetupEntry` remains setup-specific (no forced generalization)

## Implementation Notes

Extracted four shared helpers from duplicated code:
- `findBlockIndex(lines, blockName)` — finds a top-level YAML block by name
- `appendToBlock(lines, blockIdx, entryYAML)` — appends an entry to an existing block
- `appendNewBlock(content, blockName, entryYAML)` — creates a new block at EOF
- `insertIntoBlock(content, blockName, entryYAML)` — orchestrator: finds block or creates it

Both `insertPackageEntry` and `insertSetupEntry` are now one-line delegations to `insertIntoBlock`. `replaceSetupEntry` was also updated to use `findBlockIndex` instead of its inline block search loop.

9 new unit tests added for the shared helpers. All 48 writer tests pass (39 existing + 9 new). Full project test suite (21 packages) passes with zero failures.

The refactoring is purely mechanical — no behavior changes. Any future programmatic YAML block mutations (e.g., shortcuts:) can now be added with a single `insertIntoBlock` call.

## Subtasks
- [x] Extract shared YAML block helpers
- [x] Refactor `insertPackageEntry` to use helpers
- [x] Refactor `insertSetupEntry` to use helpers
- [x] Remove now-unused specialized functions
- [x] Verify all tests pass

## Blocked By
