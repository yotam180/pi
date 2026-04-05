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
	if !strings.Contains(out, "name (string, required)") {
		t.Errorf("expected required input with type, got:\n%s", out)
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
