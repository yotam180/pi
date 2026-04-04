# Implement PI Features from fzf Gap Analysis

## Type
feature

## Status
todo

## Priority
medium

## Project
07-fzf-adoption-test

## Description
Based on the findings from task 34 (clone-and-examine-fzf), implement any missing PI features or built-in automations needed to support fzf's developer workflows.

Task 34 found that **PI can model 100% of fzf's workflows today** using bash steps. The gaps are quality-of-life improvements, not blockers. The two most impactful gaps are:

### Gap 1: `env:` field on steps (medium priority)
fzf's Makefile uses environment variable overrides like `TAGS=pprof`, `SHELL=/bin/sh GOOS=`, etc. PI steps can't declare environment variables — the workaround is inlining them in bash (`TAGS=pprof go build ...`). An `env:` field on steps would be cleaner.

### Gap 2: `pi:install-go` built-in (low priority)
PI has `install-python` and `install-node` but not Go. Since PI itself is a Go tool and targets Go developers, a built-in `install-go` would be useful. Can be modeled as a local `.pi/` automation for now.

### Non-gaps
- No `pi:install-ruby` built-in — Ruby is niche; local automation is fine
- No `pi:install-shfmt` built-in — too niche for a built-in
- No matrix/cross-compile support — goreleaser handles this; not PI's job
- No `make` step type — bash works fine

### Recommendation
Since none of the gaps are blocking task 36 (Transform fzf to Use PI), this task can be done in parallel or after task 36. The `env:` feature is the only one worth implementing before task 36 for a cleaner result.

## Acceptance Criteria
- [ ] Decide on which gaps to implement vs skip
- [ ] If implementing `env:` on steps: parse, validate, inject env vars, test
- [ ] If implementing `pi:install-go`: add built-in automation YAML, tests
- [ ] `go test ./...` passes after all changes
- [ ] Documentation updated for any new features

## Implementation Notes

### Findings from task 34
- PI can model all 14 fzf Makefile targets as automations
- All gaps have bash workarounds
- The `env:` gap is the most impactful for clean automation YAML

## Subtasks
- [ ] Implement `env:` field on steps (if decided)
- [ ] Implement `pi:install-go` built-in (if decided)
- [ ] Update architecture.md and README.md

## Blocked By
34-clone-and-examine-fzf (completed)
