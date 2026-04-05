package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
	"github.com/vyper-tooling/pi/internal/runtimes"
)

// ProjectContext holds the resolved project root and configuration,
// providing a common foundation for all CLI commands. It eliminates
// duplicated resolution boilerplate across commands.
type ProjectContext struct {
	Root   string
	Config *config.ProjectConfig
}

// resolveProject finds the project root from startDir and loads the config.
// Config load errors are ignored — the caller gets nil Config and can
// proceed (useful for commands like run/list/info/doctor that tolerate
// missing or invalid config).
func resolveProject(startDir string) (*ProjectContext, error) {
	root, err := project.FindRoot(startDir)
	if err != nil {
		return nil, err
	}
	cfg, _ := config.Load(root)
	return &ProjectContext{Root: root, Config: cfg}, nil
}

// resolveProjectStrict finds the project root and loads config, returning
// an error if config loading fails. Used by commands that require a valid
// pi.yaml (setup, shell install/uninstall).
func resolveProjectStrict(startDir string) (*ProjectContext, error) {
	root, err := project.FindRoot(startDir)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	return &ProjectContext{Root: root, Config: cfg}, nil
}

// Discover runs automation discovery (local + packages + builtins).
// stderr is used for package fetch status output; nil suppresses status.
func (pc *ProjectContext) Discover(stderr io.Writer) (*discovery.Result, error) {
	return discoverAllWithConfig(pc.Root, pc.Config, stderr)
}

// ExecutorOpts configures optional Executor behavior.
type ExecutorOpts struct {
	Stdout io.Writer
	Stderr io.Writer
	Silent bool
	Loud   bool
}

// NewExecutor builds an Executor from the project context and discovery result.
// ParentEvalFile is read from the environment. Provisioner is configured
// automatically when the project config enables it.
func (pc *ProjectContext) NewExecutor(result *discovery.Result, opts ExecutorOpts) *executor.Executor {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	exec := &executor.Executor{
		RepoRoot:       pc.Root,
		Discovery:      result,
		Stdout:         stdout,
		Stderr:         stderr,
		Silent:         opts.Silent,
		Loud:           opts.Loud,
		ParentEvalFile: os.Getenv("PI_PARENT_EVAL_FILE"),
	}

	if pc.Config != nil && pc.Config.EffectiveProvisionMode() != config.ProvisionNever {
		exec.Provisioner = runtimes.NewProvisioner(pc.Config, stderr)
	}

	return exec
}

// getwd is a helper that wraps os.Getwd with a consistent error message.
func getwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return cwd, nil
}
