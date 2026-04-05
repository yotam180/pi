# Website: AI/LLM Support

## Type
feature

## Status
todo

## Priority
medium

## Project
14-documentation-website

## Description
Add first-class AI/LLM support to the documentation site. This means: a well-structured `llms.txt` that serves as a curated entry point for any AI tool, `.md` mirrors of all documentation pages, a compiled `llms-ctx.txt` for pasting into LLMs with large context windows, and UI actions on every page ("Copy as Markdown", "Open in Claude", "Open in ChatGPT"). PI is an AI-era-native tool — its documentation should be natively usable by AI assistants.

## Acceptance Criteria
- [ ] `public/llms.txt` exists, follows the llmstxt.org spec, and links to all major documentation sections with one-line descriptions
- [ ] `public/llms-ctx.txt` is generated during build: contains all doc pages concatenated as clean Markdown
- [ ] Every documentation page is accessible as clean Markdown at `<url>.md` (e.g., `/getting-started/quick-start.md`)
- [ ] `starlight-page-actions` plugin is configured and shows "Copy as Markdown" button on all doc pages
- [ ] "Open in Claude" and "Open in ChatGPT" action buttons are present on all doc pages
- [ ] The landing page mentions the LLM support (e.g., a small "Use with AI" section or a link to `/llms.txt`)
- [ ] `llms.txt` is validated: no broken links, all major pages referenced
- [ ] The CI workflow generates `llms-ctx.txt` as part of the build step (before deploying)

## Implementation Notes

### `public/llms.txt` format

Follow the llmstxt.org specification exactly. The file must be at the root of the deployed site (`/llms.txt`).

```markdown
# PI

> PI is a developer automation tool that replaces your team's shell_shortcuts.sh with structured, polyglot, version-controlled automations. Write YAML files in .pi/, run them with `pi run`, share them via GitHub packages.

PI automations are YAML files that chain steps in bash, Python, or TypeScript. The root `pi.yaml` declares shortcuts, setup sequences, and external packages.

## Getting Started

- [Introduction](/getting-started/introduction.md): What PI is and the problem it solves
- [Installation](/getting-started/installation.md): Install via Homebrew, curl, or binary download
- [Quick Start](/getting-started/quick-start.md): Run your first automation in 5 minutes

## Concepts

- [Automations](/concepts/automations.md): YAML files, naming, step types, modifiers, install blocks
- [pi.yaml](/concepts/pi-yaml.md): Root config — shortcuts, setup, packages
- [Step Types](/concepts/step-types.md): bash, python, typescript, run
- [Shell Shortcuts](/concepts/shell-shortcuts.md): pi shell, global wrapper, parent_shell
- [Packages](/concepts/packages.md): Sharing and consuming automations via GitHub

## Guides

- [Setup Automations](/guides/setup-automations.md): Write a full pi setup configuration
- [Cross-Platform Scripts](/guides/cross-platform-scripts.md): if:, first:, OS predicates
- [Publishing to GitHub](/guides/publishing-to-github.md): Create and tag a PI package repo
- [Using Packages](/guides/using-packages.md): Add, alias, and upgrade packages
- [Private Repositories](/guides/private-repos.md): SSH and GITHUB_TOKEN authentication
- [Parent Shell Steps](/guides/parent-shell-steps.md): Activate virtualenvs, cd, source files

## Reference

- [CLI Commands](/reference/cli.md): Every command and flag
- [Automation YAML](/reference/automation-yaml.md): Every field and step modifier
- [Conditions (if:)](/reference/conditions.md): All predicates and operators
- [Built-in Automations](/reference/builtins.md): The pi:* standard library
- [pi-package.yaml](/reference/pi-package-yaml.md): Package metadata spec

## Optional

- [GitHub Repository](https://github.com/yotam180/pi): Source code, issues, releases
```

### `.md` mirrors for all pages

Starlight does not generate `.md` mirrors by default. Options:

**Option A (recommended): Astro endpoint for each page**

Add a custom Astro integration or middleware that serves the raw markdown for each content page at `<slug>.md`. In Astro, you can create a catch-all route `src/pages/[...slug].md.ts` that reads the content collection entry and returns its raw markdown body.

