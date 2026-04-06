package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDryRun_ShowsStepsWithoutExecuting(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	stdout, stderr, code := runPiSplit(t, dir, "run", "--dry-run", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if strings.Contains(stdout, "Hello") {
		t.Errorf("dry-run should not produce command output on stdout, got: %q", stdout)
	}

	if !strings.Contains(stderr, "bash") {
		t.Errorf("expected 'bash' step type in dry-run stderr, got: %q", stderr)
	}
}

func TestDryRun_RunStepRecurses(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "caller.yaml"), []byte(`description: Calls target
steps:
  - run: target
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "target.yaml"), []byte(`description: The target
steps:
  - bash: echo "target ran"
`), 0o644)

	stdout, stderr, code := runPiSplit(t, root, "run", "--dry-run", "caller")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if stdout != "" {
		t.Errorf("dry-run should produce no stdout, got: %q", stdout)
	}

	if !strings.Contains(stderr, "run") {
		t.Errorf("expected 'run' step in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "target ran") {
		t.Errorf("expected target step content in stderr (recursion), got: %q", stderr)
	}
}

func TestDryRun_ConditionalStepsShowSkipReason(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "cond.yaml"), []byte(`description: Conditional steps
steps:
  - bash: echo "always"
  - bash: echo "windows only"
    if: os.windows
  - bash: echo "after conditional"
`), 0o644)

	stdout, stderr, code := runPiSplit(t, root, "run", "--dry-run", "cond")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if !strings.Contains(stderr, "always") {
		t.Errorf("expected unconditional step in output, got: %q", stderr)
	}
	if !strings.Contains(stderr, "skipped") {
		t.Errorf("expected 'skipped' for false condition, got: %q", stderr)
	}
	if !strings.Contains(stderr, "after conditional") {
		t.Errorf("expected step after conditional in output, got: %q", stderr)
	}
}

func TestDryRun_FirstBlockShowsMatchInfo(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "fb.yaml"), []byte(`description: First block test
steps:
  - first:
      - bash: echo "windows"
        if: os.windows
      - bash: echo "fallback"
`), 0o644)

	stdout, stderr, code := runPiSplit(t, root, "run", "--dry-run", "fb")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if !strings.Contains(stderr, "first") {
		t.Errorf("expected 'first' block indicator, got: %q", stderr)
	}
	if !strings.Contains(stderr, "match") {
		t.Errorf("expected 'match' indicator for fallback, got: %q", stderr)
	}
}

func TestDryRun_InstallerShowsLifecycle(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "install-foo.yaml"), []byte(`description: Install foo
install:
  test: command -v foo
  run: echo "installing foo"
  version: foo --version
`), 0o644)

	stdout, stderr, code := runPiSplit(t, root, "run", "--dry-run", "install-foo")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}

	if !strings.Contains(stderr, "install") {
		t.Errorf("expected 'install' in output, got: %q", stderr)
	}
	if !strings.Contains(stderr, "test") {
		t.Errorf("expected 'test' phase in output, got: %q", stderr)
	}
	if !strings.Contains(stderr, "run") {
		t.Errorf("expected 'run' phase in output, got: %q", stderr)
	}
	if !strings.Contains(stderr, "version") {
		t.Errorf("expected 'version' in output, got: %q", stderr)
	}
}
