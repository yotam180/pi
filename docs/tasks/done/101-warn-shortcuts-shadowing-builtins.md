# Warn when shortcuts shadow shell builtins or common commands

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
`pi shell` generates shell functions for shortcuts with no validation against POSIX/shell builtins or common command names. A user can easily define a shortcut called `test` which shadows the POSIX `test` builtin (used in `if test -f file; then...` and `[` expressions).

### Steps to Reproduce
1. Add to pi.yaml:
```yaml
shortcuts:
  test: my-test-automation
```
2. Run `pi shell`
3. Source `~/.zshrc`
4. Try: `test -f Cargo.toml && echo exists`

### Expected
Pi should either:
- Warn the user that `test` shadows a shell builtin and suggest a different name
- Refuse to create the shortcut with a note explaining why
- Or at minimum, document the risk

### Actual
The shortcut is silently installed. The POSIX `test` builtin is now inaccessible in any shell session (the function shadows it). This could break dotfiles, scripts, and other tools that depend on `test`.

Other dangerous names include: `run`, `exec`, `cd`, `echo`, `printf`, `source`, `export`, `set`, `type`, `which`, `kill`, `wait`, `time`, `read`, `eval`.

### Suggested Fix
Maintain a blocklist of POSIX builtins and common commands. When a shortcut name matches:
1. Print a warning: `warning: shortcut "test" shadows the shell builtin "test". Consider renaming to "pi-test" or "t".`
2. Still install it (user might know what they're doing), but make the risk visible.

For the `test` case specifically — suggest `t` or `check` as the shortcut name.

## Acceptance Criteria
- [x] `pi shell` warns when a shortcut name matches a known shell builtin
- [x] Warnings suggest alternative names
- [x] The shortcut is still installed (warning, not error) unless --strict is passed

## Implementation Notes

### Approach
Created `internal/shell/shadow.go` with:
- A `ShadowWarning` struct capturing the shortcut name, kind ("shell builtin" / "common command"), and a suggestion
- `shellBuiltins` map: 36 POSIX/common shell builtins (test, cd, echo, eval, exec, export, etc.)
- `commonCommands` map: 30 widely-used system commands (git, ls, rm, curl, make, etc.)
- `CheckShadowedNames()` — takes a list of names, returns sorted warnings
- `FormatWarning()` — produces human-readable warning strings
- Suggestions always use the `pi-<name>` pattern for simplicity

### Integration points
- `cli/shell.go` — `runShellInstall()` calls `warnShadowedShortcuts()` before `shell.Install()`
- `cli/setup.go` — `runSetup()` calls `warnShadowedShortcuts()` before `shell.Install()` during post-setup shortcut installation
- Warnings are written to stderr using `display.Printer.Warn()` (yellow styling when TTY)
- Shortcuts are still installed regardless of warnings — this is a warning, not an error

### Tests
- `internal/shell/shadow_test.go`: 13 tests covering no shadows, shell builtins, common commands, multiple warnings, empty input, case insensitivity, exhaustive builtin/command coverage, suggestions, formatting, and sorted output
- `internal/cli/shell_test.go`: 2 new tests — `TestShellCmd_Install_ShadowWarning` (verifies warnings appear on stderr for `test` and `echo` but not for safe names, and that shortcuts still install) and `TestShellCmd_Install_NoWarningForSafeNames` (verifies clean stderr for safe names)

## Subtasks
- [x] Define blocklist of shell builtins and common commands
- [x] Implement CheckShadowedNames and FormatWarning in internal/shell/shadow.go
- [x] Wire into cli/shell.go runShellInstall
- [x] Wire into cli/setup.go runSetup
- [x] Write unit tests for shadow.go
- [x] Write CLI integration tests for warning output
- [x] Full build + test pass
- [x] Manual QA with test workspace

## Blocked By
