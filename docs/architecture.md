# PI — Architecture

## Overview

PI is a single Go binary (`pi`) that reads a project's `.pi/` folder and `pi.yaml` config to discover and execute automations. The architecture is layered: parsing → discovery → execution → CLI.

## Package Structure

```
cmd/pi/main.go                     Entry point, calls cli.Execute()
internal/
  builtins/                        Embedded built-in automations
    builtins.go                    //go:embed, Discover() — walks embedded FS + Go-backed builtins, returns *discovery.Result; goBackedBuiltins() registers Go-native automations
    version_satisfies.go           pi:version-satisfies — Go-backed builtin for semver constraint checking; uses internal/semver
    builtins_test.go               66 tests (3 base + 7 docker + 24 installer + 16 devtool + 10 utility + 6 version-satisfies)
    embed_pi/                      Built-in automation YAML files (embedded at build time)
      hello.yaml                   Test built-in automation
      install-homebrew.yaml        pi:install-homebrew — macOS only, installs Homebrew
      install-python.yaml          pi:install-python — installs Python via pyenv/mise/brew (version input); detects pyenv-managed versions
      install-node.yaml            pi:install-node — installs Node.js via mise/brew (version input)
      install-go.yaml              pi:install-go — installs Go via mise/brew (version input)
      install-rust.yaml            pi:install-rust — installs Rust via rustup (version input)
      install-uv.yaml              pi:install-uv — installs uv via official installer
      install-tsx.yaml             pi:install-tsx — installs tsx via npm
      install-terraform.yaml       pi:install-terraform — installs Terraform via mise/brew (version input)
      install-kubectl.yaml         pi:install-kubectl — installs kubectl via mise/brew (version input)
      install-helm.yaml            pi:install-helm — installs Helm via mise/brew/script (version input)
      install-pnpm.yaml            pi:install-pnpm — installs pnpm via mise/npm/script (optional version input)
      install-bun.yaml             pi:install-bun — installs Bun via mise/brew/script (optional version input)
      install-deno.yaml            pi:install-deno — installs Deno via mise/brew/script (optional version input)
      install-aws-cli.yaml         pi:install-aws-cli — installs AWS CLI v2 via brew/official installer
      set-env.yaml                 pi:set-env — idempotently writes export to .zshrc/.bashrc (key, value, comment inputs)
      cursor/
        install-extensions/
          automation.yaml          pi:cursor/install-extensions — installs missing Cursor extensions (extensions input)
      docker/
        up.yaml                    pi:docker/up — docker compose up -d with v1 fallback
        down.yaml                  pi:docker/down — docker compose down with v1 fallback
        logs.yaml                  pi:docker/logs — docker compose logs with v1 fallback
      git/
        install-hooks.yaml         pi:git/install-hooks — copies hook scripts to .git/hooks/ (source input)
      npm/
        install.yaml               pi:npm/install — runs npm ci (preferred) or npm install (optional dir input)
      uv/
        sync.yaml                  pi:uv/sync — runs uv sync with optional groups and args inputs
  cli/                             Cobra CLI commands
    root.go                        Root command, wires subcommands, exit code handling
    context.go                     ProjectContext — shared project resolution pipeline (root + config + discovery + executor construction); resolveProject(), resolveProjectStrict(), Discover(), NewExecutor(), getwd(); eliminates boilerplate across all CLI commands
    discover.go                    discoverAll() / discoverAllWithConfig() — discovers local + package + built-in automations and merges; fetches/verifies packages from config; passes caller-provided stderr for all warnings (discovery, merge, on-demand fetch); newOnDemandFetcher() — creates on-demand GitHub package fetch callback with per-invocation dedup and advisory output; printOnDemandAdvisory() — prints fetch status and pi.yaml snippet to stderr
    discover_test.go               20 tests (advisory output format, nil writer, fetch status text, down arrow icon, discoverAllWithConfig [warnings route to stderr, nil stderr suppresses], discoverAll nil stderr, mergePackages [shadow warnings route to stderr, nil stderr suppresses], resolveFilePackage [existing dir, alias, missing, missing alias, not-a-dir, relative path, nil printer, nil stderr], resolvePackageSource file routing, mergePackages [empty list, missing file source skipped])
    run.go                         pi run — resolves and executes automations; SetInterspersed(false) stops flag parsing at automation name so post-name args forward naturally (no -- needed); --repo flag; --with key=value flag; --silent flag; --loud flag; all pi flags must come before automation name; uses ProjectContext for resolution and executor construction
    list.go                        pi list — discovers and prints automations with SOURCE column; --all flag shows grouped package sections; --builtins/-b flag includes pi:* automations; automationSource() resolves source indicator per automation
    info.go                        pi info — shows automation name, description, input docs with positional order (position N), automation-level env (sorted keys), if: conditions, dir: overrides, timeout: annotations, step descriptions, first: block details (lettered sub-steps with annotations), install lifecycle for installer automations (scalar/step-list phases, verify presence, version), and usage line for automations with inputs; stepAnnotations() shared helper for building annotation slices from step fields; printUsageLine() shows `pi run <name> <required> [optional]` syntax
    setup.go                       pi setup — runs setup entries (with if: support via conditions.Evaluator), then pi shell (CI-aware); --silent flag; --loud flag; color-coded headers via display.Printer; auto-source rc file via PI_PARENT_EVAL_FILE; uses ProjectContext for resolution and executor construction
    setup_add.go                   pi setup add — runs automation before appending setup entries to pi.yaml; short-form tool resolution (python → pi:install-python); --version, --if, --source, --groups flags; --only-add flag skips execution; key=value positional args; idempotent duplicate detection; auto-init when no pi.yaml; auto-generated help text via setupAddToolResolutionHelp()
    shell.go                       pi shell — installs/uninstalls/lists shell shortcuts; uses resolveProjectStrict() for project resolution; warnShadowedShortcuts() prints shadow warnings to stderr before installing
    version.go                     pi version — prints version string
    doctor.go                      pi doctor — scans all automations, checks requires: entries, prints health table; color-coded ✓/✗ via display.Printer
    validate.go                    pi validate — thin CLI wrapper over internal/validate package; resolves project, builds validate.Context, runs validate.DefaultRunner(), prints errors/success; exit 0/1
    add.go                         pi add — validates source via refparser.Parse(); fetches GitHub packages into cache; calls config.AddPackage() to update pi.yaml; --as for aliases; idempotent (duplicate sources print "already in pi.yaml" and exit successfully)
    init.go                        pi init — creates pi.yaml + .pi/ directory + .pi/hello.yaml example; interactive prompt with inferred kebab-case name; --name and --yes flags; non-interactive when piped; initProject() reusable by other commands; idempotent (already-initialized prints project name and exits 0)
    new.go                         pi new — scaffolds a new automation YAML file in .pi/; --bash, --python, --description flags; nested paths create subdirectories; strips .yaml extension if provided; refuses to overwrite existing files; suggests pi init if no project; ExampleAutomationContent constant shared with init
    completion.go                  pi completion — generates shell completion scripts (bash/zsh/fish/powershell) via Cobra's built-in generators; automationCompleter() — dynamic completion for automation names (used by run, info)
    root_test.go                   CLI tests (12 tests — includes doctor and validate subcommands)
    init_test.go                   pi init tests (13 tests — creates files, name flag, already initialized, .pi exists, yes flag, prompt, non-interactive, next steps, kebab-case, reusable initProject)
    new_test.go                    pi new tests (14 tests — basic, bash flag, python flag, description, nested path, already exists, no project, strip extension, output, valid content, generate default, generate nested, init creates example, init next steps)
    completion_test.go             pi completion tests (11 tests — bash/zsh/fish/powershell output, dynamic automation completion, description inclusion, builtin exclusion, graceful error handling)
    discover_test.go               20 tests (advisory output format, nil writer, fetch status text, down arrow icon, discoverAllWithConfig [warnings route to stderr, nil stderr suppresses], discoverAll nil stderr, mergePackages [shadow warnings route to stderr, nil stderr suppresses], resolveFilePackage [existing dir, alias, missing, missing alias, not-a-dir, relative path, nil printer, nil stderr], resolvePackageSource file routing, mergePackages [empty list, missing file source skipped])
    run_test.go                    pi run tests (24 tests — includes --with, inputs, --silent, excess positional args, PI_ARGS, natural arg forwarding, Cobra flag isolation tests)
    validate_test.go               pi validate tests (67 tests — integration tests via runValidate(); unit tests now delegate to validate.CheckWithInputs, validate.DetectCycles, validate.NormalizeCycleKey; covers valid project, broken refs, multiple errors, builtin refs, no pi.yaml, broken file refs, valid file refs, inline scripts, first: block file refs, subdir file refs, installer file refs, shortcut/setup/run-step with: validation, circular deps, condition validation)
    add_test.go                    pi add tests (8 tests — file source, file with alias, idempotent duplicate, no version error, invalid source, no pi.yaml, no args, builtin ref error)
    list_test.go                   pi list tests (11 tests — SOURCE column, --all flag, --builtins flag, package source, workspace source, INPUTS column)
    info_test.go                   pi info tests (35 tests — includes if: condition display, installer type, installer lifecycle detail [scalar, step-list, explicit verify, no version, truncation], first: block detail [basic, block annotations, descriptions, sub-step annotations, integration], dir: annotation, timeout: annotation, step description display, automation-level env display, stepAnnotations unit tests, positional order display, usage line)
    setup_add_test.go              pi setup add tests (21 tests — bare string, short-form expansion, pi: prefix, if flag, key=value, invalid key=value, duplicate, replace same run target, no pi.yaml yes, no pi.yaml non-interactive, local automation, source flag, groups flag, combined flags, invoke before writing, invoke failure, only-add, invoke not found, tool resolution help contains all builtins, tool resolution help deterministic, tool resolution help prefers canonical name)
    setup_test.go                  pi setup tests (8 tests — includes --silent, parent eval file)
    shell_test.go                  pi shell tests (5 tests — includes shadow warning integration tests)
    doctor_test.go                 pi doctor tests (16 tests — no-automations, no-requirements, satisfied, missing, mixed, skips, formatDoctorLabel unit [command ± version, runtime ± version], command with version integration, runtime requirement label integration)
  conditions/                      Boolean expression parser/evaluator for if: fields; also owns runtime predicate resolution
    conditions.go                  Lexer, AST, recursive-descent parser, Eval(), Predicates()
    evaluator.go                   RuntimeEnv struct (GOOS/GOARCH/Getenv/LookPath/Stat/ExecOutput), DefaultRuntimeEnv(), Evaluator (repoRoot + RuntimeEnv), NewEvaluator(), ShouldSkip(); ResolvePredicates(), ResolvePredicatesWithEnv() — predicate name → boolean resolution; resolveSingle() — dispatches os/arch/shell/env/command/file.exists/dir.exists predicates; ValidatePredicateName() — static predicate name validation; ValidateConditionExpr() — parses expression and validates all predicate names
    conditions_test.go             31 tests
    evaluator_test.go              38 tests (OS, arch, shell, env, command, file.exists, dir.exists predicates; unknown predicates; multiple predicates; empty list; DefaultRuntimeEnv; integration; ValidatePredicateName valid/invalid; ValidateConditionExpr valid/invalid; Evaluator.ShouldSkip true/false/empty/complex/nil-env/parse-error/unknown-predicate)
  config/                          pi.yaml parsing
    config.go                      ProjectConfig, Shortcut (with With field), SetupEntry (with If field, bare string support), PackageEntry (with Source/As, string or object form), RuntimesConfig + Load()
    config_test.go                 31 tests
    writer.go                      AddPackage() — reads pi.yaml, detects duplicates (DuplicatePackageError), insertPackageEntry() appends to packages: or creates the block; AddSetupEntry() — reads pi.yaml, finds matching entries via findMatchingEntry() (exact duplicate → DuplicateSetupEntryError, same run+if with different with → replaces in-place via replaceSetupEntry() → ReplacedSetupEntryError, no match → appends); FormatSetupEntry() renders setup entries as YAML; line-based raw string manipulation preserves unrelated file content
    writer_test.go                 39 tests (14 package + 25 setup: GitHub/file add, alias, duplicate, replace bare→versioned, replace versioned→bare, replace preserves position, replace multi-line, ReplacedSetupEntryError message, append, create block, missing pi.yaml, preserve content, multiple adds, format entry, insert entry, duplicate error, existing block with following content)
  automation/                      Individual automation YAML parsing
    automation.go                  Automation struct (with If, Env, Install, Requires, Inputs, GoFunc fields) + Load(path, warnWriter), LoadFromBytes(data, filePath, warnWriter), Dir(), IsInstaller(), IsGoFunc(), validate(), buildShorthandStep(); GoFunc field for Go-backed builtins; single-step shorthand support (top-level bash/python/typescript/run keys); top-level env: maps to automation-level env; warnWriter parameter replaces former global variable
    step.go                        StepType, Step (with If, Env, Silent, ParentShell, Dir, Timeout, Description, First, Pipe), stepRaw (YAML pipe + pipe_to), resolvePipe(index, warnWriter), toStep(index, warnWriter), toFirstStep(index, warnWriter), IsFirst(), InstallPhase, InstallSpec, validateSteps(), validateFirstBlock(), validateInstall(), validateInstallPhase()
    walker.go                      StepLocation (Phase, Index, FirstIndex, IsScalar, InFirstBlock(), FormatPath()), StepVisitor, WalkSteps(), WalkStepsUntil() — visitor-pattern traversal of all steps in an automation (regular steps, first: sub-steps, install phases); used by internal/validate for run-step-ref and file-ref checks
    inputs.go                      InputSpec, inputsRaw, ResolveInputs() (with excess positional args error), InputEnvVars()
    requirements.go                RequirementKind, Requirement, requirementRaw, parseNameVersion(), validateVersionString(), knownRuntimes map (python/node/go/rust), knownRuntimeNames()
    automation_test.go             33 tests (core load, validate, basic step parsing, single-step shorthand, automation-level env, shorthand parent_shell and with)
    step_test.go                   73 tests (if/env/silent/parent_shell/dir/timeout/description/pipe fields, install block, first: block)
    walker_test.go                 17 tests (regular steps, first: blocks, installer scalar/step-list, verify absent, early stop, FormatPath, field preservation)
    inputs_test.go                 18 tests (input spec, resolution, env vars, with: on steps, excess positional args error, exact positional args)
    requirements_test.go           22 tests (requires parsing for all 4 runtimes [python/node/go/rust], version validation, name-version parsing)
  display/                         Styled terminal output (color, TTY detection)
    display.go                     Printer struct, color methods (Plain, Dim, Green, Red, Bold, Warn), InstallStatus, SetupHeader, StepTrace, PackageFetch, truncateTrace, shouldColor; NewForWriter() — auto-detects TTY for arbitrary io.Writer (used by CLI commands that receive io.Writer stderr)
    tty.go                         isTerminal() via golang.org/x/term
    display_test.go                42 tests (styles, color toggle, NO_COLOR, TTY, install status variants, step trace, truncateTrace, package fetch status [including ⚠ warning icon], Warn style, NewForWriter)
  discovery/                       .pi/ folder scanning and automation lookup
    discovery.go                   Discover() (with warnWriter for name mismatch warnings), NewResult(), Result, Find()/FindWithAliases() (uses refparser for reference classification), MergeBuiltins(), MergePackage(), IsBuiltin(), IsPackage(), PackageSource(), PackageAutomations(), KnownAliases(), reconcileAutomationName(); findAlias() and findInPackage() for package resolution; OnDemandFetchFunc callback type and OnDemandFetch field for on-demand GitHub package fetching; "Did you mean?" suggestions via Levenshtein distance on not-found errors
    suggest.go                     levenshtein() — Wagner–Fischer edit distance; suggestNames() — returns close matches sorted by distance then alphabetically; used by findLocal() and findBuiltin() for "Did you mean?" error messages
    suggest_test.go                13 tests (Levenshtein distance, symmetry, suggestion sorting, max results, exact match exclusion, empty inputs, integration with Find)
    discovery_test.go              43 tests (18 base + 6 builtin merge/prefix + 5 optional name tests + 8 package merge/alias tests + 6 on-demand fetch tests)
  executor/                        Step execution engine
    executor.go                    Executor struct (with ParentEvalFile, Runners, cachedRegistry, stepOutputs, condEval fields), stepExecCtx (bundles per-automation execution state: automation, args, inputEnv — constant across all steps in a RunWithInputs call), ExitError, Run(), RunWithInputs(), execStep(), execStepSuppressed(), execParentShell(), execFirstBlock(), AppendToParentEval(), evaluateCondition() (delegates to conditions.Evaluator), conditionEvaluator() (lazy-init), pushCall()/popCall(), printer(), registry() (lazy-init with caching), newRunContext(workDir) — caller resolves dir: once and passes result, stdout()/stderr()/stdin(), lastOutput(), recordOutput(), interpolateWithCtx(), interpolateWith(), interpolateValue(); GoFunc automation dispatch; pipe: true orchestration; outputs.last inter-step output capture (via io.MultiWriter); step-level and automation-level if: conditional execution; automation-level env merged per step (not passed into run: sub-automations); step-level silent: true suppression; --loud override; parent_shell: true eval-file delegation; step dispatch via Registry; dir: resolved once per step in execStep; first: block first-match dispatch; PI_ARGS env var injection for extra args on automations without inputs; step error messages include automation name context
    runner_iface.go                StepRunner interface, RunContext (step execution context with WorkDir, RuntimePaths, InterpolateWith callback — no BuildEnv callback; env construction uses standalone BuildStepEnv), Registry (maps StepType→StepRunner), NewRegistry(), NewDefaultRegistry()
    runners.go                     Step runner implementations: SubprocessRunner (config-driven, shared by bash/python/typescript), SubprocessConfig (Binary/BinaryFunc, InlineArgs, TempFileExt, NotFoundMsg, Language), NewBashRunner(), NewPythonRunner(), NewTypeScriptRunner(), RunStepRunner; runStepCommand() shared command execution with timeout support (exec.CommandContext), calls BuildStepEnv(ctx.RuntimePaths, ...) directly; TimeoutExitCode (124); writeTempScript(), resolvePythonBin(), isCommandNotFound()
    install.go                     Installer lifecycle: execInstall(), execInstallPhase(), execInstallPhaseLive(), execInstallPhaseWithStderr(), execInstallFirstBlock(), captureVersion(), printInstallStatus(); structured test→run→verify→version lifecycle; color-coded installer status via display.Printer; all install phases (scalar and step-list) dispatch through the step runner Registry; first: block support in install phases; run phase streams stderr live to terminal; automation-level env properly applied to all phases; execInstallPhaseWithStderr() is the unified implementation — execInstallPhase() and execInstallPhaseLive() are thin wrappers that pass io.Discard or e.stderr() respectively
    helpers.go                     Shared utilities: resolveFileStep() (file-path resolution + existence check), IsFilePath() (exported), resolveScriptPath(), BuildStepEnv(runtimePaths, inputEnv, automationEnv, stepEnv) — standalone function (no Executor dependency), prependPathInEnv(), resolveStepDir(), expandTraceVars(value, inputEnv, automationEnv, stepEnv); PI_IN_* + PI_INPUT_* + PI_ARGS env injection; provisioned runtime PATH injection; automation-level and step-level env: injection; step-level dir: resolution; trace variable expansion for display
    validate.go                    ValidateRequirements() (with provisioning fallback), tryProvision(), checkRequirementImpl() (shared logic with alwaysDetectVersion flag), checkRequirement(), CheckRequirementForDoctor(), detectVersion(), extractVersion(), compareVersions(), FormatValidationError(), InstallHintFor(), CheckResult, ValidationError, installHints; pre-execution requirement validation
    predicates.go                  Thin delegation layer — re-exports conditions.RuntimeEnv (type alias), conditions.DefaultRuntimeEnv(), conditions.ResolvePredicates(), conditions.ResolvePredicatesWithEnv(), conditions.ValidatePredicateName(), conditions.ValidateConditionExpr(); keeps backward compatibility for packages that import from executor
    test_helpers_test.go           Shared test helpers: newAutomation, newAutomationInDir, newExecutor, newExecutorWithCapture, newExecutorWithEnv, step constructors (bashStep, runStep, pythonStep, typescriptStep, pipedBashStep, pipedPythonStep, bashStepIf), fakeRuntimeEnv, requirePython, requireTsx, boolPtr
    first_block_test.go            14 tests — first: block: first/middle/fallback matches, none match, mixed steps, pipe to next, pipe no match, outer if skip, run sub-step, exit error, install phase, silent, loud override
    outputs_test.go                16 tests — outputs.last: passed via with, multi-step chain, pipe capture, silent capture, indexed output, inputs interpolation, combined outputs+inputs, literal passthrough, GoFunc via run step, GoFunc failure, GoFunc condition skip, interpolateValue unit tests
    executor_test.go               20 tests — core execution: bash inline/file, run step chaining, circular deps, multi-step, working dir, mixed bash+run, exit error, IsFilePath, call stack isolation
    python_runner_test.go          9 tests — python inline/file, venv detection, mixed bash+python
    typescript_runner_test.go      8 tests — typescript inline/file, tsx not found, mixed bash+typescript
    pipe_test.go                   10 tests — pipe: true: bash→bash, bash→python, python→bash, three-step chain, failure propagation, stderr passthrough, run step piping, multiline data
    inputs_test.go                 14 tests — RunWithInputs: env var injection, positional, defaults, missing required, mixing error, excess positional args, args passthrough, short prefix, both prefixes, run step with with, PI_ARGS set/not-set/consumed/single
    conditional_step_test.go       13 tests — step-level if: true/false/not/complex, mixed conditional+unconditional, pipe passthrough on skip, file.exists/not
    conditional_automation_test.go 7 tests — automation-level if: true/false, run step calling skipped/executed automation, complex condition, skip vs circular dependency
    install_test.go                17 tests — installer lifecycle: already installed, fresh install, run fails, verify fails, verify defaults to test, no version, silent, stderr on failure (with content assertion), scalar stderr streamed, step list stderr streamed, first list with conditionals, with inputs, first block stderr surfaced, scalar phase uses automation env, version capture uses automation env, automation-level if
    step_env_test.go               14 tests — step-level env, automation-level env, bash/python, multiple vars, parent override, nil env inheritance, per-step isolation, BuildStepEnv layers, BuildStepEnv deterministic order, run: isolation
    step_dir_test.go               10 tests — step-level dir: bash inline/absolute/default, missing dir error, not-a-dir error, python step, per-step isolation, mixed with no dir, combined with env, resolveStepDir unit tests
    step_trace_test.go             22 tests — step trace lines, silent step suppression, loud override, silent still executes, silent pipe capture, expandTraceVars unit tests (11 cases: no vars, input vars, braced syntax, multiple vars, automation env, step env, override priority, unknown passthrough, mixed, deprecated prefix), input/automation/step env trace expansion, first: block trace expansion
    step_timeout_test.go           8 tests — step-level timeout: no timeout runs normally, not exceeded, exceeded (killed with exit 124), stops execution chain, with pipe: true, with silent, skipped by condition, multiple steps only timed-out killed
    parent_shell_test.go           6 tests — parent shell: writes to eval file, multiple steps append, mixed with normal, no eval file error, skipped by condition, AppendToParentEval
    coverage_gaps_test.go          53 tests — coverage gap closures: interpolateWith (empty/map/outputs/literal/multi-key), stdout/stderr/stdin nil fallback and explicit, printer lazy creation and explicit, evaluateCondition parse errors and unknown predicates (unit + automation-level + step-level), AppendToParentEval error paths (invalid path, read-only), registry custom/cached/precedence, first: block no-match pipe capture + parent shell sub-step + condition error, resolveScriptPath absolute/relative, unimplemented step type, resolveStepDir branches, isCommandNotFound, writeTempScript, resolvePythonBin ± venv, InstallHintFor known/unknown, FormatValidationError with/without hint + skip satisfied, formatRequirementLabel branches, runtimeCommand branches, detectVersion mocked (--version and version subcommand fallback), extractVersion patterns, compareVersions (various + parse error), BuildStepEnv (empty/runtimePaths only), stepExecCtx roundtrip
    validate_test.go               34 tests (version extraction, version comparison, requirement checking, validation integration, error formatting, install hints, CheckRequirementForDoctor, InstallHintFor, provisioning integration, prependPathInEnv, BuildStepEnv)
    predicates_test.go             5 delegation smoke tests (ResolvePredicatesWithEnv, DefaultRuntimeEnv, ResolvePredicates integration, ValidatePredicateName, ValidateConditionExpr); comprehensive tests are in conditions/evaluator_test.go
  cache/                           GitHub package cache manager
    cache.go                       Cache struct, Fetch(), PackagePath(), IsMutableRef(), cloneAndCache(), cloneRepo() (SSH/token/HTTPS fallback), execGit()
    package_yaml.go                PackageYAML struct, checkPackageYAML(), versionSatisfies(); optional min_pi_version enforcement
    cache_test.go                  17 tests (cache hit/miss, SSH/token/HTTPS auth order, mutable refs, atomic writes, repo files, private repo errors)
    package_yaml_test.go           15 tests (absent/empty/satisfied/unsatisfied/dev/invalid YAML, v-prefix, integration with Fetch)
  refparser/                       Automation reference string parser
    refparser.go                   Parse(), AutomationRef, RefType (Local/Builtin/GitHub/File/Alias); pure string parsing — no I/O
    refparser_test.go              46 tests (local, builtin, github, file, alias, errors, round-trip, precedence)
  project/                         Project root detection
    root.go                        FindRoot() — walks up to find pi.yaml
    root_test.go                   4 tests
  runtimes/                        Sandboxed runtime provisioning
    runtimes.go                    KnownRuntimes map (python/node/go/rust), Provisioner, Provision(), provisionWithMise(), provisionDirect(), PrependToPath(), knownRuntimeList()
    runtimes_test.go               25 tests (provisioner defaults, config, never mode, unknown runtime, already provisioned, ask mode [nil prompt, user declines, user accepts, version in prompt, no version in prompt], binDirFor [not provisioned, provisioned, default version], prependToPath, defaultVersion, runtimeBinary, knownRuntimes, mise fallback, provisionDirect [go unsupported, rust unsupported, unknown runtime], unknown manager, binDir default version path, stderr default, mise integration)
  semver/                          Semver-aware version constraint checking
    semver.go                      Satisfies() — checks version against constraint; normalises incomplete versions; wraps Masterminds/semver/v3; isChannelName() — detects non-semver channel names (stable, nightly, beta) and accepts any valid version
    semver_test.go                 50 tests (exact, >=, ^, ~, range, v-prefix, error cases, practical installer scenarios, channel names, isChannelName unit)
  shell/                           Shell shortcut file generation and management
    shell.go                       GenerateShellFile(), Install(), Uninstall(), ListInstalled(), PrimaryRCFile(), GenerateCompletionScript(); with: shortcut codegen; PI_PARENT_EVAL_FILE eval wrapper pattern; pi-setup-<project> helper function; shell completion script generation
    shadow.go                      ShadowWarning, CheckShadowedNames(), FormatWarning(); shellBuiltins (36 POSIX builtins) and commonCommands (30 system commands) blocklists; warns when shortcut names shadow known commands
    shell_test.go                  20 tests (includes completion script generation and lifecycle tests)
    shadow_test.go                 13 tests (no shadows, shell builtins, common commands, multiple, empty, case-insensitive, exhaustive coverage, suggestions, formatting, sorted output)
  validate/                        Project-wide static validation (registry-based)
    validate.go                    Check interface, CheckFunc adapter, Context (Root/Config/Discovery), Result, Runner (register+run), DefaultRunner() — pre-loads 9 built-in checks; individual check implementations: checkShortcutRefs, checkSetupRefs, checkRunStepRefs, checkFileReferences, checkShortcutInputs, checkSetupInputs, checkRunStepInputs, checkCircularDeps, checkConditions; exported helpers: CheckWithInputs(), BuildRunGraph(), DetectCycles(), NormalizeCycleKey()
    validate_test.go               55 tests (runner lifecycle [empty, register, aggregation, counts, ordering], DefaultRunner check count, CheckFunc adapter, CheckWithInputs [no-with, all-valid, unknown, no-inputs, multiple-sorted], DetectCycles [no-cycles, direct, self-loop, three-node, diamond, multiple, disconnected], NormalizeCycleKey [rotation, self-loop, single, empty], individual checks [shortcut-refs valid/broken, setup-refs valid/broken, run-step-refs valid/broken, circular-deps cycle/no-cycle, conditions valid/unknown/automation-level, shortcut-inputs valid/unknown, setup-inputs valid/unknown/no-with/broken-ref, run-step-inputs valid/unknown/no-with/broken-ref/bash-skipped, file-refs inline/run-step/missing/existing], BuildRunGraph [basic, dedup])
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
pi list [--all] [--builtins]
  │
  ├─ CLI (internal/cli)
  │    Parses args (--all, --builtins), gets CWD
  │
  ├─ Project (internal/project)
  │    Walks up from CWD to find pi.yaml → repo root path
  │
  ├─ Config (internal/config)
  │    Loads pi.yaml → ProjectConfig (packages, shortcuts)
  │
  └─ Discovery (internal/discovery + internal/builtins)
       discoverAllWithConfig(): local .pi/ + packages + builtins → merged Result
       Names() → sorted list → formatted table with SOURCE column
       Filters: builtins hidden by default (--builtins shows them)
       --all: appends grouped package sections via printPackageAutomations()
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
  ├─ Project + Config
  │    Loads pi.yaml (including packages: block)
  │
  ├─ Package Fetch (internal/cli/discover.go + internal/cache)
  │    For each packages: entry:
  │      file: → verify directory exists on disk (warn if missing)
  │      GitHub → cache.Fetch() (clone if not cached)
  │    Status output: ↓/✓/✗/⚠ per package
  │
  ├─ Discovery (internal/discovery)
  │    discoverAllWithConfig(): local .pi/ + packages + builtins → merged Result
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
  └─ Validation (internal/validate via DefaultRunner)
       9 registered checks run in order:
       1. shortcut-refs → automation names
       2. setup-refs → automation names
       3. run-step-refs → automation names (incl. install phases)
       4. file-refs → files on disk
       5. shortcut-inputs → target declared inputs
       6. setup-inputs → target declared inputs
       7. run-step-inputs → target declared inputs
       8. circular-deps → DFS graph analysis
       9. conditions → known predicate names
       Collects all errors, CLI prints to stderr, exit 0 or 1
```

