# `pi --version` Command

## Type
feature

## Status
done

## Priority
high

## Project
06-production-readiness

## Description
Expose the version string (already injected at build time via ldflags into `internal/cli.version`) through `pi --version` and `pi version`. The Makefile already sets the version to `git describe --tags --always --dirty`, so this task is purely about wiring it to the CLI output.

## Acceptance Criteria
- [x] `pi --version` prints `pi <version>` (e.g. `pi v0.1.0`, `pi v0.1.0-3-gabc1234`, `pi dev` when no tags exist)
- [x] `pi version` subcommand prints the same output
- [x] Version string is the value injected by the Makefile ldflags, not hardcoded
- [x] `go build` without the Makefile (no ldflags) prints `pi dev` as fallback, not empty or a panic
- [x] `go test ./...` passes

## Implementation Notes

### What was already in place
- `internal/cli/root.go` had `var version = "dev"` and `Version: version` on the root Cobra command
- Makefile had ldflags injection: `-X github.com/vyper-tooling/pi/internal/cli.version=$(VERSION)`
- `TestVersion` test existed, checking `--version` output contains "dev"

### Changes made
1. **Custom version template** — Set `root.SetVersionTemplate("pi {{.Version}}\n")` so `--version` prints `pi <version>` instead of Cobra's default `pi version <version>`
2. **`pi version` subcommand** — Added `internal/cli/version.go` with a simple subcommand that prints the same `pi <version>` output
3. **Enhanced tests** — Replaced the single `TestVersion` with `TestVersionFlag`, `TestVersionSubcommand`, and `TestVersionNonEmpty` in unit tests
4. **Integration tests** — Added `TestVersion_Flag`, `TestVersion_Subcommand`, and `TestVersion_FlagAndSubcommandMatch` to verify both paths produce identical output from the compiled binary

## Subtasks
- [x] Confirm `internal/cli.version` variable exists and is properly declared for ldflags injection
- [x] Wire `--version` flag on the root cobra command (customized template)
- [x] Add `version` subcommand as an alias
- [x] Set default value of `version` var to `"dev"` (fallback when built without ldflags)
- [x] Write tests (unit + integration) that version command exits 0 and prints correct format

## Blocked By
<!-- None -->
