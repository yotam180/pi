package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInfo_BasicAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Name:") {
		t.Errorf("expected Name header, got:\n%s", out)
	}
	if !strings.Contains(out, "greet") {
		t.Errorf("expected automation name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "No inputs.") {
		t.Errorf("expected 'No inputs.' for automation without inputs, got:\n%s", out)
	}
}

func TestInfo_WithInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Inputs:") {
		t.Errorf("expected Inputs header, got:\n%s", out)
	}
	if !strings.Contains(out, "name (position 1, string, required)") {
		t.Errorf("expected required input with position and type, got:\n%s", out)
	}
	if !strings.Contains(out, "Who to greet") {
		t.Errorf("expected input description, got:\n%s", out)
	}
	if !strings.Contains(out, `default: "hello"`) {
		t.Errorf("expected default value shown, got:\n%s", out)
	}
	if !strings.Contains(out, "optional") {
		t.Errorf("expected 'optional' for optional input, got:\n%s", out)
	}
}

func TestInfo_NotFound(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "info", "nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown automation, got 0: %s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in error, got: %s", out)
	}
}

func TestInfo_NoArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	_, code := runPi(t, dir, "info")
	if code == 0 {
		t.Fatal("expected non-zero exit when no argument provided")
	}
}

func TestInfo_InstallerAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "install-marker")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Type:         installer") {
		t.Errorf("expected 'Type: installer', got:\n%s", out)
	}
	if !strings.Contains(out, "Install lifecycle:") {
		t.Errorf("expected install lifecycle section, got:\n%s", out)
	}
	if !strings.Contains(out, "test:") {
		t.Errorf("expected test phase, got:\n%s", out)
	}
	if !strings.Contains(out, "run:") {
		t.Errorf("expected run phase, got:\n%s", out)
	}
	if !strings.Contains(out, "verify: (re-runs test)") {
		t.Errorf("expected default verify, got:\n%s", out)
	}
	if !strings.Contains(out, "version:") {
		t.Errorf("expected version phase, got:\n%s", out)
	}
}

func TestInfo_InstallerNoVersion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "install-no-version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Type:         installer") {
		t.Errorf("expected installer type, got:\n%s", out)
	}
	if strings.Contains(out, "version:") {
		t.Errorf("expected no version line for installer without version, got:\n%s", out)
	}
}

func TestInfo_InstallerWithInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "install-marker")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Inputs:") {
		t.Errorf("expected Inputs section, got:\n%s", out)
	}
	if !strings.Contains(out, "path") {
		t.Errorf("expected 'path' input, got:\n%s", out)
	}
}

func TestInfo_FirstBlockSubStepDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPi(t, dir, "info", "pick-platform")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "a.") {
		t.Errorf("expected lettered sub-steps (a.), got:\n%s", out)
	}
	if !strings.Contains(out, "b.") {
		t.Errorf("expected lettered sub-steps (b.), got:\n%s", out)
	}
}
