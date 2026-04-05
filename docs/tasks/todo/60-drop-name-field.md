# Drop Required `name:` Field from Automation Files

## Type
improvement

## Status
todo

## Priority
high

## Project
12-yaml-ergonomics

## Description
The `name:` field at the top of every automation YAML is redundant — it always mirrors the file path. An automation at `.pi/build/default.yaml` always has `name: build/default`. PI already knows the file path when loading the automation, so it can derive the name without the author spelling it out.

Make `name:` optional. When absent, PI derives the name from the file path: strip the `.pi/` prefix, strip `.yaml` suffix, and collapse `automation.yaml` to its parent folder name (e.g. `.pi/setup/install-cursor-extensions/automation.yaml` → `setup/install-cursor-extensions`).

When `name:` is present and matches the derived path, accept silently. When present and *mismatches*, emit a parse-time warning (but don't fail). This lets existing files keep working while the field gradually gets dropped.

## Acceptance Criteria
- [ ] `name:` is optional in all automation YAML files
- [ ] When absent, PI derives the automation name from the file path using the rules above
- [ ] When present and matching, no warning emitted
- [ ] When present and mismatching the derived name, a warning is emitted at parse time
- [ ] All existing automation files with `name:` still load correctly
- [ ] `pi info`, `pi list`, `pi run` all use the derived name correctly when `name:` is absent
- [ ] Tests cover: absent name, present-matching name, present-mismatching name
- [ ] At least one example automation updated to omit `name:` (to serve as a canonical example)

## Implementation Notes

## Subtasks
- [ ] Update automation struct and parser to make `name` optional
- [ ] Implement path-to-name derivation logic
- [ ] Add mismatch warning
- [ ] Update at least one example to omit `name:`
- [ ] Add tests

## Blocked By
