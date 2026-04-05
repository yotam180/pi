package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParentShell_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"activate-venv", "cd-tmp", "normal"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestParentShell_WritesToEvalFile(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	evalFile := filepath.Join(t.TempDir(), "eval.sh")
	env := []string{"PI_PARENT_EVAL_FILE=" + evalFile}

	out, code := runPiWithEnv(t, dir, env, "run", "cd-tmp")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "cd /tmp" {
		t.Errorf("eval file content = %q, want %q", content, "cd /tmp")
	}
}

func TestParentShell_MixedSteps(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	evalFile := filepath.Join(t.TempDir(), "eval.sh")
	env := []string{"PI_PARENT_EVAL_FILE=" + evalFile}

	out, code := runPiWithEnv(t, dir, env, "run", "activate-venv")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "preparing environment") {
		t.Errorf("expected normal step output, got:\n%s", out)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "source venv/bin/activate" {
		t.Errorf("eval file content = %q, want %q", content, "source venv/bin/activate")
	}
}

func TestParentShell_NoEvalFile_Error(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	out, code := runPi(t, dir, "run", "cd-tmp")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0: %s", out)
	}
	if !strings.Contains(out, "PI_PARENT_EVAL_FILE") {
		t.Errorf("expected error about PI_PARENT_EVAL_FILE, got:\n%s", out)
	}
}

func TestParentShell_NormalStepUnaffected(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	out, code := runPi(t, dir, "run", "normal")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "normal step") {
		t.Errorf("expected normal output, got:\n%s", out)
	}
}

func TestParentShell_InfoShowsAnnotation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	out, code := runPi(t, dir, "info", "activate-venv")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "parent_shell") {
		t.Errorf("expected parent_shell annotation in info output, got:\n%s", out)
	}
}

func TestParentShell_ShellCodegenHasEvalWrapper(t *testing.T) {
	dir := filepath.Join(examplesDir(), "parent-shell")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	env := []string{"HOME=" + tmpHome}
	out, code := runPiWithEnv(t, dir, env, "shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	shellFile := filepath.Join(tmpHome, ".pi", "shell", "parent-shell-test.sh")
	data, err := os.ReadFile(shellFile)
	if err != nil {
		t.Fatalf("reading shell file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "PI_PARENT_EVAL_FILE") {
		t.Errorf("shell file should contain PI_PARENT_EVAL_FILE, got:\n%s", content)
	}
	if !strings.Contains(content, "pi-setup-") {
		t.Errorf("shell file should contain pi-setup helper, got:\n%s", content)
	}
	if !strings.Contains(content, `source "$_pi_eval_file"`) {
		t.Errorf("shell file should contain source of eval file, got:\n%s", content)
	}
}
