# PI — Architecture

## Overview

PI is a single Go binary (`pi`) that reads a project's `.pi/` folder and `pi.yaml` config to discover and execute automations. The architecture is layered: parsing → discovery → execution → CLI.

## Package Structure

```
cmd/pi/main.go                     Entry point, calls cli.Execute()
internal/
  builtins/                        Embedded built-in automations
    builtins.go                    //go:embed, Discover() — walks embedded FS, returns *discovery.Result
    builtins_test.go               42 tests (3 base + 7 docker + 16 installer + 16 devtool)
    embed_pi/                      Built-in automation YAML files (embedded at build time)
      hello.yaml                   Test built-in automation
      install-homebrew.yaml        pi:install-homebrew — macOS only, installs Homebrew
      install-python.yaml          pi:install-python — installs Python via mise/brew (version input)
      install-node.yaml            pi:install-node — installs Node.js via mise/brew (version input)
      install-go.yaml              pi:install-go — installs Go via mise/brew (version input)
      install-rust.yaml            pi:install-rust — installs Rust via rustup (version input)
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
    discover.go                    discoverAll() — discovers local + built-in automations and merges; passes os.Stderr for name-mismatch warnings
    run.go                         pi run — resolves and executes automations; --repo flag; --with key=value flag; --silent flag; --loud flag; wires Provisioner from config
    list.go                        pi list — discovers and prints automations with [built-in] markers
    info.go                        pi info — shows automation name, description, input docs, if: conditions, dir: overrides, timeout: annotations, step descriptions, and install lifecycle for installer automations
    setup.go                       pi setup — runs setup entries (with if: support), then pi shell (CI-aware); --silent flag; --loud flag; color-coded headers via display.Printer; auto-source rc file via PI_PARENT_EVAL_FILE
    shell.go                       pi shell — installs/uninstalls/lists shell shortcuts
    version.go                     pi version — prints version string
    doctor.go                      pi doctor — scans all automations, checks requires: entries, prints health table; color-coded ✓/✗ via display.Printer
    validate.go                    pi validate — statically validates pi.yaml and .pi/ automations; cross-checks shortcut, setup, and run: step references; reports all errors; exit 0/1
    root_test.go                   CLI tests (12 tests — includes doctor and validate subcommands)
    run_test.go                    pi run tests (14 tests — includes --with, inputs, --silent tests)
    validate_test.go               pi validate tests (11 tests — valid project, broken refs, multiple errors, builtin refs, no pi.yaml)
    list_test.go                   pi list tests (7 tests — includes INPUTS column and built-in marker tests)
    info_test.go                   pi info tests (17 tests — includes if: condition display, installer type, dir: annotation, timeout: annotation, and step description display)
    setup_test.go                  pi setup tests (8 tests — includes --silent, parent eval file)
    shell_test.go                  pi shell tests (3 tests)
    doctor_test.go                 pi doctor tests (9 tests — no-automations, no-requirements, satisfied, missing, mixed, skips)
  conditions/                      Boolean expression parser/evaluator for if: fields
    conditions.go                  Lexer, AST, recursive-descent parser, Eval(), Predicates()
    conditions_test.go             31 tests
  config/                          pi.yaml parsing
    config.go                      ProjectConfig, Shortcut (with With field), SetupEntry (with If field), RuntimesConfig + Load()
    config_test.go                 17 tests
  automation/                      Individual automation YAML parsing
    automation.go                  Automation struct (with If, Install, Requires, Inputs fields) + Load(), LoadFromBytes(), Dir(), IsInstaller(), validate(), buildShorthandStep(); single-step shorthand support (top-level bash/python/typescript/run keys)
    step.go                        StepType, Step (with If, Env, Silent, ParentShell, Dir, Timeout, Description), stepRaw, toStep(), InstallPhase, InstallSpec, validateSteps(), validateInstall(), validateInstallPhase()
    inputs.go                      InputSpec, inputsRaw, ResolveInputs(), InputEnvVars()
    requirements.go                RequirementKind, Requirement, requirementRaw, parseNameVersion(), validateVersionString()
    automation_test.go             26 tests (core load, validate, basic step parsing, single-step shorthand)
    step_test.go                   53 tests (if/env/silent/parent_shell/dir/timeout/description fields, install block)
    inputs_test.go                 16 tests (input spec, resolution, env vars, with: on steps)
    requirements_test.go           20 tests (requires parsing, version validation, name-version parsing)
  display/                         Styled terminal output (color, TTY detection)
    display.go                     Printer struct, color methods (Plain, Dim, Green, Red, Bold), InstallStatus, SetupHeader, StepTrace, truncateTrace, shouldColor
    tty.go                         isTerminal() via golang.org/x/term
    display_test.go                30 tests (styles, color toggle, NO_COLOR, TTY, install status variants, step trace, truncateTrace)
  discovery/                       .pi/ folder scanning and automation lookup
    discovery.go                   Discover() (with warnWriter for name mismatch warnings), NewResult(), Result, Find() (with pi: prefix support), MergeBuiltins(), IsBuiltin(), reconcileAutomationName()
    discovery_test.go              29 tests (18 base + 6 builtin merge/prefix + 5 optional name tests)
  executor/                        Step execution engine
    executor.go                    Executor struct (with ParentEvalFile and Runners fields), ExitError, Run(), RunWithInputs(), execStep(), execStepSuppressed(), execParentShell(), AppendToParentEval(), evaluateCondition(), pushCall()/popCall(), printer(), registry(), newRunContext(), stdout()/stderr()/stdin(); pipe_to:next orchestration; step-level and automation-level if: conditional execution; step-level silent: true suppression; --loud override; parent_shell: true eval-file delegation; step dispatch via Registry; dir: validation before step execution
    runner_iface.go                StepRunner interface, RunContext (step execution context with WorkDir), Registry (maps StepType→StepRunner), NewRegistry(), NewDefaultRegistry()
    runners.go                     Step runner implementations: BashRunner, PythonRunner, TypeScriptRunner, RunStepRunner; each implements StepRunner interface; runStepCommand() shared command execution with timeout support (exec.CommandContext); TimeoutExitCode (124); resolvePythonBin(), isCommandNotFound()
    install.go                     Installer lifecycle: execInstall(), execInstallPhase(), execInstallPhaseCapture(), execBashSuppressed(), captureVersion(), printInstallStatus(), printIndentedStderr(); structured test→run→verify→version lifecycle; color-coded installer status via display.Printer; install phase step dispatch uses Registry
    helpers.go                     Shared utilities: resolveFileStep() (file-path resolution + existence check), isFilePath(), resolveScriptPath(), buildEnv(), prependPathInEnv(), resolveStepDir(); PI_INPUT_* env injection; provisioned runtime PATH injection; step-level env: injection; step-level dir: resolution
    validate.go                    ValidateRequirements() (with provisioning fallback), tryProvision(), checkRequirementImpl() (shared logic with alwaysDetectVersion flag), checkRequirement(), CheckRequirementForDoctor(), detectVersion(), extractVersion(), compareVersions(), FormatValidationError(), InstallHintFor(), CheckResult, ValidationError, installHints; pre-execution requirement validation
    predicates.go                  RuntimeEnv (with ExecOutput field), DefaultRuntimeEnv(), ResolvePredicates(), ResolvePredicatesWithEnv(); resolves if: predicate names to booleans
    test_helpers_test.go           Shared test helpers: newAutomation, newAutomationInDir, newExecutor, newExecutorWithCapture, newExecutorWithEnv, step constructors (bashStep, runStep, pythonStep, typescriptStep, pipedBashStep, pipedPythonStep, bashStepIf), fakeRuntimeEnv, requirePython, requireTsx, boolPtr
    executor_test.go               20 tests — core execution: bash inline/file, run step chaining, circular deps, multi-step, working dir, mixed bash+run, exit error, isFilePath, call stack isolation
    python_runner_test.go          9 tests — python inline/file, venv detection, mixed bash+python
    typescript_runner_test.go      8 tests — typescript inline/file, tsx not found, mixed bash+typescript
    pipe_test.go                   10 tests — pipe_to:next: bash→bash, bash→python, python→bash, three-step chain, failure propagation, stderr passthrough, run step piping, multiline data
    inputs_test.go                 7 tests — RunWithInputs: env var injection, positional, defaults, missing required, mixing error, args passthrough, run step with with
    conditional_step_test.go       13 tests — step-level if: true/false/not/complex, mixed conditional+unconditional, pipe passthrough on skip, file.exists/not
    conditional_automation_test.go 7 tests — automation-level if: true/false, run step calling skipped/executed automation, complex condition, skip vs circular dependency
    install_test.go                11 tests — installer lifecycle: already installed, fresh install, run fails, verify fails, verify defaults to test, no version, silent, stderr on failure, step list with conditionals, with inputs, automation-level if
    step_env_test.go               9 tests — step-level env: bash/python, multiple vars, parent override, nil env inheritance, per-step isolation, buildEnv with step env, buildEnv with all three, buildEnv step env deterministic order
    step_dir_test.go               10 tests — step-level dir: bash inline/absolute/default, missing dir error, not-a-dir error, python step, per-step isolation, mixed with no dir, combined with env, resolveStepDir unit tests
    step_trace_test.go             6 tests — step trace lines, silent step suppression, loud override, silent still executes, silent pipe capture
    step_timeout_test.go           8 tests — step-level timeout: no timeout runs normally, not exceeded, exceeded (killed with exit 124), stops execution chain, with pipe_to, with silent, skipped by condition, multiple steps only timed-out killed
    parent_shell_test.go           6 tests — parent shell: writes to eval file, multiple steps append, mixed with normal, no eval file error, skipped by condition, AppendToParentEval
    validate_test.go               34 tests (version extraction, version comparison, requirement checking, validation integration, error formatting, install hints, CheckRequirementForDoctor, InstallHintFor, provisioning integration, prependPathInEnv)
    predicates_test.go             12 tests (+ subtests covering all predicate types)
  project/                         Project root detection
    root.go                        FindRoot() — walks up to find pi.yaml
    root_test.go                   4 tests
  runtimes/                        Sandboxed runtime provisioning
    runtimes.go                    Provisioner, Provision(), provisionWithMise(), provisionDirect(), PrependToPath()
    runtimes_test.go               16 tests
  shell/                           Shell shortcut file generation and management
    shell.go                       GenerateShellFile(), Install(), Uninstall(), ListInstalled(), PrimaryRCFile(); with: shortcut codegen; PI_PARENT_EVAL_FILE eval wrapper pattern; pi-setup-<project> helper function
    shell_test.go                  16 tests
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
       step details with per-step if: conditions and descriptions (when
       any step has if:, description:, or other annotations), and input specs
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

```
pi doctor
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Discovery (internal/discovery + internal/builtins)
  │    discoverAll(): walks .pi/ + embeds → merged map[name]*Automation
  │
  └─ Executor/Validate (internal/executor)
       For each automation with requires:
         CheckRequirementForDoctor() → CheckResult per requirement
       Prints per-automation health table with ✓/✗ icons
       Exit 0 (all satisfied) or 1 (any missing)
