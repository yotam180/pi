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
	err := listAutomations(root, &buf, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected header with NAME")
	}
	if !strings.Contains(out, "SOURCE") {
		t.Error("expected header with SOURCE")
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
	if !strings.Contains(out, "[workspace]") {
		t.Error("expected [workspace] source indicator")
	}
}

func TestListAutomations_NoDescription(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
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

func TestListAutomations_BuiltinsHiddenByDefault(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "[built-in]") {
		t.Errorf("expected built-in automations hidden by default, got: %s", out)
	}
	if !strings.Contains(out, "No automations found") {
		t.Errorf("expected no automations message when builtins hidden, got: %s", out)
	}
}

func TestListAutomations_BuiltinsShownWithFlag(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[built-in]") {
		t.Errorf("expected built-in automations to appear with --builtins, got: %s", out)
	}
}

func TestListAutomations_FromSubdirectory(t *testing.T) {
	root := setupListWorkspace(t)
	sub := filepath.Join(root, "src", "nested")
	os.MkdirAll(sub, 0o755)

	var buf bytes.Buffer
	err := listAutomations(sub, &buf, false, false)
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
	err := listAutomations(dir, &buf, false, false)
	if err == nil {
		t.Fatal("expected error when no pi.yaml found")
	}
	if !strings.Contains(err.Error(), "pi.yaml") {
		t.Errorf("expected error to mention pi.yaml, got: %v", err)
	}
}

func TestListAutomations_ShowsInputs(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "with-inputs.yaml"), []byte(`name: with-inputs
description: Has inputs
inputs:
  service:
    type: string
    required: true
  tail:
    type: string
    default: "200"
steps:
  - bash: echo hi
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "no-inputs.yaml"), []byte(`name: no-inputs
description: No inputs
steps:
  - bash: echo hi
`), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "INPUTS") {
		t.Error("expected INPUTS header")
	}
	if !strings.Contains(out, "service, tail?") {
		t.Errorf("expected 'service, tail?' in output, got:\n%s", out)
	}
}

func TestListAutomations_Sorted(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
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

func TestListAutomations_WorkspaceSource(t *testing.T) {
	root := setupListWorkspace(t)
	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "hello") && strings.Contains(line, "Say hello") {
			if !strings.Contains(line, "[workspace]") {
				t.Errorf("expected [workspace] source for hello, line: %s", line)
			}
		}
	}
}

func TestListAutomations_PackageSource(t *testing.T) {
	root := t.TempDir()

	// Create pi.yaml with a file: package
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: test
packages:
  - source: file:./mypkg
    as: tools
`), 0o644)

	// Create local automation
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "local.yaml"), []byte(`description: Local auto
bash: echo local
`), 0o644)

	// Create package with automation
	pkgPiDir := filepath.Join(root, "mypkg", ".pi")
	os.MkdirAll(pkgPiDir, 0o755)
	os.WriteFile(filepath.Join(pkgPiDir, "from-pkg.yaml"), []byte(`description: From package
bash: echo pkg
`), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Local should show [workspace]
	if !strings.Contains(out, "[workspace]") {
		t.Errorf("expected [workspace] for local automation, got:\n%s", out)
	}

	// Package automation should show alias
	if !strings.Contains(out, "from-pkg") {
		t.Errorf("expected from-pkg in output, got:\n%s", out)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "from-pkg") {
			if !strings.Contains(line, "tools") {
				t.Errorf("expected 'tools' alias as source for from-pkg, line: %s", line)
			}
		}
	}
}

func TestListAutomations_AllFlag(t *testing.T) {
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte(`project: test
packages:
  - source: file:./mypkg
    as: tools
`), 0o644)

	// Create local automation
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "local.yaml"), []byte(`description: Local auto
bash: echo local
`), 0o644)

	// Create package with automation (including one that shadows local and one that doesn't)
	pkgPiDir := filepath.Join(root, "mypkg", ".pi")
	os.MkdirAll(pkgPiDir, 0o755)
	os.WriteFile(filepath.Join(pkgPiDir, "pkg-only.yaml"), []byte(`description: Only in package
bash: echo pkg
`), 0o644)
	os.WriteFile(filepath.Join(pkgPiDir, "another.yaml"), []byte(`description: Another pkg auto
bash: echo another
`), 0o644)

	var buf bytes.Buffer
	err := listAutomations(root, &buf, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Should have the main table
	if !strings.Contains(out, "local") {
		t.Errorf("expected local in output, got:\n%s", out)
	}

	// Should have package section header
	if !strings.Contains(out, "file:./mypkg") {
		t.Errorf("expected package header in --all output, got:\n%s", out)
	}
	if !strings.Contains(out, "alias: tools") {
		t.Errorf("expected alias in package header, got:\n%s", out)
	}

	// Should show all package automations in the grouped section
	if !strings.Contains(out, "pkg-only") {
		t.Errorf("expected pkg-only in --all output, got:\n%s", out)
	}
	if !strings.Contains(out, "another") {
		t.Errorf("expected another in --all output, got:\n%s", out)
	}
}
