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
    automation.go                  Automation, Step, StepType + Load(), FilePath, Dir()
    automation_test.go             14 tests
  discovery/                       .pi/ folder scanning and automation lookup
    discovery.go                   Discover(), Result, Find()
    discovery_test.go              18 tests
  executor/                        Step execution engine
    executor.go                    Executor, ExitError, Run(), execBash(), execRun()
    executor_test.go               20 tests
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
  └─ Executor (internal/executor)
       Runs steps in order: bash (inline/file), run: (recursive)
       Detects circular dependencies, propagates exit codes
```

## Key Design Decisions

### Automation naming
- `.pi/docker/up.yaml` → name `docker/up`
- `.pi/setup/cursor/automation.yaml` → name `setup/cursor`
- Names are always lowercase, no leading/trailing slashes
- Two files resolving to the same name is a hard error

### Package boundaries
- `config` knows only about `pi.yaml` structure
- `automation` knows only about a single automation file's structure; also stores `FilePath` for resolving relative script paths
- `discovery` ties them together: walks the filesystem, calls `automation.Load()` for each file, builds the name→automation map
- `executor` runs automation steps; depends on `automation` (types) and `discovery` (for `run:` step resolution)
- `cli` will tie discovery + config + executor together (task 05)

### Execution model
- All steps run with the repo root (directory containing `pi.yaml`) as their working directory
- Bash inline steps: `bash -c "<script>" -- [args...]` — args available as `$1`, `$2`, etc.
- Bash file steps: `bash <resolved_path> [args...]` — file path resolved relative to the automation YAML file's directory
- `run:` steps: recursive execution via `Executor.Run()` — args forwarded, circular dependencies detected via call stack
- If any step exits non-zero, execution stops immediately and the exit code propagates

### Error philosophy
- Parse errors include file path and field name
- `Find()` not-found errors list all available automations
- Collision errors mention both conflicting file paths
- Circular dependency errors show the full chain (e.g., `a → b → c → a`)

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests planned via `examples/` workspaces (task 06).

Total tests: 67 (7 CLI + 8 config + 14 automation + 18 discovery + 20 executor)
