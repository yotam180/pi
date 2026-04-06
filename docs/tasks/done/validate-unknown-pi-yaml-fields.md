# Validate Unknown Fields in pi.yaml

## Type
improvement

## Status
in_progress

## Priority
high

## Project
standalone

## Description
The `config.Load()` function silently ignores unknown fields in `pi.yaml`. If a user writes `shortcutz:` instead of `shortcuts:`, or `pakages:` instead of `packages:`, the field is silently dropped. This violates philosophy principle #2 ("Correct the action, not the developer") and principle #4 ("Guess correctly, every time").

We already have unknown-field detection for automation YAML files (in `internal/validate/unknown_fields.go`), but `pi.yaml` itself has no such protection. This task adds unknown-field detection to pi.yaml parsing, with Levenshtein-based "did you mean?" suggestions — consistent with how we already handle automation files.

The validation should:
1. Detect unknown top-level keys in pi.yaml
2. Emit warnings (not hard errors) so we don't break existing files with forward-compatible fields
3. Include "did you mean?" suggestions using Levenshtein distance
4. Be surfaced in both `config.Load()` (as warnings to stderr) and `pi validate` (as validation errors)

## Acceptance Criteria
- [x] Unknown top-level keys in pi.yaml produce warnings with "did you mean?" suggestions
- [x] `pi validate` surfaces unknown pi.yaml fields
- [x] Tests cover: valid file, unknown field, suggestion match, multiple unknown fields
- [x] Architecture doc updated
- [x] All existing tests still pass

## Implementation Notes
### Approach
- Added `UnknownFields` field to `ProjectConfig` to collect unknown top-level keys during parsing
- Used `yaml.Node` based approach: parse pi.yaml as a `yaml.Node` first to detect all keys, then do normal Unmarshal
- Known fields: `project`, `shortcuts`, `setup`, `packages`, `runtimes`
- Levenshtein distance reused from validate package pattern (reimplemented locally to avoid circular deps)
- Warnings are returned as part of `config.Load()` via a new `LoadResult` type that carries both config and warnings
- `pi validate` checks `config.UnknownFields` and reports them

### Decision: Warnings vs Errors
Chose warnings (not hard errors) because:
1. Forward compatibility: a newer pi.yaml format might have fields that an older PI binary doesn't know
2. Consistency with how automation files handle unknown fields (they're validation warnings, not parse errors)

### Decision: Where to surface
- `config.Load()` returns warnings alongside the config — callers can choose to print or ignore
- `pi validate` reports them as validation issues
- Direct `pi run` / `pi setup` print warnings to stderr (don't fail)

## Subtasks
- [x] Add unknown field detection to config package
- [x] Add did-you-mean suggestions
- [x] Add validation check to validate package
- [x] Write unit tests
- [x] Update architecture.md

## Blocked By
