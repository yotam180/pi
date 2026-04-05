# Website: Reference Documentation

## Type
feature

## Status
todo

## Priority
high

## Project
14-documentation-website

## Description
Write the five Reference pages. Reference pages are exhaustive lookup tables — every field, every flag, every valid value, with a one-line description each and a short example where needed. A developer should be able to answer "what flags does `pi run` accept?" or "what is the exact syntax for a `first:` block?" by opening Reference, not by reading Concepts again.

Source of truth: `docs/README.md`. Every item listed there must appear in the Reference. No items should be omitted.

## Acceptance Criteria
- [ ] `cli.md` documents every command listed in the CLI Reference table in `docs/README.md`, with every flag for each command
- [ ] `automation-yaml.md` documents every top-level automation field, every step type, and every step modifier — in a consistent format (field name, type, required/optional, description, example)
- [ ] `conditions.md` documents every supported predicate for `if:` expressions, with the exact syntax and a short example
- [ ] `builtins.md` documents every `pi:*` built-in automation with its description and any `inputs:` it takes
- [ ] `pi-package-yaml.md` documents the `pi-package.yaml` file format and its fields
- [ ] All pages use a consistent format (table or definition list, not prose paragraphs)
- [ ] Each reference entry has a short example where syntax is non-obvious
- [ ] Cross-links to Concepts pages for deeper context

## Implementation Notes

### cli.md — CLI Commands Reference

Use a consistent entry format for each command:
```
## pi run

Run an automation by name.

pi run <name> [args]
pi run <name> --with key=value [--with key=value ...]
pi run --repo <path> <name>
pi run --silent <name>
pi run --loud <name>

| Flag | Description |
|------|-------------|
| `--with key=value` | Pass a named input to the automation. Repeatable. |
| `--repo <path>` | Explicitly set the project root directory. |
| `--silent` | Suppress PI status lines for installer automations. |
| `--loud` | Override all `silent: true` steps; force trace line and output. |
```

Commands to document (from `docs/README.md`):
- `pi run` (with all flags)
- `pi info`
- `pi setup` (with `--no-shell`, `--silent`, `--loud`)
- `pi shell` (with `uninstall` and `list` subcommands)
- `pi list` (with `--all`, `--builtins`)
- `pi doctor`
- `pi validate`
- `pi add` (with `--as`)
- `pi version` / `pi --version`
- `pi completion` (bash, zsh, fish, powershell)

For each command, include:
- Usage line(s)
- Flags table
- Brief description of what it does
- A short example (one code block)

---

### automation-yaml.md — Automation YAML Spec

Structure this page in sections that mirror the automation structure.

#### Top-level fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | no | Human-readable description shown by `pi info` and step traces |
| `steps` | list | one of | List of steps to run in sequence. Mutually exclusive with top-level step key and `install:`. |
| `bash` | string | one of | Single-step shorthand: inline shell command or path to `.sh` file |
| `python` | string | one of | Single-step shorthand: inline Python or path to `.py` file |
| `typescript` | string | one of | Single-step shorthand: inline TypeScript or path to `.ts` file (via tsx) |
| `run` | string | one of | Single-step shorthand: automation name or GitHub reference |
| `install` | object | one of | Installer block. See `install:` section below. |
| `env` | map | no | Automation-level environment variables applied to all steps. Step-level `env:` overrides. |
| `if` | string | no | Condition expression. If false, the entire automation is skipped. |
| `inputs` | map | no | Input declarations. See Inputs section. |

Note: "one of" means exactly one of `steps`, `bash`, `python`, `typescript`, `run`, `install` must be present. Having both a top-level step key and `steps:` is a parse error.

#### Step fields (for items in `steps:`)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `bash` | string | one of step types | Inline shell or `.sh` file path |
| `python` | string | one of step types | Inline Python or `.py` file path |
| `typescript` | string | one of step types | Inline TypeScript or `.ts` file path (via tsx) |
| `run` | string | one of step types | Automation name or GitHub reference |
| `first` | list | one of step types | First-match block: list of sub-steps with `if:` conditions |
| `description` | string | no | Human-readable label for this step. Shown by `pi info`. |
| `env` | map | no | Step-level environment variables. Scoped to this step only. |
| `dir` | string | no | Working directory override. Resolved relative to project root. |
| `timeout` | string | no | Max execution time (e.g., `30s`, `5m`, `1h30m`). Exit code 124 on timeout. Not valid on `run:` or `parent_shell:` steps. |
| `silent` | bool | no | Suppress trace line and output. Overridden by `--loud` flag. |
| `if` | string | no | Condition expression. If false, step is silently skipped. |
| `pipe` | bool | no | Pipe this step's stdout to the next step's stdin. |
| `parent_shell` | bool | no | Run this bash step in the calling shell. Only valid on `bash:` steps. Cannot combine with `pipe:`. |

