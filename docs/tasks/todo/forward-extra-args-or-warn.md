# Extra args after -- should be forwarded to the command or warn if dropped

## Type
feature

## Status
todo

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
- [ ] Either extra args are forwarded (Option A) or a warning is printed (Option B)
- [ ] Silent dropping of arguments is eliminated

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
