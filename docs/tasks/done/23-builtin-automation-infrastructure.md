# Built-in Automation Infrastructure

## Type
feature

## Status
done

## Priority
high

## Project
03-built-in-library-and-pi-setup

## Description
Add the infrastructure to embed automations from the PI repo's own `.pi/` folder into the binary at build time, and resolve them when referenced with the `pi:` prefix. This is the foundation for all built-in automations — no individual built-in should be implemented until this task is done.

Currently, `discovery.Discover()` only walks a local `.pi/` directory. After this task, `discovery.Result.Find()` should also check a built-in registry when the name starts with `pi:`. The built-in automations are loaded from Go `//go:embed` at init time.

## Acceptance Criteria
- [x] A `builtins/` directory exists in the repo root containing `.pi/` automation YAML files that define built-in automations
- [x] `//go:embed` embeds the `builtins/.pi/` directory into the binary at build time
- [x] A new package `internal/builtins` provides `Discover() *discovery.Result` that returns all embedded automations
- [x] `discovery.Result.Find("pi:docker/up")` resolves to the embedded automation named `docker/up`
- [x] Local automations take precedence — if `.pi/docker/up.yaml` exists locally AND `pi:docker/up` is embedded, `Find("docker/up")` returns the local one; `Find("pi:docker/up")` always returns the built-in
- [x] `pi list` shows built-in automations (marked with a `[built-in]` indicator or similar)
- [x] `pi run pi:docker/up` executes the embedded automation
- [x] `pi info pi:docker/up` shows info for the embedded automation
- [x] All existing tests pass; new unit tests cover the embedding, resolution, and precedence
- [x] Integration test: a workspace that references `pi:` automations in its `setup:` block

## Implementation Notes

### Architecture
- Built-in automation YAML files live in `internal/builtins/embed_pi/` (embedded via `//go:embed all:embed_pi`)
- The `internal/builtins` package provides `Discover() *discovery.Result` which walks the embedded FS and parses automation YAMLs using `automation.LoadFromBytes()`
- `discovery.Result` gained new fields: `Builtins` map, `builtinSet` tracking, and methods `MergeBuiltins()`, `IsBuiltin()`, `NewResult()`
- `Find()` handles `pi:` prefix: `Find("pi:hello")` always resolves from built-ins; `Find("hello")` checks local first, then built-in fallback
- A shared `discoverAll()` helper in `internal/cli/discover.go` discovers local + built-in and merges, used by run, list, info, and setup commands
- `pi list` shows `[built-in]` marker for automations sourced from built-ins (not shadowed by local)
- `automation.LoadFromBytes()` added to support parsing YAML from embedded bytes

### Key design decisions
- Built-in files live inside `internal/builtins/embed_pi/` rather than a separate `builtins/` at repo root, because Go's `//go:embed` is relative to the source file's package
- `MergeBuiltins()` adds built-ins to the main `Automations` map for seamless resolution, but tracks origin in `builtinSet`
- Local always takes precedence when names collide — the built-in is still accessible via explicit `pi:` prefix

### Test coverage
- 3 unit tests in `internal/builtins/` (discovery, sorted names, Find integration)
- 7 new unit tests in `internal/discovery/` (MergeBuiltins add/precedence/nil, IsBuiltin, Find with pi: prefix, pi: not found)
- 9 integration tests covering: list marker, run with pi: prefix, local shadow, pi: always gets built-in, run-step calling built-in, info with pi:, list shadowed, setup with pi:, pi: not found error

## Subtasks
- [x] Create `builtins/.pi/` directory with a simple test automation (e.g., `builtins/.pi/hello.yaml` that echoes "hello from built-in")
- [x] Create `internal/builtins` package with `//go:embed` and `Discover()` function
- [x] Modify `discovery.Result` to accept and merge built-in automations (or create a composite `Find` that checks both)
- [x] Update `cli/run.go`, `cli/list.go`, `cli/info.go` to pass built-ins through to discovery
- [x] Add `[built-in]` indicator to `pi list` output for built-in automations
- [x] Write unit tests for built-in discovery, resolution, and local-takes-precedence
- [x] Write integration test with a `pi:` reference

## Blocked By
<!-- None -->
