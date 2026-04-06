package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/validate"
)

func TestValidateHelp(t *testing.T) {
	out, err := executeCmd("validate", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "validate") {
		t.Errorf("expected help to mention validate, got: %s", out)
	}
	if !strings.Contains(out, "schema") {
		t.Errorf("expected help to mention schema, got: %s", out)
	}
}

func TestValidateInRootHelp(t *testing.T) {
	out, err := executeCmd("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "validate") {
		t.Errorf("expected root help to list validate subcommand, got: %s", out)
	}
}

func TestValidate_ValidProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  greet: hello
setup:
  - run: hello
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker, got: %s", out)
	}
	if !strings.Contains(out, "automation(s)") {
		t.Errorf("expected automation count, got: %s", out)
	}
	if !strings.Contains(out, "1 shortcut") {
		t.Errorf("expected shortcut count, got: %s", out)
	}
	if !strings.Contains(out, "1 setup") {
		t.Errorf("expected setup count, got: %s", out)
	}
}

func TestValidate_NoAutomations(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: empty\n"), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success marker, got: %s", out)
	}
}

func TestValidate_BrokenShortcutRef(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  bad: nonexistent-automation
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken shortcut reference")
	}
	var exitErr *executor.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *executor.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "shortcut") {
		t.Errorf("expected error to mention shortcut, got: %s", errOut)
	}
	if !strings.Contains(errOut, "nonexistent-automation") {
		t.Errorf("expected error to mention the automation name, got: %s", errOut)
	}
}

func TestValidate_BrokenSetupRef(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
setup:
  - run: nonexistent-setup
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken setup reference")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "setup[0]") {
		t.Errorf("expected error to mention setup index, got: %s", errOut)
	}
	if !strings.Contains(errOut, "nonexistent-setup") {
		t.Errorf("expected error to mention the automation name, got: %s", errOut)
	}
}

func TestValidate_BrokenRunStep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "caller.yaml"), []byte(`name: caller
description: Calls a nonexistent automation
steps:
  - run: does-not-exist
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken run step reference")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "caller") {
		t.Errorf("expected error to mention automation name, got: %s", errOut)
	}
	if !strings.Contains(errOut, "does-not-exist") {
		t.Errorf("expected error to mention target, got: %s", errOut)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  bad1: missing-1
  bad2: missing-2
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken references")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "missing-1") {
		t.Errorf("expected first broken ref, got: %s", errOut)
	}
	if !strings.Contains(errOut, "missing-2") {
		t.Errorf("expected second broken ref, got: %s", errOut)
	}
	if !strings.Contains(errOut, "2 error") {
		t.Errorf("expected error count, got: %s", errOut)
	}
}

func TestValidate_ValidRunStep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "caller.yaml"), []byte(`name: caller
description: Calls hello
steps:
  - run: hello
`), 0644)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_BuiltinRefValid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
setup:
  - run: pi:install-uv
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success for built-in reference, got: %s", out)
	}
}

func TestValidate_NoPiYaml(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when no pi.yaml found")
	}
}

func TestValidate_BrokenFileRef(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build with script
steps:
  - bash: build.sh
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken file reference")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "file not found") {
		t.Errorf("expected 'file not found' error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "build.sh") {
		t.Errorf("expected error to mention build.sh, got: %s", errOut)
	}
}

func TestValidate_ValidFileRef(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build with script
steps:
  - bash: build.sh
`), 0644)
	os.WriteFile(filepath.Join(piDir, "build.sh"), []byte("#!/bin/bash\necho build\n"), 0755)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "✓") {
		t.Errorf("expected success, got: %s", out)
	}
}

func TestValidate_InlineScriptNotFlagged(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`description: Say hello
steps:
  - bash: echo hello world
  - python: "import sys; print('hi')"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_BrokenFileRefInFirstBlock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "multi.yaml"), []byte(`description: Multi platform
steps:
  - first:
      - bash: install-mac.sh
        if: os.macos
      - bash: install-linux.sh
        if: os.linux
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken file references in first: block")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "install-mac.sh") {
		t.Errorf("expected error to mention install-mac.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install-linux.sh") {
		t.Errorf("expected error to mention install-linux.sh, got: %s", errOut)
	}
}

