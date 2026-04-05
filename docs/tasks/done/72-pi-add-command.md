# `pi add` Command

## Type
feature

## Status
done

## Priority
medium

## Project
13-external-packages

## Description
Implement `pi add` as the ergonomic entry point for declaring a new package dependency. It writes to `pi.yaml` and fetches immediately.

**Commands:**

```bash
# Add a GitHub package
pi add yotam180/pi-common@v1.2

# Add a local folder source (no alias)
pi add file:~/shared-automations

# Add a local folder source with alias
pi add file:~/my-automations --as mytools
```

**Behavior:**
1. Parse the reference
2. For GitHub refs: fetch into cache (or confirm already cached)
3. Add the entry to `pi.yaml packages:` — if `packages:` doesn't exist yet, create it
4. Print confirmation:
```
  ✓  added yotam180/pi-common@v1.2 to pi.yaml
```

`pi add` is idempotent — adding a package that's already declared prints a message and exits cleanly without duplicating the entry.

Adding a GitHub ref without a version (`pi add yotam180/pi-common`) is an error: "version required — use pi add yotam180/pi-common@<tag>".

## Acceptance Criteria
- [x] `pi add org/repo@version` fetches and adds to `pi.yaml`
- [x] `pi add file:~/path` adds a file source to `pi.yaml`
- [x] `pi add file:~/path --as alias` adds aliased file source
- [x] `pi.yaml` is updated correctly; `packages:` created if absent
- [x] Idempotent: re-adding same package prints message, no duplicate entry
- [x] `pi add org/repo` (no version) is a clear error
- [x] `pi add` without arguments prints usage
- [x] Tests cover: GitHub add, file add with alias, idempotent re-add, missing version error

## Implementation Notes

### Architecture
- `internal/cli/add.go`: Cobra command with `--as` flag. `runAdd()` validates via `refparser.Parse()`, fetches GitHub packages into cache, then calls `config.AddPackage()`.
- `internal/config/writer.go`: `AddPackage()` reads pi.yaml, checks duplicates (returns `*DuplicatePackageError`), uses `insertPackageEntry()` for line-based string manipulation to append entries. `formatPackageEntry()` handles simple vs aliased YAML syntax.

### Key decisions
- **Line-based YAML editing** instead of unmarshal-marshal: preserves comments, formatting, and field order in the user's pi.yaml.
- **Duplicate detection by source string**: exact match on `PackageEntry.Source`. A re-add with a different alias would be treated as a new entry.
- **Re-validation after write**: `config.Load()` is called after writing to catch corruption.
- **GitHub fetch before write**: ensures the package actually exists before modifying pi.yaml.
- **Idempotent duplicates exit 0**: duplicate is not an error — it's a success with a message.

### Test coverage
- `internal/config/writer_test.go`: 14 unit tests
- `internal/cli/add_test.go`: 8 unit tests
- `tests/integration/add_test.go`: 8 integration tests

## Subtasks
- [x] Implement `pi add` subcommand
- [x] Implement `pi.yaml` writer (add to `packages:` preserving existing content)
- [x] Add tests

## Blocked By
69-github-package-cache, 70-packages-declaration-in-pi-yaml
