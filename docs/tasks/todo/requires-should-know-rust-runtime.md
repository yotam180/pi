# requires: should recognize rust as a known runtime

## Type
feature

## Status
todo

## Priority
medium

## Project
standalone

## Description
The `requires:` field in automation YAML supports `python` and `node` as known runtimes, but not `rust` — even though `pi:install-rust` exists as a builtin installer.

### Current Behavior
```yaml
requires:
  - rust    # → error: unknown runtime "rust" (known: python, node)
```
Workaround:
```yaml
requires:
  - command: rustc
```

### Expected
```yaml
requires:
  - rust    # Works! Checks for rustc in PATH.
```

This should work the same way python and node do. Having `install-rust` as a builtin but not recognizing `rust` in `requires:` is inconsistent.

### Suggested Implementation
Add `rust` to the known runtimes map:
- Runtime name: `rust`
- Version command: `rustc --version` (already used by install-rust)
- Binary check: `rustc`

Any language/runtime with a builtin installer should automatically be a known runtime for `requires:`.

## Acceptance Criteria
- [ ] `requires: [rust]` is accepted and checks for `rustc` in PATH
- [ ] `pi doctor` reports rust version correctly when using `requires: [rust]`
- [ ] Go is also a valid known runtime (since install-go exists)

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
