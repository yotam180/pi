---
title: Cross-Platform Scripts
description: Write automations that work on macOS and Linux using conditions and first-match blocks
---

This guide shows how to write automations that behave correctly across operating systems and architectures without duplicating logic.

## What you'll learn

- How to use `if:` predicates to skip platform-specific steps
- How to use `first:` blocks for cleaner mutual exclusion
- How to handle architecture-specific builds
- How to check for optional tools before using them

---

## Platform-specific steps with `if:`

The simplest cross-platform pattern uses `if:` on each step to run only on the right OS:

```yaml
# .pi/setup/install-jq.yaml
description: Install jq on any platform

steps:
  - bash: brew install jq
    if: os.macos and not command.jq

  - bash: sudo apt-get install -y jq
    if: os.linux and not command.jq

  - bash: echo "jq is ready"
```

Each step states its conditions positively. On macOS, the first step runs and the second is skipped. On Linux, the opposite happens. On either platform, if `jq` is already installed, both install steps are skipped.

## Cleaner mutual exclusion with `first:`

When you have several mutually exclusive options — "try A, else try B, else fail" — a `first:` block is cleaner than stacking `if:` conditions:

```yaml
# .pi/setup/install-python.yaml
description: Install Python 3.13

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

`first:` evaluates each sub-step's `if:` in order and runs **only the first match**. The last sub-step has no `if:` — it acts as a fallback.

The advantage over plain `if:` steps: you don't need `and not command.mise` on the brew step. `first:` stops at the first match automatically.

:::tip
`first:` works anywhere a step appears — in `steps:`, in `install.run:`, in `install.test:`. It's a generic construct, not specific to setup automations.
:::

## Architecture-specific builds

Use `os.arch.arm64` and `os.arch.amd64` to produce architecture-specific artifacts:

```yaml
# .pi/build-release.yaml
description: Build release binary for current architecture

steps:
  - bash: go build -o dist/app-arm64 ./...
    if: os.arch.arm64
    env:
      GOARCH: arm64

  - bash: go build -o dist/app-amd64 ./...
    if: os.arch.amd64
    env:
      GOARCH: amd64
```

Or cross-compile for both architectures on any machine:

```yaml
# .pi/build-all.yaml
description: Cross-compile for both architectures

steps:
  - bash: go build -o dist/app-arm64 ./...
    env:
      GOOS: linux
      GOARCH: arm64

  - bash: go build -o dist/app-amd64 ./...
    env:
      GOOS: linux
      GOARCH: amd64
```

## Checking for optional tools

Use `command.<name>` to check whether a tool is in PATH before using it:

```yaml
# .pi/format.yaml
description: Format code (skips if tools are not installed)

steps:
  - bash: prettier --write .
    if: command.prettier

  - bash: gofmt -w .
    if: command.gofmt

  - bash: black .
    if: command.black
```

Each step only runs if its tool is available. Nothing fails — everything that can run, does.

## Combining OS and tool checks

Conditions support `and`, `or`, `not`, and parentheses for complex logic:

```yaml
steps:
  # Install only on macOS when not already present
  - bash: brew install ripgrep
    if: os.macos and not command.rg

  # Install only on Linux when not already present
  - bash: sudo apt-get install -y ripgrep
    if: os.linux and not command.rg

  # Run on any platform — rg is now expected to be present
  - bash: rg --version
```

## File and directory checks

Use `file.exists()` and `dir.exists()` for conditional steps based on project structure:

```yaml
# .pi/setup/install-deps.yaml
description: Install project dependencies based on what exists

steps:
  - bash: npm ci
    if: file.exists("package-lock.json")

  - bash: pip install -r requirements.txt
    if: file.exists("requirements.txt")

  - bash: go mod download
    if: file.exists("go.mod")

  - bash: cargo fetch
    if: file.exists("Cargo.toml")
```

Paths in `file.exists()` and `dir.exists()` are resolved relative to the project root.

## Environment variable checks

Use `env.<NAME>` to check whether an environment variable is set:

```yaml
steps:
  - bash: echo "Running in CI mode..."
    if: env.CI

  - bash: echo "Running locally — starting dev server..."
    if: not env.CI
```

## Shell detection

Use `shell.zsh` and `shell.bash` to adapt to the user's shell:

```yaml
steps:
  - bash: echo "source ~/.zshrc" >> ~/.zshrc
    if: shell.zsh

  - bash: echo "source ~/.bashrc" >> ~/.bashrc
    if: shell.bash
```

## A complete cross-platform example

Here's a realistic setup automation that works on macOS and Linux, handles multiple installers, and gracefully degrades:

```yaml
# .pi/setup/configure-dev-env.yaml
description: Set up the development environment on macOS or Linux

steps:
  # Install the preferred package manager tool
  - first:
      - bash: echo "mise already available"
        if: command.mise
      - bash: curl https://mise.run | sh
        if: os.macos or os.linux

  # Install project runtimes
  - first:
      - bash: mise install && mise use
        if: command.mise and file.exists(".mise.toml")
      - bash: echo "No .mise.toml found — skipping runtime setup"

  # Platform-specific system dependencies
  - bash: brew install protobuf grpcurl
    if: os.macos and command.brew

  - bash: sudo apt-get install -y protobuf-compiler
    if: os.linux

  # Project dependencies (works on any platform)
  - bash: go mod download
    if: file.exists("go.mod")

  - bash: npm ci
    if: file.exists("package-lock.json")
```

## Available predicates

Here's a quick reference of all condition predicates:

| Predicate | True when |
|-----------|-----------|
| `os.macos` | Running on macOS |
| `os.linux` | Running on Linux |
| `os.windows` | Running on Windows |
| `os.arch.arm64` | Architecture is ARM64 (Apple Silicon, etc.) |
| `os.arch.amd64` | Architecture is x86_64 |
| `command.<name>` | `<name>` is found in PATH |
| `env.<NAME>` | Environment variable `<NAME>` is set and non-empty |
| `file.exists("<path>")` | File exists at `<path>` (relative to project root) |
| `dir.exists("<path>")` | Directory exists at `<path>` (relative to project root) |
| `shell.zsh` | Current shell is zsh |
| `shell.bash` | Current shell is bash |

Combine with `and`, `or`, `not`, and parentheses.

For the full specification, see the [Conditions reference](/reference/conditions/).

## Summary

- Use `if:` on steps to skip platform-specific commands
- Use `first:` blocks for "try A, else B, else fail" patterns — no negation chains needed
- Use `command.<name>` to check for tools before using them
- Use `file.exists()` and `dir.exists()` for project-structure-aware automations
- Combine predicates with `and`, `or`, `not`, and parentheses for complex logic