```typescript
// src/pages/[...slug].md.ts
import { getCollection } from 'astro:content';
import type { APIRoute } from 'astro';

export const GET: APIRoute = async ({ params }) => {
  const docs = await getCollection('docs');
  const entry = docs.find(e => e.id === params.slug + '.md' || e.id === params.slug);
  if (!entry) return new Response('Not found', { status: 404 });
  return new Response(entry.body, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' }
  });
};

export async function getStaticPaths() {
  const docs = await getCollection('docs');
  return docs.map(entry => ({
    params: { slug: entry.slug },
  }));
}
```

**Option B: Build-time script**

A Node.js script run as part of the build that copies all `.md` source files into `dist/` at their URL-mapped paths. Simpler to implement but less elegant.

Use Option A. It integrates cleanly with Astro's static generation and doesn't require a separate script.

### `llms-ctx.txt` generation

`llms-ctx.txt` is a single file containing all documentation pages concatenated as Markdown, with section separators. It's what you paste into a Claude or ChatGPT context window when you want the LLM to know everything about PI.

Generate it in CI using the `llms_txt2ctx` CLI tool (Python package):

```bash
pip install llms_txt2ctx
llms_txt2ctx dist/llms.txt > dist/llms-ctx.txt
```

This tool reads `llms.txt`, follows all linked `.md` URLs (now served as static files), and concatenates them.

Add this step to the GitHub Actions workflow after `npm run build`:
```yaml
- name: Generate llms-ctx.txt
  run: |
    pip install llms_txt2ctx
    llms_txt2ctx website/dist/llms.txt > website/dist/llms-ctx.txt
```

Also update `llms.txt` to include a reference to `llms-ctx.txt`:
```markdown
## Context Files

- [Full documentation context](/llms-ctx.txt): All documentation concatenated — paste into any LLM
```

### `starlight-page-actions` plugin

The package is already installed in task 79. Wire it up in `astro.config.mjs`:

```javascript
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightPageActions from 'starlight-page-actions';

export default defineConfig({
  integrations: [
    starlight({
      plugins: [
        starlightPageActions({
          actions: [
            {
              label: 'Copy as Markdown',
              icon: 'document',
              action: 'copy-markdown',
            },
            {
              label: 'Open in Claude',
              icon: 'external',
              href: (page) => `https://claude.ai/new?q=${encodeURIComponent('Help me with: ' + page.url)}`,
            },
            {
              label: 'Open in ChatGPT',
              icon: 'external',
              href: (page) => `https://chatgpt.com/?q=${encodeURIComponent('Help me with: ' + page.url)}`,
            },
          ],
        }),
      ],
      // ... rest of config
    }),
  ],
});
```

Check the `starlight-page-actions` README for the exact API — it may differ from the above. The key behaviors to implement:
1. "Copy as Markdown" — copies the raw markdown of the current page to clipboard
2. "Open in Claude" — opens Claude with a pre-filled message referencing the page
3. "Open in ChatGPT" — opens ChatGPT with a pre-filled message

### Landing page AI mention

Add a small section to the landing page (task 80):
```
Use PI documentation with your AI assistant

/llms.txt  —  curated entry point for LLMs
/llms-ctx.txt  —  full documentation context (paste into any LLM)
Every page available as clean Markdown at <url>.md
```

Keep it concise — one row of three items, not a full section.

## Subtasks
- [ ] Write `public/llms.txt` with all pages linked
- [ ] Implement `.md` mirror endpoint (`src/pages/[...slug].md.ts`)
- [ ] Add `llms-ctx.txt` generation step to CI workflow
- [ ] Configure `starlight-page-actions` in `astro.config.mjs`
- [ ] Add AI support mention to landing page
- [ ] Verify `llms.txt` has no broken links after site build
- [ ] Verify `.md` mirrors work (spot-check 3 pages)

## Blocked By
79-website-scaffold-and-ci (needs the scaffold to exist)
80-website-landing-page (to add the AI mention section)
