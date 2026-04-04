package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// Executor runs automation steps within a project.
type Executor struct {
	// RepoRoot is the project root (directory containing pi.yaml).
	// All steps run with this as their working directory.
	RepoRoot string

	// Discovery holds the resolved automations for run: step lookups.
	Discovery *discovery.Result

	// Stdout and Stderr control where step output goes.
	// Defaults to os.Stdout and os.Stderr if nil.
	Stdout io.Writer
	Stderr io.Writer

	// Stdin provides input to the first step (or any step not receiving piped data).
	// Defaults to os.Stdin if nil.
	Stdin io.Reader

	// Silent suppresses PI-managed status lines for installer automations.
	// Stderr from failed install steps is always shown regardless of this flag.
	Silent bool

	// callStack tracks the chain of automation names currently being executed,
	// used to detect circular run: dependencies.
	callStack []string

	// lastPipeBuffer holds captured stdout from the last pipe_to:next step.
	lastPipeBuffer *bytes.Buffer

	// RuntimeEnv overrides the default runtime environment for predicate resolution.
	// If nil, DefaultRuntimeEnv() is used.
	RuntimeEnv *RuntimeEnv
}

// ExitError wraps a non-zero exit code from a step.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("step exited with code %d", e.Code)
}

// Run executes all steps of the given automation in order.
// args are passed to bash steps as $1, $2, etc., or mapped to declared inputs.
// When a step declares pipe_to: next, its stdout is captured and fed as stdin
// to the following step.
func (e *Executor) Run(a *automation.Automation, args []string) error {
	return e.RunWithInputs(a, args, nil)
}

// RunWithInputs executes all steps with explicit --with input values.
// Only one of args (positional) or withArgs may be non-empty.
func (e *Executor) RunWithInputs(a *automation.Automation, args []string, withArgs map[string]string) error {
	if a.If != "" {
		skip, err := e.evaluateCondition(a.If)
		if err != nil {
			return fmt.Errorf("automation %q if: %w", a.Name, err)
		}
		if skip {
			fmt.Fprintf(e.stderr(), "[skipped] %s (condition: %s)\n", a.Name, a.If)
			return nil
		}
	}

	if err := e.pushCall(a.Name); err != nil {
		return err
	}
	defer e.popCall()

	var inputEnv []string
	if len(a.Inputs) > 0 {
		resolved, err := a.ResolveInputs(withArgs, args)
		if err != nil {
			return fmt.Errorf("automation %q: %w", a.Name, err)
		}
		inputEnv = automation.InputEnvVars(resolved)
		args = nil
	}

	if a.IsInstaller() {
		return e.execInstall(a, inputEnv)
	}

	var pipedInput io.Reader
	for i, step := range a.Steps {
		if step.If != "" {
			skip, err := e.evaluateCondition(step.If)
			if err != nil {
				return fmt.Errorf("automation %q step[%d] if: %w", a.Name, i, err)
			}
			if skip {
				if step.PipeTo == "next" && i < len(a.Steps)-1 {
					if pipedInput != nil {
						buf := &bytes.Buffer{}
						if _, err := io.Copy(buf, pipedInput); err != nil {
							return fmt.Errorf("passing piped input through skipped step[%d]: %w", i, err)
						}
						pipedInput = buf
					}
				} else {
					pipedInput = nil
				}
				continue
			}
		}

		isPipeSrc := step.PipeTo == "next" && i < len(a.Steps)-1
		if err := e.execStep(a, step, args, i, pipedInput, isPipeSrc, inputEnv); err != nil {
			return err
		}
		pipedInput = nil
		if isPipeSrc {
			pipedInput = e.lastPipeBuffer
		}
	}
	return nil
}

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
	cmd.Env = appendInputEnv(inputEnv)

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
	cmd.Env = appendInputEnv(inputEnv)

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
	if version != "" {
		fmt.Fprintf(e.stderr(), "  %s  %-25s %s (%s)\n", icon, name, status, version)
	} else {
		fmt.Fprintf(e.stderr(), "  %s  %-25s %s\n", icon, name, status)
	}
}