func TestValidate_BrokenFileRefInSubdir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi", "deploy")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "push.yaml"), []byte(`description: Push image
steps:
  - bash: push.sh
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken file reference")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "push.sh") {
		t.Errorf("expected error to mention push.sh, got: %s", errOut)
	}
}

func TestValidate_MultipleFileRefErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
steps:
  - bash: compile.sh
  - python: transform.py
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken file references")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "compile.sh") {
		t.Errorf("expected error for compile.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "transform.py") {
		t.Errorf("expected error for transform.py, got: %s", errOut)
	}
	if !strings.Contains(errOut, "2 error") {
		t.Errorf("expected 2 errors, got: %s", errOut)
	}
}

func TestValidate_InstallerScalarFileRef_Broken(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: check.sh
  run: install.sh
  version: tool --version
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken installer scalar file refs")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "check.sh") {
		t.Errorf("expected error for check.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install.sh") {
		t.Errorf("expected error for install.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install.test") {
		t.Errorf("expected error to mention install.test context, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install.run") {
		t.Errorf("expected error to mention install.run context, got: %s", errOut)
	}
}

func TestValidate_InstallerScalarFileRef_Valid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: check.sh
  run: install.sh
  version: tool --version
`), 0644)
	os.WriteFile(filepath.Join(piDir, "check.sh"), []byte("#!/bin/bash\ncommand -v tool\n"), 0755)
	os.WriteFile(filepath.Join(piDir, "install.sh"), []byte("#!/bin/bash\nbrew install tool\n"), 0755)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success, got: %s", stdout.String())
	}
}

func TestValidate_InstallerStepListFileRef_Broken(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test:
    - bash: check.sh
  run:
    - bash: install.sh
  version: tool --version
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken installer step-list file refs")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "check.sh") {
		t.Errorf("expected error for check.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install.sh") {
		t.Errorf("expected error for install.sh, got: %s", errOut)
	}
}

func TestValidate_InstallerFirstBlockFileRef_Broken(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: command -v tool
  run:
    - first:
        - bash: install-mac.sh
          if: os.macos
        - bash: install-linux.sh
          if: os.linux
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken file refs in installer first: block")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "install-mac.sh") {
		t.Errorf("expected error for install-mac.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install-linux.sh") {
		t.Errorf("expected error for install-linux.sh, got: %s", errOut)
	}
}

func TestValidate_InstallerVerifyPhaseFileRef(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: command -v tool
  run: brew install tool
  verify: verify.sh
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken verify file ref")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "verify.sh") {
		t.Errorf("expected error for verify.sh, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install.verify") {
		t.Errorf("expected error to mention install.verify context, got: %s", errOut)
	}
}

func TestValidate_InstallerInlineScriptNotFlagged(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: command -v tool >/dev/null 2>&1
  run: brew install tool
  version: tool --version | head -1
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success for inline installer scripts, got: %s", stdout.String())
	}
}

func TestValidate_ShortcutWithValidInputs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  py:
    run: install-python
    with:
      version: "3.13"
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success, got: %s", stdout.String())
	}
}

func TestValidate_ShortcutWithUnknownInput(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  py:
    run: install-python
    with:
      vrsion: "3.13"
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown input key")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "vrsion") {
		t.Errorf("expected error to mention the typo key, got: %s", errOut)
	}
	if !strings.Contains(errOut, "shortcut") {
		t.Errorf("expected error to mention shortcut context, got: %s", errOut)
	}
	if !strings.Contains(errOut, "version") {
		t.Errorf("expected error to mention available inputs, got: %s", errOut)
	}
}

func TestValidate_ShortcutWithNoInputsOnTarget(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  greet:
    run: hello
    with:
      name: world
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`description: Say hello
bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for with: on target with no inputs")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", errOut)
	}
}

func TestValidate_SetupWithUnknownInput(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
setup:
  - run: install-python
    with:
      vrsion: "3.13"
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown setup input key")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "setup[0]") {
		t.Errorf("expected error to mention setup index, got: %s", errOut)
	}
	if !strings.Contains(errOut, "vrsion") {
		t.Errorf("expected error to mention the typo key, got: %s", errOut)
	}
}

