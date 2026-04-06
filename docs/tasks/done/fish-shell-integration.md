# Fish Shell Integration

## Type
feature

## Status
done

## Priority
low

## Project
standalone

## Description

`pi shell` previously installed shell integration only for zsh and bash. Fish users who ran `pi shell` got no shortcut functions, no shell config modification, and no completion wiring. This task adds fish support to `pi shell` so the installation experience is symmetric across shells.

**What was missing:**
1. `shellConfigPaths()` only checked for `.zshrc` and `.bashrc`
2. `GenerateCompletionScript()` only emitted `$ZSH_VERSION` and `$BASH_VERSION` detection
3. No fish-syntax shortcut function generator existed
4. `buildWithArgs()` generated bash-specific `$1` positional refs instead of fish's `$argv[1]`

## Acceptance Criteria
- [x] `hasFishConfig()` detects fish and returns true when `~/.config/fish/config.fish` exists
- [x] Fish shortcut functions are generated with correct fish syntax
- [x] `FishDialect` implements `ShellDialect` interface with fish-specific syntax
- [x] `PositionalArg()` method added to `ShellDialect` interface for dialect-aware positional arg syntax
- [x] `GenerateFishCompletionScript()` generates fish completion setup
- [x] `pi shell` installs fish files when fish config exists, skips when absent
- [x] `pi shell uninstall` removes fish function files
- [x] `pi shell list` shows both `.sh` and `.fish` files, deduplicates project names
- [x] Fish source line in `config.fish` is idempotent
- [x] `go build ./...` and `go test ./...` pass
- [x] 42 new tests added (91 total in shell package, up from 49)

## Implementation Notes

### FishDialect
Implemented `FishDialect` struct in `dialect.go` implementing the `ShellDialect` interface:
- `EvalWrapperFunc`: Uses `function name ... end` syntax, `set -l` for locals, `$status` instead of `$?`, `eval (cat ...)` instead of `source`
- `InRepoCmd`: Uses `env -C "/path"` instead of `(cd "/path" && ...)` subshell
- `AnywhereCmd`: Same `--repo` flag approach as bash
- `AllArgs`: Returns `$argv` instead of `"$@"`
- `PositionalArg`: Returns `$argv[N]` instead of `"$N"`

### ShellDialect interface extension
Added `PositionalArg(n string) string` method to the `ShellDialect` interface. This was necessary because `buildWithArgs()` generates positional argument references (e.g. `$1`) which differ between shells. Bash uses `"$1"`, fish uses `$argv[1]`.

### Fish detection
- `hasFishConfig()` checks if `~/.config/fish/config.fish` exists
- Fish files are only generated when fish config is detected — no fish files are created for bash/zsh-only users
- This is a non-breaking change: existing bash/zsh users see zero behavior difference

### File layout
When fish is detected, `pi shell` generates:
- `~/.pi/shell/<project>.fish` — fish shortcut functions
- `~/.pi/shell/_pi-wrapper.fish` — global `pi` wrapper for fish
- `~/.pi/shell/_pi-completion.fish` — fish completion setup
- Source line in `~/.config/fish/config.fish`

### Uninstall
- `Uninstall()` removes both `.sh` and `.fish` project files
- Fish infrastructure files (wrapper, completion) are removed only when no fish project files remain
- Fish source line in `config.fish` is removed only when no fish project files remain

### ListInstalled
- Deduplicates projects that have both `.sh` and `.fish` files
- Excludes fish infrastructure files from the project list
- `pi shell list` now shows individual file names (`.sh` and `.fish`) per project

### Tests added (42 new)
**dialect_test.go** (14 new):
- FishDialect: EvalWrapperFunc, InRepoCmd, AnywhereCmd, AllArgs, PositionalArg, FileHeader
- FishDialect: InRepoCmd/AnywhereCmd with spaces, ComplexInnerCmd
- FishDialect: PlugsIntoGeneration, GlobalWrapper
- BashDialect: PositionalArg
- BuildWithArgs: FishDialect (3 sub-tests)

**shell_test.go** (28 new):
- Install: CreatesFishFilesWhenFishConfigExists, SkipsFishWhenNoFishConfig, FishSourceLineIdempotent
- Uninstall: RemovesFishFiles, KeepsFishFilesIfOtherProjectsExist
- ListInstalled: IncludesFishOnlyProjects, ExcludesFishInfraFiles, DeduplicatesShAndFish
- FishFilePath, GenerateFishCompletionScript, HasFishConfig, IsInfraFile (6 sub-tests)

## Subtasks
- [x] Add `FishDialect` implementing `ShellDialect` with fish syntax
- [x] Add `PositionalArg()` to `ShellDialect` interface
- [x] Add fish config path detection (`hasFishConfig`, `fishConfigPath`)
- [x] Generate and install `.fish` files alongside `.sh` files
- [x] Generate fish global wrapper and completion script
- [x] Add fish source line management (`ensureFishSourceLine`, `removeFishSourceLine`)
- [x] Update `Uninstall` to clean up fish files
- [x] Update `ListInstalled` to handle fish files
- [x] Update CLI `pi shell list` to show both `.sh` and `.fish` files
- [x] Update CLI `pi shell` to mention fish installation
- [x] Write comprehensive tests (42 new tests)
- [x] Run full test suite — all pass
- [x] Update architecture docs

## Blocked By
