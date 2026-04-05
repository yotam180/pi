---
title: Automations
description: How PI automations work — naming, steps, modifiers, and the install lifecycle
---

An automation is a YAML file in `.pi/` that defines a named, runnable unit of work. It can be a single command or a multi-step pipeline mixing bash, Python, and TypeScript.

## In this section

- [What is an automation](#what-is-an-automation) — files, naming, and the `.pi/` folder
- [Single-step shorthand](#single-step-shorthand) — the most common form
- [Multi-step automations](#multi-step-automations) — chaining steps with `steps:`
- [Step modifiers](#step-modifiers) — env, dir, timeout, silent, if, pipe, description, parent_shell
- [Step trace lines](#step-trace-lines) — what PI prints before each step
- [First-match blocks](#first-match-blocks) — mutual exclusion with `first:`
- [Installer automations](#installer-automations) — the `install:` lifecycle
- [Automation-level conditions](#automation-level-conditions) — skip the whole automation with `if:`
- [Inputs](#inputs) — parameterize automations with `inputs:`

---

## What is an automation

Every `.yaml` file inside `.pi/` is an automation. The automation name is derived from the file path — there is no `name:` field.

| File path | Automation name |
|-----------|----------------|
| `.pi/test.yaml` | `test` |
| `.pi/docker/up.yaml` | `docker/up` |
| `.pi/setup/install-deps/automation.yaml` | `setup/install-deps` |

Two files that resolve to the same name are a hard error.

:::note
The `name:` field exists for backward compatibility but is deprecated. When present and matching the derived name, it's accepted silently. When it mismatches, PI prints a warning.
:::

## Single-step shorthand

Most automations do one thing. Use the step type key directly at the top level:

```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```

This is equivalent to the full `steps:` form:

```yaml
description: Run the test suite
steps:
  - bash: go test ./...
```

PI has no preference — use the short form when there's one step, `steps:` when there are more. All step modifier fields (`env:`, `dir:`, `timeout:`, `silent:`, `pipe:`, `parent_shell:`) work alongside the shorthand key at the top level.

Having both a top-level step key and `steps:` (or `install:`) in the same file is a parse error.

## Multi-step automations

When you need to chain operations, use `steps:`:

```yaml
# .pi/deploy.yaml
description: Build and deploy the application
steps:
  - bash: go build -o bin/app ./...
  - bash: docker build -t myapp .
  - bash: docker push myapp:latest
```

Steps run sequentially. If any step exits non-zero, execution stops immediately and the exit code propagates.

## Step modifiers

Every step supports a set of modifier fields that control how it executes.

### `env:` — Environment variables

Inject environment variables into a step's execution context:

```yaml
steps:
  - bash: go build -o bin/app ./...
    env:
      GOOS: linux
      GOARCH: amd64
      CGO_ENABLED: "0"

  - bash: go test ./...
```

Environment variables are **scoped to the step** — they do not leak to subsequent steps. Step-level env vars override parent process env vars with the same name.

#### Automation-level `env:`

An automation can also declare a top-level `env:` so every step sees the same variables without repeating them:

```yaml
description: Cross-compile for Linux amd64
env:
  GOOS: linux
  GOARCH: amd64
  CGO_ENABLED: "0"

steps:
  - bash: go build -o target/app-linux ./...
  - bash: sha256sum target/app-linux > target/app-linux.sha256
```

Step-level `env:` overrides automation-level `env:` for the same key. Automation-level env does **not** propagate into sub-automations invoked by `run:` steps.

### `dir:` — Working directory

Override the working directory for a step. By default, all steps run in the project root (directory containing `pi.yaml`).

```yaml
steps:
  - bash: go test ./...
    dir: services/api

  - bash: npm install
    dir: frontend
```

The path is resolved relative to the project root. Absolute paths are used as-is. The directory must exist at execution time. Each step independently resolves its own `dir:` — there is no carry-over between steps.

### `timeout:` — Execution time limit

Set a maximum execution duration. If the step exceeds the timeout, PI kills the process and returns exit code 124 (matching the GNU `timeout` convention).

```yaml
steps:
  - bash: go build ./...
    timeout: 30s

  - bash: npm test
    timeout: 5m
```

The value is a Go-style duration string (`30s`, `5m`, `1h30m`). Timeout must be positive. It works with `bash`, `python`, and `typescript` steps but **not** with `run:` steps (set timeouts on the target automation's own steps instead).

### `silent: true` — Suppress output

Suppress both the trace line and the step's stdout/stderr output:

```yaml
steps:
  - bash: echo "Starting build..."
  - bash: rm -rf dist && mkdir dist
    silent: true
  - bash: go build -o dist/app ./...
```

The silent step still executes — only its output is hidden. Piped data (`pipe: true`) still flows through silent steps. Use `--loud` on `pi run` to override all `silent: true` flags for debugging.

### `if:` — Conditional execution

Skip a step based on a runtime condition:

```yaml
steps:
  - bash: brew install jq
    if: os.macos and not command.jq

  - bash: apt-get install -y jq
    if: os.linux and not command.jq

  - bash: echo "jq is ready"
```

When a condition evaluates to false, the step is silently skipped. Supported predicates include `os.macos`, `os.linux`, `command.<name>`, `env.<NAME>`, `file.exists("<path>")`, and more — see the [Conditions reference](/reference/conditions/) for the complete list.

When a step with `pipe: true` is skipped, any piped data passes through to the next step unchanged.

### `pipe: true` — Pipe stdout to next step

Capture a step's stdout and feed it as stdin to the next step:

```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    pipe: true

  - python: logs-formatted.py
```

Stderr is never captured — it always goes to the terminal regardless of piping.

:::note
`pipe_to: next` is the deprecated form. It still works but emits a deprecation warning. Use `pipe: true` instead.
:::

### `description:` — Human-readable label

Document what a step does. Descriptions are shown by `pi info` and have no effect on execution:

```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    description: Stream container logs
    pipe: true

  - python: logs-formatted.py
    description: Format and filter log output
```

### `parent_shell: true` — Run in the calling shell

Run a bash command in the calling shell instead of as a subprocess. Essential for commands that must affect the parent shell's state:

```yaml
steps:
  - bash: source .venv/bin/activate
    parent_shell: true
```

Only valid on `bash` steps. Cannot be combined with `pipe: true`. Requires the PI shell wrapper to be installed (`pi shell`). See [Shell Shortcuts](/concepts/shell-shortcuts/) for the full story.

## Step trace lines

Before executing each step, PI prints a trace line to stderr:

```
  → bash: echo "building project..."
  → run:  setup/install-deps
  → python: transform.py
```

Installer steps are exempt — they have their own formatted output. Steps with `silent: true` suppress the trace line. Use `--loud` to force all trace lines.

## First-match blocks

A `first:` block groups mutually exclusive sub-steps. PI evaluates each sub-step's `if:` condition in order and runs only the first one that matches:

```yaml
steps:
  - first:
      - bash: mise install python@3.13 && mise use python@3.13
        if: command.mise
      - bash: brew install python@3.13
        if: command.brew
      - bash: |
          echo "No supported installer found (tried mise, brew)" >&2
          exit 1
```

A sub-step without `if:` always matches and acts as a fallback. If all sub-steps have `if:` conditions and none match, the block is silently skipped.

This replaces the common pattern of compounding negation chains (`if: command.brew and not command.mise`) with clean, positive conditions.

The `first:` block itself supports `description:`, `if:` (to skip the entire block), and `pipe: true`. Sub-steps support all normal step fields: `env:`, `dir:`, `timeout:`, `silent:`. Nested `first:` blocks are not allowed.

`first:` works anywhere a step appears: in `steps:`, in `install.run:`, in `install.test:`.

## Installer automations

Automations that install tools use `install:` instead of `steps:`. The two are mutually exclusive. The `install:` block declares a test-run-verify lifecycle, and PI manages all status output.

```yaml
# .pi/setup/install-uv.yaml
description: Install uv (Python package manager)
install:
  test: command -v uv >/dev/null 2>&1
  run: curl -LsSf https://astral.sh/uv/install.sh | sh
  version: uv --version | awk '{print $2}'
```

PI prints one formatted status line per installer:

```
  ✓  install-uv      already installed   (0.4.1)
  →  install-node    installing...
  ✓  install-node    installed           (20.11.0)
  ✗  install-go      failed
```

The lifecycle:
1. **`test:`** — Run the test command. Zero exit means already installed → skip `run:`.
2. **`run:`** — Install the tool.
3. **`verify:`** — Verify installation succeeded. When omitted, PI re-runs `test:` as verification.
4. **`version:`** — Capture version string for the status line. Optional.

Each field can be a scalar bash string or a list of steps (same step schema as `steps:`):

```yaml
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

Use `--silent` on `pi run` or `pi setup` to suppress the status lines.

## Automation-level conditions

An automation can declare a top-level `if:` to skip the entire automation based on a condition:

```yaml
description: Install Homebrew (macOS only)
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  version: brew --version | head -1 | awk '{print $2}'
```

When the condition evaluates to false, PI prints `[skipped] <name> (condition: <expr>)` and returns success. This is different from step-level `if:`, which silently skips individual steps.

## Inputs

Automations can declare named parameters with `inputs:`:

```yaml
description: Install Python at a specific version

inputs:
  version:
    type: string
    description: Python version to install
```

At runtime, inputs are injected as `PI_IN_<NAME>` environment variables (uppercased, hyphens become underscores). Pass them with `--with` flags or positional arguments:

```bash
pi run install-python --with version=3.13
pi run install-python 3.13              # positional (mapped by declaration order)
```

For the full input specification, see the [Automation YAML reference](/reference/automation-yaml/).
