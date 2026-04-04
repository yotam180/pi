# Clone and Examine zx Workflows

## Type
research

## Status
todo

## Priority
high

## Project
10-zx-adoption-test

## Description
Clone google/zx into `~/projects/zx` and examine all developer workflows. Document every build command, test command, lint/format command, CI workflow, and release process. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/google/zx.git ~/projects/zx`
2. Read `package.json` — document build config, scripts, and dependencies
3. Read `tsconfig.json` and any build configuration
4. Read CI workflows (`.github/workflows/`)
5. Read any release configuration
6. List all tools/runtimes required (Node.js, npm, TypeScript, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [ ] zx cloned to `~/projects/zx`
- [ ] Every build/test/lint command documented
- [ ] Every CI workflow documented
- [ ] Required tools/runtimes listed
- [ ] PI feature gap analysis completed
- [ ] Findings recorded in Implementation Notes

## Implementation Notes

## Subtasks
- [ ] Clone repo
- [ ] Document build/package config
- [ ] Document CI workflows
- [ ] List required tools
- [ ] Assess PI feature coverage
- [ ] Document gaps

## Blocked By
