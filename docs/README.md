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
  - setup/install-cursor-extensions  # bare string — no modifiers needed
  - run: pi:install-python           # object form — has with: modifier
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
    pipe: true

  - python: logs-formatted.py     # path relative to this automation file
```

```yaml
# .pi/setup/install-cursor-extensions/automation.yaml
# Name is automatically "setup/install-cursor-extensions" (derived from path)
description: Install required Cursor extensions if missing

bash: |
  while IFS= read -r ext; do
    if ! cursor --list-extensions 2>/dev/null | grep -qx "$ext"; then
      cursor --install-extension "$ext" --force
    fi
  done < "$(dirname "$0")/extensions.txt"
```

> **Note:** The `name:` field is optional and deprecated. PI derives the automation name from the file path: `.pi/docker/up.yaml` → `docker/up`, `.pi/setup/cursor/automation.yaml` → `setup/cursor`. If `name:` is present and matches the derived name, it's accepted silently. If it mismatches, PI prints a warning. Existing files with `name:` continue to work.

### Single-Step Shorthand

Automations with a single step can skip the `steps:` wrapper and place the step type key at the top level:

```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```

This is equivalent to:

```yaml
description: Run the test suite
steps:
  - bash: go test ./...
```

All step types work as shorthands: `bash:`, `python:`, `typescript:`, `run:`. Step modifier fields (`env:`, `dir:`, `timeout:`, `silent:`, `pipe:`) can be used alongside the shorthand key at the top level:

```yaml
# .pi/build-linux.yaml
description: Cross-compile for Linux
bash: go build -o bin/app ./...
env:
  GOOS: linux
  GOARCH: amd64
dir: services/api
timeout: 30s
```

Having both a top-level step key and `steps:` (or `install:`) in the same file is a parse error. The `description:` at the top level remains the automation description. The `if:` at the top level is the automation-level condition.

---

## Step Types

| Type         | Usage                                                  |
|--------------|--------------------------------------------------------|
| `bash`       | Inline shell or a `.sh` file path                      |
| `python`     | Inline script or a `.py` file path                     |
| `typescript` | Inline script or a `.ts` file path (run via `tsx`)     |
| `run`        | Call another automation by name (local or marketplace) |

Steps can pass data to the next step using `pipe: true`. Full inter-step communication (env, named outputs) is planned for a future iteration.

> **Deprecation:** `pipe_to: next` is the legacy form and still works, but emits a deprecation warning. Use `pipe: true` instead.

### Step Description (`description:`)

Steps can declare an optional `description:` field to document what the step does in human-readable terms. Descriptions are displayed by `pi info` and have no effect on execution.

```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    description: Stream container logs
    pipe: true

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

### Automation-Level Environment Variables (`env:`)

Automations can declare a top-level `env:` mapping so every step sees the same variables without repeating them per step. Step-level `env:` overrides automation-level `env:` for the same key.

```yaml
description: Cross-compile fzf for Linux amd64

env:
  GOOS: linux
  GOARCH: amd64
  CGO_ENABLED: "0"

steps:
  - bash: go build -o target/fzf-linux ./...
  - bash: sha256sum target/fzf-linux > target/fzf-linux.sha256
  - bash: echo "Done. Binary at target/fzf-linux"
```

Automation-level env does not propagate into sub-automations invoked by `run:` steps — each `run:` starts with that automation’s own env (and the process environment), not the caller’s declared automation env.

Single-step shorthand can use a top-level `env:` next to `bash:`, `run:`, etc.; that `env:` is automation-level, the same as in multi-step files.

### Extra Arguments (`PI_ARGS`)

Everything after the automation name on `pi run` is forwarded as automation arguments — no `--` separator needed. PI's own flags (`--silent`, `--loud`, `--repo`, `--with`) must come **before** the automation name:

```
pi run --silent build --release --verbose
       ^^^^^^^^ PI flag (before automation name)
                ^^^^^ automation name
                      ^^^^^^^^^^^^^^^^^^ automation arguments
```

For automations **without** `inputs:`, forwarded args are available in several ways:

1. **`$PI_ARGS`** — environment variable containing all extra args, space-joined
2. **`$PI_ARG_1`, `$PI_ARG_2`, ...** — individual positional args as env vars (all step types)
3. **`$PI_ARG_COUNT`** — the number of extra args passed
4. **`$@`, `$1`, `$2`** — bash positional parameters (bash steps only)

