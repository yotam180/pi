# Website: Getting Started Documentation

## Type
feature

## Status
todo

## Priority
high

## Project
14-documentation-website

## Description
Write the three Getting Started pages: Introduction, Installation, and Quick Start. These are the most-read pages on any documentation site and must be excellent. The Quick Start has a strict pedagogical rule: **introduce the short-form syntax first, never show `steps:` on the first page**. The developer runs their first automation in under 5 minutes. Full forms and deeper explanations belong in Concepts, not here.

## Acceptance Criteria
- [ ] `introduction.md` explains what PI is, what problem it solves, and who it's for — in under 300 words
- [ ] `installation.md` covers Homebrew, curl/shell install, and manual binary download; includes a "verify install" step
- [ ] `quick-start.md` starts with the absolute minimal form (`bash: <cmd>` in a yaml file + `pi run`) and only introduces `steps:`, `pi.yaml`, and shortcuts when there is a concrete reason to need them
- [ ] Quick Start uses a "callout" box that explains the short form is syntactic sugar for the full `steps:` form — after showing it working
- [ ] Every code block in Getting Started is complete and runnable (no placeholders like `<your-project>` in the first examples)
- [ ] Each page links forward to the relevant Concepts page for deeper reading
- [ ] No page exceeds ~600 words (getting started should be fast to read)

## Implementation Notes

### introduction.md

**Structure:**
1. One-sentence definition: "PI is a developer automation tool that replaces your team's `shell_shortcuts.sh` with structured, polyglot, version-controlled automations."
2. The problem in 3 bullet points (from `docs/README.md`)
3. What PI does differently — brief, one paragraph
4. A tiny teaser example (just enough to see the shape)
5. "Next: Installation →"

**Do not** write a long background section. The problem statement should be 3 bullets, not 3 paragraphs.

**Teaser example** (only a taste — enough to intrigue, not explain):
```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```
```bash
$ pi run test
  → bash: go test ./...
ok  github.com/example/myapp  0.423s
```

---

### installation.md

**Structure:**
1. Homebrew (primary, macOS/Linux):
   ```bash
   brew install pi
   ```
2. Shell installer (macOS/Linux, no Homebrew):
   ```bash
   curl -fsSL https://getpi.dev/install | sh
   ```
3. Manual download (link to GitHub Releases for all platforms/arches)
4. Verify installation:
   ```bash
   pi version
   ```
5. Next steps callout: "→ Follow the Quick Start to run your first automation"

**Note on Windows:** Note that Windows support is limited (bash/python/typescript steps require WSL or similar). Link to GitHub issues if this is tracked.

---

### quick-start.md

This is the most carefully designed page in the whole site. Follow this exact narrative arc:

#### Step 1 — Your first automation (short form, no ceremony)

Tell the developer to create a file. Show the shortest possible valid automation:

```yaml
# .pi/greet.yaml
description: Say hello
bash: echo "Hello from PI!"
```

Then run it:
```bash
pi run greet
```

Expected output:
```
  → bash: echo "Hello from PI!"
Hello from PI!
```

**No `pi.yaml`. No `steps:`. No installs.** Just a file in `.pi/` and a command.

#### Step 2 — Something useful

Replace the toy example with a real one. Use a test runner since almost every project has one:

```yaml
# .pi/test.yaml
description: Run the test suite
bash: go test ./...
```

```bash
pi run test
```

At this point, note: "Every `.yaml` file in `.pi/` becomes an automation. The name is derived from the file path: `.pi/test.yaml` → `test`, `.pi/docker/up.yaml` → `docker/up`."

#### Step 3 — A shortcut you can type from anywhere

Now introduce `pi.yaml`, but only because there's a concrete reason: making the automation a shell command.

```yaml
# pi.yaml  (at your repo root)
project: my-app

shortcuts:
  test: test
```

```bash
pi shell         # install shortcuts into your shell (once per project)
source ~/.zshrc  # or open a new terminal

test             # runs from anywhere — no cd, no pi run
```

Show that this works from any directory. Explain `pi shell` in one sentence: it writes a shell function into your `.zshrc` (or `.bashrc`) that navigates to the project root and runs the automation.

#### Step 4 — Chaining steps (introducing `steps:` for the first time)

Now there's a concrete reason to need `steps:`: doing two things in sequence.

```yaml
# .pi/build-and-test.yaml
description: Build then run tests
steps:
  - bash: go build ./...
  - bash: go test ./...
```

```bash
pi run build-and-test
```

**Now** reveal the magic in a callout box:

> **Short form and full form are identical**
>
> The `bash:` key you used in the earlier examples is shorthand for a single-step automation. These two files are exactly equivalent:
>
> ```yaml
> # Short form
> bash: go test ./...
> ```
>
> ```yaml
> # Full form (same behavior)
> steps:
>   - bash: go test ./...
> ```
>
> Use the short form when there's only one step. Use `steps:` when you need more than one.

#### Step 5 — Where to go next

End with a "What's next" section:
- **[Concepts: Automations →]** — learn about step types, env vars, working directories, conditions
- **[Concepts: pi.yaml →]** — learn about `setup:`, `packages:`, and all root config options
- **[Concepts: Shell Shortcuts →]** — deeper look at `pi shell`, the global wrapper, and `anywhere: true`
- **[Guides: Setup Automations →]** — `pi setup` for onboarding teammates

## Subtasks
- [ ] Write `introduction.md`
- [ ] Write `installation.md`
- [ ] Write `quick-start.md` (Steps 1–5 as outlined above)
- [ ] Verify all code examples are accurate against `docs/README.md`
- [ ] Add "next page" links at the bottom of each page

## Blocked By
79-website-scaffold-and-ci
