# Runtime Provider Registry

## Status
todo

## Priority
high

## Description

All knowledge about a supported runtime (python, node, go, rust) is currently scattered across at least three packages with no shared source of truth. Adding a new runtime or changing a runtime's binary name requires coordinated edits in `automation/requirements.go`, `runtimes/runtimes.go`, `reqcheck/reqcheck.go`, and potentially `validate/`. This project introduces a single `RuntimeDescriptor` type that owns all per-runtime facts, and migrates every consumer to derive from it.

## Goals

- A single `internal/runtimes/descriptor.go` (or equivalent) defines the canonical list of supported runtimes once.
- Every place that currently has its own `knownRuntimes` map or per-runtime switch statement consumes the central descriptor instead.
- The `RuntimeCommand` function in `reqcheck` and the `runtimeBinary` function in `runtimes` are unified into one lookup.
- `defaultVersion` and direct-provisioning capability are properties of the descriptor, eliminating switch statements.
- `provisionDirect` either implements or explicitly stubs all four runtimes symmetrically — no silent asymmetry.
- Adding a fifth runtime is a single-file change.

## Background & Context

Three independent `knownRuntimes` maps exist today:
- `automation/requirements.go` — used for `requires:` YAML validation
- `runtimes/runtimes.go` — used by the provisioner
- `validate/unknown_pi_yaml.go` — used by static validation

Additionally, `reqcheck.RuntimeCommand` and `runtimes.runtimeBinary` both independently map "python" → "python3" and "rust" → "rustc". `runtimes.defaultVersion` is a switch over runtime names. `runtimes.provisionDirect` supports only node and python, silently returning an "install mise" error for go and rust — which are full members of the known-runtimes set.

The right fix is a `RuntimeDescriptor` struct that every consumer can import:

```go
type RuntimeDescriptor struct {
    Name           string // "python", "node", "go", "rust"
    Binary         string // "python3", "node", "go", "rustc"
    DefaultVersion string // "3.13", "20", "1.23", "stable"
    DirectDownload bool   // whether provisionDirect supports this runtime
    InstallHint    string // human-readable hint for pi doctor
}
```

A package-level `var Runtimes = []RuntimeDescriptor{...}` replaces all per-runtime maps and switches. Helper functions like `Find(name string) (*RuntimeDescriptor, bool)` give consumers O(1) lookup.

## Scope

### In scope
- `RuntimeDescriptor` type + `Runtimes` slice in `internal/runtimes/`
- Migrate `automation/requirements.go` knownRuntimes to use `runtimes.Runtimes`
- Migrate `runtimes/runtimes.go` knownRuntimes, runtimeBinary, defaultVersion to descriptor
- Migrate `validate/unknown_pi_yaml.go` to use descriptor
- Migrate `reqcheck/reqcheck.go` RuntimeCommand to use descriptor
- Fix `provisionDirect` asymmetry: at minimum return a consistent error; ideally implement Go/Rust direct downloads or delegate cleanly

### Out of scope
- Adding new runtimes beyond the current four
- Changes to how runtimes are provisioned at the product level (that is a separate decision)

## Success Criteria

- `grep -r "knownRuntime"` returns exactly one definition site
- `grep -r "python3\|rustc"` inside `switch` or `map` literals returns zero results (only in the descriptor definition)
- `go build ./...` passes
- `go test ./...` passes
- Adding a hypothetical 5th runtime requires editing only `runtimes/descriptor.go` (plus actual provisioning logic if needed)

## Notes

Be careful about import cycles: `automation` cannot import `runtimes`. Either move the descriptor to a lower-level package (e.g. `internal/runtimeinfo` or expose only name-lists), or keep a thin lookup function in `automation` that maps to a canonical set. The simplest approach: export `KnownRuntimeNames() []string` from `runtimes` and use that in `automation/requirements.go` for validation, keeping all richer fields inside `runtimes`.
