---
title: CLI Commands
description: Complete reference for every PI CLI command and flag
---

## pi run

Run an automation by name.

```
pi run <name> [args...]
pi run <name> --with key=value [--with key=value ...]
pi run --repo <path> <name>
pi run --silent <name>
pi run --loud <name>
```

| Flag | Description |
|------|-------------|
| `--with key=value` | Pass a named input to the automation. Repeatable. |
| `--repo <path>` | Explicitly set the project root directory. |
| `--silent` | Suppress PI status lines for installer automations. |
| `--loud` | Override all `silent: true` steps; force trace line and output. |

Positional arguments after the automation name are mapped to inputs in declaration order. Use `--with` for explicit named inputs. Mixing positional and `--with` is an error.

```bash
# Positional input
pi run setup/install-python 3.13

# Named input
pi run setup/install-python --with version=3.13

# Force all step output
pi run --loud build
```

---

## pi info

Show detailed information about an automation.

```
pi info <name>
```

Displays: automation name, description, `if:` condition (when present), automation-level `env:` variables, step count, step details with annotations (`if:`, `env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`, `description:`), `first:` block details, installer lifecycle, and input specifications with `$PI_IN_*` environment variable names.

```bash
pi info docker/up
pi info pi:install-python
```

---

## pi setup

Run all automations listed in `pi.yaml → setup:`, then install shell shortcuts.

```
pi setup
pi setup --no-shell
pi setup --silent
pi setup --loud
```

| Flag | Description |
|------|-------------|
| `--no-shell` | Run setup automations without installing shell shortcuts. |
| `--silent` | Suppress PI status lines for installer automations. |
| `--loud` | Override all `silent: true` steps; force trace line and output. |

Before running setup automations, PI fetches all declared [packages](/concepts/packages/) that aren't already cached. Setup entries support [`if:` conditions](/reference/conditions/) — entries whose condition evaluates to false are skipped with a message.

After setup completes, `pi shell` runs automatically (unless `--no-shell` is passed or a CI environment is detected). If `PI_PARENT_EVAL_FILE` is set, shortcuts are available immediately in the current terminal via auto-sourcing.

CI detection: skipped when `CI`, `GITHUB_ACTIONS`, or `GITLAB_CI` environment variables are set.

```bash
pi setup               # full setup + shell shortcuts
pi setup --no-shell    # setup only, no shortcuts
pi setup --loud        # show all step output for debugging
```

---

## pi shell

Install, remove, or list shell shortcut functions.

```
pi shell
pi shell uninstall
pi shell list
```

| Subcommand | Description |
|------------|-------------|
| *(none)* | Install shortcut functions for the current project into `~/.pi/shell/<project>.sh`. Injects a source line into `.zshrc`/`.bashrc`. See [Shell Shortcuts](/concepts/shell-shortcuts/). |
| `uninstall` | Remove shortcuts for the current project. Cleans up the source line if no projects remain. |
| `list` | List all installed shortcut files across all projects. |

`pi shell` also installs:
- A `pi-setup-<project>` helper function that wraps `pi setup` with the eval pattern
- A global `pi()` wrapper function (`~/.pi/shell/_pi-wrapper.sh`) that enables [`parent_shell: true`](/guides/parent-shell-steps/) for all `pi run` calls
- A `_pi-completion.sh` file that sets up tab completion for `pi run` and `pi info`

```bash
pi shell             # install shortcuts
pi shell uninstall   # remove shortcuts
pi shell list        # see what's installed
```

---

## pi list

List all available automations.

```
pi list
pi list --all
pi list --builtins
```

| Flag | Description |
|------|-------------|
| `--all`, `-a` | Include all automations from declared packages, grouped by source with separator headers. |
| `--builtins`, `-b` | Include built-in (`pi:*`) automations in the list. |

Output columns: NAME, SOURCE, DESCRIPTION, INPUTS. The SOURCE column shows `[workspace]` for local automations, `[built-in]` for builtins, and the alias or source string for package automations. INPUTS shows required inputs as `name` and optional as `name?`.

```bash
pi list              # local + package automations
pi list --all        # grouped by package source
pi list -b           # include pi:* built-ins
```

---

## pi doctor

Check requirement health for all automations.

```
pi doctor
```

Scans all automations (local + built-in) and checks each `requires:` entry. Displays a per-automation health table with `✓`/`✗` icons, detected versions, and install hints for missing tools. Automations without `requires:` are silently skipped.

Exit code 0 when all requirements are satisfied, 1 when any are missing.

```bash
pi doctor
```

```
docker/up
  ✓  docker                 (24.0.5)
  ✗  docker-compose         not found
     install: pip install docker-compose
```

---

## pi validate

Statically validate all configuration and automation files.

```
pi validate
```

Validation layers:
1. `pi.yaml` schema validation
2. Automation YAML parsing (all files in `.pi/`)
3. Cross-reference checks: shortcut targets, setup entry targets, and `run:` step values are verified against discovered automation names

Reports all errors (not just the first). Exit code 0 on success, 1 on validation errors. No network requests, no command execution — purely static analysis.

```bash
pi validate
```

```
✗ shortcut "vpup" references unknown automation "docker/upp"
✗ setup entry "setup/missing" not found

✓ Validated 12 automation(s), 3 shortcut(s), 2 setup entry(ies)
```

---

## pi add

Declare a package dependency in `pi.yaml`.

```
pi add <source>
pi add <source> --as <alias>
```

| Flag | Description |
|------|-------------|
| `--as <alias>` | Set an alias for the package. Enables `run: <alias>/automation-name`. Alias must not contain `/`. |

Accepts two source types:

| Source type | Format | Example |
|-------------|--------|---------|
| GitHub | `org/repo@version` | `pi add yotam180/pi-common@v1.2` |
| File | `file:~/path` or `file:./relative` | `pi add file:~/my-automations --as mytools` |

GitHub sources are fetched into `~/.pi/cache/` immediately. GitHub sources without `@version` are rejected. Adding the same source twice is a no-op ("already in pi.yaml").

```bash
# Add a GitHub package
pi add yotam180/pi-common@v1.2

# Add a local folder with an alias
pi add file:~/shared-automations --as shared

# Duplicate is a no-op
pi add yotam180/pi-common@v1.2
# → already in pi.yaml
```

---

## pi version

Print the PI version string.

```
pi version
pi --version
```

Both forms produce the same output. Dev builds print `dev`.

---

## pi completion

Generate a shell completion script.

```
pi completion <shell>
```

Supported shells: `bash`, `zsh`, `fish`, `powershell`.

The generated script provides tab completion for all PI commands and dynamic completion for automation names on `pi run` and `pi info`. Built-in automations are excluded from completion results.

```bash
# Add to ~/.zshrc
source <(pi completion zsh)

# Add to ~/.bashrc
source <(pi completion bash)

# Write to a file
pi completion fish > ~/.config/fish/completions/pi.fish
```

:::note
`pi shell` automatically installs completion scripts — you typically don't need to run `pi completion` manually.
:::
