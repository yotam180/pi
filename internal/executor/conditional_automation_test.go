package executor

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestAutomationIf_TrueExecutes(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationWithIf("macos-tool", "os.macos",
		bashStep("echo ran"),
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

func TestAutomationIf_FalseSkips(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationWithIf("macos-tool", "os.macos",
		bashStep("echo should-not-run"),
	)
	exec, stdout, stderr := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("linux"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "" {
		t.Errorf("expected no stdout, got %q", stdout.String())
	}
	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "[skipped] macos-tool") {
		t.Errorf("expected skip message in stderr, got %q", stderrStr)
	}
	if !strings.Contains(stderrStr, "condition: os.macos") {
		t.Errorf("expected condition in skip message, got %q", stderrStr)
	}
}

func TestAutomationIf_NoIfAlwaysRuns(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("always-run",
		bashStep("echo hello"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "hello" {
		t.Errorf("output = %q, want %q", got, "hello")
	}
}

func TestAutomationIf_RunStepCallsSkippedAutomation(t *testing.T) {
	dir := t.TempDir()
	child := newAutomationWithIf("child", "os.macos",
		bashStep("echo child-ran"),
	)
	parent := newAutomation("parent",
		bashStep("echo before"),
		runStep("child"),
		bashStep("echo after"),
	)
	disc := newDiscovery(map[string]*automation.Automation{
		"child":  child,
		"parent": parent,
	})
	exec, stdout, stderr := newExecutorWithEnv(dir, disc, fakeRuntimeEnv("linux"))

	err := exec.Run(parent, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if !strings.Contains(got, "before") {
		t.Errorf("expected 'before' in output, got %q", got)
	}
	if !strings.Contains(got, "after") {
		t.Errorf("expected 'after' in output, got %q", got)
	}
	if strings.Contains(got, "child-ran") {
		t.Errorf("child should not have run, got %q", got)
	}
	if !strings.Contains(stderr.String(), "[skipped] child") {
		t.Errorf("expected skip message for child, got %q", stderr.String())
	}
}

func TestAutomationIf_RunStepCallsExecutedAutomation(t *testing.T) {
	dir := t.TempDir()
	child := newAutomationWithIf("child", "os.macos",
		bashStep("echo child-ran"),
	)
	parent := newAutomation("parent",
		bashStep("echo before"),
		runStep("child"),
		bashStep("echo after"),
	)
	disc := newDiscovery(map[string]*automation.Automation{
		"child":  child,
		"parent": parent,
	})
	exec, stdout, _ := newExecutorWithEnv(dir, disc, fakeRuntimeEnv("darwin"))

	err := exec.Run(parent, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if !strings.Contains(got, "before") {
		t.Errorf("expected 'before' in output, got %q", got)
	}
	if !strings.Contains(got, "child-ran") {
		t.Errorf("expected 'child-ran' in output, got %q", got)
	}
	if !strings.Contains(got, "after") {
		t.Errorf("expected 'after' in output, got %q", got)
	}
}

func TestAutomationIf_ComplexCondition(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationWithIf("complex", "os.macos and os.arch.arm64",
		bashStep("echo ran"),
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

func TestAutomationIf_SkipDoesNotTriggerCircularDependency(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationWithIf("self-ref", "os.macos",
		runStep("self-ref"),
	)
	disc := newDiscovery(map[string]*automation.Automation{
		"self-ref": a,
	})
	exec, _, stderr := newExecutorWithEnv(dir, disc, fakeRuntimeEnv("linux"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("expected no error (skipped before push), got: %v", err)
	}
	if !strings.Contains(stderr.String(), "[skipped]") {
		t.Errorf("expected skip message, got %q", stderr.String())
	}
}
