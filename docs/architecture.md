# PI — Architecture

## Overview

PI is a single Go binary (`pi`) that reads a project's `.pi/` folder and `pi.yaml` config to discover and execute automations. The architecture is layered: parsing → discovery → execution → CLI.

## Package Structure

```
cmd/pi/main.go                     Entry point, calls cli.Execute()
internal/
  builtins/                        Embedded built-in automations
    builtins.go                    //go:embed, Discover() — walks embedded FS, returns *discovery.Result
    builtins_test.go               38 tests (3 base + 7 docker + 12 installer + 16 devtool)
    embed_pi/                      Built-in automation YAML files (embedded at build time)
      hello.yaml                   Test built-in automation
      install-homebrew.yaml        pi:install-homebrew — macOS only, installs Homebrew
      install-python.yaml          pi:install-python — installs Python via mise/brew (version input)
      install-node.yaml            pi:install-node — installs Node.js via mise/brew (version input)
      install-uv.yaml              pi:install-uv — installs uv via official installer
      install-tsx.yaml             pi:install-tsx — installs tsx via npm
      cursor/
        install-extensions/
          automation.yaml          pi:cursor/install-extensions — installs missing Cursor extensions (extensions input)
      docker/
        up.yaml                    pi:docker/up — docker compose up -d with v1 fallback
        down.yaml                  pi:docker/down — docker compose down with v1 fallback
        logs.yaml                  pi:docker/logs — docker compose logs with v1 fallback
      git/
        install-hooks.yaml         pi:git/install-hooks — copies hook scripts to .git/hooks/ (source input)
  cli/                             Cobra CLI commands
    root.go                        Root command, wires subcommands, exit code handling
    discover.go                    discoverAll() — discovers local + built-in automations and merges
    run.go                         pi run — resolves and executes automations; --repo flag; --with key=value flag
    list.go                        pi list — discovers and prints automations with [built-in] markers
    info.go                        pi info — shows automation name, description, input docs, and if: conditions
    setup.go                       pi setup — runs setup entries (with if: support), then pi shell (CI-aware)
    shell.go                       pi shell — installs/uninstalls/lists shell shortcuts
    version.go                     pi version — prints version string
    root_test.go                   CLI tests (9 tests)
    run_test.go                    pi run tests (14 tests — includes --with and inputs tests)
    list_test.go                   pi list tests (7 tests — includes INPUTS column and built-in marker tests)
    info_test.go                   pi info tests (13 tests — includes if: condition display)
    setup_test.go                  pi setup tests (6 tests)
    shell_test.go                  pi shell tests (3 tests)
  conditions/                      Boolean expression parser/evaluator for if: fields
    conditions.go                  Lexer, AST, recursive-descent parser, Eval(), Predicates()
    conditions_test.go             31 tests
  config/                          pi.yaml parsing
    config.go                      ProjectConfig, Shortcut (with With field), SetupEntry (with If field) + Load()
    config_test.go                 11 tests
  automation/                      Individual automation YAML parsing
    automation.go                  Automation (with If field), Step (with If field), StepType, InputSpec + Load(), LoadFromBytes(), FilePath, Dir(), ResolveInputs(), InputEnvVars()
    automation_test.go             45 tests
  discovery/                       .pi/ folder scanning and automation lookup
    discovery.go                   Discover(), NewResult(), Result, Find() (with pi: prefix support), MergeBuiltins(), IsBuiltin()
    discovery_test.go              24 tests (18 base + 6 builtin merge/prefix tests)
  executor/                        Step execution engine
    executor.go                    Executor (with RuntimeEnv field), ExitError, Run(), RunWithInputs(), evaluateCondition(), execBash(), execPython(), execTypeScript(), execRun(); pipe_to:next support; PI_INPUT_* env injection; step-level and automation-level if: conditional execution
    predicates.go                  RuntimeEnv, DefaultRuntimeEnv(), ResolvePredicates(), ResolvePredicatesWithEnv(); resolves if: predicate names to booleans
    executor_test.go               75 tests (55 base + 13 step-conditional + 7 automation-conditional)
    predicates_test.go             11 tests (+ subtests covering all predicate types)
  project/                         Project root detection
    root.go                        FindRoot() — walks up to find pi.yaml
    root_test.go                   4 tests
  shell/                           Shell shortcut file generation and management
    shell.go                       GenerateShellFile(), Install(), Uninstall(), ListInstalled(); with: shortcut codegen
    shell_test.go                  14 tests
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
  ├─ Discovery (internal/discovery + internal/builtins)
  │    discoverAll(): walks .pi/ + embeds → merged map[name]*Automation
  │    Find("docker/up") → local first, then built-in fallback
  │    Find("pi:docker/up") → always built-in
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
  └─ Discovery (internal/discovery + internal/builtins)
       discoverAll(): walks .pi/ + embeds → merged map[name]*Automation
       Names() → sorted list → formatted table with [built-in] markers
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
pi info <name>
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Discovery (internal/discovery)
  │    Walks .pi/ → map[name]*Automation
  │    Find(name) → *Automation
  │
  └─ Output
       Prints name, description, if: condition (when present), step count,
       step details with per-step if: conditions (when any step has if:),
       and input specs
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
- `conditions` is a pure-logic package for parsing and evaluating boolean `if:` expressions — zero dependencies on other PI packages
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

### Conditional step execution (`if:` on steps)
- Steps can declare an `if:` field containing a boolean condition expression
- Before executing a step with `if:`, the executor extracts predicates, resolves them via the predicate resolver, and evaluates the expression
- If the expression evaluates to false, the step is silently skipped (no output, no error)
- Steps without `if:` always execute (backward compatible)
- Invalid `if:` expressions are caught at YAML load time, not at runtime
- `evaluateCondition()` on `Executor` uses `RuntimeEnv` field if set (for testing), otherwise `DefaultRuntimeEnv()`
- **Pipe passthrough on skip**: when a skipped step has `pipe_to: next`, any existing piped input from a prior step passes through to the next step unchanged. If the skipped step is the first in a pipe chain, `pipedInput` remains nil.

### Conditional automation execution (`if:` on automations)
- Automations can declare a top-level `if:` field containing a boolean condition expression
- Before executing any steps, `RunWithInputs()` evaluates the automation's `if:` condition
- If the condition evaluates to false, the automation is skipped: a message `[skipped] <name> (condition: <expr>)` is printed to stderr and `nil` is returned
- Automations without `if:` always execute (backward compatible)
- The condition check happens before `pushCall()`, so skipped automations don't consume call stack slots — this naturally prevents false circular-dependency errors
- `run:` steps calling a conditionally-skipped automation succeed without error (the parent step continues normally)
- Invalid `if:` expressions are caught at YAML load time, not at runtime

### Conditional setup entries (`if:` on setup)
- Setup entries in `pi.yaml` can declare an `if:` field containing a boolean condition expression
- Before running each setup entry, `pi setup` evaluates the `if:` condition using `conditions.Predicates()` + `executor.ResolvePredicates()` + `conditions.Eval()`
- If the condition evaluates to false, the entry is skipped with a message: `==> setup[N]: <name> [skipped] (condition: <expr>)`
- Entries without `if:` always run (backward compatible)
- Invalid `if:` expressions are caught at config load time, not at runtime
- The same predicate system used by step-level and automation-level `if:` is reused (os.*, command.*, env.*, file.exists(), etc.)

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

### Automation inputs (`inputs:` / `with:`)
- Automations can declare an `inputs:` block defining named parameters with type, required, default, and description
- `InputSpec` uses `*bool` for Required so we can distinguish "not set" from "set to false"
- `IsRequired()` defaults to true when no default value is provided and `required` is not explicitly set
- Input keys are stored in declaration order (`InputKeys []string`) for positional mapping
- `ResolveInputs()` on `Automation` validates and resolves inputs from either `--with` flags or positional args (mixing is an error)
- Resolved values are injected as `PI_INPUT_<NAME>` environment variables (uppercased, hyphens become underscores)
- `appendInputEnv()` merges input env vars with the current process environment (returns nil when no inputs, inheriting parent env)
- `run:` steps support `with:` to pass named inputs to the called automation
- `pi run --with key=value` is a repeatable flag parsed by `parseWithFlags()`
- `pi.yaml` shortcuts support `with:` mapping with `$1`, `$2` positional references
- Shell codegen detects `with:` on shortcuts and emits `--with key="$N"` flags instead of `"$@"` passthrough
- `pi list` shows an INPUTS column with required inputs as `name` and optional as `name?`

### Built-in automations (`pi:` prefix)
- Built-in automation YAML files live in `internal/builtins/embed_pi/` and are embedded into the binary at build time via `//go:embed`
- `internal/builtins.Discover()` walks the embedded FS, parses YAML files with `automation.LoadFromBytes()`, and returns a `*discovery.Result`
- `cli.discoverAll()` discovers local automations, then calls `result.MergeBuiltins()` to incorporate built-ins
- Precedence: local automations shadow built-ins with the same name. `Find("hello")` returns local if it exists, otherwise built-in. `Find("pi:hello")` always returns the built-in regardless of local shadowing
- `pi list` marks built-in automations (not shadowed) with `[built-in]` in the DESCRIPTION column
- `run:` steps can reference `pi:hello` to explicitly call built-in automations
- `pi.yaml` setup entries can reference `pi:hello` for built-in setup automations
- Built-in automations use inline scripts only (no file-path steps) since they have no real filesystem directory
- Docker automations (`docker/up`, `docker/down`, `docker/logs`) detect `docker compose` (v2 plugin) first, falling back to `docker-compose` (v1 standalone); forward all CLI args via `"$@"`
- Installer automations (`install-homebrew`, `install-python`, `install-node`, `install-uv`, `install-tsx`) are idempotent check-then-install scripts:
  - Each checks `command -v` first and prints `[already installed]` with version info; installs and prints `[installed]` otherwise
  - `install-homebrew` has `if: os.macos` at the automation level (skipped on non-macOS)
  - `install-python` and `install-node` accept a `version` input; try `mise` first, fall back to `brew`
  - `install-uv` uses the official `astral.sh/uv/install.sh` script
  - `install-tsx` uses `npm install -g tsx`
