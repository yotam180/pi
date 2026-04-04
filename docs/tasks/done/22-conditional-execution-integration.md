# Conditional Execution Integration Testing & QA

## Type
improvement

## Status
done

## Priority
medium

## Project
04-conditional-step-execution

## Description
Create a comprehensive example workspace and integration test suite that exercises all aspects of conditional step execution end-to-end. This is the final QA pass for Project 04.

Create `examples/conditional/` workspace with automations that use `if:` at step, automation, and setup levels, covering all predicate types and boolean operators. Write integration tests that run the binary against this workspace and verify correct skip/execute behavior.

Also update `pi info` to show `if:` conditions on steps and automations, and update `pi list` if needed.

## Acceptance Criteria
- [x] `examples/conditional/` workspace exists with representative automations
- [x] Integration tests cover: step-level if, automation-level if, setup entry if, skipped steps, various predicates
- [x] `pi info <name>` shows the `if:` condition when present (on automation and on steps)
- [x] All existing tests still pass (no regressions)
- [x] Architecture doc updated with conditional execution details
- [x] README updated if needed

## Implementation Notes

### Example workspace (`examples/conditional/`)
Expanded the existing workspace with 5 new automations covering predicates not previously exercised:
- `env-check.yaml` — tests `env.PI_TEST_VAR` and `not env.PI_TEST_VAR`
- `command-check.yaml` — tests `command.bash` (exists) and `command.nonexistent_xyz_tool_42` (missing)
- `file-check.yaml` — tests `file.exists("pi.yaml")`, `file.exists("nonexistent.yaml")`, `dir.exists(".pi")`, `dir.exists("nonexistent-dir")`
- `complex-bool.yaml` — tests complex expressions: `os.macos or os.linux`, `command.bash and file.exists("pi.yaml")`, `(os.windows and os.linux) or (os.macos and os.windows)`, `not os.windows`
- `conditional-with-if.yaml` — tests automation-level `if:` combined with step-level `if:`

### `pi info` updates (`internal/cli/info.go`)
- Added `Condition:` line when automation has `if:` field
- Added `Step details:` section when any step has `if:` — shows each step with its type, value (truncated to 40 chars), and `[if: <expr>]` when present
- Steps without `if:` shown without the `[if:]` tag
- Section is omitted entirely when no steps have conditions (backward compatible output)

### Integration tests
Added 10 new integration tests (18 total conditional tests):
- `TestConditional_EnvCheck_WithVar` / `WithoutVar` — env predicate with/without env var set
- `TestConditional_CommandCheck` — command availability predicate
- `TestConditional_FileCheck` — file.exists/dir.exists predicates
- `TestConditional_ComplexBool` — complex boolean expressions
- `TestConditional_CombinedAutomationAndStepIf` — both levels combined
- `TestConditional_Info_AutomationLevelIf` / `StepLevelIf` / `NoCondition` — pi info condition display
- `TestConditional_List_AllAutomations` — all 14 automations discoverable

### Unit tests
Added 4 new unit tests for info.go:
- `TestShowAutomationInfo_AutomationLevelIf`
- `TestShowAutomationInfo_NoConditionWhenAbsent`
- `TestShowAutomationInfo_StepLevelIf`
- `TestShowAutomationInfo_NoStepDetailsWithoutConditions`

### Test totals
320 tests total (up from 306), all passing with race detection.

## Subtasks
- [x] Create `examples/conditional/` workspace
- [x] Write integration tests
- [x] Update `pi info` to display `if:` conditions
- [x] Update docs (architecture.md, README.md)
- [x] Full regression run

## Blocked By
19-if-on-steps
20-if-on-automations
21-if-on-setup-entries
