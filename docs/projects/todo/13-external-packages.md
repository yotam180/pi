# External Packages

## Status
todo

## Priority
high

## Description
Let PI automations be sourced from anywhere: the local project, a GitHub repo at a specific version, or a local folder on disk. Automations from all sources are resolved and executed identically — the source is transparent to the step body.

## Goals
- Any automation can be referenced as `org/repo@version/automation-path` (GitHub) or `file:~/path/automation` (local folder)
- Declared packages in `pi.yaml` are fetched by `pi setup`; undeclared packages are fetched on demand with a clear advisory to add them
- Local folder sources (`file:`) support a named alias so you don't repeat the full path everywhere
- Private GitHub repos work when SSH or a token is configured
- `pi list` shows all automations with their source; `--all` browses everything installed across all packages
- `pi add` writes a package entry to `pi.yaml` and fetches it immediately
- A `pi-package.yaml` in a GitHub repo is entirely optional — its only job is declaring a minimum PI version

## Background & Context
PI's built-in library is intentionally thin. The natural next layer is a public shared automations repo (`yotam180/pi-common`) and the ability for any team to publish their own. Without this feature, the only way to share automations across projects is copy-paste. This also supports the local development workflow: iterate on a shared automations folder without push/tag/clone cycles.

## Scope

### In scope
- Parsing and resolving three source types: local `.pi/`, `file:`, and GitHub (`org/repo@version`)
- GitHub package caching in `~/.pi/cache/`
- `packages:` block in `pi.yaml` (declared dependencies)
- On-demand fetching with advisory message
- `file:` source aliases (`as:` key)
- Private repo support via SSH / `GITHUB_TOKEN`
- `pi list` source indicators and `--all` flag
- `pi add` command writing to `pi.yaml`
- Optional `pi-package.yaml` with `min_pi_version:` field

### Out of scope
- A web-based package registry or search UI
- Non-GitHub VCS sources (GitLab, Bitbucket) — future work
- Automatic version upgrades (`pi update`)
- Package scoping / access control beyond what GitHub provides

## Success Criteria
- [ ] `run: org/repo@version/path` in any automation works end-to-end
- [ ] `run: file:~/my-automations/path` works; changes to the folder are reflected immediately
- [ ] Declared packages in `pi.yaml` are fetched during `pi setup`
- [ ] Undeclared GitHub reference triggers a fetch + advisory message with copy-pasteable `pi.yaml` snippet
- [ ] `pi add org/repo@version` adds to `pi.yaml` and fetches
- [ ] `pi add file:~/path --as alias` adds aliased file source to `pi.yaml`
- [ ] Private repos accessible via SSH work transparently
- [ ] `pi list` shows source indicator per automation
- [ ] `pi list --all` lists automations from all declared packages
- [ ] Missing or malformed package ref produces a clear, actionable error
- [ ] All new behavior documented in `docs/README.md`

## Notes
- Resolution priority: local `.pi/` > `file:` sources > GitHub packages > builtins. Local always wins.
- A `pi-package.yaml` is never required. Its only valid field for now is `min_pi_version`.
- The `@version` part of a GitHub ref is required (no silent "latest"). Use `@main` for mutable refs with a warning.
- Mutable refs (`@main`, `@HEAD`) emit a reproducibility warning at fetch time.
- `file:` sources are never cached — they're read directly from disk every time.
