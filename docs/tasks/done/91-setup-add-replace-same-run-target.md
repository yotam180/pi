# setup add: replace existing entry with same run target

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
`pi setup add` appends a new entry even when an entry with the same `run` target already exists under a different `with` map. The current duplicate check in `isSetupDuplicate` (`internal/config/writer.go`) compares `run` + `if` + `with` together, so `pi:install-node` (bare) and `pi:install-node --version 22` are seen as two distinct entries and both get written.

Concrete reproduction:
```
pi setup add pi:install-node          # writes bare entry
pi setup add pi:install-node --version 22   # should REPLACE, instead APPENDS
```

Result in `pi.yaml`:
```yaml
setup:
  - pi:install-node
  - run: pi:install-node
    with:
      version: "22"
```

The correct behaviour: when `run` (and `if`) match an existing entry, **replace** the existing entry in-place rather than append. Only treat it as a true duplicate (no-op) when the full entry — including `with` — is identical.

## Acceptance Criteria
- [x] `pi setup add pi:install-node` followed by `pi setup add pi:install-node --version 22` results in exactly one entry in `pi.yaml` (the second one, with `version: "22"`).
- [x] A fully-identical repeat call (same `run`, same `with`, same `if`) is still a no-op with a clear "Already in pi.yaml" message.
- [x] Replacement preserves the original position in the setup list (does not move the entry to the end).
- [x] `config.AddSetupEntry` is updated; the `DuplicateSetupEntryError` path now only fires for exact duplicates.
- [x] Existing tests updated / new tests added covering the replace case.
- [x] `go build ./...` and `go test ./...` pass.

## Implementation Notes
Key files:
- `internal/config/writer.go` — `isSetupDuplicate` + `AddSetupEntry` logic needs a `ReplaceOrAdd` semantic.
- `internal/cli/setup_add.go` — surfaces the result message ("Replaced" vs "Added").
- `internal/config/writer_test.go` — add `TestAddSetupEntry_SameRunDifferentWith_Replaces`.

Change `isSetupDuplicate` to two separate checks:
1. **Same-run match** (run + if match, with differs) → replace in-place.
2. **Exact match** (run + if + with all match) → return `DuplicateSetupEntryError` (no-op).

For in-place replacement, find the line range occupied by the old entry (multi-line object form or single-line string form) and substitute it.

## Subtasks
- [x] Replaced `isSetupDuplicate` with `findMatchingEntry` returning `(index, exactMatch)`.
- [x] Added `replaceSetupEntry(content string, entryIdx int, newYAML string) (string, error)` — finds the Nth setup entry's line range and substitutes.
- [x] Wired replacement path into `AddSetupEntry`; returns `ReplacedSetupEntryError` for same-target replacements.
- [x] Updated `runSetupAdd` in `setup_add.go` to print "Replaced in pi.yaml" message.
- [x] Added 7 new tests (unit + integration) covering replace scenarios: bare→versioned, versioned→bare, position preservation, multi-line to multi-line, error message format.

## Blocked By