`PI_ARG_N` env vars work in all step types — bash, python, and typescript — making individual arg access ergonomic across languages:

```yaml
# .pi/test.yaml
description: Run tests with optional extra flags
bash: cargo test $PI_ARGS
```

```yaml
# .pi/greet.py — python step accessing individual args
python: |
  import os
  name = os.environ.get("PI_ARG_1", "world")
  print(f"Hello, {name}!")
```

```bash
pi run test --ignored --nocapture
# Executes: cargo test --ignored --nocapture

pi run greet Alice
# PI_ARG_1=Alice, PI_ARG_COUNT=1
```

For automations **with** `inputs:`, positional args are mapped to declared inputs by declaration order (not to `PI_ARGS`). Excess positional args beyond the declared inputs produce an error.

```yaml
# .pi/deploy.yaml
description: Deploy to an environment
inputs:
  env:
    type: string
    required: true
  region:
    type: string
    default: us-east-1
bash: deploy.sh --env $PI_IN_ENV --region $PI_IN_REGION
```

```bash
pi run deploy prod eu-west-1
# Maps: env=prod, region=eu-west-1

pi run deploy staging
# Maps: env=staging, region=us-east-1 (default)

pi run --with env=prod deploy
# Named inputs also work; --with must come before the automation name
```

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

Timeout is compatible with all other step fields: `env:`, `dir:`, `silent:`, `if:`, and `pipe: true`. When a step with `if:` is skipped, no timeout applies. When a step with `silent: true` times out, the timeout error still propagates.

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

The silent step still executes — only its output is suppressed. Piped data (`pipe: true`) still flows through silent steps.

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
- Cannot be combined with `pipe: true`
- Works automatically when invoked via a PI shell shortcut or the global `pi()` wrapper (both set `PI_PARENT_EVAL_FILE`)
- If `PI_PARENT_EVAL_FILE` is not set (raw binary call without the shell wrapper), the step is skipped with a warning: run `pi shell` to install shell integration

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
description: Install Homebrew (macOS only)
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  version: brew --version | head -1 | awk '{print $2}'
```

```yaml
# Step list with first: block for conditional install paths
description: Install Python at a specific version

inputs:
  version:
    type: string
    description: Python version to install

install:
  test:
    - bash: python3 --version 2>&1 | grep -q "Python $PI_IN_VERSION"
  run:
    - first:
        - bash: mise install "python@$PI_IN_VERSION" && mise use "python@$PI_IN_VERSION"
          if: command.mise
        - bash: brew install "python@$PI_IN_VERSION"
          if: command.brew
        - bash: |
            echo "no suitable installer found (tried mise, brew)" >&2
            exit 1
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

When a step with `pipe: true` is skipped, any piped data passes through to the next step unchanged.

### First-Match Blocks (`first:`)

A `first:` block groups mutually exclusive sub-steps. The executor evaluates each sub-step's `if:` condition in order and runs only the first one that matches. All remaining sub-steps are skipped. A sub-step without `if:` always matches and acts as a fallback.

```yaml
steps:
  - first:
      - bash: mise install go@1.23 && mise use go@1.23
        if: command.mise
      - bash: brew install go
        if: command.brew
      - bash: |
          echo "no suitable installer found" >&2
          exit 1
```

This replaces the common pattern of compounding negation chains (`if: command.brew and not command.mise`) with clean, positive conditions. Each sub-step only needs its own condition.

`first:` is a generic step-level construct usable anywhere a step appears: in `steps:`, in `install.run:`, in `install.test:`. It is not specific to installers.

