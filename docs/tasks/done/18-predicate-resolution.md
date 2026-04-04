# Predicate Resolution

## Type
feature

## Status
done

## Priority
high

## Project
04-conditional-step-execution

## Description
Implement predicate resolution in the executor — the code that converts predicate names from `if:` expressions into actual boolean values by checking the runtime environment. This builds the `map[string]bool` that gets passed to `conditions.Eval()`.

Predicate vocabulary:

| Predicate | Resolution |
|---|---|
| `os.macos` | `runtime.GOOS == "darwin"` |
| `os.linux` | `runtime.GOOS == "linux"` |
| `os.windows` | `runtime.GOOS == "windows"` |
| `os.arch.arm64` | `runtime.GOARCH == "arm64"` |
| `os.arch.amd64` | `runtime.GOARCH == "amd64"` |
| `env.<NAME>` | `os.Getenv("NAME") != ""` |
| `command.<name>` | `exec.LookPath(name) != nil` |
| `file.exists("path")` | `os.Stat(filepath.Join(repoRoot, path))` is a file |
| `dir.exists("path")` | `os.Stat(filepath.Join(repoRoot, path))` is a directory |
| `shell.zsh` | `os.Getenv("SHELL")` ends with `/zsh` |
| `shell.bash` | `os.Getenv("SHELL")` ends with `/bash` |

The resolver should be in a new file (e.g. `internal/executor/predicates.go`) or a dedicated `internal/predicates` package. It must be testable — inject the runtime values rather than reading globals directly where possible.

## Acceptance Criteria
- [x] All predicates in the table above are implemented
- [x] `file.exists()` and `dir.exists()` resolve paths relative to the project root
- [x] Unknown predicate patterns produce a clear error listing valid predicate prefixes
- [x] `ResolvePredicates(predicateNames []string, repoRoot string) (map[string]bool, error)` is the exported interface
- [x] Unit tests cover each predicate type, including edge cases (missing env var, command not in PATH, nonexistent file)

## Implementation Notes

### Approach
- Created `internal/executor/predicates.go` — kept in the executor package since it will be called directly from the executor when wiring `if:` conditions in task #19.
- Testability achieved via `RuntimeEnv` struct with injectable `GOOS`, `GOARCH`, `Getenv`, `LookPath`, and `Stat` — no global state read directly.
- Public API: `ResolvePredicates(predicateNames, repoRoot)` for production, `ResolvePredicatesWithEnv(predicateNames, repoRoot, env)` for tests.
- `DefaultRuntimeEnv()` returns real OS-backed `RuntimeEnv`.

### Predicate routing
- Static predicates (`os.*`, `shell.*`) resolved via a `switch` statement.
- Dynamic predicates (`env.*`, `command.*`) resolved via prefix matching with validation on empty suffixes.
- Function-call predicates (`file.exists("...")`, `dir.exists("...")`) detected via prefix/suffix matching on the key format produced by `conditions.Predicates()`.

### Edge cases handled
- `file.exists("path")` returns false when path is a directory (and vice versa for `dir.exists`).
- Empty env var name (`env.`) and empty command name (`command.`) produce clear errors.
- Unknown predicates produce an error listing all valid predicate prefixes.

### Test coverage
- 14 test functions covering all predicate types with subtests: OS (9 cases), arch (4 cases), shell (8 cases), env (4 cases), command (3 cases), file.exists (3 cases), dir.exists (3 cases), unknown predicate error, multiple predicates, empty list, DefaultRuntimeEnv smoke test, integration test with real `ResolvePredicates`.

## Subtasks
- [x] Create `predicates.go` (in executor or new package)
- [x] Implement OS predicates (static based on `runtime.GOOS`/`runtime.GOARCH`)
- [x] Implement `env.*` predicates
- [x] Implement `command.*` predicates
- [x] Implement `file.exists()` and `dir.exists()` predicates
- [x] Implement `shell.*` predicates
- [x] Write unit tests

## Blocked By
17-condition-expression-parser
