# GitHub Package Cache

## Type
feature

## Status
todo

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
- [ ] Cache miss: repo is cloned at the specified tag and stored in the correct path
- [ ] Cache hit: no network call made, cached path returned immediately
- [ ] Mutable refs (`@main`) emit a reproducibility warning and are stored correctly
- [ ] Invalid tag/ref produces a clear error
- [ ] Private repo: SSH clone works when SSH key is configured
- [ ] Private repo: HTTPS with `GITHUB_TOKEN` works as fallback
- [ ] Private repo with no auth: prints actionable instructions
- [ ] `pi-package.yaml` present with satisfied `min_pi_version`: proceeds normally
- [ ] `pi-package.yaml` present with unsatisfied `min_pi_version`: fails with clear message
- [ ] `pi-package.yaml` absent: proceeds normally with no error
- [ ] Atomic write: a failed fetch does not leave a partial cache entry
- [ ] `pi cache clean` (or equivalent) can purge the cache — at minimum, removing `~/.pi/cache/` works

## Implementation Notes

## Subtasks
- [ ] Define cache directory structure and path helpers
- [ ] Implement clone-and-cache logic
- [ ] Handle mutable refs with warning
- [ ] Implement SSH + HTTPS auth fallback chain
- [ ] Implement optional `pi-package.yaml` parsing and version check
- [ ] Write tests (mock git operations where appropriate)

## Blocked By
68-automation-reference-parser
