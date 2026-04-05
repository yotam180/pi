package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IsFilePath returns true if the value looks like a file path rather than inline script.
// A file path ends in a known script extension, contains no newlines, and contains no spaces.
func IsFilePath(value string) bool {
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

// buildEnv constructs the environment for step execution, including
// PI_INPUT_* vars, provisioned runtime PATH prepends, automation-level
// env vars, and step-level env vars. Step-level env overrides automation-level
// env for the same key (both are appended; last writer wins in exec).
func (e *Executor) buildEnv(inputEnv []string, automationEnv map[string]string, stepEnv map[string]string) []string {
	if len(inputEnv) == 0 && len(e.runtimePaths) == 0 && len(automationEnv) == 0 && len(stepEnv) == 0 {
		return nil
	}
	env := os.Environ()
	if len(e.runtimePaths) > 0 {
		env = prependPathInEnv(env, e.runtimePaths)
	}
	if len(inputEnv) > 0 {
		env = append(env, inputEnv...)
	}
	if len(automationEnv) > 0 {
		envKeys := make([]string, 0, len(automationEnv))
		for k := range automationEnv {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)
		for _, k := range envKeys {
			env = append(env, k+"="+automationEnv[k])
		}
	}
	if len(stepEnv) > 0 {
		envKeys := make([]string, 0, len(stepEnv))
		for k := range stepEnv {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)
		for _, k := range envKeys {
			env = append(env, k+"="+stepEnv[k])
		}
	}
	return env
}

// resolveFileStep checks if a step value is a file path, resolves it relative
// to the automation directory, and verifies the file exists. Returns the
// resolved path and true if it's a file, or ("", false) for inline scripts.
func resolveFileStep(automationDir, value, lang string) (string, bool, error) {
	if !IsFilePath(value) {
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

// expandTraceVars expands environment variable references ($VAR and ${VAR})
// in a step value string for trace display purposes. It uses a combined map
// built from inputEnv, automation-level env, and step-level env.
// Only variables present in the provided sources are expanded; references
// to unknown variables pass through unchanged.
func expandTraceVars(value string, inputEnv []string, automationEnv map[string]string, stepEnv map[string]string) string {
	if !strings.ContainsAny(value, "$") {
		return value
	}

	vars := make(map[string]string)
	for _, entry := range inputEnv {
		if idx := strings.IndexByte(entry, '='); idx > 0 {
			vars[entry[:idx]] = entry[idx+1:]
		}
	}
	for k, v := range automationEnv {
		vars[k] = v
	}
	for k, v := range stepEnv {
		vars[k] = v
	}

	if len(vars) == 0 {
		return value
	}

	return os.Expand(value, func(key string) string {
		if v, ok := vars[key]; ok {
			return v
		}
		return "$" + key
	})
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
