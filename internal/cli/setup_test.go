package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupSetupWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: testproj

shortcuts:
  hello: greet

setup:
  - run: greet
`), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`name: greet
description: Say hello
steps:
  - bash: echo "Hello from setup"
`), 0o644)

	return root
}

func TestSetup_RunsEntries(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := setupSetupWorkspace(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "==> setup[0]: greet") {
		t.Errorf("expected setup header, got:\n%s", out)
	}
	if !strings.Contains(out, "Hello from setup") {
		t.Errorf("expected setup output, got:\n%s", out)
	}
	if !strings.Contains(out, "--no-shell") {
		t.Errorf("expected no-shell skip message, got:\n%s", out)
	}
}

func TestSetup_SkipsShellInCI(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("CI", "true")

	root := setupSetupWorkspace(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "CI environment") {
		t.Errorf("expected CI skip message, got:\n%s", out)
	}
}

func TestSetup_InstallsShortcuts(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := setupSetupWorkspace(t)
	t.Chdir(root)

	// Create .zshrc so source line is injected
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Installing shell shortcuts") {
		t.Errorf("expected shell install step, got:\n%s", out)
	}
	if !strings.Contains(out, "Setup complete") {
		t.Errorf("expected completion message, got:\n%s", out)
	}

	shellPath := filepath.Join(tmpHome, ".pi", "shell", "testproj.sh")
	if _, err := os.Stat(shellPath); err != nil {
		t.Errorf("expected shell file at %s: %v", shellPath, err)
	}
}

func TestSetup_SkipsEntryWithFalseCondition(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: testproj
setup:
  - run: greet
    if: os.windows
  - run: greet2
`), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`name: greet
description: Say hello
steps:
  - bash: echo "SKIPPED_ENTRY_OUTPUT"
`), 0o644)
	os.WriteFile(filepath.Join(piDir, "greet2.yaml"), []byte(`name: greet2
description: Second greeting
steps:
  - bash: echo "EXECUTED_ENTRY_OUTPUT"
`), 0o644)

	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "[skipped]") {
		t.Errorf("expected [skipped] in output, got:\n%s", out)
	}
	if !strings.Contains(out, "os.windows") {
		t.Errorf("expected condition in skip message, got:\n%s", out)
	}
	if strings.Contains(out, "SKIPPED_ENTRY_OUTPUT") {
		t.Errorf("greet should have been skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "EXECUTED_ENTRY_OUTPUT") {
		t.Errorf("greet2 should have run, got:\n%s", out)
	}
}

func TestSetup_RunsEntryWithTrueCondition(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: testproj
setup:
  - run: greet
    if: command.bash
`), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`name: greet
description: Say hello
steps:
  - bash: echo "Hello from setup"
`), 0o644)

	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if strings.Contains(out, "[skipped]") {
		t.Errorf("entry should not be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "Hello from setup") {
		t.Errorf("expected setup output, got:\n%s", out)
	}
}

func TestSetup_Empty(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: empty\n"), 0o644)
	os.MkdirAll(filepath.Join(root, ".pi"), 0o755)

	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runSetup(&stdout, &stderr, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Nothing to set up") {
		t.Errorf("expected nothing message, got:\n%s", stdout.String())
	}
}
