# Predicate Resolution

## Type
feature

## Status
todo

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
- [ ] All predicates in the table above are implemented
- [ ] `file.exists()` and `dir.exists()` resolve paths relative to the project root
- [ ] Unknown predicate patterns produce a clear error listing valid predicate prefixes
- [ ] `ResolvePredicates(predicateNames []string, repoRoot string) (map[string]bool, error)` is the exported interface
- [ ] Unit tests cover each predicate type, including edge cases (missing env var, command not in PATH, nonexistent file)

## Implementation Notes

## Subtasks
- [ ] Create `predicates.go` (in executor or new package)
- [ ] Implement OS predicates (static based on `runtime.GOOS`/`runtime.GOARCH`)
- [ ] Implement `env.*` predicates
- [ ] Implement `command.*` predicates
- [ ] Implement `file.exists()` and `dir.exists()` predicates
- [ ] Implement `shell.*` predicates
- [ ] Write unit tests

## Blocked By
17-condition-expression-parser
