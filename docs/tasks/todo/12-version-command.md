# `pi --version` Command

## Type
feature

## Status
todo

## Priority
high

## Project
06-production-readiness

## Description
Expose the version string (already injected at build time via ldflags into `internal/cli.version`) through `pi --version` and `pi version`. The Makefile already sets the version to `git describe --tags --always --dirty`, so this task is purely about wiring it to the CLI output.

## Acceptance Criteria
- [ ] `pi --version` prints `pi <version>` (e.g. `pi v0.1.0`, `pi v0.1.0-3-gabc1234`, `pi dev` when no tags exist)
- [ ] `pi version` subcommand prints the same output
- [ ] Version string is the value injected by the Makefile ldflags, not hardcoded
- [ ] `go build` without the Makefile (no ldflags) prints `pi dev` as fallback, not empty or a panic
- [ ] `go test ./...` passes

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] Confirm `internal/cli.version` variable exists and is properly declared for ldflags injection
- [ ] Wire `--version` flag on the root cobra command
- [ ] Add `version` subcommand as an alias
- [ ] Set default value of `version` var to `"dev"` (fallback when built without ldflags)
- [ ] Write a test that the version command exits 0 and prints something non-empty

## Blocked By
<!-- None -->
