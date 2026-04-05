# Website: Landing Page

## Type
feature

## Status
todo

## Priority
high

## Project
14-documentation-website

## Description
Build the custom landing page at `website/src/pages/index.astro`. This is a marketing page, not a docs page — it communicates what PI is, shows the install command, and makes the value proposition obvious in 10 seconds. No logo or visual identity work yet; focus on clarity, layout, and the code example.

## Acceptance Criteria
- [ ] Landing page renders at the site root `/`
- [ ] Hero section has: a punchy tagline, a one-sentence description, the install command with a copy button, and a "Get Started" + "View on GitHub" CTA
- [ ] Install command displayed is `brew install pi` (primary) with a secondary curl option shown below
- [ ] Feature section has 3–4 concise feature tiles (polyglot, shareable packages, shell shortcuts, zero-config)
- [ ] A code/YAML example section shows a real, non-trivial `pi.yaml` + automation that a developer would actually write
- [ ] "Works with" or step types section shows the supported languages (bash, Python, TypeScript) briefly
- [ ] Footer with GitHub link and "Built with Astro Starlight"
- [ ] Dark mode works (Astro Starlight dark mode applies to the landing page)
- [ ] Page is responsive (mobile + desktop)
- [ ] No broken links

## Implementation Notes

### Philosophy
This page must answer three questions in order:
1. "What is PI?" — one sentence
2. "Why should I care?" — three bullet points or tiles
3. "How do I get it?" — install command, copy button, link to docs

The developer has 10 seconds of attention. Don't waste it on philosophy. Tagline first, install command second, example third.

### Tagline options to consider
- "Replace your shell_shortcuts.sh"
- "Structured automation for your whole team"
- "The automation tool that lives in your repo"
- "Polyglot, shareable team automations"

Pick the most concrete and specific one that matches what PI actually is.

### Hero layout
```
[Tagline — large, bold]
[One-sentence description]

$ brew install pi          [copy button]

[Get Started →]   [View on GitHub ↗]
```

### Feature tiles (3–4 tiles)
Use simple icons (Heroicons or Astro's icon set). Each tile: icon + title + 1-sentence description.

1. **Polyglot by default** — Steps in bash, Python, or TypeScript. Mix languages freely within a single automation.
2. **Lives in your repo** — Automations are YAML files in `.pi/`. Version-controlled, PR-reviewed, shared with the team.
3. **Shell shortcuts, zero setup** — `pi shell` turns any automation into a terminal command. Works from anywhere.
4. **Shareable packages** — Publish a GitHub repo of automations. Teams add it with one command.

### Code example section
Show a realistic example that's interesting enough to make a developer think "oh, I have something like that." The docker workflow is a good choice since most developers have it.

```yaml
# .pi/docker/up.yaml
description: Start the development environment
steps:
  - bash: docker-compose up -d
    description: Start containers

  - bash: ./scripts/wait-for-db.sh
    description: Wait for the database to be ready

  - bash: echo "Dev environment ready at http://localhost:3000"
```

```yaml
# pi.yaml (root config)
project: my-app

shortcuts:
  up:   docker/up
  down: docker/down
  logs: docker/logs
```

Then the terminal showing:
```
$ pi shell   # install shortcuts once
$ up         # works from anywhere, forever
```

### Styling
Use Tailwind CSS classes (Astro Starlight includes it). Keep the palette neutral (use CSS vars from Starlight's theme so dark mode works automatically). Do not introduce a custom color scheme — match what Starlight uses for its own brand sections.

For the code blocks, use Starlight's `<Code>` component or a `<pre>` with syntax highlighting via Shiki (Astro's built-in).

### Copy-to-clipboard for install command
Implement a small inline `<button>` that copies `brew install pi` to clipboard on click. Use a minimal `onclick` script or a tiny Astro island — no React needed:
```html
<button onclick="navigator.clipboard.writeText('brew install pi').then(() => this.textContent = 'Copied!')">
  Copy
</button>
```

### The `src/pages/index.astro` structure
Import from Starlight's `Head`, `Header`, and `Footer` components so the navbar and dark mode toggle are present. Or use a full custom layout — either works. The key is that `/docs/` links go to the Starlight docs section and `/` is the marketing page.

## Subtasks
- [ ] Create `src/pages/index.astro` with hero section
- [ ] Add feature tiles section
- [ ] Add code example section (YAML + terminal output)
- [ ] Add copy-to-clipboard for install command
- [ ] Add footer
- [ ] Verify dark mode
- [ ] Verify mobile layout

## Blocked By
79-website-scaffold-and-ci (need the Astro project to exist first)
