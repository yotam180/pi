# Implement PI Features from fzf Gap Analysis

## Type
feature

## Status
done

## Priority
medium

## Project
07-fzf-adoption-test

## Description
Based on the findings from task 34 (clone-and-examine-fzf), implement any missing PI features or built-in automations needed to support fzf's developer workflows.

Task 34 found that PI can model 100% of fzf's workflows today using bash steps. Two quality-of-life improvements were identified and implemented:

1. **`env:` field on steps** — allows steps to declare environment variables without inlining them in bash commands
2. **`pi:install-go` built-in** — installer automation for Go via mise/brew

## Acceptance Criteria
- [x] Decide on which gaps to implement vs skip
- [x] Implement `env:` on steps: parse, validate, inject env vars, test
- [x] Implement `pi:install-go`: add built-in automation YAML, tests
- [x] `go test ./...` passes after all changes
- [x] Documentation updated for any new features

## Implementation Notes

### `env:` on steps
- Added `Env map[string]string` to `Step` struct
- Added `env` field to `stepRaw` YAML parsing
- Updated `buildEnv()` to accept step-level env vars (3rd source after input vars and runtime paths)
- Step-level env vars are scoped per step — they don't leak between steps
- Works with all step types: bash, python, typescript
- `pi info` shows `[env: KEY1, KEY2]` annotations alongside `[if: ...]` annotations
- 3 parsing tests + 8 executor tests + 5 integration tests added

### `pi:install-go`
- Added `install-go.yaml` to `internal/builtins/embed_pi/`
- Pattern matches existing `install-python` and `install-node`
- Accepts `version` input (e.g. "1.23"), checks major.minor match
- Tries mise first, falls back to brew
- Added `go` to install hints map
- 2 builtins tests + 1 integration test added

### Decisions
- **Skipped**: `pi:install-ruby` (too niche), `pi:install-shfmt` (too niche), matrix builds (not PI's job)
- **env: on install phases**: Not implemented — install phases suppress stdout/stderr, so env vars there aren't as useful. Can be added later if needed.

## Subtasks
- [x] Implement `env:` field on steps
- [x] Implement `pi:install-go` built-in
- [x] Update architecture.md and README.md

## Blocked By
34-clone-and-examine-fzf (completed)