#### `install:` block fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `test` | string or list | yes | Command or steps to check if already installed. Zero exit = installed. |
| `run` | string or list | yes | Command or steps to install the tool. |
| `version` | string | no | Command to retrieve the installed version string for status output. |

#### `first:` block

A `first:` block is a list of sub-steps. Each sub-step supports: `bash`, `python`, `typescript`, `run`, `env`, `dir`, `timeout`, `silent`, `if`. The block itself supports: `description`, `if`, `pipe`. Nested `first:` blocks are not allowed.

#### Inputs

```yaml
inputs:
  version:
    type: string
    description: The version to install
```

At runtime, inputs are injected as `PI_IN_<NAME>` environment variables. If passed via `pi run --with version=3.13`, then `PI_IN_VERSION=3.13` is set for all steps.

---

### conditions.md — Conditions Reference (if:)

Intro: conditions are boolean expressions used in `if:` fields on steps, automations, `setup:` entries, and `first:` sub-steps. They are evaluated at runtime.

#### Predicates table

| Predicate | True when... | Example |
|-----------|-------------|---------|
| `os.macos` | Running on macOS | `if: os.macos` |
| `os.linux` | Running on Linux | `if: os.linux` |
| `os.windows` | Running on Windows | `if: os.windows` |
| `os.arch.arm64` | CPU is ARM64 (Apple Silicon, etc.) | `if: os.arch.arm64` |
| `os.arch.amd64` | CPU is AMD64 (x86-64) | `if: os.arch.amd64` |
| `command.<name>` | Command is available in PATH | `if: command.brew` |
| `env.<NAME>` | Environment variable is set (non-empty) | `if: env.CI` |
| `file.exists("<path>")` | File exists at path (relative to project root) | `if: file.exists("Makefile")` |
| `dir.exists("<path>")` | Directory exists at path | `if: dir.exists(".venv")` |
| `shell.zsh` | Current shell is zsh | `if: shell.zsh` |
| `shell.bash` | Current shell is bash | `if: shell.bash` |

#### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `and` | Both sides must be true | `if: os.macos and command.brew` |
| `or` | Either side must be true | `if: os.macos or os.linux` |
| `not` | Negates the predicate | `if: not command.jq` |
| `(...)` | Grouping | `if: os.macos and (command.brew or command.curl)` |

#### Behavior notes
- When a step's `if:` is false, the step is silently skipped (no trace line, no error)
- When a `pipe: true` step is skipped, piped data passes through unchanged to the next step
- When an automation's top-level `if:` is false, the whole automation is skipped with a message

---

### builtins.md — Built-in Automations

Intro: PI ships with a standard library of automations for common tasks. Reference them with the `pi:` prefix in `run:` steps or `setup:` entries.

For each built-in, document: name, description, inputs (if any), example usage.

| Automation | Description | Inputs |
|------------|-------------|--------|
| `pi:install-python` | Check/install Python at a specific version | `version` (string) |
| `pi:install-node` | Check/install Node.js | `version` (string, optional) |
| `pi:install-go` | Check/install Go at a specific version | `version` (string) |
| `pi:install-rust` | Check/install Rust | none |
| `pi:install-ruby` | Check/install Ruby | `version` (string, optional) |
| `pi:install-uv` | Check/install uv | none |
| `pi:install-homebrew` | Check/install Homebrew (macOS only) | none |
| `pi:install-tsx` | Check/install tsx globally via npm | none |
| `pi:docker/up` | Run `docker-compose up -d` | none |
| `pi:docker/down` | Run `docker-compose down` | none |
| `pi:docker/logs` | Stream `docker-compose logs -f` | none |
| `pi:cursor/install-extensions` | Install missing Cursor extensions from a list | none |
| `pi:git/install-hooks` | Install git hooks from a source directory | none |

Show a canonical usage example:
```yaml
setup:
  - run: pi:install-python
    with:
      version: "3.13"
  - run: pi:install-node
    with:
      version: "20"
  - pi:install-homebrew
```

---

### pi-package-yaml.md

Short page. The file is optional and minimal.

Fields:

| Field | Type | Description |
|-------|------|-------------|
| `min_pi_version` | string | Minimum PI version required to use this package. Checked at fetch time. Dev builds skip this check. |

Example:
```yaml
min_pi_version: "0.5.0"
```

Behavior:
- If absent or empty, the package works with any PI version.
- If present and the running PI is older than `min_pi_version`, the fetch fails with a clear upgrade message.
- Dev builds (`version == "dev"`) skip the check.

## Subtasks
- [ ] Write `cli.md` with all commands and flags
- [ ] Write `automation-yaml.md` with all fields
- [ ] Write `conditions.md` with all predicates and operators
- [ ] Write `builtins.md` with all pi:* automations
- [ ] Write `pi-package-yaml.md`
- [ ] Verify completeness against `docs/README.md`
- [ ] Add cross-links to Concepts pages

## Blocked By
79-website-scaffold-and-ci
