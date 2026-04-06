# Unify RuntimeCommand, runtimeBinary, and defaultVersion

## Type
improvement

## Status
done

## Priority
high

## Project
15-runtime-provider-registry

## Description

Once `RuntimeDescriptor` exists (see `runtime-descriptor-type`), eliminate the three parallel switch statements that duplicate per-runtime facts: `reqcheck.RuntimeCommand`, `runtimes.runtimeBinary`, and `runtimes.defaultVersion`. Replace each with a descriptor lookup.

**Current state — three separate switch/maps doing the same thing:**

`reqcheck/reqcheck.go`:
```go
func RuntimeCommand(name string) string {
    switch name {
    case "python": return "python3"
    case "rust":   return "rustc"
    default:       return name
    }
}
```

`runtimes/runtimes.go`:
```go
func runtimeBinary(name string) string {
    switch name {
    case "python": return "python3"
    case "rust":   return "rustc"
    default:       return name
    }
}

func defaultVersion(runtimeName string) string {
    switch runtimeName {
    case "python": return "3.13"
    case "node":   return "20"
    ...
    }
}
```

**After this task:**
- `reqcheck.RuntimeCommand(name)` → `descriptor.Binary` lookup (or a single shared `RuntimeBinary(name)` helper)
- `runtimes.runtimeBinary(name)` → `descriptor.Binary` lookup
- `runtimes.defaultVersion(name)` → `descriptor.DefaultVersion` lookup
- All three functions become thin wrappers over the descriptor, or are inlined

## Acceptance Criteria
- [x] No switch statement or map literal that maps "python" → "python3" or "rust" → "rustc" outside of the descriptor definition
- [x] No `defaultVersion` switch statement outside of the descriptor definition
- [x] `reqcheck.RuntimeCommand` and `runtimes.runtimeBinary` either share a implementation or are both eliminated
- [x] `go build ./...` passes
- [x] `go test ./...` passes

## Implementation Notes

Completed as part of the `runtime-descriptor-type` task. All three functions now delegate to `runtimeinfo.Binary()` and `runtimeinfo.DefaultVersion()` respectively. The switch statements are eliminated — the data lives solely in `runtimeinfo.Runtimes`.

- `reqcheck.RuntimeCommand(name)` → `runtimeinfo.Binary(name)`
- `runtimes.runtimeBinary(name)` → `runtimeinfo.Binary(name)`
- `runtimes.defaultVersion(name)` → `runtimeinfo.DefaultVersion(name)`

The wrapper functions are kept for backward compatibility but contain no logic.

## Subtasks
- [x] Replace `reqcheck.RuntimeCommand` switch with descriptor lookup
- [x] Replace `runtimes.runtimeBinary` switch with descriptor lookup
- [x] Replace `runtimes.defaultVersion` switch with descriptor lookup
- [x] Remove any now-dead code

## Blocked By
- `runtime-descriptor-type` (completed)