```
pi completion <shell>
  │
  ├─ CLI (internal/cli/completion.go)
  │    Parses shell name (bash/zsh/fish/powershell)
  │
  └─ Cobra built-in generators
       GenBashCompletionV2() / GenZshCompletion() / GenFishCompletion() / GenPowerShellCompletionWithDesc()
       Outputs completion script to stdout
```

```
pi run <TAB>  (dynamic completion)
  │
  ├─ automationCompleter() (internal/cli/completion.go)
  │    os.Getwd() → resolveProject() → Discover()
  │
  ├─ Discovery (internal/discovery)
  │    Names() → filtered (no builtins)
  │
  └─ Output
       Returns automation names with descriptions as tab-separated completion entries
       Errors silently return empty list
```

```
pi init [--name <name>] [--yes]
  │
  ├─ CLI (internal/cli/init.go)
  │    Parses flags (--name, --yes/-y)
  │
  ├─ Already initialized?
  │    os.Stat(pi.yaml) → print "Already initialized" + next steps, exit 0
  │
  ├─ Resolve project name
  │    --name flag → use directly
  │    --yes or non-interactive → infer from directory (kebab-case)
  │    Interactive → prompt with inferred default
  │
  └─ initProject(root, name)
       Write pi.yaml (project: <name>)
       Create .pi/ directory
       Print confirmation + "Next steps" guide
```