```

```
pi validate
  │
  ├─ CLI (internal/cli)
  │    Parses args, gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Config (internal/config)
  │    Loads pi.yaml → ProjectConfig (validates schema)
  │
  ├─ Discovery (internal/discovery + internal/builtins)
  │    discoverAll(): walks .pi/ + embeds → merged map[name]*Automation
  │
  └─ Cross-reference validation (internal/cli/validate.go)
       Checks shortcut targets → automation names
       Checks setup entry targets → automation names
       Checks run: step values → automation names (incl. install phases)
       Collects all errors, prints to stderr, exit 0 or 1
```

## Build

```
Makefile                               build, vet, test, test-matrix targets
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
- The `name:` field in automation YAML is optional — PI derives it from the file path
- When `name:` is absent, `Discover()` sets `a.Name` to the path-derived name
- When `name:` is present and matches, no warning is emitted
- When `name:` is present but mismatches the derived name, a warning is printed to stderr
- Built-in automations (`internal/builtins`) apply the same rule: name from path when absent

### Single-step shorthand
- Automations with a single step can place the step type key (`bash:`, `python:`, `typescript:`, `run:`) at the top level, skipping the `steps:` wrapper
- Step modifier fields (`env:`, `dir:`, `timeout:`, `silent:`, `pipe_to:`) are supported alongside the shorthand key at the top level
- Top-level `if:` maps to the automation-level condition (not a step-level condition) — for a single-step automation, these are semantically equivalent
- Top-level `description:` remains the automation description (not a step description)
- Having both a top-level step key and `steps:` (or `install:`) in the same file is a parse error
- Having multiple top-level step keys (e.g. both `bash:` and `python:`) is a parse error
- Implementation: `buildShorthandStep()` in `automation.go` constructs a `stepRaw` from top-level keys; `UnmarshalYAML` expands it to a single-element `Steps` slice before validation — the rest of the system (executor, info, validate) sees a normal automation
- Shorthand is additive — existing `steps:` syntax continues to work unchanged

