package executor

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
)

// execInstall runs the structured install lifecycle: test → [run → verify] → version.
func (e *Executor) execInstall(a *automation.Automation, inputEnv []string) error {
	inst := a.Install

	// Phase 1: test — exit 0 means already installed
	testErr := e.execInstallPhase(a, &inst.Test, inputEnv)
	if testErr == nil {
		version := e.captureVersion(a, inst.Version, inputEnv)
		e.printInstallStatus("✓", a.Name, "already installed", version)
		return nil
	}

	// Phase 2: run — perform the installation
	// Stderr is streamed live to the terminal so the user sees errors as they happen.
	e.printInstallStatus("→", a.Name, "installing...", "")

	runErr := e.execInstallPhaseLive(a, &inst.Run, inputEnv)
	if runErr != nil {
		e.printInstallStatus("✗", a.Name, "failed", "")
		return runErr
	}

	// Phase 3: verify — confirm install succeeded (defaults to re-running test)
	verifyPhase := inst.Verify
	if verifyPhase == nil {
		verifyPhase = &inst.Test
	}
	verifyErr := e.execInstallPhaseCapture(a, verifyPhase, inputEnv, nil)
	if verifyErr != nil {
		e.printInstallStatus("✗", a.Name, "failed", "")
		return verifyErr
	}

	version := e.captureVersion(a, inst.Version, inputEnv)
	e.printInstallStatus("✓", a.Name, "installed", version)
	return nil
}

// execInstallPhase runs an install phase with stdout and stderr suppressed.
func (e *Executor) execInstallPhase(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string) error {
	return e.execInstallPhaseCapture(a, phase, inputEnv, nil)
}

// execInstallPhaseLive runs an install phase with stdout suppressed but stderr
// streamed live to the terminal. This gives the user immediate visibility into
// errors and progress from install commands.
func (e *Executor) execInstallPhaseLive(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string) error {
	steps := phase.Steps
	if phase.IsScalar {
		steps = []automation.Step{{Type: automation.StepTypeBash, Value: phase.Scalar}}
	}

	savedOutputs := e.stepOutputs
	e.stepOutputs = nil
	defer func() { e.stepOutputs = savedOutputs }()

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
			if err := e.execInstallFirstBlockLive(a, step, i, inputEnv); err != nil {
				return err
			}
			continue
		}

		runner := e.registry().Get(step.Type)
		if runner == nil {
			return fmt.Errorf("install phase step[%d]: unsupported step type %q", i, step.Type)
		}

		var outputCapture bytes.Buffer
		ctx := e.newRunContext(a, step, nil, &outputCapture, nil, inputEnv)
		ctx.Stderr = e.stderr()

		if err := runner.Run(ctx); err != nil {
			return err
		}
		e.recordOutput(outputCapture.String())
	}
	return nil
}

// execInstallPhaseCapture runs an install phase, suppressing stdout and optionally
// capturing stderr to the given buffer (in addition to suppressing it).
func (e *Executor) execInstallPhaseCapture(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string, stderrCapture *bytes.Buffer) error {
	steps := phase.Steps
	if phase.IsScalar {
		steps = []automation.Step{{Type: automation.StepTypeBash, Value: phase.Scalar}}
	}

	savedOutputs := e.stepOutputs
	e.stepOutputs = nil
	defer func() { e.stepOutputs = savedOutputs }()

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
			if err := e.execInstallFirstBlock(a, step, i, inputEnv, stderrCapture); err != nil {
				return err
			}
			continue
		}

		stderrWriter := io.Writer(io.Discard)
		if stderrCapture != nil {
			stderrWriter = stderrCapture
		}

		runner := e.registry().Get(step.Type)
		if runner == nil {
			return fmt.Errorf("install phase step[%d]: unsupported step type %q", i, step.Type)
		}

		var outputCapture bytes.Buffer
		ctx := e.newRunContext(a, step, nil, &outputCapture, nil, inputEnv)
		ctx.Stderr = stderrWriter

		if err := runner.Run(ctx); err != nil {
			return err
		}
		e.recordOutput(outputCapture.String())
	}
	return nil
}

// captureVersion runs the version command and returns the trimmed output.
func (e *Executor) captureVersion(a *automation.Automation, versionCmd string, inputEnv []string) string {
	if versionCmd == "" {
		return ""
	}

	step := automation.Step{Type: automation.StepTypeBash, Value: versionCmd}
	var buf bytes.Buffer
	ctx := e.newRunContext(a, step, nil, &buf, nil, inputEnv)
	ctx.Stderr = io.Discard

	runner := e.registry().Get(automation.StepTypeBash)
	if runner == nil {
		return ""
	}
	if err := runner.Run(ctx); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

// printInstallStatus prints a formatted installer status line unless Silent is set.
func (e *Executor) printInstallStatus(icon, name, status, version string) {
	if e.Silent {
		return
	}
	e.printer().InstallStatus(icon, name, status, version)
}

// execInstallFirstBlockLive handles a first: block inside an install phase,
// streaming stderr live to the terminal.
func (e *Executor) execInstallFirstBlockLive(a *automation.Automation, step automation.Step, index int, inputEnv []string) error {
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

		runner := e.registry().Get(sub.Type)
		if runner == nil {
			return fmt.Errorf("install phase step[%d].first[%d]: unsupported step type %q", index, j, sub.Type)
		}

		ctx := e.newRunContext(a, sub, nil, io.Discard, nil, inputEnv)
		ctx.Stderr = e.stderr()

		return runner.Run(ctx)
	}
	return nil
}

// execInstallFirstBlock handles a first: block inside an install phase.
func (e *Executor) execInstallFirstBlock(a *automation.Automation, step automation.Step, index int, inputEnv []string, stderrCapture *bytes.Buffer) error {
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

		stderrWriter := io.Writer(io.Discard)
		if stderrCapture != nil {
			stderrWriter = stderrCapture
		}

		runner := e.registry().Get(sub.Type)
		if runner == nil {
			return fmt.Errorf("install phase step[%d].first[%d]: unsupported step type %q", index, j, sub.Type)
		}

		ctx := e.newRunContext(a, sub, nil, io.Discard, nil, inputEnv)
		ctx.Stderr = stderrWriter

		return runner.Run(ctx)
	}
	return nil
}

