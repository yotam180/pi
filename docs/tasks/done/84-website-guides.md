# Website: Guides

## Type
feature

## Status
done

## Priority
high

## Project
14-documentation-website

## Description
Write six practical guides that walk through real workflows. Guides differ from Concepts: they are task-oriented ("how do I do X") rather than concept-oriented ("what is X"). Each guide should read like a short tutorial — intro, step-by-step, expected outcomes. The packaging guides (publishing to GitHub and using packages) are especially important and should be detailed enough that a developer can follow them start to finish with zero prior knowledge.

## Acceptance Criteria
- [x] `setup-automations.md` walks through writing a realistic `pi setup` configuration for a project with mixed OS support
- [x] `cross-platform-scripts.md` walks through using `if:`, `first:`, and OS predicates to write automations that work on macOS and Linux
- [x] `publishing-to-github.md` walks through creating a package repo, organizing it, tagging a release, and making it consumable — complete from `git init` to `pi add`
- [x] `using-packages.md` walks through adding a package, using it in automations, upgrading versions, and understanding the alias system
- [x] `private-repos.md` covers the SSH and GITHUB_TOKEN approaches with concrete setup steps
- [x] `parent-shell-steps.md` explains the problem, the solution, and all the ways `parent_shell: true` works (shortcuts, global wrapper, raw binary)
- [x] Every guide has a clear "What you'll learn" intro and a "Summary" at the end
- [x] Every code block is complete and runnable

## Implementation Notes

All six guides written as full-content pages matching the existing documentation style (concise, practical, with Starlight admonitions, cross-links to Concepts and Reference pages).

### Writing approach
- Each guide follows the same structure: frontmatter → intro sentence → "What you'll learn" list → horizontal rule → walkthrough sections → "Summary" at the end
- Code examples drawn from `docs/README.md` (the source of truth) and adapted for tutorial context
- Cross-links added to relevant Concepts pages (Automations, Packages, Shell Shortcuts, Conditions reference) and Reference pages
- Used Starlight admonitions (`:::tip`, `:::note`, `:::caution`) consistently with existing concept pages

### Guide summaries

**setup-automations.md** — Covers the `setup:` block in `pi.yaml`, bare string vs object form, built-in installer automations table, conditional setup entries, `pi-setup-<project>` helper, packages in setup, `--no-shell`/`--silent`/`--loud` flags, testing with Docker.

**cross-platform-scripts.md** — Covers `if:` predicates on steps, `first:` blocks for mutual exclusion, architecture-specific builds, optional tool checks, file/dir/env checks, shell detection, complete predicate reference table.

**publishing-to-github.md** — Linear walkthrough from `git init` to `pi add`. Covers repo structure, automation organization, `pi-package.yaml`, tagging releases, consumer workflow, versioning, naming conventions, composability best practices.

**using-packages.md** — Covers `pi add`, file sources, running by plain name / alias / full reference, alias setup, `pi setup` package fetching, `pi list` with SOURCE column, collision behavior, on-demand fetching, upgrading versions.

**private-repos.md** — Covers the three-method auth chain (SSH → GITHUB_TOKEN → HTTPS), SSH setup with step-by-step instructions, GITHUB_TOKEN in GitHub Actions and with PATs, verification commands, error message table, multi-repo guidance.

**parent-shell-steps.md** — Explains the subprocess limitation, the `parent_shell: true` solution, the three-part mechanism (wrapper → eval file → source), shell integration installation, common use cases (virtualenv, cd, export, nvm), multiple parent shell steps, rules/restrictions table, what doesn't work.

### Verification
- Website builds successfully with `npm run build` (21 pages, 0 errors)
- All Go tests pass (`go build ./...` and `go test ./...`)
- Sidebar correctly configured in `astro.config.mjs` — all 6 guides appear in the right order

## Subtasks
- [x] Write `setup-automations.md`
- [x] Write `cross-platform-scripts.md`
- [x] Write `publishing-to-github.md`
- [x] Write `using-packages.md`
- [x] Write `private-repos.md`
- [x] Write `parent-shell-steps.md`
- [x] Verify all code examples
- [x] Add cross-links to Concepts and Reference pages

## Blocked By
79-website-scaffold-and-ci
