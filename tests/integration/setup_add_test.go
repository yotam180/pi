package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupAdd_BareString(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:install-uv")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Added to setup") {
		t.Errorf("output should say 'Added to setup', got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if !strings.Contains(string(data), "pi:install-uv") {
		t.Errorf("pi.yaml should contain pi:install-uv, got: %q", string(data))
	}
}

func TestSetupAdd_ShortFormResolution(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "python", "--version", "3.13")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Resolved 'python'") {
		t.Errorf("output should show resolution, got: %q", out)
	}
	if !strings.Contains(out, "pi:install-python") {
		t.Errorf("output should mention resolved name, got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "pi:install-python") {
		t.Errorf("pi.yaml should contain pi:install-python, got: %q", content)
	}
	if !strings.Contains(content, `"3.13"`) {
		t.Errorf("pi.yaml should contain version, got: %q", content)
	}
}

func TestSetupAdd_WithIfFlag(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "homebrew", "--if", "os.macos")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "if: os.macos") {
		t.Errorf("pi.yaml should contain if: os.macos, got: %q", content)
	}
}

func TestSetupAdd_IdempotentDuplicate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\nsetup:\n  - pi:install-uv\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:install-uv")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Already in pi.yaml") {
		t.Errorf("output should say 'Already in pi.yaml', got: %q", out)
	}
}

func TestSetupAdd_KeyValueArgs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:cursor/install-extensions", "file=.pi/cursor/extensions.txt")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "pi:cursor/install-extensions") {
		t.Errorf("should contain automation name, got: %q", content)
	}
	if !strings.Contains(content, "file") {
		t.Errorf("should contain 'file' key, got: %q", content)
	}
}

func TestSetupAdd_NoPiYaml_InitsProject(t *testing.T) {
	dir := t.TempDir()

	out, code := runPi(t, dir, "setup", "add", "uv", "--yes")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Initialized project") {
		t.Errorf("output should mention initialization, got: %q", out)
	}
	if !strings.Contains(out, "Added to setup") {
		t.Errorf("output should mention addition, got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "project:") {
		t.Errorf("pi.yaml should have project:, got: %q", content)
	}
	if !strings.Contains(content, "pi:install-uv") {
		t.Errorf("pi.yaml should contain pi:install-uv, got: %q", content)
	}
}

func TestSetupAdd_LocalAutomation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "setup/install-deps")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if strings.Contains(out, "Resolved") {
		t.Errorf("should not show resolution for local automation, got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if !strings.Contains(string(data), "setup/install-deps") {
		t.Errorf("pi.yaml should contain setup/install-deps, got: %q", string(data))
	}
}

func TestSetupAdd_MultipleAdds(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	if _, code := runPi(t, dir, "setup", "add", "homebrew", "--if", "os.macos"); code != 0 {
		t.Fatal("first add failed")
	}
	if _, code := runPi(t, dir, "setup", "add", "uv"); code != 0 {
		t.Fatal("second add failed")
	}
	if _, code := runPi(t, dir, "setup", "add", "python", "--version", "3.13"); code != 0 {
		t.Fatal("third add failed")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "pi:install-homebrew") {
		t.Error("missing pi:install-homebrew")
	}
	if !strings.Contains(content, "pi:install-uv") {
		t.Error("missing pi:install-uv")
	}
	if !strings.Contains(content, "pi:install-python") {
		t.Error("missing pi:install-python")
	}
}

func TestSetupAdd_NoArgs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	_, code := runPi(t, dir, "setup", "add")
	if code == 0 {
		t.Error("expected non-zero exit code for no args")
	}
}

func TestSetupAdd_PiPrefixExpansion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:go", "--version", "1.23")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Resolved 'pi:go'") {
		t.Errorf("output should show resolution, got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if !strings.Contains(string(data), "pi:install-go") {
		t.Errorf("pi.yaml should contain pi:install-go, got: %q", string(data))
	}
}

func TestSetupAdd_ReplaceSameRunTarget(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\nsetup:\n  - pi:install-node\n"), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:install-node", "--version", "22")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Replaced in pi.yaml") {
		t.Errorf("output should say 'Replaced in pi.yaml', got: %q", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if strings.Count(content, "pi:install-node") != 1 {
		t.Errorf("should have exactly one pi:install-node entry, got:\n%s", content)
	}
	if !strings.Contains(content, `"22"`) {
		t.Errorf("should contain version 22, got:\n%s", content)
	}
}

func TestSetupAdd_ReplacePreservesPosition(t *testing.T) {
	dir := t.TempDir()
	initial := "project: test\nsetup:\n  - pi:install-uv\n  - pi:install-node\n  - pi:install-python\n"
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(initial), 0o644)

	out, code := runPi(t, dir, "setup", "add", "pi:install-node", "--version", "22")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	uvIdx := strings.Index(content, "pi:install-uv")
	nodeIdx := strings.Index(content, "pi:install-node")
	pythonIdx := strings.Index(content, "pi:install-python")

	if uvIdx >= nodeIdx || nodeIdx >= pythonIdx {
		t.Errorf("replacement should preserve position. uv@%d, node@%d, python@%d\ncontent:\n%s", uvIdx, nodeIdx, pythonIdx, content)
	}
}

func TestSetupAdd_PreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: my-app\n\nshortcuts:\n  up: docker/up\n"), 0o644)

	_, code := runPi(t, dir, "setup", "add", "uv")
	if code != 0 {
		t.Fatal("add failed")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	content := string(data)
	if !strings.Contains(content, "project: my-app") {
		t.Error("project should be preserved")
	}
	if !strings.Contains(content, "up: docker/up") {
		t.Error("shortcuts should be preserved")
	}
	if !strings.Contains(content, "pi:install-uv") {
		t.Error("new entry should be added")
	}
}
