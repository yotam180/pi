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

	var pipedInput io.Reader
	for i, step := range a.Steps {
		if step.If != "" {
			skip, err := e.evaluateCondition(step.If)
			if err != nil {
				return fmt.Errorf("automation %q step[%d] if: %w", a.Name, i, err)
			}
			if skip {
				// If this step is a pipe source, pass through the current piped input
				// (or discard it if there's nothing to pass through).
				if step.PipeTo == "next" && i < len(a.Steps)-1 {
					if pipedInput != nil {
						buf := &bytes.Buffer{}
						if _, err := io.Copy(buf, pipedInput); err != nil {
							return fmt.Errorf("passing piped input through skipped step[%d]: %w", i, err)
						}
						pipedInput = buf
					}
					// else pipedInput stays nil — nothing to pass through
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