- Dev tool automations (`cursor/install-extensions`, `git/install-hooks`) handle common team setup tasks:
  - `cursor/install-extensions` accepts an `extensions` input (comma or newline-separated IDs), checks `cursor --list-extensions`, installs missing ones via `cursor --install-extension`
  - `git/install-hooks` accepts a `source` input (directory path relative to repo root), copies hook files to `.git/hooks/`, makes them executable; uses `cmp` for idempotency

### Error philosophy
- Parse errors include file path and field name
- `Find()` not-found errors list all available automations
- Collision errors mention both conflicting file paths
- Circular dependency errors show the full chain (e.g., `a → b → c → a`)
- Missing `pi.yaml` errors mention the start directory

### Condition expressions (`internal/conditions`)
- Pure-logic package with zero PI internal dependencies — receives `map[string]bool` and expression string, returns `(bool, error)`
- Lexer tokenizes into IDENT, AND, OR, NOT, LPAREN, RPAREN, STRING, EOF with byte-position tracking
- Recursive-descent parser follows: `expr → orExpr → andExpr → notExpr → primary`
- `and` binds tighter than `or` (standard boolean precedence)
- Supports: bare dotted identifiers (`os.macos`), function-call syntax (`file.exists(".env")`), `and`/`or`/`not`, parentheses
- Function-call predicates are keyed in the predicate map as `name("arg")` (e.g., `file.exists(".env")`)
- `Eval(expr, predicates)` — evaluates expression; empty expression returns `true`
- `Predicates(expr)` — extracts all predicate names for pre-resolution; deduplicates preserving first-occurrence order
- Error messages include position information for malformed expressions and the predicate name for unknown predicates

