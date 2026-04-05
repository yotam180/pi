package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRequiresValidation_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"needs-bash", "needs-impossible", "needs-impossible-version", "needs-python", "no-requires"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestRequiresValidation_SatisfiedCommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-bash")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "bash is available") {
		t.Errorf("expected 'bash is available' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_SatisfiedRuntime(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-python")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "python is available") {
		t.Errorf("expected 'python is available' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_MissingCommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Missing requirements:") {
		t.Errorf("expected 'Missing requirements:' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "pi-nonexistent-tool-xyz") {
		t.Errorf("expected tool name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in output, got:\n%s", out)
	}
	if strings.Contains(out, "this should never run") {
		t.Errorf("automation steps should not execute when requirements fail, got:\n%s", out)
	}
}

func TestRequiresValidation_ImpossibleVersion(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible-version")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Missing requirements:") {
		t.Errorf("expected 'Missing requirements:' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "python >= 99.0") {
		t.Errorf("expected 'python >= 99.0' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "99.0") {
		t.Errorf("expected version requirement in output, got:\n%s", out)
	}
	if strings.Contains(out, "this should never run") {
		t.Errorf("automation steps should not execute when requirements fail, got:\n%s", out)
	}
}

func TestRequiresValidation_NoRequirements(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "no-requires")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "no requirements needed") {
		t.Errorf("expected 'no requirements needed' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_ErrorShowsInstallHint(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible-version")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "install:") {
		t.Errorf("expected install hint in output, got:\n%s", out)
	}
}
