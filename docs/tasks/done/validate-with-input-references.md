# Validate `with:` Input References in `pi validate`

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `pi validate` command checks that shortcut targets, setup entry targets, and `run:` step targets reference existing automations. It also checks that file-path step values reference existing files.

However, it does NOT validate `with:` input references — if a shortcut, setup entry, or `run:` step passes `with: { vrsion: "3.13" }` (typo) to an automation that declares `version` as its input, this is silently accepted and only fails at runtime.

This task adds three new validation checks to `pi validate`:

1. **Shortcut `with:` keys** — check that each key in a shortcut's `with:` map matches a declared input on the target automation
2. **Setup entry `with:` keys** — check that each key in a setup entry's `with:` map matches a declared input on the target automation
3. **`run:` step `with:` keys** — check that each key in a `run:` step's `with:` map matches a declared input on the target automation (skipping values that are `outputs.*` or `inputs.*` interpolation references, since those are runtime-resolved)

The validation checks `with:` **keys** (which map to input names), not `with:` **values** (which can be runtime-interpolated via `outputs.last`, etc.).

Automations without declared inputs that receive `with:` keys should also be flagged — they'll silently ignore the values at runtime.

## Acceptance Criteria
- [x] `pi validate` reports an error when a shortcut's `with:` key doesn't match a declared input
- [x] `pi validate` reports an error when a setup entry's `with:` key doesn't match a declared input
- [x] `pi validate` reports an error when a `run:` step's `with:` key doesn't match a declared input
- [x] `pi validate` reports an error when `with:` is used on a target that has no declared inputs
- [x] Built-in automations are properly resolved for input checking
- [x] Unit tests cover all new validation paths
- [x] Integration test validates the new checks end-to-end
- [x] `go build ./...` and `go test ./...` pass
- [x] Architecture docs updated

## Implementation Notes

Key design decisions:
- Only validate `with:` **keys** against declared input names. Values can be `outputs.last` etc.
- Automations without inputs receiving `with:` keys → error (they'd be silently ignored)
- `run:` steps whose target is not resolvable are already caught by `validateRunStepRefs()` — skip them here
- Built-in targets (e.g. `pi:install-python`) are resolved via the discovery result

The new validation functions follow the same pattern as existing ones:
- `validateShortcutInputs(cfg, disc, &result)` 
- `validateSetupInputs(cfg, disc, &result)`
- `validateRunStepInputs(disc, &result)`

## Subtasks
- [x] Implement `validateShortcutInputs()`
- [x] Implement `validateSetupInputs()`
- [x] Implement `validateRunStepInputs()`
- [x] Extract shared `checkWithInputs()` helper
- [x] Add unit tests
- [x] Add integration test
- [x] Update architecture doc

## Blocked By
