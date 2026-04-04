# Config and Automation Schema

## Type
feature

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Define and parse the two config formats PI uses: `pi.yaml` (project-level config) and automation YAML files (individual automation definitions). Implement loading, validation, and clear error messages. No execution yet — just parsing.

## Acceptance Criteria
- [ ] `pi.yaml` is parsed into a typed Go struct covering: `project`, `shortcuts`, `setup`
- [ ] Automation YAML is parsed into a typed Go struct covering: `name`, `description`, `steps` (each step has a type and payload)
- [ ] Validation rejects unknown fields and missing required fields with a clear error message pointing to the file and field
- [ ] A `config` package (or similar) exposes `Load(dir string)` that reads `pi.yaml` from the given directory
- [ ] An `automation` package exposes `Load(path string)` that reads a single automation YAML file
- [ ] Unit tests cover: valid parse, missing required field, unknown step type, malformed YAML

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Define `pi.yaml` Go structs
- [ ] Define automation YAML Go structs (step union type: bash/run/python/typescript — mark python+typescript as unsupported for now)
- [ ] Implement loaders with validation
- [ ] Write unit tests

## Blocked By
01-project-scaffold
