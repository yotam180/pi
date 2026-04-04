# Automation Inputs Schema

## Type
feature

## Status
todo

## Priority
high

## Project
03-built-in-library

## Description
Design and implement a first-class `inputs:` declaration schema for automation YAML files, plus the `with:` mechanism for passing values to them. This must be uniform across all automation origins — local `.pi/`, built-ins (`pi:`), and future marketplace automations. An automation that declares its inputs is self-documenting, validates its own contract, and is safe to share publicly.

The `inputs:` block goes in the automation YAML file and declares what parameters the automation accepts. The `with:` block goes at the call site (a `run:` step or a `pi.yaml` setup entry) and provides values.

### Proposed automation-side schema

```yaml
name: install-cursor-extension
description: Install a Cursor extension if not already installed

inputs:
  extension_id:
    type: string
    required: true
    description: The extension ID (e.g. "eamodio.gitlens")
  version:
    type: string
    required: false
    default: latest
    description: Extension version to pin to, or "latest"

steps:
  - bash: |
      ext="$PI_INPUT_EXTENSION_ID"
      ver="$PI_INPUT_VERSION"
      ...
```

### Proposed call-site schema

```yaml
# in pi.yaml setup:
- run: setup/install-cursor-extension
  with:
    extension_id: eamodio.gitlens
    version: "1.2.3"

# in a run: step inside another automation
steps:
  - run: setup/install-cursor-extension
    with:
      extension_id: eamodio.gitlens
```

### Input injection

Inputs are exposed to steps as environment variables, uppercased with a `PI_INPUT_` prefix:
- `extension_id` → `$PI_INPUT_EXTENSION_ID`
- `version` → `$PI_INPUT_VERSION`

This works uniformly for bash, python, and typescript steps with no per-language plumbing. The `service="${1:-all}"` pattern in bash steps becomes unnecessary — the default lives in the `inputs:` declaration, and the step just reads `$PI_INPUT_SERVICE`.

### Positional → named input mapping

When `pi run <name> arg1 arg2` is called directly (no `with:`), positional arguments are auto-mapped to declared inputs **in declaration order**:

```yaml
# docker/logs.yaml
inputs:
  service:
    type: string
    default: all
    description: Docker Compose service to tail
  tail:
    type: string
    default: "200"
    description: Number of log lines

steps:
  - bash: docker-compose logs -f --tail $PI_INPUT_TAIL $PI_INPUT_SERVICE
```

```bash
pi run docker/logs           # service=all, tail=200 (both defaults)
pi run docker/logs api       # service=api, tail=200
pi run docker/logs api 500   # service=api, tail=500
```

Explicit `with:` always takes precedence over positional mapping. Mixing positional and `with:` in the same call is an error.

### Shortcut `with:` mapping

Shortcuts can explicitly map their CLI positional args to named inputs using `$1`, `$2`, `$@` expressions in `with:` values. This allows shortcuts to reorder, rename, or provide fixed values for inputs:

```yaml
shortcuts:
  # Simple: forwards "$@" → positional → auto-mapped to inputs in order
  dlogs: docker/logs

  # Explicit: map CLI args to named inputs
  dlogs:
    run: docker/logs
    with:
      service: $1          # dlogs api → service=api
      tail: $2             # dlogs api 500 → tail=500

  # Fixed value override: always use 50 lines for this shortcut
  dlogs-short:
    run: docker/logs
    with:
      tail: "50"           # literal string, not a positional arg
      service: $1
```

For the simple `dlogs: docker/logs` form, the generated shell function is:
```bash
function dlogs() {
  (cd /path/to/repo && pi run docker/logs "$@")
}
```

For the explicit `with:` form, the generated shell function passes named inputs:
```bash
function dlogs() {
  (cd /path/to/repo && pi run docker/logs --with service="$1" --with tail="${2:-200}")
}
```

This means `pi run` must accept `--with key=value` flags as an alternative to positional arg mapping. The `--with` flags are passed by generated shell functions; users can also use them directly on the CLI.

### Validation rules
- Required inputs with no `with:` value and no positional arg → hard error before any step runs, naming the missing input
- Unknown keys in `with:` → hard error (prevents silent typos)
- Default values are applied automatically when a key is omitted and no positional arg covers it
- Mixing positional args and `--with` flags in the same `pi run` call → hard error
- `$1`/`$2` in a shortcut `with:` value that refers to an unset CLI arg → use input's default, or error if required
- Type is currently informational (string only for now); reserved for future validation

## Acceptance Criteria
- [ ] `inputs:` block is parsed from automation YAML into the `Automation` struct
- [ ] `with:` block is parsed on `run:` steps (both in automation files and `pi.yaml` setup entries)
- [ ] `with:` block is parsed on shortcut definitions in `pi.yaml`, supporting `$1`, `$2`, `$@` expressions
- [ ] Positional args to `pi run <name> arg1 arg2` are auto-mapped to declared inputs in declaration order
- [ ] `pi run --with key=value` flag is accepted as an explicit alternative to positional mapping
- [ ] Mixing positional args and `--with` flags in the same call produces a clear error
- [ ] Required inputs missing from all sources produce a clear error before execution starts
- [ ] Unknown keys in `with:` produce a clear error
- [ ] Default values are applied when a key is omitted and no positional/`--with` covers it
- [ ] Input values are injected as `PI_INPUT_<NAME>` env vars for all step types
- [ ] `pi list` shows declared inputs for automations that have them
- [ ] `pi run --help <name>` (or equivalent) prints the automation's description and input docs
- [ ] Schema is identical for local, built-in, and marketplace automations — no special cases
- [ ] `go test ./...` passes; unit tests cover validation, default injection, positional mapping, and env var generation

## Implementation Notes
<!-- Fill in as you work. -->

## Subtasks
- [ ] Extend `Automation` struct with `Inputs map[string]InputSpec`
- [ ] Extend `stepRaw` / `Step` with `With map[string]string`
- [ ] Extend `pi.yaml` `SetupEntry` and `Shortcut` with `With map[string]string`
- [ ] Implement `pi run --with key=value` flag (repeatable)
- [ ] Implement positional → named input mapping in executor (args mapped in `inputs:` declaration order)
- [ ] Implement validation in `Automation.validate()`: check required inputs, unknown keys, mixing positional + `--with`
- [ ] Inject `PI_INPUT_*` env vars in executor before running each step
- [ ] Update `pi shell` codegen to emit `--with` flags when shortcut has explicit `with:` mapping
- [ ] Update `pi list` to show inputs column or inline declaration
- [ ] Write unit tests for all validation paths
- [ ] Write integration test: automation with required input called correctly and incorrectly
- [ ] Update `examples/docker-project/.pi/docker/logs.yaml` to use `inputs:` + `$PI_INPUT_SERVICE` instead of `"${1:-all}"`

## Blocked By
<!-- None — can be designed and implemented before Project 3 work begins. -->
