# GitHub Actions CI Workflow

## Type
infra

## Status
todo

## Priority
high

## Project
06-production-readiness

## Description
Create a GitHub Actions CI workflow that runs on every push and pull request. Tests must pass on both Linux and macOS before a PR can merge.

## Acceptance Criteria
- [ ] `.github/workflows/ci.yml` exists and triggers on `push` and `pull_request`
- [ ] Workflow runs `go test ./... -race` and `go vet ./...`
- [ ] Matrix covers `ubuntu-latest` and `macos-latest`
- [ ] `go build ./...` is verified (confirms the binary compiles)
- [ ] Workflow uses the Go version from `go.mod` (via `actions/setup-go` with `go-version-file: go.mod`)
- [ ] CI passes on the current codebase
- [ ] Failed tests cause the workflow to fail (non-zero exit)

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Create `.github/workflows/ci.yml`
- [ ] Configure `actions/checkout` + `actions/setup-go` with caching
- [ ] Add `go vet`, `go build`, `go test -race` steps
- [ ] Set up OS matrix (`ubuntu-latest`, `macos-latest`)
- [ ] Push and verify the workflow runs green

## Blocked By
<!-- None -->
