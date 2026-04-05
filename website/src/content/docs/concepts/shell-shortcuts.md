---
title: Shell Shortcuts
description: How pi shell works — generated functions, the global wrapper, and parent_shell steps
---

`pi shell` turns automations into terminal commands. After running it, shortcut keys defined in `pi.yaml` become shell functions you can type from any directory.

## In this section

- [Installing shortcuts](#installing-shortcuts) — what `pi shell` does
- [Running shortcuts](#running-shortcuts) — how they work
- [`anywhere: true`](#anywhere-true) — skip the `cd` to project root
- [The global `pi()` wrapper](#the-global-pi-wrapper) — making `parent_shell` work everywhere
- [`pi-setup-<project>`](#pi-setup-project) — the setup helper function
- [The parent shell pattern](#the-parent-shell-pattern) — how `parent_shell: true` works
- [Managing shortcuts](#managing-shortcuts) — uninstall and list

---

## Installing shortcuts

Running `pi shell` does two things:

1. Writes a shell file to `~/.pi/shell/<project>.sh` containing shell functions for each shortcut defined in `pi.yaml`
2. Injects a source block into `.zshrc` (and `.bashrc` if it exists) that loads all files in `~/.pi/shell/`

After running `pi shell`, source your shell config (or open a new terminal):

```bash
pi shell
source ~/.zshrc
```

Each shortcut is a **shell function**, not an alias. The generated function:
- Creates a temp file for `PI_PARENT_EVAL_FILE`
- `cd`s to the project root
- Runs `pi run <automation>` with the temp file path
- Sources the temp file if non-empty (for `parent_shell` steps)
- Cleans up and preserves the exit code

## Running shortcuts

Once installed, shortcuts work from any directory:

```bash
cd ~/some/random/folder
vpup                      # runs pi run docker/up from the project root
```

The shortcut `cd`s to the project root before running, so automations always execute in the expected directory. Arguments are forwarded:

```bash
vpup --force-recreate     # passes --force-recreate to the automation
```

## `anywhere: true`

By default, shortcuts `cd` to the project root. With `anywhere: true`, the shortcut runs from your current directory:

```yaml
shortcuts:
  deploy:
    run: deploy/push-image
    anywhere: true
```

Use this for automations that operate on the current directory or don't depend on being at the project root — like deploy commands, global utilities, or tools that take a path argument.

When `anywhere: true` is set, the generated function uses `pi run --repo <root>` instead of `cd <root>`.

## The global `pi()` wrapper

In addition to project shortcuts, `pi shell` installs a global `pi()` shell function in `~/.pi/shell/_pi-wrapper.sh`. This wrapper wraps **every** `pi` invocation with the `PI_PARENT_EVAL_FILE` pattern:

```bash
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

This means `parent_shell: true` steps work for **all** `pi run` calls — not just shortcuts. Without this wrapper, calling the raw `pi` binary directly can't affect the parent shell.

The wrapper uses `command pi` to call the real binary and avoid infinite recursion.

## `pi-setup-<project>`

`pi shell` generates a helper function that wraps `pi setup` with the eval pattern:

```bash
pi-setup-myproject
```

This function:
1. Runs `pi setup` (installs dependencies, runs setup automations, installs shortcuts)
2. Automatically sources `~/.zshrc` after setup completes

The auto-sourcing means shortcuts are available immediately — no need to open a new terminal or manually source the rc file.

:::tip[First-time bootstrapping]
On the very first `pi setup` (before any shell wrapper exists), auto-sourcing doesn't work. Run `source ~/.zshrc` once after the first setup. After that, `pi-setup-<project>` handles everything automatically.
:::

## The parent shell pattern

When PI runs a step as a subprocess, environment changes (like `cd`, `export`, or `source venv/bin/activate`) are invisible to the parent shell — the subprocess exits, and the parent is unchanged. This is a fundamental limitation of subprocess-based tools.

`parent_shell: true` solves this:

```yaml
steps:
  - bash: python3 -m venv .venv
  - bash: source .venv/bin/activate
    parent_shell: true
```

Instead of executing the step, PI writes the command to `PI_PARENT_EVAL_FILE`. After PI exits, the shell wrapper sources that file, running the command in the parent shell.

**Requirements:**

- Only valid on `bash:` steps
- Cannot be combined with `pipe: true`
- Requires `PI_PARENT_EVAL_FILE` to be set (happens automatically via shell shortcuts or the global `pi()` wrapper)
- If `PI_PARENT_EVAL_FILE` is not set (raw binary call without the wrapper), the step is skipped with a warning

**Common use cases:**

```yaml
- bash: source .venv/bin/activate
  parent_shell: true

- bash: cd /path/to/service
  parent_shell: true

- bash: export DATABASE_URL="postgres://..."
  parent_shell: true

- bash: nvm use 20
  parent_shell: true
```

For a detailed walkthrough, see the [Parent Shell Steps guide](/guides/parent-shell-steps/).

## Managing shortcuts

### `pi shell uninstall`

Remove shortcuts for the current project:

```bash
pi shell uninstall
```

This removes the project's shell file from `~/.pi/shell/`. If no projects remain, the source block is cleaned from `.zshrc` and the global `pi()` wrapper is removed.

### `pi shell list`

List all installed shortcut files across all projects:

```bash
pi shell list
```

### Shell completion

`pi shell` also installs a shell completion file (`~/.pi/shell/_pi-completion.sh`) that enables tab completion for `pi run` and `pi info`:

```bash
pi run <TAB>    # completes automation names
pi info <TAB>   # completes automation names
```

You can also generate a completion script manually with `pi completion <shell>` (supports bash, zsh, fish, and powershell).
