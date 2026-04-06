package executor

import (
	"context"
	"io"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// StepRunner executes a single step of a given type.
// Implementations handle language-specific concerns (binary resolution,
// inline vs file dispatch, temp files, etc.).
type StepRunner interface {
	Run(ctx *RunContext) error
}

// FileExtProvider is an optional interface that StepRunner implementations
// can satisfy to declare which file extension they own (e.g. ".sh", ".py").
// The Registry uses this to detect file references without hardcoding extensions.
type FileExtProvider interface {
	FileExt() string
}

// ParentShellCapable is an optional interface that StepRunner implementations
// can satisfy to declare support for parent_shell: true steps.
type ParentShellCapable interface {
	SupportsParentShell() bool
}

// RunContext bundles everything a step runner needs to execute a step.
// Runners should treat this as read-only except for stdout/stdin which
// represent the step's I/O targets.
type RunContext struct {
	Ctx        context.Context
	Automation *automation.Automation
	Step       automation.Step
	Args       []string
	Stdout     io.Writer
	Stderr     io.Writer
	Stdin      io.Reader
	InputEnv   []string

	RepoRoot     string
	RuntimePaths []string

	// WorkDir is the resolved working directory for this step.
	// Defaults to RepoRoot when the step has no dir: override.
	WorkDir string

	// Discovery is provided for run: steps that need to resolve other automations.
	Discovery *discovery.Result

	// RunAutomation is the callback for run: steps to recursively execute
	// another automation. The runner must not hold a reference to the Executor.
	RunAutomation func(ctx context.Context, target *automation.Automation, args []string, withArgs map[string]string, stdout io.Writer, stdin io.Reader) error

	// InterpolateWith resolves outputs.last, outputs.<N>, and inputs.<name>
	// references in with: values. Returns a new map with interpolated values.
	// The second argument is the current automation's input env vars for
	// inputs.<name> resolution.
	InterpolateWith func(with map[string]string, inputEnv []string) map[string]string

	// ResolvedAutomationEnv holds the automation-level env with interpolation
	// references (outputs.*/inputs.*) already resolved. Used by BuildStepEnv
	// instead of Automation.Env when set.
	ResolvedAutomationEnv map[string]string

	// ResolvedStepEnv holds the step-level env with interpolation references
	// already resolved. Used by BuildStepEnv instead of Step.Env when set.
	ResolvedStepEnv map[string]string
}

// Registry maps step types to their runner implementations.
type Registry struct {
	runners map[automation.StepType]StepRunner
}

// NewRegistry creates an empty runner registry.
func NewRegistry() *Registry {
	return &Registry{runners: make(map[automation.StepType]StepRunner)}
}

// Register adds a runner for a step type. Overwrites any existing registration.
func (r *Registry) Register(stepType automation.StepType, runner StepRunner) {
	r.runners[stepType] = runner
}

// Get returns the runner for a step type, or nil if not registered.
func (r *Registry) Get(stepType automation.StepType) StepRunner {
	return r.runners[stepType]
}

// FileExtForStepType returns the file extension for a step type's runner,
// or "" if the runner doesn't implement FileExtProvider.
func (r *Registry) FileExtForStepType(stepType automation.StepType) string {
	runner := r.runners[stepType]
	if p, ok := runner.(FileExtProvider); ok {
		return p.FileExt()
	}
	return ""
}

// StepTypeSupportsParentShell reports whether the runner registered for
// stepType implements ParentShellCapable and returns true.
func (r *Registry) StepTypeSupportsParentShell(stepType automation.StepType) bool {
	runner := r.runners[stepType]
	if p, ok := runner.(ParentShellCapable); ok {
		return p.SupportsParentShell()
	}
	return false
}

// NewDefaultRegistry creates a registry with all built-in step runners.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(automation.StepTypeBash, NewBashRunner())
	r.Register(automation.StepTypePython, NewPythonRunner())
	r.Register(automation.StepTypeTypeScript, NewTypeScriptRunner())
	r.Register(automation.StepTypeRun, &RunStepRunner{})
	return r
}

// DefaultFileExtensions returns the file extension for each step type in the
// default registry. Used by validation to detect file references without
// hardcoding extension lists.
func DefaultFileExtensions() map[automation.StepType]string {
	reg := NewDefaultRegistry()
	exts := make(map[automation.StepType]string)
	for stepType := range reg.runners {
		if ext := reg.FileExtForStepType(stepType); ext != "" {
			exts[stepType] = ext
		}
	}
	return exts
}
