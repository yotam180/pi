# PI — Architecture

## Overview

PI is a single Go binary (`pi`) that reads a project's `.pi/` folder and `pi.yaml` config to discover and execute automations. The architecture is layered: parsing → discovery → execution → CLI.

## Package Structure

```
cmd/pi/main.go                     Entry point, calls cli.Execute()
internal/
  cli/                             Cobra CLI commands
    root.go                        Root command, wires subcommands
    run.go                         pi run (stub — wiring in task 05)
    list.go                        pi list (stub — wiring in task 05)
    setup.go                       pi setup (stub)
    root_test.go                   CLI tests (7 tests)
  config/                          pi.yaml parsing
    config.go                      ProjectConfig, Shortcut, SetupEntry + Load()
    config_test.go                 8 tests
  automation/                      Individual automation YAML parsing
    automation.go                  Automation, Step, StepType + Load()
    automation_test.go             14 tests
  discovery/                       .pi/ folder scanning and automation lookup
    discovery.go                   Discover(), Result, Find()
    discovery_test.go              18 tests
```

## Data Flow

```
pi run docker/up
  │
  ├─ CLI (internal/cli)
  │    Parses args, finds project root
  │
  ├─ Config (internal/config)
  │    Loads pi.yaml → ProjectConfig
  │
  ├─ Discovery (internal/discovery)
  │    Walks .pi/ → map[name]*Automation
  │    Find("docker/up") → *Automation
  │
  └─ Executor (internal/executor — task 04)
       Runs steps in order: bash, run:, etc.
```

## Key Design Decisions

### Automation naming
- `.pi/docker/up.yaml` → name `docker/up`
- `.pi/setup/cursor/automation.yaml` → name `setup/cursor`
- Names are always lowercase, no leading/trailing slashes
- Two files resolving to the same name is a hard error

### Package boundaries
- `config` knows only about `pi.yaml` structure
- `automation` knows only about a single automation file's structure
- `discovery` ties them together: walks the filesystem, calls `automation.Load()` for each file, builds the name→automation map
- `cli` will tie discovery + config + executor together (task 05)

### Error philosophy
- Parse errors include file path and field name
- `Find()` not-found errors list all available automations
- Collision errors mention both conflicting file paths

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests planned via `examples/` workspaces (task 06).

Total tests: 47 (7 CLI + 8 config + 14 automation + 18 discovery)
