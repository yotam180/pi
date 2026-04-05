package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/runtimes"
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

	// Loud forces all steps to print their trace line and output,
	// overriding per-step silent: true flags.
	Loud bool

	// callStack tracks the chain of automation names currently being executed,
	// used to detect circular run: dependencies.
	callStack []string

	// lastPipeBuffer holds captured stdout from the last pipe_to:next step.
	lastPipeBuffer *bytes.Buffer

	// RuntimeEnv overrides the default runtime environment for predicate resolution.
	// If nil, DefaultRuntimeEnv() is used.
	RuntimeEnv *RuntimeEnv

	// Provisioner handles sandboxed runtime provisioning. If nil, no provisioning
	// is attempted — missing requirements produce errors as before.
	Provisioner *runtimes.Provisioner

	// runtimePaths accumulates provisioned bin directories to prepend to PATH
	// for all step executions within the current automation.
	runtimePaths []string

	// Printer handles styled output. If nil, a printer is created lazily from Stderr.
	Printer *display.Printer
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

	if err := e.ValidateRequirements(a); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			fmt.Fprint(e.stderr(), FormatValidationError(ve))
			return &ExitError{Code: 1}
		}
		return err
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

		suppress := step.Silent && !e.Loud
		if !suppress {
			e.printer().StepTrace(string(step.Type), step.Value)
		}

		isPipeSrc := step.PipeTo == "next" && i < len(a.Steps)-1
		if suppress {
			if err := e.execStepSuppressed(a, step, args, i, pipedInput, isPipeSrc, inputEnv); err != nil {
				return err
			}
		} else {
			if err := e.execStep(a, step, args, i, pipedInput, isPipeSrc, inputEnv); err != nil {
				return err
			}
		}
		pipedInput = nil
		if isPipeSrc {
			pipedInput = e.lastPipeBuffer
		}
	}
	return nil
}

// execStepSuppressed runs a step with stdout and stderr suppressed (for silent: true steps).
// Pipe capture still works so that downstream steps can receive data.
func (e *Executor) execStepSuppressed(a *automation.Automation, step automation.Step, args []string, index int, stdinOverride io.Reader, capturePipe bool, inputEnv []string) error {
	origStdout, origStderr := e.Stdout, e.Stderr
	if !capturePipe {
		e.Stdout = io.Discard
	}
	e.Stderr = io.Discard
	err := e.execStep(a, step, args, index, stdinOverride, capturePipe, inputEnv)
	e.Stdout, e.Stderr = origStdout, origStderr
	return err
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

// printer returns the executor's display Printer, lazily creating one from Stderr.
func (e *Executor) printer() *display.Printer {
	if e.Printer != nil {
		return e.Printer
	}
	return display.New(e.stderr())
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
