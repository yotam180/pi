package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_BasicScaffold(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "build")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "build.yaml"))
	if err != nil {
		t.Fatalf("expected .pi/build.yaml to be created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "description:") {
		t.Errorf("expected description field, got:\n%s", content)
	}
	if !strings.Contains(content, "bash:") {
		t.Errorf("expected bash field, got:\n%s", content)
	}
}

func TestNew_OutputConfirmation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "deploy")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Created") {
		t.Errorf("expected 'Created' confirmation, got:\n%s", out)
	}
	if !strings.Contains(out, ".pi/deploy.yaml") {
		t.Errorf("expected file path in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Next steps:") {
		t.Errorf("expected 'Next steps:' guidance, got:\n%s", out)
	}
	if !strings.Contains(out, "pi run deploy") {
		t.Errorf("expected 'pi run deploy' in next steps, got:\n%s", out)
	}
}

func TestNew_BashFlag(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "compile", "--bash", "go build ./...")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "compile.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "go build ./...") {
		t.Errorf("expected bash command in file, got:\n%s", string(data))
	}
}

func TestNew_PythonFlag(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "transform", "--python", "transform.py")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "transform.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "python: transform.py") {
		t.Errorf("expected python field in file, got:\n%s", string(data))
	}
}

func TestNew_DescriptionFlag(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "test", "-d", "Run the test suite")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "test.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "Run the test suite") {
		t.Errorf("expected custom description in file, got:\n%s", string(data))
	}
}

func TestNew_NestedPath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "setup/install-deps")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	targetPath := filepath.Join(dir, ".pi", "setup", "install-deps.yaml")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected nested file to be created at %s: %v", targetPath, err)
	}

	if !strings.Contains(out, "setup/install-deps.yaml") {
		t.Errorf("expected nested path in output, got:\n%s", out)
	}
}

func TestNew_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)
	os.WriteFile(filepath.Join(dir, ".pi", "build.yaml"), []byte("bash: echo hi\n"), 0o644)

	out, code := runPi(t, dir, "new", "build")
	if code == 0 {
		t.Fatalf("expected non-zero exit for existing automation, got 0\noutput: %s", out)
	}

	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists' error, got:\n%s", out)
	}
}

func TestNew_NoProject(t *testing.T) {
	dir := t.TempDir()

	out, code := runPi(t, dir, "new", "build")
	if code == 0 {
		t.Fatalf("expected non-zero exit when no project, got 0\noutput: %s", out)
	}

	if !strings.Contains(out, "pi init") {
		t.Errorf("expected suggestion to run 'pi init', got:\n%s", out)
	}
}

func TestNew_StripYamlExtension(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "build.yaml")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if _, err := os.Stat(filepath.Join(dir, ".pi", "build.yaml")); err != nil {
		t.Fatalf("expected .pi/build.yaml (not .pi/build.yaml.yaml): %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".pi", "build.yaml.yaml")); err == nil {
		t.Error("should not create build.yaml.yaml — extension should be stripped")
	}
}

func TestNew_NoArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	_, code := runPi(t, dir, "new")
	if code == 0 {
		t.Fatal("expected non-zero exit when no argument provided")
	}
}

func TestNew_CreatedFileIsRunnable(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	_, code := runPi(t, dir, "new", "hello", "--bash", "echo works")
	if code != 0 {
		t.Fatalf("new failed with code %d", code)
	}

	out, code := runPi(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("run failed with code %d\noutput: %s", code, out)
	}

	if !strings.Contains(out, "works") {
		t.Errorf("expected 'works' in output, got:\n%s", out)
	}
}

func TestNew_CombinedFlags(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "lint", "--bash", "golangci-lint run", "-d", "Run the linter")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "lint.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "golangci-lint run") {
		t.Errorf("expected bash command, got:\n%s", content)
	}
	if !strings.Contains(content, "Run the linter") {
		t.Errorf("expected description, got:\n%s", content)
	}
}

func TestNew_DeeplyNestedPath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "infra/k8s/deploy-staging")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	targetPath := filepath.Join(dir, ".pi", "infra", "k8s", "deploy-staging.yaml")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected deeply nested file: %v", err)
	}

	if !strings.Contains(out, "infra/k8s/deploy-staging.yaml") {
		t.Errorf("expected nested path in output, got:\n%s", out)
	}
}

func TestNew_StripYmlExtension(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "new", "test.yml")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if _, err := os.Stat(filepath.Join(dir, ".pi", "test.yaml")); err != nil {
		t.Fatalf("expected .pi/test.yaml: %v", err)
	}
}
