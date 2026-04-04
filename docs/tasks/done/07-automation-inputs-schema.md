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
- [x] `inputs:` block is parsed from automation YAML into the `Automation` struct
- [x] `with:` block is parsed on `run:` steps (both in automation files and `pi.yaml` setup entries)
- [x] `with:` block is parsed on shortcut definitions in `pi.yaml`, supporting `$1`, `$2`, `$@` expressions
- [x] Positional args to `pi run <name> arg1 arg2` are auto-mapped to declared inputs in declaration order
- [x] `pi run --with key=value` flag is accepted as an explicit alternative to positional mapping
- [x] Mixing positional args and `--with` flags in the same call produces a clear error
- [x] Required inputs missing from all sources produce a clear error before execution starts
- [x] Unknown keys in `with:` produce a clear error
- [x] Default values are applied when a key is omitted and no positional/`--with` covers it
- [x] Input values are injected as `PI_INPUT_<NAME>` env vars for all step types
- [x] `pi list` shows declared inputs for automations that have them
- [ ] `pi run --help <name>` (or equivalent) prints the automation's description and input docs — deferred to a follow-up task
- [x] Schema is identical for local, built-in, and marketplace automations — no special cases
- [x] `go test ./...` passes; unit tests cover validation, default injection, positional mapping, and env var generation

## Implementation Notes

### Architecture decisions
- `InputSpec.Required` is `*bool` to distinguish "not set" from "false". `IsRequired()` defaults to true when no default is provided.
- `InputKeys []string` preserves YAML declaration order for positional mapping, using a custom `inputsRaw` unmarshaller.
- Validation happens in `ResolveInputs()` on the `Automation` struct, not in the executor — clean separation.
- `appendInputEnv()` returns nil when there are no input env vars, so `cmd.Env = nil` inherits the parent process environment (no behavior change for automations without inputs).
- `run:` steps with `with:` call `RunWithInputs()` with the with map, while steps without `with:` forward positional args.
- Shell codegen for `with:` shortcuts emits `--with key="$N"` flags. Literal values (not `$N` references) are quoted directly.

### Test coverage
- 37 new tests added across automation (13), executor (8), config (1), cli (8), shell (3), integration (8)
- Total test count: 205

## Subtasks
- [x] Extend `Automation` struct with `Inputs map[string]InputSpec`
- [x] Extend `stepRaw` / `Step` with `With map[string]string`
- [x] Extend `pi.yaml` `SetupEntry` and `Shortcut` with `With map[string]string`
- [x] Implement `pi run --with key=value` flag (repeatable)
- [x] Implement positional → named input mapping in executor (args mapped in `inputs:` declaration order)
- [x] Implement validation in `ResolveInputs()`: check required inputs, unknown keys, mixing positional + `--with`
- [x] Inject `PI_INPUT_*` env vars in executor before running each step
- [x] Update `pi shell` codegen to emit `--with` flags when shortcut has explicit `with:` mapping
- [x] Update `pi list` to show inputs column or inline declaration
- [x] Write unit tests for all validation paths
- [x] Write integration test: automation with required input called correctly and incorrectly
- [x] Update `examples/docker-project/.pi/docker/logs.yaml` to use `inputs:` + `$PI_INPUT_SERVICE` instead of `"${1:-all}"`

## Blocked By
<!-- None — can be designed and implemented before Project 3 work begins. -->
