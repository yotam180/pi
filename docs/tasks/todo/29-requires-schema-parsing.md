# Requires Schema Parsing

## Type
feature

## Status
todo

## Priority
high

## Project
05-environment-robustness

## Description
Add a `requires:` block to the automation YAML schema. This block declares the tools and runtimes an automation needs before any step runs. The parser should handle four requirement forms:

1. `python >= 3.11` — runtime with minimum version
2. `node` — runtime, any version
3. `command: docker` — any command in PATH
4. `command: kubectl >= 1.28` — command with minimum version

The `requires:` block should be parsed into a `[]Requirement` type on the `Automation` struct. Each `Requirement` has fields: `Name`, `Kind` (runtime or command), `MinVersion` (optional). The parsing should handle the four forms above via custom YAML unmarshalling.

`requires:` is valid on both `steps:`-based and `install:`-based automations.

### YAML examples

```yaml
requires:
  - python >= 3.11
  - command: docker
  - command: jq
```

```yaml
requires:
  - node >= 18
  - command: kubectl >= 1.28
```

## Acceptance Criteria
- [ ] `Requirement` type added to `internal/automation/automation.go` with `Name`, `Kind` (runtime/command), `MinVersion` fields
- [ ] Custom YAML unmarshalling handles all 4 forms: `<runtime>`, `<runtime> >= <version>`, `command: <name>`, `command: <name> >= <version>`
- [ ] `Automation.Requires` field populated during parsing
- [ ] Invalid `requires:` entries produce clear parse errors
- [ ] Unit tests for all forms, edge cases (bad version syntax, unknown runtime names)
- [ ] `go test ./...` passes

## Implementation Notes

## Subtasks
- [ ] Define `RequirementKind` (runtime/command) and `Requirement` struct
- [ ] Implement `requirementRaw` YAML unmarshalling (handles scalar strings and `command:` mappings)
- [ ] Parse `requires:` in `Automation.UnmarshalYAML`
- [ ] Add semver parsing utility for `>= X.Y.Z` version constraints
- [ ] Write unit tests for all requirement forms
- [ ] Write unit tests for error cases

## Blocked By
