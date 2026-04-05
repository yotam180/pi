# Automation Reference Parser

## Type
feature

## Status
done

## Priority
high

## Project
13-external-packages

## Description
Build the foundation layer that parses and classifies any automation reference string into one of three source types, before any I/O happens. Everything else in project 13 depends on this.

Reference formats to support:

| Format | Source type | Example |
|--------|-------------|---------|
| `path/to/automation` | Local `.pi/` | `docker/up`, `build/default` |
| `pi:name` | Built-in | `pi:install-go` |
| `org/repo@version/automation-path` | GitHub package | `yotam180/pi-common@v1.2/docker/up` |
| `file:~/path/automation` | Local folder | `file:~/my-automations/docker/up` |
| `alias/automation-path` | Alias (resolved via `packages:`) | `common/docker/up` |

Rules:
- A string containing `@` and at least two `/`-separated segments before `@` → GitHub ref
- A string starting with `file:` → file source
- A string matching a declared alias prefix → alias ref
- Everything else → local `.pi/` lookup (existing behavior)

The parser returns a typed struct (`LocalRef`, `BuiltinRef`, `GitHubRef`, `FileRef`, `AliasRef`) with all extracted fields (org, repo, version, path, etc.). It does no filesystem access or network calls — pure parsing.

## Acceptance Criteria
- [x] Each of the five reference formats is correctly parsed to the right struct type
- [x] `org/repo@version/path` extracts org, repo, version, and automation path correctly
- [x] `file:~/path/automation` extracts the filesystem path (with `~` expanded)
- [x] A ref with `@` but missing org/repo prefix is a parse error with a clear message
- [x] A ref starting with `file:` but with an empty path is a parse error
- [x] Alias prefix matching only triggers when the alias is declared in the loaded `packages:` config
- [x] Round-trip: parsing then re-serializing produces the canonical string form
- [x] Comprehensive unit tests for all formats and edge cases (missing `@`, extra `/`, etc.)

## Implementation Notes

### Package location
Created `internal/refparser/refparser.go` — a pure-logic package with no dependencies on other PI internal packages (except `os` for tilde expansion).

### Design decisions
- **Single struct, not interface**: Used a single `AutomationRef` struct with a `Type` field (enum) rather than an interface with separate types. This is simpler and avoids type assertions — callers just switch on `ref.Type`.
- **Precedence order**: `pi:` > `file:` > `@` (GitHub) > alias match > local. This order ensures prefixed formats are detected first, then @ detection (which would otherwise be ambiguous with email-like strings), then alias lookup (which needs the packages config), and finally the default local resolution.
- **No I/O**: The parser does zero filesystem or network access. Tilde expansion uses `os.UserHomeDir()` but that's a syscall, not I/O. This makes the parser fast and testable.
- **`FindWithAliases()`**: Added to `discovery.Result` alongside the existing `Find()`. `Find()` delegates to `FindWithAliases(nil)` for backward compatibility. When packages support is added, callers with alias config will use `FindWithAliases()`.
- **Clear error messages for unsupported types**: GitHub, file, and alias refs return explicit "not yet supported" errors rather than "not found" — this guides users toward understanding the roadmap.

### Integration
Updated `discovery.Find()` to use `refparser.Parse()` internally. The existing behavior for local and builtin refs is preserved exactly. All 749+ existing tests pass without modification.

### Test coverage
46 test cases in `refparser_test.go` covering:
- Local refs (simple, nested, single, trailing/leading slash, mixed case)
- Builtin refs (simple, nested, deep)
- Builtin errors (empty name)
- GitHub refs (full, no path, main branch, deep path, semver)
- GitHub errors (missing org, missing repo, missing version, no org prefix, repo with slash)
- File refs (tilde, absolute, relative)
- File errors (empty path)
- Alias refs (with path, single segment, different alias)
- Alias non-match falls to local
- Empty ref error
- Whitespace handling
- String round-trip for all types
- Precedence rules
- Edge cases (@-at-position-0)

## Subtasks
- [x] Define `AutomationRef` interface and concrete types
- [x] Implement parser with format detection
- [x] Write unit tests
- [x] Integrate into existing automation resolver (replace raw string lookup)

## Blocked By