```
pi add <source> [--as <alias>]
  │
  ├─ CLI (internal/cli/add.go)
  │    Parses args (--as), requires project root with pi.yaml
  │
  ├─ Refparser (internal/refparser)
  │    Parse(source) — rejects invalid refs (e.g. built-ins without a version where required)
  │
  ├─ Project + Cache (internal/project, internal/cache)
  │    GitHub sources → cache.Fetch() before writing config
  │    file: sources — no fetch; validated by refparser / add path
  │
  └─ Config writer (internal/config/writer.go)
       AddPackage() — duplicate normalized entries → success + "already in pi.yaml"
       Else append YAML line(s) under packages: (create block if missing)
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

### Automation reference parsing (`internal/refparser`)
- All automation references pass through `refparser.Parse()` before resolution
- Five reference types: `RefLocal` (`.pi/` path), `RefBuiltin` (`pi:` prefix), `RefGitHub` (`org/repo@version[/path]`), `RefFile` (`file:` prefix), `RefAlias` (alias-prefixed path)
- Detection precedence: `pi:` → `file:` → `@` (GitHub) → alias match → local (default)
- `AutomationRef` struct holds all parsed fields: `Type`, `Raw`, `Path`, `Org`, `Repo`, `Version`, `FSPath`, `Alias`
- `String()` produces the canonical form; round-trip (`Parse` → `String`) preserves canonical representation
- Pure string parsing — no filesystem or network access; safe to call from validators and completers
- `discovery.Find()` delegates to `refparser.Parse()` internally; local and builtin refs resolve as before
- `discovery.FindWithAliases()` accepts `knownAliases` for alias resolution (used when packages are configured)
- GitHub, file, and alias refs currently return clear "not yet supported" errors — infrastructure for project 13

### Single-step shorthand
- Automations with a single step can place the step type key (`bash:`, `python:`, `typescript:`, `run:`) at the top level, skipping the `steps:` wrapper
- Step modifier fields (`env:`, `dir:`, `timeout:`, `silent:`, `pipe:`, `parent_shell:`, `with:`) are supported alongside the shorthand key at the top level
- Top-level `if:` maps to the automation-level condition (not a step-level condition) — for a single-step automation, these are semantically equivalent
- Top-level `description:` remains the automation description (not a step description)
- Having both a top-level step key and `steps:` (or `install:`) in the same file is a parse error
- Having multiple top-level step keys (e.g. both `bash:` and `python:`) is a parse error
- Implementation: `buildShorthandStep()` in `automation.go` constructs a `stepRaw` from top-level keys; `UnmarshalYAML` expands it to a single-element `Steps` slice before validation — the rest of the system (executor, info, validate) sees a normal automation
- Shorthand is additive — existing `steps:` syntax continues to work unchanged

### External packages (`packages:` in `pi.yaml`)
- Teams declare external automation sources in `pi.yaml → packages:`
- Two source types: GitHub (`org/repo@version`) and local file system (`file:~/path` or `file:./relative`)
- `PackageEntry` supports both simple string form and object form (`source:` + optional `as:` alias)
- Aliases let you write `run: mytools/docker/up` instead of the full source path
- `discoverAllWithConfig()` orchestrates the pipeline: discover local → fetch/verify packages → discover package automations → merge into result → merge builtins
- GitHub packages are fetched via `cache.Cache` — SSH/token/HTTPS auth fallback, atomic writes
- `file:` sources are verified to exist; missing ones print a warning but don't halt (non-fatal)
- Relative `file:` paths are resolved relative to the project root
- `discovery.Result.MergePackage()` discovers automations from each package's `.pi/` and merges them
- Package automations that don't collide with local names are added to the main `Automations` map — accessible via plain names in `run:` steps
- Resolution priority: local `.pi/` > package automations > builtins — local always wins with a shadow warning
- Alias resolution: `refparser.Parse()` detects alias-prefixed paths when `knownAliases` is provided; `FindWithAliases()` auto-populates aliases from the Result's own map
- `pi setup` shows package fetch status before running setup automations (↓ fetching, ✓ cached/found, ✗ failed, ⚠ not found)
- All CLI commands (run, list, info, validate, doctor) use `discoverAllWithConfig()` to see package automations
- Validation: duplicate aliases are a parse error; aliases with `/` are rejected; empty source is rejected

### On-demand GitHub package fetching
- When a `run:` step references an undeclared GitHub package (e.g. `org/repo@v1.0/path`), PI fetches it automatically instead of failing
- `OnDemandFetchFunc` type on `discovery.Result` — callback invoked by `findInPackage()` when a GitHub ref has no declared package
- `newOnDemandFetcher()` in `cli/discover.go` creates the callback with a closure-scoped `fetched` map for per-invocation dedup
- The callback uses the same `cache.Cache` as declared packages — identical auth fallback (SSH/token/HTTPS)
- Advisory output: shown only when a live network fetch happens (not from cache); prints `↓ fetched (on demand)` status plus a `tip:` with a ready-to-paste `packages:` YAML snippet
- Advisory is written to stderr so it doesn't interfere with piped automation output
- `file:` refs are never fetched on demand — `findInPackage()` checks `ref.Type == refparser.RefFile` and returns an error immediately
- `PackageAutomations()` method on `Result` exposes the per-source automation map for lookup after on-demand merge

### `pi init` command (`internal/cli/init.go`)
- `pi init [--name <name>] [--yes]` bootstraps a new PI project
- Creates `pi.yaml` with `project: <name>` and an empty `.pi/` directory
- Project name resolution: `--name` flag → use as-is; `--yes` flag or non-interactive stdin → infer from directory name (kebab-cased); interactive → prompt with inferred default
- Kebab-case conversion: lowercase, spaces/underscores → hyphens, strip non-alphanumeric characters, trim leading/trailing hyphens
- Non-interactive detection: `isInteractive()` checks if stdin is a `*os.File` backed by a character device; piped input (CI, scripts) silently uses the inferred name
- Already-initialized handling: if `pi.yaml` exists, prints "Already initialized (project: <name>)" with the existing project name, shows "Next steps", exits 0 — not an error
- If `.pi/` exists but `pi.yaml` does not, only `pi.yaml` is created (`.pi/` creation is a no-op via `MkdirAll`)
- `initProject(root, name, stdout)` is the exported core logic — writes the files and prints confirmation; callable from other commands (e.g. `pi setup add` when no project exists)
- "Next steps" block shown on both success and already-initialized paths: suggests `pi setup add`, `pi shell`, `pi run`
- Output uses `display.Printer` for consistent styling: green for "Initialized project", dim for file creation lines and next steps

### `pi setup add` command (`internal/cli/setup_add.go`, `internal/config/writer.go`)
- `pi setup add <name> [key=value ...] [--version <v>] [--if <expr>] [--source <path>] [--groups <list>] [--yes] [--only-add]`
- **Invoke-before-write**: by default, the automation is executed first; if it succeeds, the entry is written to `pi.yaml`; if it fails, `pi.yaml` is untouched. This prevents misconfigured or failing entries from polluting the config.
- `--only-add`: skips execution and writes the entry directly to `pi.yaml` (the original behavior); useful for CI bootstrapping or when the tool is already set up
- `invokeSetupAutomation()` reuses the same `ProjectContext → Discover → NewExecutor → RunWithInputs` pipeline as `pi setup`, applied to a single entry
- Short-form tool name resolution via `setupAddKnownTools` map: `python` → `pi:install-python`, `pi:go` → `pi:install-go`, `rust` → `pi:install-rust`, etc.
- `setupAddToolResolutionHelp()` auto-generates the help text "Tool names are resolved automatically" section from the map; deduplicates aliases, prefers canonical names matching `pi:install-<name>` suffix, sorts alphabetically
- Resolution advisory: when a short-form expansion happens, prints `Resolved '<input>' → <expanded>` (philosophy principle 6: explain the magic)
- YAML output: bare string form when no modifiers, object form (`run:` + `if:` + `with:`) when modifiers are present
- `--version`, `--source`, `--groups` flags populate `with:` entries; `--if` sets the condition
- Positional `key=value` arguments add additional `with:` entries (e.g. `file=.pi/cursor/extensions.txt`)
- Idempotent: if the exact entry already exists (same run, same if, same with map), prints "Already in pi.yaml" and exits 0
- Same run with different with: replaces the existing entry in-place (preserving position in the list), prints "Replaced in pi.yaml" and exits 0; this prevents accidental duplicate entries (e.g. `pi setup add node` then `pi setup add node --version 22`)
- No pi.yaml: offers to initialize the project first; `--yes` or non-interactive auto-accepts; declined exits 1
- Uses `initProject()` from `init.go` for project initialization — same output and behavior as `pi init`
- Uses `config.AddSetupEntry()` for file manipulation — same line-based approach as `config.AddPackage()`
- `config.FormatSetupEntry()` renders entries for both confirmation output and file insertion
- `config.DuplicateSetupEntryError` type enables the CLI to distinguish exact duplicates (no-op) from real errors
- `config.ReplacedSetupEntryError` type enables the CLI to distinguish replacements ("Replaced in pi.yaml") from fresh adds ("Added to setup in pi.yaml")

### `pi add` command (`internal/cli/add.go`, `internal/config/writer.go`)
- `pi add <source> [--as <alias>]` is the ergonomic entry point for declaring a package dependency
- Source is validated via `refparser.Parse()`: only `RefGitHub` and `RefFile` types are accepted; `RefLocal`, `RefBuiltin`, etc. return a clear error
- GitHub sources without `@version` are detected specially: if refparser classifies `org/repo` as `RefLocal` (no `@`), `validateGitHubSource()` detects the `org/repo` pattern and returns "version required — use pi add org/repo@<tag>"
- For GitHub sources, the package is fetched into cache before writing to `pi.yaml` — fetch failure prevents the config change
- For `file:` sources, no fetching is needed — the entry is written directly
- `config.AddPackage()` reads the existing `pi.yaml`, checks for duplicates via source string matching, and writes back
- Duplicate detection returns `*DuplicatePackageError` which the CLI handles as a success: prints "already in pi.yaml" and exits 0
- File modification uses line-based string manipulation (`insertPackageEntry`): finds the `packages:` block, locates the end of its list items, and appends the new entry. If no `packages:` block exists, one is appended at the end of the file
- After writing, `config.Load()` is called to re-validate the updated file — catches any corruption
- `formatPackageEntry()` renders simple sources as `  - source` and aliased sources as `  - source: ...\n    as: ...`
- The command is idempotent by design — the same package can be added multiple times without duplication

### GitHub package cache (`internal/cache`)
- `Cache` struct holds configuration: `Root` (cache directory), `WarnWriter` (for mutable ref warnings), `PIVersion` (for version checks), `GitFunc` (injectable git executor), `GetenvFunc` (injectable env reader)
- `DefaultCacheRoot()` returns `~/.pi/cache`
- `PackagePath(org, repo, version)` returns `<root>/github/<org>/<repo>/<version>/`
- `Fetch(org, repo, version)` is the main entry point: checks cache hit first, clones on miss
- Cache hit: `os.Stat()` the target directory — if exists, return immediately with no network call
- Cache miss: `cloneAndCache()` → clone into temp dir → checkout version → remove `.git` → atomic `os.Rename()` to final path
- Atomic writes: if any step fails, `defer os.RemoveAll(tmpDir)` ensures no partial cache entry remains
- Auth fallback chain: SSH (`git@github.com:org/repo.git`) → HTTPS with `GITHUB_TOKEN` → plain HTTPS. Each attempt uses a fresh temp dir
- Mutable refs (`main`, `master`, `HEAD`): date-stamped cache key (e.g. `main~20260405`) and warning to stderr
- `pi-package.yaml`: optional file in package root; only `min_pi_version` field is checked. Dev builds skip the check
- `versionSatisfies()` does component-wise numeric comparison with `v` prefix stripping
- `GitFunc` and `GetenvFunc` allow full test isolation without real git or environment variables

### Package boundaries
- `config` knows only about `pi.yaml` structure; includes `PackageEntry` (source + optional alias), `PackageAliases()` helper
- `automation` knows only about a single automation file's structure; also stores `FilePath` for resolving relative script paths; `Load()`/`LoadFromBytes()` accept a `warnWriter io.Writer` parameter for deprecation warnings (no global state); `WalkSteps()`/`WalkStepsUntil()` provide visitor-pattern traversal of all steps (used by `internal/validate` and future analysis passes)
- `cache` manages `~/.pi/cache/` — clones GitHub repos at specific versions, handles auth fallback, validates `pi-package.yaml`; depends on `gopkg.in/yaml.v3` for package YAML parsing; injectable `GitFunc`/`GetenvFunc` for testing
- `conditions` is a pure-logic package for parsing and evaluating boolean `if:` expressions — zero dependencies on other PI packages
- `refparser` is a pure-logic package for parsing automation reference strings into typed structs — zero dependencies on other PI packages (uses only `os` for tilde expansion)
- `discovery` ties them together: walks the filesystem, calls `automation.Load()` for each file, builds the name→automation map; uses `refparser` to classify references in `Find()`/`FindWithAliases()`; `MergePackage()` discovers and merges package automations with alias tracking
- `semver` is a pure-logic package for version constraint checking — wraps `Masterminds/semver/v3` with normalisation; no dependencies on other PI packages
- `executor` runs automation steps; depends on `automation` (types), `discovery` (for `run:` step resolution), and provides `outputs.last` interpolation
- `builtins` embeds YAML files and registers Go-backed builtins; depends on `automation`, `discovery`, and `semver` (for `version-satisfies`)
- `project` handles finding the repo root (directory containing `pi.yaml`)
- `validate` provides a registry-based validation system with `Check` interface, `Runner`, and `DefaultRunner()`; depends on `automation`, `config`, `discovery`, `conditions`, and `executor` (for `IsFilePath`)
- `cli` ties project + config + discovery + cache + executor + validate together; `ProjectContext` in `context.go` provides the shared resolution pipeline; `discoverAllWithConfig()` orchestrates package fetching and automation discovery

### CLI project resolution (`ProjectContext`)
- `ProjectContext` encapsulates the common CLI resolution pipeline: project root + config + discovery + executor construction
- `resolveProject(startDir)` finds root and loads config (ignoring config errors — used by `run`, `list`, `info`, `doctor`, `add`)
- `resolveProjectStrict(startDir)` finds root and loads config strictly (returning errors — used by `setup`, `shell install/uninstall`)
- `(pc *ProjectContext).Discover(stderr)` runs full discovery (local + packages + builtins)
- `(pc *ProjectContext).NewExecutor(result, opts)` builds an `Executor` with ParentEvalFile from env and Provisioner from config
- `ExecutorOpts` bundles Stdout, Stderr, Silent, Loud for executor construction
- `getwd()` is a shared helper wrapping `os.Getwd()` with consistent error messaging
- All CLI commands use this pipeline, eliminating duplicated boilerplate across `run.go`, `list.go`, `info.go`, `doctor.go`, `validate.go`, `setup.go`, `add.go`, `shell.go`
- Adding a new CLI command requires only: call `resolveProject()`, then `pc.Discover()` and optionally `pc.NewExecutor()` — ~3 lines instead of ~15

### Execution model
- By default, all steps run with the repo root (directory containing `pi.yaml`) as their working directory. Steps can override this with `dir:`.
- Step runner registry: `Registry` in `runner_iface.go` maps `StepType` → `StepRunner`. `NewDefaultRegistry()` registers config-driven `SubprocessRunner` instances for bash/python/typescript plus `RunStepRunner`. `Executor.execStep()` dispatches through the registry instead of a switch statement. Adding a new step type only requires implementing `StepRunner` and registering it — no executor changes needed.
- `SubprocessRunner` in `runners.go` is a config-driven runner that handles the full subprocess lifecycle: file-path resolution → inline-script handling (via `InlineArgs` or temp-file) → command execution → error wrapping. `SubprocessConfig` captures per-language differences (binary, inline arg format, temp file extension, not-found message). All language-specific runners (`NewBashRunner`, `NewPythonRunner`, `NewTypeScriptRunner`) are thin config constructors. Adding a new language runner requires only a ~10-line config struct — no new types needed.
- `RunContext` in `runner_iface.go` bundles everything a runner needs: automation, step, args, I/O writers, env, repo root, and `WorkDir` (the resolved working directory). Runners are decoupled from `Executor` internals — they receive a callback for recursive `run:` calls.
- Common execution substrate: `runStepCommand()` in `runners.go` handles `exec.Command` setup (Dir from `WorkDir`, Env, Stdout, Stderr, Stdin) and error wrapping (`*ExitError` for non-zero exits). All language runners delegate to this via `SubprocessRunner`.
- File-path resolution: `resolveFileStep()` in `helpers.go` combines `IsFilePath()` + `resolveScriptPath()` + existence check into a single call, eliminating per-runner duplication. `IsFilePath()` is exported so `pi validate` can reuse the same detection logic for static file reference validation.
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
- Adding a new subprocess step type: (1) register the `StepType` constant in `automation.go` and add it to `validStepTypes`, (2) create a `New<Lang>Runner()` function in `runners.go` that returns a `&SubprocessRunner{Config: SubprocessConfig{...}}` with the language-specific config, (3) register it in `NewDefaultRegistry()` in `runner_iface.go` — no new types or executor changes needed. For non-subprocess step types (like `run:`), implement the `StepRunner` interface directly.

### Installer automation lifecycle (`install:` block)
- Automations can use `install:` instead of `steps:` — the two are mutually exclusive
- `InstallSpec` has four fields: `Test`, `Run`, `Verify` (optional), `Version` (optional)
- `InstallPhase` is polymorphic: either a scalar bash string or a list of steps (same step schema as `steps:`)
- When `verify:` is absent, the `test:` phase is re-run as verification after `run:` completes
- Executor runs `execInstall()`: test → [run → verify] → version
- All install phases (scalar and step-list) dispatch through the step runner Registry, ensuring consistent behavior: automation-level env, file-path resolution, and future features like timeout all work uniformly
- `execInstallPhaseWithStderr()` is the unified install phase executor — takes an explicit `stderrWriter io.Writer` parameter. `execInstallPhase()` (suppressed, passes `io.Discard`) and `execInstallPhaseLive()` (live, passes `e.stderr()`) are thin wrappers. `execInstallFirstBlock()` similarly takes an explicit stderr writer.
- All phase stdout is suppressed; stderr from `run:` is streamed live to the terminal so the user sees errors as they happen; test/verify phase stderr is suppressed
- `version:` command stdout is trimmed and displayed in the status line; version capture dispatches through the runner registry so automation-level env is available
- PI prints one formatted status line per installer: `✓ / → / ✗  name  status  (version)`
- `--silent` flag on `pi run` and `pi setup` suppresses PI status lines (stderr from failures always shown)
- Step lists in install phases support `if:` conditions, all step types (bash, python, typescript), and `run:` references. Step dispatch uses the same `Registry` as normal steps.
- A `run:` step inside an `install:` phase that references another installer automation runs that automation's own `install:` lifecycle

### Pipe support (`pipe: true`)
- `pipe_to: next` is the deprecated form — still accepted for backward compatibility but emits a deprecation warning at parse time
- When a step declares `pipe: true`, its stdout is captured to a `bytes.Buffer` instead of printed to terminal
- The captured buffer is fed as stdin to the next step
- If `pipe: true` appears on the last step, it's a no-op — output goes to terminal normally
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
- Compatible with all other step fields: `if:`, `env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`, `pipe: true`

### Step trace lines and silent/loud
- Before executing each non-installer step, PI prints a trace line to stderr: `  → <type>: <truncated-command>`
- Trace lines use dim styling via `display.Printer.StepTrace()`
- Multiline commands are collapsed to first line with `...`; long commands are truncated at 80 chars
- Environment variable references (`$VAR` and `${VAR}`) in trace values are expanded using the combined env from inputEnv (`PI_IN_*`), automation-level env, and step-level env via `expandTraceVars()`. Unknown variables (e.g. `$HOME`) pass through unchanged. This is display-only — execution is not affected.
- Installer steps are exempt — they have their own formatted status output
- A step with `silent: true` suppresses its trace line AND its stdout/stderr output
- Silent steps still execute — only their output is hidden; pipe data (`pipe: true`) still flows through
- `Executor.Loud` overrides all `silent: true` flags — when set, every step prints trace + output
- `--loud` flag on `pi run` and `pi setup` sets `Executor.Loud = true`
- `execStepSuppressed()` wraps `execStep()` with stdout/stderr redirected to `io.Discard` for silent steps
- When a silent step uses `pipe: true`, pipe capture still works (only non-pipe stdout is discarded)
- `pi info` shows `[silent]` annotation on steps with `silent: true`

### Parent shell execution (`parent_shell: true` on steps)
- Bash steps can declare `parent_shell: true` to run in the calling shell instead of as a subprocess
- `parent_shell` is only valid on bash steps — error on python, typescript, or run steps
- `parent_shell` cannot be combined with `pipe` — error at parse time
- When a parent_shell step executes, PI does **not** run it; instead it appends the command to `Executor.ParentEvalFile`
- `ParentEvalFile` is populated from the `PI_PARENT_EVAL_FILE` env var by `cli/run.go` and `cli/setup.go`
- If `ParentEvalFile` is empty and a parent_shell step is encountered, a warning is printed to stderr and the step is skipped (non-fatal): `⚠  parent_shell step skipped: not running inside a PI shell wrapper. Run 'pi shell' to install shell integration.`
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
- Pattern: create a temp file → run PI with `PI_PARENT_EVAL_FILE` set as a command prefix → source the file if non-empty → clean up → preserve exit code
- `PI_PARENT_EVAL_FILE` is set **inside** the subshell as a simple command prefix (`(cd "/path" && PI_PARENT_EVAL_FILE=... pi run ...)`), not outside — `VAR=value (subshell)` is a syntax error in both bash and zsh
- `pi-setup-<project>` uses the same pattern wrapping `pi setup`
- The temp file is always cleaned up, even if PI exits with an error
- The exit code from `pi run` is preserved via `local _pi_exit=$?` → `return $_pi_exit`

### Global `pi()` wrapper function
- `pi shell` installs a global `pi()` shell function in `~/.pi/shell/_pi-wrapper.sh`
- The wrapper wraps every `pi` invocation with the eval pattern so `parent_shell: true` works for direct `pi run` calls, not just shortcuts
- Uses `command pi` to call the real binary and avoid infinite recursion
- The wrapper file is created/updated during `shell.Install()` and removed during `shell.Uninstall()` when the last project is uninstalled
- `ListInstalled()` excludes the wrapper file from the project list
- The `_` prefix ensures the wrapper sorts before project files and is sourced first

### Step-level environment variables (`env:` on steps)
- Steps can declare an `env:` mapping of key-value pairs to inject into the step's execution environment
- `env:` vars are merged into the process environment after PI_IN_*/PI_INPUT_* vars and provisioned runtime PATH
- Step-level env vars do not leak between steps — each step gets a fresh copy of the process environment plus its own `env:` overlay
- Steps without `env:` inherit the parent process environment as before (backward compatible)
- Step-level env vars override parent env vars with the same name (last-writer-wins since they're appended)
- Works with all step types: bash, python, typescript
- `pi info` shows `[env: KEY1, KEY2]` annotations on steps that declare env vars
- Install phases (`install:` block) do not support step-level `env:` — they use only input env vars

### Automation-level environment variables (`env:` on automations)
- Automations can declare a top-level `env:` mapping; those variables apply to every step in that automation (both `steps:` and `install:` phase steps)
- Step-level `env:` overrides automation-level `env:` for the same key: both layers are merged into the env slice in order (automation env before step env), so duplicate keys resolve last-writer-wins when the process environment is built
- Automation-level env does not propagate to sub-automations invoked via `run:` — each nested `Run()` uses only that automation’s declared automation env (and inputs), not the caller’s
- Single-step shorthand: a top-level `env:` beside `bash:` / `python:` / `typescript:` / `run:` is automation-level env (not a per-step-only modifier); semantics match multi-step automations
- `pi info` prints an `Env:` line listing automation-level keys in sorted order when present
- `buildEnv()` in `helpers.go` takes three env layers: `inputEnv` (PI_IN_* + PI_INPUT_* + PI_ARGS), `automationEnv`, and `stepEnv` — merged in that order before subprocess execution

### Step working directory (`dir:` on steps)
- Steps can declare a `dir:` field to override the working directory for that step's execution
- When `dir:` is set, the path is resolved relative to the repo root; absolute paths are used as-is
- The resolved directory must exist at execution time — `resolveStepDir()` in `helpers.go` validates existence and confirms it's a directory
- Validation happens in `execStep()` before the runner is invoked; non-existent or non-directory paths produce a clear error
- Steps without `dir:` use the repo root as their working directory (backward compatible)
- `dir:` is per-step — each step independently resolves its own directory, no carry-over between steps
- `WorkDir` on `RunContext` carries the resolved directory to runners; `runStepCommand()` uses `cmd.Dir = ctx.WorkDir`
- Works with all step types: bash, python, typescript
- `dir:` is independent of other step fields: combinable with `env:`, `if:`, `silent:`, `pipe: true`
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
- Compatible with all other step fields: `env:`, `dir:`, `silent:`, `if:`, `pipe: true`
- When a step with `if:` evaluates to false, no timeout applies (step is skipped)
- When a step with `silent: true` times out, the timeout error still propagates
- `pi info` shows `[timeout: <value>]` annotations on steps that declare `timeout:`

### First-match blocks (`first:`)
- Steps can declare a `first:` field containing a list of sub-steps
- A step with `First != nil` is a first-match block; `Step.IsFirst()` returns true
- The executor evaluates each sub-step's `if:` condition in order and runs only the first one that matches
- A sub-step without `if:` always matches and acts as a fallback
- If no sub-step matches, the block is silently skipped
- Nested `first:` blocks are rejected at parse time
- Block-level fields: only `description:`, `if:`, and `pipe:` are valid; `env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`, `with:` are rejected with messages pointing to sub-steps
- Sub-steps support all normal step fields: `env:`, `dir:`, `timeout:`, `silent:`, `parent_shell:`
- `pipe: true` on a `first:` block correctly pipes the matched sub-step's stdout to the next step
- When a piped `first:` block has no matching sub-step, an empty pipe buffer is set (not stale data)
- Works in all step contexts: `steps:`, `install.run:`, `install.test:`, `install.verify:`
- `execFirstBlock()` in `executor.go` handles `first:` in the main step loop
- `execInstallFirstBlock()` in `install.go` handles `first:` in install phases
- `pi info` renders `first:` blocks with lettered sub-steps (a, b, c) and sub-step annotations
- `pi validate` traverses into `first:` blocks to check `run:` step references
- Silent/loud mode works on sub-steps: each sub-step's `silent:` flag is respected, `Loud` overrides it

### Conditional step execution (`if:` on steps)
- Steps can declare an `if:` field containing a boolean condition expression
- Before executing a step with `if:`, the executor extracts predicates, resolves them via the predicate resolver, and evaluates the expression
- If the expression evaluates to false, the step is silently skipped (no output, no error)
- Steps without `if:` always execute (backward compatible)
- Invalid `if:` expressions are caught at YAML load time, not at runtime
- `evaluateCondition()` on `Executor` uses `RuntimeEnv` field if set (for testing), otherwise `DefaultRuntimeEnv()`
- **Pipe passthrough on skip**: when a skipped step has `pipe: true`, any existing piped input from a prior step passes through to the next step unchanged. If the skipped step is the first in a pipe chain, `pipedInput` remains nil.

### Conditional automation execution (`if:` on automations)
- Automations can declare a top-level `if:` field containing a boolean condition expression
- Before executing any steps, `RunWithInputs()` evaluates the automation's `if:` condition
- If the condition evaluates to false, the automation is skipped: a message `[skipped] <name> (condition: <expr>)` is printed to stderr and `nil` is returned
- Automations without `if:` always execute (backward compatible)
- The condition check happens before `pushCall()`, so skipped automations don't consume call stack slots — this naturally prevents false circular-dependency errors
- `run:` steps calling a conditionally-skipped automation succeed without error (the parent step continues normally)
- Invalid `if:` expressions are caught at YAML load time, not at runtime

### Setup entry syntax (`setup:` in `pi.yaml`)
- Setup entries accept both bare strings and objects (like shortcuts)
- Bare strings (`- setup/install-go`) are shorthand for `run: <string>` with no modifiers
- Object form (`run:` + optional `if:` + optional `with:`) is required for entries that need modifiers
- Both forms can be mixed in the same list
- `SetupEntry.UnmarshalYAML()` handles the polymorphic parsing (scalar → bare, mapping → object)
- Bare string entries cannot have `if:` or `with:` — YAML syntax prevents it naturally

### Conditional setup entries (`if:` on setup)
- Setup entries in `pi.yaml` can declare an `if:` field containing a boolean condition expression (requires object form)
- Before running each setup entry, `pi setup` evaluates the `if:` condition using `conditions.NewEvaluator().ShouldSkip()`
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
- Runtime requirements map names to commands: `python` → `python3`, `node` → `node`, `rust` → `rustc`, `go` → `go`
- Version detection: runs `<cmd> --version`, captures stdout+stderr, extracts semver via regex `(\d+(?:\.\d+)+)`; falls back to `<cmd> version` (no `--`) when `--version` fails or produces no version string (needed for `go version` etc.)
- Handles all common version output formats: `Python 3.13.0`, `v20.11.0`, `jq-1.7.1`, `docker version 24.0.5`
- Version comparison: splits on `.`, compares numeric components pairwise; missing trailing components are treated as 0
- Install hints: built-in map of common tool names → install instructions (python, node, go, rust, docker, jq, kubectl, rustc, cargo, etc.)
- Formatted error output shows automation name, missing requirements with version info, and install hints
- Testability: `RuntimeEnv.ExecOutput` field allows mocking `--version` calls without real command execution
- Validation happens after input resolution but before step/install execution
- `InstallHintFor()` is the exported version of `installHint()` for use by `pi doctor`

### Step walker (`internal/automation/walker.go`)
- `WalkSteps(a, fn)` visits every step in an automation: regular steps, `first:` block sub-steps (expanded inline), and install phase steps (test → run → verify order)
- `WalkStepsUntil(a, fn)` is the early-stop variant — visitor returns `true` to stop walking
- `StepLocation` encodes where the step lives: `Phase` ("" or "test"/"run"/"verify"), `Index`, `FirstIndex` (-1 when not in a `first:` block), `IsScalar` (for synthetic steps from scalar install phases)
- `FormatPath(automationName)` on `StepLocation` produces human-readable location strings for error messages (e.g. `"my-auto: step[2]"`, `"my-auto: install.run step[0].first[1]"`, `"my-auto: install.test"`)
- Scalar install phases generate a synthetic `Step{Type: StepTypeBash, Value: phase.Scalar}` so consumers don't need separate handling
- Used by `internal/validate` for both run-step-ref and file-ref validation — eliminates duplicated traversal code
- Pure traversal — no I/O, no side effects; visitors capture results externally

### Validate command (`pi validate`)
- Statically validates all config and automation files without executing anything
- CLI layer (`cli/validate.go`) is a thin wrapper: resolves project → builds `validate.Context` → runs `validate.DefaultRunner()` → prints results
- Validation logic lives in `internal/validate/` with a registry-based `Check` interface for extensibility
- `DefaultRunner()` pre-registers 9 checks: shortcut-refs, setup-refs, run-step-refs, file-refs, shortcut-inputs, setup-inputs, run-step-inputs, circular-deps, conditions
- Adding a new check = implement `validate.Check` interface + register with `Runner.Register()`
- Validation layers: (1) pi.yaml schema via `config.Load()`, (2) automation discovery via `discoverAll()`, (3) pluggable validation checks via `validate.Runner`
- File reference checks: when a step value looks like a file path (detected by `executor.IsFilePath()` — ends in `.sh`/`.py`/`.ts`, no newlines, no spaces), the path is resolved relative to the automation YAML file's directory and checked for existence on disk. Built-in automations are skipped (they use inline scripts only). Covers steps, `first:` block sub-steps, and install phase steps (including scalar install phases).
- All step-walking checks use `automation.WalkSteps()` for traversal, with `StepLocation.FormatPath()` for error location formatting
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
- Scalar entries are parsed as runtimes; `python`, `node`, `go`, and `rust` are known runtimes — unknown names produce an error suggesting `command:` syntax
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
- Resolved values are injected as both `PI_IN_<NAME>` (canonical) and `PI_INPUT_<NAME>` (deprecated) environment variables (uppercased, hyphens become underscores)
- `InputEnvVars()` converts resolved inputs to `PI_IN_*` + `PI_INPUT_*` env vars in sorted key order; both prefixes are always set for backward compatibility
- `pi info` shows `→ $PI_IN_<NAME>` next to each input spec
- `run:` steps support `with:` to pass named inputs to the called automation
- `pi run --with key=value` is a repeatable flag parsed by `parseWithFlags()`
- `pi.yaml` shortcuts support `with:` mapping with `$1`, `$2` positional references
- Shell codegen detects `with:` on shortcuts and emits `--with key="$N"` flags instead of `"$@"` passthrough
- `pi list` shows an INPUTS column with required inputs as `name` and optional as `name?`

### Built-in automations (`pi:` prefix)
- Built-in automation YAML files live in `internal/builtins/embed_pi/` and are embedded into the binary at build time via `//go:embed`
- Go-backed builtins are registered programmatically in `builtins.goBackedBuiltins()` and merged into the discovery result
- `internal/builtins.Discover()` walks the embedded FS, parses YAML files with `automation.LoadFromBytes()`, merges Go-backed builtins, and returns a `*discovery.Result`
- `cli.discoverAll()` discovers local automations, then calls `result.MergeBuiltins()` to incorporate built-ins
- Precedence: local automations shadow built-ins with the same name. `Find("hello")` returns local if it exists, otherwise built-in. `Find("pi:hello")` always returns the built-in regardless of local shadowing
- `pi list` hides built-in automations by default; `--builtins` / `-b` flag includes them with `[built-in]` in the SOURCE column
- `run:` steps can reference `pi:hello` to explicitly call built-in automations
- `pi.yaml` setup entries can reference `pi:hello` for built-in setup automations
- Built-in automations use inline scripts only (no file-path steps) since they have no real filesystem directory
- All built-in YAML files use the new concise syntax: no `name:` fields, single-step shorthand where applicable, `first:` blocks in installer run phases, `PI_IN_*` input variables
- Docker automations (`docker/up`, `docker/down`, `docker/logs`) use single-step shorthand; detect `docker compose` (v2 plugin) first, falling back to `docker-compose` (v1 standalone); forward all CLI args via `"$@"`
- Installer automations (`install-homebrew`, `install-python`, `install-node`, `install-go`, `install-rust`, `install-uv`, `install-tsx`, `install-terraform`, `install-kubectl`, `install-helm`, `install-pnpm`, `install-bun`, `install-deno`, `install-aws-cli`) use the structured `install:` block:
  - Each defines `test:`, `run:`, and optional `verify:` and `version:` fields
  - PI manages all user-facing output — automations only provide commands
  - `test` exits 0 → `✓  <name>  already installed  (<version>)`; no `run` executed
  - `test` exits non-zero → `→  <name>  installing...` → `run` executes → `verify` (or re-run `test`) → `✓  <name>  installed  (<version>)` or `✗  <name>  failed`
  - `install-homebrew` has `if: os.macos` at the automation level (skipped on non-macOS)
  - `install-python`, `install-node`, `install-go`, `install-rust`, `install-terraform`, `install-kubectl`, and `install-helm` use `pi:version-satisfies` in their `test:` phase for semver-aware version checking via `outputs.last` interpolation
  - `install-python`, `install-node`, and `install-go` accept a `version` input; use `first:` blocks in the `run:` phase to try `mise` first, fall back to `brew`, with a clear error fallback
  - `install-rust` accepts a `version` input; uses `rustup` if available, otherwise installs via the official `rustup.rs` installer script
  - `install-uv` uses the official `astral.sh/uv/install.sh` script
  - `install-tsx` uses `npm install -g tsx`
  - `install-terraform`, `install-kubectl`, `install-helm` accept a required `version` input; use `first:` blocks prioritizing mise, then brew/script fallbacks
  - `install-pnpm`, `install-bun`, `install-deno` accept an optional `version` input (omit for latest); use `first:` blocks prioritizing mise, then platform-specific fallbacks
  - `install-aws-cli` has no version input; uses platform-specific installers (brew on macOS, official zip on Linux amd64/arm64)
