package executor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestPipe_BashToBash(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo hello world"),
		bashStep("tr a-z A-Z"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO WORLD" {
		t.Errorf("output = %q, want %q", got, "HELLO WORLD")
	}
}

func TestPipe_BashToPython(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo hello world"),
		pythonStep("import sys; print(sys.stdin.read().strip().upper())"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO WORLD" {
		t.Errorf("output = %q, want %q", got, "HELLO WORLD")
	}
}

func TestPipe_ThreeStepChain(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo apple banana cherry"),
		pipedBashStep("tr ' ' '\\n'"),
		bashStep("sort"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	want := "apple\nbanana\ncherry"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPipe_FailureInMiddleStopsExecution(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	a := newAutomation("test",
		pipedBashStep("echo data"),
		pipedBashStep("exit 1"),
		bashStep("cat > "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error when piped step fails")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.Code)
	}

	if _, err := os.Stat(outFile); err == nil {
		t.Error("third step should not have run, but output file exists")
	}
}

func TestPipe_StderrPassesThrough(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo stdout-data; echo stderr-data >&2"),
		bashStep("cat"),
	)
	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotStdout := strings.TrimSpace(stdout.String())
	if gotStdout != "stdout-data" {
		t.Errorf("stdout = %q, want %q", gotStdout, "stdout-data")
	}

	gotStderr := stderr.String()
	if !strings.Contains(gotStderr, "stderr-data") {
		t.Errorf("stderr = %q, want it to contain %q", gotStderr, "stderr-data")
	}
}

func TestPipe_NoPipeDefaultBehavior(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo step1"),
		bashStep("echo step2"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	want := "step1\nstep2"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPipe_LastStepPipeToNextIsNoop(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo piped-first"),
		automation.Step{Type: automation.StepTypeBash, Value: "echo last-step", Pipe: true},
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "last-step" {
		t.Errorf("output = %q, want %q", got, "last-step")
	}
}

func TestPipe_PythonToBash(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	a := newAutomation("test",
		pipedPythonStep("print('hello from python')"),
		bashStep("tr a-z A-Z"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO FROM PYTHON" {
		t.Errorf("output = %q, want %q", got, "HELLO FROM PYTHON")
	}
}

func TestPipe_ThroughRunStep(t *testing.T) {
	dir := t.TempDir()
	inner := newAutomation("upper", bashStep("tr a-z A-Z"))
	outer := newAutomation("test",
		pipedBashStep("echo hello"),
		runStep("upper"),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"upper": inner,
		"test":  outer,
	})
	exec, stdout, _ := newExecutorWithCapture(dir, disc)

	err := exec.Run(outer, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO" {
		t.Errorf("output = %q, want %q", got, "HELLO")
	}
}

func TestPipe_MultilineData(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep(`printf "line1\nline2\nline3\n"`),
		bashStep("wc -l | tr -d ' '"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "3" {
		t.Errorf("output = %q, want %q", got, "3")
	}
}
