package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SubprocessConfig defines the language-specific behavior for a subprocess-based
// step runner. All subprocess runners (bash, python, typescript, etc.) share the
// same lifecycle: resolve file path → build command args → execute → handle errors.
// The config captures only the parts that differ between languages.
type SubprocessConfig struct {
	// Binary is the fixed command name (e.g. "bash", "tsx"). If BinaryFunc is set,
	// it takes precedence. One of Binary or BinaryFunc must be provided.
	Binary string

	// BinaryFunc dynamically resolves the command binary. Used when the binary
	// depends on runtime state (e.g. Python venv detection). Takes precedence
	// over Binary when set.
	BinaryFunc func() string

	// InlineArgs builds the argument list for an inline script.
	// Receives the script text and returns the args to pass after the binary name.
	// For example, bash returns ["-c", script, "--"] while python returns ["-c", script].
	// If nil, TempFileExt must be set (temp-file mode).
	InlineArgs func(script string) []string

	// TempFileExt, when non-empty, writes inline scripts to a temp file with
	// this extension instead of using InlineArgs. Used by languages whose
	// interpreters don't support -c (e.g. tsx). The temp file is cleaned up
	// after execution.
	TempFileExt string

	// NotFoundMsg is returned when the binary is not found in PATH.
	// If empty, a generic error is used.
	NotFoundMsg string

	// Language is the human-readable name for error messages (e.g. "bash", "python").
	Language string
}

// SubprocessRunner executes steps by shelling out to an external command.
// It handles file-path resolution, inline-to-temp-file conversion, command
// execution, and error wrapping — the common lifecycle shared by all
// language-specific step runners.
type SubprocessRunner struct {
	Config SubprocessConfig
}

func (r *SubprocessRunner) Run(ctx *RunContext) error {
	bin := r.Config.Binary
	if r.Config.BinaryFunc != nil {
		bin = r.Config.BinaryFunc()
	}

	var cmdArgs []string

	resolved, isFile, err := resolveFileStep(ctx.Automation.Dir(), ctx.Step.Value, r.Config.Language)
	if err != nil {
		return err
	}

	if isFile {
		cmdArgs = append([]string{resolved}, ctx.Args...)
	} else if r.Config.TempFileExt != "" {
		tmpPath, cleanup, err := writeTempScript(ctx.Step.Value, r.Config.TempFileExt)
		if err != nil {
			return err
		}
		defer cleanup()
		cmdArgs = append([]string{tmpPath}, ctx.Args...)
	} else if r.Config.InlineArgs != nil {
		cmdArgs = append(r.Config.InlineArgs(ctx.Step.Value), ctx.Args...)
	} else {
		return fmt.Errorf("step type %q: no inline execution method configured", r.Config.Language)
	}

	if err := runStepCommand(bin, cmdArgs, ctx); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			return err
		}
		if isCommandNotFound(err) && r.Config.NotFoundMsg != "" {
			return fmt.Errorf("%s", r.Config.NotFoundMsg)
		}
		return fmt.Errorf("running %s step: %w", r.Config.Language, err)
	}
	return nil
}

// writeTempScript writes script content to a temporary file with the given
// extension and returns the path, a cleanup function, and any error.
func writeTempScript(content, ext string) (string, func(), error) {
	tmp, err := os.CreateTemp("", "pi-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("creating temp file for step: %w", err)
	}
	cleanup := func() { os.Remove(tmp.Name()) }

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		cleanup()
		return "", nil, fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()
	return tmp.Name(), cleanup, nil
}

// resolvePythonBin returns the python binary to use. If VIRTUAL_ENV is set,
// uses $VIRTUAL_ENV/bin/python; otherwise falls back to python3.
func resolvePythonBin() string {
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		return filepath.Join(venv, "bin", "python")
	}
	return "python3"
}

// NewBashRunner creates a SubprocessRunner configured for bash steps.
func NewBashRunner() *SubprocessRunner {
	return &SubprocessRunner{Config: SubprocessConfig{
		Binary:   "bash",
		Language: "bash",
		InlineArgs: func(script string) []string {
			return []string{"-c", script, "--"}
		},
	}}
}

// NewPythonRunner creates a SubprocessRunner configured for python steps.
func NewPythonRunner() *SubprocessRunner {
	return &SubprocessRunner{Config: SubprocessConfig{
		BinaryFunc:  resolvePythonBin,
		Language:    "python",
		NotFoundMsg: "python3 not found in PATH — install Python 3 or activate a virtualenv",
		InlineArgs: func(script string) []string {
			return []string{"-c", script}
		},
	}}
}

// NewTypeScriptRunner creates a SubprocessRunner configured for typescript steps.
func NewTypeScriptRunner() *SubprocessRunner {
	return &SubprocessRunner{Config: SubprocessConfig{
		Binary:      "tsx",
		Language:    "typescript",
		NotFoundMsg: "tsx not found in PATH — install it with: npm install -g tsx",
		TempFileExt: ".ts",
	}}
}

// RunStepRunner executes run: steps by recursively invoking another automation.
type RunStepRunner struct{}

func (r *RunStepRunner) Run(ctx *RunContext) error {
	target, err := ctx.Discovery.Find(ctx.Step.Value)
	if err != nil {
		return fmt.Errorf("run step: %w", err)
	}

	if len(ctx.Step.With) > 0 {
		with := ctx.Step.With
		if ctx.InterpolateWith != nil {
			with = ctx.InterpolateWith(with, ctx.InputEnv)
		}
		return ctx.RunAutomation(target, nil, with, ctx.Stdout, ctx.Stdin)
	}
	return ctx.RunAutomation(target, ctx.Args, nil, ctx.Stdout, ctx.Stdin)
}

// TimeoutExitCode is the exit code returned when a step exceeds its timeout.
// Matches the GNU timeout(1) convention.
const TimeoutExitCode = 124

// runStepCommand executes a command using the RunContext for directory, env, and I/O.
// Non-zero exit codes are returned as *ExitError. Other exec failures are returned as-is.
// When the step has a timeout, exec.CommandContext is used with a deadline.
func runStepCommand(bin string, args []string, ctx *RunContext) error {
	var cmd *exec.Cmd

	if ctx.Step.Timeout > 0 {
		deadline, cancel := context.WithTimeout(context.Background(), ctx.Step.Timeout)
		defer cancel()
		cmd = exec.CommandContext(deadline, bin, args...)
	} else {
		cmd = exec.Command(bin, args...)
	}

	cmd.Dir = ctx.WorkDir
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	cmd.Env = ctx.BuildEnv(ctx.InputEnv, ctx.Automation.Env, ctx.Step.Env)

	if err := cmd.Run(); err != nil {
		if ctx.Step.Timeout > 0 && cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			return &ExitError{Code: TimeoutExitCode}
		}
		if ctx.Step.Timeout > 0 && err.Error() == "signal: killed" {
			return &ExitError{Code: TimeoutExitCode}
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}

// isCommandNotFound checks if an exec error indicates the binary wasn't found.
func isCommandNotFound(err error) bool {
	return strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file or directory")
}
