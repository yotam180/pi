# PI — Architecture

## Overview

PI is a single Go binary (`pi`) that reads a project's `.pi/` folder and `pi.yaml` config to discover and execute automations. The architecture is layered: parsing → discovery → execution → CLI.

## Package Structure

```
cmd/pi/main.go                     Entry point, calls cli.Execute()
internal/
  cli/                             Cobra CLI commands
    root.go                        Root command, wires subcommands, exit code handling
    run.go                         pi run — resolves and executes automations; --repo flag
    list.go                        pi list — discovers and prints automations
    setup.go                       pi setup — runs setup entries, then pi shell (CI-aware)
    shell.go                       pi shell — installs/uninstalls/lists shell shortcuts
    root_test.go                   CLI tests (7 tests)
    run_test.go                    pi run tests (8 tests)
    list_test.go                   pi list tests (6 tests)
    setup_test.go                  pi setup tests (4 tests)
    shell_test.go                  pi shell tests (3 tests)
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
    executor.go                    Executor, ExitError, Run(), execBash(), execPython(), execTypeScript(), execRun(); pipe_to:next support
    executor_test.go               47 tests
  project/                         Project root detection
    root.go                        FindRoot() — walks up to find pi.yaml
    root_test.go                   4 tests
  shell/                           Shell shortcut file generation and management
    shell.go                       GenerateShellFile(), Install(), Uninstall(), ListInstalled()
    shell_test.go                  11 tests
```

## Data Flow

```
pi run docker/up
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Discovery (internal/discovery)
  │    Walks .pi/ → map[name]*Automation
  │    Find("docker/up") → *Automation
  │
  └─ Executor (internal/executor)
       Runs steps in order: bash (inline/file), run: (recursive)
       Detects circular dependencies, propagates exit codes
```

```
pi list
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  └─ Discovery (internal/discovery)
       Walks .pi/ → map[name]*Automation
       Names() → sorted list → formatted table output
```

```
pi shell
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Config (internal/config)
  │    Loads pi.yaml → ProjectConfig (project name, shortcuts)
  │
  └─ Shell (internal/shell)
       GenerateShellFile() → builds function definitions
       Install() → writes to ~/.pi/shell/<project>.sh
       ensureSourceLine() → injects source block into .zshrc/.bashrc
```

```
pi setup
  │
  ├─ CLI (internal/cli)
  │    Parses args (--no-shell), gets CWD
  │
  ├─ Project + Config + Discovery
  │    Loads pi.yaml, discovers automations
  │
  ├─ Executor (internal/executor)
  │    Runs each setup entry sequentially
  │
  └─ Shell (internal/shell)  [unless CI or --no-shell]
       Install() → writes shortcuts, injects source line
```

## Key Design Decisions

### Project root detection
- `FindRoot()` walks up from the current directory, checking each level for `pi.yaml`
- Stops at the filesystem root with a clear error if not found
- Picks the closest `pi.yaml` (like `git` finds `.git/`)
- Used by both `pi run` and `pi list` so they work from any subdirectory

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
- `project` handles finding the repo root (directory containing `pi.yaml`)
- `cli` ties project + discovery + executor together

### Execution model
- All steps run with the repo root (directory containing `pi.yaml`) as their working directory
- Bash inline steps: `bash -c "<script>" -- [args...]` — args available as `$1`, `$2`, etc.
- Bash file steps: `bash <resolved_path> [args...]` — file path resolved relative to the automation YAML file's directory
- Python inline steps: `python3 -c "<script>" [args...]` — args available as `sys.argv[1:]`
- Python file steps: `python3 <resolved_path> [args...]` — file path resolved relative to the automation YAML file's directory
- Python uses `$VIRTUAL_ENV/bin/python` when a virtualenv is active, otherwise `python3`
- TypeScript inline steps: written to a temp file (`pi-ts-*.ts`), run via `tsx <tmpfile> [args...]` — args available as `process.argv.slice(2)`
- TypeScript file steps: `tsx <resolved_path> [args...]` — file path resolved relative to the automation YAML file's directory
- TypeScript requires `tsx` in PATH; clear error with install hint (`npm install -g tsx`) if not found
- `run:` steps: recursive execution via `Executor.Run()` — args forwarded, circular dependencies detected via call stack
- If any step exits non-zero, execution stops immediately and the exit code propagates

### Pipe support (`pipe_to: next`)
- When a step declares `pipe_to: next`, its stdout is captured to a `bytes.Buffer` instead of printed to terminal
- The captured buffer is fed as stdin to the next step
- If `pipe_to: next` appears on the last step, it's a no-op — output goes to terminal normally
- Stderr is never captured — it always goes to the terminal regardless of piping
- Works across all step types (bash, python, typescript, run)
- Exit code propagation: if a piping step fails, execution stops immediately
- Executor fields use `io.Writer`/`io.Reader` interfaces (not `*os.File`) to support buffer-based piping

### Shell shortcuts (`pi shell`)
- Shortcuts are defined in `pi.yaml → shortcuts:` as either a string (`"docker/up"`) or an object (`{run: ..., anywhere: true}`)
- `pi shell` writes shell functions to `~/.pi/shell/<project>.sh` — one file per project
- A source block (`for f in ~/.pi/shell/*.sh; do source "$f"; done`) is injected into `.zshrc` (and `.bashrc` if it exists)
- Source line injection is idempotent — checked before appending
- Default shortcuts `cd` to the repo root and call `pi run <automation> "$@"`
- `anywhere: true` shortcuts use `pi run --repo <root> <automation> "$@"` without cd
- `pi shell uninstall` removes the project file and cleans the source line if no repos remain
- `pi shell list` shows all installed shortcut files
- `pi setup` runs `pi shell` as its final step, skipping in CI environments (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.)
- `pi setup --no-shell` explicitly skips the shell step

### Error philosophy
- Parse errors include file path and field name
- `Find()` not-found errors list all available automations
- Collision errors mention both conflicting file paths
- Circular dependency errors show the full chain (e.g., `a → b → c → a`)
- Missing `pi.yaml` errors mention the start directory

### CLI output
- `pi list` uses `text/tabwriter` for aligned columns (NAME, DESCRIPTION)
- Automations without descriptions show `-` as placeholder
- Empty project (no automations) shows a friendly message, not an error
- `pi run` with unknown automation lists available automations in the error

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests in `tests/integration/` build the `pi` binary and run it against `examples/` workspaces using `exec.Command`.

Total tests: 152 (14 automation + 28 CLI + 8 config + 18 discovery + 47 executor + 4 project + 11 shell + 23 integration)

### Integration tests
- Build `pi` binary once in `TestMain`
- Run `pi list` and `pi run` against `examples/basic/`, `examples/docker-project/`, and `examples/pipe/`
- Assert exit codes, output content, and step ordering
- Pipe tests verify cross-language piping (bash→python→bash) end-to-end
- Shell tests: install, idempotent re-install, uninstall, list, `--repo` flag, setup integration, `--no-shell`
