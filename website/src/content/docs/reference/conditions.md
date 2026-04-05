---
title: "Conditions (if:)"
description: Every supported predicate and operator for conditional execution
---

Conditions are boolean expressions used in `if:` fields on [steps](/reference/automation-yaml/#step-modifier-fields), [automations](/reference/automation-yaml/#top-level-fields), [`setup:` entries](/concepts/pi-yaml/#setup), and [`first:` sub-steps](/reference/automation-yaml/#first-match-blocks). They are evaluated at runtime.

## Predicates

| Predicate | True when… | Example |
|-----------|-----------|---------|
| `os.macos` | Running on macOS | `if: os.macos` |
| `os.linux` | Running on Linux | `if: os.linux` |
| `os.windows` | Running on Windows | `if: os.windows` |
| `os.arch.arm64` | CPU architecture is ARM64 (Apple Silicon, etc.) | `if: os.arch.arm64` |
| `os.arch.amd64` | CPU architecture is AMD64 (x86-64) | `if: os.arch.amd64` |
| `command.<name>` | Command `<name>` is available in PATH | `if: command.brew` |
| `env.<NAME>` | Environment variable `<NAME>` is set and non-empty | `if: env.CI` |
| `file.exists("<path>")` | File exists at `<path>` (relative to project root) | `if: file.exists("Makefile")` |
| `dir.exists("<path>")` | Directory exists at `<path>` (relative to project root) | `if: dir.exists(".venv")` |
| `shell.zsh` | Current shell is zsh | `if: shell.zsh` |
| `shell.bash` | Current shell is bash | `if: shell.bash` |

## Operators

| Operator | Description | Precedence | Example |
|----------|-------------|------------|---------|
| `not` | Negates the following predicate | highest | `if: not command.jq` |
| `and` | Both sides must be true | medium | `if: os.macos and command.brew` |
| `or` | Either side must be true | lowest | `if: os.macos or os.linux` |
| `(...)` | Grouping — overrides precedence | — | `if: os.macos and (command.brew or command.curl)` |

`and` binds tighter than `or`, so `a or b and c` is parsed as `a or (b and c)`.

## Behavior

### On steps

When a step's `if:` evaluates to false, the step is silently skipped — no trace line, no error, no output.

```yaml
steps:
  - bash: brew install jq
    if: os.macos and not command.jq

  - bash: apt-get install -y jq
    if: os.linux and not command.jq
```

### Pipe passthrough on skip

When a `pipe: true` step is skipped by its condition, any piped data from a prior step passes through unchanged to the next step:

```yaml
steps:
  - bash: echo "hello"
    pipe: true
  - bash: tr a-z A-Z
    if: env.TRANSFORM
    pipe: true
  - bash: cat
```

If `TRANSFORM` is not set, the middle step is skipped and `"hello"` flows directly to `cat`.

### On automations

When an automation's top-level `if:` evaluates to false, the entire automation is skipped with a message: `[skipped] <name> (condition: <expr>)`. A `run:` step calling a skipped automation succeeds without error.

```yaml
description: Install Homebrew (macOS only)
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL ...)"
```

### On setup entries

Setup entries in `pi.yaml` support `if:` in object form. When the condition is false, the entry is skipped with a message.

```yaml
setup:
  - run: setup/install-xcode-tools
    if: os.macos
  - run: setup/install-build-essential
    if: os.linux
```

### On `first:` sub-steps

Each sub-step's `if:` is evaluated in order. Only the first matching sub-step executes:

```yaml
steps:
  - first:
      - bash: mise install python@3.13
        if: command.mise
      - bash: brew install python@3.13
        if: command.brew
      - bash: echo "no installer found" >&2 && exit 1
```

### Validation

Invalid `if:` expressions are caught at YAML parse time, not at runtime. Unknown predicates produce an error listing valid prefixes.

## Combining Conditions

```yaml
# Multiple checks
if: os.macos and not command.jq

# Platform alternatives
if: os.macos or os.linux

# Grouped logic
if: os.macos and (command.brew or command.curl)

# Nested negation
if: not (os.windows or os.linux)

# Environment-based
if: env.CI and command.docker

# File checks
if: file.exists("docker-compose.yml") and command.docker
```