// printIndentedStderr prints stderr output indented for the error block.
func (e *Executor) printIndentedStderr(text string) {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		fmt.Fprintf(e.stderr(), "      %s\n", line)
	}
}

func (e *Executor) execStep(a *automation.Automation, step automation.Step, args []string, index int, stdinOverride io.Reader, capturePipe bool, inputEnv []string) error {
	stdout := e.stdout()
	if capturePipe {
		buf := &bytes.Buffer{}
		stdout = buf
		defer func() { e.lastPipeBuffer = buf }()
	}

	stdin := e.stdin()
	if stdinOverride != nil {
		stdin = stdinOverride
	}

	switch step.Type {
	case automation.StepTypeBash:
		return e.execBash(a, step, args, stdout, stdin, inputEnv)
	case automation.StepTypePython:
		return e.execPython(a, step, args, stdout, stdin, inputEnv)
	case automation.StepTypeTypeScript:
		return e.execTypeScript(a, step, args, stdout, stdin, inputEnv)
	case automation.StepTypeRun:
		return e.execRun(step, args, stdout, stdin, inputEnv)
	default:
		return fmt.Errorf("step[%d]: step type %q is not implemented", index, step.Type)
	}
}

func (e *Executor) execBash(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	var cmdArgs []string

	if isFilePath(step.Value) {
		resolved := resolveScriptPath(a.Dir(), step.Value)
		if _, err := os.Stat(resolved); err != nil {
			return fmt.Errorf("bash script file not found: %s (resolved from %q relative to %s)", resolved, step.Value, a.Dir())
		}
		cmdArgs = append([]string{resolved}, args...)
	} else {
		cmdArgs = append([]string{"-c", step.Value, "--"}, args...)
	}

	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = stdout
	cmd.Stderr = e.stderr()
	cmd.Stdin = stdin
	cmd.Env = appendInputEnv(inputEnv)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("running bash step: %w", err)
	}
	return nil
}

func (e *Executor) execPython(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	pythonBin := e.resolvePythonBin()

	var cmdArgs []string
	if isFilePath(step.Value) {
		resolved := resolveScriptPath(a.Dir(), step.Value)
		if _, err := os.Stat(resolved); err != nil {
			return fmt.Errorf("python script file not found: %s (resolved from %q relative to %s)", resolved, step.Value, a.Dir())
		}
		cmdArgs = append([]string{resolved}, args...)
	} else {
		cmdArgs = append([]string{"-c", step.Value}, args...)
	}

	cmd := exec.Command(pythonBin, cmdArgs...)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = stdout
	cmd.Stderr = e.stderr()
	cmd.Stdin = stdin
	cmd.Env = appendInputEnv(inputEnv)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		if isCommandNotFound(err) {
			return fmt.Errorf("python3 not found in PATH — install Python 3 or activate a virtualenv")
		}
		return fmt.Errorf("running python step: %w", err)
	}
	return nil
}

// resolvePythonBin returns the python binary to use. If VIRTUAL_ENV is set,
// uses $VIRTUAL_ENV/bin/python; otherwise falls back to python3.
func (e *Executor) resolvePythonBin() string {
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		return filepath.Join(venv, "bin", "python")
	}
	return "python3"
}

// isCommandNotFound checks if an exec error indicates the binary wasn't found.
func isCommandNotFound(err error) bool {
	return strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file or directory")
}

