package executor

import (
	"bytes"
	"errors"
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

	// lastPipeBuffer holds captured stdout from the last pipe: true step.
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

	// ParentEvalFile is the path to a file where parent-shell commands are written.
	// Steps with parent_shell: true append their commands to this file instead of
	// executing them. The calling shell wrapper evals the file after PI exits.
	// If empty, parent_shell steps produce an error.
	ParentEvalFile string

	// Runners is the step runner registry. If nil, NewDefaultRegistry() is used.
	Runners *Registry

	// stepOutputs stores trimmed stdout from each executed step (0-indexed).
	// Used to implement outputs.last interpolation in with: values.
	stepOutputs []string
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
// When a step declares pipe: true, its stdout is captured and fed as stdin
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
	var resolvedInputs map[string]string
	if len(a.Inputs) > 0 {
		var err error
		resolvedInputs, err = a.ResolveInputs(withArgs, args)
		if err != nil {
			return fmt.Errorf("automation %q: %w", a.Name, err)
		}
		inputEnv = automation.InputEnvVars(resolvedInputs)
		args = nil
	}

	if err := e.ValidateRequirements(a); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Fprint(e.stderr(), FormatValidationError(ve))
			return &ExitError{Code: 1}
		}
		return err
	}

	if a.IsGoFunc() {
		if resolvedInputs == nil {
			resolvedInputs = make(map[string]string)
		}
		if err := a.GoFunc(resolvedInputs); err != nil {
			fmt.Fprintf(e.stderr(), "%s\n", err.Error())
			return &ExitError{Code: 1}
		}
		return nil
	}

	if a.IsInstaller() {
		return e.execInstall(a, inputEnv)
	}

	savedOutputs := e.stepOutputs
	e.stepOutputs = nil
	defer func() { e.stepOutputs = savedOutputs }()

	var pipedInput io.Reader
	for i, step := range a.Steps {
		if step.If != "" {
			skip, err := e.evaluateCondition(step.If)
			if err != nil {
				return fmt.Errorf("automation %q step[%d] if: %w", a.Name, i, err)
			}
			if skip {
				if step.Pipe && i < len(a.Steps)-1 {
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

		if step.IsFirst() {
			isPipeSrc := step.Pipe && i < len(a.Steps)-1
			if err := e.execFirstBlock(a, step, args, i, pipedInput, isPipeSrc, inputEnv); err != nil {
				return err
			}
			pipedInput = nil
			if isPipeSrc {
				pipedInput = e.lastPipeBuffer
			}
			continue
		}

		if step.ParentShell {
			if err := e.execParentShell(a, step, inputEnv); err != nil {
				return fmt.Errorf("automation %q step[%d]: %w", a.Name, i, err)
			}
			continue
		}

		suppress := step.Silent && !e.Loud
		if !suppress {
			e.printer().StepTrace(string(step.Type), expandTraceVars(step.Value, inputEnv, a.Env, step.Env))
		}

		isPipeSrc := step.Pipe && i < len(a.Steps)-1
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

// execFirstBlock runs a first: block — evaluates sub-step conditions in order and
// executes only the first sub-step whose if: condition passes (or has no if:).
// If no sub-step matches, the block is silently skipped.
func (e *Executor) execFirstBlock(a *automation.Automation, step automation.Step, args []string, index int, stdinOverride io.Reader, capturePipe bool, inputEnv []string) error {
	for j, sub := range step.First {
		if sub.If != "" {
			skip, err := e.evaluateCondition(sub.If)
			if err != nil {
				return fmt.Errorf("automation %q step[%d].first[%d] if: %w", a.Name, index, j, err)
			}
			if skip {
				continue
			}
		}

		// Found our match — execute this sub-step
		if sub.ParentShell {
			if err := e.execParentShell(a, sub, inputEnv); err != nil {
				return fmt.Errorf("automation %q step[%d].first[%d]: %w", a.Name, index, j, err)
			}
			return nil
		}

		suppress := sub.Silent && !e.Loud
		if !suppress {
			e.printer().StepTrace(string(sub.Type), expandTraceVars(sub.Value, inputEnv, a.Env, sub.Env))
		}

		if suppress {
			return e.execStepSuppressed(a, sub, args, index, stdinOverride, capturePipe, inputEnv)
		}
		return e.execStep(a, sub, args, index, stdinOverride, capturePipe, inputEnv)
	}

	// No sub-step matched — silently skip the entire first: block.
	// If this block is a pipe source, set an empty buffer so downstream steps
	// receive empty stdin rather than stale data.
	if capturePipe {
		e.lastPipeBuffer = &bytes.Buffer{}
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
	if step.Dir != "" {
		if _, err := resolveStepDir(e.RepoRoot, step.Dir); err != nil {
			return fmt.Errorf("step[%d]: %w", index, err)
		}
	}

	stdout := e.stdout()
	var outputCapture bytes.Buffer
	if capturePipe {
		buf := &bytes.Buffer{}
		stdout = buf
		defer func() {
			e.lastPipeBuffer = buf
			e.recordOutput(buf.String())
		}()
	} else {
		stdout = io.MultiWriter(stdout, &outputCapture)
	}

	stdin := e.stdin()
	if stdinOverride != nil {
		stdin = stdinOverride
	}

	runner := e.registry().Get(step.Type)
	if runner == nil {
		return fmt.Errorf("step[%d]: step type %q is not implemented", index, step.Type)
	}

	err := runner.Run(e.newRunContext(a, step, args, stdout, stdin, inputEnv))
	if !capturePipe {
		e.recordOutput(outputCapture.String())
	}
	return err
}

// registry returns the executor's runner registry, lazily creating the default.
func (e *Executor) registry() *Registry {
	if e.Runners != nil {
		return e.Runners
	}
	return NewDefaultRegistry()
}

// newRunContext builds a RunContext from the executor's current state.
// If the step declares dir:, it must be resolved before calling this method
// and the resolved path passed via workDir. If workDir is empty, RepoRoot is used.
func (e *Executor) newRunContext(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) *RunContext {
	workDir := e.RepoRoot
	if step.Dir != "" {
		if resolved, err := resolveStepDir(e.RepoRoot, step.Dir); err == nil {
			workDir = resolved
		}
	}
	return &RunContext{
		Automation:   a,
		Step:         step,
		Args:         args,
		Stdout:       stdout,
		Stderr:       e.stderr(),
		Stdin:        stdin,
		InputEnv:     inputEnv,
		RepoRoot:     e.RepoRoot,
		RuntimePaths: e.runtimePaths,
		WorkDir:      workDir,
		Discovery:    e.Discovery,
		BuildEnv:     e.buildEnv,
		RunAutomation: func(target *automation.Automation, args []string, withArgs map[string]string, targetStdout io.Writer, targetStdin io.Reader) error {
			origStdout, origStdin := e.Stdout, e.Stdin
			e.Stdout, e.Stdin = targetStdout, targetStdin
			var runErr error
			if len(withArgs) > 0 {
				runErr = e.RunWithInputs(target, nil, withArgs)
			} else {
				runErr = e.RunWithInputs(target, args, nil)
			}
			e.Stdout, e.Stdin = origStdout, origStdin
			return runErr
		},
		InterpolateWith: e.interpolateWithCtx,
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

// execParentShell writes a parent_shell step's command to the eval file
// instead of executing it. The calling shell wrapper sources the file after PI exits.
func (e *Executor) execParentShell(a *automation.Automation, step automation.Step, inputEnv []string) error {
	if e.ParentEvalFile == "" {
		e.printer().Warn("  ⚠  parent_shell step skipped: not running inside a PI shell wrapper. Run 'pi shell' to install shell integration.\n")
		return nil
	}
	e.printer().StepTrace("parent", expandTraceVars(step.Value, inputEnv, a.Env, step.Env))
	return AppendToParentEval(e.ParentEvalFile, step.Value)
}

// AppendToParentEval appends a command to the parent eval file.
// The file is created if it doesn't exist.
func AppendToParentEval(path, command string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening parent eval file: %w", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, command); err != nil {
		return fmt.Errorf("writing to parent eval file: %w", err)
	}
	return nil
}

// lastOutput returns the trimmed stdout of the most recently executed step,
// or "" if no step has produced output yet.
func (e *Executor) lastOutput() string {
	if len(e.stepOutputs) == 0 {
		return ""
	}
	return e.stepOutputs[len(e.stepOutputs)-1]
}

// recordOutput appends a step's captured stdout to the outputs list.
func (e *Executor) recordOutput(output string) {
	e.stepOutputs = append(e.stepOutputs, strings.TrimSpace(output))
}

// interpolateWithCtx resolves outputs.last, outputs.<N>, and inputs.<name>
// references in with: values, given the current automation's resolved inputs.
func (e *Executor) interpolateWithCtx(with map[string]string, currentInputEnv []string) map[string]string {
	if len(with) == 0 {
		return with
	}
	result := make(map[string]string, len(with))
	for k, v := range with {
		result[k] = e.interpolateValue(v, currentInputEnv)
	}
	return result
}

// interpolateWith resolves output references without input context.
func (e *Executor) interpolateWith(with map[string]string) map[string]string {
	return e.interpolateWithCtx(with, nil)
}

// interpolateValue replaces "outputs.last", "outputs.<N>", and "inputs.<name>"
// references in a string value.
func (e *Executor) interpolateValue(v string, inputEnv []string) string {
	if v == "outputs.last" {
		return e.lastOutput()
	}
	if strings.HasPrefix(v, "outputs.") {
		suffix := strings.TrimPrefix(v, "outputs.")
		var n int
		if _, err := fmt.Sscanf(suffix, "%d", &n); err == nil {
			if n >= 0 && n < len(e.stepOutputs) {
				return e.stepOutputs[n]
			}
		}
	}
	if strings.HasPrefix(v, "inputs.") {
		name := strings.TrimPrefix(v, "inputs.")
		envKey := "PI_IN_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		for _, entry := range inputEnv {
			if strings.HasPrefix(entry, envKey+"=") {
				return entry[len(envKey)+1:]
			}
		}
	}
	return v
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