- Utility automations (`uv/sync`, `npm/install`, `set-env`) use single-step shorthand:
  - `uv/sync` accepts optional `dir`, `groups`, `args` inputs; runs `uv sync` with computed flags
  - `npm/install` accepts optional `dir` input; prefers `npm ci` when `package-lock.json` exists, falls back to `npm install`
  - `set-env` accepts required `key` and `value` inputs, optional `comment`; idempotently writes `export KEY=VALUE` to `~/.zshrc` and `~/.bashrc`
- Dev tool automations (`cursor/install-extensions`, `git/install-hooks`) use single-step shorthand:
  - `cursor/install-extensions` accepts an `extensions` input (comma or newline-separated IDs), checks `cursor --list-extensions`, installs missing ones via `cursor --install-extension`
  - `git/install-hooks` accepts a `source` input (directory path relative to repo root), copies hook files to `.git/hooks/`, makes them executable; uses `cmp` for idempotency
- Go-backed builtins (no YAML file; registered programmatically):
  - `version-satisfies` — semver-aware version constraint checking; accepts `version` and `required` inputs; exits 0 if satisfied, 1 with error message if not; uses `internal/semver` package; intended for use in `install.test:` phases via `outputs.last` interpolation

### Go-backed builtins (`GoFunc` on `Automation`)
- Automations can have a `GoFunc func(inputs map[string]string) error` field instead of steps or install blocks
- `IsGoFunc()` returns true when `GoFunc` is non-nil
- `validate()` skips step/install checks for GoFunc automations
- `Executor.RunWithInputs()` handles GoFunc dispatch: resolves inputs, calls GoFunc, prints errors to stderr, returns `ExitError{Code: 1}` on failure
- GoFunc automations respect `if:` conditions (checked before GoFunc call)
- GoFunc automations are registered in `builtins.goBackedBuiltins()` and merged into the discovery result during `Discover()`
- `pi info` shows `Type: go-native` for GoFunc automations
- Adding a new Go-backed builtin: create a constructor function in `internal/builtins/`, add it to `goBackedBuiltins()` map

