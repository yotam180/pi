---
title: Automation YAML
description: Complete reference for the automation YAML file format — every field, every type
---

Automations are YAML files in your project's `.pi/` folder. Each file defines a named, runnable unit of work. The automation name is derived from the file path — `.pi/docker/up.yaml` becomes `docker/up`. For an overview of how automations fit into the PI model, see [Automations](/concepts/automations/).

## Naming

| File path | Automation name |
|-----------|----------------|
| `.pi/docker/up.yaml` | `docker/up` |
| `.pi/setup/cursor/automation.yaml` | `setup/cursor` |
| `.pi/test.yaml` | `test` |

When a folder contains an `automation.yaml` file, the automation name is the folder path (without the filename).

:::note
The `name:` field is deprecated. PI derives the name from the file path. If `name:` is present and matches the derived name, it's accepted silently. If it mismatches, PI prints a warning. Existing files with `name:` continue to work.
:::

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | no | Human-readable description shown by `pi info` and `pi list`. |
| `steps` | list | one of | List of steps to execute in sequence. |
| `bash` | string | one of | [Single-step shorthand](#single-step-shorthand): inline shell or path to `.sh` file. |
| `python` | string | one of | Single-step shorthand: inline Python or path to `.py` file. |
| `typescript` | string | one of | Single-step shorthand: inline TypeScript or path to `.ts` file (via tsx). |
| `run` | string | one of | Single-step shorthand: call another automation by name. |
| `install` | object | one of | [Installer block](#install-block). Mutually exclusive with `steps` and step shorthands. |
| `env` | map | no | [Automation-level environment variables](#automation-level-env) applied to all steps. |
| `if` | string | no | [Condition expression](/reference/conditions/). If false, the entire automation is skipped. |
| `inputs` | map | no | [Input declarations](#inputs). |

Exactly one of `steps`, `bash`, `python`, `typescript`, `run`, or `install` must be present. Having both a top-level step key and `steps` (or `install`) is a parse error.

---

## Steps

Each item in the `steps` list is a step. A step must have exactly one step type key.

### Step Type Keys

| Key | Description |
|-----|-------------|
| `bash` | Inline shell command or path to a `.sh` file. |
| `python` | Inline Python script or path to a `.py` file. |
| `typescript` | Inline TypeScript or path to a `.ts` file (run via `tsx`). |
| `run` | Call another automation by name (local, [built-in](/reference/builtins/), [package](/concepts/packages/), or GitHub reference). |
| `first` | [First-match block](#first-match-blocks) — list of conditional sub-steps, only the first match executes. |

For details on how each step type works (inline vs. file, argument passing, virtual environments), see [Step Types](/concepts/step-types/).

### Step Modifier Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | — | Human-readable label. Shown by `pi info`. No effect on execution. |
| `env` | map | — | Step-level environment variables. Scoped to this step only; overrides automation-level and parent process env. See [env](#step-level-env). |
| `dir` | string | project root | Working directory override. Resolved relative to project root; absolute paths used as-is. Must exist at execution time. See [dir](#step-working-directory). |
| `timeout` | string | — | Maximum execution time as a Go duration (`30s`, `5m`, `1h30m`). Process killed with exit code 124 on timeout. Not valid on `run` or `parent_shell` steps. See [timeout](#step-timeout). |
| `silent` | bool | `false` | Suppress trace line and stdout/stderr output. Overridden by `--loud`. Piped data still flows. See [silent](#silent-steps). |
| `if` | string | — | [Condition expression](/reference/conditions/). If false, step is silently skipped. |
| `pipe` | bool | `false` | Pipe this step's stdout to the next step's stdin. Stderr always goes to terminal. See [pipe](#pipe). |
| `parent_shell` | bool | `false` | Run this bash command in the calling shell instead of as a subprocess. Only valid on `bash` steps. Cannot combine with `pipe`. See [parent_shell](#parent-shell). |
| `with` | map | — | Named inputs to pass when the step type is `run`. |

---

## Single-Step Shorthand

Automations with a single step can place the step type key at the top level, skipping the `steps` wrapper:

```yaml
# Short form
description: Run the test suite
bash: go test ./...
```

```yaml
# Equivalent full form
description: Run the test suite
steps:
  - bash: go test ./...
```

All step types work as shorthands: `bash`, `python`, `typescript`, `run`. Step modifier fields (`env`, `dir`, `timeout`, `silent`, `pipe`, `parent_shell`, `with`) can be used alongside the shorthand key at the top level.

```yaml
description: Cross-compile for Linux
bash: go build -o bin/app ./...
env:
  GOOS: linux
  GOARCH: amd64
dir: services/api
timeout: 30s
```

Top-level `if` maps to the automation-level condition. Top-level `description` remains the automation description.

---

## Step-Level `env` {#step-level-env}

Steps can inject environment variables into their execution context:

```yaml
steps:
  - bash: go build -o bin/app ./...
    env:
      GOOS: linux
      GOARCH: amd64
      CGO_ENABLED: "0"
```

- Scoped to the step — does not leak to subsequent steps
- Overrides both automation-level `env` and parent process variables with the same name
- Works with `bash`, `python`, and `typescript` steps

## Automation-Level `env` {#automation-level-env}

Automations can declare a top-level `env` so every step sees the same variables:

```yaml
description: Cross-compile for Linux
env:
  GOOS: linux
  GOARCH: amd64

steps:
  - bash: go build -o target/app ./...
  - bash: sha256sum target/app > target/app.sha256
```

- Step-level `env` overrides automation-level `env` for the same key
- Does not propagate into sub-automations invoked by `run` steps

---

## Step Working Directory (`dir`) {#step-working-directory}

```yaml
steps:
  - bash: go test ./...
    dir: services/api

  - bash: npm install
    dir: frontend
```

- Resolved relative to the project root (directory containing `pi.yaml`)
- Absolute paths are used as-is
- Must exist at execution time — PI reports an error if it doesn't
- Per-step — no carry-over between steps; steps without `dir` use the project root

---

## Step Timeout (`timeout`) {#step-timeout}

```yaml
steps:
  - bash: go build ./...
    timeout: 30s

  - python: train_model.py
    timeout: 1h30m
```

- Value is a Go duration string: `30s`, `5m`, `1h30m`
- Must be positive — zero or negative values rejected at parse time
- On timeout, the process is killed and exit code 124 is returned (matching GNU `timeout` convention)
- Works with `bash`, `python`, `typescript` steps
- Not valid on `run` steps (set timeouts on the target automation's steps instead) or `parent_shell` steps

---

## Silent Steps {#silent-steps}

```yaml
steps:
  - bash: rm -rf dist && mkdir dist
    silent: true
  - bash: go build -o dist/app ./...
```

- Suppresses the step's trace line and stdout/stderr output
- Step still executes — only output is hidden
- Piped data (`pipe: true`) still flows through silent steps
- `--loud` flag on `pi run` / `pi setup` overrides all `silent: true` flags

---

## Pipe (`pipe`) {#pipe}

```yaml
steps:
  - bash: docker-compose logs -f --tail 200
    pipe: true

  - python: format-logs.py
```

- Captures stdout into a buffer and feeds it as stdin to the next step
- Stderr always goes to the terminal regardless of piping
- If `pipe: true` is on the last step, output goes to terminal normally
- Works across all step types
- When a piped step with `if` is skipped, piped data passes through unchanged to the next step

:::note
`pipe_to: next` is the legacy form. It still works but emits a deprecation warning. Use `pipe: true` instead.
:::

---

## Parent Shell (`parent_shell`) {#parent-shell}

```yaml
steps:
  - bash: source .venv/bin/activate
    parent_shell: true
```

- Only valid on `bash` steps
- Cannot combine with `pipe: true`
- PI does not execute the step — instead it writes the command to `PI_PARENT_EVAL_FILE`
- After PI exits, the shell wrapper `eval`s the file, running the command in the calling shell
- Requires `PI_PARENT_EVAL_FILE` to be set (automatic when using shell shortcuts or the global `pi()` wrapper installed by `pi shell`)
- If `PI_PARENT_EVAL_FILE` is not set, the step is skipped with a warning

See the [Parent Shell Steps guide](/guides/parent-shell-steps/) for detailed usage.

---

## First-Match Blocks (`first`) {#first-match-blocks}

A `first` block groups mutually exclusive sub-steps. The executor evaluates each sub-step's `if` condition in order and runs only the first match:

```yaml
steps:
  - first:
      - bash: mise install go@1.23 && mise use go@1.23
        if: command.mise
      - bash: brew install go
        if: command.brew
      - bash: |
          echo "no installer found" >&2
          exit 1
```

**Block-level fields** (on the `first` step itself):

| Field | Valid | Description |
|-------|-------|-------------|
| `description` | yes | Human-readable label for the block. |
| `if` | yes | Skip the entire block if false. |
| `pipe` | yes | Pipe the matched sub-step's output to the next step. |
| `env`, `dir`, `timeout`, `silent`, `parent_shell`, `with` | **no** | Set these on individual sub-steps instead. |

**Sub-step fields:** Each sub-step supports all normal step fields: `bash`, `python`, `typescript`, `run`, `env`, `dir`, `timeout`, `silent`, `if`.

- A sub-step without `if` always matches (acts as a fallback)
- If no sub-step matches, the block is silently skipped
- Nested `first` blocks are not allowed
- Works anywhere a step appears: `steps`, `install.run`, `install.test`

---

## `install` Block {#install-block}

Automations that install tools use `install` instead of `steps`. PI manages the test-run-verify lifecycle and all status output.

```yaml
description: Install Homebrew
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  version: brew --version | head -1 | awk '{print $2}'
```

### `install` Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `test` | string or list | yes | Command(s) to check if already installed. Exit 0 = installed. |
| `run` | string or list | yes | Command(s) to install the tool. |
| `verify` | string or list | no | Command(s) to verify after install. Defaults to re-running `test`. |
| `version` | string | no | Command to retrieve the installed version string for status output. |

When `test`, `run`, or `verify` is a list, each item is a step (same schema as `steps` items, supporting `if`, `first`, etc.).

### Status Output

```
✓  install-homebrew      already installed   (4.2.1)
→  install-python        installing...
✓  install-python        installed           (3.13.0)
✗  install-node          failed
```

Use `--silent` on `pi run` or `pi setup` to suppress status lines.

---

## Inputs {#inputs}

Automations can declare named parameters:

```yaml
inputs:
  version:
    type: string
    description: Python version to install
  debug:
    type: string
    required: false
    default: "false"
```

### Input Spec Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | — | Input type (currently only `string`). |
| `description` | string | — | Human-readable description shown by `pi info`. |
| `required` | bool | `true` (when no default) | Whether the input must be provided. |
| `default` | string | — | Default value when not provided. |

### Passing Inputs

```bash
# Positional (mapped in declaration order)
pi run setup/install-python 3.13

# Named
pi run setup/install-python --with version=3.13
```

### Environment Variables

Resolved inputs are injected as environment variables available to all steps:

| Variable | Description |
|----------|-------------|
| `PI_IN_<NAME>` | Canonical form. Name is uppercased, hyphens become underscores. |
| `PI_INPUT_<NAME>` | Deprecated form. Set alongside `PI_IN_*` for backward compatibility. |

Example: input `version` → `PI_IN_VERSION` and `PI_INPUT_VERSION`.

### `with` on `run` Steps

`run` steps can pass inputs to the called automation:

```yaml
steps:
  - run: pi:install-python
    with:
      version: "3.13"
```

### `with` on Shortcuts

Shortcuts in `pi.yaml` can use `$1`, `$2` positional references:

```yaml
shortcuts:
  install-py:
    run: pi:install-python
    with:
      version: "$1"
```
