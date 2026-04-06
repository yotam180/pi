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
	"github.com/vyper-tooling/pi/internal/interpolation"
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

	// DryRun prints what steps would be executed without actually running them.
	// Conditions are evaluated, run: targets are resolved and recursed into,
	// but no subprocess commands or side effects occur.
	DryRun bool

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

	// Runners is the step runner registry. If nil, NewDefaultRegistry() is used
	// and cached for subsequent calls.
	Runners *Registry

	// cachedRegistry holds the lazily-created default registry so that
	// registry() doesn't allocate a new one on every step execution.
	cachedRegistry *Registry

	// Outputs tracks step outputs for interpolation (outputs.last, outputs.<N>).
	Outputs interpolation.OutputTracker

	// condEval is the lazily-created condition evaluator for if: expressions.
	condEval *conditions.Evaluator
}

// stepExecCtx bundles the per-automation execution state that is constant
// across all steps within a single RunWithInputs call. Methods receive this
// instead of repeating the same parameter list.
type stepExecCtx struct {
	automation    *automation.Automation
	args          []string
	inputEnv      []string
	automationEnv map[string]string
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

	if len(args) > 0 {
		inputEnv = append(inputEnv, "PI_ARGS="+strings.Join(args, " "))
		for i, arg := range args {
			inputEnv = append(inputEnv, fmt.Sprintf("PI_ARG_%d=%s", i+1, arg))
		}
		inputEnv = append(inputEnv, fmt.Sprintf("PI_ARG_COUNT=%d", len(args)))
	}

	if err := e.ValidateRequirements(a); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Fprint(e.stderr(), FormatValidationError(ve))
			return &ExitError{Code: 1}
		}
		return err
	}

	if e.DryRun {
		return e.dryRunAutomation(a, args, inputEnv)
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

	ctx := &stepExecCtx{
		automation:    a,
		args:          args,
		inputEnv:      inputEnv,
		automationEnv: interpolation.ResolveEnv(a.Env, &e.Outputs, inputEnv),
	}

	saved := e.Outputs.Snapshot()
	e.Outputs.Reset()
	defer func() { e.Outputs.Restore(saved) }()

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
							return fmt.Errorf("automation %q: passing piped input through skipped step[%d]: %w", a.Name, i, err)
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
			if err := e.execFirstBlock(ctx, step, i, pipedInput, isPipeSrc); err != nil {
				return err
			}
			pipedInput = nil
			if isPipeSrc {
				pipedInput = e.lastPipeBuffer
			}
			continue
		}

		if step.ParentShell {
			if err := e.execParentShell(ctx, step); err != nil {
				return fmt.Errorf("automation %q step[%d]: %w", a.Name, i, err)
			}
			continue
		}

		suppress := step.Silent && !e.Loud
		if !suppress {
			e.printer().StepTrace(string(step.Type), expandTraceVars(step.Value, ctx.inputEnv, ctx.automationEnv, step.Env))
		}

		isPipeSrc := step.Pipe && i < len(a.Steps)-1
		stdoutW, stderrW := e.stdout(), e.stderr()
		if suppress {
			stdoutW, stderrW = io.Discard, io.Discard
		}
		if err := e.execStep(ctx, step, i, pipedInput, isPipeSrc, stdoutW, stderrW); err != nil {
			return err
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
func (e *Executor) execFirstBlock(ctx *stepExecCtx, step automation.Step, index int, stdinOverride io.Reader, capturePipe bool) error {
	a := ctx.automation
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

		if sub.ParentShell {
			if err := e.execParentShell(ctx, sub); err != nil {
				return fmt.Errorf("automation %q step[%d].first[%d]: %w", a.Name, index, j, err)
			}
			return nil
		}

		suppress := sub.Silent && !e.Loud
		if !suppress {
			e.printer().StepTrace(string(sub.Type), expandTraceVars(sub.Value, ctx.inputEnv, ctx.automationEnv, sub.Env))
		}

		stdoutW, stderrW := e.stdout(), e.stderr()
		if suppress {
			stdoutW, stderrW = io.Discard, io.Discard
		}
		return e.execStep(ctx, sub, index, stdinOverride, capturePipe, stdoutW, stderrW)
	}

	if capturePipe {
		e.lastPipeBuffer = &bytes.Buffer{}
	}
	return nil
}

