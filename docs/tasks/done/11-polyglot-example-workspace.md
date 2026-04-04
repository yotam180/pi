# Polyglot Example Workspace

## Type
infra

## Status
done

## Priority
medium

## Project
02-polyglot-runner-and-shell-integration

## Description
Create an `examples/polyglot/` workspace that demonstrates mixing bash, Python, and TypeScript steps with pipe support. This workspace serves as the integration test bed for Project 2 features and as documentation for users.

## Acceptance Criteria
- [x] `examples/polyglot/` workspace exists with `pi.yaml`, `.pi/` folder, and `README.md`
- [x] Contains at least 3 automations demonstrating:
  - A Python step (inline and file)
  - A TypeScript step (inline and file)
  - A multi-step automation with `pipe_to: next` piping bash output through Python
- [x] All automations work end-to-end with `pi run`
- [x] Integration tests added to `tests/integration/` covering these automations
- [x] No references to Vyper or internal tooling

## Implementation Notes

Created 6 automations organized in 3 domains:

**text/** — Python demonstrations
- `text/reverse` — inline Python: reverses text (with optional arg)
- `text/transform` — Python file (`transform.py`): reads piped input, formats into numbered box

**data/** — TypeScript demonstrations
- `data/generate` — inline TypeScript: generates JSON array
- `data/format` — TypeScript file (`format.ts`): reads piped JSON, outputs sorted leaderboard; uses `run:` step to call `data/generate` with `pipe_to: next`

**pipeline/** — Cross-language pipe chains
- `pipeline/etl` — 3-step chain: bash (CSV) → Python (JSON transform) → TypeScript (formatted output)
- `pipeline/wordcount` — 2-step chain: bash (generate text) → Python (count words)

10 integration tests in `tests/integration/polyglot_test.go` covering list, each automation, argument passing, step ordering, and subdirectory discovery.

## Subtasks
- [x] Create `examples/polyglot/` workspace structure
- [x] Write automations with Python, TypeScript, and pipe examples
- [x] Add integration tests
- [x] Write README

## Blocked By
09-pipe-support
