package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestStepTrace_DefaultBehavior(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo hello"),
		bashStep("echo world"),
	)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "→ bash: echo hello") {
		t.Errorf("expected trace line for first step, got: %q", got)
	}
	if !strings.Contains(got, "→ bash: echo world") {
		t.Errorf("expected trace line for second step, got: %q", got)
	}
}

func TestStepTrace_RunStep(t *testing.T) {
	dir := t.TempDir()
	child := newAutomation("child", bashStep("echo child"))
	a := newAutomation("parent", runStep("child"))

	disc := newDiscovery(map[string]*automation.Automation{
		"child": child,
	})
	exec, _, stderr := newExecutorWithCapture(dir, disc)
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "→ run: child") {
		t.Errorf("expected trace for run step, got: %q", got)
	}
	if !strings.Contains(got, "→ bash: echo child") {
		t.Errorf("expected trace for child bash step, got: %q", got)
	}
}

func TestStepTrace_SilentStep(t *testing.T) {
	dir := t.TempDir()
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo silent-output",
		Silent: true,
	}
	normalStep := bashStep("echo normal-output")
	a := newAutomation("test", silentStep, normalStep)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if strings.Contains(stderrOut, "echo silent-output") {
		t.Errorf("silent step should not print trace, got: %q", stderrOut)
	}
	if !strings.Contains(stderrOut, "→ bash: echo normal-output") {
		t.Errorf("normal step should print trace, got: %q", stderrOut)
	}

	stdoutOut := stdout.String()
	if strings.Contains(stdoutOut, "silent-output") {
		t.Errorf("silent step should suppress stdout, got: %q", stdoutOut)
	}
	if !strings.Contains(stdoutOut, "normal-output") {
		t.Errorf("normal step should print stdout, got: %q", stdoutOut)
	}
}

func TestStepTrace_LoudOverridesSilent(t *testing.T) {
	dir := t.TempDir()
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo loud-output",
		Silent: true,
	}
	a := newAutomation("test", silentStep)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	exec.Loud = true
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "→ bash: echo loud-output") {
		t.Errorf("loud should override silent and print trace, got: %q", stderrOut)
	}

	stdoutOut := stdout.String()
	if !strings.Contains(stdoutOut, "loud-output") {
		t.Errorf("loud should override silent and print output, got: %q", stdoutOut)
	}
}

func TestStepTrace_SilentStepStillExecutes(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "proof.txt")
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo done > " + outFile,
		Silent: true,
	}
	a := newAutomation("test", silentStep)

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("silent step should still execute, but output file missing: %v", err)
	}
}

func TestStepTrace_SilentPipeCapture(t *testing.T) {
	dir := t.TempDir()
	silentPiped := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo piped-data",
		Silent: true,
		PipeTo: "next",
	}
	receiver := bashStep("cat")
	a := newAutomation("test", silentPiped, receiver)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if strings.Contains(stderrOut, "echo piped-data") {
		t.Errorf("silent piped step should not print trace, got: %q", stderrOut)
	}

	stdoutOut := strings.TrimSpace(stdout.String())
	if stdoutOut != "piped-data" {
		t.Errorf("pipe should still work for silent steps, got: %q", stdoutOut)
	}
}
