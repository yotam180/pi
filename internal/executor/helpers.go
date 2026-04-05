package executor

import (
	"fmt"
	"os"
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

// resolveStepDir resolves the working directory for a step. If stepDir is empty,
// returns repoRoot. If stepDir is absolute, returns it as-is. Otherwise, joins
// it with repoRoot. Returns an error if the resolved directory doesn't exist.
func resolveStepDir(repoRoot, stepDir string) (string, error) {
	if stepDir == "" {
		return repoRoot, nil
	}
	dir := stepDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(repoRoot, dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("step dir %q does not exist (resolved to %s)", stepDir, dir)
		}
		return "", fmt.Errorf("checking step dir %q: %w", stepDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("step dir %q is not a directory (resolved to %s)", stepDir, dir)
	}
	return dir, nil
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
