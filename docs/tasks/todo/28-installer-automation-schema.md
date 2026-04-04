# Installer Automation Schema

## Type
feature

## Status
todo

## Priority
high

## Project
03-built-in-library

## Description
Replace free-form `steps:` in installer automations with a structured `install:` block that explicitly declares the test-run-verify lifecycle. PI itself owns all status output (`[installing]`, `[✓ installed]`, `[already installed]`, `[✗ failed]`) — the automation only provides the bash commands. A `--silent` flag suppresses PI's own output for scripted/CI contexts.

The current installer automations each manually echo their own status strings, resulting in inconsistent formatting, duplicated logic, and bash scripts that mix installation commands with output concerns. The `install:` block solves this by separating the "what to do" (automation's job) from "what to tell the user" (PI's job).

## The Schema

An automation with an `install:` block replaces `steps:`. The two are mutually exclusive.

### Scalar shorthand (simple installs)

When a field is a string, it is treated as inline bash — the same as `bash: |` in a step:

```yaml
name: install-homebrew
description: Install Homebrew (macOS only)
if: os.macos

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  verify: brew --version >/dev/null 2>&1
  version: brew --version | head -1 | awk '{print $2}'
```

### Step list (rich installs with `if:` conditions and automation references)

`test:`, `run:`, and `verify:` also accept a **list of step objects** — the full step schema, including `if:` predicates, all step types (`bash:`, `python:`, `typescript:`), and `run:` references to other automations. This allows conditional install paths without nesting bash `if/elif`:

```yaml
name: install-python
description: Install Python at a specific version

inputs:
  version:
    type: string
    required: true
    description: Python version to install (e.g. "3.13")

install:
  test:
    - bash: python3 --version 2>&1 | grep -q "Python $PI_INPUT_VERSION"

  run:
    - bash: mise install "python@$PI_INPUT_VERSION" && mise use "python@$PI_INPUT_VERSION"
      if: command.mise
    - bash: brew install "python@$PI_INPUT_VERSION"
      if: command.brew and not command.mise
    - bash: |
        echo "no suitable installer found (tried mise, brew)" >&2
        exit 1
      if: not command.mise and not command.brew

  verify:
    - bash: python3 --version 2>&1 | grep -q "Python $PI_INPUT_VERSION"

  version: python3 --version 2>&1 | awk '{print $2}'
```

Referencing another automation in a `run:` list works the same as a `run:` step anywhere else — the sub-automation's own lifecycle (including its own `install:` block if it has one) runs in full:

```yaml
install:
  test: command -v node >/dev/null 2>&1
  run:
    - run: pi:install-homebrew
      if: os.macos and not command.brew
    - bash: brew install node
      if: os.macos
    - bash: apt-get install -y nodejs
      if: os.linux
  version: node --version | sed 's/^v//'
```

### Field reference

| Field | Required | Accepts | Description |
|---|---|---|---|
| `test` | yes | string or step list | Exits 0 if already installed. All stdout/stderr suppressed. |
| `run` | yes | string or step list | Performs the installation. Stdout suppressed; stderr shown on failure. |
| `verify` | no | string or step list | Confirms install succeeded. Stdout/stderr suppressed. **Defaults to re-running `test` if omitted.** |
| `version` | no | string only | Stdout captured as the version string for display. |

String fields accept inline bash or a `.sh` / `.py` / `.ts` file path relative to the automation file — same rules as the corresponding step type. `version:` is always a string (it is a capture command, not a workflow).

### `verify` defaults to `test`

When `verify:` is not declared, PI re-runs `test:` after a successful `install:` as the verification step. This is the right default for nearly all installer automations — if "is it installed?" was the right check before, it's the right check after. Declaring an explicit `verify:` is only needed when the post-install check differs from the pre-install check (e.g. verifying a specific version was actually activated after a `mise use`).

`install-homebrew` therefore needs no `verify:` at all:

```yaml
install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  version: brew --version | head -1 | awk '{print $2}'
```

### Execution semantics for step lists

