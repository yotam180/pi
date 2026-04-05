package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStepVisibility_DefaultTraceLines(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	stdout, stderr, code := runPiSplit(t, dir, "run", "normal")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !strings.Contains(stderr, "→ bash: echo \"step-one-output\"") {
		t.Errorf("expected trace line for step 1 in stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "→ bash: echo \"step-two-output\"") {
		t.Errorf("expected trace line for step 2 in stderr, got:\n%s", stderr)
	}

	if !strings.Contains(stdout, "step-one-output") {
		t.Errorf("expected step 1 output in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "step-two-output") {
		t.Errorf("expected step 2 output in stdout, got:\n%s", stdout)
	}
}

func TestStepVisibility_SilentSuppressesTraceAndOutput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	stdout, stderr, code := runPiSplit(t, dir, "run", "mixed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !strings.Contains(stderr, "→ bash: echo \"visible-output\"") {
		t.Errorf("expected trace for visible step, got stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "hidden-output") {
		t.Errorf("silent step trace should be suppressed, got stderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "→ bash: echo \"also-visible\"") {
		t.Errorf("expected trace for last step, got stderr:\n%s", stderr)
	}

	if !strings.Contains(stdout, "visible-output") {
		t.Errorf("expected visible output, got stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, "hidden-output") {
		t.Errorf("silent step output should be suppressed, got stdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "also-visible") {
		t.Errorf("expected also-visible output, got stdout:\n%s", stdout)
	}
}

func TestStepVisibility_LoudOverridesSilent(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	stdout, stderr, code := runPiSplit(t, dir, "run", "--loud", "mixed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !strings.Contains(stderr, "→ bash: echo \"hidden-output\"") {
		t.Errorf("loud should show trace for silent step, got stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "hidden-output") {
		t.Errorf("loud should show output for silent step, got stdout:\n%s", stdout)
	}
}

func TestStepVisibility_AllSilentNoOutput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	stdout, stderr, code := runPiSplit(t, dir, "run", "all-silent")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if strings.Contains(stderr, "→") {
		t.Errorf("all-silent should produce no trace lines, got stderr:\n%s", stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("all-silent should produce no stdout, got:\n%s", stdout)
	}
}

func TestStepVisibility_AllSilentLoudShowsAll(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	stdout, stderr, code := runPiSplit(t, dir, "run", "--loud", "all-silent")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if !strings.Contains(stderr, "→ bash: echo \"quiet-one\"") {
		t.Errorf("loud should show trace for silent step 1, got stderr:\n%s", stderr)
	}
	if !strings.Contains(stderr, "→ bash: echo \"quiet-two\"") {
		t.Errorf("loud should show trace for silent step 2, got stderr:\n%s", stderr)
	}
	if !strings.Contains(stdout, "quiet-one") {
		t.Errorf("loud should show output for silent step 1, got stdout:\n%s", stdout)
	}
}

func TestStepVisibility_InfoShowsSilent(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-visibility")
	out, code := runPi(t, dir, "info", "mixed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "silent") {
		t.Errorf("expected 'silent' annotation in info output, got:\n%s", out)
	}
}
