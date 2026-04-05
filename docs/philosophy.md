# PI — Product Philosophy

This document defines PI's core product values. Every user-facing feature, command, error message, output line, and builtin automation must be evaluated against these principles before shipping.

**Agents:** read this before designing or implementing anything a developer will touch.

---

## The product promise

PI replaces `setup_environment.sh` and `shell_shortcuts.sh`. These files exist in almost every engineering repo. They grow to 1000 lines. They're written by the person who set up the repo, poorly documented, and terrifying to touch. New developers copy commands from Slack instead of reading them.

PI's promise: a developer clones a repo, runs `pi setup`, and is productive. No Slack messages. No shell archaeology. No manual steps. The repo documents itself, and the documentation runs.

Every product decision is evaluated against: does this get someone from `git clone` to productive faster, with more confidence, and with less chance of making a mistake?

---

## Principles

### 1. The first two minutes are sacred

A developer's first experience with PI must work, completely, with no prior knowledge.

That means the Quick Start shows `bash: go test ./...` — a single file, a single command — before it shows anything else. Not `steps:`. Not `pi.yaml`. Not shortcuts. Those come later, when there's a reason for them.

**The rule:** every entry point — the first page of docs, the output of `pi init`, the first error message a new user sees — must be designed for someone who doesn't know what PI is. The path from zero to first successful `pi run` should take under two minutes and require zero googling.

---

### 2. Correct the action, not the developer

When a developer types `pi setup add python`, they mean `pi:install-python`. PI should do that — not return an error saying "unknown automation 'python'."

When they type `pi run dokcer/up`, PI should suggest `docker/up` — not print a list of all automations and leave them guessing.

The difference between a good tool and an annoying one is often this: does it meet the developer where they are, or does it demand they come to it?

**The rule:** if PI can determine the developer's intent with high confidence, it takes the action and prints what it did. If there's genuine ambiguity (multiple plausible matches, a destructive action), it confirms. It never returns a bare error when a corrective action exists.

This applies to:
- Automation name fuzzy matching (`dokcer/up` → suggest `docker/up`)
- Short-form builtin resolution (`python` → `pi:install-python`, `pi:python` → `pi:install-python`)
- Missing `pi.yaml` when running `pi setup add` → offer to init the project first
- Already-done operations (`pi setup add uv` when uv is already in setup:) → "already there, nothing to do"

---

### 3. Short form is the real form

`bash: go test ./...` is not shorthand for `steps: [{bash: go test ./...}]`. It is the real form. `steps:` is the expanded form for when you need it.

This is not just about syntax. It's about what PI optimizes for. The thing developers write a hundred times should be as short as possible. The thing they write once (a complex multi-step automation) can afford a bit more ceremony.

PI should never make a developer write more than they need to. If they need to add a version to their Python install, they type `--version 3.13`, not `with:\n  version: "3.13"` in the terminal. PI generates the YAML.

**The rule:** in commands, in docs, in error messages, in generated output — always show the minimal form first. Only introduce more syntax when the developer has a concrete reason to need it. Never start with the most general case.

---

### 4. Guess correctly, every time

If PI's builtins are named consistently enough, a developer who knows `pi:install-python` can guess `pi:install-node`, `pi:install-terraform`, `pi:install-kubectl` without looking anything up. The pattern is the documentation.

This is why consistency is not a nice-to-have — it's a product feature. A developer who can guess the command correctly the first time will trust PI. A developer who has to check the docs for every new tool will not.

**The rule:** all installers use `version` as the input name. All installers use the `install:` block. All installers try `mise` first, then platform-specific fallbacks. All installers produce the same `✓ / → / ✗` status output. No exceptions without a written reason in `docs/architecture.md`.

The naming convention: `pi:install-<tool>` for the canonical name. A short alias `pi:<tool>` resolves to it. Both work everywhere.

---

### 5. The repo is the documentation, and the documentation runs

`.pi/` is committed. `pi.yaml` is committed. A new developer clones the repo, looks in `.pi/`, and sees exactly how the team works — what commands they run, what setup steps they need, what shortcuts they use. And then they can run it.

This is PI's biggest idea. A `setup_environment.sh` documents what you should do and also tries to do it. But it's written for one machine, by one person, tested once. PI's automation model is designed so that setup steps are:
- Written declaratively (YAML, not bash spaghetti)
- Idempotent (safe to run again after pulling changes)
- Composable (each step does one thing; chain them to do more)
- Testable (the `test:` phase in `install:` blocks proves the step ran correctly)

