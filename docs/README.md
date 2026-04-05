# PI — Product Definition

PI (`pi`) is a developer automation platform for teams managing complex repositories. It replaces shell shortcut files and setup scripts with a structured, polyglot, and shareable automation model.

---

## The Problem

Most engineering teams accumulate a `shell_shortcuts.sh` and a `setup_environment.sh` in their repo. These files:
- Are written entirely in bash, which is readable for simple things but ugly for logic
- Grow to 1000+ lines over time with no clear structure
- Can't easily mix languages (step in Python, step in bash, step in TypeScript)
- Have no standard way to share automations between projects or teams
- Require manual setup by every developer (copy this into your `.zshrc`, run this script...)

PI solves all of this.

---

## Core Concepts

### The `.pi/` folder

Every project using PI has a `.pi/` folder at the repo root. This is where all local automations live, organized however the team prefers — by domain, by tool, by lifecycle stage.

```
.pi/
  docker/
    up.yaml
    down.yaml
    logs-formatted.yaml
    logs-formatted.py        ← scripts live right next to their automation
  setup/
    install-cursor-extensions/
      automation.yaml
      extensions.txt         ← bundled assets (vsix lists, configs, etc.)
    install-uv.yaml
  deploy/
    push-image.yaml
    helm-upgrade.yaml
```

Each `.yaml` file in `.pi/` defines one automation. Scripts and assets used by an automation live in the same folder as the automation file.

### `pi.yaml`

The root config file. Stays small — it only defines:
- Which automations are exposed as shell shortcuts
- Which automations run during `pi setup`
- Project-level settings

```yaml
project: vyper-platform

shortcuts:
  vpup:    docker/up
  vpdown:  docker/down
  vplf:    docker/logs-formatted
  vpbrl:   docker/build-run-logs
  vpbrlf:  docker/build-run-logs-formatted
  sk_deploy:
    run: deploy/push-image
    anywhere: true    # default is repo-root; this flag lets it run from anywhere

setup:
  - run: setup/install-brew
    if: os.macos
  - run: setup/install-uv
    if: not command.uv
  - run: setup/install-cursor-extensions
  - run: pi:install-python          # built-in PI automation
    with:
      version: "3.13"
```

### Automations

An automation is a YAML file that defines a named, runnable unit of work. It can chain steps written in different languages. The automation name is derived from the file path — no `name:` field is needed.

```yaml
# .pi/docker/logs-formatted.yaml
# Name is automatically "docker/logs-formatted" (derived from path)
description: Stream container logs through the log formatter

steps:
  - bash: docker-compose logs -f --tail 200
    pipe_to: next

  - python: logs-formatted.py     # path relative to this automation file
```

```yaml
# .pi/setup/install-cursor-extensions/automation.yaml
# Name is automatically "setup/install-cursor-extensions" (derived from path)
description: Install required Cursor extensions if missing

steps:
  - bash: |
      while IFS= read -r ext; do
        if ! cursor --list-extensions 2>/dev/null | grep -qx "$ext"; then
          cursor --install-extension "$ext" --force
        fi
      done < "$(dirname "$0")/extensions.txt"
```

> **Note:** The `name:` field is optional and deprecated. PI derives the automation name from the file path: `.pi/docker/up.yaml` → `docker/up`, `.pi/setup/cursor/automation.yaml` → `setup/cursor`. If `name:` is present and matches the derived name, it's accepted silently. If it mismatches, PI prints a warning. Existing files with `name:` continue to work.

---

## Step Types

| Type         | Usage                                                  |
|--------------|--------------------------------------------------------|
| `bash`       | Inline shell or a `.sh` file path                      |
| `python`     | Inline script or a `.py` file path                     |
| `typescript` | Inline script or a `.ts` file path (run via `tsx`)     |
| `run`        | Call another automation by name (local or marketplace) |

Steps can pass data to the next step using `pipe_to: next`. Full inter-step communication (env, named outputs) is planned for a future iteration.

### Step Description (`description:`)

Steps can declare an optional `description:` field to document what the step does in human-readable terms. Descriptions are displayed by `pi info` and have no effect on execution.

