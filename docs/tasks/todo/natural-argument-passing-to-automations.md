# Natural argument passing to automations (positional params + drop --)

## Type
feature

## Status
todo

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
pi run build --with profile=release  # explicit: same result
```

#### Shell shortcuts

Shortcuts should inherit this behavior naturally — since shortcuts call `pi run <automation> "$@"`, any args passed to the shortcut function are forwarded:

```bash
build release      # calls pi run build release
build              # calls pi run build (uses default: dev)
```

## Acceptance Criteria
- [ ] `pi run build release` works when `profile` is a positional input
- [ ] `pi run build release` works without `--` and without explicit `--with`
- [ ] `pi run --silent build release` correctly applies `--silent` to pi, not to the automation
- [ ] `--with key=value` still works and takes precedence over positional args
- [ ] `--` still works as an explicit separator for backwards compatibility
- [ ] `pi info build` shows which inputs are positional and in what order
- [ ] Shell shortcuts transparently forward positional args

## Implementation Notes
This is a breaking change only for the edge case where someone passes a flag-shaped string as the automation name. In practice this is rare/impossible. The `--` separator can remain supported but optional.

Consider documenting the before-vs-after convention (pi flags before name, automation args after) clearly in `pi run --help`.

## Subtasks
- [ ] Design the `positional:` input field (or auto-infer for single-input automations)
- [ ] Update `pi run` arg parsing to stop at the automation name
- [ ] Pass post-name args as positional inputs, then fall through to `$PI_ARGS`
- [ ] Update `pi info` to show positional order
- [ ] Update `pi run --help` to document the convention
- [ ] Update shell shortcut codegen to pass `"$@"`

## Blocked By