func TestValidate_SetupWithValidInputs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
setup:
  - run: install-python
    with:
      version: "3.13"
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success, got: %s", stdout.String())
	}
}

func TestValidate_RunStepWithUnknownInput(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)
	os.WriteFile(filepath.Join(piDir, "setup-all.yaml"), []byte(`description: Setup everything
steps:
  - run: install-python
    with:
      vrsion: "3.13"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown run step input key")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "vrsion") {
		t.Errorf("expected error to mention the typo key, got: %s", errOut)
	}
	if !strings.Contains(errOut, "install-python") {
		t.Errorf("expected error to mention target automation, got: %s", errOut)
	}
}

func TestValidate_RunStepWithValidInputs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)
	os.WriteFile(filepath.Join(piDir, "setup-all.yaml"), []byte(`description: Setup everything
steps:
  - run: install-python
    with:
      version: "3.13"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success, got: %s", stdout.String())
	}
}

func TestValidate_RunStepWithNoInputsOnTarget(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`description: Say hello
bash: echo hello
`), 0644)
	os.WriteFile(filepath.Join(piDir, "caller.yaml"), []byte(`description: Calls hello with bogus inputs
steps:
  - run: hello
    with:
      name: world
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for with: on target with no inputs")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", errOut)
	}
}

func TestValidate_RunStepWithInFirstBlock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)
	os.WriteFile(filepath.Join(piDir, "setup-all.yaml"), []byte(`description: Setup everything
steps:
  - first:
      - run: install-python
        with:
          vrsion: "3.13"
        if: os.macos
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown input in first: block run step")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "vrsion") {
		t.Errorf("expected error to mention the typo key, got: %s", errOut)
	}
	if !strings.Contains(errOut, "first") {
		t.Errorf("expected error to mention first: block context, got: %s", errOut)
	}
}

func TestValidate_ShortcutWithBrokenRefSkipsInputCheck(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  bad:
    run: nonexistent
    with:
      version: "3.13"
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`description: Say hello
bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for broken reference")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "nonexistent") {
		t.Errorf("expected error to mention broken ref, got: %s", errOut)
	}
	// Should only have the broken ref error, not an input error
	if strings.Contains(errOut, "no declared inputs") {
		t.Errorf("should not report input errors when target is unresolvable, got: %s", errOut)
	}
}

func TestValidate_MultipleWithErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  py:
    run: install-python
    with:
      vrsion: "3.13"
      platform: linux
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-python.yaml"), []byte(`description: Install Python
inputs:
  version:
    type: string
    description: Python version
bash: echo "installing python $PI_IN_VERSION"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for multiple unknown input keys")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "vrsion") {
		t.Errorf("expected error for 'vrsion', got: %s", errOut)
	}
	if !strings.Contains(errOut, "platform") {
		t.Errorf("expected error for 'platform', got: %s", errOut)
	}
}

