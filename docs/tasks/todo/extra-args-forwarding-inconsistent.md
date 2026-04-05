# Extra args via -- are inconsistently forwarded

## Type
bug

## Status
todo

## Priority
medium

## Project
standalone

## Description
When passing extra arguments after `--` to `pi run`, the behavior is inconsistent between automations that have `inputs:` defined and those that don't.

### Observed behavior

**Automation with inputs (build):** args ARE appended, but the input default is dropped:
```
# build.yaml has: inputs.profile (default: dev), step: cargo build --profile $PI_IN_PROFILE
$ pi run build -- --release
  → bash: cargo build --profile --release
# Expected: cargo build --profile dev --release
# The default "dev" for $PI_IN_PROFILE was dropped, leaving --profile with no value
```

**Automation without inputs (test):** args are NOT appended at all:
```
# test.yaml has: step: cargo test
$ pi run test -- --ignored
  → bash: cargo test
# Expected: cargo test --ignored
```

### Expected behavior
Either:
- **Option A:** `--` args are always appended to the last bash step, regardless of whether the automation has inputs
- **Option B:** `--` args are never forwarded, and a warning is printed explaining to use `--with`

The current mix of "sometimes forwarded, sometimes not, and when forwarded they break input defaults" is the worst outcome.

### Additional concern
When `--` args are forwarded to an automation with inputs, the input defaults should still be applied. `$PI_IN_PROFILE` should expand to `dev` even when extra args are present, giving `cargo build --profile dev --release`.

## Acceptance Criteria
- [ ] `pi run <automation> -- <args>` behaves consistently whether or not the automation has `inputs:`
- [ ] When args are forwarded, input defaults are still applied (not dropped)
- [ ] Behavior is documented in `pi run --help`

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
