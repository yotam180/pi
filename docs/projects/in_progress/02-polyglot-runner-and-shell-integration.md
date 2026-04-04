# Polyglot Runner & Shell Integration

## Status
in_progress

## Priority
high

## Description
Extend the core engine with Python and TypeScript step types, pipe support between steps, and the `pi shell` command that installs project shortcuts into the developer's shell. This project delivers the two highest-value things PI offers over existing task runners: polyglot step chaining and first-class shell shortcut management.

## Goals
- Automations can mix bash, Python, and TypeScript steps naturally
- Scripts live in the same `.pi/` folder as the automation file — no separate `ops/` directory needed
- `pi shell` injects shortcut functions that work from anywhere in the terminal, always running from the repo root
- Steps can pipe stdout to the next step
- No runtime dependencies required for the developer beyond what they already have (Python, Node.js)

## Background & Context
The immediate use case is replacing the `vplf` shortcut, which pipes `docker-compose logs` output through a Python log formatter (`logf.py`). Today both the bash and Python parts live in `shell_shortcuts.sh` and `ops/local/logf.py`. With this project, both live under `.pi/docker/` and are described in a single automation YAML.

Shell shortcut injection solves the onboarding problem: instead of "copy this source line into your .zshrc", a developer runs `pi shell` once and all shortcuts from all PI-enabled repos they've installed are available.

## Scope

### In scope
- Step type: `python` — inline script string or `.py` file path (relative to automation file), run via the system Python or an active venv
- Step type: `typescript` — inline script string or `.ts` file path, run via `tsx` (must be installed)
- `pipe_to: next` on any step — stdout of that step becomes stdin of the next
- `pi shell` command:
  - Reads all shortcuts from `pi.yaml`
  - Writes shell functions to a PI-managed file (e.g. `~/.pi/shell/<repo-slug>.sh`)
  - Adds a single `source` line to `.zshrc` / `.bashrc` (only once)
  - Each shortcut function `cd`s to the repo root before running, unless `anywhere: true`
- `pi shell --uninstall` removes the shortcuts for the current repo
- `pi shell --list` shows all currently installed shortcuts across all repos

### Out of scope
- Built-in automation library (Project 3)
- Marketplace automations
- Named outputs / full inter-step variable passing (future)
- Windows support

## Success Criteria
- [ ] `vplf` equivalent works: a `docker/logs-formatted` automation with a bash step piped to a Python step, log formatter script living in `.pi/docker/`
- [ ] A TypeScript step runs a `.ts` file via `tsx` with no extra config
- [ ] `pi shell` writes working shortcut functions to the shell; `vpup` works from `~` after running `pi shell` in the repo
- [ ] Shortcuts `cd` to repo root by default; `anywhere: true` shortcuts do not
- [ ] `pi shell --uninstall` cleanly removes shortcuts for the repo without touching others
- [ ] Port `vplf`, `vpbrl`, `vpbrlf`, `vprl`, `vprlf` from Vyper's `shell_shortcuts.sh` as integration tests
- [ ] `go test ./...` passes

## Dependencies & Design Constraints

**Shortcut `with:` mapping (task `07-automation-inputs-schema`):** The `pi shell` codegen must be input-aware. Shortcuts can declare explicit `with:` mappings that bind CLI positional args to named automation inputs:

```yaml
shortcuts:
  dlogs: docker/logs           # simple: forwards "$@", auto-maps to inputs in order
  dlogs:
    run: docker/logs
    with:
      service: $1              # explicit: generates --with service="$1" in shell function
      tail: $2
```

The generated shell function for an explicit `with:` shortcut passes `--with key=value` flags to `pi run` rather than raw `"$@"`. This means `pi shell` implementation must check whether a shortcut has a `with:` block and emit different function bodies accordingly. Do not implement the simple `"$@"` passthrough in a way that makes the explicit `with:` case hard to add later — design for both forms from the start.

## Notes
- Python step should detect and use an active virtualenv (`VIRTUAL_ENV` env var) if present, otherwise fall back to `python3` in PATH.
- TypeScript step assumes `tsx` is available; PI should emit a clear error if it's not, with an install hint.
- The `~/.pi/shell/<repo-slug>.sh` approach means multiple PI repos on one machine can all have their shortcuts active simultaneously without conflicts.
- `anywhere: true` is the mechanism for shortcuts like `sk_deploy` that the developer might want to run without being inside the repo.