```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    description: Stream container logs
    pipe_to: next

  - python: logs-formatted.py
    description: Format and filter log output
```

Steps without `description:` behave exactly as before. The description is purely informational — it helps developers understand complex automations when browsing with `pi info`.

### Environment Variables (`env:`)

Steps can declare an `env:` mapping to inject environment variables into the step's execution context:

```yaml
steps:
  - bash: go build -o bin/app ./...
    env:
      GOOS: linux
      GOARCH: amd64
      CGO_ENABLED: "0"

  - bash: go test ./...
```

Environment variables are scoped to the step — they do not leak to subsequent steps. Step-level env vars override parent process env vars with the same name.

### Working Directory (`dir:`)

Steps can declare a `dir:` field to override the working directory for that step's execution. By default, all steps run in the project root (directory containing `pi.yaml`).

```yaml
steps:
  - bash: go test ./...
    dir: services/api

  - bash: npm install
    dir: frontend

  - bash: echo "back in root"
```

The `dir:` path is resolved relative to the project root. Absolute paths are used as-is. The directory must exist at execution time — if it doesn't, PI reports a clear error.

Working directories are per-step — each step independently resolves its own `dir:`, and steps without `dir:` always use the project root (no implicit carry-over from a previous step).

### Timeout (`timeout:`)

Steps can declare a `timeout:` field to set a maximum execution duration. If the step exceeds the timeout, PI kills the process and returns exit code 124 (matching the GNU `timeout` command convention).

```yaml
steps:
  - bash: go build ./...
    timeout: 30s

  - bash: npm test
    timeout: 5m

  - python: train_model.py
    timeout: 1h30m
```

The value is a Go-style duration string (e.g., `30s`, `5m`, `1h30m`). Timeout must be positive — zero or negative values are rejected at parse time.

Timeout works with all subprocess step types (`bash`, `python`, `typescript`). It cannot be used on `run:` steps (set timeouts on the target automation's own steps instead) or `parent_shell` steps (they don't execute as subprocesses).

Timeout is compatible with all other step fields: `env:`, `dir:`, `silent:`, `if:`, and `pipe_to`. When a step with `if:` is skipped, no timeout applies. When a step with `silent: true` times out, the timeout error still propagates.

### Step Trace Lines

By default, PI prints a trace line to stderr before executing each step:

```
  → bash: echo "building project..."
  → run:  setup/install-deps
  → python: transform.py
```

Installer steps are exempt — they have their own formatted output.

### Silent Steps (`silent: true`)

A step can declare `silent: true` to suppress both its trace line and its stdout/stderr output. This is useful for noisy housekeeping commands that clutter the terminal:

```yaml
steps:
  - bash: echo "Starting build..."
  - bash: rm -rf dist && mkdir dist
    silent: true
  - bash: go build -o dist/app ./...
```

The silent step still executes — only its output is suppressed. Piped data (`pipe_to: next`) still flows through silent steps.

### Loud Mode (`--loud`)

Passing `--loud` to `pi run` or `pi setup` overrides all `silent: true` flags, forcing every step to print its trace line and output. This is the escape hatch for debugging:

```bash
pi run --loud build
pi setup --loud
```

### Parent Shell Steps (`parent_shell: true`)

A bash step can declare `parent_shell: true` to run the command in the **calling shell** instead of as a subprocess. This is essential for commands that must affect the parent shell's state, such as activating a virtualenv or changing directory.

```yaml
steps:
  - bash: echo "Setting up..."
  - bash: source venv/bin/activate
    parent_shell: true
  - bash: cd /some/dir
    parent_shell: true
```

When `parent_shell: true` is set:
- PI does **not** execute the step itself
- Instead, it appends the command to a temp file (`PI_PARENT_EVAL_FILE`)
- After PI exits, the shell wrapper `eval`s the file, running the command in the parent shell

**Requirements**:
- Only valid on `bash` steps
- Cannot be combined with `pipe_to`
- Must be invoked via a PI shell shortcut (which sets `PI_PARENT_EVAL_FILE`); running directly with `pi run` will error unless the env var is set manually

