# Requires Schema Parsing

## Type
feature

## Status
done

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
- [x] `Requirement` type added to `internal/automation/automation.go` with `Name`, `Kind` (runtime/command), `MinVersion` fields
- [x] Custom YAML unmarshalling handles all 4 forms: `<runtime>`, `<runtime> >= <version>`, `command: <name>`, `command: <name> >= <version>`
- [x] `Automation.Requires` field populated during parsing
- [x] Invalid `requires:` entries produce clear parse errors
- [x] Unit tests for all forms, edge cases (bad version syntax, unknown runtime names)
- [x] `go test ./...` passes

## Implementation Notes

### Types added to `internal/automation/automation.go`
- `RequirementKind` — string enum: `RequirementRuntime` ("runtime") or `RequirementCommand` ("command")
- `Requirement` — struct with `Name`, `Kind`, `MinVersion` fields
- `requirementRaw` — internal type with `UnmarshalYAML` for polymorphic parsing (scalar strings for runtimes, mapping nodes for commands)
- `knownRuntimes` — allowlist of valid runtime identifiers: `python`, `node`

### Parsing strategy
- Scalar YAML values (strings) are parsed as runtimes. Unknown runtimes produce an error with a hint to use `command:` instead.
- Mapping YAML values with `command:` key are parsed as commands.
- Both forms support `>= X.Y.Z` version constraints with `parseNameVersion()`.
- `validateVersionString()` enforces dot-separated numeric components.

### Error messages
- Unknown runtime: suggests using `command: <name>` instead
- Empty entries: clear "cannot be empty" message
- Bad version syntax: mentions the specific non-numeric character or empty component
- Missing version after `>=`: clear "missing version after >=" message
- Invalid format (spaces without `>=`): suggests using `name >= version`
- Unknown mapping key (not "command"): lists expected key

### Tests added
14 new test functions + 19 subtests covering all 4 forms, mixed requirements, three-part versions, no-requires (backward compat), requires on installers, and all error cases (unknown runtime, empty entry, empty command, bad version, missing version, invalid key, empty version component, invalid format, bad command version). Also tested `LoadFromBytes` path.

Total automation tests: 75 (was 61).

## Subtasks
- [x] Define `RequirementKind` (runtime/command) and `Requirement` struct
- [x] Implement `requirementRaw` YAML unmarshalling (handles scalar strings and `command:` mappings)
- [x] Parse `requires:` in `Automation.UnmarshalYAML`
- [x] Add semver parsing utility for `>= X.Y.Z` version constraints
- [x] Write unit tests for all requirement forms
- [x] Write unit tests for error cases

## Blocked By
