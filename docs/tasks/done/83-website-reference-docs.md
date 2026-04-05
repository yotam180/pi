# Website: Reference Documentation

## Type
feature

## Status
done

## Priority
high

## Project
14-documentation-website

## Description
Write the five Reference pages. Reference pages are exhaustive lookup tables — every field, every flag, every valid value, with a one-line description each and a short example where needed. A developer should be able to answer "what flags does `pi run` accept?" or "what is the exact syntax for a `first:` block?" by opening Reference, not by reading Concepts again.

Source of truth: `docs/README.md`. Every item listed there must appear in the Reference. No items should be omitted.

## Acceptance Criteria
- [x] `cli.md` documents every command listed in the CLI Reference table in `docs/README.md`, with every flag for each command
- [x] `automation-yaml.md` documents every top-level automation field, every step type, and every step modifier — in a consistent format (field name, type, required/optional, description, example)
- [x] `conditions.md` documents every supported predicate for `if:` expressions, with the exact syntax and a short example
- [x] `builtins.md` documents every `pi:*` built-in automation with its description and any `inputs:` it takes
- [x] `pi-package-yaml.md` documents the `pi-package.yaml` file format and its fields
- [x] All pages use a consistent format (table or definition list, not prose paragraphs)
- [x] Each reference entry has a short example where syntax is non-obvious
- [x] Cross-links to Concepts pages for deeper context

## Implementation Notes

### Approach
All five reference pages were written from scratch, replacing the stub content. The source of truth (`docs/README.md`) was used for all feature descriptions.

### Content structure
- **cli.md**: Each command gets its own `##` section with usage lines, flags table, description, and example. All 11 commands documented with all flags.
- **automation-yaml.md**: Organized by concept (top-level fields → steps → step types → modifiers → shorthand → env → dir → timeout → silent → pipe → parent_shell → first → install → inputs). Tables for structured fields, code examples for syntax.
- **conditions.md**: Predicates table, operators table with precedence, behavior sections for steps/automations/setup/first blocks, examples of combining conditions.
- **builtins.md**: Grouped by category (Installers, Docker, Dev Tools). Each automation has a description table (condition, inputs, test, install, version) and usage example.
- **pi-package-yaml.md**: Short page covering the single `min_pi_version` field, behavior, and file location.

### Cross-links added
- automation-yaml.md → Automations concept, Step Types concept, Builtins reference, Packages concept, Conditions reference, Parent Shell guide
- cli.md → Packages concept, Shell Shortcuts concept, Conditions reference, Parent Shell guide
- conditions.md → automation-yaml reference, pi-yaml concept
- builtins.md → automation-yaml install block reference
- pi-package-yaml.md → Packages concept

### Completeness notes
- `pi:install-ruby` is listed in docs/README.md but no built-in YAML file exists in `internal/builtins/embed_pi/`. Omitted from builtins.md since it's not a real built-in yet. The README should be corrected when this is addressed.
- Deprecated `name:` field documented as a note in automation-yaml.md.
- Website builds successfully with all 21 pages.

## Subtasks
- [x] Write `cli.md` with all commands and flags
- [x] Write `automation-yaml.md` with all fields
- [x] Write `conditions.md` with all predicates and operators
- [x] Write `builtins.md` with all pi:* automations
- [x] Write `pi-package-yaml.md`
- [x] Verify completeness against `docs/README.md`
- [x] Add cross-links to Concepts pages

## Blocked By
79-website-scaffold-and-ci (done)