### Auto-Sourcing After Setup

When `pi setup` is run via a shell shortcut (or with `PI_PARENT_EVAL_FILE` set), it automatically writes `source ~/.zshrc` (or the equivalent) to the eval file after installing shell shortcuts. This means shortcuts are available immediately in the current terminal — no need to open a new terminal or manually source the rc file.

The generated shell file includes a `pi-setup-<project>` helper function that wraps `pi setup` with the eval pattern:

```bash
# Generated by pi shell — use this for zero-friction setup
pi-setup-myproject
```

### Installer Automations (`install:`)

Automations that install tools use the `install:` block instead of `steps:`. The two are mutually exclusive. The `install:` block explicitly declares a test-run-verify lifecycle, and PI manages all status output.

```yaml
# Scalar shorthand for simple installs
name: install-homebrew
description: Install Homebrew (macOS only)
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  version: brew --version | head -1 | awk '{print $2}'
```

```yaml
# Step list for conditional install paths
name: install-python
description: Install Python at a specific version

inputs:
  version:
    type: string
    description: Python version to install

install:
  test:
    - bash: python3 --version 2>&1 | grep -q "Python $PI_INPUT_VERSION"
  run:
    - bash: mise install "python@$PI_INPUT_VERSION" && mise use "python@$PI_INPUT_VERSION"
      if: command.mise
    - bash: brew install "python@$PI_INPUT_VERSION"
      if: command.brew and not command.mise
  version: python3 --version 2>&1 | awk '{print $2}'
```

PI prints one status line per installer:
```
  ✓  install-homebrew      already installed   (4.2.1)
  →  install-python        installing...
  ✓  install-python        installed           (3.13.0)
  ✗  install-node          failed
```

When `verify:` is omitted, PI re-runs `test:` as verification after `run:` completes. Use `--silent` to suppress status lines.

### Conditional Steps (`if:`)

Steps can declare an `if:` field to conditionally execute based on the runtime environment. If the condition evaluates to false, the step is silently skipped.

```yaml
steps:
  - bash: brew install jq
    if: os.macos and not command.jq

  - bash: apt-get install -y jq
    if: os.linux and not command.jq

  - bash: echo "jq is ready"
```

Supported predicates: `os.macos`, `os.linux`, `os.windows`, `os.arch.arm64`, `os.arch.amd64`, `env.<NAME>`, `command.<name>`, `file.exists("<path>")`, `dir.exists("<path>")`, `shell.zsh`, `shell.bash`. Combine with `and`, `or`, `not`, and parentheses.

When a step with `pipe_to: next` is skipped, any piped data passes through to the next step unchanged.

---

## Shell Shortcuts

Running `pi shell` writes shortcut functions into the user's shell config (`.zshrc` / `.bashrc`). Each shortcut is a shell function that:

1. By default, `cd`s to the repo root and runs `pi run <automation>` from there
2. With `anywhere: true`, runs `pi run` without changing directory (useful for global ops)
3. Wraps `pi run` with a `PI_PARENT_EVAL_FILE` eval pattern so that `parent_shell: true` steps work

This means shortcuts like `vpup` work from anywhere in the terminal — the developer never has to think about their working directory.

```bash
# Generated by: pi shell
function vpup() {
  local _pi_eval_file
  _pi_eval_file="$(mktemp)"
  PI_PARENT_EVAL_FILE="$_pi_eval_file" (cd /path/to/vyper-platform && pi run docker/up "$@")
  local _pi_exit=$?
  if [ -s "$_pi_eval_file" ]; then
    source "$_pi_eval_file"
  fi
  rm -f "$_pi_eval_file"
  return $_pi_exit
}
```

In addition, `pi shell` generates a `pi-setup-<project>` helper function that wraps `pi setup` with the same eval pattern, enabling auto-sourcing of shell shortcuts after setup completes.

---

## Environment Setup

`pi setup` runs all automations listed in `pi.yaml → setup:` sequentially. Setup automations are expected to be idempotent — check first, act only if needed.

Setup entries support the same `if:` conditions as automation steps. When a condition evaluates to false, the entry is skipped with a message. Entries without `if:` always run.

