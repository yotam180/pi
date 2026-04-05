# `setup:` Bare Path Shorthand in `pi.yaml`

## Type
improvement

## Status
done

## Priority
medium

## Project
12-yaml-ergonomics

## Description
`setup:` entries in `pi.yaml` currently require the `run:` key even though every entry is always a run:

```yaml
setup:
  - run: setup/install-go
  - run: setup/install-ruby
    if: os.macos
```

Since all entries are runs, the `run:` key is redundant. Allow bare strings as shorthand:

```yaml
setup:
  - setup/install-go
  - run: setup/install-ruby
    if: os.macos
```

Bare strings (no `if:`, no `with:`) become simple run entries. The expanded object form (`run:` + optional `if:` + optional `with:`) remains valid for entries that need modifiers. Both forms work in the same list.

## Acceptance Criteria
- [x] A bare string entry in `setup:` is treated as `run: <string>` with no conditions
- [x] Object entries with `run:` continue to work as before
- [x] Mixed lists (some bare, some object) parse correctly
- [x] `pi setup` executes bare entries correctly
- [x] Bare entries with `if:` are NOT valid (YAML syntax prevents it naturally — no special error needed)
- [x] Tests cover: bare string, object with if, mixed list, empty bare string, builtin bare path
- [x] `pi.yaml` in this repo and example `pi.yaml` files updated to use bare paths where possible

## Implementation Notes
- Added `UnmarshalYAML` to `SetupEntry` (same pattern as `Shortcut`): scalar → bare, mapping → object
- YAML naturally prevents bare+if — you can't put mapping keys on a scalar list entry, so no special error handling is needed
- 4 new tests added to config_test.go: `TestLoad_SetupBareString`, `TestLoad_SetupMixedBareAndObject`, `TestLoad_SetupBareStringEmpty`, `TestLoad_SetupBareStringWithBuiltin`
- All 7 pi.yaml files updated to use bare paths where entries have no `if:` or `with:` modifiers
- Backward compatible — existing `- run:` syntax continues to work unchanged
- Total config tests: 21 (was 17)

## Subtasks
- [x] Update `pi.yaml` setup entry parser — added `UnmarshalYAML` to `SetupEntry`
- [x] Add tests — 4 new unit tests
- [x] Update example and repo `pi.yaml` files — all 7 files updated
- [x] Update docs/README.md with bare path syntax examples
- [x] Update docs/architecture.md with setup entry syntax section

## Blocked By
