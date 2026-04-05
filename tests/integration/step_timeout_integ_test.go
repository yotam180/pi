package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStepTimeout_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-timeout")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"fast", "slow", "mixed", "timeout-info"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got: %s", name, out)
		}
	}
}

func TestStepTimeout_FastCompletesWithinTimeout(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-timeout")
	stdout, _, code := runPiSplit(t, dir, "run", "fast")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "completed-fast") {
		t.Errorf("expected 'completed-fast' in output, got: %s", stdout)
	}
}

func TestStepTimeout_SlowExceedsTimeout(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-timeout")
	_, code := runPi(t, dir, "run", "slow")
	if code == 0 {
		t.Fatal("expected non-zero exit for timed-out step")
	}
	if code != 124 {
		t.Errorf("expected exit code 124 (timeout), got %d", code)
	}
}

func TestStepTimeout_MixedTimedAndUntimed(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-timeout")
	stdout, _, code := runPiSplit(t, dir, "run", "mixed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "no-timeout") {
		t.Errorf("expected 'no-timeout' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "with-timeout") {
		t.Errorf("expected 'with-timeout' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "also-no-timeout") {
		t.Errorf("expected 'also-no-timeout' in output, got: %s", stdout)
	}
}

func TestStepTimeout_InfoShowsAnnotation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-timeout")
	out, code := runPi(t, dir, "info", "timeout-info")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected 'Step details:' section, got: %s", out)
	}
	if !strings.Contains(out, "timeout: 30s") {
		t.Errorf("expected 'timeout: 30s' annotation, got: %s", out)
	}
	if !strings.Contains(out, "timeout: 5m") {
		t.Errorf("expected 'timeout: 5m' annotation, got: %s", out)
	}
	if !strings.Contains(out, "silent") {
		t.Errorf("expected 'silent' annotation on step 2, got: %s", out)
	}
}
