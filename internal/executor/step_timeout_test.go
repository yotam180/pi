package executor

import (
	"strings"
	"testing"
	"time"

	"github.com/vyper-tooling/pi/internal/automation"
)

func bashStepWithTimeout(value string, timeout time.Duration) automation.Step {
	return automation.Step{
		Type:       automation.StepTypeBash,
		Value:      value,
		Timeout:    timeout,
		TimeoutRaw: timeout.String(),
	}
}

func TestStep_Timeout_NoTimeout_RunsNormally(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("normal", bashStep(`echo "fast"`))
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "fast") {
		t.Errorf("output = %q, want to contain 'fast'", stdout.String())
	}
}

func TestStep_Timeout_NotExceeded(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("fast-with-timeout",
		bashStepWithTimeout(`echo "quick"`, 5*time.Second),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "quick") {
		t.Errorf("output = %q, want to contain 'quick'", stdout.String())
	}
}

func TestStep_Timeout_Exceeded(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("slow",
		bashStepWithTimeout(`sleep 30`, 200*time.Millisecond),
	)
	exec, _, _ := newExecutorWithCapture(dir, nil)

	start := time.Now()
	err := exec.Run(a, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for timed-out step")
	}
	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != TimeoutExitCode {
		t.Errorf("exit code = %d, want %d", exitErr.Code, TimeoutExitCode)
	}
	if elapsed > 3*time.Second {
		t.Errorf("timeout took too long: %v (expected ~200ms)", elapsed)
	}
}

func TestStep_Timeout_StopsExecution(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("chain",
		bashStepWithTimeout(`sleep 30`, 200*time.Millisecond),
		bashStep(`echo "should not run"`),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(stdout.String(), "should not run") {
		t.Error("second step should not have executed after timeout")
	}
}

func TestStep_Timeout_WithPipeTo(t *testing.T) {
	dir := t.TempDir()
	step := bashStepWithTimeout(`echo "piped data"`, 5*time.Second)
	step.Pipe = true
	a := newAutomation("pipe-timeout",
		step,
		bashStep(`cat`),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "piped data") {
		t.Errorf("output = %q, want to contain 'piped data'", stdout.String())
	}
}

func TestStep_Timeout_WithSilent(t *testing.T) {
	dir := t.TempDir()
	step := bashStepWithTimeout(`echo "silent output"`, 5*time.Second)
	step.Silent = true
	a := newAutomation("silent-timeout", step)
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdout.String(), "silent output") {
		t.Error("silent step should suppress output")
	}
}

func TestStep_Timeout_SkippedByCondition(t *testing.T) {
	dir := t.TempDir()
	step := bashStepWithTimeout(`sleep 30`, 200*time.Millisecond)
	step.If = "os.windows"

	a := newAutomation("skipped-timeout", step)
	exec, _, _ := newExecutorWithEnv(dir, nil, fakeRuntimeEnv("linux"))

	start := time.Now()
	err := exec.Run(a, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed > 1*time.Second {
		t.Errorf("skipped step should not have waited for timeout: %v", elapsed)
	}
}

func TestStep_Timeout_MultipleSteps_OnlyTimedOutStepKilled(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("multi",
		bashStep(`echo "first"`),
		bashStepWithTimeout(`sleep 30`, 200*time.Millisecond),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, nil)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stdout.String(), "first") {
		t.Error("first step should have completed before timeout")
	}
}
