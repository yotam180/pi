package executor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/display"
)

// execInstall runs the structured install lifecycle: test → [run → verify] → version.
func (e *Executor) execInstall(a *automation.Automation, inputEnv []string) error {
	inst := a.Install

	// Phase 1: test — exit 0 means already installed
	testErr := e.execInstallPhase(a, &inst.Test, inputEnv)
	if testErr == nil {
		version := e.captureVersion(a, inst.Version, inputEnv)
		e.printInstallStatus(display.StatusSuccessCached, a.Name, "already installed", version)
		return nil
	}

	// Phase 2: run — perform the installation
	// Stderr is streamed live to the terminal so the user sees errors as they happen.
	e.printInstallStatus(display.StatusInProgress, a.Name, "installing...", "")

	runErr := e.execInstallPhaseLive(a, &inst.Run, inputEnv)
	if runErr != nil {
		e.printInstallStatus(display.StatusFailed, a.Name, "failed", "")
		return runErr
	}

	// Phase 3: verify — confirm install succeeded (defaults to re-running test)
	verifyPhase := inst.Verify
	if verifyPhase == nil {
		verifyPhase = &inst.Test
	}
	verifyErr := e.execInstallPhase(a, verifyPhase, inputEnv)
	if verifyErr != nil {
		e.printInstallStatus(display.StatusFailed, a.Name, "failed", "")
		return verifyErr
	}

	version := e.captureVersion(a, inst.Version, inputEnv)
	e.printInstallStatus(display.StatusSuccess, a.Name, "installed", version)
	return nil
}

// execInstallPhase runs an install phase with stdout and stderr suppressed.
func (e *Executor) execInstallPhase(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string) error {
	return e.execInstallPhaseWithStderr(a, phase, inputEnv, io.Discard)
}

// execInstallPhaseLive runs an install phase with stdout suppressed but stderr
// streamed live to the terminal.
func (e *Executor) execInstallPhaseLive(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string) error {
	return e.execInstallPhaseWithStderr(a, phase, inputEnv, e.stderr())
}

// execInstallPhaseWithStderr runs an install phase with stdout suppressed and
// stderr routed to the given writer. Steps are dispatched through execStep,
// sharing the same condition evaluation, runner lookup, and output capture
// logic as the main execution loop.
func (e *Executor) execInstallPhaseWithStderr(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string, stderrWriter io.Writer) error {
	steps := phase.Steps
	if phase.IsScalar {
		steps = []automation.Step{{Type: automation.StepTypeBash, Value: phase.Scalar}}
	}

	ctx := &stepExecCtx{
		automation: a,
		inputEnv:   inputEnv,
	}

	saved := e.Outputs.Snapshot()
	e.Outputs.Reset()
	defer func() { e.Outputs.Restore(saved) }()

	for i, step := range steps {
		if step.If != "" {
			skip, err := e.evaluateCondition(step.If)
			if err != nil {
				return fmt.Errorf("install phase step[%d] if: %w", i, err)
			}
			if skip {
				continue
			}
		}

		if step.IsFirst() {
			if err := e.execInstallFirstBlock(ctx, step, i, stderrWriter); err != nil {
				return err
			}
			continue
		}

		if err := e.execStep(ctx, step, i, nil, false, io.Discard, stderrWriter); err != nil {
			return err
		}
	}
	return nil
}

// captureVersion runs the version command and returns the trimmed output.
func (e *Executor) captureVersion(a *automation.Automation, versionCmd string, inputEnv []string) string {
	if versionCmd == "" {
		return ""
	}

	const versionStepType = automation.StepTypeBash
	runner := e.registry().Get(versionStepType)
	if runner == nil {
		return ""
	}

	step := automation.Step{Type: versionStepType, Value: versionCmd}
	var buf bytes.Buffer
	rc := e.newRunContext(a, step, nil, &buf, nil, inputEnv, e.RepoRoot, io.Discard)

	if err := runner.Run(rc); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

// printInstallStatus prints a formatted installer status line unless Silent is set.
func (e *Executor) printInstallStatus(kind display.StatusKind, name, status, version string) {
	if e.Silent {
		return
	}
	e.printer().InstallStatus(kind, name, status, version)
}

// execInstallFirstBlock handles a first: block inside an install phase.
// Each sub-step is dispatched through execStep, sharing runner lookup and
// output capture with the main execution loop.
func (e *Executor) execInstallFirstBlock(ctx *stepExecCtx, step automation.Step, index int, stderrWriter io.Writer) error {
	for j, sub := range step.First {
		if sub.If != "" {
			skip, err := e.evaluateCondition(sub.If)
			if err != nil {
				return fmt.Errorf("install phase step[%d].first[%d] if: %w", index, j, err)
			}
			if skip {
				continue
			}
		}

		return e.execStep(ctx, sub, index, nil, false, io.Discard, stderrWriter)
	}
	return nil
}