### Inter-step output capture (`outputs.last`)
- Every step's stdout is captured via `io.MultiWriter` — output goes to both the original writer and a capture buffer
- Captured output is trimmed and stored in `Executor.stepOutputs` (a `[]string` indexed by step position)
- `stepOutputs` is scoped per-automation: saved and restored via `defer` when entering `RunWithInputs()` and install phases
- Pipe steps: the pipe buffer content is also recorded as the step's output
- Silent steps: output is still captured (MultiWriter to io.Discard + capture buffer)
- Install phase steps: output capture added to `execInstallPhaseWithStderr()` so `outputs.last` works in `install.test:` pipelines

### `with:` value interpolation
- `InterpolateWith` callback on `RunContext` enables `RunStepRunner` to resolve special references in `with:` values before passing them to the called automation
- `interpolateWithCtx(with, inputEnv)` resolves all references in a with map
- Three interpolation sources:
  - `outputs.last` — trimmed stdout of the most recently executed step (via `lastOutput()`)
  - `outputs.<N>` — stdout of step N (0-indexed) from `stepOutputs` slice
  - `inputs.<name>` — current automation's input value (looked up in `PI_IN_*` env vars)
- Non-matching values pass through unchanged (literal strings work as before)
- The interpolation happens in `RunStepRunner.Run()` before calling `RunAutomation`