At the block level, `first:` supports `description:`, `if:` (to conditionally skip the entire block), and `pipe: true` (to pipe the matched sub-step's output). Sub-steps inside `first:` support all normal step fields: `env:`, `dir:`, `timeout:`, `silent:`.

If all sub-steps have `if:` conditions and none match, the block is silently skipped (consistent with how a skipped `if:` step behaves). Nested `first:` blocks are not allowed.

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
  (cd /path/to/vyper-platform && PI_PARENT_EVAL_FILE="$_pi_eval_file" pi run docker/up "$@")
  local _pi_exit=$?
  if [ -s "$_pi_eval_file" ]; then
    source "$_pi_eval_file"
  fi
  rm -f "$_pi_eval_file"
  return $_pi_exit
}
```

In addition, `pi shell` generates:

- A `pi-setup-<project>` helper function that wraps `pi setup` with the same eval pattern, enabling auto-sourcing of shell shortcuts after setup completes
- A global `pi()` wrapper function (`~/.pi/shell/_pi-wrapper.sh`) that wraps every `pi` invocation with the eval pattern, so `parent_shell: true` steps work for all `pi run` calls — not just shortcuts:

```bash
# Generated by: pi shell (global wrapper)
function pi() {
  local _pi_eval_file
  _pi_eval_file="$(mktemp)"
  PI_PARENT_EVAL_FILE="$_pi_eval_file" command pi "$@"
  local _pi_exit=$?
  if [ -s "$_pi_eval_file" ]; then
    source "$_pi_eval_file"
  fi
  rm -f "$_pi_eval_file"
  return $_pi_exit
}
```

---

## Environment Setup

`pi setup` runs all automations listed in `pi.yaml → setup:` sequentially. Setup automations are expected to be idempotent — check first, act only if needed.

Setup entries can be bare strings or objects. Bare strings are shorthand for `run: <string>` with no conditions or inputs. The object form (`run:` + optional `if:` + optional `with:`) is required for entries that need modifiers. Both forms can be mixed in the same list:

```yaml
setup:
  - setup/install-go                # bare string — simple run
  - run: setup/install-ruby         # object form — has if: modifier
    if: os.macos
  - run: pi:install-python          # object form — has with: modifier
    with:
      version: "3.13"
```

Setup entries support the same `if:` conditions as automation steps. When a condition evaluates to false, the entry is skipped with a message. Entries without `if:` always run.

```yaml
# .pi/setup/install-uv.yaml
bash: |
  if ! command -v uv &> /dev/null; then
    curl -LsSf https://astral.sh/uv/install.sh | sh
  fi
```

---

## Packages (`packages:` in `pi.yaml`)

Teams can declare external automation sources in `pi.yaml` using the `packages:` block. Package automations are discovered and merged into the project — they work just like local automations.

```yaml
packages:
  # Simple GitHub package — most common case
  - yotam180/pi-common@v1.2

  # GitHub package with explicit source key (same as above, verbose form)
  - source: yotam180/pi-common@v1.2

  # Local folder source — for developing automations without push/clone cycle
  - source: file:~/my-automations
    as: mytools           # optional alias; enables `run: mytools/docker/up`

  # Local folder without alias — reference with full file: prefix
  - source: file:~/shared-automations

  # Relative local path — resolved relative to project root
  - source: file:./packages/shared
    as: shared
```

### Source types

| Source | Format | Caching |
|--------|--------|---------|
| GitHub | `org/repo@version` | Cached in `~/.pi/cache/`; fetched once per version |
| File | `file:~/path` or `file:./relative` | Read directly from disk; no caching |

GitHub packages require a version tag after `@` — there is no implicit "latest". Mutable refs like `@main` or `@HEAD` are accepted but emit a reproducibility warning and use a date-stamped cache key (e.g. `main~20260405`). Pin to a tag for stable, reproducible builds.

File sources use `~` for the home directory and `./` for paths relative to the project root. Changes to a `file:` source are reflected immediately (no cache invalidation needed).

### Aliases (`as:`)

The `as:` key lets you write `run: mytools/docker/up` instead of referencing the full source. Aliases must be unique within a `pi.yaml`. An alias that collides with a local `.pi/` automation path emits a warning (local wins per resolution order). Aliases must not contain `/`.

### `pi add`

The `pi add` command is the ergonomic way to declare a package dependency:

```bash
# Add a GitHub package (fetches into cache immediately)
pi add yotam180/pi-common@v1.2

# Add a local folder source
pi add file:~/shared-automations

# Add a local folder with an alias
pi add file:~/my-automations --as mytools
```

The command validates the source, fetches GitHub packages into `~/.pi/cache/`, and appends the entry to `pi.yaml`. Adding the same source twice is a no-op — PI prints "already in pi.yaml" and exits successfully.

GitHub sources without `@version` are rejected with a clear message: `version required — use pi add org/repo@<tag>`.

### `pi setup` integration

Before running any setup automations, `pi setup` fetches all declared GitHub packages that aren't already cached. `file:` entries are verified to exist on disk (a warning is printed if not, but setup continues). Status output:

```
  ↓  yotam180/pi-common@v1.2          fetching...
  ✓  yotam180/pi-common@v1.2          cached
  ✓  file:~/my-automations            found  (alias: mytools)
  ⚠  file:~/missing-path              not found
