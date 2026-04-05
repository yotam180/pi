# Website: Concepts Documentation

## Type
feature

## Status
todo

## Priority
high

## Project
14-documentation-website

## Description
Write the five Concepts pages that explain PI's mental model in depth. These pages go deeper than Getting Started — they cover every feature, every modifier, every behavioral detail. A developer who has finished Getting Started should be able to read the Concepts pages and understand *everything* PI can do, without needing to read the Reference. Reference is for looking things up; Concepts is for understanding.

Source of truth: `docs/README.md`. Do not invent behavior. Every claim must be verifiable against that document.

## Acceptance Criteria
- [ ] `automations.md` covers: what an automation is, path-based naming, single-step shorthand (and its equivalence to full form), multi-step, all step types, all step modifiers (env, dir, timeout, silent, if, pipe, description), first: blocks, install: blocks — with examples for all
- [ ] `pi-yaml.md` covers: file location, `project:`, `shortcuts:` (with and without `anywhere: true`), `setup:` (bare strings vs object form, if: on setup entries), `packages:` (all source types: GitHub, file:, aliases)
- [ ] `step-types.md` covers: bash (inline and file), python (inline and file), typescript (inline and file, via tsx), run: (local, package, on-demand GitHub), and what each step type actually does at runtime
- [ ] `shell-shortcuts.md` covers: what `pi shell` does, the generated function shape, the global `pi()` wrapper, `pi-setup-<project>`, `pi shell uninstall`, `pi shell list`, `anywhere: true`, how parent_shell: steps interact with shortcuts
- [ ] `packages.md` covers: the full package story from consumer side and author side, GitHub packages, file: sources, aliases, `pi add`, version pinning, mutable refs and the reproducibility warning, `pi-package.yaml`, on-demand fetching and the advisory message, private repos overview (link to guide)
- [ ] Each page has an "In this section" summary at the top listing what's covered
- [ ] Each page links to the Reference for the exhaustive field-by-field listing
- [ ] All code examples are accurate

## Implementation Notes

### automations.md

**Narrative flow:**

1. **What is an automation?** — A YAML file in `.pi/`. Each file = one named, runnable unit of work. Name is derived from file path, not a `name:` field.

2. **Naming** — Explain the path-to-name derivation with examples:
   - `.pi/test.yaml` → `test`
   - `.pi/docker/up.yaml` → `docker/up`
   - `.pi/setup/install-deps/automation.yaml` → `setup/install-deps`
   
   Mention the deprecated `name:` field and what happens if it mismatches.

3. **Single-step shorthand** — The most common form. Show it, then show the equivalent `steps:` form, and state clearly: "These are identical. PI has no preference — use the short form when there's one step, `steps:` when there are more."

4. **Multi-step automations** — Show a `steps:` automation. Explain sequential execution, exit-on-failure behavior.

5. **Step modifiers** — Go through each one with a concrete use case:
   - `env:` — per-step env vars (and automation-level env:). Scope: does not leak to next step.
   - `dir:` — override working directory. Resolved relative to project root.
   - `timeout:` — duration format (30s, 5m, 1h30m). Exit code 124 on timeout.
   - `silent: true` — suppresses trace line and output. `--loud` overrides.
   - `if:` — skip step based on condition. Link to Conditions reference.
   - `pipe: true` — stdout of this step becomes stdin of next step.
   - `description:` — human-readable label shown by `pi info`. No effect on execution.
   - `parent_shell: true` — runs in calling shell (see Shell Shortcuts page for details).

6. **Step trace lines** — What PI prints before each step by default: `→ bash: go test ./...`

7. **`first:` blocks** — Mutual exclusion: only the first sub-step whose `if:` matches runs. Show the "which package manager do I have?" pattern. Explain why this is better than compounding `and not` chains.

8. **`install:` block** — For tool installation. Declare `test:`, `run:`, `version:` — PI manages the status output (✓ already installed / → installing... / ✗ failed). Show the `first:` pattern inside `install.run:` for multiple installer paths. Mention the status output format.

9. **Automation-level `if:`** — Skip the whole automation based on a condition. Different from step-level `if:`.

10. **Inputs** — Brief mention that automations can declare inputs. Link to Reference for the full schema.

---

### pi-yaml.md

**Narrative flow:**

1. **What is `pi.yaml`?** — The root config file. Lives at the repo root. It is small by design — it only declares *which* automations are exposed as shortcuts, *which* run during setup, and *where* to find external automations. It does not define automations itself.

2. **`project:`** — The project name. Used in generated shell function names.

3. **`shortcuts:`** — Show a minimal example (2–3 shortcuts). Explain that the key is the shell command name and the value is the automation name. Then introduce the object form with `anywhere: true`: what it does (runs `pi run` without changing directory first), and when to use it (global tools, deploy commands).

4. **`setup:`** — The list of automations `pi setup` runs. Show both forms (bare string = no modifiers, object form = has `if:` or `with:`). Walk through a realistic setup list with comments. Explain idempotency expectation.

5. **`packages:`** — Overview of all source types. Show each form. Explain `as:` alias. Note resolution order (local > packages > builtins). Explain `pi add` as the ergonomic way to write this section. Link to the Packages concept page for the full story.

---

### step-types.md

**Narrative flow:**

For each step type, cover: what it is, how to use it (inline vs file reference), what happens at runtime (what PI actually does to execute it).

1. **`bash:`** — Inline shell or path to a `.sh` file. Runs via `/bin/sh`. Show both:
   ```yaml
   bash: echo "hello"          # inline
   bash: scripts/setup.sh      # file path (relative to automation file)
   ```
   Note: `parent_shell: true` is only valid on bash steps.

