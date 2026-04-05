# Website: Getting Started Documentation

## Type
feature

## Status
done

## Priority
high

## Project
14-documentation-website

## Description
Write the three Getting Started pages: Introduction, Installation, and Quick Start. These are the most-read pages on any documentation site and must be excellent. The Quick Start has a strict pedagogical rule: **introduce the short-form syntax first, never show `steps:` on the first page**. The developer runs their first automation in under 5 minutes. Full forms and deeper explanations belong in Concepts, not here.

## Acceptance Criteria
- [x] `introduction.md` explains what PI is, what problem it solves, and who it's for — in under 300 words
- [x] `installation.md` covers Homebrew and manual binary download; includes a "verify install" step
- [x] `quick-start.md` starts with the absolute minimal form (`bash: <cmd>` in a yaml file + `pi run`) and only introduces `steps:`, `pi.yaml`, and shortcuts when there is a concrete reason to need them
- [x] Quick Start uses a "callout" box that explains the short form is syntactic sugar for the full `steps:` form — after showing it working
- [x] Every code block in Getting Started is complete and runnable (no placeholders like `<your-project>` in the first examples)
- [x] Each page links forward to the relevant Concepts page for deeper reading
- [x] No page exceeds ~600 words (getting started should be fast to read)

## Implementation Notes

### Decisions
- **No curl install script**: The task template mentions a `curl -fsSL https://getpi.dev/install | sh` option, but this doesn't exist yet (no `getpi.dev` domain or install script). Installation covers Homebrew (primary) and manual binary download from GitHub Releases. A curl install method can be added when the infrastructure exists.
- **Starlight callout syntax**: Used `:::tip[...]` for the short-form/full-form equivalence callout in Quick Start, which is Starlight's native admonition syntax.
- **Starlight note syntax**: Used `:::note[Windows]` in Installation for the Windows caveat.
- **Forward links**: Each page ends with a "Next" link to the next page in sequence. Quick Start ends with a "What's Next" section linking to four concept/guide pages.
- **Word counts**: introduction (245), installation (182), quick-start (454) — all well under the 600-word cap.
- **Narrative arc in Quick Start**: Follows the exact 5-step arc from the task spec — greet → test → shortcuts → chaining steps → what's next. The `steps:` keyword only appears in Step 4, never before.

### Files modified
- `website/src/content/docs/getting-started/introduction.md` — full content
- `website/src/content/docs/getting-started/installation.md` — full content
- `website/src/content/docs/getting-started/quick-start.md` — full content

## Subtasks
- [x] Write `introduction.md`
- [x] Write `installation.md`
- [x] Write `quick-start.md` (Steps 1–5 as outlined above)
- [x] Verify all code examples are accurate against `docs/README.md`
- [x] Add "next page" links at the bottom of each page

## Blocked By
79-website-scaffold-and-ci (done)
