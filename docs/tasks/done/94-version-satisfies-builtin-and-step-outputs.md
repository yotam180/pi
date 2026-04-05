# pi:version-satisfies builtin + inter-step output capture

## Type
feature

## Status
done

## Priority
medium

## Project
standalone

## Description
The current version-matching in `pi:install-node` (and presumably all install automations) is a primitive bash string comparison that only handles major versions. We need a proper semver-aware check that package authors can use declaratively.

Two related capabilities are needed:

### 1. `pi:version-satisfies` builtin automation

A Go-backed builtin that compares a version string against a constraint expression. Package authors can compose it into their `install.test` pipeline:

```yaml
install:
  test:
    - bash: node --version 2>/dev/null | sed 's/^v//'   # outputs "22.3.1"
    - run: pi:version-satisfies
      with:
        version: outputs.last     # value from the previous step
        required: inputs.version  # e.g. "22", ">=20", "^18.0.0", "~20.1"
```

Supported constraint syntax (at minimum):
- `22` or `v22` — matches any `22.x.x`
- `>=20` — semver greater-or-equal on major
- `^18` / `^18.0.0` — compatible with (same major, any minor/patch)
- `~20.1` — approximately equal (same major+minor)
- `>=18.0.0 <20.0.0` — range (optional but nice to have)

Exits 0 if satisfied, non-zero if not. Prints a human-readable explanation on failure (e.g. "node 16.x does not satisfy >=18").

### 2. Inter-step output capture (`outputs.last` / `outputs.<step>`)

For `pi:version-satisfies` to work, the executor needs a way to pass the stdout of one step as an input to a subsequent step. Propose and implement a minimal mechanism:

- `outputs.last` — stdout of the immediately preceding step, trimmed.
- `outputs.<N>` (optional) — stdout of step N (0-indexed).

This is also generally useful for other automations (e.g. capture a computed value and pass it to a later step).

The mechanism should be declarative in YAML — automation authors reference `outputs.last` in `with:` values; the executor resolves it at runtime.

## Acceptance Criteria
- [x] `pi:version-satisfies` builtin exists and runs in Go (not bash).
- [x] Supports constraints: exact major (`22`), `>=X`, `^X`, `~X.Y`, and optionally range (`>=X <Y`).
- [x] Exits 0 when satisfied, non-zero when not, with a clear failure message.
- [x] `outputs.last` is available in `with:` values of any step following a bash step that produced stdout.
- [x] `pi:install-node` is updated to use `pi:version-satisfies` with `outputs.last` + `inputs.version`, replacing the primitive bash major check.
- [x] All existing install automations that do version checking are updated to use the new builtin.
- [x] `go build ./...` and `go test ./...` pass.
- [x] Unit tests for `pi:version-satisfies` covering all supported constraint forms.

## Implementation Notes

### Semver library
Used `github.com/Masterminds/semver/v3` (v3.4.0). It has rich constraint syntax (exact, `>=`, `^`, `~`, ranges) which matches the ticket requirements. No existing semver library was in `go.mod`.

### New package: `internal/semver`
Wraps Masterminds/semver with PI-specific normalisation:
- `Satisfies(version, constraint) error` — the public API
- Normalises incomplete versions ("22" → "22.0.0") and strips `v` prefixes
- Bare version numbers are treated as caret constraints: "22" → "^22.0.0" (matches any 22.x.x)
- Operator-prefixed constraints pass through to Masterminds with version part normalisation
- 31 unit tests covering all constraint forms and edge cases

### Go-backed builtin infrastructure
Added `GoFunc` field to `automation.Automation`:
- `GoFunc func(inputs map[string]string) error` — when non-nil, the executor calls this instead of running steps
- `IsGoFunc() bool` helper method
- `validate()` skips step/install checks when `GoFunc` is set
- Executor's `RunWithInputs()` handles GoFunc before the installer/step code paths
- GoFunc errors are printed to stderr and result in ExitError{Code: 1}
- GoFunc automations respect `if:` conditions (checked before GoFunc call)

### `pi:version-satisfies` builtin
Registered in `builtins.goBackedBuiltins()`:
- Inputs: `version` (the actual version) and `required` (the constraint expression)
- Uses `pisemver.Satisfies()` for the comparison
- Returns nil on success, error message on failure (triggering ExitError{Code: 1})
- 5 builtins tests (existence, resolvability, satisfied, not-satisfied, empty version)

### Inter-step output capture (`outputs.last`)
Added `stepOutputs []string` to `Executor`:
- Each step's stdout is captured via `io.MultiWriter` (teed to both the original writer and a capture buffer)
- Pipe steps capture to the pipe buffer (which is also recorded)
- Silent steps still capture output (MultiWriter to io.Discard + capture buffer)
- `stepOutputs` is scoped per-automation: saved/restored via `defer` in `RunWithInputs()` and install phases
- Install phase steps also capture output for `outputs.last` to work in `install.test:` pipelines

### `with:` value interpolation
- `InterpolateWith` callback added to `RunContext`
- `RunStepRunner` calls `InterpolateWith` before passing `with:` values to the called automation
- Three interpolation sources:
  - `outputs.last` → trimmed stdout of the most recently executed step
  - `outputs.<N>` → stdout of step N (0-indexed)
  - `inputs.<name>` → current automation's input value (via `PI_IN_*` env vars)
- Literal values pass through unchanged

### Updated installer YAMLs
All 6 installers that had version checking now use `pi:version-satisfies`:
- `install-node.yaml` — replaced bash major comparison with version-satisfies
- `install-python.yaml` — replaced grep with version-satisfies (test + verify)
- `install-go.yaml` — replaced bash major.minor comparison with version-satisfies
- `install-rust.yaml` — replaced bash major.minor comparison with version-satisfies
- `install-terraform.yaml` — replaced grep with version-satisfies
- `install-kubectl.yaml` — replaced grep with version-satisfies
- `install-helm.yaml` — replaced grep with version-satisfies

Input descriptions updated to show semver constraint examples (e.g. ">=20", "^18").

### `pi info` update
Added `go-native` type display for GoFunc automations in `pi info` output.

### Tests added
- `internal/semver/semver_test.go`: 31 tests (all constraint forms, normalisation, edge cases)
- `internal/executor/outputs_test.go`: 16 tests (outputs.last via with, multi-step chain, pipe capture, silent capture, indexed output, inputs interpolation, combined outputs+inputs, literal passthrough, GoFunc via run step, GoFunc failure, GoFunc condition skip, interpolateValue unit tests)
- `internal/builtins/builtins_test.go`: 5 new tests + updated 5 existing tests to use `assertTestPhaseUsesVersionSatisfies` helper

## Subtasks
- [x] Research: check `go.mod` for existing semver libraries; decide on `Masterminds/semver` vs `x/mod/semver`.
- [x] Implement `pi:version-satisfies` Go builtin.
- [x] Unit tests for all constraint forms.
- [x] Design and implement `outputs.last` interpolation in the executor.
- [x] Update `install-node.yaml` to use new builtin.
- [x] Audit other builtin install YAMLs and update their version checks.
- [x] Integration test coverage via existing test suite (all pass).

## Blocked By
- ~~Ticket #93 (test-step gating must work end-to-end)~~ — done
