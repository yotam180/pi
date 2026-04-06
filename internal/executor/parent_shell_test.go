package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestParentShell_WritesToEvalFile(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	step := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "source venv/bin/activate",
		ParentShell: true,
	}
	a := newAutomation("test", step)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "source venv/bin/activate" {
		t.Errorf("eval file content = %q, want %q", content, "source venv/bin/activate")
	}

	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "parent:") {
		t.Errorf("expected parent: trace in stderr, got: %q", stderrOut)
	}
}

func TestParentShell_MultipleStepsAppend(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	step1 := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "cd /tmp",
		ParentShell: true,
	}
	step2 := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "export FOO=bar",
		ParentShell: true,
	}
	a := newAutomation("test", step1, step2)

	exec := newExecutor(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines in eval file, got %d: %q", len(lines), string(data))
	}
	if lines[0] != "cd /tmp" {
		t.Errorf("line 1 = %q, want %q", lines[0], "cd /tmp")
	}
	if lines[1] != "export FOO=bar" {
		t.Errorf("line 2 = %q, want %q", lines[1], "export FOO=bar")
	}
}

func TestParentShell_MixedWithNormalSteps(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	normalStep := bashStep("echo hello-from-normal")
	parentStep := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "cd /tmp",
		ParentShell: true,
	}
	a := newAutomation("test", normalStep, parentStep)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "hello-from-normal") {
		t.Errorf("normal step output missing, got: %q", stdout.String())
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "cd /tmp" {
		t.Errorf("eval file content = %q, want %q", string(data), "cd /tmp")
	}
}

func TestParentShell_NoEvalFile_Warning(t *testing.T) {
	dir := t.TempDir()

	step := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "source venv/bin/activate",
		ParentShell: true,
	}
	a := newAutomation("test", step)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "parent_shell step skipped") {
		t.Errorf("expected skip warning in stderr, got: %q", stderrOut)
	}
	if !strings.Contains(stderrOut, "pi shell") {
		t.Errorf("expected 'pi shell' hint in stderr, got: %q", stderrOut)
	}
}

func TestParentShell_SkippedByCondition(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	step := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "source venv/bin/activate",
		ParentShell: true,
		If:          "os.windows",
	}
	a := newAutomation("test", step)

	exec := newExecutor(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(evalFile); !os.IsNotExist(err) {
		data, _ := os.ReadFile(evalFile)
		if len(data) > 0 {
			t.Errorf("eval file should be empty for skipped parent_shell step, got: %q", string(data))
		}
	}
}

func TestAppendToParentEval(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	if err := AppendToParentEval(evalFile, "line1"); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := AppendToParentEval(evalFile, "line2"); err != nil {
		t.Fatalf("second append: %v", err)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	expected := "line1\nline2\n"
	if string(data) != expected {
		t.Errorf("eval file content = %q, want %q", string(data), expected)
	}
}

func TestParentShell_PythonStep_Rejected(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	step := automation.Step{
		Type:        automation.StepTypePython,
		Value:       "print('hello')",
		ParentShell: true,
	}
	a := newAutomation("test", step)

	exec := newExecutor(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for parent_shell on python step")
	}
	if !strings.Contains(err.Error(), "does not support parent_shell") {
		t.Errorf("error = %q, want mention of 'does not support parent_shell'", err.Error())
	}
}

func TestParentShell_TypeScriptStep_Rejected(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	step := automation.Step{
		Type:        automation.StepTypeTypeScript,
		Value:       "console.log('hi')",
		ParentShell: true,
	}
	a := newAutomation("test", step)

	exec := newExecutor(dir, newDiscovery(nil))
	exec.ParentEvalFile = evalFile

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for parent_shell on typescript step")
	}
	if !strings.Contains(err.Error(), "does not support parent_shell") {
		t.Errorf("error = %q, want mention of 'does not support parent_shell'", err.Error())
	}
}

func TestRegistry_SupportsParentShell(t *testing.T) {
	reg := NewDefaultRegistry()

	if !reg.StepTypeSupportsParentShell(automation.StepTypeBash) {
		t.Error("bash should support parent_shell")
	}
	if reg.StepTypeSupportsParentShell(automation.StepTypePython) {
		t.Error("python should not support parent_shell")
	}
	if reg.StepTypeSupportsParentShell(automation.StepTypeTypeScript) {
		t.Error("typescript should not support parent_shell")
	}
	if reg.StepTypeSupportsParentShell(automation.StepTypeRun) {
		t.Error("run should not support parent_shell")
	}
}
