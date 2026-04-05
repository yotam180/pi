# Documentation Website

## Status
in_progress

## Priority
high

## Description
Build a public-facing documentation website for PI that serves as both a marketing page and a complete learning resource. The site should project the right "aura" — it is an AI-era-native tool, and the documentation should reflect that through excellent structure, fast navigation, and first-class LLM/AI support (llms.txt, markdown mirrors, copy-to-clipboard, open-in-Claude actions). It must be auto-deployed from the repo via CI/CD.

## Goals
- A compelling landing page that communicates what PI is and how to install it in 10 seconds
- Complete Getting Started path: from zero to first running automation in 5 minutes, using short forms first
- Full Concepts section: every core abstraction explained with mental models, not just feature descriptions
- Complete Reference section: every CLI flag, every YAML field, every condition predicate documented
- Practical Guides: real workflows developers actually face (publishing packages, cross-platform scripts, parent shell, private repos)
- First-class AI/LLM support: `llms.txt`, markdown page mirrors, compiled `llms-ctx.txt`, copy-to-markdown button, open-in-Claude action
- Auto-deployed to GitHub Pages on every push to `main` that touches `website/`

## Background & Context
PI has deep, well-considered documentation in `docs/README.md` but it is agent-facing memory, not user-facing documentation. There is no public website. As PI matures toward distribution, a great website is critical for credibility, adoption, and making it easy for developers (and their AI assistants) to use the tool correctly. This also serves as a reference for the growing number of users who find PI via Homebrew or GitHub.

The key pedagogical decision: start with short forms. A first-time developer should never be confronted with `steps:` and `pi.yaml` on their first page. They should see `bash: go test ./...` running immediately, then discover the full form naturally as they need more power.

## Scope

### In scope
- Astro Starlight site in `website/` directory
- Custom landing page (Astro page, not a Starlight docs page)
- Getting Started section (Introduction, Installation, Quick Start)
- Concepts section (Automations, pi.yaml, Step Types, Shell Shortcuts, Packages)
- Reference section (CLI, YAML Spec, Conditions, Built-ins, pi-package.yaml)
- Guides section (Setup automations, cross-platform scripts, publishing to GitHub, private repos, parent shell steps)
- `llms.txt`, `.md` page mirrors, `llms-ctx.txt`, starlight-page-actions plugin
- GitHub Actions workflow deploying to GitHub Pages on push to `main`
- Dark mode, full-text search (Pagefind, built-in with Starlight)

### Out of scope
- Logo / visual identity work (deferred)
- Interactive YAML playground in the browser (deferred)
- Versioned docs (single version for now)
- Blog / changelog (deferred)
- Custom domain setup (deferred; GitHub Pages subdomain fine for now)
- i18n / translations

## Success Criteria
- [ ] `cd website && npm run build` completes without errors
- [ ] GitHub Actions workflow deploys the site to GitHub Pages on push to `main`
- [x] Landing page has install command with copy button, feature summary, and link to docs
- [ ] Getting Started Quick Start uses short-form syntax first, reveals full form later with explanation
- [ ] Every CLI command listed in `docs/README.md` has a corresponding reference page entry
- [ ] Every YAML field and step modifier listed in `docs/README.md` has a documented reference entry
- [ ] All `if:` predicates are documented in the Conditions reference
- [x] Guides cover: writing a setup automation, cross-platform scripts, publishing a package to GitHub, consuming packages, private repos, parent shell steps
- [ ] `public/llms.txt` exists and links to key documentation pages
- [ ] `starlight-page-actions` plugin provides "Copy as Markdown" button on all doc pages
- [ ] Site scores 90+ on Lighthouse (Starlight default achieves this)

## Notes
- The `docs/` folder is agent memory and stays as-is. The website content in `website/src/content/docs/` is authored separately for human readers.
- Starlight is the framework. Do not switch to Docusaurus or Mintlify without a clear reason.
- The `website/` directory lives at the repo root alongside `cmd/`, `internal/`, `docs/`.
- Task execution order matters: task 79 (scaffold) must be completed before any content tasks. Content tasks (81-85) can proceed in parallel after 79. Task 80 (landing page) and task 85 (AI support) can also proceed in parallel with content.
- When writing content, use the `docs/README.md` as the source of truth for all feature descriptions, behaviors, and examples. Do not invent behavior.
