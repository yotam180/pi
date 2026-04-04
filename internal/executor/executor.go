package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
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
	Stdout *os.File
	Stderr *os.File

	// callStack tracks the chain of automation names currently being executed,
	// used to detect circular run: dependencies.
	callStack []string
}

// ExitError wraps a non-zero exit code from a step.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("step exited with code %d", e.Code)
}

// Run executes all steps of the given automation in order.
// args are passed to bash steps as $1, $2, etc.
func (e *Executor) Run(a *automation.Automation, args []string) error {
	if err := e.pushCall(a.Name); err != nil {
		return err
	}
	defer e.popCall()

	for i, step := range a.Steps {
		if err := e.execStep(a, step, args, i); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) execStep(a *automation.Automation, step automation.Step, args []string, index int) error {
	switch step.Type {
	case automation.StepTypeBash:
		return e.execBash(a, step, args)
	case automation.StepTypeRun:
		return e.execRun(step, args)
	default:
		return fmt.Errorf("step[%d]: step type %q is not implemented", index, step.Type)
	}
}

func (e *Executor) execBash(a *automation.Automation, step automation.Step, args []string) error {
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
	cmd.Stdout = e.stdout()
	cmd.Stderr = e.stderr()
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("running bash step: %w", err)
	}
	return nil
}

func (e *Executor) execRun(step automation.Step, args []string) error {
	target, err := e.Discovery.Find(step.Value)
	if err != nil {
		return fmt.Errorf("run step: %w", err)
	}
	return e.Run(target, args)
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
// A file path ends in .sh, contains no newlines, and contains no spaces.
func isFilePath(value string) bool {
	return strings.HasSuffix(value, ".sh") &&
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

func (e *Executor) stdout() *os.File {
	if e.Stdout != nil {
		return e.Stdout
	}
	return os.Stdout
}

func (e *Executor) stderr() *os.File {
	if e.Stderr != nil {
		return e.Stderr
	}
	return os.Stderr
}
