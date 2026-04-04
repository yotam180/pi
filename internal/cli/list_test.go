package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupListWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(filepath.Join(piDir, "docker"), 0o755)

	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hi
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "docker", "up.yaml"), []byte(`name: docker/up
description: Start containers
steps:
  - bash: echo up
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "no-desc.yaml"), []byte(`name: no-desc
steps:
  - bash: echo x
`), 0o644)

	return root
}

func TestListAutomations_Success(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected header with NAME")
	}
	if !strings.Contains(out, "DESCRIPTION") {
		t.Error("expected header with DESCRIPTION")
	}
	if !strings.Contains(out, "hello") {
		t.Error("expected hello in output")
	}
	if !strings.Contains(out, "docker/up") {
		t.Error("expected docker/up in output")
	}
	if !strings.Contains(out, "Say hello") {
		t.Error("expected description in output")
	}
}

func TestListAutomations_NoDescription(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "no-desc") {
		t.Error("expected no-desc in output")
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "no-desc") && !strings.Contains(line, "-") {
			t.Errorf("expected dash placeholder for empty description, line: %s", line)
		}
	}
}

func TestListAutomations_Empty(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No automations found") {
		t.Errorf("expected friendly message, got: %s", buf.String())
	}
}

func TestListAutomations_FromSubdirectory(t *testing.T) {
	root := setupListWorkspace(t)
	sub := filepath.Join(root, "src", "nested")
	os.MkdirAll(sub, 0o755)

	var buf bytes.Buffer
	err := listAutomations(sub, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Error("expected automations listed when called from subdirectory")
	}
}

func TestListAutomations_NoPiYaml(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := listAutomations(dir, &buf)
	if err == nil {
		t.Fatal("expected error when no pi.yaml found")
	}
	if !strings.Contains(err.Error(), "pi.yaml") {
		t.Errorf("expected error to mention pi.yaml, got: %v", err)
	}
}

func TestListAutomations_Sorted(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	dockerIdx := strings.Index(out, "docker/up")
	helloIdx := strings.Index(out, "hello")
	noDescIdx := strings.Index(out, "no-desc")

	if dockerIdx > helloIdx || helloIdx > noDescIdx {
		t.Errorf("expected sorted order docker/up < hello < no-desc, got indices: docker/up=%d hello=%d no-desc=%d",
			dockerIdx, helloIdx, noDescIdx)
	}
}
