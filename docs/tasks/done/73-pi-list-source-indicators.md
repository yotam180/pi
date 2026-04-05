# `pi list` Source Indicators and `--all` Flag

## Type
improvement

## Status
done

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
- [x] `pi list` default output includes a source column for every automation
- [x] Workspace automations show `[workspace]`
- [x] Package automations show the alias (if set) or `org/repo@version`
- [x] `pi list --all` includes automations from all declared packages, grouped by package with a header
- [x] `pi list --all` requires packages to be fetched — if not cached, prints "run pi setup first" for that package
- [x] `--builtins` flag (or `-b`) includes `pi:*` automations in the output
- [x] Output is well-formatted even with long package names
- [x] Tests cover: workspace-only, mixed workspace+package, --all, --builtins

## Implementation Notes

### Approach
- Added SOURCE column between NAME and DESCRIPTION in `pi list` output (4-column tabwriter: NAME, SOURCE, DESCRIPTION, INPUTS)
- `automationSource()` helper resolves the source indicator for each automation:
  - `[workspace]` for local `.pi/` automations
  - `[built-in]` for built-in automations
  - Package alias (e.g. `mytools`) if the package has an alias, otherwise the full source string (e.g. `org/repo@version`)
- `--builtins` / `-b` flag controls whether built-in automations appear in the list (default: hidden, cleaner output)
- `--all` / `-a` flag triggers `printPackageAutomations()` which appends grouped sections with `──` headers per package
- This is a breaking change: builtins no longer show by default. The `[built-in]` marker moved from being prepended to the DESCRIPTION to being the value in the SOURCE column.

### Files changed
- `internal/cli/list.go` — main implementation: SOURCE column, --all, --builtins flags, `automationSource()`, `printPackageAutomations()`
- `internal/cli/list_test.go` — 11 unit tests covering workspace source, package source, --all flag, --builtins flag, and existing functionality updated for new API
- `tests/integration/builtins_test.go` — updated all tests that checked `[built-in]` in list output to use `--builtins` flag; added `TestBuiltins_List_HiddenByDefault` test
- `tests/integration/packages_test.go` — added SOURCE column assertions to `TestPackages_List`; added `TestPackages_ListAll` test

### Test counts
- 11 unit tests in `list_test.go` (was 7)
- 26 integration tests in `builtins_test.go` (was 25, added HiddenByDefault)
- 11 integration tests in `packages_test.go` (was 9, added ListAll)

## Subtasks
- [x] Update `pi list` output to include source column
- [x] Implement `--all` flag logic
- [x] Implement `--builtins` flag
- [x] Add tests

## Blocked By
70-packages-declaration-in-pi-yaml
