# `pi list` Source Indicators and `--all` Flag

## Type
improvement

## Status
todo

## Priority
medium

## Project
13-external-packages

## Description
Update `pi list` to show where each automation comes from, and add `--all` to browse automations from declared packages alongside local ones.

**Default `pi list` (workspace automations only):**

```
  build/default          [workspace]   Build fzf binary for current platform
  build/install          [workspace]   Build and copy to bin/
  test/unit              [workspace]   Run Go unit tests
  docker/up              mytools       Start all containers
  docker/down            mytools       Stop all containers
```

The source column shows `[workspace]` for local `.pi/` automations and the alias (or `org/repo@version` if no alias) for package automations. This makes the origin of each automation immediately visible.

**`pi list --all`:**
Shows automations from local `.pi/` AND all automations available in every declared package. This is how you browse what a package offers before writing `run:` steps.

```
  build/default          [workspace]   Build fzf binary for current platform
  ...
  ── yotam180/pi-common@v1.2 ──────────────────────────────────────────
  docker/up              pi-common     Start Docker Compose services
  docker/down            pi-common     Stop Docker Compose services
  node/install-deps      pi-common     Run npm ci
  python/setup-venv      pi-common     Create Python virtualenv
  ...
```

Built-in automations (`pi:*`) are not shown unless `--builtins` flag is also passed (keeping the default output focused).

## Acceptance Criteria
- [ ] `pi list` default output includes a source column for every automation
- [ ] Workspace automations show `[workspace]`
- [ ] Package automations show the alias (if set) or `org/repo@version`
- [ ] `pi list --all` includes automations from all declared packages, grouped by package with a header
- [ ] `pi list --all` requires packages to be fetched — if not cached, prints "run pi setup first" for that package
- [ ] `--builtins` flag (or `-b`) includes `pi:*` automations in the output
- [ ] Output is well-formatted even with long package names
- [ ] Tests cover: workspace-only, mixed workspace+package, --all, --builtins

## Implementation Notes

## Subtasks
- [ ] Update `pi list` output to include source column
- [ ] Implement `--all` flag logic
- [ ] Implement `--builtins` flag
- [ ] Add tests

## Blocked By
70-packages-declaration-in-pi-yaml
