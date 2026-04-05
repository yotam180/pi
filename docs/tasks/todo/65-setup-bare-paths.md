# `setup:` Bare Path Shorthand in `pi.yaml`

## Type
improvement

## Status
todo

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
  - setup/install-ruby
    if: os.macos
```

Wait — YAML doesn't support inline keys on a scalar list entry. The `if:` modifier needs a key. So the shorthand should be:

```yaml
setup:
  - setup/install-go
  - run: setup/install-ruby
    if: os.macos
```

Bare strings (no `if:`, no `with:`) become simple run entries. The expanded object form (`run:` + optional `if:` + optional `with:`) remains valid for entries that need modifiers. Both forms work in the same list.

## Acceptance Criteria
- [ ] A bare string entry in `setup:` is treated as `run: <string>` with no conditions
- [ ] Object entries with `run:` continue to work as before
- [ ] Mixed lists (some bare, some object) parse correctly
- [ ] `pi setup` executes bare entries correctly
- [ ] Bare entries with `if:` are NOT valid (clear parse error message pointing to object form)
- [ ] Tests cover: bare string, object with if, mixed list, invalid bare+if attempt
- [ ] `pi.yaml` in this repo and example `pi.yaml` files updated to use bare paths where possible

## Implementation Notes

## Subtasks
- [ ] Update `pi.yaml` setup entry parser
- [ ] Add tests
- [ ] Update example and repo `pi.yaml` files

## Blocked By
