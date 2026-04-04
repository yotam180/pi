# Condition Expression Parser

## Type
feature

## Status
done

## Priority
high

## Project
04-conditional-step-execution

## Description
Implement the `internal/conditions` package: a recursive-descent parser and evaluator for boolean expressions used in `if:` fields. The package is pure logic — it receives a predicate-resolution map (`map[string]bool`) and an expression string, and returns `(bool, error)`. No dependencies on executor, automation, or any other PI package.

The expression language supports:
- Bare dotted identifiers as predicates: `os.macos`, `command.docker`, `env.CI`, `shell.zsh`
- Function-call syntax for path predicates: `file.exists("path")`, `dir.exists("path")`
- Boolean operators: `and`, `or`, `not`
- Parentheses for grouping
- Whitespace is insignificant

Examples:
```
os.macos
not env.CI
os.macos and command.brew
(os.macos or os.linux) and command.docker
os.macos and not command.brew
env.DOCKER_HOST or command.docker
file.exists(".env") and not env.CI
```

## Acceptance Criteria
- [x] `conditions.Eval(expr string, predicates map[string]bool) (bool, error)` works for all expression forms
- [x] Parser handles: dotted identifiers, `and`, `or`, `not`, parentheses, `file.exists("...")`, `dir.exists("...")`
- [x] Unknown predicates (not in the provided map) produce a clear error with the predicate name
- [x] Malformed expressions produce a clear error with position information
- [x] Empty expression returns `true` (no condition = always run)
- [x] `conditions.Predicates(expr string) ([]string, error)` extracts all predicate names from an expression (for pre-resolving predicates before evaluation)
- [x] Package has zero dependencies on other PI internal packages
- [x] Comprehensive unit tests: all operators, nesting, error cases, edge cases (empty, whitespace-only, nested parentheses, consecutive `not not`)

## Implementation Notes

### Architecture
Single file `internal/conditions/conditions.go` (~300 lines) containing:
- **Lexer**: Tokenizes input into IDENT, AND, OR, NOT, LPAREN, RPAREN, STRING, EOF tokens. Tracks byte positions for error reporting.
- **AST**: Four node types — `identNode` (bare predicates), `funcCallNode` (function-call predicates like `file.exists("...")`), `notNode`, `binaryNode` (and/or).
- **Parser**: Recursive descent following the grammar: `expr → orExpr → andExpr → notExpr → primary`. `and` binds tighter than `or`, matching boolean convention.
- **Evaluator**: Walks the AST. Bare identifiers lookup directly in the predicate map. Function calls are keyed as `name("arg")` in the map.
- **Predicate extractor**: Walks the AST collecting unique predicate names in declaration order.

### Design Decisions
- Function-call predicates (e.g., `file.exists(".env")`) are stored in the predicate map with the canonical key format `file.exists(".env")`. This keeps the evaluator generic — it doesn't need to know what `file.exists` means.
- The lexer rejects numeric characters as ident-starters, so `file.exists(123)` is a lexer error rather than a parser error. This is fine — the error message is clear.
- Deduplication in `Predicates()` preserves first-occurrence order.

### Test Coverage
31 unit tests covering:
- Empty/whitespace expressions
- Bare identifiers, dotted identifiers
- All operators (and, or, not), double not
- Operator precedence (and > or)
- Parentheses, nested parentheses
- Function-call syntax
- Complex combinations
- Multiple chained and/or
- Unknown predicate errors
- Malformed expression errors (9 cases)
- Unterminated string
- Predicate extraction (simple, complex, dedup, function calls)
- Lexer edge cases (empty, positions)
- All spec examples from the task description

## Subtasks
- [x] Define the token types (IDENT, AND, OR, NOT, LPAREN, RPAREN, STRING, EOF)
- [x] Implement lexer (tokenizer) that produces a token stream from the expression string
- [x] Implement recursive-descent parser: `expr → orExpr`, `orExpr → andExpr ("or" andExpr)*`, `andExpr → notExpr ("and" notExpr)*`, `notExpr → "not" notExpr | primary`, `primary → IDENT | IDENT "(" STRING ")" | "(" expr ")"`
- [x] Implement `Eval()` that walks the AST and evaluates against the predicate map
- [x] Implement `Predicates()` that walks the AST and collects predicate names
- [x] Write unit tests for all cases

## Blocked By