```yaml
# .pi/setup/install-uv.yaml
steps:
  - bash: |
      if ! command -v uv &> /dev/null; then
        curl -LsSf https://astral.sh/uv/install.sh | sh
      fi
```

---

## Automation Resolution

When PI encounters an automation name, it resolves in this order:

1. **Local** — `.pi/<name>.yaml` or `.pi/<name>/automation.yaml`
2. **Built-in** — automations shipped with the PI binary (prefixed `pi:`)
3. **Marketplace** — cached from GitHub, referenced as `org/package@version`

### Built-in library (`pi:`)

PI ships with a standard collection of automations for common tasks:

- `pi:install-python` — check/install Python at a specific version
- `pi:install-node` — check/install Node.js
- `pi:install-go` — check/install Go at a specific version
- `pi:install-rust` — check/install Rust at a specific version
- `pi:install-ruby` — check/install Ruby at a specific version
- `pi:install-uv` — check/install uv
- `pi:install-homebrew` — check/install Homebrew
- `pi:install-tsx` — check/install tsx globally via npm
- `pi:docker/up`, `pi:docker/down`, `pi:docker/logs` — standard Docker Compose ops
- `pi:cursor/install-extensions` — install missing Cursor extensions from a list
- `pi:git/install-hooks` — install git hooks from a source directory

These are defined in the PI repository's own `.pi/` folder and compiled into the binary.

### Marketplace (`org/package@version`)

Marketplace automations live in GitHub repos and are referenced by `org/package-name@version` (version = git tag). PI downloads and caches them in `~/.pi/cache/`.

```yaml
setup:
  - run: python-foundation/install-python@v3.13.0
    with:
      version: "3.13"
```

A marketplace package is just a GitHub repo with a `pi-package.yaml` at the root — no registry needed.

---

## CLI Reference

| Command                                  | Description                                              |
|-----------------------------------------|----------------------------------------------------------|
| `pi run <name> [args]`                  | Run an automation by name (args mapped to inputs)        |
| `pi run <name> --with key=value`        | Run with explicit named inputs (repeatable)              |
| `pi run --repo <path> <name>`           | Run an automation with explicit project root             |
| `pi run --silent <name>`                | Suppress PI status lines for installer automations       |
| `pi run --loud <name>`                  | Force all steps to print trace lines and output          |
| `pi info <name>`                        | Show name, description, and input docs for an automation |
| `pi setup`                    | Run all setup automations, then install shell shortcuts  |
| `pi setup --no-shell`         | Run setup automations without installing shortcuts       |
| `pi setup --silent`           | Suppress PI status lines for installer automations       |
| `pi setup --loud`             | Force all steps to print trace lines and output          |
| `pi shell`                    | Install shortcut functions into the current shell config |
| `pi shell uninstall`          | Remove shortcuts for the current project                 |
| `pi shell list`               | List all installed shortcut files across all projects    |
| `pi list`                     | List all available automations in the project            |
| `pi doctor`                   | Check requirement health for all automations             |
| `pi validate`                 | Statically validate all config and automation files      |
| `pi version`                  | Print the PI version                                     |
| `pi --version`                | Same as `pi version`                                     |
| `pi add org/package@version`  | Download and cache a marketplace automation              |

---

## Agent Workflow

> This section is for agents working on this codebase.

This folder is your memory. Read it at the start of every session.

### Folder structure

```
docs/
  projects/
    todo/          # Defined, not started
    in_progress/   # Currently being worked on
    done/          # Completed
  tasks/
    todo/          # Ready to pick up
    in_progress/   # In flight (keep to 1)
    done/          # Completed
  templates/
    project.md
    task.md
```

### Each session

1. Check `docs/tasks/in_progress/` — resume before starting anything new.
2. If nothing in progress, pick the highest-priority task from `docs/tasks/todo/`.
3. Read the parent project doc for context.
4. Work the task. Update its file with decisions and progress as you go.
5. Move to `done/` when complete. Commit before session ends.

See `AGENTS.md` for full instructions on creating tasks and projects.
