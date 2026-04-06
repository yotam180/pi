# Tool Registry Centralization

## Status
done

## Priority
medium

## Description

Two separate, manually-maintained maps in the CLI layer encode knowledge about installable tools: `setupAddKnownTools` in `cli/setup_add.go` (maps short names to `pi:install-*` automations) and `installHints` in `reqcheck/reqcheck.go` (maps tool names to install commands). Every new builtin added to the library requires updating both maps independently, with no structural enforcement. This project introduces a `ToolDescriptor` type and a central registry that both maps derive from, making tool additions a single-change operation.

## Goals

- A `ToolDescriptor` type encodes: canonical builtin name, short-form aliases, and human-readable install hint.
- `setupAddKnownTools` is generated/derived from the tool registry — not manually written.
- `installHints` is derived from the tool registry — not maintained as a separate map.
- Adding a new tool to the builtin library requires editing only the tool registry definition.

## Success Criteria

- [x] `setupAddKnownTools` contains no hand-written entries; it is derived from `tools.Registry`
- [x] `installHints` in `reqcheck` is derived from `tools.Registry`
- [x] `pi:ruby` and all other builtins have correct short-form and `pi:` prefix aliases
- [x] A test fails if a `pi:install-*` builtin exists without a `ToolDescriptor` entry
- [x] `go build ./...` and `go test ./...` pass

## Notes

Completed in one session. Both project tasks (`tool-descriptor-type` and `install-hints-from-tool-registry`) were completed together since the install-hints migration was a natural extension of the registry creation.

Key decisions:
- `ToolDescriptor` does not include `CommandName` (initially proposed in the design). The `ShortNames` field is sufficient — `InstallHintFor()` matches by short name.
- `ruby` was moved to a command-only entry (no `BuiltinName`) since `pi:install-ruby` doesn't exist as a builtin. A reverse-coverage test (`TestRegistryBuiltinNamesExist`) enforces this.
- Dependency direction is clean: `cli → tools`, `reqcheck → tools`. No import cycles.
- `tools` package only imports `builtins` in test files.