### Version constraint checking (`internal/semver`)
- Wraps `github.com/Masterminds/semver/v3` with PI-specific normalisation
- `Satisfies(version, constraint) error` — the public API; returns nil if satisfied, error with explanation if not
- Normalises incomplete versions: "22" → "22.0.0", "22.3" → "22.3.0"
- Strips `v` prefix from both version and constraint
- Bare version numbers are treated as caret constraints: "22" → "^22.0.0" (matches any 22.x.x)
- Supports all Masterminds/semver constraint syntax: `>=`, `<=`, `!=`, `>`, `<`, `~`, `^`, `=`, comma-separated ranges
- Operator-prefixed constraints have their version parts normalised (e.g. ">= 20" → ">= 20.0.0")

### Sandboxed runtime provisioning (`internal/runtimes`)
- Opt-in via `runtimes:` block in `pi.yaml` — `provision: never` (default), `ask`, or `auto`
- `manager: mise` (default, falls back to direct if mise not installed) or `manager: direct`
- Provisioned runtimes are installed into `~/.pi/runtimes/<name>/<version>/bin/`
- `python`, `node`, `go`, and `rust` are known runtimes — `command:` requirements are never provisioned
- Integration point: `Executor.Provisioner` field — when set, `ValidateRequirements()` calls `tryProvision()` for failed runtime requirements
- PATH scoping: `buildEnv()` prepends provisioned bin directories to PATH for all step executions via `prependPathInEnv()`
- `buildEnv()` handles input env vars, provisioned runtime PATH, automation-level env, and step-level env vars; automation and step env keys are each iterated in sorted order for deterministic behavior
- Mise backend: calls `mise install <runtime>@<version>`, then `mise where` to find the install path, symlinks binaries into the managed directory
- Direct backend: downloads from official CDN (nodejs.org for node, python-build-standalone for python), extracts tar.gz, places binaries
- `Provisioner.PromptFunc` controls interactive "ask" mode — nil means non-interactive (skip provisioning)
- Already-provisioned runtimes are detected by checking for the binary at the expected path — no re-download

