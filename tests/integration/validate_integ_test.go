package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_ValidProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-valid")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success marker, got: %s", stdout)
	}
	if !strings.Contains(stdout, "shortcut") {
		t.Errorf("expected shortcut count, got: %s", stdout)
	}
	if !strings.Contains(stdout, "setup") {
		t.Errorf("expected setup count, got: %s", stdout)
	}
	if !strings.Contains(stdout, "automation") {
		t.Errorf("expected automation count, got: %s", stdout)
	}
}

func TestValidate_InvalidProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-invalid")
	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "nonexistent-automation") {
		t.Errorf("expected broken shortcut error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "also-nonexistent") {
		t.Errorf("expected broken setup error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "ghost-automation") {
		t.Errorf("expected broken run step error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "3 error") {
		t.Errorf("expected 3 errors counted, got: %s", stderr)
	}
}

func TestValidate_InvalidProject_AllErrorsReported(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-invalid")
	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	errorCount := strings.Count(stderr, "✗")
	if errorCount != 3 {
		t.Errorf("expected 3 error lines (✗), got %d in:\n%s", errorCount, stderr)
	}
}

func TestValidate_BasicProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for basic project, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success for basic project, got: %s", stdout)
	}
}

func TestValidate_BuiltinRefsValid(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for builtins project, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success for builtins project, got: %s", stdout)
	}
}

func TestValidate_BrokenFileReferences(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-file-refs")
	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1 for broken file references, got %d", code)
	}
	if !strings.Contains(stderr, "does-not-exist.sh") {
		t.Errorf("expected error for does-not-exist.sh, got: %s", stderr)
	}
	if !strings.Contains(stderr, "also-missing.py") {
		t.Errorf("expected error for also-missing.py, got: %s", stderr)
	}
	if !strings.Contains(stderr, "file not found") {
		t.Errorf("expected 'file not found' message, got: %s", stderr)
	}
	if !strings.Contains(stderr, "2 error") {
		t.Errorf("expected 2 errors, got: %s", stderr)
	}
}

func TestValidate_ValidFileReferencePasses(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "build.yaml"), `description: Build
steps:
  - bash: build.sh
`)
	writeFile(t, filepath.Join(piDir, "build.sh"), "#!/bin/bash\necho build\n")

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for valid file references, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success marker, got: %s", stdout)
	}
}

func TestValidate_InlineScriptsNotFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "hello.yaml"), `description: Inline
steps:
  - bash: echo hello world
`)

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success, got: %s", stdout)
	}
}

func TestValidate_InstallerScalarFileRefBroken(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-tool.yaml"), `description: Install tool
install:
  test: check.sh
  run: install.sh
  version: tool --version
`)

	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1 for broken installer file refs, got %d", code)
	}
	if !strings.Contains(stderr, "check.sh") {
		t.Errorf("expected error for check.sh, got: %s", stderr)
	}
	if !strings.Contains(stderr, "install.sh") {
		t.Errorf("expected error for install.sh, got: %s", stderr)
	}
}

func TestValidate_InstallerScalarFileRefValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-tool.yaml"), `description: Install tool
install:
  test: check.sh
  run: install.sh
  version: tool --version
`)
	writeFile(t, filepath.Join(piDir, "check.sh"), "#!/bin/bash\ncommand -v tool\n")
	writeFile(t, filepath.Join(piDir, "install.sh"), "#!/bin/bash\nbrew install tool\n")

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for valid installer file refs, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success, got: %s", stdout)
	}
}

func TestValidate_InstallerInlineScriptsNotFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-tool.yaml"), `description: Install tool
install:
  test: command -v tool >/dev/null 2>&1
  run: brew install tool
  version: tool --version
`)

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for inline installer scripts, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success, got: %s", stdout)
	}
}

func TestValidate_WithInputsValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), `project: test
shortcuts:
  py:
    run: install-python
    with:
      version: "3.13"
setup:
  - run: install-python
    with:
      version: "3.12"
`)
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-python.yaml"), `description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`)

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for valid with: inputs, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success marker, got: %s", stdout)
	}
}

func TestValidate_WithInputsInvalid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), `project: test
shortcuts:
  py:
    run: install-python
    with:
      vrsion: "3.13"
`)
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-python.yaml"), `description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`)

	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1 for invalid with: input key, got %d", code)
	}
	if !strings.Contains(stderr, "vrsion") {
		t.Errorf("expected error to mention typo key, got: %s", stderr)
	}
	if !strings.Contains(stderr, "shortcut") {
		t.Errorf("expected error to mention shortcut context, got: %s", stderr)
	}
	if !strings.Contains(stderr, "version") {
		t.Errorf("expected error to mention available inputs, got: %s", stderr)
	}
}

func TestValidate_WithNoInputsOnTarget(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), `project: test
setup:
  - run: hello
    with:
      name: world
`)
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "hello.yaml"), `description: Say hello
bash: echo hello
`)

	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1 for with: on target with no inputs, got %d", code)
	}
	if !strings.Contains(stderr, "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", stderr)
	}
}

func TestValidate_RunStepWithInvalidInput(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), "project: test\n")
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)
	writeFile(t, filepath.Join(piDir, "install-python.yaml"), `description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`)
	writeFile(t, filepath.Join(piDir, "setup-all.yaml"), `description: Setup everything
steps:
  - run: install-python
    with:
      vrsion: "3.13"
`)

	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "vrsion") {
		t.Errorf("expected error to mention typo key, got: %s", stderr)
	}
}

func TestValidate_BuiltinWithValidInputs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), `project: test
setup:
  - run: pi:install-python
    with:
      version: "3.13"
`)
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)

	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for valid builtin with: input, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success, got: %s", stdout)
	}
}

func TestValidate_BuiltinWithInvalidInput(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pi.yaml"), `project: test
setup:
  - run: pi:install-python
    with:
      vrsion: "3.13"
`)
	piDir := filepath.Join(dir, ".pi")
	mkdirAll(t, piDir)

	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1 for invalid builtin with: input, got %d", code)
	}
	if !strings.Contains(stderr, "vrsion") {
		t.Errorf("expected error to mention typo key, got: %s", stderr)
	}
	if !strings.Contains(stderr, "version") {
		t.Errorf("expected error to mention available inputs, got: %s", stderr)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("creating %s: %v", path, err)
	}
}
