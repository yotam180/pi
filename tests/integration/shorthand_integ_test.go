package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestShorthand_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"greet", "build", "delegate", "greet-input", "silent-build"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got: %s", name, out)
		}
	}
}

func TestShorthand_RunBash(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	stdout, _, code := runPiSplit(t, dir, "run", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Hello from shorthand!") {
		t.Errorf("expected greeting in output, got: %s", stdout)
	}
}

func TestShorthand_RunWithEnv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	stdout, _, code := runPiSplit(t, dir, "run", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "GOOS=linux") {
		t.Errorf("expected GOOS=linux in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "GOARCH=amd64") {
		t.Errorf("expected GOARCH=amd64 in output, got: %s", stdout)
	}
}

func TestShorthand_RunStep(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	stdout, _, code := runPiSplit(t, dir, "run", "delegate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Hello from shorthand!") {
		t.Errorf("expected greeting from delegated automation, got: %s", stdout)
	}
}

func TestShorthand_RunWithInput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	stdout, _, code := runPiSplit(t, dir, "run", "greet-input", "Alice")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Hello, Alice!") {
		t.Errorf("expected 'Hello, Alice!' in output, got: %s", stdout)
	}
}

func TestShorthand_InfoShowsCorrectly(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Greet the user") {
		t.Errorf("expected description in info output, got: %s", out)
	}
	if !strings.Contains(out, "Steps:") {
		t.Errorf("expected 'Steps:' in info output, got: %s", out)
	}
}

func TestShorthand_InfoWithModifiers(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	out, code := runPi(t, dir, "info", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Cross-compile for Linux") {
		t.Errorf("expected description in info output, got: %s", out)
	}
	if !strings.Contains(out, "Env:") || !strings.Contains(out, "GOOS") {
		t.Errorf("expected automation-level Env: in info output, got: %s", out)
	}
}

func TestShorthand_Validate(t *testing.T) {
	dir := filepath.Join(examplesDir(), "shorthand")
	out, code := runPi(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Validated") {
		t.Errorf("expected 'Validated' in output, got: %s", out)
	}
}
