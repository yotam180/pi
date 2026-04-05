# Website: Scaffold and CI/CD

## Type
infra

## Status
done

## Priority
high

## Project
14-documentation-website

## Description
Set up the Astro Starlight website skeleton in `website/` at the repo root, with all navigation stubs in place and a GitHub Actions workflow that builds and deploys to GitHub Pages on push to `main`. This is the foundation all content tasks depend on — it must be done first.

## Acceptance Criteria
- [x] `website/` directory exists at repo root with a working Astro Starlight installation
- [x] `cd website && npm run build` succeeds (stub pages are fine, no broken links in nav)
- [x] `cd website && npm run dev` starts a local dev server
- [x] Full navigation structure is in place as stub pages (correct filenames, correct sidebar config)
- [x] `astro.config.mjs` has the correct site title, description, and sidebar structure
- [x] `.github/workflows/docs.yml` workflow builds and deploys to GitHub Pages on push to `main` (path filter: `website/**`)
- [x] `website/` has a `.gitignore` ignoring `node_modules/` and `dist/`
- [x] `public/llms.txt` stub file exists (will be filled in by task 85)
- [x] `starlight-page-actions` npm package is installed (wired up in task 85, but installed here)
- [x] README note in repo root `README.md` mentions the website directory

## Implementation Notes

### Framework
Astro Starlight. Install with:
```
npm create astro@latest website -- --template starlight
```
Then install additional deps:
```
npm install starlight-page-actions
```

### Directory structure to create
```
website/
  src/
    content/
      docs/
        index.mdx                          ← redirects to /getting-started/introduction
        getting-started/
          introduction.md                  ← stub
          installation.md                  ← stub
          quick-start.md                   ← stub
        concepts/
          automations.md                   ← stub
          pi-yaml.md                       ← stub
          step-types.md                    ← stub
          shell-shortcuts.md               ← stub
          packages.md                      ← stub
        reference/
          cli.md                           ← stub
          automation-yaml.md               ← stub
          conditions.md                    ← stub
          builtins.md                      ← stub
          pi-package-yaml.md               ← stub
        guides/
          setup-automations.md             ← stub
          cross-platform-scripts.md        ← stub
          publishing-to-github.md          ← stub
          using-packages.md                ← stub
          private-repos.md                 ← stub
          parent-shell-steps.md            ← stub
    pages/
      index.astro                          ← custom landing page (task 80)
  public/
    llms.txt                               ← stub (task 85)
  astro.config.mjs
  package.json
  tsconfig.json
  .gitignore
```

### Sidebar configuration (in `astro.config.mjs`)
```js
sidebar: [
  {
    label: 'Getting Started',
    items: [
      { label: 'Introduction', slug: 'getting-started/introduction' },
      { label: 'Installation', slug: 'getting-started/installation' },
      { label: 'Quick Start', slug: 'getting-started/quick-start' },
    ],
  },
  {
    label: 'Concepts',
    items: [
      { label: 'Automations', slug: 'concepts/automations' },
      { label: 'pi.yaml', slug: 'concepts/pi-yaml' },
      { label: 'Step Types', slug: 'concepts/step-types' },
      { label: 'Shell Shortcuts', slug: 'concepts/shell-shortcuts' },
      { label: 'Packages', slug: 'concepts/packages' },
    ],
  },
  {
    label: 'Guides',
    items: [
      { label: 'Setup Automations', slug: 'guides/setup-automations' },
      { label: 'Cross-Platform Scripts', slug: 'guides/cross-platform-scripts' },
      { label: 'Publishing to GitHub', slug: 'guides/publishing-to-github' },
      { label: 'Using Packages', slug: 'guides/using-packages' },
      { label: 'Private Repositories', slug: 'guides/private-repos' },
      { label: 'Parent Shell Steps', slug: 'guides/parent-shell-steps' },
    ],
  },
  {
    label: 'Reference',
    items: [
      { label: 'CLI Commands', slug: 'reference/cli' },
      { label: 'Automation YAML', slug: 'reference/automation-yaml' },
      { label: 'Conditions (if:)', slug: 'reference/conditions' },
      { label: 'Built-in Automations', slug: 'reference/builtins' },
      { label: 'pi-package.yaml', slug: 'reference/pi-package-yaml' },
    ],
  },
],
```

### GitHub Actions workflow
File: `.github/workflows/docs.yml`

```yaml
name: Deploy Documentation

on:
  push:
    branches: [main]
    paths:
      - 'website/**'

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: website/package-lock.json
      - name: Install dependencies
        run: npm ci
        working-directory: website
      - name: Build site
        run: npm run build
        working-directory: website
      - uses: actions/upload-pages-artifact@v3
        with:
          path: website/dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

### Stub page format
Each stub should have a title and a one-line placeholder:
```md
---
title: Page Title
description: One-line description for SEO and social sharing
---

This page is coming soon.
```

### The docs root `index.mdx`
Should redirect to the Introduction page using a `<meta http-equiv="refresh">` or Astro's redirect. In Starlight, the simplest approach is to set the `template: splash` in frontmatter on the root and have the landing page at `src/pages/index.astro` (which is what task 80 will build). The `src/content/docs/index.mdx` should not exist — the root `/` is handled by the custom `src/pages/index.astro`.

## Subtasks
- [x] Scaffold Astro Starlight project in `website/`
- [x] Install `starlight-page-actions`
- [x] Create all stub pages with correct filenames
- [x] Configure sidebar in `astro.config.mjs`
- [x] Create `public/llms.txt` stub
- [x] Create `.github/workflows/docs.yml`
- [x] Verify `npm run build` passes
- [x] Add `website/` mention to root `README.md`

## Completion Notes

Scaffolded using `npm create astro@latest website -- --template starlight`. Astro v6.1.3, Starlight v0.38.2, starlight-page-actions v0.5.0.

Key decisions:
- Root `/` handled by `src/pages/index.astro` (currently a redirect to `/getting-started/introduction/`, task 80 will replace with landing page)
- No `src/content/docs/index.mdx` — avoids conflict with the custom landing page route
- `starlight-page-actions` wired into the Starlight plugin system in `astro.config.mjs` (ready for task 85 to configure actions)
- `.gitignore` covers `node_modules/`, `dist/`, and `.astro/` (generated types directory)
- Build produces 21 pages including Pagefind search index
- Dev server verified at `localhost:4321` returning 200

## Blocked By
None — this is the first task in the project.
