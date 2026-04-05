---
title: Setup Automations
description: Write a pi setup configuration that installs dependencies and configures your team's environment
---

This guide walks through building a `pi setup` configuration for a real project — one that installs tools, configures the environment, and leaves shell shortcuts ready in the terminal.

## What you'll learn

- How to declare setup entries in `pi.yaml`
- The difference between bare string and object form entries
- How to write idempotent custom setup automations
- How `pi-setup-<project>` auto-sources shortcuts after setup

---

## The `setup:` block

Setup entries live in `pi.yaml` under the `setup:` key. Each entry references an automation to run during `pi setup`:

```yaml
# pi.yaml
project: my-app

setup:
  - pi:install-homebrew
  - run: pi:install-python
    with:
      version: "3.13"
  - setup/install-project-deps
```

Entries run sequentially, top to bottom. If any entry fails, setup stops.

## Bare strings vs object form

Setup entries come in two forms:

**Bare string** — no conditions, no inputs:

```yaml
setup:
  - setup/install-go
  - pi:install-homebrew
```

**Object form** — when you need `if:` or `with:`:

```yaml
setup:
  - run: pi:install-python
    with:
      version: "3.13"
  - run: setup/install-ruby
    if: os.macos
```

Both forms can be mixed in the same list. Use bare strings when there are no modifiers — they're shorter and easier to scan.

## Building a realistic setup sequence

Here's a complete setup for a project with a Python backend, a Node frontend, and platform-specific dependencies:

```yaml
# pi.yaml
project: vyper-platform

setup:
  # 1. Platform package manager
  - pi:install-homebrew

  # 2. Language runtimes
  - run: pi:install-python
    with:
      version: "3.13"
  - run: pi:install-node
    with:
      version: "20"

  # 3. Project-specific dependencies
  - setup/install-project-deps

  # 4. Dev environment configuration
  - run: setup/configure-git-hooks
    if: dir.exists(".git")
```

Step by step:

1. **`pi:install-homebrew`** is a built-in automation. It has its own `if: os.macos` condition, so it's skipped automatically on Linux — no need to add `if:` on the setup entry.

2. **`pi:install-python`** and **`pi:install-node`** are built-in installers that accept a `version` input. They check if the right version is already installed before doing anything.

3. **`setup/install-project-deps`** is a custom automation (see below) that installs npm and pip dependencies.

4. **`setup/configure-git-hooks`** only runs if the project is a git repo — useful when the same setup config is used in CI containers where `.git/` may not exist.

## Writing a custom setup automation

Setup automations should be **idempotent** — check first, act only if needed. Here's the project-specific dependency installer:

```yaml
# .pi/setup/install-project-deps.yaml
description: Install all project dependencies

steps:
  - bash: npm ci
    dir: frontend

  - bash: pip install -r requirements.txt
    dir: backend
```

And a git hooks installer:

```yaml
# .pi/setup/configure-git-hooks.yaml
description: Install git hooks from hooks/ directory
bash: |
  for hook in hooks/*; do
    name="$(basename "$hook")"
    cp "$hook" ".git/hooks/$name"
    chmod +x ".git/hooks/$name"
  done
```

:::tip
PI ships a built-in for this: `pi:git/install-hooks` with a `source` input. Use it instead of writing your own when possible.
:::

## Using built-in installer automations

PI ships with installer automations that handle the test-install-verify lifecycle automatically. They check whether a tool is already installed before running the installer, and they print clean status output:

```
  ✓  install-homebrew      already installed   (4.2.1)
  →  install-python        installing...
  ✓  install-python        installed           (3.13.0)
  ✓  install-node          already installed   (20.11.0)
```

Available built-in installers:

| Automation | Inputs | Notes |
|-----------|--------|-------|
| `pi:install-homebrew` | — | macOS only (has `if: os.macos`) |
| `pi:install-python` | `version` | Tries mise, then brew |
| `pi:install-node` | `version` | Tries mise, then brew |
| `pi:install-go` | `version` | Tries mise, then brew |
| `pi:install-rust` | `version` | Uses rustup |
| `pi:install-uv` | — | Official installer script |
| `pi:install-tsx` | — | `npm install -g tsx` |

## Conditional setup entries

Use `if:` on setup entries to skip platform-specific or environment-specific steps:

```yaml
setup:
  # Only on macOS
  - run: setup/install-xcode-tools
    if: os.macos

  # Only when Docker isn't already available
  - run: setup/install-docker
    if: not command.docker

  # Only on ARM machines
  - run: setup/install-arm-toolchain
    if: os.arch.arm64

  # Only when a .env file exists
  - run: setup/load-env
    if: file.exists(".env")
```

See the [Conditions reference](/reference/conditions/) for all available predicates.

## The `pi-setup-<project>` helper

After running `pi shell`, your shell gets a helper function named `pi-setup-<project>`. For a project named `vyper-platform`, it's `pi-setup-vyper-platform`.

This function wraps `pi setup` with the eval pattern, which means:

1. It runs all setup automations
2. It installs shell shortcuts
3. It **automatically sources `.zshrc`** so shortcuts are available immediately — no need to open a new terminal

```bash
pi-setup-vyper-platform
# All done — shortcuts are live in this terminal
```

:::tip[First-time bootstrapping]
On the very first setup (before any shell wrapper exists), run:

```bash
pi setup
source ~/.zshrc
```

After that, `pi-setup-<project>` handles everything automatically.
:::

## Packages in setup

If your `pi.yaml` declares [packages](/concepts/packages/), `pi setup` fetches them before running any setup automations:

```
  ↓  yotam180/pi-common@v1.2          fetching...
  ✓  yotam180/pi-common@v1.2          cached
  ✓  file:~/my-automations            found  (alias: mytools)
```

This means you can use package automations in your setup entries:

```yaml
packages:
  - source: your-org/pi-shared@v2.0
    as: shared

setup:
  - shared/setup/install-deps
  - shared/setup/configure-env
```

## The `--no-shell` and `--silent` flags

**`--no-shell`** skips the shell shortcut installation step. Useful in CI or when you only want to run setup automations without touching the shell config:

```bash
pi setup --no-shell
```

**`--silent`** suppresses the PI status lines for installer automations:

```bash
pi setup --silent
```

**`--loud`** forces all steps to print trace lines and output, overriding any `silent: true` flags:

```bash
pi setup --loud
```

:::note
`pi setup` automatically skips shell installation when it detects a CI environment (via `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, or similar environment variables).
:::

## Testing your setup

The best way to validate your setup configuration is to run it on a clean machine. Use Docker:

```yaml
# .pi/test-setup.yaml
description: Test pi setup in a clean container
bash: |
  docker run --rm -v "$(pwd):/workspace" -w /workspace golang:1.23-bookworm \
    sh -c "go install ./... && pi setup --no-shell"
```

Or simply ask a colleague to clone the repo and run:

```bash
pi setup
```

If something is missing, `pi doctor` can help diagnose which requirements are not satisfied:

```bash
pi doctor
```

## Summary

- Setup entries go in `pi.yaml → setup:`
- Use bare strings for simple entries, object form for entries with `if:` or `with:`
- Write idempotent setup automations — check before acting
- Use PI's built-in installers (`pi:install-*`) for common tools
- `pi-setup-<project>` handles auto-sourcing after setup
- Test on a clean machine with Docker or `pi doctor`