### Package boundaries
- `config` knows only about `pi.yaml` structure
- `automation` knows only about a single automation file's structure; also stores `FilePath` for resolving relative script paths
- `conditions` is a pure-logic package for parsing and evaluating boolean `if:` expressions — zero dependencies on other PI packages
- `discovery` ties them together: walks the filesystem, calls `automation.Load()` for each file, builds the name→automation map
- `executor` runs automation steps; depends on `automation` (types) and `discovery` (for `run:` step resolution)
- `project` handles finding the repo root (directory containing `pi.yaml`)
- `cli` ties project + discovery + executor together

### Execution model
- By default, all steps run with the repo root (directory containing `pi.yaml`) as their working directory. Steps can override this with `dir:`.
- Step runner registry: `Registry` in `runner_iface.go` maps `StepType` → `StepRunner`. `NewDefaultRegistry()` registers `BashRunner`, `PythonRunner`, `TypeScriptRunner`, `RunStepRunner`. `Executor.execStep()` dispatches through the registry instead of a switch statement. Adding a new step type only requires implementing `StepRunner` and registering it — no executor changes needed.
- `RunContext` in `runner_iface.go` bundles everything a runner needs: automation, step, args, I/O writers, env, repo root, and `WorkDir` (the resolved working directory). Runners are decoupled from `Executor` internals — they receive a callback for recursive `run:` calls.
- Common execution substrate: `runStepCommand()` in `runners.go` handles `exec.Command` setup (Dir from `WorkDir`, Env, Stdout, Stderr, Stdin) and error wrapping (`*ExitError` for non-zero exits). All language runners delegate to this.
- File-path resolution: `resolveFileStep()` in `helpers.go` combines `isFilePath()` + `resolveScriptPath()` + existence check into a single call, eliminating per-runner duplication
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
- Adding a new step type: (1) register the `StepType` constant in `automation.go` and add it to `validStepTypes`, (2) implement `StepRunner` in `runners.go` (using `resolveFileStep()` for file-path handling and `runStepCommand()` for execution), (3) register it in `NewDefaultRegistry()` in `runner_iface.go` — no changes to `executor.go` needed

