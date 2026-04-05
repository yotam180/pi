---
title: Using Packages
description: Add and use external automation packages in your project
---

This guide walks through adding a package to your project, running its automations, using aliases, and understanding how packages interact with local automations.

## What you'll learn

- How to add a package with `pi add`
- How to reference package automations in `pi run` and `run:` steps
- How aliases work and when to use them
- How packages are fetched during `pi setup`
- How to browse available automations
- How collision resolution works
- How to upgrade package versions

---

## Adding a package

Use `pi add` to declare a package dependency:

```bash
pi add yotam180/pi-common@v1.2
```

This does three things:

1. Validates the reference format
2. Fetches the package into `~/.pi/cache/github/yotam180/pi-common/v1.2/`
3. Appends the entry to `pi.yaml`:

```yaml
packages:
  - yotam180/pi-common@v1.2
```

Adding the same source twice is a no-op — PI prints "already in pi.yaml" and exits successfully.

:::note
GitHub sources require a version tag after `@`. Without it, `pi add` rejects the input: `version required — use pi add org/repo@<tag>`.
:::

## Adding a local folder source

For developing automations locally without a push/clone cycle, use a `file:` source:

```bash
# Absolute path
pi add file:~/my-automations

# With an alias
pi add file:~/my-automations --as mytools

# Relative to project root
pi add file:./packages/shared --as shared
```

`file:` sources are read directly from disk — no caching, no fetching. Changes are reflected immediately.

## Running package automations

After adding a package, its automations are available just like local ones.

### By plain name

If no local automation has the same name, you can use the plain name:

```bash
pi run docker/up
```

```yaml
# In an automation file
steps:
  - run: docker/up
```

### By alias

When you've declared an alias, use it as a prefix:

```yaml
packages:
  - source: yotam180/pi-common@v1.2
    as: common
```

```bash
pi run common/docker/up
```

```yaml
steps:
  - run: common/docker/up
```

Aliases make references unambiguous — even if multiple packages provide a `docker/up`, the alias tells PI exactly which one you mean.

### By full GitHub reference

You can always use the full reference, even without declaring the package:

```bash
pi run yotam180/pi-common@v1.2/docker/up
```

This works even if the package isn't in `pi.yaml`. See [on-demand fetching](#on-demand-references) below.

## Using aliases

The `as:` key provides a short name for a package:

```yaml
packages:
  - source: your-org/pi-docker-utils@v2.0
    as: docker-utils

  - source: your-org/pi-ci-helpers@v1.0
    as: ci
```

Now you can write:

```yaml
steps:
  - run: docker-utils/compose/up
  - run: ci/lint
  - run: ci/test
```

Aliases must be unique within a `pi.yaml` and must not contain `/`. An alias that collides with a local `.pi/` path prints a warning — the local automation wins.

## What happens during `pi setup`

Before running any setup automations, `pi setup` fetches all declared GitHub packages that aren't already cached:

```
  ↓  yotam180/pi-common@v1.2          fetching...
  ✓  yotam180/pi-common@v1.2          cached
  ✓  file:~/my-automations            found  (alias: mytools)
  ⚠  file:~/missing-path              not found
```

- **GitHub packages:** Fetched if not in cache. Once cached, subsequent runs are instant.
- **File sources:** Verified to exist on disk. A missing path prints a warning but doesn't halt setup.

After package fetching, setup entries run sequentially. You can reference package automations in your setup sequence:

```yaml
setup:
  - common/setup/install-deps
  - common/setup/configure-env
```

## Browsing available automations

Use `pi list` to see all available automations with their source:

```bash
pi list
```

```
NAME                    SOURCE                          DESCRIPTION
docker/up               [workspace]                     Start Docker containers
docker/down             yotam180/pi-common@v1.2         Stop Docker containers
setup/install-deps      yotam180/pi-common@v1.2         Install project dependencies
```

The SOURCE column shows where each automation comes from: `[workspace]` for local, the package source for packages.

Use `--all` to see automations from all packages, grouped by source:

```bash
pi list --all
```

Use `--builtins` to also include built-in `pi:*` automations:

```bash
pi list --builtins
```

## Collision behavior

When a local automation has the same name as a package automation, **local always wins**:

```
.pi/docker/up.yaml          ← your local automation
yotam180/pi-common@v1.2     ← package also has docker/up
```

Running `pi run docker/up` uses the local version. PI prints a warning to let you know the package automation is shadowed.

The resolution order is:

1. **Local** — `.pi/<name>.yaml`
2. **Package** — from declared `packages:` sources
3. **Built-in** — PI's built-in `pi:*` automations
4. **On-demand** — undeclared GitHub references

## On-demand references

You can reference a GitHub package inline without declaring it in `packages:`:

```yaml
steps:
  - run: org/repo@v1.0/docker/up
```

PI fetches it automatically (once, then cached) and prints an advisory:

```
  ↓  org/repo@v1.0          fetched (on demand)

  tip: add to pi.yaml to avoid fetching on every fresh clone:

    packages:
      - org/repo@v1.0
```

The advisory only appears when a live network fetch happens — not when the package is already cached.

:::tip
On-demand references are convenient for trying out a package. For production use, always declare packages in `pi.yaml` so they're fetched reliably during `pi setup`.
:::

`file:` sources are never fetched on demand — they must be declared in `pi.yaml`.

## Upgrading package versions

To upgrade a package, change the version tag in `pi.yaml`:

```yaml
packages:
  - yotam180/pi-common@v1.3    # was @v1.2
```

Or run `pi add` with the new version:

```bash
pi add yotam180/pi-common@v1.3
```

The next `pi run` or `pi setup` fetches the new version into cache. The old version stays in cache but is no longer used.

## Summary

- Use `pi add` to declare package dependencies — it validates, fetches, and writes to `pi.yaml`
- Reference package automations by plain name, alias, or full GitHub reference
- Use aliases (`as:`) for unambiguous references when multiple packages are involved
- `pi setup` fetches all declared packages before running setup automations
- Local automations always take priority over package automations
- On-demand fetching works for quick experiments — declare in `pi.yaml` for production use
