package automation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func boolPtr(b bool) *bool { return &b }

func TestLoad_ValidBashStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "hello.yaml", `
name: hello
description: Say hello
steps:
  - bash: echo "Hello, World!"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Name != "hello" {
		t.Errorf("name = %q, want %q", a.Name, "hello")
	}
	if a.Description != "Say hello" {
		t.Errorf("description = %q, want %q", a.Description, "Say hello")
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeBash)
	}
	if a.Steps[0].Value != `echo "Hello, World!"` {
		t.Errorf("step value = %q", a.Steps[0].Value)
	}
}

func TestLoad_MultipleSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "multi.yaml", `
name: multi
description: Multiple steps
steps:
  - bash: echo "step 1"
  - bash: echo "step 2"
  - bash: echo "step 3"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(a.Steps))
	}

	for i, step := range a.Steps {
		if step.Type != StepTypeBash {
			t.Errorf("step[%d] type = %q, want %q", i, step.Type, StepTypeBash)
		}
	}
}

func TestLoad_PipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "pipe.yaml", `
name: pipe-test
steps:
  - bash: echo data
    pipe_to: next
  - bash: cat
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Steps[0].PipeTo != "next" {
		t.Errorf("step[0].PipeTo = %q, want %q", a.Steps[0].PipeTo, "next")
	}
	if a.Steps[1].PipeTo != "" {
		t.Errorf("step[1].PipeTo = %q, want empty", a.Steps[1].PipeTo)
	}
}

func TestLoad_InlineBashMultiline(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "multiline.yaml", `
name: multiline
steps:
  - bash: |
      echo "line 1"
      echo "line 2"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(a.Steps[0].Value, "line 1") {
		t.Errorf("expected multiline bash, got: %q", a.Steps[0].Value)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestLoad_MissingName_Allowed(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-name.yaml", `
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "" {
		t.Errorf("expected empty name (to be set by discovery), got %q", a.Name)
	}
}

func TestLoad_NoSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty.yaml", `
name: empty
description: No steps
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing steps")
	}
	if !strings.Contains(err.Error(), "steps") {
		t.Errorf("expected 'steps' in error, got: %v", err)
	}
}

func TestLoad_NoStepType(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-type.yaml", `
name: no-type
steps:
  - pipe_to: next
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for step with no type")
	}
	if !strings.Contains(err.Error(), "must specify one of") {
		t.Errorf("expected 'must specify' error, got: %v", err)
	}
}

func TestLoad_MultipleStepTypes(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
steps:
  - bash: echo hello
    run: other/thing
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for multiple step types")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Errorf("expected 'exactly one' error, got: %v", err)
	}
}

func TestLoad_PythonStep_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "python.yaml", `
name: py-test
steps:
  - python: |
      import sys
      print("hello from python")
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Type != StepTypePython {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypePython)
	}
}

func TestLoad_TypeScriptStep_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "typescript.yaml", `
name: ts-test
steps:
  - typescript: |
      const msg: string = "hello from typescript";
      console.log(msg);
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Type != StepTypeTypeScript {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeTypeScript)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
steps:
  - bash: echo hello
  invalid yaml here: [[[
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoad_EmptyStepValue(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty-step.yaml", `
name: empty-step
steps:
  - bash: ""
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty step value")
	}
	if !strings.Contains(err.Error(), "empty value") {
		t.Errorf("expected 'empty value' error, got: %v", err)
	}
}

func TestStepType_IsValid(t *testing.T) {
	tests := []struct {
		st   StepType
		want bool
	}{
		{StepTypeBash, true},
		{StepTypeRun, true},
		{StepTypePython, true},
		{StepTypeTypeScript, true},
		{StepType("ruby"), false},
		{StepType(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.st), func(t *testing.T) {
			if got := tt.st.IsValid(); got != tt.want {
				t.Errorf("StepType(%q).IsValid() = %v, want %v", tt.st, got, tt.want)
			}
		})
	}
}
