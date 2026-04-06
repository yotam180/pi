package executor

import (
	"bytes"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/discovery"
)

func newAutomation(name string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		Steps:    steps,
		FilePath: "/fake/path/automation.yaml",
	}
}

func newAutomationInDir(name, dir string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		Steps:    steps,
		FilePath: filepath.Join(dir, "automation.yaml"),
	}
}

func bashStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeBash, Value: value}
}

func runStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeRun, Value: value}
}

func pythonStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypePython, Value: value}
}

func typescriptStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeTypeScript, Value: value}
}

func pipedBashStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeBash, Value: value, Pipe: true}
}

func pipedPythonStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypePython, Value: value, Pipe: true}
}

func bashStepIf(value, cond string) automation.Step {
	return automation.Step{Type: automation.StepTypeBash, Value: value, If: cond}
}

func newDiscovery(automations map[string]*automation.Automation) *discovery.Result {
	return &discovery.Result{Automations: automations}
}

func newExecutor(repoRoot string, disc *discovery.Result) *Executor {
	return &Executor{
		RepoRoot:  repoRoot,
		Discovery: disc,
		Stdout:    io.Discard,
		Stderr:    io.Discard,
	}
}

func newExecutorWithCapture(repoRoot string, disc *discovery.Result) (*Executor, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return &Executor{
		RepoRoot:  repoRoot,
		Discovery: disc,
		Stdout:    stdout,
		Stderr:    stderr,
	}, stdout, stderr
}

func newExecutorWithEnv(repoRoot string, disc *discovery.Result, env *conditions.RuntimeEnv) (*Executor, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return &Executor{
		RepoRoot:   repoRoot,
		Discovery:  disc,
		Stdout:     stdout,
		Stderr:     stderr,
		RuntimeEnv: env,
	}, stdout, stderr
}

func fakeRuntimeEnv(goos string) *conditions.RuntimeEnv {
	return &conditions.RuntimeEnv{
		GOOS:     goos,
		GOARCH:   "arm64",
		Getenv:   func(s string) string { return "" },
		LookPath: func(s string) (string, error) { return "", osexec.ErrNotFound },
		Stat:     os.Stat,
	}
}

func newAutomationWithIf(name, cond string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		If:       cond,
		Steps:    steps,
		FilePath: "/fake/path/automation.yaml",
	}
}

func newInstallerAutomation(name string, inst *automation.InstallSpec) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		Install:  inst,
		FilePath: "/fake/path/automation.yaml",
	}
}

func automationWithInputs(name string, inputs map[string]automation.InputSpec, inputKeys []string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:      name,
		Inputs:    inputs,
		InputKeys: inputKeys,
		Steps:     steps,
		FilePath:  "/fake/path/automation.yaml",
	}
}

func runStepWith(value string, with map[string]string) automation.Step {
	return automation.Step{Type: automation.StepTypeRun, Value: value, With: with}
}

func boolPtr(b bool) *bool { return &b }

func requirePython(t *testing.T) {
	t.Helper()
	if _, err := osexec.LookPath("python3"); err != nil {
		t.Skip("python3 not found in PATH, skipping Python test")
	}
}

func requireTsx(t *testing.T) {
	t.Helper()
	if _, err := osexec.LookPath("tsx"); err != nil {
		t.Skip("tsx not found in PATH, skipping TypeScript test")
	}
}
