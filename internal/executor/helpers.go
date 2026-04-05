package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

// buildEnv constructs the environment for step execution, including
// PI_INPUT_* vars, provisioned runtime PATH prepends, and step-level env vars.
func (e *Executor) buildEnv(inputEnv []string, stepEnv map[string]string) []string {
	if len(inputEnv) == 0 && len(e.runtimePaths) == 0 && len(stepEnv) == 0 {
		return nil
	}
	env := os.Environ()
	if len(e.runtimePaths) > 0 {
		env = prependPathInEnv(env, e.runtimePaths)
	}
	if len(inputEnv) > 0 {
		env = append(env, inputEnv...)
	}
	for k, v := range stepEnv {
		env = append(env, k+"="+v)
	}
	return env
}

// runCommand executes a command with standard PI conventions: working directory,
// environment, and error wrapping. This is the common substrate for all step runners.
// Non-zero exit codes are returned as *ExitError. Other exec failures (including
// command-not-found) are returned as-is for callers to handle with context-specific messages.
func (e *Executor) runCommand(bin string, args []string, stdout io.Writer, stdin io.Reader, inputEnv []string, stepEnv map[string]string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = e.RepoRoot
	cmd.Stdout = stdout
	cmd.Stderr = e.stderr()
	cmd.Stdin = stdin
	cmd.Env = e.buildEnv(inputEnv, stepEnv)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}

// resolveFileStep checks if a step value is a file path, resolves it relative
// to the automation directory, and verifies the file exists. Returns the
// resolved path and true if it's a file, or ("", false) for inline scripts.
func resolveFileStep(automationDir, value, lang string) (string, bool, error) {
	if !isFilePath(value) {
		return "", false, nil
	}
	resolved := resolveScriptPath(automationDir, value)
	if _, err := os.Stat(resolved); err != nil {
		return "", false, fmt.Errorf("%s script file not found: %s (resolved from %q relative to %s)", lang, resolved, value, automationDir)
	}
	return resolved, true, nil
}

// prependPathInEnv finds the PATH entry in env and prepends the given dirs.
func prependPathInEnv(env []string, dirs []string) []string {
	prefix := strings.Join(dirs, string(os.PathListSeparator))
	for i, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			env[i] = "PATH=" + prefix + string(os.PathListSeparator) + entry[5:]
			return env
		}
	}
	env = append(env, "PATH="+prefix)
	return env
}
