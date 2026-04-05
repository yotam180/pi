package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPipe_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"upper", "count-lines"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestPipe_Upper(t *testing.T) {
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPi(t, dir, "run", "upper")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "HELLO FROM PIPE") {
		t.Errorf("expected uppercased output, got:\n%s", out)
	}
}

func TestPipe_CountLines(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPiStdout(t, dir, "run", "count-lines")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "3" {
		t.Errorf("expected line count '3', got %q", trimmed)
	}
}
