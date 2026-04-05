package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupShellWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: testproj

shortcuts:
  hello: greet
  deploy:
    run: deploy/push
    anywhere: true
`), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`name: greet
description: Say hello
steps:
  - bash: echo hello
`), 0o644)
	os.MkdirAll(filepath.Join(piDir, "deploy"), 0o755)
	os.WriteFile(filepath.Join(piDir, "deploy", "push.yaml"), []byte(`name: deploy/push
description: Push deploy
steps:
  - bash: echo deploying
`), 0o644)

	return root
}

func TestShellCmd_Install(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := setupShellWorkspace(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runShellInstall(&stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "Installed 2 shortcut(s)") {
		t.Errorf("expected install summary, got:\n%s", out)
	}
	if !strings.Contains(out, "hello → greet") {
		t.Errorf("expected hello shortcut, got:\n%s", out)
	}
	if !strings.Contains(out, "deploy → deploy/push (anywhere)") {
		t.Errorf("expected deploy shortcut, got:\n%s", out)
	}

	// Verify file exists
	shellPath := filepath.Join(tmpHome, ".pi", "shell", "testproj.sh")
	data, err := os.ReadFile(shellPath)
	if err != nil {
		t.Fatalf("reading shell file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "function hello()") {
		t.Error("shell file missing hello function")
	}
	if !strings.Contains(content, "function deploy()") {
		t.Error("shell file missing deploy function")
	}
	if !strings.Contains(content, "--repo") {
		t.Error("deploy function should use --repo")
	}
}

func TestShellCmd_Install_ShadowWarning(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: testproj

shortcuts:
  test: my-tests
  vpup: docker/up
  echo: my-echo
`), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "my-tests.yaml"), []byte("bash: echo test\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "my-echo.yaml"), []byte("bash: echo hello\n"), 0o644)
	os.MkdirAll(filepath.Join(piDir, "docker"), 0o755)
	os.WriteFile(filepath.Join(piDir, "docker", "up.yaml"), []byte("bash: echo up\n"), 0o644)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runShellInstall(&stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, `shortcut "echo"`) {
		t.Errorf("expected shadow warning for 'echo', got stderr:\n%s", stderrStr)
	}
	if !strings.Contains(stderrStr, `shortcut "test"`) {
		t.Errorf("expected shadow warning for 'test', got stderr:\n%s", stderrStr)
	}
	if strings.Contains(stderrStr, "vpup") {
		t.Errorf("should not warn about 'vpup', got stderr:\n%s", stderrStr)
	}

	// Shortcuts should still be installed despite warnings
	stdoutStr := stdout.String()
	if !strings.Contains(stdoutStr, "Installed 3 shortcut(s)") {
		t.Errorf("shortcuts should still install despite warnings, got:\n%s", stdoutStr)
	}
}

func TestShellCmd_Install_NoWarningForSafeNames(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := setupShellWorkspace(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	err := runShellInstall(&stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stderr.Len() > 0 {
		t.Errorf("expected no warnings for safe shortcut names, got stderr:\n%s", stderr.String())
	}
}

func TestShellCmd_Uninstall(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	root := setupShellWorkspace(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	runShellInstall(&stdout, &stderr)

	stdout.Reset()
	err := runShellUninstall(&stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Removed") {
		t.Errorf("expected removal message, got:\n%s", stdout.String())
	}

	shellPath := filepath.Join(tmpHome, ".pi", "shell", "testproj.sh")
	if _, err := os.Stat(shellPath); !os.IsNotExist(err) {
		t.Error("shell file should be removed")
	}
}

func TestShellCmd_List(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Empty
	var stdout bytes.Buffer
	runShellList(&stdout)
	if !strings.Contains(stdout.String(), "No shell shortcuts") {
		t.Errorf("expected empty message, got:\n%s", stdout.String())
	}

	// Create some files
	shellDir := filepath.Join(tmpHome, ".pi", "shell")
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, "proj-a.sh"), []byte("# a"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "proj-b.sh"), []byte("# b"), 0o644)

	stdout.Reset()
	runShellList(&stdout)
	out := stdout.String()
	if !strings.Contains(out, "proj-a.sh") {
		t.Errorf("expected proj-a, got:\n%s", out)
	}
	if !strings.Contains(out, "proj-b.sh") {
		t.Errorf("expected proj-b, got:\n%s", out)
	}
}
