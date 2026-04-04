# Condition Expression Parser

## Type
feature

## Status
todo

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
- [ ] `conditions.Eval(expr string, predicates map[string]bool) (bool, error)` works for all expression forms
- [ ] Parser handles: dotted identifiers, `and`, `or`, `not`, parentheses, `file.exists("...")`, `dir.exists("...")`
- [ ] Unknown predicates (not in the provided map) produce a clear error with the predicate name
- [ ] Malformed expressions produce a clear error with position information
- [ ] Empty expression returns `true` (no condition = always run)
- [ ] `conditions.Predicates(expr string) ([]string, error)` extracts all predicate names from an expression (for pre-resolving predicates before evaluation)
- [ ] Package has zero dependencies on other PI internal packages
- [ ] Comprehensive unit tests: all operators, nesting, error cases, edge cases (empty, whitespace-only, nested parentheses, consecutive `not not`)

## Implementation Notes

## Subtasks
- [ ] Define the token types (IDENT, AND, OR, NOT, LPAREN, RPAREN, STRING, EOF)
- [ ] Implement lexer (tokenizer) that produces a token stream from the expression string
- [ ] Implement recursive-descent parser: `expr → orExpr`, `orExpr → andExpr ("or" andExpr)*`, `andExpr → notExpr ("and" notExpr)*`, `notExpr → "not" notExpr | primary`, `primary → IDENT | IDENT "(" STRING ")" | "(" expr ")"`
- [ ] Implement `Eval()` that walks the AST and evaluates against the predicate map
- [ ] Implement `Predicates()` that walks the AST and collects predicate names
- [ ] Write unit tests for all cases

## Blocked By