func TestValidate_SetupWithNoInputsOnTarget(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
setup:
  - run: hello
    with:
      name: world
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`description: Say hello
bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for with: on target with no inputs")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "setup[0]") {
		t.Errorf("expected error to mention setup index, got: %s", errOut)
	}
	if !strings.Contains(errOut, "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", errOut)
	}
}

func TestCheckWithInputs_NoWith(t *testing.T) {
	a := &automation.Automation{Name: "test"}
	msgs := validate.CheckWithInputs(nil, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors for nil with, got: %v", msgs)
	}
	msgs = validate.CheckWithInputs(map[string]string{}, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors for empty with, got: %v", msgs)
	}
}

func TestCheckWithInputs_AllValid(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}, "arch": {}},
		InputKeys: []string{"version", "arch"},
	}
	msgs := validate.CheckWithInputs(map[string]string{"version": "3.13", "arch": "arm64"}, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors, got: %v", msgs)
	}
}

func TestCheckWithInputs_UnknownKey(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}},
		InputKeys: []string{"version"},
	}
	msgs := validate.CheckWithInputs(map[string]string{"vrsion": "3.13"}, a)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "vrsion") {
		t.Errorf("expected error to mention 'vrsion', got: %s", msgs[0])
	}
	if !strings.Contains(msgs[0], "version") {
		t.Errorf("expected error to list available inputs, got: %s", msgs[0])
	}
}

func TestCheckWithInputs_NoInputsOnTarget(t *testing.T) {
	a := &automation.Automation{Name: "hello"}
	msgs := validate.CheckWithInputs(map[string]string{"name": "world"}, a)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", msgs[0])
	}
}

func TestCheckWithInputs_MultipleUnknownSorted(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}},
		InputKeys: []string{"version"},
	}
	msgs := validate.CheckWithInputs(map[string]string{"platform": "linux", "arch": "arm64"}, a)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "arch") {
		t.Errorf("expected first error (sorted) to be about 'arch', got: %s", msgs[0])
	}
	if !strings.Contains(msgs[1], "platform") {
		t.Errorf("expected second error (sorted) to be about 'platform', got: %s", msgs[1])
	}
}

// --- Circular dependency detection tests ---

func TestValidate_DirectCircularDep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "alpha.yaml"), []byte(`description: Alpha
steps:
  - run: beta
`), 0644)
	os.WriteFile(filepath.Join(piDir, "beta.yaml"), []byte(`description: Beta
steps:
  - run: alpha
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "alpha") {
		t.Errorf("expected error to mention 'alpha', got: %s", errOut)
	}
	if !strings.Contains(errOut, "beta") {
		t.Errorf("expected error to mention 'beta', got: %s", errOut)
	}
	if !strings.Contains(errOut, "→") {
		t.Errorf("expected error to contain chain arrow, got: %s", errOut)
	}
}

func TestValidate_IndirectCircularDep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "a.yaml"), []byte(`description: A
steps:
  - run: b
`), 0644)
	os.WriteFile(filepath.Join(piDir, "b.yaml"), []byte(`description: B
steps:
  - run: c
`), 0644)
	os.WriteFile(filepath.Join(piDir, "c.yaml"), []byte(`description: C
steps:
  - run: a
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for indirect circular dependency")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "a") && !strings.Contains(errOut, "b") && !strings.Contains(errOut, "c") {
		t.Errorf("expected error to mention all three automations, got: %s", errOut)
	}
}

func TestValidate_SelfReferencing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "loop.yaml"), []byte(`description: Loop
steps:
  - run: loop
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for self-referencing automation")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "loop") {
		t.Errorf("expected error to mention 'loop', got: %s", errOut)
	}
}

func TestValidate_CircularDepInFirstBlock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "start.yaml"), []byte(`description: Start
steps:
  - first:
      - run: end
        if: os.macos
      - bash: echo fallback
`), 0644)
	os.WriteFile(filepath.Join(piDir, "end.yaml"), []byte(`description: End
steps:
  - run: start
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for circular dep through first: block")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %s", errOut)
	}
}

func TestValidate_DiamondDep_NoCycle(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "top.yaml"), []byte(`description: Top
steps:
  - run: left
  - run: right
`), 0644)
	os.WriteFile(filepath.Join(piDir, "left.yaml"), []byte(`description: Left
steps:
  - run: bottom
`), 0644)
	os.WriteFile(filepath.Join(piDir, "right.yaml"), []byte(`description: Right
steps:
  - run: bottom
`), 0644)
	os.WriteFile(filepath.Join(piDir, "bottom.yaml"), []byte(`description: Bottom
bash: echo done
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("diamond dependencies should not be flagged: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success marker, got: %s", stdout.String())
	}
}

