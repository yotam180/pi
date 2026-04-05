package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_AllSatisfied(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")

	if code != 1 {
		t.Fatalf("expected exit 1 (some requirements missing), got %d: %s", code, out)
	}

	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ for satisfied requirements, got:\n%s", out)
	}

	if !strings.Contains(out, "needs-bash") {
		t.Errorf("expected needs-bash in output, got:\n%s", out)
	}
}

func TestDoctor_ShowsMissingRequirements(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	if !strings.Contains(out, "needs-impossible") {
		t.Errorf("expected needs-impossible in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✗") {
		t.Errorf("expected ✗ for missing requirements, got:\n%s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' for missing command, got:\n%s", out)
	}
}

func TestDoctor_ShowsVersionMismatch(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	if !strings.Contains(out, "needs-impossible-version") {
		t.Errorf("expected needs-impossible-version in output, got:\n%s", out)
	}
	if !strings.Contains(out, "python >= 99.0") {
		t.Errorf("expected 'python >= 99.0' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "need >= 99.0") {
		t.Errorf("expected version mismatch message, got:\n%s", out)
	}
}

func TestDoctor_SkipsNoRequiresAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, _ := runPi(t, dir, "doctor")

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "no-requires" {
			t.Errorf("doctor should skip automations without requires:, but found 'no-requires' in output:\n%s", out)
		}
	}
}

func TestDoctor_ShowsDetectedVersion(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, _ := runPi(t, dir, "doctor")

	if !strings.Contains(out, "needs-python") {
		t.Errorf("expected needs-python in output, got:\n%s", out)
	}
	if !strings.Contains(out, "(") || !strings.Contains(out, ")") {
		t.Errorf("expected version in parentheses, got:\n%s", out)
	}
}

func TestDoctor_ShowsInstallHint(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	if !strings.Contains(out, "install") || !strings.Contains(out, "python") {
		t.Errorf("expected install hint for python, got:\n%s", out)
	}
}

func TestDoctor_HealthyWorkspace(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: healthy\n"), 0644); err != nil {
		t.Fatal(err)
	}
	piDir := filepath.Join(dir, ".pi")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(piDir, "needs-bash.yaml"), []byte(`name: needs-bash
description: Needs bash
requires:
  - command: bash
steps:
  - bash: echo ok
`), 0644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "doctor")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ for satisfied requirements, got:\n%s", out)
	}
	if strings.Contains(out, "✗") {
		t.Errorf("should not have ✗ in healthy workspace, got:\n%s", out)
	}
}
