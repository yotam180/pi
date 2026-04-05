---
title: Built-in Automations
description: Reference for all pi:* built-in automations shipped with PI
---

PI ships with a standard library of automations for common tasks. Reference them with the `pi:` prefix in `run:` steps or `setup:` entries.

```yaml
setup:
  - run: pi:install-python
    with:
      version: "3.13"
  - run: pi:install-node
    with:
      version: "22"
  - pi:install-homebrew
```

Built-in automations are hidden from `pi list` by default. Use `pi list --builtins` to see them.

---

## Installers

Installer automations use the [`install:` block](/reference/automation-yaml/#install-block). PI manages all status output — test, install, verify, and version display.

### `pi:install-homebrew`

Install Homebrew (macOS only).

| | |
|---|---|
| **Condition** | `if: os.macos` — skipped on non-macOS |
| **Inputs** | none |
| **Test** | `command -v brew` |
| **Install** | Official Homebrew install script |
| **Version** | `brew --version` |

```yaml
- pi:install-homebrew
```

### `pi:install-python`

Install Python at a specific version. Tries [mise](https://mise.jdx.dev/) first, then Homebrew.

| | |
|---|---|
| **Inputs** | `version` (string, required) — e.g. `"3.13"` |
| **Test** | `python3 --version` output matches requested version |
| **Install** | `first:` block — mise → brew → error |
| **Version** | `python3 --version` |

```yaml
- run: pi:install-python
  with:
    version: "3.13"
```

### `pi:install-node`

Install Node.js at a specific major version. Tries mise first, then Homebrew.

| | |
|---|---|
| **Inputs** | `version` (string, required) — major version, e.g. `"22"` |
| **Test** | `node --version` major component matches requested version |
| **Install** | `first:` block — mise → brew → error |
| **Version** | `node --version` |

```yaml
- run: pi:install-node
  with:
    version: "22"
```

### `pi:install-go`

Install Go at a specific version. Tries mise first, then Homebrew.

| | |
|---|---|
| **Inputs** | `version` (string, required) — major.minor, e.g. `"1.23"` |
| **Test** | `go version` major.minor matches requested version |
| **Install** | `first:` block — mise → brew → error |
| **Version** | `go version` |

```yaml
- run: pi:install-go
  with:
    version: "1.23"
```

### `pi:install-rust`

Install Rust at a specific version. Uses rustup if available, otherwise installs via the official rustup.rs installer.

| | |
|---|---|
| **Inputs** | `version` (string, required) — major.minor, e.g. `"1.88"` |
| **Test** | `rustc --version` major.minor matches requested version |
| **Install** | rustup install + default (or curl installer) |
| **Version** | `rustc --version` |

```yaml
- run: pi:install-rust
  with:
    version: "1.88"
```

### `pi:install-uv`

Install the [uv](https://docs.astral.sh/uv/) Python package manager.

| | |
|---|---|
| **Inputs** | none |
| **Test** | `command -v uv` |
| **Install** | Official uv install script |
| **Version** | `uv --version` |

```yaml
- pi:install-uv
```

### `pi:install-tsx`

Install [tsx](https://tsx.is/) globally for TypeScript execution.

| | |
|---|---|
| **Inputs** | none |
| **Test** | `command -v tsx` |
| **Install** | `npm install -g tsx` |
| **Version** | `tsx --version` |

```yaml
- pi:install-tsx
```

---

## Docker

Docker automations use single-step shorthand. They detect `docker compose` (v2 plugin) first, falling back to `docker-compose` (v1 standalone). All CLI arguments are forwarded via `"$@"`.

### `pi:docker/up`

Start Docker Compose services in detached mode.

```yaml
- run: pi:docker/up
```

```bash
# Equivalent to:
docker compose up -d
```

### `pi:docker/down`

Stop and remove Docker Compose services.

```yaml
- run: pi:docker/down
```

### `pi:docker/logs`

Stream Docker Compose service logs (last 200 lines, follow mode).

```yaml
- run: pi:docker/logs
```

---

## Dev Tools

### `pi:cursor/install-extensions`

Install missing Cursor editor extensions from a list.

| | |
|---|---|
| **Inputs** | `extensions` (string, required) — extension IDs, newline-separated or comma-separated |

Checks currently installed extensions via `cursor --list-extensions` and only installs missing ones. Reports how many were installed vs. already present.

```yaml
- run: pi:cursor/install-extensions
  with:
    extensions: |
      ms-python.python
      dbaeumer.vscode-eslint
      esbenp.prettier-vscode
```

### `pi:git/install-hooks`

Install git hooks from a source directory into `.git/hooks/`. Uses `cmp` for idempotency — only copies changed files.

| | |
|---|---|
| **Inputs** | `source` (string, required) — directory containing hook scripts, relative to repo root |

```yaml
- run: pi:git/install-hooks
  with:
    source: .githooks
```

---

## Using Built-ins

### In setup entries

```yaml
setup:
  - pi:install-homebrew
  - run: pi:install-python
    with:
      version: "3.13"
```

### In run steps

```yaml
steps:
  - run: pi:docker/up
  - bash: echo "Services started"
```

### Listing built-ins

```bash
pi list --builtins     # include pi:* in the list
pi info pi:install-go  # detailed info for a built-in
```

### Local shadowing

If your project has a `.pi/docker/up.yaml`, it takes priority over `pi:docker/up` when referenced as `docker/up`. Use the `pi:` prefix to explicitly call the built-in version regardless of local automations.
