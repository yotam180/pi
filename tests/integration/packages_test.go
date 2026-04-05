package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func packagesDir() string {
	return filepath.Join(examplesDir(), "packages")
}

func TestPackages_List(t *testing.T) {
	out, code := runPi(t, packagesDir(), "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}

	// Should show local automation
	if !strings.Contains(out, "local/hello") {
		t.Errorf("expected local/hello in list, got:\n%s", out)
	}

	// Should show package automations
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected docker/up from package in list, got:\n%s", out)
	}
	if !strings.Contains(out, "docker/down") {
		t.Errorf("expected docker/down from package in list, got:\n%s", out)
	}
	if !strings.Contains(out, "utils/greet") {
		t.Errorf("expected utils/greet from package in list, got:\n%s", out)
	}
}

func TestPackages_RunLocal(t *testing.T) {
	out, code := runPi(t, packagesDir(), "run", "local/hello")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "hello from local") {
		t.Errorf("expected 'hello from local' output, got:\n%s", out)
	}
}

func TestPackages_RunPackageAutomation(t *testing.T) {
	out, code := runPi(t, packagesDir(), "run", "docker/up")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "docker up from package") {
		t.Errorf("expected 'docker up from package' output, got:\n%s", out)
	}
}

func TestPackages_RunViaAlias(t *testing.T) {
	out, code := runPi(t, packagesDir(), "run", "mytools/docker/up")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "docker up from package") {
		t.Errorf("expected 'docker up from package' output, got:\n%s", out)
	}
}

func TestPackages_RunUtilsGreet(t *testing.T) {
	out, code := runPi(t, packagesDir(), "run", "utils/greet")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "greetings from the package") {
		t.Errorf("expected 'greetings from the package' output, got:\n%s", out)
	}
}

func TestPackages_Info(t *testing.T) {
	out, code := runPi(t, packagesDir(), "info", "docker/up")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected 'docker/up' in info, got:\n%s", out)
	}
	if !strings.Contains(out, "Start containers (from package)") {
		t.Errorf("expected description in info, got:\n%s", out)
	}
}

func TestPackages_Validate(t *testing.T) {
	out, code := runPi(t, packagesDir(), "validate")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "Validated") {
		t.Errorf("expected validation success, got:\n%s", out)
	}
}

func TestPackages_Setup_FetchesPackages(t *testing.T) {
	stdout, stderr, code := runPiSplit(t, packagesDir(), "setup", "--no-shell")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	combined := stdout + stderr

	// Should show package fetch status
	if !strings.Contains(combined, "found") {
		t.Errorf("expected 'found' in setup output for file: package, got:\n%s", combined)
	}

	// Should run setup automation
	if !strings.Contains(combined, "hello from local") {
		t.Errorf("expected 'hello from local' from setup, got:\n%s", combined)
	}
}

func TestPackages_LocalShadowsPackage(t *testing.T) {
	// Create a temp workspace where local shadows a package automation
	dir := t.TempDir()

	// Create pi.yaml with package
	writeIntegFile(t, filepath.Join(dir, "pi.yaml"), `
project: shadow-test
packages:
  - source: file:./pkg
    as: mypkg
`)

	// Create local .pi/shared.yaml
	writeIntegFile(t, filepath.Join(dir, ".pi", "shared.yaml"), `
description: Local shared
bash: echo "local wins"
`)

	// Create package with same name
	writeIntegFile(t, filepath.Join(dir, "pkg", ".pi", "shared.yaml"), `
description: Package shared
bash: echo "package loses"
`)

	out, code := runPi(t, dir, "run", "shared")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "local wins") {
		t.Errorf("expected local to shadow package, got:\n%s", out)
	}
}

func writeIntegFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}