### Predicate resolution (`internal/executor/predicates.go`)
- Converts predicate names (from `conditions.Predicates()`) into `map[string]bool` for `conditions.Eval()`
- `ResolvePredicates(names, repoRoot)` is the public API; `ResolvePredicatesWithEnv(names, repoRoot, env)` accepts an injected `RuntimeEnv` for testing
- `RuntimeEnv` struct captures injectable `GOOS`, `GOARCH`, `Getenv()`, `LookPath()`, `Stat()` — no direct global reads
- Static predicates: `os.macos`, `os.linux`, `os.windows`, `os.arch.arm64`, `os.arch.amd64`, `shell.zsh`, `shell.bash`
- Dynamic predicates: `env.<NAME>` checks `Getenv(NAME) != ""`; `command.<name>` checks `LookPath(name) == nil`
- Function-call predicates: `file.exists("path")` checks path exists and is a file; `dir.exists("path")` checks path exists and is a directory — both resolve relative to `repoRoot`
- Unknown predicates produce an error listing all valid prefixes

### CLI output
- `pi list` uses `text/tabwriter` for aligned columns (NAME, DESCRIPTION)
- Automations without descriptions show `-` as placeholder
- Empty project (no automations) shows a friendly message, not an error
- `pi run` with unknown automation lists available automations in the error

