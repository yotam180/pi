---
title: Introduction
description: What PI is, what problem it solves, and why you should use it
---

PI is a developer automation tool that replaces your team's `shell_shortcuts.sh` with structured, polyglot, version-controlled automations.

## The Problem

Most engineering teams accumulate a `shell_shortcuts.sh` and a `setup_environment.sh` in their repo. These files:

- Are written entirely in bash, which becomes unreadable as they grow
- Reach 1000+ lines with no clear structure or discoverability
- Can't mix languages — your log formatter wants Python, your setup script is bash, your linter config is TypeScript

PI replaces all of that with a structured model: each automation is a YAML file in a `.pi/` folder, runnable with `pi run`, and shareable across projects via GitHub packages.

## What It Looks Like

```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```

```bash
$ pi run test
  → bash: go test ./...
ok  github.com/example/myapp  0.423s
```

That's a complete automation — one file, one command. No boilerplate.

## Key Ideas

**Polyglot by default** — Steps can be bash, Python, or TypeScript. Mix them freely within a single automation.

**Lives in your repo** — Automations are YAML files in `.pi/`. They're version-controlled, PR-reviewed, and shared with the team automatically.

**Shell shortcuts** — `pi shell` turns any automation into a terminal command that works from any directory.

**Shareable packages** — Publish automations to a GitHub repo. Teams add them with `pi add org/repo@v1.0`.

## Next

[Install PI →](/getting-started/installation/)
