# Docker-Based Test Matrix

## Type
infra

## Status
done

## Priority
medium

## Project
05-environment-robustness

## Description
Create Docker-based test environments that prove PI works on fresh systems. These serve as both CI infrastructure and documentation of what "a fresh machine" means.

### Test environments
```
tests/docker/
  ubuntu-fresh/       Dockerfile — Ubuntu 24.04 (bookworm), no runtimes, no tools
  ubuntu-node/        Dockerfile — Ubuntu 24.04 (bookworm), Node 20 + tsx
  ubuntu-python/      Dockerfile — Ubuntu 24.04 (bookworm), Python 3 only
  alpine-fresh/       Dockerfile — Alpine (golang:1.26.1-alpine), bare minimum
```

Each Dockerfile installs only Go (via official golang image) and the specified runtime (if any). No other tools are pre-installed.

### Test runner
A `make test-matrix` target runs `tests/docker/test-matrix.sh` which:
1. Builds each Docker image
2. Runs `go test ./... -count=1` inside each container
3. Reports pass/fail per environment with ✓/✗ icons
4. Exits with non-zero if any environment fails

### Key test scenarios per environment
- `ubuntu-fresh`: all tests pass, Python/TypeScript tests skipped gracefully via `t.Skip()`
- `ubuntu-node`: TypeScript tests work, Python tests skipped gracefully
- `ubuntu-python`: Python tests work, TypeScript tests skipped gracefully
- `alpine-fresh`: basic tests pass, confirms no glibc assumptions

## Acceptance Criteria
- [x] 4 Dockerfiles created in `tests/docker/`
- [x] Each image is self-contained (no external dependencies beyond Docker)
- [x] `make test-matrix` target builds images and runs tests
- [x] CI job added for matrix testing (separate workflow: `.github/workflows/docker-matrix.yml`)
- [x] All environments pass their expected tests
- [x] Pass/fail summary printed at the end
- [x] `go test ./...` passes on host

## Implementation Notes

### Runtime skip guards
Added `requirePython(t)`, `requireNode(t)`, and `requireTsx(t)` helper functions that call `t.Skip()` when the respective runtime isn't in PATH. These are used in:

- `internal/executor/executor_test.go`: All Python inline/file/pipe tests and the `requirePython()` helper (matching the existing `requireTsx()` pattern)
- `tests/integration/helpers_test.go`: Shared helpers for integration tests
- `tests/integration/polyglot_test.go`: All Python and TypeScript execution tests
- `tests/integration/examples_test.go`: Tests that require Python (validation, doctor, provisioning) and tsx (installer idempotency)

This approach lets the same test suite run on any environment — tests that can't execute naturally skip rather than fail.

### Docker images
- All Ubuntu images use `golang:1.26.1-bookworm` as base
- Alpine uses `golang:1.26.1-alpine`
- Each image copies the full repo, downloads Go modules, and runs `go test ./...`
- Node image installs Node 20 via NodeSource and tsx globally via npm

### CI workflow
Separate workflow (`.github/workflows/docker-matrix.yml`) that runs on PRs to main. Uses GitHub Actions matrix strategy with `fail-fast: false` so all environments always run.

### Makefile
Created `Makefile` with `build`, `vet`, `test`, and `test-matrix` targets.

## Subtasks
- [x] Write Dockerfiles for all 4 environments
- [x] Write `test-matrix.sh` script
- [x] Add Makefile target
- [x] Add CI workflow (separate job)
- [x] Verify all environments pass

## Blocked By
32-sandboxed-runtime-provisioning (done)
