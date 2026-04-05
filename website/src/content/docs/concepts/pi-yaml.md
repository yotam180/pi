---
title: pi.yaml
description: The root configuration file — shortcuts, setup sequences, and package declarations
---

`pi.yaml` is the root configuration file for a PI project. It lives at the repo root alongside the `.pi/` folder. It is small by design — it only declares which automations are exposed as shell shortcuts, which run during `pi setup`, and where to find external automations. It does not define automations itself.

## In this section

- [`project:`](#project) — the project name
- [`shortcuts:`](#shortcuts) — expose automations as shell commands
- [`setup:`](#setup) — what `pi setup` runs
- [`packages:`](#packages) — external automation sources
- [`runtimes:`](#runtimes) — sandboxed runtime provisioning

---

## `project:`

The project name. Used in generated shell function names (e.g., `pi-setup-<project>`).

```yaml
project: my-app
```

## `shortcuts:`

Map shortcut names to automations. Each shortcut becomes a shell command after running `pi shell`.

```yaml
shortcuts:
  up:    docker/up
  down:  docker/down
  test:  test
```

The key is the command you type in the terminal. The value is the automation name (the path inside `.pi/`, minus the `.yaml` extension).

### Object form with `anywhere: true`

By default, a shortcut `cd`s to the project root before running. With `anywhere: true`, it runs from your current directory:

```yaml
shortcuts:
  deploy:
    run: deploy/push-image
    anywhere: true
```

Use `anywhere: true` for automations that operate on the current directory or work globally — like deploy commands that don't depend on being at the project root.

### Shortcuts with `with:`

Shortcuts can map positional shell arguments to named inputs:

```yaml
shortcuts:
  install-py:
    run: pi:install-python
    with:
      version: "$1"
```

This generates a shell function where `install-py 3.13` becomes `pi run pi:install-python --with version=3.13`.

## `setup:`

The list of automations that `pi setup` runs sequentially. Setup automations are expected to be idempotent — check first, act only if needed.

```yaml
setup:
  - pi:install-homebrew
  - run: pi:install-python
    with:
      version: "3.13"
  - run: pi:install-node
    with:
      version: "20"
  - setup/install-project-deps
  - run: setup/configure-git-hooks
    if: dir.exists(".git")
```

Entries can be **bare strings** or **objects**:

| Form | When to use |
|------|-------------|
| `- setup/install-go` | No conditions, no inputs |
| `- run: setup/install-go` | Object form — when you need `if:` or `with:` |

Both forms can be mixed in the same list. Bare strings are shorthand for `run: <string>` with no modifiers.

### Conditional setup entries

Object-form entries support `if:` to skip based on the runtime environment:

```yaml
setup:
  - run: setup/install-brew
    if: os.macos
  - run: setup/install-apt-deps
    if: os.linux
  - setup/install-project-deps    # always runs
```

When a condition evaluates to false, the entry is skipped with a message. The same predicate system used by step-level `if:` is available — see the [Conditions reference](/reference/conditions/).

## `packages:`

Declare external automation sources. Package automations are discovered and merged into the project — they work just like local automations.

```yaml
packages:
  - yotam180/pi-common@v1.2

  - source: file:~/my-automations
    as: mytools

  - source: file:./packages/shared
    as: shared
```

### Source types

| Source | Format | Caching |
|--------|--------|---------|
| GitHub | `org/repo@version` | Cached in `~/.pi/cache/`; fetched once per version |
| File | `file:~/path` or `file:./relative` | Read directly from disk; no caching |

GitHub packages require a version tag after `@`. Mutable refs like `@main` work but emit a reproducibility warning.

### Aliases (`as:`)

The `as:` key lets you write `run: mytools/docker/up` instead of referencing the full source. Aliases must be unique and must not contain `/`.

### `pi add`

The ergonomic way to add a package:

```bash
pi add yotam180/pi-common@v1.2
pi add file:~/shared-automations
pi add file:~/my-automations --as mytools
```

For the full package story — resolution order, on-demand fetching, writing packages — see [Packages](/concepts/packages/).

## `runtimes:`

Configure sandboxed runtime provisioning. When enabled, PI can automatically install required runtimes (Python, Node.js) into `~/.pi/runtimes/` instead of failing on missing requirements.

```yaml
runtimes:
  provision: auto    # never | ask | auto
  manager: mise      # mise | direct
```

| Field | Values | Default | Description |
|-------|--------|---------|-------------|
| `provision` | `never`, `ask`, `auto` | `never` | Whether to provision missing runtimes |
| `manager` | `mise`, `direct` | `mise` | Which tool to use for provisioning |

With `provision: never` (the default), PI behaves as before — it checks for requirements but never installs them. With `auto`, missing runtimes are installed automatically. With `ask`, PI prompts before installing.

Provisioned runtimes are installed into `~/.pi/runtimes/<name>/<version>/bin/` and their PATH is scoped to the step execution — they don't affect your system installation.