## CI/CD

### GitHub Actions CI (`.github/workflows/ci.yml`)
- Triggers on push to `main` and pull requests targeting `main`
- OS matrix: `ubuntu-latest`, `macos-latest`
- Go version from `go.mod` via `actions/setup-go@v5` (includes module caching)
- Node.js 22 + `tsx` installed for TypeScript step runner tests
- Python 3 pre-installed on both runners
- Steps: `go vet ./...` → `go build ./...` → `go test ./... -race -count=1`

### Release workflow (`.github/workflows/release.yml`)
- Triggers on tag push matching `v*`
- Runs full test suite before releasing as a safety gate
- Uses `goreleaser/goreleaser-action@v6` with GoReleaser v2
- Produces cross-compiled binaries for: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- Each binary archived as `.tar.gz` with `README.md`
- `checksums.txt` included in every release
- Version injected via ldflags (`cli.version`) — same variable as the `.pi/build.yaml` automation
- Uses default `GITHUB_TOKEN` for GitHub Release creation
- Passes `HOMEBREW_TAP_TOKEN` secret to GoReleaser for Homebrew tap updates

### GoReleaser (`.goreleaser.yaml`)
- Config version 2
- CGO disabled, binaries stripped (`-s -w`)
- Changelog auto-generated, excludes docs/test/chore commits
- `pi run snapshot` runs a local snapshot build for testing
- `homebrew_casks:` section auto-publishes a Homebrew Cask to `yotam180/homebrew-pi` on each non-prerelease tag
- Cask includes xattr quarantine removal postflight hook for unsigned macOS binaries

### Homebrew distribution
- Tap repo: `github.com/yotam180/homebrew-pi`
- Install: `brew install yotam180/pi/pi`
- GoReleaser generates `Casks/pi.rb` and pushes it to the tap on each release
- Uses `HOMEBREW_TAP_TOKEN` (fine-grained PAT with write access to `homebrew-pi`)
- `skip_upload: auto` prevents pre-release tags from updating the tap

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `gopkg.in/yaml.v3` — YAML parsing
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests in `tests/integration/` build the `pi` binary and run it against `examples/` workspaces using `exec.Command`.

Total tests: 390 (41 automation + 38 builtins + 48 CLI + 30 conditions + 11 config + 24 discovery + 86 executor + 4 project + 14 shell + 94 integration)

### Integration tests
- Build `pi` binary once in `TestMain`
- Run `pi list` and `pi run` against `examples/basic/`, `examples/docker-project/`, and `examples/pipe/`
- Assert exit codes, output content, and step ordering
- Pipe tests verify cross-language piping (bash→python→bash) end-to-end
- Polyglot tests cover Python (inline/file), TypeScript (inline/file), multi-step pipe chains (bash→Python→TypeScript), and `run:` step piping
- Shell tests: install, idempotent re-install, uninstall, list, `--repo` flag, setup integration, `--no-shell`, setup with conditional entries (skip/run), conditional skip shows condition
- Inputs tests: positional mapping, `--with` flags, defaults, missing required errors, unknown input errors, `run:` step with `with:`, `pi list` INPUTS column
- Info tests: basic automation details, automation with inputs (required/optional/defaults), not-found error, missing argument error
- Conditional tests: list, platform-info (OS-aware step skipping), skip-all (all conditional steps skipped), pipe-conditional (pipe passthrough on skipped step), automation-level-if list, impossible (always-skipped automation), macos-only (OS-aware automation), run-step calling skipped automation, env predicate (with/without var), command predicate (available/missing), file.exists/dir.exists predicates, complex boolean expressions (and/or/not/parentheses), combined automation+step level if, pi info showing conditions (automation-level, step-level, absent)
- Docker built-in tests: all three docker automations appear in `pi list` with `[built-in]` marker, `pi info` shows details for each, `run:` step resolution works via `docker-builtins` example workspace
- Installer built-in tests: all 5 installer automations (`install-homebrew`, `install-python`, `install-node`, `install-uv`, `install-tsx`) appear in `pi list` with `[built-in]` marker, `pi info` shows details/inputs/conditions for each, `pi run pi:install-tsx` executes idempotently, `pi list` shows INPUTS column for versioned installers
- Dev tool built-in tests: `cursor/install-extensions` and `git/install-hooks` appear in `pi list` with `[built-in]` marker, `pi info` shows details and inputs for each, `pi list` shows INPUTS column with required input names
