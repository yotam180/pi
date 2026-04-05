# Update `docs/README.md` for External Packages

## Type
chore

## Status
done

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
- [x] All four reference formats documented with examples
- [x] `packages:` block fully documented (both simple and object form)
- [x] `pi add` documented with all flag variants
- [x] On-demand advisory behavior explained
- [x] Private repo auth documented
- [x] "Writing a package repo" section exists and is concise
- [x] `pi-package.yaml` documented as optional
- [x] CLI reference table updated with any new/changed commands

## Implementation Notes
- Expanded the existing Packages section with new subsections: `pi add`, On-demand fetching, Private repositories, Writing a package repo, `pi-package.yaml`
- Replaced the outdated "Marketplace" subsection in Automation Resolution with "On-demand" entry in the resolution priority list
- Added detail about mutable refs, date-stamped cache keys, and reproducibility warnings to Source types
- Added alias validation rule (`/` not allowed) to Aliases section
- CLI reference table already had `pi add` and `pi list --all` — no changes needed
- Verified all docs match the implemented behavior in `cli/add.go`, `cli/discover.go`, `config/config.go`, `cache/cache.go`, `refparser/refparser.go`

## Subtasks
- [x] Draft documentation for all sections
- [x] Review against implemented behavior
- [x] Update CLI reference table

## Blocked By
68-automation-reference-parser, 69-github-package-cache, 70-packages-declaration-in-pi-yaml, 71-on-demand-fetch-with-advisory, 72-pi-add-command, 73-pi-list-source-indicators
