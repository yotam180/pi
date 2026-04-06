# Natural argument passing to automations (positional params + drop --)

## Type
feature

## Status
done

## Priority
medium

## Project
standalone

## Description
`pi run build -- --release` is clumsy. The `--` separator is a POSIX convention for ending flag parsing, not how developers think about running commands. The goal is to let developers pass arguments to automations as naturally as possible.

Two related improvements:

### 1. Positional parameters

Automations should support positional inputs — unnamed arguments that map to inputs by position:

```yaml
name: build
inputs:
  profile:
    type: string
    default: dev
    positional: 1       # accepts it as the first bare argument
```

This would let you write:
```
pi run build release
```
instead of:
```
pi run build --with profile=release
```

For automations with a single input, `positional: 1` should probably be the default or auto-inferred. The `--with` syntax stays supported for explicitness and for use in YAML `run:` steps.

### 2. Drop the -- requirement (non-backwards-compatible)

Extra bare arguments after the automation name should be forwarded to the automation's bash steps via `$PI_ARGS`, without requiring `--`:

```
pi run build release
  → bash: cargo build --profile release
```

#### Handling conflicts with pi's own flags

`pi run` has flags of its own (`--silent`, `--loud`, `--repo`, `--with`). The key insight is that **pi flags and automation arguments don't need to share the same namespace** if pi's flags are required to come before the automation name:

```
pi run --silent build release
#      ^^^^^^^^ pi flag (before automation name)
#               ^^^^^ automation name
#                     ^^^^^^^ automation argument
```

This is the same convention `docker run` uses: `docker run --rm nginx nginx -g "daemon off;"` — Docker's own flags come before the image name, everything after is for the container.

So the rule would be:
- **Before `<automation>`:** flags are parsed as `pi run` flags (`--silent`, `--loud`, `--repo`, `--with`)
- **After `<automation>`:** everything is passed as automation arguments, no `--` needed
- `--` still works as an explicit separator for the rare case where an automation name could be ambiguous

#### Interaction with `--with`

`--with` still works for named inputs and should take precedence over positional:
```
pi run build release               # positional: profile=release
pi run --with profile=release build # explicit: same result
```

#### Shell shortcuts

Shortcuts should inherit this behavior naturally — since shortcuts call `pi run <automation> "$@"`, any args passed to the shortcut function are forwarded:

```bash
build release      # calls pi run build release
build              # calls pi run build (uses default: dev)
```

## Acceptance Criteria
- [x] `pi run build release` works when `profile` is a positional input
- [x] `pi run build release` works without `--` and without explicit `--with`
- [x] `pi run --silent build release` correctly applies `--silent` to pi, not to the automation
- [x] `--with key=value` still works and takes precedence over positional args
- [x] `--` still works as an explicit separator for backwards compatibility
- [x] `pi info build` shows which inputs are positional and in what order
- [x] Shell shortcuts transparently forward positional args

## Implementation Notes

### Approach: Cobra's SetInterspersed(false)

The solution was elegant: Cobra's `cmd.Flags().SetInterspersed(false)` tells the flag parser to stop processing flags when it encounters the first non-flag argument. Since the automation name is always the first non-flag argument, this means:

- `pi run --silent build release --verbose` → `--silent` parsed as pi flag, `build` is args[0], `release --verbose` are args[1:]
- `pi run build --release` → `build` is args[0], `--release` is args[1] (forwarded as automation arg)

This follows the same convention as `docker run`, `kubectl exec`, etc.

### No `positional:` field needed

The original spec proposed a `positional: N` field on InputSpec. This turned out to be unnecessary — inputs already have a natural positional order from YAML declaration order, and `ResolveInputs()` already maps positional args to inputs by that order. All inputs are positional by default.

### Breaking change: `--with` placement

`--with` must now come **before** the automation name, not after. This is a behavior change but is consistent with how all other pi flags work. Updated all integration tests.

### `--` backwards compat caveat

With `SetInterspersed(false)`, a literal `--` after the automation name is passed through as an argument (it's no longer consumed by Cobra). So `pi run build -- --release` would set PI_ARGS to `-- --release`. Since the `--` is no longer needed, developers should simply write `pi run build --release`. The `--` before the automation name still works as a Cobra separator.

### `pi info` improvements

- Added `position N` to each input's metadata display
- Added a `Usage:` line showing `pi run <name> <required> [optional]` syntax

### Files changed

- `internal/cli/run.go` — SetInterspersed(false), updated help text
- `internal/cli/info.go` — position display, printUsageLine()
- `internal/cli/run_test.go` — 8 new tests for arg forwarding and Cobra flag isolation
- `internal/cli/info_test.go` — updated format assertions, 3 new tests for usage line
- `tests/integration/inputs_test.go` — moved --with before automation name, dropped --
- `tests/integration/installer_schema_test.go` — moved --with before automation name
- `tests/integration/info_test.go` — updated format assertion for position display
- `docs/README.md` — updated PI_ARGS docs, CLI reference, added positional input examples
- `docs/architecture.md` — updated run.go and info.go descriptions

## Subtasks
- [x] ~~Design the `positional:` input field~~ Not needed — declaration order is sufficient
- [x] Update `pi run` arg parsing to stop at the automation name (SetInterspersed(false))
- [x] Pass post-name args as positional inputs, then fall through to `$PI_ARGS` (already worked)
- [x] Update `pi info` to show positional order
- [x] Update `pi run --help` to document the convention
- [x] ~~Update shell shortcut codegen to pass `"$@"`~~ Already passes `"$@"`

## Blocked By
