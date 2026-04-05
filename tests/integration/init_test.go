package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CreatesProjectFiles(t *testing.T) {
	dir := t.TempDir()

	out, code := runPi(t, dir, "init", "--name", "test-project")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatalf("pi.yaml not created: %v", err)
	}
	if string(data) != "project: test-project\n" {
		t.Errorf("pi.yaml content = %q, want %q", string(data), "project: test-project\n")
	}

	info, err := os.Stat(filepath.Join(dir, ".pi"))
	if err != nil {
		t.Fatalf(".pi/ not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf(".pi should be a directory")
	}

	if !strings.Contains(out, "Initialized project 'test-project'") {
		t.Errorf("output should confirm init, got: %q", out)
	}
	if !strings.Contains(out, "Created pi.yaml") {
		t.Errorf("output should mention pi.yaml, got: %q", out)
	}
	if !strings.Contains(out, "Created .pi/") {
		t.Errorf("output should mention .pi/, got: %q", out)
	}
}

func TestInit_YesFlag(t *testing.T) {
	dir := t.TempDir()

	out, code := runPi(t, dir, "init", "--yes")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if _, err := os.Stat(filepath.Join(dir, "pi.yaml")); err != nil {
		t.Fatalf("pi.yaml should exist: %v", err)
	}
	if !strings.Contains(out, "Initialized project") {
		t.Errorf("output should confirm init, got: %q", out)
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: existing-app\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "init", "--name", "new-name")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Already initialized") {
		t.Errorf("output should say 'Already initialized', got: %q", out)
	}
	if !strings.Contains(out, "existing-app") {
		t.Errorf("output should mention existing project name, got: %q", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "existing-app") {
		t.Errorf("pi.yaml should not be overwritten, got: %q", string(data))
	}
}

func TestInit_PiDirExistsNoPiYaml(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "init", "--name", "fresh")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatalf("pi.yaml should be created: %v", err)
	}
	if !strings.Contains(string(data), "fresh") {
		t.Errorf("pi.yaml should contain 'fresh', got: %q", string(data))
	}
}

func TestInit_NextStepsShown(t *testing.T) {
	dir := t.TempDir()

	out, code := runPi(t, dir, "init", "--name", "my-proj")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Next steps:") {
		t.Errorf("output should show 'Next steps:', got: %q", out)
	}
	if !strings.Contains(out, "pi setup add") {
		t.Errorf("output should mention 'pi setup add', got: %q", out)
	}
	if !strings.Contains(out, "pi shell") {
		t.Errorf("output should mention 'pi shell', got: %q", out)
	}
	if !strings.Contains(out, "pi run") {
		t.Errorf("output should mention 'pi run', got: %q", out)
	}
}

func TestInit_NonInteractiveFallsBackToYes(t *testing.T) {
	dir := t.TempDir()

	// Running without --name and without --yes in a non-interactive context
	// (CI, exec.Command) should behave like --yes
	out, code := runPi(t, dir, "init")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Initialized project") {
		t.Errorf("output should confirm init, got: %q", out)
	}
}

func TestInit_CreatedProjectIsValid(t *testing.T) {
	dir := t.TempDir()

	_, code := runPi(t, dir, "init", "--name", "valid-project")
	if code != 0 {
		t.Fatalf("init failed with code %d", code)
	}

	// pi list should work on the newly created project
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("pi list failed with code %d\noutput: %s", code, out)
	}
}

func TestInit_AlreadyInitialized_NextSteps(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "init")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Next steps:") {
		t.Errorf("already-initialized should show 'Next steps:', got: %q", out)
	}
}
