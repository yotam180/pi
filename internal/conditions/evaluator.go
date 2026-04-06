package conditions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RuntimeEnv captures the runtime values needed for predicate resolution
// and requirement checking. All fields are populated from the real environment
// by DefaultRuntimeEnv(), but can be overridden in tests.
type RuntimeEnv struct {
	GOOS     string
	GOARCH   string
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)

	// ExecOutput is an optional function for testing that returns the output
	// of running a command with given args. Used by requirement validation
	// (in the executor package) to mock `<cmd> --version` calls. If nil,
	// the real command is executed. Not used by condition evaluation.
	ExecOutput func(cmd string, args ...string) string
}

// DefaultRuntimeEnv returns a RuntimeEnv backed by the real OS.
func DefaultRuntimeEnv() *RuntimeEnv {
	return &RuntimeEnv{
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,
		Getenv:   os.Getenv,
		LookPath: exec.LookPath,
		Stat:     os.Stat,
	}
}

// Evaluator provides a self-contained way to evaluate if: condition expressions.
// It resolves predicate names against the runtime environment and evaluates
// the boolean expression in one step.
type Evaluator struct {
	env      *RuntimeEnv
	repoRoot string
}

// NewEvaluator creates an Evaluator. If env is nil, DefaultRuntimeEnv() is used.
func NewEvaluator(repoRoot string, env *RuntimeEnv) *Evaluator {
	if env == nil {
		env = DefaultRuntimeEnv()
	}
	return &Evaluator{env: env, repoRoot: repoRoot}
}

// ShouldSkip evaluates a condition expression and returns true if the
// condition is false (meaning the step/automation should be skipped).
// An empty expression never skips.
func (ev *Evaluator) ShouldSkip(expr string) (bool, error) {
	predNames, err := Predicates(expr)
	if err != nil {
		return false, err
	}

	resolved, err := ResolvePredicatesWithEnv(predNames, ev.repoRoot, ev.env)
	if err != nil {
		return false, err
	}

	result, err := Eval(expr, resolved)
	if err != nil {
		return false, err
	}

	return !result, nil
}

// ResolvePredicates resolves a list of predicate names to boolean values
// using the default runtime environment.
func ResolvePredicates(predicateNames []string, repoRoot string) (map[string]bool, error) {
	return ResolvePredicatesWithEnv(predicateNames, repoRoot, DefaultRuntimeEnv())
}

// ResolvePredicatesWithEnv resolves predicate names to boolean values using
// the provided RuntimeEnv.
func ResolvePredicatesWithEnv(predicateNames []string, repoRoot string, env *RuntimeEnv) (map[string]bool, error) {
	result := make(map[string]bool, len(predicateNames))

	for _, name := range predicateNames {
		val, err := resolveSingle(name, repoRoot, env)
		if err != nil {
			return nil, err
		}
		result[name] = val
	}

	return result, nil
}

func resolveSingle(name string, repoRoot string, env *RuntimeEnv) (bool, error) {
	switch name {
	case "os.macos":
		return env.GOOS == "darwin", nil
	case "os.linux":
		return env.GOOS == "linux", nil
	case "os.windows":
		return env.GOOS == "windows", nil
	case "os.arch.arm64":
		return env.GOARCH == "arm64", nil
	case "os.arch.amd64":
		return env.GOARCH == "amd64", nil
	case "shell.zsh":
		return strings.HasSuffix(env.Getenv("SHELL"), "/zsh"), nil
	case "shell.bash":
		return strings.HasSuffix(env.Getenv("SHELL"), "/bash"), nil
	}

	if strings.HasPrefix(name, "env.") {
		envVar := name[len("env."):]
		if envVar == "" {
			return false, fmt.Errorf("invalid predicate %q: env variable name is empty", name)
		}
		return env.Getenv(envVar) != "", nil
	}

	if strings.HasPrefix(name, "command.") {
		cmdName := name[len("command."):]
		if cmdName == "" {
			return false, fmt.Errorf("invalid predicate %q: command name is empty", name)
		}
		_, err := env.LookPath(cmdName)
		return err == nil, nil
	}

	if strings.HasPrefix(name, "file.exists(\"") && strings.HasSuffix(name, "\")") {
		path := name[len("file.exists(\"") : len(name)-len("\")")]
		resolved := filepath.Join(repoRoot, path)
		info, err := env.Stat(resolved)
		if err != nil {
			return false, nil
		}
		return !info.IsDir(), nil
	}

	if strings.HasPrefix(name, "dir.exists(\"") && strings.HasSuffix(name, "\")") {
		path := name[len("dir.exists(\"") : len(name)-len("\")")]
		resolved := filepath.Join(repoRoot, path)
		info, err := env.Stat(resolved)
		if err != nil {
			return false, nil
		}
		return info.IsDir(), nil
	}

	return false, fmt.Errorf(
		"unknown predicate %q; valid prefixes: os.macos, os.linux, os.windows, "+
			"os.arch.arm64, os.arch.amd64, env.<NAME>, command.<name>, "+
			"file.exists(\"<path>\"), dir.exists(\"<path>\"), shell.zsh, shell.bash",
		name,
	)
}

// knownExactPredicates lists predicates that must match exactly.
var knownExactPredicates = map[string]bool{
	"os.macos":      true,
	"os.linux":      true,
	"os.windows":    true,
	"os.arch.arm64": true,
	"os.arch.amd64": true,
	"shell.zsh":     true,
	"shell.bash":    true,
}

// ValidatePredicateName checks whether a predicate name is statically valid
// without resolving its runtime value. This enables pi validate to catch
// typos like "os.macoss" at validation time instead of at execution time.
func ValidatePredicateName(name string) error {
	if knownExactPredicates[name] {
		return nil
	}

	if strings.HasPrefix(name, "env.") {
		if name[len("env."):] == "" {
			return fmt.Errorf("invalid predicate %q: env variable name is empty", name)
		}
		return nil
	}

	if strings.HasPrefix(name, "command.") {
		if name[len("command."):] == "" {
			return fmt.Errorf("invalid predicate %q: command name is empty", name)
		}
		return nil
	}

	if strings.HasPrefix(name, "file.exists(\"") && strings.HasSuffix(name, "\")") {
		return nil
	}

	if strings.HasPrefix(name, "dir.exists(\"") && strings.HasSuffix(name, "\")") {
		return nil
	}

	return fmt.Errorf(
		"unknown predicate %q; valid predicates: os.macos, os.linux, os.windows, "+
			"os.arch.arm64, os.arch.amd64, env.<NAME>, command.<name>, "+
			"file.exists(\"<path>\"), dir.exists(\"<path>\"), shell.zsh, shell.bash",
		name,
	)
}

// ValidateConditionExpr statically validates an if: expression by parsing it
// and checking that all predicate names are recognized. Returns nil if valid.
func ValidateConditionExpr(expr string) error {
	preds, err := Predicates(expr)
	if err != nil {
		return err
	}
	for _, pred := range preds {
		if err := ValidatePredicateName(pred); err != nil {
			return err
		}
	}
	return nil
}
