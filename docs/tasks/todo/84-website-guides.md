# Website: Guides

## Type
feature

## Status
todo

## Priority
high

## Project
14-documentation-website

## Description
Write six practical guides that walk through real workflows. Guides differ from Concepts: they are task-oriented ("how do I do X") rather than concept-oriented ("what is X"). Each guide should read like a short tutorial — intro, step-by-step, expected outcomes. The packaging guides (publishing to GitHub and using packages) are especially important and should be detailed enough that a developer can follow them start to finish with zero prior knowledge.

## Acceptance Criteria
- [ ] `setup-automations.md` walks through writing a realistic `pi setup` configuration for a project with mixed OS support
- [ ] `cross-platform-scripts.md` walks through using `if:`, `first:`, and OS predicates to write automations that work on macOS and Linux
- [ ] `publishing-to-github.md` walks through creating a package repo, organizing it, tagging a release, and making it consumable — complete from `git init` to `pi add`
- [ ] `using-packages.md` walks through adding a package, using it in automations, upgrading versions, and understanding the alias system
- [ ] `private-repos.md` covers the SSH and GITHUB_TOKEN approaches with concrete setup steps
- [ ] `parent-shell-steps.md` explains the problem, the solution, and all the ways `parent_shell: true` works (shortcuts, global wrapper, raw binary)
- [ ] Every guide has a clear "What you'll learn" intro and a "Summary" at the end
- [ ] Every code block is complete and runnable

## Implementation Notes

### setup-automations.md

**What you'll learn:** Write a `pi setup` configuration that installs dependencies on any developer's machine, works on macOS and Linux, checks before installing, and leaves shortcuts ready in the terminal.

**Walkthrough:**

1. Anatomy of a setup entry — bare string vs object form. Explain: bare string = no conditions, no inputs; object form = use when you need `if:` or `with:`.

2. Build a realistic setup sequence step by step:
   ```yaml
   setup:
     - pi:install-homebrew          # macOS only (automation has its own if: os.macos)
     - run: pi:install-python
       with:
         version: "3.13"
     - run: pi:install-node
       with:
         version: "20"
     - setup/install-project-deps   # custom automation: npm install, pip install, etc.
     - run: setup/configure-git-hooks
       if: dir.exists(".git")
   ```

3. Writing an idempotent custom setup automation:
   ```yaml
   # .pi/setup/install-project-deps.yaml
   description: Install all project dependencies
   steps:
     - bash: npm ci
       dir: frontend
     - bash: pip install -r requirements.txt
       dir: backend
   ```

4. The `pi-setup-<project>` helper: after running `pi shell`, the generated `pi-setup-myproject` function runs `pi setup` AND sources the shell immediately, so shortcuts are available without opening a new terminal.

5. Testing your setup on a clean machine: mention Docker for CI-style testing, or just running on a colleague's machine.

---

### cross-platform-scripts.md

**What you'll learn:** Write automations that behave correctly on macOS and Linux without duplicating logic.

**Walkthrough:**

1. The `if:` predicate on steps: skip platform-specific steps. Show installing a tool with platform-specific package managers:
   ```yaml
   steps:
     - bash: brew install jq
       if: os.macos and not command.jq
     - bash: apt-get install -y jq
       if: os.linux and not command.jq
   ```

2. The `first:` block — the cleaner alternative when the logic is "try A, then B, then C":
   ```yaml
   steps:
     - first:
         - bash: mise install python@3.13 && mise use python@3.13
           if: command.mise
         - bash: brew install python@3.13
           if: command.brew
         - bash: |
             echo "No supported installer found (tried mise, brew)" >&2
             exit 1
   ```
   Explain: no need for `and not command.mise` in the brew condition — `first:` stops at the first match.

3. Architecture-specific builds:
   ```yaml
   steps:
     - bash: go build -o dist/app-arm64 ./...
       if: os.arch.arm64
       env:
         GOARCH: arm64
     - bash: go build -o dist/app-amd64 ./...
       if: os.arch.amd64
       env:
         GOARCH: amd64
   ```

4. Checking for optional tools:
   ```yaml
   steps:
     - bash: prettier --write .
       if: command.prettier
     - bash: echo "Prettier not installed — skipping formatting"
       if: not command.prettier
       silent: true
   ```

5. Tips: use `command.<name>` to check for tools before using them; prefer `first:` over long `and not` chains; keep each step's condition positive when possible.

