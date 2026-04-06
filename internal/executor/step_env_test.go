package executor

import (
	"bytes"
	"fmt"
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

func TestBuildStepEnv_WithStepEnv(t *testing.T) {
	env := BuildStepEnv(nil, nil, nil, map[string]string{"FOO": "bar"})
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

func TestBuildStepEnv_StepEnvDeterministicOrder(t *testing.T) {
	stepEnv := map[string]string{
		"ZEBRA":  "z",
		"ALPHA":  "a",
		"MIDDLE": "m",
	}

	for i := 0; i < 20; i++ {
		env := BuildStepEnv(nil, nil, nil, stepEnv)
		if env == nil {
			t.Fatal("expected non-nil env")
		}
		var stepVars []string
		for _, e := range env {
			if e == "ALPHA=a" || e == "MIDDLE=m" || e == "ZEBRA=z" {
				stepVars = append(stepVars, e)
			}
		}
		if len(stepVars) != 3 {
			t.Fatalf("iteration %d: expected 3 step vars, got %d", i, len(stepVars))
		}
		if stepVars[0] != "ALPHA=a" || stepVars[1] != "MIDDLE=m" || stepVars[2] != "ZEBRA=z" {
			t.Errorf("iteration %d: step env vars not in sorted order: %v", i, stepVars)
		}
	}
}

func TestBuildStepEnv_WithAllThree(t *testing.T) {
	env := BuildStepEnv(
		[]string{"/provisioned/bin"},
		[]string{"PI_INPUT_X=1"},
		map[string]string{"AUTO_VAR": "av"},
		map[string]string{"STEP_VAR": "sv"},
	)
	if env == nil {
		t.Fatal("expected non-nil env")
	}

	hasInput, hasAuto, hasStep, hasPath := false, false, false, false
	for _, e := range env {
		if e == "PI_INPUT_X=1" {
			hasInput = true
		}
		if e == "AUTO_VAR=av" {
			hasAuto = true
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
	if !hasAuto {
		t.Error("missing AUTO_VAR=av")
	}
	if !hasStep {
		t.Error("missing STEP_VAR=sv")
	}
	if !hasPath {
		t.Error("missing provisioned PATH")
	}
}

func TestAutomationEnv_AppliesToAllSteps(t *testing.T) {
	dir := t.TempDir()
	outFile1 := filepath.Join(dir, "out1.txt")
	outFile2 := filepath.Join(dir, "out2.txt")

	a := &automation.Automation{
		Name: "auto-env",
		Env: map[string]string{
			"MY_VAR": "from_automation",
		},
		Steps: []automation.Step{
			bashStep(fmt.Sprintf(`echo "${MY_VAR:-empty}" > %s`, outFile1)),
			bashStep(fmt.Sprintf(`echo "${MY_VAR:-empty}" > %s`, outFile2)),
		},
		FilePath: "/fake/path/automation.yaml",
	}

	exec := newExecutor(dir, newDiscovery(nil))

	if err := exec.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got1, _ := os.ReadFile(outFile1)
	if strings.TrimSpace(string(got1)) != "from_automation" {
		t.Errorf("step1 got %q, want %q", strings.TrimSpace(string(got1)), "from_automation")
	}
	got2, _ := os.ReadFile(outFile2)
	if strings.TrimSpace(string(got2)) != "from_automation" {
		t.Errorf("step2 got %q, want %q", strings.TrimSpace(string(got2)), "from_automation")
	}
}

func TestAutomationEnv_StepOverrides(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	a := &automation.Automation{
		Name: "auto-env-override",
		Env: map[string]string{
			"MY_VAR": "from_automation",
		},
		Steps: []automation.Step{
			{
				Type:  automation.StepTypeBash,
				Value: fmt.Sprintf(`echo "${MY_VAR}" > %s`, outFile),
				Env:   map[string]string{"MY_VAR": "from_step"},
			},
		},
		FilePath: "/fake/path/automation.yaml",
	}

	exec := newExecutor(dir, newDiscovery(nil))

	if err := exec.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "from_step" {
		t.Errorf("step got %q, want %q (step env should override automation env)", strings.TrimSpace(string(got)), "from_step")
	}
}

func TestAutomationEnv_DoesNotPropagateToRunStep(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	child := &automation.Automation{
		Name: "child",
		Steps: []automation.Step{
			bashStep(fmt.Sprintf(`echo "${MY_VAR:-empty}" > %s`, outFile)),
		},
		FilePath: "/fake/path/automation.yaml",
	}

	parent := &automation.Automation{
		Name: "parent",
		Env: map[string]string{
			"MY_VAR": "from_parent",
		},
		Steps: []automation.Step{
			runStep("child"),
		},
		FilePath: "/fake/path/automation.yaml",
	}

	disc := newDiscovery(map[string]*automation.Automation{
		"parent": parent,
		"child":  child,
	})
	exec := newExecutor(dir, disc)

	if err := exec.Run(parent, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(outFile)
	if strings.TrimSpace(string(got)) != "empty" {
		t.Errorf("child got %q, want %q (automation env should not propagate to sub-automations)", strings.TrimSpace(string(got)), "empty")
	}
}

func TestBuildStepEnv_WithAutomationEnv(t *testing.T) {
	env := BuildStepEnv(nil, nil, map[string]string{"AUTO_KEY": "auto_val"}, nil)
	if env == nil {
		t.Fatal("expected non-nil env when automation env is set")
	}

	found := false
	for _, e := range env {
		if e == "AUTO_KEY=auto_val" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected AUTO_KEY=auto_val in env")
	}
}

func TestBuildStepEnv_AutomationEnvOverriddenByStepEnv(t *testing.T) {
	env := BuildStepEnv(
		nil,
		nil,
		map[string]string{"SHARED": "auto"},
		map[string]string{"SHARED": "step"},
	)
	if env == nil {
		t.Fatal("expected non-nil env")
	}

	var sharedVals []string
	for _, e := range env {
		if strings.HasPrefix(e, "SHARED=") {
			sharedVals = append(sharedVals, e)
		}
	}
	if len(sharedVals) != 2 {
		t.Fatalf("expected 2 SHARED entries (auto + step), got %d: %v", len(sharedVals), sharedVals)
	}
	if sharedVals[len(sharedVals)-1] != "SHARED=step" {
		t.Errorf("last SHARED entry should be step override, got %q", sharedVals[len(sharedVals)-1])
	}
}