func (e *Executor) execTypeScript(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	var cmdArgs []string
	var tmpFile string

	if isFilePath(step.Value) {
		resolved := resolveScriptPath(a.Dir(), step.Value)
		if _, err := os.Stat(resolved); err != nil {
			return fmt.Errorf("typescript file not found: %s (resolved from %q relative to %s)", resolved, step.Value, a.Dir())
		}
		cmdArgs = append([]string{resolved}, args...)
	} else {
		tmp, err := os.CreateTemp("", "pi-ts-*.ts")
		if err != nil {
			return fmt.Errorf("creating temp file for typescript step: %w", err)
		}
		tmpFile = tmp.Name()
		defer os.Remove(tmpFile)

		if _, err := tmp.WriteString(step.Value); err != nil {
			tmp.Close()
			return fmt.Errorf("writing typescript temp file: %w", err)
		}
		tmp.Close()

		cmdArgs = append([]string{tmpFile}, args...)
	}

	cmd := exec.Command("tsx", cmdArgs...)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = stdout
	cmd.Stderr = e.stderr()
	cmd.Stdin = stdin
	cmd.Env = appendInputEnv(inputEnv)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		if isCommandNotFound(err) {
			return fmt.Errorf("tsx not found in PATH — install it with: npm install -g tsx")
		}
		return fmt.Errorf("running typescript step: %w", err)
	}
	return nil
}

func (e *Executor) execRun(step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	target, err := e.Discovery.Find(step.Value)
	if err != nil {
		return fmt.Errorf("run step: %w", err)
	}

	origStdout, origStdin := e.Stdout, e.Stdin
	e.Stdout, e.Stdin = stdout, stdin
	var runErr error
	if len(step.With) > 0 {
		runErr = e.RunWithInputs(target, nil, step.With)
	} else {
		runErr = e.RunWithInputs(target, args, nil)
	}
	e.Stdout, e.Stdin = origStdout, origStdin
	return runErr
}

// pushCall adds a name to the call stack, returning an error on circular dependency.
func (e *Executor) pushCall(name string) error {
	for _, called := range e.callStack {
		if called == name {
			chain := append(e.callStack, name)
			return fmt.Errorf("circular automation dependency: %s", strings.Join(chain, " → "))
		}
	}
	e.callStack = append(e.callStack, name)
	return nil
}

func (e *Executor) popCall() {
	if len(e.callStack) > 0 {
		e.callStack = e.callStack[:len(e.callStack)-1]
	}
}

// isFilePath returns true if the value looks like a file path rather than inline script.
// A file path ends in a known script extension, contains no newlines, and contains no spaces.
func isFilePath(value string) bool {
	hasKnownExt := strings.HasSuffix(value, ".sh") ||
		strings.HasSuffix(value, ".py") ||
		strings.HasSuffix(value, ".ts")
	return hasKnownExt &&
		!strings.Contains(value, "\n") &&
		!strings.Contains(value, " ")
}

// resolveScriptPath resolves a script path relative to the automation's directory.
func resolveScriptPath(automationDir, scriptPath string) string {
	if filepath.IsAbs(scriptPath) {
		return scriptPath
	}
	return filepath.Join(automationDir, scriptPath)
}

// appendInputEnv merges PI_INPUT_* env vars into the current environment.
// If inputEnv is empty, returns nil (cmd.Env=nil inherits parent env).
func appendInputEnv(inputEnv []string) []string {
	if len(inputEnv) == 0 {
		return nil
	}
	env := os.Environ()
	return append(env, inputEnv...)
}

// evaluateCondition resolves and evaluates an if: expression.
// Returns true if the step should be skipped (condition is false).
func (e *Executor) evaluateCondition(expr string) (bool, error) {
	predNames, err := conditions.Predicates(expr)
	if err != nil {
		return false, err
	}

	env := e.RuntimeEnv
	if env == nil {
		env = DefaultRuntimeEnv()
	}

	resolved, err := ResolvePredicatesWithEnv(predNames, e.RepoRoot, env)
	if err != nil {
		return false, err
	}

	result, err := conditions.Eval(expr, resolved)
	if err != nil {
		return false, err
	}

	return !result, nil
}

func (e *Executor) stdout() io.Writer {
	if e.Stdout != nil {
		return e.Stdout
	}
	return os.Stdout
}

func (e *Executor) stderr() io.Writer {
	if e.Stderr != nil {
		return e.Stderr
	}
	return os.Stderr
}

func (e *Executor) stdin() io.Reader {
	if e.Stdin != nil {
		return e.Stdin
	}
	return os.Stdin
}
