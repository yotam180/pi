# Update `docs/README.md` for External Packages

## Type
chore

## Status
todo

## Priority
medium

## Project
13-external-packages

## Description
Update `docs/README.md` to document the full external packages system once all feature tasks (68–73) are complete. Replace the existing stub "Marketplace" section with complete, accurate documentation.

Sections to add/update:

1. **Reference formats** — the four source types with syntax and examples
2. **`packages:` in `pi.yaml`** — full example with GitHub and file sources, alias syntax
3. **`pi add`** — usage examples
4. **`pi setup` behavior** — how packages are fetched during setup
5. **On-demand fetching** — explain the advisory and how to respond to it
6. **`pi list --all`** — how to browse available automations
7. **Private repos** — SSH and `GITHUB_TOKEN` instructions
8. **Writing a package repo** — minimal instructions: create a GitHub repo, put automations in `.pi/`, optionally add `pi-package.yaml` with `min_pi_version`
9. **`pi-package.yaml` reference** — optional file, supported fields

## Acceptance Criteria
- [ ] All four reference formats documented with examples
- [ ] `packages:` block fully documented (both simple and object form)
- [ ] `pi add` documented with all flag variants
- [ ] On-demand advisory behavior explained
- [ ] Private repo auth documented
- [ ] "Writing a package repo" section exists and is concise
- [ ] `pi-package.yaml` documented as optional
- [ ] CLI reference table updated with any new/changed commands

## Implementation Notes

## Subtasks
- [ ] Draft documentation for all sections
- [ ] Review against implemented behavior
- [ ] Update CLI reference table

## Blocked By
68-automation-reference-parser, 69-github-package-cache, 70-packages-declaration-in-pi-yaml, 71-on-demand-fetch-with-advisory, 72-pi-add-command, 73-pi-list-source-indicators
