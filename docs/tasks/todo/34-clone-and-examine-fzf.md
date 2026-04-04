# Clone and Examine fzf Workflows

## Type
research

## Status
todo

## Priority
high

## Project
07-fzf-adoption-test

## Description
Clone junegunn/fzf into `~/projects/fzf` and examine all developer workflows. Document every build command, test command, lint/format command, Docker operation, setup step, and release command. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/junegunn/fzf.git ~/projects/fzf`
2. Read the Makefile thoroughly — document every target
3. Read install/uninstall scripts
4. Read CI workflows (`.github/workflows/`)
5. Read `.goreleaser.yml` if present
6. List all tools/runtimes required (Go, Ruby, shellcheck, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [ ] fzf cloned to `~/projects/fzf`
- [ ] Every Makefile target documented
- [ ] Every script documented
- [ ] Every CI workflow documented
- [ ] Required tools/runtimes listed
- [ ] PI feature gap analysis completed
- [ ] Findings recorded in Implementation Notes

## Implementation Notes

## Subtasks
- [ ] Clone repo
- [ ] Document Makefile targets
- [ ] Document scripts
- [ ] Document CI workflows
- [ ] List required tools
- [ ] Assess PI feature coverage
- [ ] Document gaps

## Blocked By
