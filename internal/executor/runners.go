package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BashRunner executes bash steps (inline scripts or .sh files).
type BashRunner struct{}

func (r *BashRunner) Run(ctx *RunContext) error {
	var cmdArgs []string

	resolved, isFile, err := resolveFileStep(ctx.Automation.Dir(), ctx.Step.Value, "bash")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, ctx.Args...)
	} else {
		cmdArgs = append([]string{"-c", ctx.Step.Value, "--"}, ctx.Args...)
	}

	if err := runStepCommand("bash", cmdArgs, ctx); err != nil {
		if _, ok := err.(*ExitError); ok {
			return err
		}
		return fmt.Errorf("running bash step: %w", err)
	}
	return nil
}

// PythonRunner executes python steps (inline scripts or .py files).
type PythonRunner struct{}

func (r *PythonRunner) Run(ctx *RunContext) error {
	pythonBin := resolvePythonBin()

	var cmdArgs []string
	resolved, isFile, err := resolveFileStep(ctx.Automation.Dir(), ctx.Step.Value, "python")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, ctx.Args...)
	} else {
		cmdArgs = append([]string{"-c", ctx.Step.Value}, ctx.Args...)
	}

	if err := runStepCommand(pythonBin, cmdArgs, ctx); err != nil {
		if _, ok := err.(*ExitError); ok {
			return err
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
func resolvePythonBin() string {
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		return filepath.Join(venv, "bin", "python")
	}
	return "python3"
}

// TypeScriptRunner executes typescript steps (inline scripts or .ts files).
type TypeScriptRunner struct{}

func (r *TypeScriptRunner) Run(ctx *RunContext) error {
	var cmdArgs []string

	resolved, isFile, err := resolveFileStep(ctx.Automation.Dir(), ctx.Step.Value, "typescript")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, ctx.Args...)
	} else {
		tmp, err := os.CreateTemp("", "pi-ts-*.ts")
		if err != nil {
			return fmt.Errorf("creating temp file for typescript step: %w", err)
		}
		defer os.Remove(tmp.Name())

		if _, err := tmp.WriteString(ctx.Step.Value); err != nil {
			tmp.Close()
			return fmt.Errorf("writing typescript temp file: %w", err)
		}
		tmp.Close()

		cmdArgs = append([]string{tmp.Name()}, ctx.Args...)
	}

	if err := runStepCommand("tsx", cmdArgs, ctx); err != nil {
		if _, ok := err.(*ExitError); ok {
			return err
		}
		if isCommandNotFound(err) {
			return fmt.Errorf("tsx not found in PATH — install it with: npm install -g tsx")
		}
		return fmt.Errorf("running typescript step: %w", err)
	}
	return nil
}

// RunStepRunner executes run: steps by recursively invoking another automation.
type RunStepRunner struct{}

func (r *RunStepRunner) Run(ctx *RunContext) error {
	target, err := ctx.Discovery.Find(ctx.Step.Value)
	if err != nil {
		return fmt.Errorf("run step: %w", err)
	}

	if len(ctx.Step.With) > 0 {
		return ctx.RunAutomation(target, nil, ctx.Step.With, ctx.Stdout, ctx.Stdin)
	}
	return ctx.RunAutomation(target, ctx.Args, nil, ctx.Stdout, ctx.Stdin)
}

// runStepCommand executes a command using the RunContext for directory, env, and I/O.
// Non-zero exit codes are returned as *ExitError. Other exec failures are returned as-is.
func runStepCommand(bin string, args []string, ctx *RunContext) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = ctx.RepoRoot
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	cmd.Env = ctx.BuildEnv(ctx.InputEnv, ctx.Step.Env)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
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