func TestValidate_NoCircularDep_LinearChain(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "first.yaml"), []byte(`description: First
steps:
  - run: second
`), 0644)
	os.WriteFile(filepath.Join(piDir, "second.yaml"), []byte(`description: Second
steps:
  - run: third
`), 0644)
	os.WriteFile(filepath.Join(piDir, "third.yaml"), []byte(`description: Third
bash: echo end
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("linear chain should not be flagged: %v\nstderr: %s", err, stderr.String())
	}
}

func TestValidate_CircularDepChainFormat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "x.yaml"), []byte(`description: X
steps:
  - run: y
`), 0644)
	os.WriteFile(filepath.Join(piDir, "y.yaml"), []byte(`description: Y
steps:
  - run: x
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	errOut := stderr.String()
	// Chain should show the full cycle: x → y → x
	if !strings.Contains(errOut, " → ") {
		t.Errorf("expected chain with arrows, got: %s", errOut)
	}
}

func TestValidate_CircularDepWithOtherErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  bad: nonexistent
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "a.yaml"), []byte(`description: A
steps:
  - run: b
`), 0644)
	os.WriteFile(filepath.Join(piDir, "b.yaml"), []byte(`description: B
steps:
  - run: a
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected errors")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "nonexistent") {
		t.Errorf("expected broken ref error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "circular dependency") {
		t.Errorf("expected circular dependency error, got: %s", errOut)
	}
}

// --- Unit tests for cycle detection internals ---

func TestDetectCycles_NoCycles(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": nil,
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}
}

func TestDetectCycles_DirectCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	cycle := cycles[0]
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node, got: %v", cycle)
	}
}

func TestDetectCycles_SelfLoop(t *testing.T) {
	graph := map[string][]string{
		"x": {"x"},
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	if len(cycles[0]) != 2 {
		t.Errorf("self-loop cycle should have 2 elements (x → x), got: %v", cycles[0])
	}
}

func TestDetectCycles_ThreeNodeCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_DiamondNoCycle(t *testing.T) {
	graph := map[string][]string{
		"top":    {"left", "right"},
		"left":   {"bottom"},
		"right":  {"bottom"},
		"bottom": nil,
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("diamond should have no cycles, got: %v", cycles)
	}
}

func TestDetectCycles_MultipleCycles(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"},
		"c": {"d"},
		"d": {"c"},
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_DisconnectedGraphWithCycle(t *testing.T) {
	graph := map[string][]string{
		"a":      {"b"},
		"b":      {"a"},
		"island": nil,
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle (island should not affect detection), got %d: %v", len(cycles), cycles)
	}
}

func TestNormalizeCycleKey_Rotation(t *testing.T) {
	k1 := validate.NormalizeCycleKey([]string{"b", "c", "a", "b"})
	k2 := validate.NormalizeCycleKey([]string{"a", "b", "c", "a"})
	if k1 != k2 {
		t.Errorf("rotated cycles should normalize to same key: %q vs %q", k1, k2)
	}
}

func TestNormalizeCycleKey_SelfLoop(t *testing.T) {
	key := validate.NormalizeCycleKey([]string{"x", "x"})
	if key != "x" {
		t.Errorf("self-loop key = %q, want %q", key, "x")
	}
}

func TestBuildRunGraph_SkipsBrokenRefs(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": nil,
	}
	cycles := validate.DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles for valid graph, got: %v", cycles)
	}
}

// --- Condition validation tests ---

func TestValidate_ValidConditions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build with conditions
if: os.macos
steps:
  - bash: echo step1
    if: command.docker
  - bash: echo step2
    if: os.linux or os.macos
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error for valid conditions, got: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("expected success, got: %s", stdout.String())
	}
}

func TestValidate_ConditionSyntaxErrorCaughtAtLoadTime(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Broken syntax
steps:
  - bash: echo hello
    if: os.macos and and os.linux
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for syntax error in if: condition")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "invalid if expression") {
		t.Errorf("expected error to mention 'invalid if expression', got: %s", errOut)
	}
}

func TestValidate_UnknownPredicateInCondition(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Unknown predicate
steps:
  - bash: echo hello
    if: os.macoss
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown predicate")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "unknown predicate") {
		t.Errorf("expected error to mention 'unknown predicate', got: %s", errOut)
	}
	if !strings.Contains(errOut, "os.macoss") {
		t.Errorf("expected error to mention the bad predicate name, got: %s", errOut)
	}
}

func TestValidate_UnknownPredicateOnAutomation(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Bad automation-level condition
if: os.freebsd
bash: echo hello
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown predicate on automation-level if:")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "unknown predicate") {
		t.Errorf("expected error to mention 'unknown predicate', got: %s", errOut)
	}
	if !strings.Contains(errOut, "os.freebsd") {
		t.Errorf("expected error to mention 'os.freebsd', got: %s", errOut)
	}
	if !strings.Contains(errOut, "build") {
		t.Errorf("expected error to mention automation name 'build', got: %s", errOut)
	}
}

func TestValidate_ConditionInFirstBlock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: First block with bad condition
steps:
  - first:
      - bash: echo one
        if: command.mise
      - bash: echo two
        if: bogus.pred
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown predicate in first: block")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "unknown predicate") {
		t.Errorf("expected error to mention 'unknown predicate', got: %s", errOut)
	}
	if !strings.Contains(errOut, "bogus.pred") {
		t.Errorf("expected error to mention 'bogus.pred', got: %s", errOut)
	}
}

func TestValidate_DynamicPredicatesPass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Dynamic predicates
steps:
  - bash: echo step1
    if: command.docker
  - bash: echo step2
    if: env.CI
  - bash: echo step3
    if: "file.exists(\".env\")"
  - bash: echo step4
    if: "dir.exists(\"src\")"
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected no error for dynamic predicates, got: %v\nstderr: %s", err, stderr.String())
	}
}

func TestValidate_MultipleConditionErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Multiple bad conditions
if: os.bsd
steps:
  - bash: echo one
    if: shell.fish
  - bash: echo two
    if: os.macoss
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected errors for multiple bad conditions")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "os.bsd") {
		t.Errorf("expected error to mention 'os.bsd', got: %s", errOut)
	}
	if !strings.Contains(errOut, "shell.fish") {
		t.Errorf("expected error to mention 'shell.fish', got: %s", errOut)
	}
	if !strings.Contains(errOut, "os.macoss") {
		t.Errorf("expected error to mention 'os.macoss', got: %s", errOut)
	}
	if !strings.Contains(errOut, "3 error(s)") {
		t.Errorf("expected 3 errors, got: %s", errOut)
	}
}

func TestValidate_ConditionInInstallPhase(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "install-tool.yaml"), []byte(`description: Install tool
install:
  test: command -v tool
  run:
    - first:
        - bash: mise install tool
          if: command.mise
        - bash: brew install tool
          if: os.macoss
  version: tool --version
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown predicate in install phase")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "unknown predicate") {
		t.Errorf("expected error to mention 'unknown predicate', got: %s", errOut)
	}
	if !strings.Contains(errOut, "os.macoss") {
		t.Errorf("expected error to mention 'os.macoss', got: %s", errOut)
	}
}

