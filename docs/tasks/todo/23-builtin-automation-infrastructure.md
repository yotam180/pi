# Built-in Automation Infrastructure

## Type
feature

## Status
todo

## Priority
high

## Project
03-built-in-library-and-pi-setup

## Description
Add the infrastructure to embed automations from the PI repo's own `.pi/` folder into the binary at build time, and resolve them when referenced with the `pi:` prefix. This is the foundation for all built-in automations — no individual built-in should be implemented until this task is done.

Currently, `discovery.Discover()` only walks a local `.pi/` directory. After this task, `discovery.Result.Find()` should also check a built-in registry when the name starts with `pi:`. The built-in automations are loaded from Go `//go:embed` at init time.

## Acceptance Criteria
- [ ] A `builtins/` directory exists in the repo root containing `.pi/` automation YAML files that define built-in automations
- [ ] `//go:embed` embeds the `builtins/.pi/` directory into the binary at build time
- [ ] A new package `internal/builtins` provides `Discover() *discovery.Result` that returns all embedded automations
- [ ] `discovery.Result.Find("pi:docker/up")` resolves to the embedded automation named `docker/up`
- [ ] Local automations take precedence — if `.pi/docker/up.yaml` exists locally AND `pi:docker/up` is embedded, `Find("docker/up")` returns the local one; `Find("pi:docker/up")` always returns the built-in
- [ ] `pi list` shows built-in automations (marked with a `[built-in]` indicator or similar)
- [ ] `pi run pi:docker/up` executes the embedded automation
- [ ] `pi info pi:docker/up` shows info for the embedded automation
- [ ] All existing tests pass; new unit tests cover the embedding, resolution, and precedence
- [ ] Integration test: a workspace that references `pi:` automations in its `setup:` block

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `builtins/.pi/` directory with a simple test automation (e.g., `builtins/.pi/hello.yaml` that echoes "hello from built-in")
- [ ] Create `internal/builtins` package with `//go:embed` and `Discover()` function
- [ ] Modify `discovery.Result` to accept and merge built-in automations (or create a composite `Find` that checks both)
- [ ] Update `cli/run.go`, `cli/list.go`, `cli/info.go` to pass built-ins through to discovery
- [ ] Add `[built-in]` indicator to `pi list` output for built-in automations
- [ ] Write unit tests for built-in discovery, resolution, and local-takes-precedence
- [ ] Write integration test with a `pi:` reference

## Blocked By
<!-- None -->
