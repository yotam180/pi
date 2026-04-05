# On-Demand Package Fetching with Advisory

## Type
feature

## Status
done

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
- [x] Undeclared GitHub ref triggers automatic fetch with the advisory message
- [x] Advisory is printed to stderr, not stdout
- [x] Advisory includes a ready-to-paste `pi.yaml` snippet
- [x] Advisory shown once per package per invocation (not per step referencing it)
- [x] Subsequent runs using cached result are fully silent (no advisory)
- [x] A `file:` ref that doesn't exist is an error, not on-demand — clear message
- [x] On-demand fetch respects the same auth logic (SSH / `GITHUB_TOKEN`) as declared packages
- [x] Tests cover: first-time fetch advisory, cached-no-advisory, file-missing error

## Implementation Notes

### Architecture
- `OnDemandFetchFunc` type and `OnDemandFetch` field added to `discovery.Result`
- `findInPackage()` checks `ref.Type == refparser.RefFile` first — file: refs always error, never on-demand
- For GitHub refs not in declared packages, `findInPackage()` delegates to `OnDemandFetch` callback
- `newOnDemandFetcher()` in `cli/discover.go` creates a closure with a `fetched` map for per-invocation dedup
- Advisory output uses `display.Printer.PackageFetch()` for styled status + plain `fmt.Fprintf` for tip text
- `wasCached` check: stat the cache path before calling `Fetch()` — advisory shown only on live network fetch
- `PackageAutomations()` method added to `Result` for looking up automations within a specific package source

### Tests added
- **Discovery unit tests** (6 tests in `discovery_test.go`): callback invocation, nil callback error, file: ref never triggers, declared package skips callback, error propagation, PackageAutomations accessor
- **CLI unit tests** (4 tests in `discover_test.go`): advisory output format, nil writer safety, fetch status text, down arrow icon
- **Integration tests** (6 tests in `on_demand_test.go`): file: ref never on-demand, GitHub ref shows on-demand error, declared package no advisory, cached package silent, file: missing path clear error, advisory to stderr

## Subtasks
- [x] Hook on-demand fetch into the automation resolver
- [x] Implement advisory output (deduped per invocation)
- [x] Add tests

## Blocked By
69-github-package-cache, 70-packages-declaration-in-pi-yaml
