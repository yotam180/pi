---
title: Parent Shell Steps
description: Run commands in the calling shell with parent_shell — virtualenvs, cd, exports
---

This guide explains the problem that `parent_shell: true` solves, how it works under the hood, and all the ways to use it effectively.

## What you'll learn

- Why subprocess-based tools can't affect the parent shell
- How `parent_shell: true` solves this
- How the mechanism works under the hood
- How to install shell integration
- Common use cases and patterns
- What doesn't work with parent shell steps

---

## The problem

When PI runs a step, it executes the command as a **subprocess**. Any changes the subprocess makes to its environment — `cd`, `export`, `source` — are invisible to the parent shell. The subprocess exits, and the parent shell is unchanged.

This is a fundamental limitation of all subprocess-based tools, not specific to PI.

Try it yourself:

```yaml
# .pi/go-to-service.yaml
description: Navigate to the service directory
bash: cd services/api
```

```bash
pi run go-to-service
pwd    # still in the original directory
```

The `cd` ran inside a subprocess that exited immediately. Your terminal didn't move.

## The solution

`parent_shell: true` tells PI: "don't execute this step yourself — write it to a file, and after PI exits, have the calling shell run it directly."

```yaml
# .pi/go-to-service.yaml
description: Navigate to the service directory
bash: cd services/api
parent_shell: true
```

```bash
pi run go-to-service
pwd    # now in services/api
```

The `cd` runs in the parent shell, so the terminal actually changes directory.

## How it works

The mechanism relies on three parts:

### 1. The shell wrapper sets `PI_PARENT_EVAL_FILE`

When you run `pi run` via a shell shortcut or the global `pi()` wrapper, the wrapper creates a temp file and passes its path as `PI_PARENT_EVAL_FILE`:

```bash
# What the wrapper does (simplified)
_pi_eval_file="$(mktemp)"
PI_PARENT_EVAL_FILE="$_pi_eval_file" pi run go-to-service
```

### 2. PI writes to the eval file instead of executing

When PI encounters a step with `parent_shell: true`, it appends the command to `PI_PARENT_EVAL_FILE` instead of running it as a subprocess:

```
  → parent: cd services/api
```

### 3. The wrapper sources the eval file

After PI exits, the wrapper sources the temp file:

```bash
if [ -s "$_pi_eval_file" ]; then
  source "$_pi_eval_file"
fi
rm -f "$_pi_eval_file"
```

The command runs in the parent shell, so `cd`, `source`, and `export` actually take effect.

## Installing shell integration

`parent_shell: true` requires the `PI_PARENT_EVAL_FILE` mechanism. This is set up automatically when you install shell integration:

```bash
pi shell
source ~/.zshrc
```

After this, `parent_shell: true` works in two ways:

1. **Via shell shortcuts** — each shortcut function wraps `pi run` with the eval pattern
2. **Via the global `pi()` wrapper** — wraps every `pi` invocation, so `pi run` always supports parent shell steps

:::caution
If you call the raw `pi` binary directly (without the shell wrapper), `PI_PARENT_EVAL_FILE` is not set. Parent shell steps are skipped with a warning:

```
⚠  parent_shell step skipped: not running inside a PI shell wrapper.
   Run 'pi shell' to install shell integration.
```

This is non-fatal — other steps continue to run normally.
:::

## Common use cases

### Activate a Python virtualenv

```yaml
# .pi/activate.yaml
description: Create and activate the Python virtualenv

steps:
  - bash: python3 -m venv .venv
    if: not dir.exists(".venv")

  - bash: source .venv/bin/activate
    parent_shell: true
```

After running `pi run activate`, your terminal is inside the virtualenv.

### Change directory

```yaml
# .pi/go-to-api.yaml
description: Navigate to the API service
bash: cd services/api
parent_shell: true
```

### Load environment variables

```yaml
# .pi/load-env.yaml
description: Load environment variables from .env
bash: source .env
parent_shell: true
```

### Switch Node versions with nvm

```yaml
# .pi/use-node-20.yaml
description: Switch to Node 20
bash: nvm use 20
parent_shell: true
```

### Combine normal and parent shell steps

Parent shell steps mix freely with regular steps:

```yaml
# .pi/setup-dev.yaml
description: Set up development environment

steps:
  - bash: python3 -m venv .venv
    description: Create virtualenv

  - bash: source .venv/bin/activate
    parent_shell: true
    description: Activate virtualenv in parent shell

  - bash: pip install -r requirements.txt
    description: Install dependencies (runs in subprocess)
```

The regular steps run as subprocesses (as normal). The parent shell step is deferred to the eval file. After PI exits, the `source` command runs in the parent shell.

:::note
The regular steps (like `pip install`) run **before** the parent shell steps take effect. Parent shell steps only execute after PI exits. This means you can't install packages into a virtualenv that hasn't been activated yet via `parent_shell: true` — the activation happens after PI.
:::

## Multiple parent shell steps

When an automation has multiple `parent_shell: true` steps, all of them are appended to the eval file and execute in order after PI exits:

```yaml
steps:
  - bash: source .venv/bin/activate
    parent_shell: true

  - bash: export API_URL="http://localhost:8080"
    parent_shell: true

  - bash: cd services/api
    parent_shell: true
```

After PI exits, the parent shell runs all three commands: activate the virtualenv, set the env var, then change directory.

## Rules and restrictions

| Rule | Details |
|------|---------|
| Only `bash:` steps | `parent_shell: true` is invalid on `python:`, `typescript:`, and `run:` steps |
| No piping | Cannot combine with `pipe: true` |
| Requires shell wrapper | `PI_PARENT_EVAL_FILE` must be set (via `pi shell` or shortcuts) |
| Fire-and-forget | No output capture — parent shell steps don't produce stdout for piping |
| Condition-aware | `if:` is evaluated before the parent shell check — skipped steps don't write to the eval file |

## What doesn't work

**Output capture:** Parent shell steps don't produce output that PI can capture. You can't pipe from a parent shell step to the next step.

**Immediate effect:** Parent shell steps don't take effect until after PI exits. If a subsequent regular step depends on the parent shell state (like an activated virtualenv), it won't see it.

**Raw binary calls:** Calling `pi` directly (not via the shell wrapper) means `PI_PARENT_EVAL_FILE` is not set. Parent shell steps are skipped with a warning.

## Summary

- `parent_shell: true` runs bash commands in the calling shell instead of a subprocess
- Essential for `cd`, `source`, `export`, and virtualenv activation
- Works automatically via shell shortcuts or the global `pi()` wrapper
- Install shell integration with `pi shell` to enable the mechanism
- Parent shell steps execute after PI exits — they can't affect subsequent regular steps
- Only valid on `bash:` steps; cannot combine with `pipe: true`
