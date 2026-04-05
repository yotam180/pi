# Product Philosophy Document

## Type
chore

## Status
done

## Priority
high

## Project
standalone

## Description
Create `docs/philosophy.md` with PI's core product values, and update `AGENTS.md` to require agents to read it before implementing any user-facing feature. The philosophy document is the governing framework for all product decisions — it must exist before the `pi init`, `pi setup add`, and builtin expansion work begins.

## Acceptance Criteria
- [ ] `docs/philosophy.md` exists and covers: intent over syntax, progressive disclosure, forgiving by default, idempotent by design, transparent (not magic), consistent patterns, composable primitives, AI-native by default
- [ ] Each principle has a "The rule:" statement that makes it actionable
- [ ] `AGENTS.md` has a new "User-facing features" section that requires reading `docs/philosophy.md` before implementing user-facing work
- [ ] The AGENTS.md section includes a concrete checklist (not just a reference)

## Implementation Notes
The philosophy document should be written as a decision framework, not a manifesto. Each principle should be short, specific, and include a concrete example of how it applies to PI.

The AGENTS.md update must be prominent — not buried at the bottom. It should appear before the Code section since it governs a broader category of decisions.

## Subtasks
- [ ] Write `docs/philosophy.md` (8 principles)
- [ ] Update `AGENTS.md` with "User-facing features" section

## Blocked By
None
