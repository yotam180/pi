# Color-based display improvements

## Type
improvement

## Status
done

## Priority
medium

## Project
11-improve-display-ux

## Description
The current terminal output produced by `pi setup` and `pi run` treats all output lines with equal visual weight. Already-satisfied installer steps look identical to steps that actually did work, and errors have no visual distinction. This task introduces intentional colour-coding across all pi output.

Desired palette:
- **Dim/grey** — steps that were skipped or already satisfied (e.g. `already installed`)
- **Normal/white** — steps currently executing
- **Bold green** — steps that completed successfully and did real work (e.g. `installed`, `done`)
- **Bold red** — steps that failed

All colour output must:
- Respect the `NO_COLOR` environment variable (disable colours when set)
- Detect whether stdout/stderr is a TTY; output plain text when piped
- Use a thin internal `color` package (or wrap `github.com/fatih/color`) rather than scattering ANSI codes

## Acceptance Criteria
- [x] Already-installed installer lines are rendered in dim/grey
- [x] Newly-installed installer lines are rendered bold green
- [x] Failed installer lines are rendered bold red
- [x] Setup entry header lines (`==> setup[N]: ...`) are rendered in a distinct muted style (dim white or grey)
- [x] Colour is disabled when `NO_COLOR` env var is set
- [x] Colour is disabled when stdout is not a TTY (e.g. piped to `grep`)
- [x] `go test ./...` passes
- [ ] Manual smoke test: `pi setup` on bat project shows visually distinct lines for installed vs already-installed

## Implementation Notes

### Approach
Created a new `internal/display` package that provides a `Printer` abstraction for styled terminal output. Key design decisions:

1. **Zero-dependency approach initially considered, then chose `golang.org/x/term`** — reliable cross-platform TTY detection via `term.IsTerminal()` rather than rolling our own with platform-specific syscalls.

2. **Printer is a value type, not a global** — each caller creates a Printer wrapping their `io.Writer`. This keeps testability high (in tests, writers are `bytes.Buffer`, which is not `*os.File`, so color is automatically disabled without any test changes).

3. **Lazy initialization on Executor** — the `Executor` has an optional `Printer` field. If nil, `printer()` creates one from the stderr writer on first use. The CLI's `setup.go` explicitly creates printers for both stdout and stderr.

4. **`InstallStatus()` encapsulates the icon→style mapping** — moved the formatting logic into the display package so the executor just calls `p.InstallStatus(icon, name, status, version)`.

5. **Also styled `pi doctor` output** — ✓ lines in green, ✗ lines in red, automation names in bold. Same pattern as installer output.

6. **`pi setup` "Setup complete." styled bold green** — gives a clear visual signal that setup finished successfully.

### Files changed
- `internal/display/display.go` — Printer struct, color methods (Plain, Dim, Green, Red, Bold), InstallStatus, SetupHeader, shouldColor
- `internal/display/tty.go` — isTerminal via golang.org/x/term
- `internal/display/display_test.go` — 25 unit tests covering all styles, color on/off, NO_COLOR, TTY detection, install status variants
- `internal/executor/executor.go` — added Printer field, printer() lazy init, updated printInstallStatus to delegate to display.Printer.InstallStatus
- `internal/cli/setup.go` — creates display.Printer for stdout/stderr, uses SetupHeader for `==>` lines, Green for "Setup complete."
- `internal/cli/doctor.go` — creates display.Printer, uses Bold for automation names, Green for ✓ lines, Red for ✗ lines

### Backward compatibility
All existing tests pass unchanged because:
- Tests use `bytes.Buffer` as writers, which is not `*os.File` — `shouldColor()` returns false → no ANSI codes → identical text output
- The text content of all messages is unchanged; only ANSI wrapping is added in TTY mode

## Subtasks
- [x] Audit all `fmt.Print*` / `log.*` calls in the renderer/printer layer and centralise them
- [x] Introduce a colour abstraction (check if `fatih/color` is already a dep, else add it)
- [x] Apply dim style to "already installed" installer output
- [x] Apply green style to "installed" installer output
- [x] Apply red style to "failed" installer output
- [x] Apply dim style to setup section headers
- [x] Add TTY detection + NO_COLOR support
- [x] Add unit tests for the colour/printer layer

## Blocked By
