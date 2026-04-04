package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runPiWithHome(t *testing.T, dir, home string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+home)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v\n%s", args, err, string(out))
		}
	}
	return string(out), exitCode
}

func TestShell_Install(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)

	out, code := runPiWithHome(t, dir, tmpHome, "shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "Installed") {
		t.Errorf("expected install message, got:\n%s", out)
	}
	if !strings.Contains(out, "shortcut(s)") {
		t.Errorf("expected shortcut count, got:\n%s", out)
	}

	shellFile := filepath.Join(tmpHome, ".pi", "shell", "docker-project.sh")
	data, err := os.ReadFile(shellFile)
	if err != nil {
		t.Fatalf("reading shell file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "function dup()") {
		t.Error("missing dup function")
	}
	if !strings.Contains(content, "function ddown()") {
		t.Error("missing ddown function")
	}
	if !strings.Contains(content, "docker/up") {
		t.Error("missing docker/up in function body")
	}

	zshData, _ := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	if !strings.Contains(string(zshData), "# Added by PI") {
		t.Error("source line missing from .zshrc")
	}
}

func TestShell_Idempotent(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	runPiWithHome(t, dir, tmpHome, "shell")
	runPiWithHome(t, dir, tmpHome, "shell")

	zshData, _ := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	count := strings.Count(string(zshData), "# Added by PI")
	if count != 1 {
		t.Errorf("expected source line exactly once, found %d times", count)
	}
}

func TestShell_Uninstall(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	runPiWithHome(t, dir, tmpHome, "shell")

	out, code := runPiWithHome(t, dir, tmpHome, "shell", "uninstall")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Removed") {
		t.Errorf("expected removed message, got:\n%s", out)
	}

	shellFile := filepath.Join(tmpHome, ".pi", "shell", "docker-project.sh")
	if _, err := os.Stat(shellFile); !os.IsNotExist(err) {
		t.Error("shell file should be removed after uninstall")
	}

	zshData, _ := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
	if strings.Contains(string(zshData), "# Added by PI") {
		t.Error("source line should be removed after uninstall")
	}
}

func TestShell_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	// Empty list
	out, code := runPiWithHome(t, dir, tmpHome, "shell", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "No shell shortcuts") {
		t.Errorf("expected empty message, got:\n%s", out)
	}

	// Install and list again
	runPiWithHome(t, dir, tmpHome, "shell")

	out, code = runPiWithHome(t, dir, tmpHome, "shell", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "docker-project.sh") {
		t.Errorf("expected docker-project in list, got:\n%s", out)
	}
}

func TestShell_RunWithRepoFlag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpDir := t.TempDir()

	out, code := runPi(t, tmpDir, "run", "--repo", dir, "docker/up")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "All containers started") {
		t.Errorf("expected docker up output with --repo flag, got:\n%s", out)
	}
}

func TestSetup_Integration(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	out, code := runPiWithHome(t, dir, tmpHome, "setup")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "Installing shell shortcuts") {
		t.Errorf("expected shell install step, got:\n%s", out)
	}
	if !strings.Contains(out, "Setup complete") {
		t.Errorf("expected setup complete message, got:\n%s", out)
	}

	shellFile := filepath.Join(tmpHome, ".pi", "shell", "docker-project.sh")
	if _, err := os.Stat(shellFile); err != nil {
		t.Errorf("expected shell file created by setup: %v", err)
	}
}

func TestSetup_WithNoShell(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()

	out, code := runPiWithHome(t, dir, tmpHome, "setup", "--no-shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "--no-shell") {
		t.Errorf("expected no-shell skip message, got:\n%s", out)
	}

	shellFile := filepath.Join(tmpHome, ".pi", "shell", "docker-project.sh")
	if _, err := os.Stat(shellFile); !os.IsNotExist(err) {
		t.Error("shell file should not be created with --no-shell")
	}
}

func TestSetup_ConditionalEntries(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	tmpHome := t.TempDir()

	out, code := runPiWithHome(t, dir, tmpHome, "setup", "--no-shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	// setup-always (no if:) should always run
	if !strings.Contains(out, "setup-always executed") {
		t.Errorf("setup-always should run, got:\n%s", out)
	}

	// setup-never (if: os.windows and os.linux) should be skipped
	if strings.Contains(out, "setup-never executed") {
		t.Errorf("setup-never should be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "[skipped]") {
		t.Errorf("expected [skipped] marker, got:\n%s", out)
	}

	// setup-platform (if: os.macos or os.linux) runs on CI (ubuntu) and dev (macOS)
	if !strings.Contains(out, "setup-platform executed") {
		t.Errorf("setup-platform should run on macOS/Linux, got:\n%s", out)
	}
}

func TestSetup_ConditionalSkipShowsCondition(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	tmpHome := t.TempDir()

	out, code := runPiWithHome(t, dir, tmpHome, "setup", "--no-shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "os.windows and os.linux") {
		t.Errorf("expected condition expression in skip message, got:\n%s", out)
	}
}