### Installer automation lifecycle (`install:` block)
- Automations can use `install:` instead of `steps:` — the two are mutually exclusive
- `InstallSpec` has four fields: `Test`, `Run`, `Verify` (optional), `Version` (optional)
- `InstallPhase` is polymorphic: either a scalar bash string or a list of steps (same step schema as `steps:`)
- When `verify:` is absent, the `test:` phase is re-run as verification after `run:` completes
- Executor runs `execInstall()`: test → [run → verify] → version
- All phase stdout is suppressed; stderr is captured from `run:` and shown only on failure
- `version:` command stdout is trimmed and displayed in the status line
- PI prints one formatted status line per installer: `✓ / → / ✗  name  status  (version)`
- `--silent` flag on `pi run` and `pi setup` suppresses PI status lines (stderr from failures always shown)
- Step lists in install phases support `if:` conditions, all step types (bash, python, typescript), and `run:` references. Step dispatch uses the same `Registry` as normal steps.
- A `run:` step inside an `install:` phase that references another installer automation runs that automation's own `install:` lifecycle

### Pipe support (`pipe_to: next`)
- When a step declares `pipe_to: next`, its stdout is captured to a `bytes.Buffer` instead of printed to terminal
- The captured buffer is fed as stdin to the next step
- If `pipe_to: next` appears on the last step, it's a no-op — output goes to terminal normally
- Stderr is never captured — it always goes to the terminal regardless of piping
- Works across all step types (bash, python, typescript, run)
- Exit code propagation: if a piping step fails, execution stops immediately
- Executor fields use `io.Writer`/`io.Reader` interfaces (not `*os.File`) to support buffer-based piping

### Step description (`description:` on steps)
- Steps can declare an optional `description:` field with a human-readable string
- Descriptions are purely informational — they have no effect on step execution
- `pi info` shows step descriptions as an indented line below the step detail line when the "Step details" section is displayed
- The presence of a `description:` on any step triggers the "Step details" section (same as `if:`, `env:`, etc.)
- Steps without `description:` have an empty string (backward compatible)
- Works with all step types: bash, python, typescript, run
- Compatible with all other step fields: `if:`, `env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`, `pipe_to`

### Step trace lines and silent/loud
- Before executing each non-installer step, PI prints a trace line to stderr: `  → <type>: <truncated-command>`
- Trace lines use dim styling via `display.Printer.StepTrace()`
- Multiline commands are collapsed to first line with `...`; long commands are truncated at 80 chars
- Installer steps are exempt — they have their own formatted status output
- A step with `silent: true` suppresses its trace line AND its stdout/stderr output
- Silent steps still execute — only their output is hidden; pipe data (`pipe_to: next`) still flows through
- `Executor.Loud` overrides all `silent: true` flags — when set, every step prints trace + output
- `--loud` flag on `pi run` and `pi setup` sets `Executor.Loud = true`
- `execStepSuppressed()` wraps `execStep()` with stdout/stderr redirected to `io.Discard` for silent steps
- When a silent step uses `pipe_to: next`, pipe capture still works (only non-pipe stdout is discarded)
- `pi info` shows `[silent]` annotation on steps with `silent: true`

### Parent shell execution (`parent_shell: true` on steps)
- Bash steps can declare `parent_shell: true` to run in the calling shell instead of as a subprocess
- `parent_shell` is only valid on bash steps — error on python, typescript, or run steps
- `parent_shell` cannot be combined with `pipe_to` — error at parse time
- When a parent_shell step executes, PI does **not** run it; instead it appends the command to `Executor.ParentEvalFile`
- `ParentEvalFile` is populated from the `PI_PARENT_EVAL_FILE` env var by `cli/run.go` and `cli/setup.go`
- If `ParentEvalFile` is empty and a parent_shell step is encountered, an error is returned with a message about `PI_PARENT_EVAL_FILE`
- After PI exits, the shell wrapper function (generated by `pi shell`) sources the eval file, running the commands in the parent shell
- `execParentShell()` prints a trace line `  → parent: <command>` before writing to the eval file
- `AppendToParentEval(path, command)` is the public helper that appends a line to the eval file
- The `if:` condition check on parent_shell steps happens before the parent_shell check — skipped steps don't write to the eval file
- `pi info` shows `[parent_shell]` annotation on steps with `parent_shell: true`
- Use cases: `source venv/bin/activate`, `cd /some/dir`, `export VAR=value`

