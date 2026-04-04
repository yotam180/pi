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

This works uniformly for bash, python, and typescript steps with no per-language plumbing.

### Validation rules
- Required inputs with no `with:` value → hard error before any step runs, naming the missing input
- Unknown keys in `with:` → hard error (prevents silent typos)
- Default values are injected automatically when the caller omits the key
- Type is currently informational (string only for now); reserved for future validation

## Acceptance Criteria
- [ ] `inputs:` block is parsed from automation YAML into the `Automation` struct
- [ ] `with:` block is parsed on `run:` steps (both in automation files and `pi.yaml` setup entries)
- [ ] Required inputs missing from `with:` produce a clear error before execution starts
- [ ] Unknown keys in `with:` produce a clear error
- [ ] Default values are applied when a key is omitted
- [ ] Input values are injected as `PI_INPUT_<NAME>` env vars for all step types
- [ ] `pi list` shows declared inputs for automations that have them
- [ ] `pi run --help <name>` (or equivalent) prints the automation's description and input docs
- [ ] Schema is identical for local, built-in, and marketplace automations — no special cases
- [ ] `go test ./...` passes; unit tests cover validation, default injection, and env var generation

## Implementation Notes
<!-- Fill in as you work. -->

## Subtasks
- [ ] Extend `Automation` struct with `Inputs map[string]InputSpec`
- [ ] Extend `stepRaw` / `Step` with `With map[string]string`
- [ ] Extend `pi.yaml` `SetupEntry` with `With map[string]string`
- [ ] Implement validation in `Automation.validate()`: check required inputs, unknown keys
- [ ] Inject `PI_INPUT_*` env vars in executor before running each step
- [ ] Update `pi list` to show inputs column or inline declaration
- [ ] Write unit tests for all validation paths
- [ ] Write integration test: automation with required input called correctly and incorrectly

## Blocked By
<!-- None — can be designed and implemented before Project 3 work begins. -->