2. **`python:`** — Inline script or path to a `.py` file. Runs via `python3`. The `.py` file path is relative to the automation file (important: next to the `.yaml`, not the project root). Show both forms.

3. **`typescript:`** — Inline TypeScript or path to a `.ts` file. Runs via `tsx`. Requires `tsx` to be installed. Show both forms. Note that `pi:install-tsx` can install it.

4. **`run:`** — Call another automation by name. Show local, package (via alias), and on-demand GitHub forms:
   ```yaml
   run: docker/up                           # local
   run: mytools/docker/up                   # package alias
   run: org/repo@v1.0/docker/up             # on-demand GitHub
   ```
   Explain what `run:` does NOT do: it does not inherit the caller's `env:`, `dir:`, or `timeout:`. Each `run:` target is an independent automation execution.

5. **File paths vs inline** — Explain the rule: if the value ends in `.sh`, `.py`, `.ts` it's a file path; otherwise it's inline. File paths are relative to the automation YAML file (not the project root). This lets assets live next to their automation.

---

### shell-shortcuts.md

**Narrative flow:**

1. **What `pi shell` does** — Writes shortcut functions into `~/.zshrc` (or `~/.bashrc` for bash users). Each shortcut is a shell function, not an alias. Explain the function shape briefly (cd + PI_PARENT_EVAL_FILE pattern) — the full generated code is in the Reference.

2. **Running shortcuts** — After `pi shell` + source, any shortcut key from `pi.yaml` is available as a terminal command from any directory.

3. **`anywhere: true`** — By default, a shortcut `cd`s to the project root before running. With `anywhere: true`, it doesn't. When to use it: automations that operate on the current directory or work globally (e.g., a deploy command that needs to know where you are).

4. **The global `pi()` wrapper** — `pi shell` also installs a global `pi()` function that wraps every `pi` invocation. This is what makes `parent_shell: true` steps work for *all* `pi run` calls — not just shortcuts. Explain: without the wrapper, calling the `pi` binary directly can't affect the parent shell. The wrapper creates a temp file, passes it as `PI_PARENT_EVAL_FILE`, and `eval`s it after `pi` exits.

5. **`pi-setup-<project>`** — The helper function installed by `pi shell` that wraps `pi setup` with the eval pattern and auto-sources the shell after setup completes. Explain why: `pi setup` installs shortcuts, but without re-sourcing, they're not available yet. This helper makes it seamless — run `pi-setup-myproject` once, shortcuts are ready immediately.

6. **`pi shell uninstall`** — Removes shortcuts for the current project from the shell config. Does not remove the global `pi()` wrapper.

7. **`pi shell list`** — Shows all installed shortcut files across all projects.

8. **`parent_shell: true` interaction** — Brief explanation of why some steps need to run in the parent shell (virtualenv activation, `cd` that persists). Direct the reader to the Parent Shell Steps guide for the full story.

---

### packages.md

**Narrative flow:**

1. **Why packages?** — The problem: copy-pasting automations between repos doesn't scale. With packages, a team maintains one repo of shared automations and every project consumes it. Changes propagate when you upgrade the version pin.

2. **Source types** — Introduce the three source types, each with a short example in `packages:`:
   - GitHub: `yotam180/pi-common@v1.2` — fetched once, cached, pinned
   - File (absolute): `file:~/my-automations` — read directly from disk, no cache
   - File (relative): `file:./local-package` — relative to project root

3. **Declaring packages** — Show a complete `packages:` block in `pi.yaml`. Show both the simple form and the form with `as:` alias. Explain what the alias does: lets you write `run: mytools/docker/up` instead of the full path.

4. **`pi add`** — The ergonomic way to add a package. Show all three forms:
   ```bash
   pi add yotam180/pi-common@v1.2
   pi add file:~/shared-automations
   pi add file:~/my-automations --as mytools
   ```
   Note: adding the same source twice is a no-op.

5. **On-demand fetching** — If you write `run: org/repo@v1.0/docker/up` without declaring the package, PI fetches it automatically and shows an advisory with the snippet to add to `pi.yaml`. Cached afterwards.

6. **Version pinning** — The `@version` part is required for GitHub packages. Mutable refs like `@main` work but emit a reproducibility warning. Explain why: `@main` today ≠ `@main` in six months. Use release tags for stability.

7. **Running package automations** — Show how `pi list`, `pi list --all`, and `pi run` work with packages. Show the SOURCE column in `pi list`.

8. **Writing a package** — The author side. A package is just a GitHub repo with a `.pi/` folder. No registry. No special tooling. Steps:
   1. Create a GitHub repo
   2. Add a `.pi/` folder with your automations
   3. Tag a release: `git tag v1.0 && git push origin v1.0`
   4. Users consume it with `pi add your-org/your-repo@v1.0`
   
   Optional: add `pi-package.yaml` with `min_pi_version:` to enforce a minimum PI version.

9. **Private repos** — Brief overview (SSH / GITHUB_TOKEN). Link to the Private Repos guide for full details.

## Subtasks
- [ ] Write `automations.md`
- [ ] Write `pi-yaml.md`
- [ ] Write `step-types.md`
- [ ] Write `shell-shortcuts.md`
- [ ] Write `packages.md`
- [ ] Verify all examples against `docs/README.md`
- [ ] Add cross-links between concept pages and to Reference pages

## Blocked By
79-website-scaffold-and-ci