- Steps in a list run in order, same as `steps:` in a regular automation
- `if:` conditions are evaluated before each step; steps that don't match are skipped silently
- If any step exits non-zero, the entire phase (`test`, `run`, `verify`) is considered failed
- A `run:` step inside an `install:` list that references an installer automation runs that automation's `install:` lifecycle, not its `steps:`

## PI-Managed Output

PI prints one status line per installer automation. The automation's own stdout is suppressed — it does not print anything. The `run:` bash should be silent; PI manages the narrative.

```
pi setup

  ✓  install-homebrew      already installed   (4.2.1)
  →  install-python        installing...
  ✓  install-python        installed           (3.13.0)
  ✗  install-node          failed

      stderr from run: command:
      error: no release found for node@20 on darwin/arm64
```

Rules:
- `test` exits 0 → print `✓  <name>  already installed  (<version>)`, skip `run`
- `test` exits non-zero → print `→  <name>  installing...` → run `run`
- `run` exits 0 and `verify` passes → print `✓  <name>  installed  (<version>)`
- `run` or `verify` exits non-zero → print `✗  <name>  failed` + indented stderr

When `version:` is not declared, the version column is omitted.

## `--silent` Flag

`pi setup --silent` and `pi run --silent <name>` suppress PI's own status lines. The automation's stderr is still shown on failure regardless of `--silent` (errors must always surface).

`--silent` does not affect regular `steps:`-based automations — their stdout is the product, not status noise.

## Acceptance Criteria
- [ ] `install:` block is parsed from automation YAML into the `Automation` struct (`InstallSpec` type)
- [ ] `steps:` and `install:` are mutually exclusive — validation errors if both are present
- [ ] Executor detects `install:` automations and uses the structured lifecycle instead of step iteration
- [ ] `test` bash runs first; exit 0 short-circuits to "already installed" output
- [ ] `run` bash stdout is suppressed; stderr is captured and shown only on failure
- [ ] When `verify:` is absent, `test:` is re-run as the verification step after `run:` completes
- [ ] When `verify:` is present, it runs instead of re-running `test:`; failure is treated as install failure
- [ ] `version` bash (if present) is run after successful `test` or `verify`; its stdout is trimmed and used in the status line
- [ ] PI prints the status table described above during `pi setup` and `pi run` on installer automations
- [ ] `--silent` flag suppresses PI status lines on both `pi setup` and `pi run`
- [ ] All existing built-in installer automations migrated from `steps:` to `install:` schema:
  - `install-homebrew.yaml`
  - `install-python.yaml`
  - `install-node.yaml`
  - `install-uv.yaml`
  - `install-tsx.yaml`
- [ ] `go test ./...` passes
- [ ] Integration test: run `pi:install-homebrew` on a machine where brew is already installed; verify output is `✓ already installed` with no extra noise

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Add `InstallSpec` struct to `internal/automation/automation.go` with `Test`, `Run`, `Verify` as `InstallPhase` (either string or `[]Step`) and `Version` as string
- [ ] Implement `InstallPhase` YAML unmarshalling: detect scalar vs sequence and parse accordingly
- [ ] Parse `install:` block in `UnmarshalYAML`; validate mutual exclusion with `steps:`
- [ ] Add installer execution path to executor: `execInstall()`
- [ ] For string phases: run as inline bash with stdout/stderr suppression
- [ ] For step-list phases: reuse existing step execution with `if:` evaluation; suppress stdout, capture stderr
- [ ] Implement `version:` capture (run bash, trim stdout)
- [ ] Add status line printer to executor (format: `✓ / → / ✗  name  status  (version)`)
- [ ] Add `--silent` flag to `pi setup` and `pi run` cobra commands; thread through to executor
- [ ] Migrate all 5 built-in installer automations to `install:` schema
- [ ] Write unit tests for `InstallSpec` parsing: scalar, step list, mixed, mutual exclusion, verify-defaults-to-test
- [ ] Write integration test for already-installed and fresh-install paths
- [ ] Write integration test: `run:` reference to another installer automation inside an `install:` block

## Blocked By
<!-- None — can be worked independently of inputs schema -->
