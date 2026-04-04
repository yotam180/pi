# Agent Instructions

This document contains persistent instructions for all agents working on this repo.

## Documentation is your memory

`docs/` is the single source of truth. Read `docs/README.md` at the start of every session to orient yourself. No state is preserved between sessions — everything important must be written down.

## Working on tasks

1. Check `docs/tasks/in_progress/` first — resume any open task before starting something new.
2. If nothing is in_progress, pick the highest-priority task from `docs/tasks/todo/` and move it to `in_progress/`.
3. Keep the task file updated as you work: record decisions, approaches, and gotchas in `## Implementation Notes`.
4. When done, move the task file to `docs/tasks/done/`.
5. Commit all changes before the session ends.

## Creating a new task

1. Copy `docs/templates/task.md` into `docs/tasks/todo/`.
2. Name the file in kebab-case describing the work, e.g. `add-login-flag.md`.
3. Fill in every section. The description must be specific enough that any agent can start without extra context.
4. Set `## Type` to one of: `feature | bug | improvement | research | infra | chore`.
5. Set `## Priority` to `high | medium | low`.
6. Set `## Project` to the name of the parent project file, or `standalone`.

## Creating a new project

1. Copy `docs/templates/project.md` into `docs/projects/todo/`.
2. Name the file in kebab-case, e.g. `config-file-support.md`.
3. Fill in all sections, especially **Goals** and **Success Criteria** — these define done.
4. When a project moves to `in_progress`, create all its task files in `docs/tasks/todo/` at that time.
5. A project is done only when all its tasks are in `docs/tasks/done/` and the success criteria are met. Move the project file to `docs/projects/done/`.

## Priorities and balance

Balance work across: infrastructure, self-improvement, and feature development. Don't let any one area starve the others.

## Code

- Language: Go. Follow standard Go best practices.
- Test everything. Do QA before marking a task done.
- Run `go build ./...` and `go test ./...` before committing.