---

### publishing-to-github.md

**What you'll learn:** Create a package repository on GitHub, organize it, tag a release, and make it available for others to use with `pi add`.

This is the "author" guide. A detailed, linear walkthrough.

**Walkthrough:**

#### 1. Create the package repo

```bash
mkdir my-pi-package
cd my-pi-package
git init
git remote add origin git@github.com:your-org/my-pi-package.git
```

#### 2. Create the `.pi/` folder and write your first automation

Structure advice: organize by tool or lifecycle phase. Be explicit about what namespacing does:
```
.pi/
  docker/
    up.yaml
    down.yaml
    logs.yaml
  setup/
    install-deps.yaml
    configure-hooks.yaml
```

A user who adds this package and writes `run: docker/up` will get `.pi/docker/up.yaml`. If they have a local `.pi/docker/up.yaml`, that wins (local always takes priority).

#### 3. The optional `pi-package.yaml`

Add one if you want to enforce a minimum PI version:
```yaml
min_pi_version: "0.3.0"
```

If absent, any PI version can use the package.

#### 4. Commit and tag a release

```bash
git add .
git commit -m "initial package release"
git tag v1.0
git push origin main
git push origin v1.0
```

The version tag is what consumers reference. Without a tag, there's no stable version to pin.

#### 5. Consumers add it

```bash
pi add your-org/my-pi-package@v1.0
```

This writes to their `pi.yaml`:
```yaml
packages:
  - your-org/my-pi-package@v1.0
```

And immediately fetches the package into `~/.pi/cache/`.

#### 6. Releasing updates

When you make changes:
```bash
git add .
git commit -m "add setup/install-deps automation"
git tag v1.1
git push origin main
git push origin v1.1
```

Consumers upgrade by editing `pi.yaml` (change `@v1.0` to `@v1.1`) or running:
```bash
pi add your-org/my-pi-package@v1.1
```

Note: `pi add` with a new version appends a new entry — the old one should be removed manually (for now — `pi update` is planned).

#### 7. Naming conventions

Recommendations:
- Use `kebab-case` for repo names: `my-org/pi-docker-utils`
- Group by domain in the `.pi/` folder: `docker/`, `setup/`, `deploy/`
- Prefix the package name with `pi-` so it's searchable on GitHub

#### 8. Making automations composable

Advice on design: write small, single-purpose automations. Let callers compose with `run:`. Avoid monolithic scripts. Use `inputs:` for parameterization. Write idempotent automations (check before acting).

---

### using-packages.md

**What you'll learn:** Add a package to your project, reference its automations, use aliases, and understand how packages interact with local automations.

**Walkthrough:**

1. **Adding a package**:
   ```bash
   pi add yotam180/pi-common@v1.2
   ```
   What happens: validates the ref, fetches into `~/.pi/cache/yotam180/pi-common@v1.2/`, writes the entry to `pi.yaml`.

2. **Using it directly** — no alias, full reference:
   ```yaml
   steps:
     - run: yotam180/pi-common@v1.2/docker/up
   ```
   This works but is verbose. Better: declare in `packages:` and use the short name.

3. **Using with just the package short name** — after declaring in `packages:`, automations that don't collide with local names are accessible by their plain name:
   ```bash
   pi run docker/up    # runs from the package if no local .pi/docker/up.yaml
   ```

4. **Using an alias** — cleaner, unambiguous:
   ```yaml
   packages:
     - source: yotam180/pi-common@v1.2
       as: common
   ```
   ```yaml
   steps:
     - run: common/docker/up
   ```

5. **What happens on `pi setup`** — packages declared in `packages:` are fetched before setup automations run. Status output shows fetch state.

6. **Browsing available automations**:
   ```bash
   pi list          # local + packages, with SOURCE column
   pi list --all    # grouped by package source
   ```

7. **Collision behavior** — local `.pi/` always wins. If you have `.pi/docker/up.yaml` and a package also has `docker/up`, your local version runs and a warning is printed.

8. **On-demand references** — you can reference a package inline in a `run:` step without declaring it:
   ```yaml
   run: org/repo@v1.0/docker/up
   ```
   PI fetches it automatically (once, then cached) and prints an advisory to add it to `pi.yaml` for reproducible builds.

9. **Upgrading** — change the version tag in `pi.yaml` and run `pi setup` or any `pi run` (PI fetches the new version automatically).

