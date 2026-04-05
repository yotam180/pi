package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStepDir_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"run-in-subdir", "mixed-dirs", "dir-with-env", "bad-dir"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list, got: %s", name, out)
		}
	}
}

func TestStepDir_RunInSubdir(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	stdout, _, code := runPiSplit(t, dir, "run", "run-in-subdir")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "subdir") {
		t.Errorf("expected 'subdir' in pwd output, got: %s", stdout)
	}
}

func TestStepDir_MixedDirs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	stdout, _, code := runPiSplit(t, dir, "run", "mixed-dirs")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "sub=") {
		t.Errorf("expected 'sub=' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "subdir") {
		t.Errorf("expected subdir path in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "root=") {
		t.Errorf("expected 'root=' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "back=") {
		t.Errorf("expected 'back=' in output, got: %s", stdout)
	}
}

func TestStepDir_DirWithEnv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	stdout, _, code := runPiSplit(t, dir, "run", "dir-with-env")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "working-in") {
		t.Errorf("expected 'working-in' from env var, got: %s", stdout)
	}
	if !strings.Contains(stdout, "subdir") {
		t.Errorf("expected subdir path, got: %s", stdout)
	}
}

func TestStepDir_BadDirFails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	out, code := runPi(t, dir, "run", "bad-dir")
	if code == 0 {
		t.Fatal("expected non-zero exit for bad dir")
	}
	if !strings.Contains(out, "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %s", out)
	}
}

func TestStepDir_Info(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-dir")
	out, code := runPi(t, dir, "info", "run-in-subdir")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "dir: subdir") {
		t.Errorf("expected 'dir: subdir' annotation in info output, got: %s", out)
	}
}