```

### On-demand fetching

When a `run:` step references an undeclared GitHub package (e.g. `org/repo@v1.0/docker/up`), PI fetches it automatically instead of failing. After fetching, PI prints an advisory to stderr with a ready-to-paste `packages:` snippet:

```
  ↓  org/repo@v1.0          fetched (on demand)

  tip: add to pi.yaml to avoid fetching on every fresh clone:

    packages:
      - org/repo@v1.0
```

On-demand fetching uses the same cache — subsequent runs in the same environment won't re-fetch. The advisory only appears when a live network fetch happens, not when the package is already cached.

`file:` sources are never fetched on demand — they must be declared in `pi.yaml`.

### Package automations in `pi list` / `pi run`

Package automations that don't collide with local names appear in `pi list` with a SOURCE column showing the package source or alias. They are directly runnable via `pi run`:

```bash
# Run a package automation by its plain name (if no local collision)
pi run docker/up

# Run via alias
pi run mytools/docker/up

# Run via full GitHub reference (works even if not declared in packages:)
pi run org/repo@v1.0/docker/up
```

Use `pi list --all` to browse automations from all declared packages, grouped by source with separator headers.

### Private repositories

PI tries multiple auth methods when fetching a GitHub package, in order:

1. **SSH** — `git@github.com:org/repo.git`. Works if you have an SSH key configured for GitHub.
2. **HTTPS with token** — uses the `GITHUB_TOKEN` environment variable if set.
3. **Plain HTTPS** — works for public repos only.

For private repos, ensure either:
- An SSH key is configured (`ssh -T git@github.com` should succeed), or
- `GITHUB_TOKEN` is set to a personal access token with repo read access

If all methods fail, PI prints an error with instructions:
```
could not fetch org/repo: check network and that the repo exists.
For private repos:
  • Ensure an SSH key is configured (git@github.com:org/repo.git)
  • Or set GITHUB_TOKEN env var for HTTPS auth
```

### Writing a package repo

A PI package is just a GitHub repo with automations in a `.pi/` folder — the same structure as any PI project. No registry, no special tooling required.

Minimal example:
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

Tag a release to make it available:
```bash
git tag v1.0
git push origin v1.0
```

Users consume it with:
```bash
pi add your-org/my-pi-package@v1.0
```

### `pi-package.yaml`

An optional file at the root of a package repo. Its only supported field is `min_pi_version`:

```yaml
min_pi_version: "0.5.0"
```

When present, PI checks the running version against `min_pi_version` at fetch time. If the running PI is older, the fetch fails with a clear upgrade message. Dev builds (`version == "dev"`) skip this check.

If `pi-package.yaml` is absent or empty, the package works with any PI version.

---

## Automation Resolution

When PI encounters an automation name, it resolves in this order:

1. **Local** — `.pi/<name>.yaml` or `.pi/<name>/automation.yaml`
2. **Package** — automations from declared `packages:` sources (GitHub or `file:`)
3. **Built-in** — automations shipped with the PI binary (prefixed `pi:`)
4. **On-demand** — undeclared GitHub references (`org/repo@version/path`) are fetched automatically

Local always wins. If a local automation shadows a package automation, a warning is printed.

### Built-in library (`pi:`)

PI ships with a standard collection of automations for common tasks:

- `pi:install-python` — check/install Python at a specific version
- `pi:install-node` — check/install Node.js
- `pi:install-go` — check/install Go at a specific version
- `pi:install-rust` — check/install Rust at a specific version
- `pi:install-uv` — check/install uv
- `pi:install-homebrew` — check/install Homebrew
- `pi:install-tsx` — check/install tsx globally via npm
- `pi:install-terraform` — check/install Terraform at a specific version
- `pi:install-kubectl` — check/install kubectl at a specific version
- `pi:install-helm` — check/install Helm at a specific version
- `pi:install-pnpm` — check/install pnpm (optional version)
- `pi:install-bun` — check/install Bun JavaScript runtime (optional version)
- `pi:install-deno` — check/install Deno runtime (optional version)
- `pi:install-aws-cli` — check/install AWS CLI v2
- `pi:docker/up`, `pi:docker/down`, `pi:docker/logs` — standard Docker Compose ops
- `pi:uv/sync` — sync Python project dependencies using uv
- `pi:npm/install` — install Node.js dependencies using npm ci or npm install
- `pi:set-env` — idempotently add an environment variable export to the shell config
- `pi:cursor/install-extensions` — install missing Cursor extensions from a list
- `pi:git/install-hooks` — install git hooks from a source directory
- `pi:version-satisfies` — check whether a version string satisfies a semver constraint (Go-native)

These are defined in the PI repository's own `.pi/` folder and compiled into the binary.

### Inter-step output capture

Steps can pass their stdout to subsequent steps via `outputs.last` in `with:` values:

```yaml
steps:
  - bash: node --version 2>/dev/null | sed 's/^v//'
  - run: pi:version-satisfies
    with:
      version: outputs.last      # stdout of the previous step, trimmed
      required: inputs.version   # current automation's input value
