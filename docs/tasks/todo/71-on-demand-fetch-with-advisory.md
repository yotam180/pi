# On-Demand Package Fetching with Advisory

## Type
feature

## Status
todo

## Priority
high

## Project
13-external-packages

## Description
When PI encounters a GitHub automation reference (`org/repo@version/path`) that is not declared in `pi.yaml packages:` and is not already cached, it should fetch it automatically rather than failing — but it must clearly tell the user what happened and what to do.

**User-facing output during on-demand fetch:**

```
  ↓  fetching yotam180/pi-common@v1.2 (on demand)...
  ✓  yotam180/pi-common@v1.2  cached

  tip: add to pi.yaml to avoid fetching on every fresh clone:

    packages:
      - yotam180/pi-common@v1.2

```

The advisory is printed to stderr so it doesn't interfere with piped automation output. It is shown once per package per PI invocation, not once per step.

If the package is already cached (previous on-demand fetch), it is used silently — no advisory is shown again. The advisory is only shown when a live network fetch happens.

The goal is: things just work, but the user is nudged toward declaring dependencies explicitly. The `file:` source type is never on-demand — a `file:` ref that doesn't exist is always an immediate error.

## Acceptance Criteria
- [ ] Undeclared GitHub ref triggers automatic fetch with the advisory message
- [ ] Advisory is printed to stderr, not stdout
- [ ] Advisory includes a ready-to-paste `pi.yaml` snippet
- [ ] Advisory shown once per package per invocation (not per step referencing it)
- [ ] Subsequent runs using cached result are fully silent (no advisory)
- [ ] A `file:` ref that doesn't exist is an error, not on-demand — clear message
- [ ] On-demand fetch respects the same auth logic (SSH / `GITHUB_TOKEN`) as declared packages
- [ ] Tests cover: first-time fetch advisory, cached-no-advisory, file-missing error

## Implementation Notes

## Subtasks
- [ ] Hook on-demand fetch into the automation resolver
- [ ] Implement advisory output (deduped per invocation)
- [ ] Add tests

## Blocked By
69-github-package-cache, 70-packages-declaration-in-pi-yaml
