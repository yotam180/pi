package executor

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func firstStep(subSteps ...automation.Step) automation.Step {
	return automation.Step{First: subSteps}
}

func firstStepPiped(subSteps ...automation.Step) automation.Step {
	return automation.Step{First: subSteps, Pipe: true}
}

func firstStepIf(cond string, subSteps ...automation.Step) automation.Step {
	return automation.Step{First: subSteps, If: cond}
}

func TestFirstBlock_FirstMatches(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo first", "os.macos"),
			bashStepIf("echo second", "os.linux"),
			bashStep("echo fallback"),
		),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "first" {
		t.Errorf("output = %q, want %q", got, "first")
	}
}

func TestFirstBlock_MiddleMatches(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo first", "os.macos"),
			bashStepIf("echo second", "os.linux"),
			bashStep("echo fallback"),
		),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("linux"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "second" {
		t.Errorf("output = %q, want %q", got, "second")
	}
}

func TestFirstBlock_FallbackMatches(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo first", "os.macos"),
			bashStepIf("echo second", "os.linux"),
			bashStep("echo fallback"),
		),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("windows"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "fallback" {
		t.Errorf("output = %q, want %q", got, "fallback")
	}
}

func TestFirstBlock_NoneMatch(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo first", "os.macos"),
			bashStepIf("echo second", "os.linux"),
		),
		bashStep("echo after"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("windows"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "after" {
		t.Errorf("output = %q, want %q", got, "after")
	}
}

func TestFirstBlock_MixedWithRegularSteps(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo before"),
		firstStep(
			bashStepIf("echo first", "os.linux"),
			bashStep("echo fallback"),
		),
		bashStep("echo after"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 output lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "before" {
		t.Errorf("line[0] = %q, want %q", lines[0], "before")
	}
	if lines[1] != "fallback" {
		t.Errorf("line[1] = %q, want %q", lines[1], "fallback")
	}
	if lines[2] != "after" {
		t.Errorf("line[2] = %q, want %q", lines[2], "after")
	}
}

func TestFirstBlock_PipeToNext(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStepPiped(
			bashStepIf("echo WRONG", "os.linux"),
			bashStep("echo piped-data"),
		),
		bashStep("cat"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "piped-data" {
		t.Errorf("output = %q, want %q", got, "piped-data")
	}
}

func TestFirstBlock_PipeToNextNoneMatch(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo upstream"),
		firstStepPiped(
			bashStepIf("echo WRONG", "os.linux"),
		),
		bashStep("cat"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When no sub-step matches in a piped first: block, the pipe buffer is empty
	got := strings.TrimSpace(stdout.String())
	if got != "" {
		t.Errorf("output = %q, want empty (no match, pipe consumed)", got)
	}
}

func TestFirstBlock_OuterIfSkipsEntireBlock(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStepIf("os.linux",
			bashStep("echo should-not-run"),
		),
		bashStep("echo after"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "after" {
		t.Errorf("output = %q, want %q", got, "after")
	}
}

func TestFirstBlock_SubStepWithRunStep(t *testing.T) {
	dir := t.TempDir()
	helper := newAutomation("helper", bashStep("echo from-helper"))
	disc := newDiscovery(map[string]*automation.Automation{"helper": helper})

	a := newAutomation("test",
		firstStep(
			bashStepIf("echo wrong", "os.linux"),
			automation.Step{Type: automation.StepTypeRun, Value: "helper"},
		),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, disc, fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "from-helper" {
		t.Errorf("output = %q, want %q", got, "from-helper")
	}
}

func TestFirstBlock_ExitError(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStep("exit 42"),
		),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from exit 42")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if exitErr.Code != 42 {
		t.Errorf("exit code = %d, want 42", exitErr.Code)
	}
}

func TestFirstBlock_InInstallPhase(t *testing.T) {
	dir := t.TempDir()

	// Use a marker file to simulate test-before-install (fail) and test-after-install (pass)
	markerFile := dir + "/installed"
	inst := &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "test -f " + markerFile},
		Run: automation.InstallPhase{
			Steps: []automation.Step{
				firstStep(
					bashStepIf("echo wrong-path", "os.linux"),
					bashStep("touch " + markerFile),
				),
			},
		},
		Version: "echo 1.0.0",
	}

	a := newInstallerAutomation("test-installer", inst)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  newDiscovery(nil),
		Stdout:     stdout,
		Stderr:     stderr,
		RuntimeEnv: fakeRuntimeEnv("darwin"),
		Silent:     true,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the marker file was created (fallback sub-step ran)
	if _, err := os.Stat(markerFile); err != nil {
		t.Errorf("expected marker file to exist (install ran), got error: %v", err)
	}
}

func TestFirstBlock_SilentSubStep(t *testing.T) {
	dir := t.TempDir()
	silentBash := automation.Step{Type: automation.StepTypeBash, Value: "echo silent-output", Silent: true}
	a := newAutomation("test",
		firstStep(silentBash),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := stdout.String()
	if strings.Contains(got, "silent-output") {
		t.Errorf("output should be suppressed, got %q", got)
	}
}

func TestFirstBlock_LoudOverridesSilent(t *testing.T) {
	dir := t.TempDir()
	silentBash := automation.Step{Type: automation.StepTypeBash, Value: "echo loud-output", Silent: true}
	a := newAutomation("test",
		firstStep(silentBash),
	)
	stdout := &bytes.Buffer{}
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    stdout,
		Stderr:    io.Discard,
		Loud:      true,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "loud-output") {
		t.Errorf("output should contain loud-output when Loud=true, got %q", got)
	}
}
