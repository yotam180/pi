# setup add help text should list rust alongside python/node/go

## Type
improvement

## Status
todo

## Priority
low

## Project
standalone

## Description
The `pi setup add --help` text shows automatic resolution examples for python, node, and go:

```
Tool names are resolved automatically:
  python  →  pi:install-python
  node    →  pi:install-node
  go      →  pi:install-go
```

But `rust` is missing from this list, even though `pi setup add rust` does resolve to `pi:install-rust` correctly. A user might assume rust isn't supported and write a custom installer instead.

### Fix
Add rust to the help text:
```
Tool names are resolved automatically:
  python  →  pi:install-python
  node    →  pi:install-node
  go      →  pi:install-go
  rust    →  pi:install-rust
```

Ideally, this list should be auto-generated from whatever builtins exist with the `install-*` naming pattern, so it stays in sync as new installers are added.

## Acceptance Criteria
- [ ] `pi setup add --help` lists rust in the tool name resolution examples
- [ ] The list is either auto-generated or there's a test ensuring it matches available builtins

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
