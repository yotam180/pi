package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestConditionalStep_TrueExecutes(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo ran", "os.macos"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "ran" {
		t.Errorf("output = %q, want %q", got, "ran")
	}
}

func TestConditionalStep_FalseSkips(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo should-not-run", "os.linux"),
		bashStep("echo should-run"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "should-run" {
		t.Errorf("output = %q, want %q", got, "should-run")
	}
}

func TestConditionalStep_NoIfAlwaysRuns(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo always"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "always" {
		t.Errorf("output = %q, want %q", got, "always")
	}
}

func TestConditionalStep_AllSkipped(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo a", "os.linux"),
		bashStepIf("echo b", "os.windows"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "" {
		t.Errorf("expected no output, got %q", got)
	}
}

func TestConditionalStep_NotOperator(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo ran", "not os.linux"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "ran" {
		t.Errorf("output = %q, want %q", got, "ran")
	}
}

func TestConditionalStep_ComplexExpression(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo ran", "os.macos and os.arch.arm64"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "ran" {
		t.Errorf("output = %q, want %q", got, "ran")
	}
}

func TestConditionalStep_ComplexExpressionFalse(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStepIf("echo should-not-run", "os.macos and os.arch.amd64"),
	)
	// arm64, not amd64
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "" {
		t.Errorf("expected no output, got %q", got)
	}
}

func TestConditionalStep_MixedConditionalAndUnconditional(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo step1"),
		bashStepIf("echo step2-skipped", "os.linux"),
		bashStep("echo step3"),
		bashStepIf("echo step4-ran", "os.macos"),
		bashStep("echo step5"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	want := "step1\nstep3\nstep4-ran\nstep5"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestConditionalStep_PipeSkipped_PassesThrough(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo hello", PipeTo: "next"},
		automation.Step{Type: automation.StepTypeBash, Value: "tr a-z A-Z", PipeTo: "next", If: "os.linux"},
		bashStep("cat"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	// Middle step is skipped, so input passes through unchanged
	if got != "hello" {
		t.Errorf("output = %q, want %q (pipe should pass through skipped step)", got, "hello")
	}
}

func TestConditionalStep_PipeSkipped_NoPriorPipe(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo skipped-source", PipeTo: "next", If: "os.linux"},
		bashStep("echo fallback"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "fallback" {
		t.Errorf("output = %q, want %q", got, "fallback")
	}
}

func TestConditionalStep_PipeSkipped_MultipleSkipped(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo data", PipeTo: "next"},
		automation.Step{Type: automation.StepTypeBash, Value: "tr a-z A-Z", PipeTo: "next", If: "os.linux"},
		automation.Step{Type: automation.StepTypeBash, Value: "rev", PipeTo: "next", If: "os.windows"},
		bashStep("cat"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "data" {
		t.Errorf("output = %q, want %q (pipe should pass through multiple skipped steps)", got, "data")
	}
}

func TestConditionalStep_FileExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=yes\n"), 0644)

	a := newAutomation("test",
		bashStepIf("echo found", `file.exists(".env")`),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "found" {
		t.Errorf("output = %q, want %q", got, "found")
	}
}

func TestConditionalStep_FileNotExists(t *testing.T) {
	dir := t.TempDir()

	a := newAutomation("test",
		bashStepIf("echo should-not-run", `file.exists(".env")`),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "" {
		t.Errorf("expected no output, got %q", got)
	}
}
