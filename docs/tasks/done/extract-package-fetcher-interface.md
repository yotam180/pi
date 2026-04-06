# Extract PackageFetcher Interface for Testable GitHub Package Operations

## Type
improvement

## Status
done

## Priority
high

## Project
standalone

## Description

The CLI package (`internal/cli`) has three functions that directly construct `cache.Cache` objects and perform filesystem I/O to fetch/check GitHub packages:

1. `resolveGitHubPackage()` in `discover.go` — 0% coverage
2. `newOnDemandFetcher()` in `discover.go` — 6.9% coverage
3. `fetchGitHubPackage()` in `add.go` — 0% coverage

All three follow the same pattern:
- Call `cache.DefaultCacheRoot()` to get the cache directory
- Construct a `cache.Cache{}`
- Call `c.PackagePath()` and `os.Stat()` to check if cached
- Call `c.Fetch()` if not cached
- Print status via `display.Printer`

This pattern is duplicated across all three functions and prevents unit testing because it requires a real filesystem and network access.

**Solution:** Extract a `PackageFetcher` interface that abstracts the cache lookup + fetch lifecycle. Implement it with the existing `cache.Cache` behavior. Inject it into the CLI functions so tests can provide a mock. Write comprehensive unit tests for all three functions.

This also makes the codebase extensible for future package source types (OCI registries, S3, etc.).

## Acceptance Criteria
- [x] `PackageFetcher` interface defined with `Fetch(org, repo, version) (path string, wasCached bool, err error)` signature
- [x] `CachePackageFetcher` concrete type wraps `cache.Cache` and implements the interface
- [x] `resolveGitHubPackage` accepts a `PackageFetcher` parameter
- [x] `newOnDemandFetcher` accepts a `PackageFetcher` parameter
- [x] `fetchGitHubPackage` accepts a `PackageFetcher` parameter
- [x] Callers updated to construct `CachePackageFetcher` and pass it
- [x] Unit tests cover: cached hit, fresh fetch, fetch failure, on-demand dedup, advisory output
- [x] All existing tests pass
- [x] `go build ./...` succeeds
- [x] CLI package coverage improves (target: >82%) — achieved 84.2% (up from 78.9%)

## Implementation Notes

### Approach
Introduced a `PackageFetcher` interface in `internal/cli/package_fetcher.go` that abstracts the cache-check → fetch → return-path lifecycle for GitHub packages. The interface has a single method:

```go
type PackageFetcher interface {
    Fetch(org, repo, version string) (path string, wasCached bool, err error)
}
```

The `wasCached` return value is important — it lets callers distinguish cache hits from fresh fetches so they can print appropriate status (e.g., the on-demand fetcher only prints advisories on fresh fetches).

### Changes
1. **`package_fetcher.go`** — new file: `PackageFetcher` interface + `CachePackageFetcher` concrete impl wrapping `cache.Cache`
2. **`discover.go`** — `resolveGitHubPackage()`, `mergePackages()`, `newOnDemandFetcher()` all accept `PackageFetcher`; `discoverAllWithConfig()` creates `CachePackageFetcher` at the top and passes it down
3. **`add.go`** — `fetchGitHubPackage()` accepts `PackageFetcher`; extracted `runAddWithFetcher()` for testing; `runAdd()` delegates with nil fetcher (lazy creation)
4. **`package_fetcher_test.go`** — 22 tests covering all three refactored functions via `mockFetcher`

### Design decisions
- **Interface lives in CLI package** — it's only used there. If other packages need it later, it can be promoted.
- **`wasCached` in return value** — considered making this a separate `IsCached()` method but the single-call pattern is simpler and matches all three call sites.
- **Nil fetcher with lazy creation** — `newOnDemandFetcher` and `runAddWithFetcher` accept nil and create a `CachePackageFetcher` lazily. This keeps the default path (production code) unchanged and only requires explicit fetcher injection in tests.
- **Removed `cache` import from `add.go`** — now only `package_fetcher.go` imports `cache`, reducing coupling.

### Coverage improvement
- CLI package: 78.9% → 84.2% (+5.3pp)
- `fetchGitHubPackage`: 0% → 100%
- `resolveGitHubPackage`: 0% → 94.4%
- `newOnDemandFetcher`: 6.9% → 77.8%
- `runAdd`: 83.3% → 100%

## Subtasks
- [x] Define interface and concrete type
- [x] Refactor discover.go
- [x] Refactor add.go
- [x] Write tests (22 new tests)
- [x] Verify coverage improvement (84.2%)

## Blocked By
