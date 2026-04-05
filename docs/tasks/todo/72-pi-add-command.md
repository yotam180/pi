# `pi add` Command

## Type
feature

## Status
todo

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
- [ ] `pi add org/repo@version` fetches and adds to `pi.yaml`
- [ ] `pi add file:~/path` adds a file source to `pi.yaml`
- [ ] `pi add file:~/path --as alias` adds aliased file source
- [ ] `pi.yaml` is updated correctly; `packages:` created if absent
- [ ] Idempotent: re-adding same package prints message, no duplicate entry
- [ ] `pi add org/repo` (no version) is a clear error
- [ ] `pi add` without arguments prints usage
- [ ] Tests cover: GitHub add, file add with alias, idempotent re-add, missing version error

## Implementation Notes

## Subtasks
- [ ] Implement `pi add` subcommand
- [ ] Implement `pi.yaml` writer (add to `packages:` preserving existing content)
- [ ] Add tests

## Blocked By
69-github-package-cache, 70-packages-declaration-in-pi-yaml
