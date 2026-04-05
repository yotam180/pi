---
title: Packages
description: Share and consume automations across projects via GitHub repos or local folders
---

Packages let teams share automations between repositories. A package is a GitHub repo (or local folder) with a `.pi/` directory — the same structure as any PI project. No registry, no special tooling.

## In this section

- [Why packages](#why-packages) — the problem they solve
- [Declaring packages](#declaring-packages) — the `packages:` block in `pi.yaml`
- [Source types](#source-types) — GitHub and file sources
- [Aliases](#aliases) — naming shortcuts for packages
- [`pi add`](#pi-add) — the ergonomic way to add packages
- [Resolution order](#resolution-order) — what happens when names collide
- [On-demand fetching](#on-demand-fetching) — inline GitHub references without declaration
- [Version pinning](#version-pinning) — why `@version` matters
- [Running package automations](#running-package-automations) — `pi list`, `pi run`, and the SOURCE column
- [Writing a package](#writing-a-package) — the author side
- [`pi-package.yaml`](#pi-packageyaml) — optional package metadata
- [Private repositories](#private-repositories) — SSH and token authentication

---

## Why packages

Most teams start with automations in `.pi/` that work for one repo. Eventually patterns emerge — the Docker setup, the CI helpers, the git hook installer — that belong in every repo. Copy-pasting doesn't scale: a fix in one repo doesn't propagate to the others.

Packages solve this. Maintain one repo of shared automations, and every project consumes it with a version pin. Changes propagate when you upgrade the pin.

## Declaring packages

Add a `packages:` block to `pi.yaml`:

```yaml
packages:
  # Simple GitHub package
  - yotam180/pi-common@v1.2

  # GitHub package with alias
  - source: yotam180/pi-common@v1.2
    as: common

  # Local folder (absolute)
  - source: file:~/my-automations
    as: mytools

  # Local folder (relative to project root)
  - source: file:./packages/shared
    as: shared
```

Package automations are discovered and merged into the project — they work just like local automations in `pi run`, `pi list`, and `pi info`.

## Source types

| Source | Format | Caching |
|--------|--------|---------|
| GitHub | `org/repo@version` | Cached in `~/.pi/cache/`; fetched once per version |
| File | `file:~/path` or `file:./relative` | Read directly from disk; no caching |

**GitHub sources** require a version tag after `@`. The package is cloned into `~/.pi/cache/github/<org>/<repo>/<version>/` on first use and reused from cache on subsequent runs.

**File sources** point to a local directory containing a `.pi/` folder. `~` expands to the home directory. `./` is relative to the project root. Changes to file sources are reflected immediately — no cache invalidation needed.

## Aliases

The `as:` key provides a short name for a package, letting you write cleaner references:

```yaml
packages:
  - source: yotam180/pi-common@v1.2
    as: common
```

```yaml
# In an automation
steps:
  - run: common/docker/up       # instead of the full path
```

Aliases must be unique within a `pi.yaml` and must not contain `/`. An alias that collides with a local `.pi/` automation path prints a warning (local wins).

## `pi add`

The command-line way to declare a package dependency:

```bash
# GitHub package
pi add yotam180/pi-common@v1.2

# Local folder
pi add file:~/shared-automations

# Local folder with alias
pi add file:~/my-automations --as mytools
```

`pi add` validates the source, fetches GitHub packages into `~/.pi/cache/`, and appends the entry to `pi.yaml`. Adding the same source twice is a no-op — PI prints "already in pi.yaml" and exits successfully.

GitHub sources without `@version` are rejected: `version required — use pi add org/repo@<tag>`.

## Resolution order

When PI encounters an automation name, it resolves in this order:

1. **Local** — `.pi/<name>.yaml` or `.pi/<name>/automation.yaml`
2. **Package** — automations from declared `packages:` sources
3. **Built-in** — automations shipped with the PI binary (prefixed `pi:`)
4. **On-demand** — undeclared GitHub references fetched automatically

**Local always wins.** If a local automation shadows a package automation, a warning is printed.

## On-demand fetching

You can reference a GitHub package inline in a `run:` step without declaring it in `packages:`:

```yaml
steps:
  - run: org/repo@v1.0/docker/up
```

PI fetches it automatically (once, then cached) and prints an advisory to stderr:

```
  ↓  org/repo@v1.0          fetched (on demand)

  tip: add to pi.yaml to avoid fetching on every fresh clone:

    packages:
      - org/repo@v1.0
```

The advisory only appears when a live network fetch happens, not when the package is already cached.

`file:` sources are never fetched on demand — they must be declared in `pi.yaml`.

## Version pinning

The `@version` part is required for GitHub packages. Use release tags for stable, reproducible builds:

```yaml
packages:
  - yotam180/pi-common@v1.2    # pinned — same content everywhere
```

Mutable refs like `@main` or `@HEAD` are accepted but emit a reproducibility warning and use a date-stamped cache key (e.g., `main~20260405`):

```yaml
packages:
  - yotam180/pi-common@main    # ⚠ mutable — not reproducible
```

`@main` today may not equal `@main` in six months. Pin to a tag for stability.

## Running package automations

Package automations appear in `pi list` with a SOURCE column showing where they come from:

```bash
pi list             # local + packages, with SOURCE column
pi list --all       # includes grouped package sections
```

Run them like any other automation:

```bash
# By plain name (if no local collision)
pi run docker/up

# By alias
pi run common/docker/up

# By full GitHub reference
pi run org/repo@v1.0/docker/up
```

## Writing a package

A PI package is a GitHub repo with automations in `.pi/`:

```
my-pi-package/
  .pi/
    docker/
      up.yaml
      down.yaml
    setup/
      install-deps.yaml
  pi-package.yaml          ← optional
```

Steps to publish:

```bash
git init
# ... create .pi/ automations ...
git add .
git commit -m "initial package release"
git tag v1.0
git push origin main
git push origin v1.0
```

Users consume it with:

```bash
pi add your-org/my-pi-package@v1.0
```

**Tips for package authors:**
- Write small, single-purpose automations — let callers compose with `run:`
- Use `inputs:` for parameterization
- Write idempotent automations (check before acting)
- Prefix the repo name with `pi-` so it's discoverable on GitHub

For a complete walkthrough, see the [Publishing to GitHub guide](/guides/publishing-to-github/).

## `pi-package.yaml`

An optional file at the root of a package repo. Its only field is `min_pi_version`:

```yaml
min_pi_version: "0.5.0"
```

When present, PI checks the running version at fetch time. If PI is older than `min_pi_version`, the fetch fails with a clear upgrade message. Dev builds skip this check.

If `pi-package.yaml` is absent or empty, the package works with any PI version.

See the [pi-package.yaml reference](/reference/pi-package-yaml/) for details.

## Private repositories

PI tries multiple auth methods when fetching a GitHub package:

1. **SSH** — `git@github.com:org/repo.git` (works if you have an SSH key configured)
2. **HTTPS with token** — uses the `GITHUB_TOKEN` environment variable
3. **Plain HTTPS** — public repos only

For private repos, ensure either an SSH key is configured or `GITHUB_TOKEN` is set. For detailed setup instructions, see the [Private Repositories guide](/guides/private-repos/).
