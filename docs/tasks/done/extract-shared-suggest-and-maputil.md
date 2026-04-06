# Extract Shared suggest and maputil Packages

## Type
improvement

## Status
done

## Priority
medium

## Project
standalone

## Description

The codebase has duplicated utility code that should be consolidated into shared packages:

1. **Levenshtein distance** — identical implementations in `internal/discovery/suggest.go` and `internal/validate/unknown_fields.go`. Both compute Wagner–Fischer edit distance with the same algorithm.

2. **Suggest functions** — `suggestNames()` in discovery and `suggestField()` in validate follow the same pattern: compute distances, filter by threshold, sort by distance then alphabetically. They differ only in return type (slice vs single string) and threshold calculation.

3. **`sortedKeys()`** — four independent implementations across `shell/shell.go`, `config/writer.go`, `validate/unknown_fields.go`, and `cli/list.go`. Each takes a different map type. With Go generics (available since 1.18, project uses 1.26), these can be unified into a single generic function.

### Plan

- Create `internal/suggest` package: exports `Levenshtein()`, `Best()` (single closest match), `TopN()` (top N matches)
- Create `internal/maputil` package: exports `SortedKeys[K cmp.Ordered, V any](m map[K]V) []K`
- Migrate all callers to use the shared packages
- Remove duplicated code
- Maintain 100% test coverage on the new packages

## Acceptance Criteria
- [x] No duplicate `levenshtein()` implementations remain
- [x] No duplicate `sortedKeys()` implementations remain
- [x] `internal/suggest` package exists with tests and 100% coverage
- [x] `sortedKeys()` replaced with `slices.Sorted(maps.Keys())` (no maputil package needed)
- [x] All existing tests pass unchanged
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach
- `internal/suggest` will contain the core Levenshtein algorithm and two suggest helpers:
  - `Best(query string, candidates []string, maxDist int) string` — returns single best match (used by validate)
  - `TopN(query string, candidates []string, maxDist, maxResults int) []string` — returns top N matches (used by discovery)
- `internal/maputil` will use Go generics with `cmp.Ordered` constraint for the key type
- The `min3` helper will be unexported in the suggest package
- Tests will be ported from both source packages

### Decisions
- Package name `suggest` chosen over `levenshtein` or `strdist` because the primary use case is suggesting alternatives, not raw distance computation
- Instead of creating a `maputil` package, used stdlib `slices.Sorted(maps.Keys(m))` directly at call sites — cleaner and zero new code needed
- The `min3` helper was replaced with Go's built-in `min()` variadic function (available since Go 1.21)
- Threshold computation stays in callers (discovery uses 30% of query length min 3; validate uses 50% of field length min 2) — the shared package takes maxDist as a parameter

## Subtasks
- [x] Create `internal/suggest` package with Levenshtein + suggest functions
- [x] Replace `sortedKeys()` with `slices.Sorted(maps.Keys())` at all 4 call sites
- [x] Migrate `discovery/suggest.go` to use shared package
- [x] Migrate `validate/unknown_fields.go` to use shared package
- [x] Migrate `validate/unknown_pi_yaml.go` to use renamed `suggestFieldName()`
- [x] Update tests in discovery and validate packages
- [x] Run tests and verify — all pass, coverage maintained

## Blocked By
