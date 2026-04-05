package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
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
	e.printInstallStatus("→", a.Name, "installing...", "")

	stderrBuf := &bytes.Buffer{}
	runErr := e.execInstallPhaseCapture(a, &inst.Run, inputEnv, stderrBuf)
	if runErr != nil {
		e.printInstallStatus("✗", a.Name, "failed", "")
		if stderrBuf.Len() > 0 {
			e.printIndentedStderr(stderrBuf.String())
		}
		return runErr
	}

	// Phase 3: verify — confirm install succeeded (defaults to re-running test)
	verifyPhase := inst.Verify
	if verifyPhase == nil {
		verifyPhase = &inst.Test
	}
	verifyErr := e.execInstallPhase(a, verifyPhase, inputEnv)
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

// execInstallPhaseCapture runs an install phase, suppressing stdout and optionally
// capturing stderr to the given buffer (in addition to suppressing it).
func (e *Executor) execInstallPhaseCapture(a *automation.Automation, phase *automation.InstallPhase, inputEnv []string, stderrCapture *bytes.Buffer) error {
	if phase.IsScalar {
		return e.execBashSuppressed(a, phase.Scalar, inputEnv, stderrCapture)
	}

	for i, step := range phase.Steps {
		if step.If != "" {
			skip, err := e.evaluateCondition(step.If)
			if err != nil {
				return fmt.Errorf("install phase step[%d] if: %w", i, err)
			}
			if skip {
				continue
			}
		}

		switch step.Type {
		case automation.StepTypeBash:
			if err := e.execBashSuppressed(a, step.Value, inputEnv, stderrCapture); err != nil {
				return err
			}
		case automation.StepTypeRun:
			target, err := e.Discovery.Find(step.Value)
			if err != nil {
				return fmt.Errorf("install phase run step: %w", err)
			}
			origStdout, origStdin, origSilent := e.Stdout, e.Stdin, e.Silent
			e.Stdout, e.Stdin = io.Discard, nil
			var runErr error
			if len(step.With) > 0 {
				runErr = e.RunWithInputs(target, nil, step.With)
			} else {
				runErr = e.RunWithInputs(target, nil, nil)
			}
			e.Stdout, e.Stdin, e.Silent = origStdout, origStdin, origSilent
			if runErr != nil {
				return runErr
			}
		case automation.StepTypePython:
			if err := e.execScriptSuppressed(a, "python", step.Value, inputEnv, stderrCapture); err != nil {
				return err
			}
		case automation.StepTypeTypeScript:
			if err := e.execScriptSuppressed(a, "typescript", step.Value, inputEnv, stderrCapture); err != nil {
				return err
			}
		default:
			return fmt.Errorf("install phase step[%d]: unsupported step type %q", i, step.Type)
		}
	}
	return nil
}

// execBashSuppressed runs inline bash with stdout suppressed and stderr optionally captured.
func (e *Executor) execBashSuppressed(a *automation.Automation, script string, inputEnv []string, stderrCapture *bytes.Buffer) error {
	var cmdArgs []string
	if isFilePath(script) {
		resolved := resolveScriptPath(a.Dir(), script)
		cmdArgs = []string{resolved}
	} else {
		cmdArgs = []string{"-c", script}
	}

	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = io.Discard
	cmd.Stdin = nil
	cmd.Env = e.buildEnv(inputEnv, nil)

	if stderrCapture != nil {
		cmd.Stderr = stderrCapture
	} else {
		cmd.Stderr = io.Discard
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("running install bash: %w", err)
	}
	return nil
}

// execScriptSuppressed runs a python or typescript step with stdout suppressed.
func (e *Executor) execScriptSuppressed(a *automation.Automation, lang, script string, inputEnv []string, stderrCapture *bytes.Buffer) error {
	step := automation.Step{Value: script}
	switch lang {
	case "python":
		step.Type = automation.StepTypePython
	case "typescript":
		step.Type = automation.StepTypeTypeScript
	}

	origStdout, origStderr, origStdin := e.Stdout, e.Stderr, e.Stdin
	e.Stdout = io.Discard
	if stderrCapture != nil {
		e.Stderr = stderrCapture
	} else {
		e.Stderr = io.Discard
	}
	e.Stdin = nil

	var err error
	switch step.Type {
	case automation.StepTypePython:
		err = e.execPython(a, step, nil, e.Stdout, e.Stdin, inputEnv)
	case automation.StepTypeTypeScript:
		err = e.execTypeScript(a, step, nil, e.Stdout, e.Stdin, inputEnv)
	}

	e.Stdout, e.Stderr, e.Stdin = origStdout, origStderr, origStdin
	return err
}

// captureVersion runs the version command and returns the trimmed output.
func (e *Executor) captureVersion(a *automation.Automation, versionCmd string, inputEnv []string) string {
	if versionCmd == "" {
		return ""
	}

	var buf bytes.Buffer
	cmd := exec.Command("bash", "-c", versionCmd)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	cmd.Env = e.buildEnv(inputEnv, nil)

	if err := cmd.Run(); err != nil {
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

// printIndentedStderr prints stderr output indented for the error block.
func (e *Executor) printIndentedStderr(text string) {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		fmt.Fprintf(e.stderr(), "      %s\n", line)
	}
}
