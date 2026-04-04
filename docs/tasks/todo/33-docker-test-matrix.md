# Docker-Based Test Matrix

## Type
infra

## Status
todo

## Priority
medium

## Project
05-environment-robustness

## Description
Create Docker-based test environments that prove PI works on fresh systems. These serve as both CI infrastructure and documentation of what "a fresh machine" means.

### Test environments
```
tests/docker/
  ubuntu-fresh/       Dockerfile — Ubuntu 24.04, no runtimes, no tools
  ubuntu-node/        Dockerfile — Ubuntu 24.04, Node 20 only
  ubuntu-python/      Dockerfile — Ubuntu 24.04, Python 3.13 only
  alpine-fresh/       Dockerfile — Alpine 3.19, bare minimum
```

Each Dockerfile should install only Go (for running tests) and the specified runtime (if any). No other tools should be pre-installed.

### Test runner
A `make test-matrix` target should:
1. Build each Docker image
2. Run `go test ./...` + integration tests inside each container
3. Report pass/fail per environment
4. Exit with non-zero if any environment fails

The CI pipeline should run this matrix on PRs (optionally, as a separate job to avoid slowing down the main CI).

### Key test scenarios per environment
- `ubuntu-fresh`: all tests pass, TypeScript steps fail gracefully with clear error, `requires:` validation catches missing runtimes
- `ubuntu-node`: TypeScript steps work, Python steps fail gracefully
- `ubuntu-python`: Python steps work, TypeScript steps fail gracefully
- `alpine-fresh`: basic tests pass, confirms no glibc assumptions

## Acceptance Criteria
- [ ] 4 Dockerfiles created in `tests/docker/`
- [ ] Each image is self-contained (no external dependencies beyond Docker)
- [ ] `make test-matrix` target builds images and runs tests
- [ ] CI job added for matrix testing (can be a separate workflow)
- [ ] All environments pass their expected tests
- [ ] Pass/fail summary printed at the end
- [ ] `go test ./...` passes on host

## Implementation Notes

## Subtasks
- [ ] Write Dockerfiles for all 4 environments
- [ ] Write `test-matrix.sh` script
- [ ] Add Makefile target
- [ ] Add CI workflow (optional separate job)
- [ ] Verify all environments pass

## Blocked By
32-sandboxed-runtime-provisioning
