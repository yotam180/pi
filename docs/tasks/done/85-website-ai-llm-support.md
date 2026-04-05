# Website: AI/LLM Support

## Type
feature

## Status
done

## Priority
medium

## Project
14-documentation-website

## Description
Add first-class AI/LLM support to the documentation site. This means: a well-structured `llms.txt` that serves as a curated entry point for any AI tool, `.md` mirrors of all documentation pages, a compiled `llms-ctx.txt` for pasting into LLMs with large context windows, and UI actions on every page ("Copy as Markdown", "Open in Claude", "Open in ChatGPT"). PI is an AI-era-native tool — its documentation should be natively usable by AI assistants.

## Acceptance Criteria
- [x] `public/llms.txt` exists, follows the llmstxt.org spec, and links to all major documentation sections with one-line descriptions
- [x] `public/llms-ctx.txt` is generated during build: contains all doc pages concatenated as clean Markdown
- [x] Every documentation page is accessible as clean Markdown at `<url>.md` (e.g., `/getting-started/quick-start.md`)
- [x] `starlight-page-actions` plugin is configured and shows "Copy as Markdown" button on all doc pages
- [x] "Open in Claude" and "Open in ChatGPT" action buttons are present on all doc pages
- [x] The landing page mentions the LLM support (e.g., a small "Use with AI" section or a link to `/llms.txt`)
- [x] `llms.txt` is validated: no broken links, all major pages referenced
- [x] The CI workflow generates `llms-ctx.txt` as part of the build step (before deploying)

## Implementation Notes

### Key decisions

1. **Plugin does the heavy lifting**: `starlight-page-actions@0.5.0` handles all three major features:
   - `.md` mirrors — via `vite-plugin-static-copy` that copies, cleans, and renames all `src/content/docs/**/*.md` files into `dist/`
   - Page action buttons — "Copy Markdown", "Open in ChatGPT", "Open in Claude" buttons on every doc page
   - The plugin can auto-generate `llms.txt` but only when `baseUrl` is set

2. **Curated `llms.txt`**: We intentionally do NOT set `baseUrl` on the plugin, which means the plugin skips auto-generating `llms.txt`. Instead, we write our own curated `public/llms.txt` following the llmstxt.org specification exactly. Our version has a proper description block, organized sections matching the sidebar, and one-line descriptions per page. It uses relative links (`getting-started/introduction.md`) which work correctly for the deployed site.

3. **`llms-ctx.txt` generation**: Instead of using the `llms_txt2ctx` Python tool (which requires a running server to follow URLs), we wrote `website/scripts/generate-llms-ctx.sh` — a simple bash script that parses `llms.txt`, extracts `.md` references, and concatenates the corresponding files from `dist/`. This runs after `npm run build` in the CI workflow and produces ~112KB of clean, concatenated documentation.

4. **No custom Astro endpoint needed**: The task originally planned a custom `src/pages/[...slug].md.ts` endpoint for `.md` mirrors. The plugin's `vite-plugin-static-copy` approach is better — it strips frontmatter, cleans Starlight components, and produces clean markdown at build time. No custom code needed.

5. **Landing page AI section**: Added a compact "Use with your AI assistant" section with three linked cards: `/llms.txt`, `/llms-ctx.txt`, and `<url>.md`. Styled consistently with the existing landing page design. Placed between the "Start simple" section and the footer.

### Files changed
- `website/public/llms.txt` — curated llmstxt.org-spec file with all 19 doc pages + context file link
- `website/astro.config.mjs` — configured `starlightPageActions()` with chatgpt, claude, and markdown actions
- `website/scripts/generate-llms-ctx.sh` — build script to concatenate all .md mirrors into llms-ctx.txt
- `website/src/pages/index.astro` — added "Use with your AI assistant" section
- `.github/workflows/docs.yml` — added llms-ctx.txt generation step after build

### Verification
- `npm run build` succeeds
- 19/19 `.md` mirrors generated in `dist/`
- 19/19 `llms.txt` links validated against `dist/` files
- `llms-ctx.txt` generated (111,584 bytes)
- Page action buttons (Copy Markdown, ChatGPT, Claude) present in built HTML
- Go build/vet/tests all pass (no Go changes in this task)

## Subtasks
- [x] Write `public/llms.txt` with all pages linked
- [x] Configure `starlight-page-actions` with ChatGPT, Claude, Copy Markdown actions
- [x] Write `scripts/generate-llms-ctx.sh` to concatenate .md mirrors
- [x] Add `llms-ctx.txt` generation step to CI workflow
- [x] Add AI support mention to landing page
- [x] Verify `llms.txt` has no broken links after site build
- [x] Verify `.md` mirrors work (spot-checked 3 pages: introduction, automations, cli)

## Blocked By
79-website-scaffold-and-ci (done)
80-website-landing-page (done)
