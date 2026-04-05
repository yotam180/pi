---
title: Quick Start
description: Run your first PI automation in under 5 minutes
---

This guide takes you from zero to a working automation in under 5 minutes. No configuration files, no setup — just a YAML file and a command.

## Step 1 — Your First Automation

Create a `.pi/` folder at your project root and add a file:

```yaml
# .pi/greet.yaml
description: Say hello
bash: echo "Hello from PI!"
```

Run it:

```bash
pi run greet
```

Output:

```
  → bash: echo "Hello from PI!"
Hello from PI!
```

That's it. No `pi.yaml`, no `steps:` block, no configuration. Just a file in `.pi/` and a command.

## Step 2 — Something Useful

Replace the toy example with something real. Most projects have a test suite:

```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```

```bash
pi run test
```

Every `.yaml` file in `.pi/` becomes an automation. The name is derived from the file path:

| File | Automation name |
|------|----------------|
| `.pi/test.yaml` | `test` |
| `.pi/docker/up.yaml` | `docker/up` |
| `.pi/setup/install-deps.yaml` | `setup/install-deps` |

## Step 3 — A Shortcut You Can Type from Anywhere

Now create a `pi.yaml` at your project root. This is the only reason you need one: to declare shell shortcuts.

```yaml
# pi.yaml
project: my-app

shortcuts:
  test: test
```

Install the shortcut into your shell:

```bash
pi shell
source ~/.zshrc  # or open a new terminal
```

Now `test` is a terminal command that works from any directory:

```bash
cd ~/some/other/folder
test    # runs pi run test, from your project root
```

`pi shell` writes a shell function into your `.zshrc` (or `.bashrc`) that navigates to the project root and runs the automation.

## Step 4 — Chaining Steps

When you need to do two things in sequence, use `steps:`:

```yaml
# .pi/build-and-test.yaml
description: Build then run tests
steps:
  - bash: go build ./...
  - bash: go test ./...
```

```bash
pi run build-and-test
```

:::tip[Short form and full form are identical]
The `bash:` key you used earlier is shorthand for a single-step automation. These two files are exactly equivalent:

```yaml
# Short form
bash: go test ./...
```

```yaml
# Full form
steps:
  - bash: go test ./...
```

Use the short form when there's only one step. Use `steps:` when you need more.
:::

## What's Next

You've created automations, run them, set up a shortcut, and chained steps. Here's where to go deeper:

- **[Automations](/concepts/automations/)** — step types, environment variables, working directories, conditions, and more
- **[pi.yaml](/concepts/pi-yaml/)** — `setup:`, `packages:`, and all root config options
- **[Shell Shortcuts](/concepts/shell-shortcuts/)** — the global wrapper, `anywhere: true`, and `parent_shell: true`
- **[Setup Automations](/guides/setup-automations/)** — onboard teammates with `pi setup`
