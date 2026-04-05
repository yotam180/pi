# Add "Did you mean?" suggestions for mistyped automation names

## Type
improvement

## Status
in_progress

## Priority
high

## Project
standalone

## Description

When a developer types `pi run dokcer/up` (a typo), PI currently prints "not found" and lists all available automations. This is unhelpful — the developer has to scan the full list to find the right name.

Per philosophy principle 2 ("Correct the action, not the developer"), PI should suggest close matches. For example:

```
automation "dokcer/up" not found

Did you mean?
  docker/up       Docker Compose up

Available automations:
  build           ...
  docker/up       ...
  ...
```

This applies to `findLocal()` in `internal/discovery/discovery.go`, which is the resolution path for local and package automation names.

## Acceptance Criteria
- [x] Levenshtein distance function implemented and tested
- [x] `findLocal()` includes "Did you mean?" suggestions for close matches (edit distance ≤ 3)
- [x] Suggestions are sorted by edit distance (closest first), max 3 shown
- [x] Unit tests cover: exact match (no suggestions), close typo, transposition, prefix match, no close matches
- [x] Integration test covers the error output format
- [x] `go build ./...` and `go test ./...` pass

## Implementation Notes

### Approach
- Implement a minimal Levenshtein distance function directly in the `discovery` package (no external dependency, no new package — it's ~15 lines and only used here)
- Threshold: edit distance ≤ 3 or ≤ 30% of the name length (whichever is more permissive), to handle both short and long names
- Show max 3 suggestions, sorted by distance, then alphabetically for ties
- The "Did you mean?" section appears before the "Available automations:" list
- Also applied to `findBuiltin()` for `pi:` references

### Tech Decisions
- Kept the distance function in `discovery` package rather than creating a new `internal/stringutil` package — YAGNI, it's a single function with a single call site
- Used dynamic programming Levenshtein (not Damerau-Levenshtein) — simpler and sufficient for automation name typos
- Threshold formula: `maxDist = max(3, len(name)*30/100)` — allows proportionally more edits for longer names while keeping a minimum of 3 for short names

## Subtasks
- [x] Implement `levenshtein()` function in discovery package
- [x] Implement `suggestNames()` function
- [x] Update `findLocal()` error message
- [x] Add unit tests
- [x] Add integration test
- [x] Update docs

## Blocked By