```

Supported interpolation references in `with:` values:
- `outputs.last` — trimmed stdout of the immediately preceding step
- `outputs.<N>` — stdout of step N (0-indexed)
- `inputs.<name>` — current automation's input value

Literal strings in `with:` values pass through unchanged.

---

## CLI Reference

| Command                                  | Description                                              |
|-----------------------------------------|----------------------------------------------------------|
| `pi run <name> [args]`                  | Run an automation by name (args after name forwarded)    |
| `pi run --with key=value <name>`        | Run with explicit named inputs (repeatable; before name) |
| `pi run --repo <path> <name>`           | Run an automation with explicit project root             |
| `pi run --silent <name>`                | Suppress PI status lines for installer automations       |
| `pi run --loud <name>`                  | Force all steps to print trace lines and output          |
| `pi info <name>`                        | Show name, description, and input docs for an automation |
| `pi setup`                    | Run all setup automations, then install shell shortcuts  |
| `pi setup --no-shell`         | Run setup automations without installing shortcuts       |
| `pi setup --silent`           | Suppress PI status lines for installer automations       |
| `pi setup --loud`             | Force all steps to print trace lines and output          |
| `pi setup add <name> [key=value ...]` | Run automation then add to `pi.yaml`; resolves short-form names (`python` → `pi:install-python`); idempotent |
| `pi setup add <name> --version <v>`   | Add with `with: version: "<v>"`                          |
| `pi setup add <name> --if <expr>`     | Add with `if: <expr>` condition                          |
| `pi setup add <name> --only-add`      | Skip execution, write entry directly (CI/pre-configured) |
| `pi setup add <name> --yes`           | Auto-confirm project initialization if no `pi.yaml`      |
| `pi shell`                    | Install shortcut functions into the current shell config |
| `pi shell uninstall`          | Remove shortcuts for the current project                 |
| `pi shell list`               | List all installed shortcut files across all projects    |
| `pi list`                     | List all available automations with SOURCE column        |
| `pi list --all`               | Include all automations from declared packages, grouped  |
| `pi list --builtins`          | Include built-in (`pi:*`) automations in the list        |
| `pi doctor`                   | Check requirement health for all automations             |
| `pi validate`                 | Statically validate all config and automation files      |
| `pi validate --warnings`      | Include non-fatal warnings (missing descriptions, unused automations, shortcut shadowing) |
| `pi version`                  | Print the PI version                                     |
| `pi --version`                | Same as `pi version`                                     |
| `pi init`                         | Initialize a new PI project (`pi.yaml` + `.pi/` + example) |
| `pi init --name <name>`          | Initialize with a specific project name (non-interactive) |
| `pi init --yes`                   | Accept inferred project name without prompting            |
| `pi new <name>`                   | Scaffold a new automation file in `.pi/`                  |
| `pi new <name> --bash "cmd"`     | Pre-fill with a bash command                              |
| `pi new <name> --python "file"`  | Pre-fill with a python script reference                   |
| `pi new <name> -d "description"` | Set the automation description                            |
| `pi add <source> [--as <alias>]` | Append a package to `pi.yaml` (`org/repo@version` or `file:...`); fetch GitHub into cache; duplicate entry is a no-op ("already in pi.yaml") |
| `pi completion <shell>`         | Generate shell completion script (`bash`, `zsh`, `fish`, `powershell`) |

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
