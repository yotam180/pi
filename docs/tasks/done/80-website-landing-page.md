# Website: Landing Page

## Type
feature

## Status
done

## Priority
high

## Project
14-documentation-website

## Description
Build the custom landing page at `website/src/pages/index.astro`. This is a marketing page, not a docs page — it communicates what PI is, shows the install command, and makes the value proposition obvious in 10 seconds. No logo or visual identity work yet; focus on clarity, layout, and the code example.

## Acceptance Criteria
- [x] Landing page renders at the site root `/`
- [x] Hero section has: a punchy tagline, a one-sentence description, the install command with a copy button, and a "Get Started" + "View on GitHub" CTA
- [x] Install command displayed is `brew install yotam180/pi/pi` (primary) with a secondary GitHub Releases link below
- [x] Feature section has 4 concise feature tiles (polyglot, lives in repo, shell shortcuts, shareable packages)
- [x] A code/YAML example section shows a real, non-trivial `pi.yaml` + automation that a developer would actually write
- [x] "One automation, any language" step types section shows the supported languages (bash, Python, TypeScript, run)
- [x] "Start simple" section shows single-step shorthand syntax
- [x] Footer with GitHub link, Docs, CLI Reference, and "Built with Astro Starlight"
- [x] Dark mode works (uses Starlight's `StarlightPage` with splash template — inherits theme toggle)
- [x] Page is responsive (mobile + desktop via CSS Grid and media queries)
- [x] No broken links

## Implementation Notes

### Approach
Used `StarlightPage` component with `template: 'splash'` instead of a fully standalone page. This gives us:
- Starlight's header/navbar with dark mode toggle and search
- Consistent theming via Starlight's CSS custom properties (`--sl-color-*`)
- No sidebar (splash template)
- Starlight's `hero` prop for the hero section with CTAs and tagline

### Tagline chosen
"Replace your `shell_shortcuts.sh`" — the most concrete and recognizable option. Every developer has one of these files.

### Sections
1. **Hero** — Starlight's built-in hero component with tagline, description, and two CTAs (Get Started, View on GitHub)
2. **Install command** — `brew install yotam180/pi/pi` with copy-to-clipboard button (vanilla JS, no framework needed). Fallback link to GitHub Releases
3. **Feature tiles** — 4 cards in a responsive grid: Polyglot, Lives in repo, Shell shortcuts, Shareable packages
4. **Code example** — 3-panel grid: `.pi/docker/up.yaml`, `pi.yaml`, and terminal output showing `pi shell` + shortcut usage
5. **Step types** — 4 cards showing `bash:`, `python:`, `typescript:`, `run:` with descriptions
6. **Start simple** — 2-panel grid showing single-step shorthand syntax (the "easy entry point")
7. **Footer** — Links to GitHub, Docs, CLI Reference + attribution

### Styling
All styling uses Starlight CSS custom properties (`--sl-color-gray-*`, `--sl-color-green-high`, `--sl-color-text-accent`, `--sl-font-mono`) with fallback values. This ensures dark/light mode works automatically. Scoped `<style>` block — no global CSS files added. Responsive breakpoint at 640px.

### Copy-to-clipboard
Vanilla JS event listener on the copy button. Changes icon to `✓` for 1.5s on successful copy, then reverts. Uses `navigator.clipboard.writeText()`.

## Subtasks
- [x] Create `src/pages/index.astro` with hero section (via StarlightPage splash template)
- [x] Add feature tiles section
- [x] Add code example section (YAML + terminal output, 3-panel grid)
- [x] Add step types section
- [x] Add "start simple" shorthand section
- [x] Add copy-to-clipboard for install command
- [x] Add footer
- [x] Verify dark mode (inherited from StarlightPage)
- [x] Verify mobile layout (responsive CSS grid with media queries)
- [x] Verify site builds (`npm run build` — 0 errors)

## Blocked By
79-website-scaffold-and-ci (need the Astro project to exist first) — ✅ done
