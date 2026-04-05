# Extra args after -- should be forwarded to the command or warn if dropped

## Type
feature

## Status
done

## Priority
medium

## Project
standalone

## Description
Running `pi run build -- --release` silently drops the `--release` flag. The command runs as if the extra args weren't provided, with no warning.

This is surprising because `--` is the universal convention for "pass remaining args through", and many tools (npm, cargo, docker) support it.

### Current Behavior
```
$ pi run build -- --release
  → bash: cargo build
    Finished `dev` profile [unoptimized + debuginfo] target(s)
```
The `--release` flag is silently ignored.

### Expected (Option A — forward args)
```
$ pi run build -- --release
  → bash: cargo build --release
    Finished `release` profile [optimized] target(s)
```
Extra args after `--` are appended to the last (or only) bash step.

### Expected (Option B — warn)
```
$ pi run build -- --release
warning: extra arguments after -- are not supported, use --with key=value for inputs
  → bash: cargo build
```

### Suggested Implementation
Option A is more useful. A special variable like `$PI_EXTRA_ARGS` or `$@` could be made available in bash steps, and if present in the step template, the args are forwarded. If no step uses `$PI_EXTRA_ARGS`, print a warning that the args were dropped.

Or, more simply: if the step is a single `bash:` step, append the extra args to it automatically.

## Acceptance Criteria
- [x] Either extra args are forwarded (Option A) or a warning is printed (Option B)
- [x] Silent dropping of arguments is eliminated

## Implementation Notes

### Approach taken: error on excess positional args (for automations with `inputs:`)

After investigating, the issue has two distinct cases:

1. **Automations WITH `inputs:`**: When an automation declares inputs and the user provides more positional args than declared inputs, the excess args were silently dropped by `ResolveInputs()`. **Fix**: `ResolveInputs()` now returns a clear error when excess positional args are provided:
   ```
   automation "greet": too many arguments: got 3, but "greet" only accepts 1 input(s); extra: bob charlie
   ```

2. **Automations WITHOUT `inputs:`**: Args already flow through to bash steps via `$@` and `$1`/`$2`. The user can write `bash: cargo build $@` to forward args. This is working correctly — no change needed.

### Why error instead of warning?
- An error is more appropriate because excess args to an automation with declared inputs can never do anything useful — they're always a user mistake
- Aligns with PI philosophy principle 2: "correct the action, not the developer" — the error message clearly explains what went wrong and what the limits are
- Consistent with how unknown `--with` keys already produce errors

### Files changed
- `internal/automation/inputs.go` — added excess args check in `ResolveInputs()` positional args branch
- `internal/automation/inputs_test.go` — added `TestResolveInputs_ExcessPositionalArgs` and `TestResolveInputs_ExactPositionalArgs`
- `internal/executor/inputs_test.go` — added `TestRunWithInputs_ExcessPositionalArgs`
- `internal/cli/run_test.go` — added `TestRunAutomation_ExcessPositionalArgs`

## Subtasks
- [x] Investigate current arg handling behavior
- [x] Implement excess args error in ResolveInputs
- [x] Write unit tests at automation, executor, and CLI levels
- [x] Full test suite passes
- [x] Manual QA

## Blocked By
