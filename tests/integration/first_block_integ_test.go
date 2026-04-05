package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFirstBlock_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"pick-platform", "no-match", "with-pipe", "mixed"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestFirstBlock_PickPlatform(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPiStdout(t, dir, "run", "pick-platform")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	switch runtime.GOOS {
	case "darwin":
		if trimmed != "Platform is macOS" {
			t.Errorf("output = %q, want %q", trimmed, "Platform is macOS")
		}
	case "linux":
		if trimmed != "Platform is Linux" {
			t.Errorf("output = %q, want %q", trimmed, "Platform is Linux")
		}
	default:
		if trimmed != "Platform is unknown" {
			t.Errorf("output = %q, want %q", trimmed, "Platform is unknown")
		}
	}
}

func TestFirstBlock_NoMatch(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPiStdout(t, dir, "run", "no-match")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "continued" {
		t.Errorf("output = %q, want %q", trimmed, "continued")
	}
}

func TestFirstBlock_WithPipe(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPiStdout(t, dir, "run", "with-pipe")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hello_from_first" {
		t.Errorf("output = %q, want %q", trimmed, "hello_from_first")
	}
}

func TestFirstBlock_Mixed(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPiStdout(t, dir, "run", "mixed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "before" {
		t.Errorf("line[0] = %q, want %q", lines[0], "before")
	}
	if lines[1] != "middle" {
		t.Errorf("line[1] = %q, want %q", lines[1], "middle")
	}
	if lines[2] != "after" {
		t.Errorf("line[2] = %q, want %q", lines[2], "after")
	}
}

func TestFirstBlock_Info(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPi(t, dir, "info", "pick-platform")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "first") {
		t.Errorf("expected 'first' in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "Step details") {
		t.Errorf("expected 'Step details' section in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "if: os.macos") {
		t.Errorf("expected sub-step condition in info output, got:\n%s", out)
	}
}

func TestFirstBlock_Validate(t *testing.T) {
	dir := filepath.Join(examplesDir(), "first-block")
	out, code := runPi(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Validated") {
		t.Errorf("expected 'Validated' in output, got:\n%s", out)
	}
}

func TestFirstBlock_InstallerWithFirst(t *testing.T) {
	marker := "/tmp/pi-first-block-test-marker"
	os.Remove(marker)
	defer os.Remove(marker)

	dir := filepath.Join(examplesDir(), "first-block")
	_, code := runPi(t, dir, "run", "installer-first")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("expected marker file %s to exist after install", marker)
	}
}
