# Environment Robustness

## Status
todo

## Priority
medium

## Description
Make PI automations portable and self-describing with respect to their runtime requirements. Automations that need Node.js, Python, or specific tools should declare this upfront; PI validates before running, fails fast with actionable errors, and — optionally — provisions missing runtimes automatically into a sandboxed directory so automations work on a fresh machine without a prior `pi setup`.

## Goals
- An automation can declare exactly what it needs to run (`requires:` block)
- PI validates all requirements before executing any step — no mid-run failures from missing runtimes
- Missing runtimes produce a clear error with concrete install instructions, not a cryptic "command not found" from bash
- `pi doctor` gives an at-a-glance health check of all requirements across a project's automations
- Sandboxed runtime provisioning: PI can download runtimes into `~/.pi/runtimes/` without touching the system, so a shared automation "just works" on a fresh machine
- The full test suite runs against multiple clean Docker environments, proving portability

## Background & Context

Today, a TypeScript step on a machine without `tsx` fails mid-execution with a shell error. The same for Python steps without `python3`. The developer has to trace back the error, figure out what's missing, install it, and retry — all friction that PI should eliminate.

The `pi:install-node` and `pi:install-python` built-ins (Project 3) solve this for teams who run `pi setup` first. But for ad-hoc `pi run` usage, or when sharing automations with people who haven't run setup, there needs to be a better fallback than a raw shell error.

Sandboxed runtime provisioning is the key insight: rather than requiring system-level installs, PI can download a self-contained runtime into `~/.pi/runtimes/` and use it only for that step execution. The system is untouched. This is opt-in and integrates with `mise` when available (which already handles multi-platform runtime downloads and version management) rather than reimplementing a downloader.

Testing across clean Docker environments is necessary because "works on my machine" is not sufficient for a tool whose job is to work on every machine.

## Scope

### In scope

**`requires:` block in automation YAML**
Declare the tools and runtimes an automation needs before any step runs:
```yaml
name: format-logs
requires:
  - python >= 3.11
  - command: docker
  - command: jq

steps:
  - python: formatter.py
```

Supported requirement forms:
- `node >= 18`, `python >= 3.11` — runtime with minimum version
- `node`, `python` — runtime, any version
- `command: docker` — any command in PATH
- `command: kubectl >= 1.28` — command with minimum version (parsed from `kubectl version`)

**Pre-execution validation**
Before running any step, PI checks all `requires:`. If anything is missing, it fails with a clear error listing every missing requirement and a concrete install hint for each:
```
✗ pi run format-logs

  Missing requirements:
    python >= 3.11   not found (python3 --version returned 3.9.7)
                     → install: brew install python@3.13   or  pi setup (if configured)
    command: jq      not found
                     → install: brew install jq
```

**`pi doctor` command**
Scan all automations in the project, collect their `requires:`, and print a health table:
```
pi doctor

  docker/logs-formatted
    ✓ python >= 3.11     (3.13.0)
    ✓ command: docker    (24.0.5)

  setup/build-image
    ✓ command: docker    (24.0.5)
    ✗ command: kubectl   not found → brew install kubectl
```

**Sandboxed runtime provisioning (opt-in)**
When enabled in `pi.yaml`, PI provisions missing runtimes into `~/.pi/runtimes/<runtime>/<version>/` instead of erroring:
```yaml
# pi.yaml
runtimes:
  provision: auto     # auto | ask | never (default: never)
  manager: mise       # mise | direct (default: mise if available, else direct)
```

Provisioning flow:
1. `provision: never` (default) — error with install hint, do nothing
2. `provision: ask` — prompt the developer, provision if confirmed
3. `provision: auto` — silently provision, print one line: `[provisioned] node 20.11.0 → ~/.pi/runtimes/node/20.11.0/`

When `manager: mise`, PI calls `mise install node@20` and prepends the mise shim path. When `manager: direct`, PI downloads a pre-built binary from the official release CDN for the current OS/arch.

Provisioned runtimes are used only for PI step execution — they are never added to the system PATH permanently.

**Docker-based test matrix**
Test environments for CI and local verification:
```
tests/docker/
  ubuntu-fresh/       Dockerfile — Ubuntu 24.04, no runtimes, no tools
  ubuntu-node/        Dockerfile — Ubuntu 24.04, Node 20 only
  ubuntu-python/      Dockerfile — Ubuntu 24.04, Python 3.13 only
  alpine-fresh/       Dockerfile — Alpine 3.19, bare minimum
```

A `make test-matrix` target builds each image, runs `go test ./...` + integration tests inside each, and reports pass/fail per environment. The CI pipeline runs this matrix on every PR.

### Out of scope
- Windows support
- Full dependency graph resolution between automations
- Rollback if a provisioned runtime fails mid-install
- Network-isolated / air-gapped provisioning
- Provisioning for `command:` requirements (only runtimes — node, python, etc.)

## Success Criteria
- [ ] `requires:` block is parsed and validated in automation YAML
- [ ] Pre-execution validation runs before any step and produces the formatted error table above
- [ ] `pi doctor` scans all automations and prints per-automation requirement status
- [ ] `provision: auto` mode silently provisions node and python via `mise` (or direct download as fallback) and runs the step successfully
- [ ] Provisioned runtimes are isolated to `~/.pi/runtimes/` — nothing written to system PATH
- [ ] `tests/docker/ubuntu-fresh/` image has a passing integration test that runs a TypeScript automation from scratch via `provision: auto`
- [ ] All 4 Docker environments pass `go test ./...`
- [ ] `go test ./...` passes on the host as well

## Notes
- `requires:` validation intentionally runs before `if:` conditions are evaluated — a step that's skipped by `if:` doesn't need its requirements checked. Implementation should evaluate `if:` first, then check `requires:` only for steps that will actually run.
- `mise` is the preferred provisioning backend because it already handles multi-platform binary downloads, version resolution, and shim management. Don't reinvent this. If `mise` is not on the machine, fall back to direct CDN download for node and python only.
- `pi doctor` should be fast — it does PATH lookups and `--version` calls, not network requests.
- The Docker test environments should be checked in and buildable with no external dependencies beyond Docker. They serve as both CI infrastructure and documentation of what "a fresh machine" means.
- Version comparison for `requires:` uses semver. If a tool's `--version` output doesn't parse cleanly, treat it as "version unknown" and satisfy any `>= X` requirement with a warning, not a hard failure.
