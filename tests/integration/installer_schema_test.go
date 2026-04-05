package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerSchema_ListShowsInstallerAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"check-ready", "install-marker", "install-conditional", "install-no-version", "steps-automation"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestInstallerSchema_AlreadyInstalled(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' icon in output, got:\n%s", out)
	}
}

func TestInstallerSchema_FreshInstall(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	marker := filepath.Join(t.TempDir(), "test-marker")
	out, code := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "installing...") {
		t.Errorf("expected 'installing...' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "installed") {
		t.Errorf("expected 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected version '1.0.0' in output, got:\n%s", out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("expected marker file to exist at %s", marker)
	}
}

func TestInstallerSchema_FreshInstallThenAlreadyInstalled(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	marker := filepath.Join(t.TempDir(), "test-marker")

	out1, code1 := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code1 != 0 {
		t.Fatalf("first run: expected exit 0, got %d: %s", code1, out1)
	}
	if !strings.Contains(out1, "installing...") {
		t.Errorf("first run: expected 'installing...', got:\n%s", out1)
	}

	out2, code2 := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code2 != 0 {
		t.Fatalf("second run: expected exit 0, got %d: %s", code2, out2)
	}
	if !strings.Contains(out2, "already installed") {
		t.Errorf("second run: expected 'already installed', got:\n%s", out2)
	}
}

func TestInstallerSchema_NoVersion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "install-no-version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed', got:\n%s", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "install-no-version") && strings.Contains(line, "(") {
			t.Errorf("expected no version parenthetical, got:\n%s", line)
		}
	}
}

func TestInstallerSchema_InfoShowsInstallerType(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Type:") {
		t.Errorf("expected 'Type:' in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "installer") {
		t.Errorf("expected 'installer' in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "Install lifecycle") {
		t.Errorf("expected 'Install lifecycle' in info output, got:\n%s", out)
	}
}

func TestInstallerSchema_InfoShowsStepsForRegularAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "steps-automation")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Steps:") {
		t.Errorf("expected 'Steps:' in info output, got:\n%s", out)
	}
	if strings.Contains(out, "Type:         installer") {
		t.Errorf("unexpected 'Type: installer' in info output for steps-based automation, got:\n%s", out)
	}
}

func TestInstallerSchema_ConditionalRunSteps(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "install-conditional")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "installed") {
		t.Errorf("expected 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2.0.0") {
		t.Errorf("expected version '2.0.0' in output, got:\n%s", out)
	}
}

func TestInstallerSchema_SilentFlag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "--silent", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if strings.Contains(out, "already installed") {
		t.Errorf("expected no status output with --silent, got:\n%s", out)
	}
	if strings.Contains(out, "✓") {
		t.Errorf("expected no status icon with --silent, got:\n%s", out)
	}
}

func TestInstallerSchema_RegularAutomationUnaffectedBySilent(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "--silent", "steps-automation")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "I am a regular automation") {
		t.Errorf("expected regular automation output even with --silent, got:\n%s", out)
	}
}

func TestInstallerSchema_BuiltinInstallerAlreadyInstalled(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "pi:install-tsx")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' icon for already-installed tsx, got:\n%s", out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' for tsx, got:\n%s", out)
	}
}
