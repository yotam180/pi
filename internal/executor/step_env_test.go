package executor

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestStepEnv_BashInline(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$MY_VAR" > ` + outFile,
		Env:   map[string]string{"MY_VAR": "hello_from_env"},
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "hello_from_env" {
		t.Errorf("got %q, want %q", strings.TrimSpace(string(got)), "hello_from_env")
	}
}

func TestStepEnv_MultipleVars(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$GOOS-$GOARCH" > ` + outFile,
		Env:   map[string]string{"GOOS": "linux", "GOARCH": "arm64"},
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "linux-arm64" {
		t.Errorf("got %q, want %q", strings.TrimSpace(string(got)), "linux-arm64")
	}
}

func TestStepEnv_OverridesParent(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	os.Setenv("PI_TEST_STEP_ENV_VAR", "original")
	defer os.Unsetenv("PI_TEST_STEP_ENV_VAR")

	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$PI_TEST_STEP_ENV_VAR" > ` + outFile,
		Env:   map[string]string{"PI_TEST_STEP_ENV_VAR": "overridden"},
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "overridden" {
		t.Errorf("got %q, want %q", strings.TrimSpace(string(got)), "overridden")
	}
}

func TestStepEnv_NilEnvInheritsParent(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	os.Setenv("PI_TEST_PARENT_VAR", "from_parent")
	defer os.Unsetenv("PI_TEST_PARENT_VAR")

	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$PI_TEST_PARENT_VAR" > ` + outFile,
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "from_parent" {
		t.Errorf("got %q, want %q", strings.TrimSpace(string(got)), "from_parent")
	}
}

func TestStepEnv_PythonStep(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	var stdout bytes.Buffer
	step := automation.Step{
		Type:  automation.StepTypePython,
		Value: `import os; print(os.environ.get("MY_PY_VAR", ""))`,
		Env:   map[string]string{"MY_PY_VAR": "python_env_works"},
	}
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(stdout.String()) != "python_env_works" {
		t.Errorf("got %q, want %q", strings.TrimSpace(stdout.String()), "python_env_works")
	}
}

func TestStepEnv_PerStepIsolation(t *testing.T) {
	dir := t.TempDir()
	outFile1 := filepath.Join(dir, "out1.txt")
	outFile2 := filepath.Join(dir, "out2.txt")

	step1 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$STEP_SPECIFIC" > ` + outFile1,
		Env:   map[string]string{"STEP_SPECIFIC": "step1_val"},
	}
	step2 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "${STEP_SPECIFIC:-empty}" > ` + outFile2,
	}
	a := newAutomation("test", step1, step2)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got1, _ := os.ReadFile(outFile1)
	if strings.TrimSpace(string(got1)) != "step1_val" {
		t.Errorf("step1 got %q, want %q", strings.TrimSpace(string(got1)), "step1_val")
	}

	got2, _ := os.ReadFile(outFile2)
	if strings.TrimSpace(string(got2)) != "empty" {
		t.Errorf("step2 got %q, want %q (env from step1 should not leak)", strings.TrimSpace(string(got2)), "empty")
	}
}

func TestBuildEnv_WithStepEnv(t *testing.T) {
	exec := &Executor{
		RepoRoot: t.TempDir(),
	}

	env := exec.buildEnv(nil, map[string]string{"FOO": "bar"})
	if env == nil {
		t.Fatal("expected non-nil env when step env is set")
	}

	found := false
	for _, e := range env {
		if e == "FOO=bar" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected FOO=bar in env")
	}
}

func TestBuildEnv_WithAllThree(t *testing.T) {
	exec := &Executor{
		RepoRoot:     t.TempDir(),
		runtimePaths: []string{"/provisioned/bin"},
	}

	env := exec.buildEnv(
		[]string{"PI_INPUT_X=1"},
		map[string]string{"STEP_VAR": "sv"},
	)
	if env == nil {
		t.Fatal("expected non-nil env")
	}

	hasInput, hasStep, hasPath := false, false, false
	for _, e := range env {
		if e == "PI_INPUT_X=1" {
			hasInput = true
		}
		if e == "STEP_VAR=sv" {
			hasStep = true
		}
		if strings.HasPrefix(e, "PATH=") && strings.Contains(e, "/provisioned/bin") {
			hasPath = true
		}
	}
	if !hasInput {
		t.Error("missing PI_INPUT_X=1")
	}
	if !hasStep {
		t.Error("missing STEP_VAR=sv")
	}
	if !hasPath {
		t.Error("missing provisioned PATH")
	}
}