**The rule:** every setup step must be safe to run multiple times. `test:` before `run:`. Check before writing. If a file-writing operation (`.zshrc`, `pi.yaml`, `.git/hooks/`) doesn't check first, it's a bug.

---

### 6. Explain the magic, don't hide it

PI does things automatically that used to require manual steps. When it does, it tells you what it did and why. Not because developers can't handle automation — but because developers who understand their tools trust them more and debug them better.

When PI installs a shortcut, it shows you the generated function. When it runs a `parent_shell:` step, it shows you the trace `→ parent: source .venv/bin/activate`. When it expands `python` to `pi:install-python`, it prints `Resolved 'python' → pi:install-python`.

The website docs follow the same principle: the Quick Start shows the short form working, then immediately explains what it expanded to. "This is equivalent to..." is one of the most useful sentences in a tech doc.

**The rule:** never hide what PI is doing. Trace lines before every step. Advisory messages when implicit actions happen. The "what's next" message after every init/add/install. If a developer is confused about why something happened, the answer should be findable in the terminal output, not just in the docs.

---

### 7. One thing, done completely

Each builtin does one thing. `pi:install-python` installs Python. It doesn't create a virtualenv, configure pip, or install packages. Those are separate automations.

This is how you get a library of automations that compose without surprising interactions. A developer who wants Python + a virtualenv + project deps writes three automations (or uses `pi setup add` three times). Each step is independently understandable, independently testable, independently replaceable.

The flip side: "one thing" means doing it completely. `pi:install-python` handles `mise`, `brew`, and a clear fallback error. It provides a `version:` output. It works on macOS and Linux. It's not a thin wrapper that works 80% of the time.

**The rule:** each builtin's description is a complete sentence explaining what it does, not just its name repeated. If a builtin needs to do two things, it calls two automations via `run:` steps.

---

### 8. PI is a team tool, not a personal tool

PI is not `~/.aliases`. It's `.pi/` — checked into the repo, shared with the team, reviewed in PRs.

The unit of value is not "I can type `vpup` instead of `docker-compose up -d`". It's "any developer on this team can clone this repo and have a working environment in one command, and if something breaks, we fix the automation and everyone gets the fix."

This shapes how PI is designed:
- Shortcuts are declared in `pi.yaml` (team-visible), not in a local config
- Setup is `pi setup` (reproducible), not a wiki page telling you what to install
- Packages are pinned versions (reproducible), not "grab the latest from GitHub"
- Automations are committed and reviewed (trustworthy), not local scripts that work on one machine

**The rule:** every feature should be evaluated from the team perspective first. Does this help a new teammate onboard faster? Does this make the automation more portable? Does this make the setup more reproducible?

---

### 9. AI is a first-class user

PI is built in the AI coding era. An LLM reading `pi info docker/up` should get enough context to help a developer use it correctly. An LLM reading `docs/README.md` should be able to answer questions about PI's YAML syntax. A developer should be able to paste `pi.yaml` into Claude and get a useful code review.

This is why descriptions are full sentences. Why error messages include the file path and field name. Why `pi info` output is structured. Why there's a `llms.txt` and `.md` mirrors on the website.

**The rule:** descriptions must be complete sentences explaining intent ("Install Python at a specific version using mise or Homebrew") not name-echo ("Installs python"). Error messages must include the context needed to fix the problem without opening a browser. `docs/README.md` and `docs/architecture.md` are the ground truth — if behavior isn't documented there, it doesn't exist as far as future agents are concerned.

---

## Summary

| Principle | The rule |
|-----------|----------|
| First two minutes are sacred | Every entry point is designed for someone who doesn't know PI yet |
| Correct the action | High-confidence intent → take action + explain. Ambiguous → confirm. Never bare errors. |
| Short form is the real form | Minimal syntax first, everywhere. Generate the YAML; don't make them write it. |
| Guess correctly, every time | Identical patterns across all builtins. The pattern is the documentation. |
| The repo is the documentation | Every setup step idempotent. Check before writing. `test:` before `run:`. |
| Explain the magic | Trace lines. Advisory messages. "What's next." The answer is in the terminal. |
| One thing, done completely | Each builtin does one thing but does it fully. Compose via `run:`. |
| PI is a team tool | Evaluate every feature from the perspective of a new teammate on day one. |
| AI is a first-class user | Full-sentence descriptions. Structured output. Docs as ground truth. |
