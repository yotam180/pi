# Playbook: Real-World Adoption Test

Test PI by adopting it in a real open-source repository. This validates that PI's feature set handles real-world workflows and surfaces missing capabilities.

## Phase 1: Clone & Examine

1. Clone the target repo into `~/projects/<repo-name>`
2. Read the repo's README, Makefile, scripts/, docker-compose files, CI workflows
3. List every developer workflow you find:
   - Build commands
   - Test commands
   - Lint/format commands
   - Docker operations
   - Setup/install steps
   - Deploy/release commands
4. For each workflow, note:
   - What tool/runtime it requires
   - Whether PI can model it today
   - If not, what PI feature is missing
5. Document findings in the task file under Implementation Notes

## Phase 2: Create Feature Tasks

For each gap identified in Phase 1:
- If it's a missing PI feature, create a task in `docs/tasks/todo/`
- If it's a missing built-in automation, create a task for adding it
- Set proper priority and link to the parent project
- Be specific about what the feature should do and why it's needed

## Phase 3: Transform the Repo

1. Create `.pi/` folder in the target repo
2. Create `pi.yaml` with project config, shortcuts, and setup entries
3. Write automation YAML files for each workflow identified in Phase 1
4. Write setup automations for developer onboarding
5. Define shell shortcuts for common operations
6. Test every automation: `pi run <name>` should work identically to the original command
7. Test `pi setup` on a fresh state
8. Test `pi shell` and verify shortcuts work

## Key Principles

- Use the **development version** of `pi` from this repo (build with `go build -o ~/go/bin/pi ./cmd/pi/`)
- Every automation should produce identical output to the original command
- If PI can't model something cleanly, prefer creating a new PI feature over a hacky workaround
- Document every quirk and decision in the task file
