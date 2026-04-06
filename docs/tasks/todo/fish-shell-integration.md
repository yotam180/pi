# Fish Shell Integration

## Type
feature

## Status
todo

## Priority
low

## Project
standalone

## Description

`pi shell` currently installs shell integration only for zsh and bash. `pi completion` generates fish completions. Fish users who run `pi shell` get no shortcut functions, no shell config modification, and no completion wiring — despite the CLI claiming fish completion support. This task adds fish support to `pi shell` so the installation experience is symmetric across shells.

**Current gaps:**
1. `shellConfigPaths()` in `internal/shell/shell.go` checks only for `.zshrc` and `.bashrc`
2. `GenerateCompletionScript()` only emits `$ZSH_VERSION` and `$BASH_VERSION` detection — no fish
3. No fish-syntax shortcut function generator exists

**Fish-specific considerations:**
- Fish config: `~/.config/fish/config.fish`
- Fish doesn't source `.sh` files — fish function files must be `.fish` and placed in `~/.config/fish/functions/` or sourced via config.fish
- Fish function syntax differs from bash: `function name; ...; end` not `function name() { ... }`
- Fish uses `$argv` not `$@` for variadic args
- Fish eval: `eval (...)` not `eval "$(...)"`

**Proposed approach:**
- Detect fish by checking for `~/.config/fish/config.fish` or `$FISH_VERSION`
- Generate fish-syntax function files into `~/.pi/shell/<project>.fish`
- Write a source line into `~/.config/fish/config.fish`
- The completion script should also handle fish: `if set -q FISH_VERSION; eval (pi completion fish); end`

**Fish shortcut template:**
```fish
function vpup
  set _pi_eval_file (mktemp)
  cd /path/to/project && PI_PARENT_EVAL_FILE=$_pi_eval_file pi run docker/up $argv
  set _pi_exit $status
  if test -s $_pi_eval_file
    eval (cat $_pi_eval_file)
  end
  rm -f $_pi_eval_file
  return $_pi_exit
end
```

## Acceptance Criteria
- [ ] `shellConfigPaths()` detects fish and returns `~/.config/fish/config.fish` when present
- [ ] Fish shortcut functions are generated with correct fish syntax
- [ ] `GenerateCompletionScript()` includes fish completion detection
- [ ] `pi shell uninstall` removes fish function files
- [ ] Fish users get working shortcuts after `pi shell`
- [ ] `go build ./...` and `go test ./...` pass

## Implementation Notes

## Subtasks
- [ ] Add fish config path detection to `shellConfigPaths()`
- [ ] Write `generateFishFunction()` with correct fish syntax
- [ ] Generate and install `.fish` files alongside `.sh` files
- [ ] Update `GenerateCompletionScript` for fish
- [ ] Update `Uninstall` to clean up fish files
- [ ] Write tests for fish function generation
- [ ] Manual QA on fish shell

## Blocked By
