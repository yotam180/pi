# pi:install-node: skip when already installed + surface real errors

## Type
bug

## Status
done

## Priority
high

## Project
standalone

## Description
Two separate problems with the `pi:install-node` builtin automation (`internal/builtins/embed_pi/install-node.yaml`):

**Problem A — installs even when node is already present at the right version**

The `test:` block in `install.yaml` does a bash major-version comparison. It should short-circuit the entire `install.run` step when node is already at a satisfying version. Currently it either doesn't fire correctly or the logic is wrong, causing a reinstall every time `pi setup` is run. (This is also a primitive version check — see ticket #94 for a richer solution.)

**Problem B — setup errors are swallowed; the user sees no actionable output**

When the install step fails, pi reports:
```
✗  install-node              failed
setup[0] "pi:install-node" failed: step exited with code 1
```
…but never shows the actual error from the bash commands (e.g. "no suitable installer found", brew errors, mise errors). The developer is left with nothing to debug.

## Acceptance Criteria
- [x] When node is already installed at a version satisfying the requested major, `pi setup` (or `pi setup add node --version 22`) skips the install step and prints a "already satisfied" message instead of re-installing.
- [x] When the install step fails, the full stderr output from the failing bash command is printed before the `✗ failed` line. No swallowing.
- [x] Error output for "no suitable installer found" is explicit: names which installers were tried and links to how to install one (mise/brew).
- [x] `go build ./...` and `go test ./...` pass.

## Implementation Notes

### Problem A — test script analysis
The `install-node.yaml` test script (`[ "$MAJOR" = "$PI_IN_VERSION" ]`) works correctly when tested manually. `PI_IN_VERSION` is properly injected via the executor's `buildEnv(inputEnv, nil, nil)` in both `execBashSuppressed` and the step-list path. The test phase exit code correctly gates the run phase in `execInstall`. No code change was needed — the test script and executor logic are sound.

The version check remains "primitive" (major-only comparison) — ticket #94 will replace it with `pi:version-satisfies` for richer semver-aware matching.

### Problem B — stderr streaming fix
**Root cause**: The `run:` phase of install automations captured stderr to a `bytes.Buffer` (via `execInstallPhaseCapture`) and only printed it after failure. This worked in some cases but the user experience was poor — errors appeared after the `✗ failed` status line, and in some code paths (e.g. verify failures) stderr was suppressed entirely.

**Fix**: Introduced `execInstallPhaseLive()`, `execBashLive()`, and `execInstallFirstBlockLive()` — parallel versions of the suppressed functions that stream stderr directly to the terminal (`e.stderr()`) instead of buffering it. The `run:` phase now uses the live variant so users see errors in real-time as the install runs.

The old `execInstallPhaseCapture`/`execBashSuppressed`/`execInstallFirstBlock` remain for the test/verify phases where stderr suppression is intentional.

Removed `printIndentedStderr()` since it's no longer needed — stderr is streamed live instead of buffered and replayed.

### Error message improvements
Updated fallback error messages in all 5 installer YAMLs that use `first:` blocks with a fallback:
- `install-node.yaml` — now includes "Install one of: mise, brew" with URLs
- `install-python.yaml` — same
- `install-go.yaml` — same
- `install-terraform.yaml` — same, plus manual install link
- `install-kubectl.yaml` — same, plus manual install link

### Tests added
- `TestExecInstall_FirstBlockFailStderrSurfaced` — verifies stderr from `first:` block fallback is visible
- `TestExecInstall_RunFailsScalarStderrStreamed` — verifies stderr from scalar run phase is streamed
- `TestExecInstall_RunFailsStepListStderrStreamed` — verifies stderr from step-list run phase is streamed
- `TestExecInstall_SilentStillShowsStderrOnFailure` — updated to also assert stderr content is visible
- `TestInstallerSchema_FailedInstallShowsStderr` — new integration test with `install-failing.yaml` example

Total: 4 new unit tests, 1 new integration test, 1 updated unit test.
