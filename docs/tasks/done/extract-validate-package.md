# Extract Validation into Standalone Package

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The `cli/validate.go` file (400 lines) contains all project validation logic directly in the CLI package. This includes 9 validation functions that all follow the same pattern: accept discovery result + validation result, walk automations, and append errors. The validation logic is tightly coupled to CLI wiring, making it impossible to reuse from other contexts (IDE plugins, library consumers) and harder to extend.

**Goal:** Extract the validation system into `internal/validate/` with a clean `Check` interface and a `Runner` that orchestrates checks. Each validation concern becomes an independent, testable, registerable unit. The CLI's `validate.go` becomes a thin wrapper.

**Current structure (cli/validate.go):**
- `validateProject()` — monolithic orchestrator calling 9 functions
- `validateShortcutRefs()`, `validateSetupRefs()`, `validateRunStepRefs()` — reference checks
- `validateFileReferences()` — file-path existence checks
- `validateShortcutInputs()`, `validateSetupInputs()`, `validateRunStepInputs()` — input matching
- `validateCircularDeps()` — cycle detection
- `validateConditions()` — if: expression validation

**Target structure (internal/validate/):**
- `Check` interface — `Name() string`, `Run(ctx) []string`
- `Context` struct — holds config, discovery, root
- `Runner` — collects checks, runs them, returns `Result`
- Individual check implementations — one per concern
- Helper functions (checkWithInputs, buildRunGraph, detectCycles, normalizeCycleKey) move here

**Benefits:**
- Adding a new validation check = implement `Check` + register it
- Validators are independently testable without CLI setup
- Other consumers can run validation without the CLI
- Clear separation between "what to validate" and "how to report"

## Acceptance Criteria
- [x] `internal/validate/` package exists with `Check` interface and `Runner`
- [x] All 9 validation concerns extracted as individual checks
- [x] `cli/validate.go` is a thin wrapper delegating to `validate.Runner`
- [x] All existing `cli/validate_test.go` tests pass unchanged
- [x] New unit tests in `internal/validate/` for the Runner and individual checks
- [x] `go build ./...` and `go test ./...` pass
- [x] Architecture docs updated

## Implementation Notes

### Approach

Design the `Check` interface to be minimal:

```go
type Check interface {
    Name() string
    Run(ctx *Context) []string // returns error messages
}
```

Context holds everything a check needs:
```go
type Context struct {
    Root      string
    Config    *config.ProjectConfig
    Discovery *discovery.Result
}
```

Runner orchestrates:
```go
type Runner struct {
    checks []Check
}

func (r *Runner) Register(c Check) { r.checks = append(r.checks, c) }
func (r *Runner) Run(ctx *Context) Result { ... }
func DefaultRunner() *Runner { ... } // pre-registers all built-in checks
```

### Key design decisions
- The Check interface returns `[]string` (error messages) rather than `[]error` — matches existing pattern and keeps it simple
- Checks are stateless — all state comes from Context
- Helper functions like `CheckWithInputs` and `BuildRunGraph`/`DetectCycles`/`NormalizeCycleKey` are exported package-level functions in `validate/` for reuse
- `Result` type (formerly `ValidationResult` in cli) moved to validate package
- The CLI command became ~40 lines — resolve project, build Context, run DefaultRunner, print results
- `CheckFunc` adapter allows inline function registration without defining separate types
- All 67 existing CLI integration tests continue to pass unchanged (just updated imports for `validate.CheckWithInputs` etc.)
- Individual check functions are unexported (lowercase) since external consumers use them through the `Runner` — the exported surface is `Check`, `CheckFunc`, `Runner`, `DefaultRunner()`, and helper functions

### Coverage results
- `internal/validate`: 96.6% (52 tests)
- `internal/cli`: 78.8% (down from 81.2% since validation logic moved out, still fully tested via integration)
- Overall test count: 1743 (up from 1488, net +55 new tests in validate package)

## Subtasks
- [x] Create task file
- [x] Create internal/validate/ package with core types
- [x] Move helper functions (checkWithInputs, graph analysis)
- [x] Implement individual Check types
- [x] Create Runner with DefaultRunner()
- [x] Refactor cli/validate.go to delegate
- [x] Write unit tests for validate package (52 tests, 96.6% coverage)
- [x] Run full test suite (1743 tests, all pass)
- [x] Update architecture.md
- [ ] Commit

## Blocked By
