# Static Circular Dependency Detection in pi validate

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

`pi validate` performs extensive static analysis of a PI project — checking automation references, file paths, and input mismatches — but it does not check for circular `run:` dependencies. Currently, circular deps are only caught at runtime by `executor.go`'s `pushCall()` method, which means developers only discover them when actually running the automation.

Since `pi validate` already has full access to the automation graph via `discovery.Result`, it should statically detect circular dependency cycles and report them as validation errors.

**Examples of what should be caught:**

1. Direct cycle: A → B → A
2. Indirect cycle: A → B → C → A
3. Self-referencing: A → A
4. Cycles within `first:` blocks (any sub-step's `run:` target could create a cycle)

**What should NOT be flagged:**
- Diamond dependencies (A → B, A → C, B → D, C → D) are fine — only cycles are errors
- Automations referenced via `run:` from `first:` blocks where the other branch doesn't create a cycle

**Implementation approach:**
Build a directed graph of automation → automation edges from all `run:` steps (including inside `first:` blocks). Detect cycles using standard DFS cycle detection. Report each cycle as a validation error with the full chain (e.g., "circular dependency: A → B → C → A").

## Acceptance Criteria
- [x] `pi validate` detects direct circular dependencies (A → B → A)
- [x] `pi validate` detects indirect circular dependencies (A → B → C → A)
- [x] `pi validate` detects self-referencing automations (A → A)
- [x] `pi validate` detects cycles through `first:` block `run:` steps
- [x] Diamond dependencies are NOT flagged
- [x] Each cycle is reported with the full chain in the error message
- [x] Builtin automations (pi:*) are included in cycle detection where referenced
- [x] All existing tests pass unchanged
- [x] New tests cover all cycle patterns above
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach
Added three functions to `cli/validate.go`:

1. **`buildRunGraph(disc)`** — Walks all discovered automations using `automation.WalkSteps`, collects `run:` step targets by resolving them via `disc.Find()`. Broken refs are silently skipped (already caught by `validateRunStepRefs`). Returns `map[string][]string` (adjacency list).

2. **`detectCycles(graph)`** — Standard DFS with three-color marking (white/gray/black). When a gray node is revisited, the cycle is extracted from the current path. Uses `normalizeCycleKey()` to deduplicate cycles found from different starting nodes (e.g., A→B→A found when starting at A vs starting at B).

3. **`normalizeCycleKey(cycle)`** — Rotates the cycle ring to start from the lexicographically smallest node, producing a canonical key for deduplication.

### Design decisions
- Graph includes all automations (local, builtin, package) since cycles can span across sources
- Broken `run:` references are skipped silently — they're already reported by `validateRunStepRefs`, no need to double-report
- Self-references (A→A) produce `["A", "A"]` which formats as `A → A`
- Multiple independent cycles are all reported (not just the first one)
- Error format: `circular dependency: A → B → C → A` — consistent with the runtime error format in `executor.go`

### Test coverage
18 new tests total:
- 8 integration tests: direct cycle, indirect cycle, self-reference, first: block cycle, diamond (OK), linear chain (OK), chain format verification, circular dep combined with other errors
- 10 unit tests: `detectCycles` (7 cases covering no-cycles, direct, self-loop, three-node, diamond, multiple, disconnected), `normalizeCycleKey` (2 cases), `buildRunGraph` (1 case)

## Subtasks
- [x] Add `validateCircularDeps` function in `cli/validate.go`
- [x] Build directed graph from `run:` steps using `automation.WalkSteps`
- [x] Implement DFS cycle detection
- [x] Format cycle errors with full chain
- [x] Wire into `validateProject()`
- [x] Write tests for all cycle patterns
- [x] Write tests for valid (non-cycle) patterns
- [x] Update `pi validate` help text
- [x] Update architecture.md
- [x] Run full test suite

## Blocked By