// execStep runs a single step using the provided I/O destinations.
// displayStdout is where visible output goes (io.Discard for suppressed steps).
// stderrW is where stderr goes (io.Discard for suppressed steps).
// Output is always captured for outputs.last regardless of displayStdout.
func (e *Executor) execStep(ctx *stepExecCtx, step automation.Step, index int, stdinOverride io.Reader, capturePipe bool, displayStdout io.Writer, stderrW io.Writer) error {
	a := ctx.automation

	workDir := e.RepoRoot
	if step.Dir != "" {
		resolved, err := resolveStepDir(e.RepoRoot, step.Dir)
		if err != nil {
			return fmt.Errorf("automation %q step[%d]: %w", a.Name, index, err)
		}
		workDir = resolved
	}

	var stdout io.Writer
	var outputCapture bytes.Buffer
	if capturePipe {
		buf := &bytes.Buffer{}
		stdout = buf
		defer func() {
			e.lastPipeBuffer = buf
			e.Outputs.Record(buf.String())
		}()
	} else {
		stdout = io.MultiWriter(displayStdout, &outputCapture)
	}

	stdin := e.stdin()
	if stdinOverride != nil {
		stdin = stdinOverride
	}

	runner := e.registry().Get(step.Type)
	if runner == nil {
		return fmt.Errorf("automation %q step[%d]: step type %q is not implemented", a.Name, index, step.Type)
	}

	rc := e.newRunContext(a, step, ctx.args, stdout, stdin, ctx.inputEnv, workDir, stderrW)
	rc.ResolvedAutomationEnv = ctx.automationEnv
	rc.ResolvedStepEnv = interpolation.ResolveEnv(step.Env, &e.Outputs, ctx.inputEnv)
	err := runner.Run(rc)
	if !capturePipe {
		e.Outputs.Record(outputCapture.String())
	}
	return err
}

// registry returns the executor's runner registry, lazily creating and caching the default.
func (e *Executor) registry() *Registry {
	if e.Runners != nil {
		return e.Runners
	}
	if e.cachedRegistry == nil {
		e.cachedRegistry = NewDefaultRegistry()
	}
	return e.cachedRegistry
}

// newRunContext builds a RunContext from the executor's current state.
// workDir is the resolved working directory — pass e.RepoRoot when the step
// has no dir: override. stderrW is the stderr destination for this step.
// The caller is responsible for resolving and validating the directory
// before calling this method.
func (e *Executor) newRunContext(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string, workDir string, stderrW io.Writer) *RunContext {
	return &RunContext{
		Automation:   a,
		Step:         step,
		Args:         args,
		Stdout:       stdout,
		Stderr:       stderrW,
		Stdin:        stdin,
		InputEnv:     inputEnv,
		RepoRoot:     e.RepoRoot,
		RuntimePaths: e.runtimePaths,
		WorkDir:      workDir,
		Discovery:    e.Discovery,
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
		InterpolateWith: func(with map[string]string, inputEnv []string) map[string]string {
			return interpolation.ResolveWith(with, &e.Outputs, inputEnv)
		},
	}
}

// evaluateCondition resolves and evaluates an if: expression.
// Returns true if the step should be skipped (condition is false).
func (e *Executor) evaluateCondition(expr string) (bool, error) {
	return e.conditionEvaluator().ShouldSkip(expr)
}

// conditionEvaluator returns the executor's condition evaluator, lazily creating one.
func (e *Executor) conditionEvaluator() *conditions.Evaluator {
	if e.condEval != nil {
		return e.condEval
	}
	e.condEval = conditions.NewEvaluator(e.RepoRoot, e.RuntimeEnv)
	return e.condEval
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
func (e *Executor) execParentShell(ctx *stepExecCtx, step automation.Step) error {
	if !e.registry().StepTypeSupportsParentShell(step.Type) {
		return fmt.Errorf("step type %q does not support parent_shell", step.Type)
	}
	if e.ParentEvalFile == "" {
		e.printer().Warn("  ⚠  parent_shell step skipped: not running inside a PI shell wrapper. Run 'pi shell' to install shell integration.\n")
		return nil
	}
	e.printer().StepTrace("parent", expandTraceVars(step.Value, ctx.inputEnv, ctx.automationEnv, step.Env))
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

// interpolateWithCtx delegates to interpolation.ResolveWith for backward
// compatibility with tests that reference the method name.
func (e *Executor) interpolateWithCtx(with map[string]string, currentInputEnv []string) map[string]string {
	return interpolation.ResolveWith(with, &e.Outputs, currentInputEnv)
}

// interpolateWith resolves output references without input context.
func (e *Executor) interpolateWith(with map[string]string) map[string]string {
	return interpolation.ResolveWith(with, &e.Outputs, nil)
}

// interpolateEnv delegates to interpolation.ResolveEnv.
func (e *Executor) interpolateEnv(env map[string]string, inputEnv []string) map[string]string {
	return interpolation.ResolveEnv(env, &e.Outputs, inputEnv)
}

// interpolateValue delegates to interpolation.ResolveValue.
func (e *Executor) interpolateValue(v string, inputEnv []string) string {
	return interpolation.ResolveValue(v, &e.Outputs, inputEnv)
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
