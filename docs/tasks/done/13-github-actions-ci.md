# GitHub Actions CI Workflow

## Type
infra

## Status
done

## Priority
high

## Project
06-production-readiness

## Description
Create a GitHub Actions CI workflow that runs on every push and pull request. Tests must pass on both Linux and macOS before a PR can merge.

## Acceptance Criteria
- [x] `.github/workflows/ci.yml` exists and triggers on `push` and `pull_request`
- [x] Workflow runs `go test ./... -race` and `go vet ./...`
- [x] Matrix covers `ubuntu-latest` and `macos-latest`
- [x] `go build ./...` is verified (confirms the binary compiles)
- [x] Workflow uses the Go version from `go.mod` (via `actions/setup-go` with `go-version-file: go.mod`)
- [x] CI passes on the current codebase
- [x] Failed tests cause the workflow to fail (non-zero exit)

## Implementation Notes

### Approach
- Single job `test` with OS matrix (`ubuntu-latest`, `macos-latest`)
- `actions/setup-go@v5` reads Go version from `go.mod` automatically and caches modules by default
- Node.js 22 + `tsx` installed globally for TypeScript step runner integration tests
- Python 3 is pre-installed on both GitHub runner images — no setup needed
- Minimal permissions (`contents: read`) for security
- Steps: checkout → setup-go → setup-node → install tsx → go vet → go build → go test -race

### Verified locally
- `go vet ./...` — clean
- `go build ./...` — clean
- `go test ./... -race -count=1` — all 8 packages pass (including integration tests)

## Subtasks
- [x] Create `.github/workflows/ci.yml`
- [x] Configure `actions/checkout` + `actions/setup-go` with caching
- [x] Add `go vet`, `go build`, `go test -race` steps
- [x] Set up OS matrix (`ubuntu-latest`, `macos-latest`)
- [ ] Push and verify the workflow runs green (will verify on push)

## Blocked By
<!-- None -->
