# Clone and Examine bat Workflows

## Type
research

## Status
todo

## Priority
high

## Project
08-bat-adoption-test

## Description
Clone sharkdp/bat into `~/projects/bat` and examine all developer workflows. Document every build command, test command, lint/format command, CI workflow, and release process. For each workflow, assess whether PI can model it today or if a new feature is needed.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 1).

### Steps
1. `git clone https://github.com/sharkdp/bat.git ~/projects/bat`
2. Read `Cargo.toml` thoroughly — document build config and dependencies
3. Read any Makefile or build scripts
4. Read CI workflows (`.github/workflows/`)
5. Read any release configuration (goreleaser, cargo-dist, etc.)
6. List all tools/runtimes required (Rust, cargo, clippy, rustfmt, etc.)
7. For each workflow, note whether PI can model it and what's missing
8. Record all findings in Implementation Notes

## Acceptance Criteria
- [ ] bat cloned to `~/projects/bat`
- [ ] Every build/test/lint command documented
- [ ] Every CI workflow documented
- [ ] Required tools/runtimes listed
- [ ] PI feature gap analysis completed
- [ ] Findings recorded in Implementation Notes

## Implementation Notes

## Subtasks
- [ ] Clone repo
- [ ] Document Cargo/build config
- [ ] Document CI workflows
- [ ] List required tools
- [ ] Assess PI feature coverage
- [ ] Document gaps

## Blocked By
