# Config and Automation Schema

## Type
feature

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Define and parse the two config formats PI uses: `pi.yaml` (project-level config) and automation YAML files (individual automation definitions). Implement loading, validation, and clear error messages. No execution yet — just parsing.

## Acceptance Criteria
- [x] `pi.yaml` is parsed into a typed Go struct covering: `project`, `shortcuts`, `setup`
- [x] Automation YAML is parsed into a typed Go struct covering: `name`, `description`, `steps` (each step has a type and payload)
- [x] Validation rejects unknown fields and missing required fields with a clear error message pointing to the file and field
- [x] A `config` package (or similar) exposes `Load(dir string)` that reads `pi.yaml` from the given directory
- [x] An `automation` package exposes `Load(path string)` that reads a single automation YAML file
- [x] Unit tests cover: valid parse, missing required field, unknown step type, malformed YAML

## Implementation Notes

### Decisions
- **YAML library**: `gopkg.in/yaml.v3` — the standard Go YAML library, well-maintained, supports custom unmarshallers.
- **Package structure**: Two separate packages `internal/config` and `internal/automation`, each with its own `Load()` entrypoint. Clean separation of concerns — config is the project-level pi.yaml, automation is an individual automation file.
- **Shortcut union type**: Used custom `UnmarshalYAML` on the `Shortcut` struct to support both plain string (`up: docker/up`) and object form (`deploy: {run: ..., anywhere: true}`). Checks for `ScalarNode` vs `MappingNode`.
- **Step union type**: Used an intermediate `stepRaw` struct with nullable pointers for each step type (`*string`). During unmarshalling, exactly one must be set. This gives clear errors for: no type, multiple types, unknown types.
- **Python/TypeScript**: Defined as valid step types in the schema (they parse successfully) but marked as "not yet implemented" during validation. This means the schema is forward-compatible and the error is clear.
- **Validation strategy**: Validate after parsing. Errors include the file path and field name for easy debugging.

### File structure
```
internal/config/config.go        — ProjectConfig, Shortcut, SetupEntry structs + Load()
internal/config/config_test.go   — 8 tests
internal/automation/automation.go — Automation, Step, StepType structs + Load()
internal/automation/automation_test.go — 14 tests
```

### Test coverage
- **Config tests (8)**: valid full config, minimal valid, missing file, missing project field, empty shortcut run, empty setup run, malformed YAML, string vs object shortcuts
- **Automation tests (14)**: valid bash step, multiple steps, pipe_to, inline multiline bash, missing file, missing name, no steps, no step type specified, multiple step types, python not implemented, typescript not implemented, malformed YAML, empty step value, StepType.IsImplemented()

## Subtasks
- [x] Define `pi.yaml` Go structs
- [x] Define automation YAML Go structs (step union type: bash/run/python/typescript — mark python+typescript as unsupported for now)
- [x] Implement loaders with validation
- [x] Write unit tests (22 total — 8 config, 14 automation)

## Blocked By
01-project-scaffold