### Auto-sourcing after `pi setup`
- After `pi setup` installs shell shortcuts, if `PI_PARENT_EVAL_FILE` is set, PI writes `source <rc-file>` to the eval file
- The rc file is determined by `shell.PrimaryRCFile()` — prefers `.zshrc`, falls back to `.bashrc`
- This makes shortcuts immediately available in the current terminal without manual sourcing
- `pi shell` generates a `pi-setup-<project>` helper function that wraps `pi setup` with the eval pattern
- First-run bootstrapping: on the very first `pi setup` (before any shell wrapper exists), auto-sourcing doesn't work — the user runs `source ~/.zshrc` once, then `pi-setup-<project>` handles it automatically going forward

### Shell wrapper eval pattern
- All shell shortcut functions generated by `pi shell` use the `PI_PARENT_EVAL_FILE` eval pattern
- Pattern: create a temp file → set `PI_PARENT_EVAL_FILE` → run `pi run` → source the file if non-empty → clean up → preserve exit code
- `pi-setup-<project>` uses the same pattern wrapping `pi setup`
- The temp file is always cleaned up, even if PI exits with an error
- The exit code from `pi run` is preserved via `local _pi_exit=$?` → `return $_pi_exit`

### Step-level environment variables (`env:` on steps)
- Steps can declare an `env:` mapping of key-value pairs to inject into the step's execution environment
- `env:` vars are merged into the process environment after PI_INPUT_* vars and provisioned runtime PATH
- Step-level env vars do not leak between steps — each step gets a fresh copy of the process environment plus its own `env:` overlay
- Steps without `env:` inherit the parent process environment as before (backward compatible)
- Step-level env vars override parent env vars with the same name (last-writer-wins since they're appended)
- Works with all step types: bash, python, typescript
- `pi info` shows `[env: KEY1, KEY2]` annotations on steps that declare env vars
- Install phases (`install:` block) do not support step-level `env:` — they use only input env vars

### Step working directory (`dir:` on steps)
- Steps can declare a `dir:` field to override the working directory for that step's execution
- When `dir:` is set, the path is resolved relative to the repo root; absolute paths are used as-is
- The resolved directory must exist at execution time — `resolveStepDir()` in `helpers.go` validates existence and confirms it's a directory
- Validation happens in `execStep()` before the runner is invoked; non-existent or non-directory paths produce a clear error
- Steps without `dir:` use the repo root as their working directory (backward compatible)
- `dir:` is per-step — each step independently resolves its own directory, no carry-over between steps
- `WorkDir` on `RunContext` carries the resolved directory to runners; `runStepCommand()` uses `cmd.Dir = ctx.WorkDir`
- Works with all step types: bash, python, typescript
- `dir:` is independent of other step fields: combinable with `env:`, `if:`, `silent:`, `pipe_to`
- `parent_shell` steps: `dir:` has no effect since parent_shell steps don't execute as subprocesses
- `pi info` shows `[dir: <path>]` annotations on steps that declare `dir:`

### Step timeout (`timeout:` on steps)
- Steps can declare a `timeout:` field with a Go-style duration string (e.g., `30s`, `5m`, `1h30m`)
- `timeout:` is parsed via `time.ParseDuration()` during YAML unmarshalling; stored as `time.Duration` on `Step` struct
- `TimeoutRaw` preserves the original string for display in `pi info`
- Non-positive durations (zero or negative) are rejected at parse time
- `timeout:` is invalid on `run:` steps (parse-time error) — set timeouts on the target automation's own steps instead
- `timeout:` is invalid on `parent_shell:` steps (parse-time error) — they don't execute as subprocesses
- Enforcement: `runStepCommand()` uses `exec.CommandContext` with `context.WithTimeout` when `Step.Timeout > 0`
- On timeout, the process is killed via context cancellation and `*ExitError{Code: 124}` is returned — exit code 124 matches the GNU `timeout(1)` convention
- Works with all subprocess step types: bash, python, typescript
- Compatible with all other step fields: `env:`, `dir:`, `silent:`, `if:`, `pipe_to`
- When a step with `if:` evaluates to false, no timeout applies (step is skipped)
- When a step with `silent: true` times out, the timeout error still propagates
- `pi info` shows `[timeout: <value>]` annotations on steps that declare `timeout:`

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
- All shortcut functions use the `PI_PARENT_EVAL_FILE` eval wrapper pattern: create temp file → set env var → run `pi run` → source eval file if non-empty → clean up → preserve exit code
- Default shortcuts `cd` to the repo root and call `pi run <automation> "$@"`
- `anywhere: true` shortcuts use `pi run --repo <root> <automation> "$@"` without cd
- Each project's shell file includes a `pi-setup-<project>` helper that wraps `pi setup` with the same eval pattern
- `pi shell uninstall` removes the project file and cleans the source line if no repos remain
- `pi shell list` shows all installed shortcut files
- `pi setup` runs `pi shell` as its final step, skipping in CI environments (`CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, etc.)
- `pi setup --no-shell` explicitly skips the shell step
- After `pi setup` installs shortcuts, if `PI_PARENT_EVAL_FILE` is set, `source <rc-file>` is written to the eval file for auto-sourcing

### Pre-execution requirement validation
- Before executing any steps, `RunWithInputs()` calls `ValidateRequirements()` on the automation
- If any requirement is not satisfied, execution stops immediately with a formatted error table and exit code 1
- `checkRequirementImpl(req, env, alwaysDetectVersion)` is the shared core: PATH lookup via `LookPath()`, optional version detection and comparison. `checkRequirement()` (normal execution) and `CheckRequirementForDoctor()` (always-detect) both delegate to it.
- Runtime requirements map names to commands: `python` → `python3`, `node` → `node`
- Version detection: runs `<cmd> --version`, captures stdout+stderr, extracts semver via regex `(\d+(?:\.\d+)+)`; falls back to `<cmd> version` (no `--`) when `--version` fails or produces no version string (needed for `go version` etc.)
- Handles all common version output formats: `Python 3.13.0`, `v20.11.0`, `jq-1.7.1`, `docker version 24.0.5`
- Version comparison: splits on `.`, compares numeric components pairwise; missing trailing components are treated as 0
- Install hints: built-in map of common tool names → install instructions (python, node, docker, jq, kubectl, rustc, cargo, etc.)
- Formatted error output shows automation name, missing requirements with version info, and install hints
- Testability: `RuntimeEnv.ExecOutput` field allows mocking `--version` calls without real command execution
- Validation happens after input resolution but before step/install execution
- `InstallHintFor()` is the exported version of `installHint()` for use by `pi doctor`

### Validate command (`pi validate`)
- Statically validates all config and automation files without executing anything
- Validation layers: (1) pi.yaml schema via `config.Load()`, (2) automation discovery via `discoverAll()`, (3) cross-reference checks
- Cross-reference checks: shortcuts → automations, setup entries → automations, `run:` steps → automations (including install phase steps)
- Reports all errors (not just the first) with `✗` prefixed lines to stderr
- Prints summary on success: `✓ Validated N automation(s), M shortcut(s), K setup entry(ies)`
- Exit code 0 on success, 1 on validation errors (uses `*executor.ExitError{Code: 1}` for CLI exit handling)
- Built-in automations are included in the resolution target set — `pi:install-python` references are valid
- Designed for CI pipelines: no interactive prompts, clear exit codes, structured error output
- No network requests, no command execution — purely static analysis

### Doctor command (`pi doctor`)
- Scans all automations (local + built-in), filters to those with `requires:` entries
- For each automation, checks every requirement using `CheckRequirementForDoctor()` from `internal/executor`
- Output format: per-automation section with `✓`/`✗` icons, detected version in parentheses, install hints
- Automations without `requires:` are silently skipped — not shown in output
- Exit code 0 when all requirements are satisfied, exit code 1 when any are missing
- No network requests — only PATH lookups and `--version` calls

### Requirement declarations (`requires:`)
- Automations can declare a `requires:` block listing tools/runtimes needed before execution
- `Requirement` struct has `Name`, `Kind` (RequirementRuntime or RequirementCommand), `MinVersion` (optional)
- Four forms supported in YAML:
  - `python` — runtime, any version
  - `python >= 3.11` — runtime with minimum version constraint
  - `command: docker` — command in PATH, any version
  - `command: kubectl >= 1.28` — command with minimum version
- Scalar entries are parsed as runtimes; only `python` and `node` are known runtimes — unknown names produce an error suggesting `command:` syntax
- Mapping entries with `command:` key are parsed as commands (any name is valid)
- Version constraints use `>=` operator with dot-separated numeric components (validated at parse time)
- `requires:` is valid on both `steps:` and `install:` automations
- Automations without `requires:` have an empty slice (backward compatible)
- Parsing happens via `requirementRaw.UnmarshalYAML()` with custom scalar/mapping handling

### Automation inputs (`inputs:` / `with:`)
- Automations can declare an `inputs:` block defining named parameters with type, required, default, and description
- `InputSpec` uses `*bool` for Required so we can distinguish "not set" from "set to false"
- `IsRequired()` defaults to true when no default value is provided and `required` is not explicitly set
- Input keys are stored in declaration order (`InputKeys []string`) for positional mapping
- `ResolveInputs()` on `Automation` validates and resolves inputs from either `--with` flags or positional args (mixing is an error)
- Resolved values are injected as `PI_INPUT_<NAME>` environment variables (uppercased, hyphens become underscores)
- `InputEnvVars()` converts resolved inputs to `PI_INPUT_*` env vars in sorted key order
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
- Installer automations (`install-homebrew`, `install-python`, `install-node`, `install-go`, `install-rust`, `install-uv`, `install-tsx`) use the structured `install:` block:
  - Each defines `test:`, `run:`, and optional `verify:` and `version:` fields
  - PI manages all user-facing output — automations only provide commands
  - `test` exits 0 → `✓  <name>  already installed  (<version>)`; no `run` executed
  - `test` exits non-zero → `→  <name>  installing...` → `run` executes → `verify` (or re-run `test`) → `✓  <name>  installed  (<version>)` or `✗  <name>  failed`
  - `install-homebrew` has `if: os.macos` at the automation level (skipped on non-macOS)
  - `install-python`, `install-node`, and `install-go` accept a `version` input; use step lists with `if:` conditions to try `mise` first, fall back to `brew`
  - `install-rust` accepts a `version` input; uses `rustup` if available, otherwise installs via the official `rustup.rs` installer script
  - `install-uv` uses the official `astral.sh/uv/install.sh` script
  - `install-tsx` uses `npm install -g tsx`
- Dev tool automations (`cursor/install-extensions`, `git/install-hooks`) handle common team setup tasks:
  - `cursor/install-extensions` accepts an `extensions` input (comma or newline-separated IDs), checks `cursor --list-extensions`, installs missing ones via `cursor --install-extension`
  - `git/install-hooks` accepts a `source` input (directory path relative to repo root), copies hook files to `.git/hooks/`, makes them executable; uses `cmp` for idempotency

### Sandboxed runtime provisioning (`internal/runtimes`)
- Opt-in via `runtimes:` block in `pi.yaml` — `provision: never` (default), `ask`, or `auto`
- `manager: mise` (default, falls back to direct if mise not installed) or `manager: direct`
- Provisioned runtimes are installed into `~/.pi/runtimes/<name>/<version>/bin/`
- Only `python` and `node` are known runtimes — `command:` requirements are never provisioned
- Integration point: `Executor.Provisioner` field — when set, `ValidateRequirements()` calls `tryProvision()` for failed runtime requirements
- PATH scoping: `buildEnv()` prepends provisioned bin directories to PATH for all step executions via `prependPathInEnv()`
- `buildEnv()` handles input env vars, provisioned runtime PATH, and step-level env vars; step env keys are iterated in sorted order for deterministic behavior
- Mise backend: calls `mise install <runtime>@<version>`, then `mise where` to find the install path, symlinks binaries into the managed directory
- Direct backend: downloads from official CDN (nodejs.org for node, python-build-standalone for python), extracts tar.gz, places binaries
- `Provisioner.PromptFunc` controls interactive "ask" mode — nil means non-interactive (skip provisioning)
- Already-provisioned runtimes are detected by checking for the binary at the expected path — no re-download

### Styled terminal output (`internal/display`)
- `Printer` wraps an `io.Writer` with optional ANSI color codes
- Color is auto-detected: enabled only when the writer is a `*os.File` backed by a terminal and `NO_COLOR` is not set
- `NewWithColor(w, bool)` allows explicit control for testing
- Style methods: `Plain()`, `Dim()`, `Green()`, `Red()`, `Bold()` — all accept `fmt.Sprintf`-style format strings
- `InstallStatus(icon, name, status, version)` encapsulates the icon→style mapping: `✓`+already→dim, `✓`+installed→bold green, `✗`→bold red, `→`→plain
- `SetupHeader()` renders `==>` lines in dim style
- `StepTrace(stepType, value)` renders step trace lines (`  → bash: echo hello`) in dim style with multiline collapse and 80-char truncation
- TTY detection uses `golang.org/x/term.IsTerminal()` for cross-platform reliability
- Backward compatible: tests use `bytes.Buffer` (not `*os.File`) so color is always off in tests; identical text output
- Used by: `Executor.printInstallStatus()`, `Executor.RunWithInputs()` for step trace lines, `cli/setup.go` headers, `cli/doctor.go` health output

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
- `golang.org/x/term` — TTY detection for color output
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests in `tests/integration/` build the `pi` binary and run it against `examples/` workspaces using `exec.Command`.

Total tests: 714 (115 automation + 42 builtins + 74 CLI + 30 conditions + 17 config + 30 display + 29 discovery + 170 executor [across 14 test files] + 4 project + 15 runtimes + 16 shell + 172 integration)

### Runtime skip guards
Tests that require specific runtimes use `requirePython(t)`, `requireNode(t)`, or `requireTsx(t)` helpers that call `t.Skip()` when the runtime isn't in PATH. This allows the full test suite to run on any environment — tests naturally skip rather than fail when their runtime is unavailable.

- `internal/executor/executor_test.go`: `requirePython()` and `requireTsx()` helpers for unit tests
- `tests/integration/helpers_test.go`: shared `requirePython()`, `requireNode()`, `requireTsx()` for integration tests

### Docker test matrix
Four Docker environments prove PI works on fresh systems (`make test-matrix`):

```
tests/docker/
  ubuntu-fresh/       golang:1.26.1-bookworm — Go + bash only
  ubuntu-node/        golang:1.26.1-bookworm — Go + Node 20 + tsx
  ubuntu-python/      golang:1.26.1-bookworm — Go + Python 3
  alpine-fresh/       golang:1.26.1-alpine   — Go + bash (musl libc)
```

- `tests/docker/test-matrix.sh` — builds each image, runs `go test ./...`, reports pass/fail summary
- `Makefile` — `make test-matrix` target invokes the script
- `.github/workflows/docker-matrix.yml` — CI workflow runs each environment as a separate matrix job on PRs

### Integration tests

Integration tests live in `tests/integration/` and are split by feature domain into focused test files:

```
tests/integration/
  main_test.go                    TestMain (builds pi binary), shared helpers (runPi, runPiStdout, runPiSplit, runPiWithEnv, examplesDir, findRepoRoot)
  helpers_test.go                 Runtime skip guards: requirePython, requireNode, requireTsx
  basic_test.go                   7 tests — basic example: list, greet, greet with args, build/compile, deploy (run chaining), not-found, from subdirectory
  docker_test.go                  6 tests — docker-project example: list, up, down, logs, logs with args, build-and-up (ordering)
  pipe_test.go                    3 tests — pipe example: list, upper (bash pipe), count-lines (bash→python pipe)
  version_test.go                 3 tests — version: --version flag, version subcommand, flag and subcommand match
  inputs_test.go                  8 tests — inputs: positional args, both args, --with flags, defaults, missing required, unknown input, run step with with, list INPUTS column
  info_test.go                    4 tests — info command: basic automation, with inputs, not-found, no args
  conditionals_test.go            20 tests — conditional execution: list, platform-info, skip-all, pipe passthrough, automation-level if, env/command/file predicates, complex booleans, combined automation+step if, info conditions
  builtins_test.go                25 tests — built-in automations: pi: prefix, local shadow, run step calls, docker builtins (list, info, run), installer builtins (list, info, inputs, conditions, idempotent), dev tool builtins (list, info, inputs)
  installer_schema_test.go        11 tests — installer schema: list, already-installed, fresh install, install-then-already, no-version, info type, info steps, conditional run, --silent, regular unaffected, built-in installer
  requires_test.go                7 tests — requires validation: list, satisfied command, satisfied runtime, missing command, impossible version, no-requires, install hint
  doctor_test.go                  7 tests — pi doctor: all satisfied, missing, version mismatch, skips no-requires, detected version, install hint, healthy workspace
  runtime_provisioning_test.go    7 tests — runtime provisioning: list, no-requirements, already-installed, never-mode, config parsing, auto/ask modes
  step_env_integ_test.go          5 tests — step env: build with env, multi-env isolation, list, info, info with condition
  step_visibility_integ_test.go   7 tests — step visibility: default trace, silent suppression, loud override, all-silent, all-silent loud, info silent annotation
  parent_shell_integ_test.go      7 tests — parent shell: list, eval file write, mixed steps, no eval file error, normal unaffected, info annotation, shell codegen
  step_dir_integ_test.go          6 tests — step dir: list, run in subdir, mixed dirs, dir with env, bad dir error, info annotation
  step_timeout_integ_test.go      5 tests — step timeout: list, fast completes, slow exceeds (exit 124), mixed timed/untimed, info annotation
  step_description_integ_test.go  5 tests — step description: list, run, info descriptions, info with annotations, info no-desc no-details
  validate_integ_test.go          5 tests — pi validate: valid project, invalid project, all errors reported, basic project, builtin refs
  shorthand_integ_test.go         8 tests — single-step shorthand: list, run bash, run with env, run step delegation, run with input, info, info with modifiers, validate
  polyglot_test.go                Polyglot runner tests (Python inline/file, TypeScript inline/file, multi-step pipe chains)
  shell_test.go                   Shell shortcut tests (install, uninstall, list, --repo, setup integration, --no-shell, conditional entries)
```

- Build `pi` binary once in `TestMain` (`main_test.go`)
- Run `pi list` and `pi run` against `examples/` workspaces using `exec.Command`
- Assert exit codes, output content, and step ordering
