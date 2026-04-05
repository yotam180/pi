package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInputs_PositionalArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "alice")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hello alice" {
		t.Errorf("output = %q, want %q", trimmed, "hello alice")
	}
}

func TestInputs_PositionalBothArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "bob", "hi")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hi bob" {
		t.Errorf("output = %q, want %q", trimmed, "hi bob")
	}
}

func TestInputs_WithFlags(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "--with", "name=charlie", "--with", "greeting=hey")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hey charlie" {
		t.Errorf("output = %q, want %q", trimmed, "hey charlie")
	}
}

func TestInputs_DefaultApplied(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "--with", "name=dave")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hello dave" {
		t.Errorf("output = %q, want %q", trimmed, "hello dave")
	}
}

func TestInputs_MissingRequired(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "greet")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing required input, got 0: %s", out)
	}
	if !strings.Contains(out, "required input") {
		t.Errorf("expected 'required input' in error, got: %s", out)
	}
}

func TestInputs_UnknownInput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "greet", "--with", "typo=val")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown input, got 0: %s", out)
	}
	if !strings.Contains(out, "unknown input") {
		t.Errorf("expected 'unknown input' in error, got: %s", out)
	}
}

func TestInputs_RunStepWithWith(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "caller")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hey world" {
		t.Errorf("output = %q, want %q", trimmed, "hey world")
	}
}

func TestInputs_List_ShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "INPUTS") {
		t.Error("expected INPUTS column in list output")
	}
	if !strings.Contains(out, "name, greeting?") {
		t.Errorf("expected 'name, greeting?' in list output, got:\n%s", out)
	}
}
