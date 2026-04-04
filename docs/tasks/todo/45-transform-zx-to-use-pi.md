# Transform zx to Use PI

## Type
feature

## Status
todo

## Priority
medium

## Project
10-zx-adoption-test

## Description
Create PI automations for zx's developer workflows in `~/projects/zx`. This is the final phase of the adoption test — actually using PI to model a real TypeScript project's workflows.

Follow the playbook at `docs/playbooks/real-world-adoption-test.md` (Phase 3).

### Steps
1. Build the development `pi` binary: `cd ~/projects/vyper-tooling && go build -o ~/go/bin/pi ./cmd/pi/`
2. Create `~/projects/zx/.pi/` folder
3. Create `~/projects/zx/pi.yaml` with project config
4. Write automation YAML files for each workflow from task 43:
   - Build/compile
   - Test: unit tests, integration tests
   - Lint/format
   - Setup: Node.js, npm install
5. Define setup automations for dev environment
6. Define shell shortcuts for common operations
7. Test every automation against the original commands
8. Test `pi setup` and `pi shell`
9. Document any quirks and whether they should be solved in YAML or as new PI features

## Acceptance Criteria
- [ ] `~/projects/zx/.pi/` contains automations for all major workflows
- [ ] `~/projects/zx/pi.yaml` has shortcuts and setup entries
- [ ] `pi setup` in zx installs all required tools
- [ ] `pi shell` in zx installs working shortcuts
- [ ] `pi run <name>` for each automation produces correct results
- [ ] `pi list` shows all automations with descriptions
- [ ] Quirks documented in Implementation Notes with feature-vs-YAML analysis

## Implementation Notes

## Subtasks
- [ ] Create `.pi/` and `pi.yaml`
- [ ] Write build automations
- [ ] Write test automations
- [ ] Write lint/format automations
- [ ] Write setup automations
- [ ] Define shortcuts
- [ ] Test all automations
- [ ] Test `pi setup` and `pi shell`
- [ ] Document quirks and lessons learned

## Blocked By
43-clone-and-examine-zx
44-implement-zx-feature-gaps
