# Project Scaffold

## Type
infra

## Status
todo

## Priority
high

## Project
01-core-engine

## Description
Set up the Go module, CLI skeleton, and project folder structure for the `pi` binary. This is the first task — everything else builds on top of it. By the end, `pi --help` and `pi run --help` should work, even if `run` does nothing yet.

## Acceptance Criteria
- [ ] Go module initialized at repo root (`go.mod`, module name `github.com/vyper-tooling/pi` or similar)
- [ ] CLI entry point at `cmd/pi/main.go` using `cobra`
- [ ] Subcommands registered: `run`, `list` (stubs — no logic yet, just print "not implemented")
- [ ] `go build ./...` succeeds and produces a `pi` binary
- [ ] `go test ./...` passes (no tests yet is fine, just must not fail)
- [ ] `Makefile` with targets: `build`, `test`, `install` (copies binary to `/usr/local/bin`)
- [ ] `.gitignore` covers build artifacts
- [ ] `README.md` at repo root with one-liner description and `go install` instructions

## Implementation Notes
<!-- Fill in as you work -->

## Subtasks
- [ ] `go mod init`
- [ ] Add cobra dependency
- [ ] Wire root command + `run` and `list` subcommands
- [ ] Write Makefile
- [ ] Write README

## Blocked By
nothing
