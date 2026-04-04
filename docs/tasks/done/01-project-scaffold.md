# Project Scaffold

## Type
infra

## Status
done

## Priority
high

## Project
01-core-engine

## Description
Set up the Go module, CLI skeleton, and project folder structure for the `pi` binary. This is the first task — everything else builds on top of it. By the end, `pi --help` and `pi run --help` should work, even if `run` does nothing yet.

## Acceptance Criteria
- [x] Go module initialized at repo root (`go.mod`, module name `github.com/vyper-tooling/pi`)
- [x] CLI entry point at `cmd/pi/main.go` using `cobra`
- [x] Subcommands registered: `run`, `list`, `setup` (stubs — print "not implemented")
- [x] `go build ./...` succeeds and produces a `pi` binary
- [x] `go test ./...` passes (7 tests covering help, stubs, version, arg validation)
- [x] `Makefile` with targets: `build`, `test`, `install` (copies binary to `/usr/local/bin`)
- [x] `.gitignore` covers build artifacts
- [x] `README.md` at repo root with one-liner description and `go install` instructions

## Implementation Notes

### Decisions
- **Module path**: `github.com/vyper-tooling/pi` — matches the org/repo convention.
- **CLI framework**: Cobra v1.10.2 — standard Go CLI library, well-tested, good help generation.
- **Package layout**: `cmd/pi/main.go` as entry point, `internal/cli/` for all command definitions. Keeps main thin and commands testable.
- **Added `setup` subcommand**: Not in the original task spec but it's part of the core product (pi.yaml has a setup section), so added it as a stub alongside run/list.
- **Version injection**: Using `-ldflags` to inject version at build time via the Makefile. Defaults to "dev" in development.
- **Test approach**: Used a helper `executeCmd()` that creates a fresh root command, captures stdout to a buffer, and runs it. Tests cover: root help, run help, run stub output, run arg validation, list stub, setup stub, version flag.

### File structure
```
cmd/pi/main.go           — entry point, calls cli.Execute()
internal/cli/root.go     — root command, wires subcommands
internal/cli/run.go      — pi run (stub)
internal/cli/list.go     — pi list (stub)
internal/cli/setup.go    — pi setup (stub)
internal/cli/root_test.go — 7 tests
Makefile                 — build/test/install/clean
.gitignore               — build artifacts, IDE, OS files
README.md                — project description and install instructions
```

## Subtasks
- [x] `go mod init`
- [x] Add cobra dependency
- [x] Wire root command + `run`, `list`, `setup` subcommands
- [x] Write tests (7 tests)
- [x] Write Makefile
- [x] Write README
- [x] Write .gitignore

## Blocked By
nothing
