# Create RuntimeDescriptor and Consolidate Known Runtimes

## Type
improvement

## Status
done

## Priority
high

## Project
15-runtime-provider-registry

## Description

Create `internal/runtimes/descriptor.go` with a `RuntimeDescriptor` struct and a `Runtimes` slice that becomes the single source of truth for all runtime knowledge. Migrate the three independent `knownRuntimes` maps (in `automation/requirements.go`, `runtimes/runtimes.go`, and `validate/unknown_pi_yaml.go`) to derive from this single definition.

**New file `internal/runtimes/descriptor.go`:**
```go
// RuntimeDescriptor describes everything PI knows about a supported runtime.
type RuntimeDescriptor struct {
    Name           string // "python", "node", "go", "rust"
    Binary         string // "python3", "node", "go", "rustc"
    DefaultVersion string // "3.13", "20", "1.23", "stable"
    DirectDownload bool   // whether provisionDirect supports this runtime natively
    InstallHint    string // human-readable hint for pi doctor
}

var Runtimes = []RuntimeDescriptor{
    {Name: "python", Binary: "python3", DefaultVersion: "3.13", DirectDownload: true,
     InstallHint: "brew install python3  or  https://www.python.org/downloads/"},
    {Name: "node",   Binary: "node",    DefaultVersion: "20",   DirectDownload: true,
     InstallHint: "brew install node  or  https://nodejs.org/"},
    {Name: "go",     Binary: "go",      DefaultVersion: "1.23", DirectDownload: false,
     InstallHint: "brew install go  or  https://go.dev/dl/"},
    {Name: "rust",   Binary: "rustc",   DefaultVersion: "stable", DirectDownload: false,
     InstallHint: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"},
}

// Find returns the RuntimeDescriptor for the given name, or (nil, false) if unknown.
func Find(name string) (*RuntimeDescriptor, bool) { ... }

// KnownNames returns the set of known runtime names (for validation use).
func KnownNames() map[string]bool { ... }
```

**Migration targets:**
1. `runtimes/runtimes.go` — replace `KnownRuntimes` map with `runtimes.KnownNames()`
2. `automation/requirements.go` — replace `knownRuntimes` with a call to `runtimes.KnownNames()` (note: avoid import cycle — `automation` importing `runtimes` is new; check if this creates a cycle)
3. `validate/unknown_pi_yaml.go` — replace its copy with `runtimes.KnownNames()`

**Import cycle note:** If `automation` cannot import `runtimes`, expose the names via a thin `internal/runtimeinfo` package that neither imports the other, or export `KnownNames()` as a package-level var that `runtimes` populates at init. The simplest safe approach: move the descriptor to a new `internal/runtimeinfo` package that `automation`, `runtimes`, `reqcheck`, and `validate` all import.

## Acceptance Criteria
- [x] `RuntimeDescriptor` struct and `Runtimes` slice defined in one place
- [x] `automation/requirements.go` derives its known-runtime list from the shared definition
- [x] `runtimes/runtimes.go` derives its known-runtime set from the shared definition
- [x] `validate/unknown_pi_yaml.go` — N/A: its `knownRuntimesKeys` is about pi.yaml config fields (`provision`, `manager`), not runtime names
- [x] `grep -r 'knownRuntime' --include="*.go"` returns only derived variable declarations (no duplicate map literals)
- [x] `go build ./...` passes
- [x] `go test ./...` passes

## Implementation Notes

### Package placement: `internal/runtimeinfo`
Created a new leaf package `internal/runtimeinfo` instead of putting the descriptor in `internal/runtimes`. This avoids import cycles because:
- `automation` has zero internal imports (leaf package) — it can import `runtimeinfo`
- `runtimes` imports `config` — it can also import `runtimeinfo`
- `reqcheck` imports `automation` and `conditions` — it can import `runtimeinfo`

Placing the descriptor in `internal/runtimes` would have been impossible because `automation` cannot import `runtimes` (many packages import `automation`, creating a cycle).

### `validate/unknown_pi_yaml.go` — not migrated
On investigation, `knownRuntimesKeys` in `validate/unknown_pi_yaml.go` contains `{"provision": true, "manager": true}` — these are pi.yaml config block field names, not runtime names. This is a different concept entirely and does not need migration.

### Additional helpers: `Binary()` and `DefaultVersion()`
Added convenience functions `runtimeinfo.Binary(name)` and `runtimeinfo.DefaultVersion(name)` that both `runtimes` and `reqcheck` now use. This eliminates the duplicated switch statements for name→binary mapping (`runtimeBinary` and `RuntimeCommand`) and the `defaultVersion` switch.

### What remains as thin wrappers
- `runtimes.defaultVersion()` → delegates to `runtimeinfo.DefaultVersion()`
- `runtimes.runtimeBinary()` → delegates to `runtimeinfo.Binary()`
- `reqcheck.RuntimeCommand()` → delegates to `runtimeinfo.Binary()`

These wrappers are kept for backward compatibility (callers reference them), but the actual data lives solely in `runtimeinfo.Runtimes`.

### `executor/runners.go` `resolvePythonBin()` — not migrated
This function has VIRTUAL_ENV special-case logic (returns `$VIRTUAL_ENV/bin/python` when set). It's about step execution, not runtime identity. Left as-is; a future task could wire it through the descriptor if desired.

## Subtasks
- [x] Create `RuntimeDescriptor` in `internal/runtimeinfo` package (checked for import cycles)
- [x] Export `KnownNames()`, `SortedNames()`, `Binary()`, `DefaultVersion()`, `Find()` helpers
- [x] Migrate `automation/requirements.go`
- [x] Migrate `runtimes/runtimes.go` (KnownRuntimes, knownRuntimeList, defaultVersion, runtimeBinary)
- [x] Migrate `reqcheck/reqcheck.go` (RuntimeCommand)
- [x] Write tests for `Find()`, `KnownNames()`, `SortedNames()`, `Binary()`, `DefaultVersion()`
- [x] Fix `runtimes_test.go` to reference `runtimeinfo.SortedNames()` instead of removed `knownRuntimeList()`

## Blocked By