func TestValidate_ConditionWithOtherErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
shortcuts:
  bad: nonexistent
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Bad condition and broken ref
steps:
  - bash: echo hello
    if: os.macoss
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected errors")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "nonexistent") {
		t.Errorf("expected broken ref error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "os.macoss") {
		t.Errorf("expected condition error, got: %s", errOut)
	}
}

func TestValidate_UnknownFieldDetected(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`descrption: Build the project
bash: go build ./...
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown field 'descrption'")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "descrption") {
		t.Errorf("expected error to mention 'descrption', got: %s", errOut)
	}
	if !strings.Contains(errOut, "unknown field") {
		t.Errorf("expected 'unknown field' in error, got: %s", errOut)
	}
	if !strings.Contains(errOut, "description") {
		t.Errorf("expected 'description' suggestion, got: %s", errOut)
	}
}

func TestValidate_UnknownFieldInStep(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
steps:
  - bash: go build ./...
    timout: 30s
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown step field 'timout'")
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "timout") {
		t.Errorf("expected error to mention 'timout', got: %s", errOut)
	}
	if !strings.Contains(errOut, "step[0]") {
		t.Errorf("expected error to mention step index, got: %s", errOut)
	}
	if !strings.Contains(errOut, "timeout") {
		t.Errorf("expected 'timeout' suggestion, got: %s", errOut)
	}
}

func TestValidate_ValidFilePassesUnknownFieldCheck(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`project: test
`), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
bash: go build ./...
env:
  GOOS: linux
dir: .
timeout: 30s
silent: true
`), 0644)

	var stdout, stderr bytes.Buffer
	err := runValidate(dir, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error for valid file: %v\nstderr: %s", err, stderr.String())
	}
}
