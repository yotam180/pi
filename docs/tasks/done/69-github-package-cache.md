# GitHub Package Cache

## Type
feature

## Status
done

## Priority
high

## Project
13-external-packages

## Description
Implement the package cache manager that fetches GitHub repos at specific versions and stores them in `~/.pi/cache/`. Once a version is cached it is immutable — PI never re-fetches it unless explicitly asked.

**Cache layout:**
```
~/.pi/cache/
  github/
    yotam180/
      pi-common/
        v1.2/
          .pi/
            docker/up.yaml
            ...
          pi-package.yaml   ← optional
        v2.0/
          .pi/
            ...
```

**Fetch behavior:**
1. Check `~/.pi/cache/github/org/repo/version/` — if exists, return path immediately (cache hit)
2. On cache miss: clone `https://github.com/org/repo` at the given tag/ref into a temp dir, then move atomically to the cache path
3. Mutable refs (`@main`, `@HEAD`, `@<branch>`) get a special subfolder `main~<date>` and emit a warning: "Using mutable ref @main — result may not be reproducible. Pin to a version tag for stability."
4. Network failures produce a clear error: "Could not fetch org/repo@version. Check network and that the repo/tag exists."

**Private repos:**
- Attempt clone via SSH first (`git@github.com:org/repo.git`) if an SSH key is configured
- Fall back to HTTPS with `GITHUB_TOKEN` env var if set (`https://<token>@github.com/org/repo.git`)
- If both fail and the repo is private, print actionable auth instructions

**`pi-package.yaml` handling (optional):**
After fetching, check if `pi-package.yaml` exists in the repo root. If it does, parse it and validate `min_pi_version` against the running PI binary. If the PI version is too old, fail with a clear message. If `pi-package.yaml` is absent, proceed silently.

## Acceptance Criteria
- [x] Cache miss: repo is cloned at the specified tag and stored in the correct path
- [x] Cache hit: no network call made, cached path returned immediately
- [x] Mutable refs (`@main`) emit a reproducibility warning and are stored correctly
- [x] Invalid tag/ref produces a clear error
- [x] Private repo: SSH clone works when SSH key is configured
- [x] Private repo: HTTPS with `GITHUB_TOKEN` works as fallback
- [x] Private repo with no auth: prints actionable instructions
- [x] `pi-package.yaml` present with satisfied `min_pi_version`: proceeds normally
- [x] `pi-package.yaml` present with unsatisfied `min_pi_version`: fails with clear message
- [x] `pi-package.yaml` absent: proceeds normally with no error
- [x] Atomic write: a failed fetch does not leave a partial cache entry
- [x] `pi cache clean` (or equivalent) can purge the cache — at minimum, removing `~/.pi/cache/` works

## Implementation Notes

### Package: `internal/cache`

Two files:
- `cache.go` — `Cache` struct with `Fetch()`, `PackagePath()`, `IsMutableRef()`, clone logic with SSH/token/HTTPS fallback chain
- `package_yaml.go` — `PackageYAML` struct, `checkPackageYAML()`, `versionSatisfies()` for min_pi_version checks

### Key design decisions:
1. **Dependency injection for testability**: `GitFunc` and `GetenvFunc` fields on `Cache` allow full mocking without real git or env vars
2. **Atomic writes**: clone goes into `os.MkdirTemp()` temp dir, then `os.Rename()` to final path. If anything fails, `defer os.RemoveAll(tmpDir)` cleans up
3. **Auth fallback order**: SSH → HTTPS+token → plain HTTPS. Each attempt gets a fresh temp dir
4. **Mutable refs**: date-stamped subfolder (e.g. `main~20260405`) to avoid stale caches while maintaining immutability per fetch
5. **.git removal**: after checkout, `.git` is removed — cached packages are working trees only
6. **Version checking**: dev builds and empty PIVersion skip min_pi_version checks; `versionSatisfies()` does component-wise comparison with v-prefix stripping
7. **Cache cleanup on failed version check**: if pi-package.yaml check fails after fetch, the cache entry is removed

### Test coverage: 32 tests
- `cache_test.go`: 17 tests — cache hit, cache miss (SSH/token/HTTPS), auth fallback order, mutable ref warning, atomic write, second-call cache hit, repo files, private repo error messages
- `package_yaml_test.go`: 15 tests — absent/empty/satisfied/unsatisfied/exact/dev/invalid YAML/v-prefix versions, integration with Fetch (satisfied/unsatisfied/absent/cache-hit checked)

## Subtasks
- [x] Define cache directory structure and path helpers
- [x] Implement clone-and-cache logic
- [x] Handle mutable refs with warning
- [x] Implement SSH + HTTPS auth fallback chain
- [x] Implement optional `pi-package.yaml` parsing and version check
- [x] Write tests (mock git operations where appropriate)

## Blocked By
68-automation-reference-parser ✅ (done)