### Styled terminal output (`internal/display`)
- `Printer` wraps an `io.Writer` with optional ANSI color codes
- Color is auto-detected: enabled only when the writer is a `*os.File` backed by a terminal and `NO_COLOR` is not set
- `NewForWriter(w)` accepts an arbitrary `io.Writer`: uses TTY-aware color detection when `w` is `*os.File`, disables color otherwise. Preferred over `New()` in CLI commands that receive `io.Writer` parameters.
- `NewWithColor(w, bool)` allows explicit control for testing
- Style methods: `Plain()`, `Dim()`, `Green()`, `Red()`, `Bold()` — all accept `fmt.Sprintf`-style format strings
- `InstallStatus(icon, name, status, version)` encapsulates the icon→style mapping: `✓`+already→dim, `✓`+installed→bold green, `✗`→bold red, `→`→plain
- `SetupHeader()` renders `==>` lines in dim style
- `StepTrace(stepType, value)` renders step trace lines (`  → bash: echo hello`) in dim style with multiline collapse and 80-char truncation
- TTY detection uses `golang.org/x/term.IsTerminal()` for cross-platform reliability
- Backward compatible: tests use `bytes.Buffer` (not `*os.File`) so color is always off in tests; identical text output
- Used by: `Executor.printInstallStatus()`, `Executor.RunWithInputs()` for step trace lines, `cli/setup.go` headers, `cli/doctor.go` health output