---

### private-repos.md

**What you'll learn:** Use PI packages from private GitHub repositories.

**Walkthrough:**

1. **How PI authenticates** — PI tries three methods in order:
   1. SSH (`git@github.com:org/repo.git`) — works if your SSH key is configured for GitHub
   2. HTTPS with token (`GITHUB_TOKEN` env var)
   3. Plain HTTPS — public repos only

2. **SSH setup** (recommended for developer machines):
   ```bash
   ssh-keygen -t ed25519 -C "your-email@example.com"
   # Add the public key to GitHub → Settings → SSH and GPG keys
   ssh -T git@github.com     # verify it works
   ```
   After that, `pi add` and `pi setup` work transparently with private repos.

3. **GITHUB_TOKEN setup** (recommended for CI):
   ```bash
   export GITHUB_TOKEN="ghp_your_token_here"
   ```
   The token needs `repo` read access (or `contents: read` for fine-grained tokens). In GitHub Actions, `GITHUB_TOKEN` is automatically available — set it in the workflow env:
   ```yaml
   - name: Setup
     run: pi setup
     env:
       GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```

4. **Error messages** — if all methods fail, PI prints a clear error with instructions. Show the error and map each fix.

5. **Testing your setup** — verify SSH with `ssh -T git@github.com`, verify token with `curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user`.

---

### parent-shell-steps.md

**What you'll learn:** Use `parent_shell: true` to run bash commands that must affect the calling shell — like activating a virtualenv or changing directory.

**The problem:**

When PI runs a step as a subprocess, any changes the subprocess makes to its environment (like `cd`, `export`, or `source venv/bin/activate`) are invisible to the parent shell. The process exits, the parent shell is unchanged.

This is the fundamental limitation of subprocess-based automation tools.

**The solution:**

PI's `parent_shell: true` flag tells PI: "don't execute this step yourself — write it to a file, and after PI exits, have the shell run it directly."

```yaml
steps:
  - bash: python3 -m venv .venv
    description: Create virtualenv if needed
  - bash: source .venv/bin/activate
    parent_shell: true
    description: Activate virtualenv in the current shell
```

After PI exits, the calling shell `source`s the temp file containing `source .venv/bin/activate`. Your terminal is now in the virtualenv.

**How it works under the hood:**

1. When PI starts (via a shell shortcut or the global `pi()` wrapper), a temp file path is passed in `PI_PARENT_EVAL_FILE`.
2. When PI encounters a `parent_shell: true` step, it appends the command to that file instead of running it.
3. When PI exits, the wrapper `eval`s the file.

**Requirements:**

- Only valid on `bash:` steps. Cannot be used with `python:`, `typescript:`, or `run:`.
- Cannot combine with `pipe: true`.
- Requires `PI_PARENT_EVAL_FILE` to be set. This happens automatically when:
  - Running via a shell shortcut (installed by `pi shell`)
  - Running via the global `pi()` wrapper (also installed by `pi shell`)
- If `PI_PARENT_EVAL_FILE` is not set (e.g., calling the raw `pi` binary directly), the step is skipped with a warning: "run `pi shell` to install shell integration"

**Installing shell integration:**

```bash
pi shell    # installs shortcuts AND the global pi() wrapper
```

After this, every `pi run` call (via the wrapper) will respect `parent_shell: true` steps.

**Common use cases:**

```yaml
# Activate a Python virtualenv
- bash: source .venv/bin/activate
  parent_shell: true

# Change directory persistently
- bash: cd /path/to/service
  parent_shell: true

# Load env vars from a file
- bash: source .env
  parent_shell: true

# Use nvm to switch Node versions
- bash: nvm use 20
  parent_shell: true
```

**What doesn't work:**

- `parent_shell: true` does not support capturing output or piping. It's fire-and-forget into the parent shell.
- Do not use it for steps that should NOT affect the parent shell — use it only when the parent shell state is the actual goal.

## Subtasks
- [ ] Write `setup-automations.md`
- [ ] Write `cross-platform-scripts.md`
- [ ] Write `publishing-to-github.md`
- [ ] Write `using-packages.md`
- [ ] Write `private-repos.md`
- [ ] Write `parent-shell-steps.md`
- [ ] Verify all code examples
- [ ] Add cross-links to Concepts and Reference pages

## Blocked By
79-website-scaffold-and-ci
