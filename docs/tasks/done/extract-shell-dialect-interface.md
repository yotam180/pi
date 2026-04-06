# Extract Shell Dialect Interface for Multi-Shell Extensibility

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The `internal/shell/shell.go` file hardcodes bash/zsh syntax for shell function generation. The eval-file wrapper pattern (mktemp → run → capture exit → source eval file → cleanup → return) is duplicated 4 times across `generateFunction`, `generateFunctionWithInputs`, `generateSetupHelper`, and `GenerateGlobalWrapper`. Each duplicate is nearly identical — only the inner command varies.

Additionally, adding support for new shells (fish, powershell) requires duplicating each of these generation functions with shell-specific syntax. The `GenerateCompletionScript` function similarly only handles zsh and bash detection.

**This refactor:**
1. Extracts the eval-file wrapper into a reusable `evalWrapper(funcName, innerCmd)` helper, eliminating 4× duplication in bash/zsh generation
2. Introduces a `ShellDialect` interface that encapsulates shell-specific syntax differences (function definition, subshell, eval, variable capture, arg forwarding)
3. Implements `bashDialect` as the first (and currently only) concrete implementation
4. Makes future fish/powershell shell support a matter of implementing a new dialect — not duplicating generation logic

This directly prepares the codebase for the fish-shell-integration task while reducing current duplication.

## Acceptance Criteria
- [x] Eval-file wrapper pattern is defined once and reused for all 4 function types
- [x] `ShellDialect` interface defined with methods for shell-specific syntax
- [x] `BashDialect` implements the interface
- [x] `GenerateShellFile` and `GenerateGlobalWrapper` use the dialect
- [x] All 20 existing shell tests pass unchanged
- [x] No behavior change in generated shell file content
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach
- Created `dialect.go` with `ShellDialect` interface (5 methods: `EvalWrapperFunc`, `InRepoCmd`, `AnywhereCmd`, `AllArgs`, `FileHeader`)
- `BashDialect` struct implements the interface — the eval-file wrapper template lives in `EvalWrapperFunc` as the single source of truth
- `generateFunction`, `generateSetupHelper`, and `GenerateGlobalWrapper` all delegate to the dialect instead of duplicating the wrapper pattern
- Extracted `buildWithArgs()` as a standalone helper (previously embedded in `generateFunctionWithInputs`)
- Added `GenerateShellFileWithDialect()` and `GenerateGlobalWrapperWithDialect()` as the dialect-aware entry points; the existing `GenerateShellFile()` and `GenerateGlobalWrapper()` delegate with `DefaultDialect()`
- No public API changes — all callers see identical behavior

### What was deduplicated
Before: 4 copies of the eval wrapper (mktemp → run → capture exit → source → cleanup → return) in `generateFunction`, `generateFunctionWithInputs`, `generateSetupHelper`, `GenerateGlobalWrapper`
After: 1 copy in `BashDialect.EvalWrapperFunc()`, called from all 4 sites

### Testing
- All 20 original `shell_test.go` tests pass unchanged, verifying behavioral equivalence
- Added 16 new tests in `dialect_test.go`:
  - `BashDialect` method tests (EvalWrapperFunc, InRepoCmd, AnywhereCmd, AllArgs, FileHeader)
  - DefaultDialect type assertion
  - Mock dialect integration (plugs into `GenerateShellFileWithDialect` and `GenerateGlobalWrapperWithDialect`)
  - Edge cases (paths with spaces, complex inner commands, empty shortcuts)
  - `buildWithArgs` unit tests (positional refs, literals, mixed, sort order)
- Shell package coverage: 84.7% → 85.8%

### Why this matters
Adding fish shell support (the existing `fish-shell-integration` task) now requires implementing one interface (`ShellDialect`) rather than duplicating 4 function generators. The mock dialect test proves the interface is pluggable.

## Subtasks
- [x] Extract `evalWrapper` helper to deduplicate the 4 shell function templates
- [x] Define `ShellDialect` interface
- [x] Implement `BashDialect`
- [x] Wire dialect into generation functions
- [x] Verify all existing tests pass
- [x] Add tests for dialect interface

## Blocked By
