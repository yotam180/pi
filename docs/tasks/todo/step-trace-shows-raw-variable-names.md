# Step trace lines show raw variable names instead of resolved values

## Type
bug

## Status
todo

## Priority
medium

## Project
standalone

## Description
When running an automation with inputs, the step trace line shows the raw environment variable reference instead of the resolved value.

### Steps to Reproduce
1. Create an automation with an input:
```yaml
name: build
inputs:
  profile:
    type: string
    default: dev
steps:
  - bash: cargo build --profile $PI_IN_PROFILE
```
2. Run: `pi run build --with profile=release`

### Expected
```
  → bash: cargo build --profile release
```

### Actual
```
  → bash: cargo build --profile $PI_IN_PROFILE
```

The command DOES execute correctly (release profile is used), but the trace output shows the unexpanded template. This makes it harder to debug what's actually running, especially with complex commands with multiple variables.

### Suggested Fix
Expand environment variable references in the trace output before printing. This is purely a display issue — execution is correct.

## Acceptance Criteria
- [ ] Step trace lines show resolved variable values, not raw `$PI_IN_*` references
- [ ] Complex commands with multiple variables all show resolved values
- [ ] Non-input env vars (like `$HOME`) are also expanded in trace (or at least `$PI_IN_*` vars are)

## Implementation Notes

## Subtasks
- [ ] 

## Blocked By
