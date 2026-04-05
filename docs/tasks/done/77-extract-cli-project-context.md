# Extract CLI ProjectContext to Eliminate Boilerplate

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

Every CLI command repeats the same project resolution pipeline:
1. `os.Getwd()` → `project.FindRoot()` → `config.Load()` → `discoverAllWithConfig()`
2. Create executor with RepoRoot, Discovery, Stdout, Stderr, Silent, Loud, ParentEvalFile, Provisioner

This boilerplate is duplicated across `run.go`, `setup.go`, `list.go`, `info.go`, `doctor.go`, `validate.go`, `add.go`. There's also a `loadProjectConfig()` helper in `shell.go` that partially deduplicates root+config but isn't used outside `shell.go`.

**Goal**: Extract a `ProjectContext` type that encapsulates the common resolution pipeline and provides a single method to build an `Executor`. This:
- Eliminates duplicated boilerplate
- Makes adding new CLI commands trivial (just call `resolveProject()`)
- Provides a single extension point for cross-cutting concerns
- Reduces parameter passing in helper functions

## Acceptance Criteria
- [x] `ProjectContext` type exists in `internal/cli/context.go`
- [x] All CLI commands use `resolveProject()` (or similar) to get a `ProjectContext`
- [x] `loadProjectConfig()` in `shell.go` replaced by `ProjectContext`
- [x] All existing tests pass without modification
- [x] No functional behavior changes — pure refactor
- [x] `go vet ./...` clean

## Implementation Notes

**Approach**: Create a `ProjectContext` struct holding `Root`, `Config`, `Discovery`, plus a `resolveProject(startDir)` constructor and a `newExecutor(opts)` builder. Some commands need different levels of resolution:

- `shell list` doesn't need project root at all
- `shell install/uninstall` need root + config (no discovery)
- `list`, `info`, `doctor`, `validate` need root + config + discovery
- `run`, `setup` need root + config + discovery + executor

So `ProjectContext` should support incremental resolution:
- `resolveProject(startDir)` → Root + Config (config errors ignored for commands that don't need it, like `run`)
- `resolveProjectStrict(startDir)` → Root + Config (config errors fatal)
- `(pc *ProjectContext).Discover(stderr)` → populates Discovery
- `(pc *ProjectContext).NewExecutor(opts)` → builds Executor

Actually, looking at the pattern more carefully:
- `run.go` ignores config load errors: `cfg, _ := config.Load(root)`
- `setup.go` requires config: `cfg, err := config.Load(root)`
- `doctor.go`, `list.go`, `info.go` ignore config errors
- `validate.go` treats config errors as validation errors (different handling)

The best approach is to keep `resolveProject` simple — it always resolves root + attempts config load — and let the caller decide whether to check `pc.ConfigErr`. This avoids over-engineering.

Final design:
- `ProjectContext` with `Root string`, `Config *config.ProjectConfig`
- `resolveProject(startDir string) (*ProjectContext, error)` — finds root, loads config (ignoring error)
- `resolveProjectStrict(startDir string) (*ProjectContext, error)` — finds root, loads config (fatal error)
- `(pc *ProjectContext).Discover(stderr io.Writer) (*discovery.Result, error)` — runs discovery
- `ExecutorOpts` for executor construction parameters

After implementation: commands that currently have 10-15 lines of boilerplate become 2-3 lines.

## Subtasks
- [x] Create `internal/cli/context.go` with `ProjectContext`
- [x] Refactor `run.go` to use it
- [x] Refactor `list.go`, `info.go`, `doctor.go`
- [x] Refactor `setup.go`, `validate.go`, `add.go`
- [x] Replace `loadProjectConfig()` in `shell.go`
- [x] Run full test suite

## Blocked By
