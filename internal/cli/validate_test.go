package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/executor"
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
	exitErr, ok := err.(*executor.ExitError)
	if !ok {
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
