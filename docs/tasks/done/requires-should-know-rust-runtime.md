# requires: should recognize rust and go as known runtimes

## Type
feature

## Status
done

## Priority
medium

## Project
standalone

## Description
The `requires:` field in automation YAML supports `python` and `node` as known runtimes, but not `rust` or `go` — even though `pi:install-rust` and `pi:install-go` exist as builtin installers.

### Current Behavior
```yaml
requires:
  - rust    # → error: unknown runtime "rust" (known: python, node)
  - go      # → error: unknown runtime "go" (known: python, node)
```
Workaround:
```yaml
requires:
  - command: rustc
  - command: go
```

### Expected
```yaml
requires:
  - rust    # Works! Checks for rustc in PATH.
  - go      # Works! Checks for go in PATH.
```

## PM Concern — Addressed

The PM questioned whether Rust and Go should be "known runtimes" since users don't write PI steps in those languages (only bash, python, typescript are step types).

**Decision: Yes, they should be known runtimes.** The reasoning:

1. **`requires:` declares dependencies, not step languages.** A bash step that runs `go build ./...` needs Go installed. A bash step that runs `cargo build` needs Rust. The `requires:` field exists to declare these dependencies — it's orthogonal to which languages PI supports for inline steps.

2. **Consistency with builtins.** PI ships `pi:install-go` and `pi:install-rust`. If PI knows how to install them, PI should recognize them in `requires:`. Having an installer but not accepting the name in `requires:` is a paper cut.

3. **User intent is clear.** When a user writes `requires: [go]`, their intent is unambiguous. Making them write `command: go` instead is ceremony without benefit.

4. **The alternative (`command: go`) still works.** Adding Go/Rust as known runtimes doesn't remove any capability — it adds a more readable shorthand.

The `knownRuntimes` map is about "runtimes PI recognizes by name for dependency declaration," not "languages you can write steps in." The step type system (`bash:`, `python:`, `typescript:`) is a separate concept.

## Acceptance Criteria
- [x] `requires: [rust]` is accepted and checks for `rustc` in PATH
- [x] `requires: [go]` is accepted and checks for `go` in PATH
- [x] `pi doctor` reports rust/go version correctly when using `requires: [rust]` / `requires: [go]`
- [x] Known runtimes list in error messages is auto-generated and sorted
- [x] All existing tests pass, new tests cover go and rust requirements

## Implementation Notes

### Changes made
1. **`internal/automation/requirements.go`** — Added `go` and `rust` to `knownRuntimes` map. Added `knownRuntimeNames()` helper that generates a sorted list dynamically instead of hardcoding "python, node" in error messages.

2. **`internal/runtimes/runtimes.go`** — Added `go` and `rust` to `KnownRuntimes` map, `defaultVersion()` (go=1.23, rust=stable), `runtimeBinary()` (rust→rustc, go→go), and `provisionDirect()` (explicit "not supported" message pointing to mise). Added `knownRuntimeList()` for dynamic error messages.

3. **`internal/executor/validate.go`** — Added `rust` → `rustc` mapping in `runtimeCommand()`. Added install hints for `rust`, `go`, `rustc`, `cargo`, `rustup`.

4. **`internal/cli/setup_add.go`** — Auto-generated tool resolution help text via `setupAddToolResolutionHelp()` so the help text stays in sync as new builtins are added.

### Test coverage
- `requirements_test.go`: Extended to test all 4 known runtimes (python, node, go, rust) in bare and versioned forms. Added assertion that error messages list all known runtimes.
- `validate_test.go`: Added `TestCheckRequirement_RustRuntimeFound/NotFound`, `TestCheckRequirement_RustVersionSatisfied`, `TestCheckRequirement_GoRuntimeFound`, `TestCheckRequirement_GoVersionSatisfied`, `TestRuntimeCommand` table-driven refactor, install hint tests for rust/go.
- `runtimes_test.go`: Extended `TestDefaultVersion`, `TestRuntimeBinary`, `TestKnownRuntimes` to cover all 4 runtimes.

## Subtasks
- [x] Add go and rust to knownRuntimes in requirements.go
- [x] Add go and rust to KnownRuntimes in runtimes.go
- [x] Add rust→rustc mapping in executor/validate.go runtimeCommand()
- [x] Add install hints for rust/go
- [x] Make error message runtime list dynamic
- [x] Write tests for all new paths
- [x] Auto-generate setup add help text

## Blocked By
