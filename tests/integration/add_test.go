package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupAddProject(t *testing.T, piYaml string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(piYaml), 0o644); err != nil {
		t.Fatal(err)
	}
	piDir := filepath.Join(dir, ".pi")
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readPiYaml(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestAdd_FileSource(t *testing.T) {
	dir := setupAddProject(t, "project: test-add\n")

	out, code := runPi(t, dir, "add", "file:~/my-automations")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}
	if !strings.Contains(out, "added") {
		t.Errorf("expected 'added' in output, got: %s", out)
	}

	content := readPiYaml(t, dir)
	if !strings.Contains(content, "file:~/my-automations") {
		t.Errorf("pi.yaml should contain the added source, got:\n%s", content)
	}
}

func TestAdd_FileSourceWithAlias(t *testing.T) {
	dir := setupAddProject(t, "project: test-add\n")

	out, code := runPi(t, dir, "add", "file:~/my-automations", "--as", "mytools")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	content := readPiYaml(t, dir)
	if !strings.Contains(content, "file:~/my-automations") {
		t.Errorf("pi.yaml should contain source, got:\n%s", content)
	}
	if !strings.Contains(content, "mytools") {
		t.Errorf("pi.yaml should contain alias, got:\n%s", content)
	}
}

func TestAdd_Idempotent(t *testing.T) {
	dir := setupAddProject(t, `project: test-add

packages:
  - file:~/my-automations
`)

	out, code := runPi(t, dir, "add", "file:~/my-automations")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}
	if !strings.Contains(out, "already in") {
		t.Errorf("expected 'already in' message, got: %s", out)
	}

	content := readPiYaml(t, dir)
	count := strings.Count(content, "file:~/my-automations")
	if count != 1 {
		t.Errorf("source should appear exactly once, appeared %d times", count)
	}
}

func TestAdd_NoVersionError(t *testing.T) {
	dir := setupAddProject(t, "project: test-add\n")

	out, code := runPi(t, dir, "add", "yotam180/pi-common")
	if code == 0 {
		t.Fatal("expected non-zero exit for missing version")
	}
	if !strings.Contains(out, "version required") {
		t.Errorf("expected 'version required' error, got: %s", out)
	}
}

func TestAdd_NoArgs(t *testing.T) {
	dir := setupAddProject(t, "project: test-add\n")

	out, code := runPi(t, dir, "add")
	if code == 0 {
		t.Fatal("expected non-zero exit for no args")
	}
	if !strings.Contains(out, "accepts 1 arg") {
		t.Errorf("expected usage error, got: %s", out)
	}
}

func TestAdd_CreatesPackagesBlock(t *testing.T) {
	dir := setupAddProject(t, `project: test-add

shortcuts:
  up: docker/up
`)

	out, code := runPi(t, dir, "add", "file:~/shared")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	content := readPiYaml(t, dir)
	if !strings.Contains(content, "packages:") {
		t.Error("pi.yaml should contain packages: block")
	}
	if !strings.Contains(content, "docker/up") {
		t.Error("existing shortcuts should be preserved")
	}
}

func TestAdd_AppendsToExisting(t *testing.T) {
	dir := setupAddProject(t, `project: test-add

packages:
  - file:~/first
`)

	out, code := runPi(t, dir, "add", "file:~/second")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	content := readPiYaml(t, dir)
	if !strings.Contains(content, "file:~/first") {
		t.Error("first package should be preserved")
	}
	if !strings.Contains(content, "file:~/second") {
		t.Error("second package should be added")
	}
}

func TestAdd_InvalidSourceError(t *testing.T) {
	dir := setupAddProject(t, "project: test-add\n")

	out, code := runPi(t, dir, "add", "just-a-name")
	if code == 0 {
		t.Fatal("expected non-zero exit for invalid source")
	}
	if !strings.Contains(out, "invalid package source") {
		t.Errorf("expected 'invalid package source' error, got: %s", out)
	}
}
