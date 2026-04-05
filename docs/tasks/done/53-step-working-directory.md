# Step Working Directory (`dir:`)

## Type
feature

## Status
done

## Priority
high

## Project
standalone

## Description
Add a `dir:` field to steps that allows overriding the working directory for a step's execution. Currently all steps run with `cmd.Dir = repoRoot`, which forces users to use `cd dir && command` hacks in bash steps when they need to run in a subdirectory. This is a natural customization point that users will want.

The `dir:` field is optional. When specified, it's resolved relative to the repo root. Absolute paths are used as-is. When absent, behavior is unchanged (repo root is used).

This feature exercises the full stack: YAML parsing → automation model → RunContext → runner execution → CLI display → tests.

### Behavior
- `dir:` is valid on all step types: bash, python, typescript, run
- `dir:` value is resolved relative to the repo root (not the automation file)
- Absolute paths are used as-is
- The resolved directory must exist at execution time; if not, an error is returned
- `pi info` shows `[dir: <path>]` annotation on steps with `dir:` set
- `dir:` is independent of other step flags (silent, env, if, parent_shell, pipe_to)
- `parent_shell` steps: `dir:` is ignored (parent_shell commands don't execute in a subprocess)

## Acceptance Criteria
- [x] `dir:` field added to Step struct and parsed from YAML
- [x] Step execution uses `dir:` to override working directory
- [x] Works across all step types (bash, python, typescript)
- [x] `dir:` resolved relative to repo root; absolute paths preserved
- [x] Missing directory produces a clear error at execution time
- [x] `pi info` shows `dir:` annotation on steps
- [x] Unit tests for parsing, validation, execution
- [x] Integration tests with example workspace
- [x] All existing tests pass (no regressions)
- [x] Architecture docs updated

## Implementation Notes
### Design decisions:
- `dir:` is resolved relative to repo root, not the automation file's directory. Rationale: the repo root is the "anchor" for all PI operations; users think in terms of project-relative paths. Script file paths are different — they're relative to the automation file because they're assets of that automation.
- Validation of directory existence happens at execution time, not parse time. This matches how PI handles other runtime concerns (command availability, etc.) and avoids false errors when automations are loaded but not executed.
- `parent_shell` steps ignore `dir:` since they write commands to an eval file, not execute them. The parent shell is responsible for its own directory context.
- The `dir:` override is passed through `RunContext` to keep runners decoupled from executor internals.
- `run:` steps with `dir:` don't change the working directory of the called automation — each automation manages its own context.

### Key files modified:
- `internal/automation/automation.go` — add Dir field to Step, parse from stepRaw
- `internal/executor/runner_iface.go` — add DirOverride to RunContext
- `internal/executor/runners.go` — use DirOverride in runStepCommand()
- `internal/executor/executor.go` — resolve dir: and populate RunContext
- `internal/executor/helpers.go` — add resolveStepDir() helper
- `internal/cli/info.go` — display dir: annotation
- Tests across executor and integration packages

## Subtasks
- [x] Create task file
- [x] Add Dir to Step struct and stepRaw
- [x] Add resolveStepDir() helper
- [x] Wire through RunContext
- [x] Update runStepCommand() to use DirOverride
- [x] Update pi info display
- [x] Add unit tests
- [x] Add integration tests
- [x] Update docs
- [x] Full QA pass

## Blocked By
