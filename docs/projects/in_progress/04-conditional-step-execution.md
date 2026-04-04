# Conditional Step Execution (`if:`)

## Status
in_progress

## Priority
medium

## Description
Add a first-class `if:` field to steps (and automation-level declarations) that evaluates a boolean expression against a small set of built-in predicates. Steps whose `if:` evaluates to false are skipped silently. This makes automations portable across OSes and environments without bash conditionals inside every step.

## Goals
- A step can be skipped on the wrong OS, missing command, or absent env var — declared directly in YAML
- The expression syntax is immediately readable by any developer without documentation
- No external runtime (Python, Node, etc.) is required to evaluate conditions
- The predicate set covers the real cases PI users actually hit: OS, env vars, command availability, filesystem checks
- `if:` works identically on steps, `run:` steps in automations, and `pi.yaml` setup entries

## Background & Context

Built-in automations like `pi:install-homebrew` and `pi:install-python` need OS routing to be useful — Homebrew only makes sense on macOS, `apt-get` only on Linux. Without `if:`, the only option is bash `if`-blocks inside steps, which hides the intent from the YAML and makes `pi list` useless for understanding what an automation does.

The original design discussion considered three approaches:

1. **`os:` filter on steps** — simple but OS-only, creates a second mechanism for every other condition
2. **Predefined dot-notation predicates only** — `os.macos`, `command.docker` — no combining, limited but zero-ambiguity
3. **Expression language with `and` / `or` / `not`** — covers all real cases in one mechanism

The chosen design is option 3 with a deliberately small predicate set. The expression syntax uses Python-style keywords (`and`, `or`, `not`) which are already familiar and don't require any operator knowledge. Evaluation is implemented as a custom Go recursive-descent parser over a fixed predicate vocabulary — no external dependencies, no runtime requirements, consistent with the single-binary philosophy.

Using Python or TypeScript as the expression evaluator was considered and rejected: it would require those runtimes to be available before any setup step runs, defeating the bootstrapping purpose of `pi:install-python`.

## Scope

### In scope

**`if:` field on steps**
```yaml
steps:
  - bash: brew install python@3.13
    if: os.macos

  - bash: apt-get install -y python3.13
    if: os.linux

  - bash: winget install python3
    if: os.windows
```

**`if:` field on `run:` steps and `pi.yaml` setup entries**
```yaml
# in pi.yaml
setup:
  - run: setup/install-brew
    if: os.macos
  - run: setup/install-uv
    if: not command.uv

# in an automation file
steps:
  - run: docker/up
    if: command.docker
```

**`if:` at the automation level** — automation is skipped (not errored) if false
```yaml
name: install-homebrew
if: os.macos
steps:
  - bash: ...
```

**Boolean operators:** `and`, `or`, `not`, parentheses for grouping
```yaml
if: os.macos and command.brew
if: not env.CI
if: os.macos or os.linux
if: (os.macos or os.linux) and command.docker
if: env.DOCKER_HOST or command.docker
if: os.macos and not command.brew
```

**Predicate vocabulary:**

| Predicate | Evaluates to true when |
|---|---|
| `os.macos` | running on macOS |
| `os.linux` | running on Linux |
| `os.windows` | running on Windows |
| `os.arch.arm64` | CPU architecture is ARM64 (e.g. Apple Silicon) |
| `os.arch.amd64` | CPU architecture is x86-64 |
| `env.<NAME>` | environment variable `NAME` is set and non-empty |
| `command.<name>` | `<name>` resolves in PATH (like `which <name>`) |
| `file.exists("<path>")` | file exists at `<path>` relative to project root |
| `dir.exists("<path>")` | directory exists at `<path>` relative to project root |
| `shell.zsh` | current shell is zsh |
| `shell.bash` | current shell is bash |

**Error handling:**
- Unknown predicate → hard error with the predicate name and a list of valid ones
- Malformed expression → hard error with the expression string and position of the problem
- A step with `if: false` result → silently skipped, nothing printed
- An automation with `if: false` at the automation level → skip with a brief note (`[skipped] install-homebrew (condition: os.macos)`)

**`pi list` integration:** Show the `if:` condition alongside each step when printing automation details

### Out of scope
- Equality checks (`env.FOO == "bar"`) — use bash for value comparisons
- Arithmetic or string operations
- Custom user-defined predicates
- `else:` / `elif:` — write two steps with complementary conditions instead
- Parallel step execution gated on conditions

## Success Criteria
- [ ] All predicates in the vocabulary table are implemented and tested
- [ ] `and`, `or`, `not`, and parentheses work correctly in all combinations
- [ ] Unknown predicates and malformed expressions produce clear, actionable errors
- [ ] `if:` works on: steps, `run:` steps, `pi.yaml` setup entries, and automation-level declarations
- [ ] Skipped steps are silent; skipped automations print a one-line note
- [ ] `pi list` shows `if:` conditions in step details
- [ ] `go test ./...` passes; unit tests cover the parser and all predicates
- [ ] At least one built-in automation (`pi:install-homebrew` or `pi:install-python`) uses `if:` as a real-world integration test

## Notes
- Implement the evaluator in a dedicated `internal/conditions` package so it has no dependencies on the executor or automation packages — it receives a `map[string]bool` of resolved predicates and an expression string, returns bool + error.
- Predicate resolution (actually calling `which`, checking `GOOS`, reading env) should happen in the executor, not the conditions package. This keeps the conditions package pure and unit-testable with any input map.
- `file.exists()` and `dir.exists()` take a string argument — the only predicate form that isn't a bare dotted identifier. The parser needs to handle this one case of function-call syntax.
- `command.<name>` uses PATH lookup at evaluation time, not at parse time — the command might be installed by an earlier step.
- `os.macos` / `os.linux` / `os.windows` map directly to `runtime.GOOS` values (`darwin`, `linux`, `windows`).
