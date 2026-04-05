package executor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
)

func (e *Executor) execBash(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	var cmdArgs []string

	resolved, isFile, err := resolveFileStep(a.Dir(), step.Value, "bash")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, args...)
	} else {
		cmdArgs = append([]string{"-c", step.Value, "--"}, args...)
	}

	if err := e.runCommand("bash", cmdArgs, stdout, stdin, inputEnv, step.Env); err != nil {
		if _, ok := err.(*ExitError); ok {
			return err
		}
		return fmt.Errorf("running bash step: %w", err)
	}
	return nil
}

func (e *Executor) execPython(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	pythonBin := e.resolvePythonBin()

	var cmdArgs []string
	resolved, isFile, err := resolveFileStep(a.Dir(), step.Value, "python")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, args...)
	} else {
		cmdArgs = append([]string{"-c", step.Value}, args...)
	}

	if err := e.runCommand(pythonBin, cmdArgs, stdout, stdin, inputEnv, step.Env); err != nil {
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
func (e *Executor) resolvePythonBin() string {
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		return filepath.Join(venv, "bin", "python")
	}
	return "python3"
}

func (e *Executor) execTypeScript(a *automation.Automation, step automation.Step, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string) error {
	var cmdArgs []string

	resolved, isFile, err := resolveFileStep(a.Dir(), step.Value, "typescript")
	if err != nil {
		return err
	}
	if isFile {
		cmdArgs = append([]string{resolved}, args...)
	} else {
		tmp, err := os.CreateTemp("", "pi-ts-*.ts")
		if err != nil {
			return fmt.Errorf("creating temp file for typescript step: %w", err)
		}
		defer os.Remove(tmp.Name())

		if _, err := tmp.WriteString(step.Value); err != nil {
			tmp.Close()
			return fmt.Errorf("writing typescript temp file: %w", err)
		}
		tmp.Close()

		cmdArgs = append([]string{tmp.Name()}, args...)
	}

	if err := e.runCommand("tsx", cmdArgs, stdout, stdin, inputEnv, step.Env); err != nil {
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

// isCommandNotFound checks if an exec error indicates the binary wasn't found.
func isCommandNotFound(err error) bool {
	return strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file or directory")
}