### Shell completion (`pi completion`)
- `pi completion <shell>` generates a completion script for bash, zsh, fish, or powershell using Cobra's built-in generators
- Dynamic completion for automation names: `automationCompleter()` returns a `ValidArgsFunction` that resolves the project, discovers automations, and returns names with descriptions
- Wired to `pi run` and `pi info` via `cmd.ValidArgsFunction` — `pi run <TAB>` and `pi info <TAB>` complete automation names
- Built-in automations are excluded from completion results (consistent with `pi list` default behavior)
- Completion functions silently return empty on any error — completion should never crash the shell
- When more than one arg is already provided (the automation name), no further completions are offered
- `pi shell` automatically generates a `_pi-completion.sh` file in `~/.pi/shell/` that sets up completion for both bash and zsh
- The completion file detects `$ZSH_VERSION` / `$BASH_VERSION` and eval's the appropriate `pi completion <shell>` output
- `_pi-completion.sh` is sourced via the existing `*.sh` glob in the source block, alongside the wrapper and shortcut files
- Cleanup: `Uninstall()` removes `_pi-completion.sh` when the last project is uninstalled; `ListInstalled()` excludes it from the project list

### Error philosophy
- Parse errors include file path and field name
- `Find()` not-found errors show "Did you mean?" suggestions for close matches (Levenshtein distance ≤ 3 or ≤ 30% of name length), then list all available automations
- `findBuiltin()` not-found errors suggest close built-in names with `pi:` prefix
- Collision errors mention both conflicting file paths
- Circular dependency errors show the full chain (e.g., `a → b → c → a`)
- Missing `pi.yaml` errors mention the start directory

### Error type checking convention
- All error type checking uses `errors.As()` instead of direct type assertions (`err.(*SomeType)`)
- This ensures error unwrapping works correctly if errors are wrapped with `fmt.Errorf("...: %w", err)` in the future
- Pattern: `var exitErr *ExitError; if errors.As(err, &exitErr) { ... }` — never `if exitErr, ok := err.(*ExitError); ok { ... }`
- Applies to production code and test code alike
- Key error types checked via `errors.As()`: `*executor.ExitError`, `*executor.ValidationError`, `*config.DuplicatePackageError`, `*config.DuplicateSetupEntryError`, `*config.ReplacedSetupEntryError`, `*exec.ExitError`

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

### Predicate resolution and evaluation (`internal/conditions/evaluator.go`)
- Converts predicate names (from `conditions.Predicates()`) into `map[string]bool` for `conditions.Eval()`
- `Evaluator` struct wraps repoRoot + RuntimeEnv, provides `ShouldSkip(expr)` — the one-call API for condition evaluation (parse → resolve → eval → negate); used by both `Executor.evaluateCondition()` and `cli/setup.go`
- `ResolvePredicates(names, repoRoot)` is the low-level public API; `ResolvePredicatesWithEnv(names, repoRoot, env)` accepts an injected `RuntimeEnv` for testing
- `RuntimeEnv` struct captures injectable `GOOS`, `GOARCH`, `Getenv()`, `LookPath()`, `Stat()`, `ExecOutput()` — no direct global reads
- Static predicates: `os.macos`, `os.linux`, `os.windows`, `os.arch.arm64`, `os.arch.amd64`, `shell.zsh`, `shell.bash`
- Dynamic predicates: `env.<NAME>` checks `Getenv(NAME) != ""`; `command.<name>` checks `LookPath(name) == nil`
- Function-call predicates: `file.exists("path")` checks path exists and is a file; `dir.exists("path")` checks path exists and is a directory — both resolve relative to `repoRoot`
- Unknown predicates produce an error listing all valid prefixes
- `ValidatePredicateName(name)` — static validation that a predicate name is structurally valid (matches a known exact name or known prefix pattern like `env.`, `command.`) without resolving runtime values; used by `pi validate` for static analysis
- `ValidateConditionExpr(expr)` — parses an `if:` expression and validates all extracted predicate names; combines `conditions.Predicates()` for syntax checking with `ValidatePredicateName()` for semantic checking
- The `executor` package re-exports all these symbols via thin delegation in `predicates.go` for backward compatibility

### CLI output
- `pi list` uses `text/tabwriter` for aligned columns (NAME, SOURCE, DESCRIPTION, INPUTS)
- SOURCE column shows `[workspace]` for local, `[built-in]` for builtins, alias or source string for packages
- `--all` / `-a` appends grouped sections per declared package with `──` separator headers
- `--builtins` / `-b` includes `pi:*` automations (hidden by default for cleaner output)
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

### Documentation website (`.github/workflows/docs.yml`)
- Triggers on push to `main` that touches `website/`
- Builds the Astro Starlight site in `website/`, then generates `llms-ctx.txt`
- Deploys to GitHub Pages
- Build pipeline: `npm ci` → `npm run build` → `./scripts/generate-llms-ctx.sh` → upload artifact → deploy
- AI/LLM support:
  - `public/llms.txt` — curated llmstxt.org-spec entry point linking all doc pages as `.md`
  - `.md` mirrors — every doc page available as clean Markdown (generated by `starlight-page-actions` plugin via `vite-plugin-static-copy`)
  - `llms-ctx.txt` — all docs concatenated as Markdown, generated by `website/scripts/generate-llms-ctx.sh`
  - Page action buttons — "Copy Markdown", "Open in ChatGPT", "Open in Claude" on every doc page (via `starlight-page-actions` plugin)

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
- `github.com/Masterminds/semver/v3` — Semver constraint parsing and matching (used by `pi:version-satisfies`)
- No CGO, no runtime dependencies

## Test Strategy

Unit tests per package using `testing` and `t.TempDir()` fixtures. Integration tests in `tests/integration/` build the `pi` binary and run it against `examples/` workspaces using `exec.Command`.

Total tests: 1743 (189 automation [177 base + 12 walker] + 142 builtins + 32 cache + 261 CLI [includes 6 completion, 12 init, 21 setup add, 20 info, 67 validate] + 30 conditions + 67 config + 42 display + 55 discovery [46 base + 9 suggest] + 305 executor [across 16 test files] + 4 project + 46 refparser + 16 runtimes + 57 semver + 40 shell + 52 validate + 277 integration [includes 8 completion, 8 init, 17 setup add, 8 info, 11 validate])

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
  add_test.go                     8 tests — pi add: file source, file with alias, idempotent, no version error, no args, creates packages block, appends to existing, invalid source
  init_test.go                    8 tests — pi init: creates project files, --yes flag, already initialized, .pi exists, next steps, non-interactive fallback, created project is valid, already-initialized next steps
  setup_add_test.go               17 tests — pi setup add: bare string, short-form resolution, if flag, idempotent duplicate, key=value args, no pi.yaml auto-init, local automation, multiple adds, no args, pi: prefix expansion, replace same run target, replace preserves position, preserves existing content, invoke before writing, invoke failure, only-add skips execution, not found without only-add
  basic_test.go                   9 tests — basic example: list, greet, greet with args, build/compile, deploy (run chaining), not-found, not-found did-you-mean, not-found no-suggestions, from subdirectory
  docker_test.go                  6 tests — docker-project example: list, up, down, logs, logs with args, build-and-up (ordering)
  pipe_test.go                    3 tests — pipe example: list, upper (bash pipe), count-lines (bash→python pipe)
  version_test.go                 3 tests — version: --version flag, version subcommand, flag and subcommand match
  inputs_test.go                  9 tests — inputs: positional args, both args, --with flags, defaults, missing required, unknown input, run step with with, info env var prefix, list INPUTS column
  info_test.go                    8 tests — info command: basic automation, with inputs (position display), not-found, no args, installer type and lifecycle, installer without version, installer with inputs, first: block sub-step details
  conditionals_test.go            20 tests — conditional execution: list, platform-info, skip-all, pipe passthrough, automation-level if, env/command/file predicates, complex booleans, combined automation+step if, info conditions
  builtins_test.go                27 tests — built-in automations: pi: prefix, local shadow, run step calls, docker builtins (list, info, run), installer builtins (list, info, inputs, conditions, idempotent), dev tool builtins (list, info, inputs), builtins hidden by default, not-found did-you-mean
  installer_schema_test.go        12 tests — installer schema: list, already-installed, fresh install, install-then-already, no-version, info type, info steps, conditional run, --silent, regular unaffected, failed install shows stderr, built-in installer
  requires_test.go                7 tests — requires validation: list, satisfied command, satisfied runtime, missing command, impossible version, no-requires, install hint
  doctor_test.go                  7 tests — pi doctor: all satisfied, missing, version mismatch, skips no-requires, detected version, install hint, healthy workspace
  runtime_provisioning_test.go    7 tests — runtime provisioning: list, no-requirements, already-installed, never-mode, config parsing, auto/ask modes
  step_env_integ_test.go          11 tests — step env and automation-level env: build, isolation, overrides, run: no bleed, install phases, list, info, shorthand, conditions
  step_visibility_integ_test.go   7 tests — step visibility: default trace, silent suppression, loud override, all-silent, all-silent loud, info silent annotation
  parent_shell_integ_test.go      7 tests — parent shell: list, eval file write, mixed steps, no eval file error, normal unaffected, info annotation, shell codegen
  step_dir_integ_test.go          6 tests — step dir: list, run in subdir, mixed dirs, dir with env, bad dir error, info annotation
  step_timeout_integ_test.go      5 tests — step timeout: list, fast completes, slow exceeds (exit 124), mixed timed/untimed, info annotation
  step_description_integ_test.go  5 tests — step description: list, run, info descriptions, info with annotations, info no-desc no-details
  validate_integ_test.go          17 tests — pi validate: valid project, invalid project, all errors reported, basic project, builtin refs, broken file references, valid file reference, inline scripts not flagged, installer scalar file refs (broken + valid), installer inline scripts not flagged, with: inputs valid, with: inputs invalid, with: no inputs on target, run step with: invalid input, builtin with: valid input, builtin with: invalid input
  first_block_integ_test.go       8 tests — first: block: list, pick-platform, no-match, with-pipe, mixed, info, validate, installer
  shorthand_integ_test.go         8 tests — single-step shorthand: list, run bash, run with env, run step delegation, run with input, info, info with modifiers, validate
  packages_test.go                11 tests — packages: list (SOURCE column), list --all (grouped sections), run local, run package automation, run via alias, run utils, info, validate, setup fetches packages, local shadows package
  completion_test.go              8 tests — completion: bash/zsh/fish output, no arg error, works without pi.yaml, shell install creates completion file, dynamic run/info completion, builtin exclusion
  on_demand_test.go               6 tests — on-demand fetch: file: ref never on-demand, GitHub ref undeclared shows advisory, declared package no advisory, cached package silent, file: missing path clear error, advisory to stderr
  polyglot_test.go                Polyglot runner tests (Python inline/file, TypeScript inline/file, multi-step pipe chains)
  shell_test.go                   Shell shortcut tests (install, uninstall, list, --repo, setup integration, --no-shell, conditional entries)
```

- Build `pi` binary once in `TestMain` (`main_test.go`)
- Run `pi list` and `pi run` against `examples/` workspaces using `exec.Command`
- Assert exit codes, output content, and step ordering
