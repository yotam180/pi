# install-rust --version stable fails with semver parse error

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
Running `pi setup add rust --version stable` fails with:

```
invalid constraint "stable": improper constraint: ^stable.0.0
```

The Rust ecosystem uses channel names like `stable`, `nightly`, and `beta` alongside semver versions like `1.94`. The install-rust builtin correctly calls rustup and installs the toolchain, but then fails at the version constraint check because it tries to parse "stable" as a semver constraint (`^stable.0.0`).

This is particularly bad because a brand new Rust developer would naturally type `--version stable` — it's what rustup itself uses.

### Steps to Reproduce
1. `pi init --yes`
2. `pi setup add rust --version stable`

### Expected
The install should succeed, recognizing `stable` as a valid Rust channel name.

### Actual
The install runs (rustup succeeds) but pi reports failure because it can't validate the version.

### Suggested Fix
The version constraint system should recognize Rust-specific channel names (`stable`, `nightly`, `beta`) as valid values. When a channel name is given instead of a semver range:
- During install: pass the channel name directly to rustup (already works)
- During version check: verify that the installed version was obtained from the requested channel, OR just accept any installed version when the channel is "stable"

## Acceptance Criteria
- [x] `pi setup add rust --version stable` succeeds
- [x] `pi setup add rust --version nightly` succeeds
- [x] `pi setup add rust --version beta` succeeds
- [x] Semver versions like `1.94` continue to work as before

## Implementation Notes

### Approach
The fix is in the `semver` package, not in the Rust-specific builtin. This keeps it generic — any tool that uses channel names (e.g. `stable`, `nightly`, `beta`, `lts`, `latest`) will benefit.

### Changes
1. **`internal/semver/semver.go`**: Added `isChannelName()` function that detects purely lowercase-alpha strings (no digits, dots, operators). When `Satisfies()` receives a channel name as the constraint and the version is a valid semver, it returns nil (satisfied). The logic: a channel name is not a version constraint, so any installed version satisfies it.

2. **`internal/builtins/embed_pi/install-rust.yaml`**: Updated the input description to mention channel names (`stable`, `nightly`, `beta`) as valid values.

3. **Tests added**:
   - `TestSatisfies` — 4 new cases for channel name constraints
   - `TestIsChannelName` — 15 cases covering valid channels, digits, mixed, uppercase, empty, hyphenated
   - `TestDiscover_VersionSatisfiesChannelName` — integration test in builtins covering stable/nightly/beta through the GoFunc

### Design decision
Considered putting the channel-name logic in `version_satisfies.go` (the GoFunc), but chose `semver.go` because:
- It's the right abstraction layer (constraint parsing)
- It's generic — not Rust-specific
- Any future tool that uses channel names will just work
- The `isChannelName` check is a simple O(n) scan — no performance concern

## Subtasks
- [x] Add `isChannelName()` to semver package
- [x] Update `Satisfies()` to short-circuit on channel names
- [x] Add semver tests (unit + integration)
- [x] Update install-rust.yaml description
- [x] Full build + test pass

## Blocked By
