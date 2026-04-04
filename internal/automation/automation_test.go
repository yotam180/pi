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

func TestLoad_ValidBashStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "up.yaml", `
name: docker/up
description: Start all containers

steps:
  - bash: docker-compose up -d
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Name != "docker/up" {
		t.Errorf("name = %q, want %q", a.Name, "docker/up")
	}
	if a.Description != "Start all containers" {
		t.Errorf("description = %q, want %q", a.Description, "Start all containers")
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeBash)
	}
	if a.Steps[0].Value != "docker-compose up -d" {
		t.Errorf("step value = %q, want %q", a.Steps[0].Value, "docker-compose up -d")
	}
}

func TestLoad_MultipleSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "deploy.yaml", `
name: deploy
description: Build and deploy

steps:
  - bash: make build
  - run: docker/push
  - bash: kubectl apply -f deploy.yaml
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(a.Steps))
	}

	want := []struct {
		t StepType
		v string
	}{
		{StepTypeBash, "make build"},
		{StepTypeRun, "docker/push"},
		{StepTypeBash, "kubectl apply -f deploy.yaml"},
	}
	for i, w := range want {
		if a.Steps[i].Type != w.t {
			t.Errorf("step[%d].Type = %q, want %q", i, a.Steps[i].Type, w.t)
		}
		if a.Steps[i].Value != w.v {
			t.Errorf("step[%d].Value = %q, want %q", i, a.Steps[i].Value, w.v)
		}
	}
}

func TestLoad_PipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "logs.yaml", `
name: logs
description: Stream and format logs

steps:
  - bash: docker-compose logs -f
    pipe_to: next
  - bash: format-logs.sh
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
	path := writeFile(t, dir, "setup.yaml", `
name: setup/install-uv
description: Install uv if missing

steps:
  - bash: |
      if ! command -v uv &> /dev/null; then
        curl -LsSf https://astral.sh/uv/install.sh | sh
      fi
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeBash)
	}
	if !strings.Contains(a.Steps[0].Value, "command -v uv") {
		t.Errorf("step value should contain multiline bash, got: %q", a.Steps[0].Value)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestLoad_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
description: No name field

steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
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
		t.Fatal("expected error for no steps")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("error should mention 'at least one step', got: %v", err)
	}
}

func TestLoad_NoStepType(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "nostep.yaml", `
name: bad
steps:
  - pipe_to: next
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for step with no type")
	}
	if !strings.Contains(err.Error(), "must specify one of") {
		t.Errorf("error should mention 'must specify one of', got: %v", err)
	}
}

func TestLoad_MultipleStepTypes(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "multi.yaml", `
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
		t.Errorf("error should mention 'exactly one', got: %v", err)
	}
}

func TestLoad_PythonStep_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "py.yaml", `
name: test
steps:
  - python: script.py
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypePython {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypePython)
	}
	if a.Steps[0].Value != "script.py" {
		t.Errorf("step value = %q, want %q", a.Steps[0].Value, "script.py")
	}
}

func TestLoad_UnsupportedStepType_TypeScript(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "ts.yaml", `
name: test
steps:
  - typescript: script.ts
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unimplemented step type")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("error should mention 'not yet implemented', got: %v", err)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: test
steps: [[[invalid
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention 'parsing', got: %v", err)
	}
}

func TestLoad_EmptyStepValue(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty-val.yaml", `
name: test
steps:
  - bash: ""
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty step value")
	}
	if !strings.Contains(err.Error(), "empty value") {
		t.Errorf("error should mention 'empty value', got: %v", err)
	}
}

func TestStepType_IsImplemented(t *testing.T) {
	tests := []struct {
		st   StepType
		want bool
	}{
		{StepTypeBash, true},
		{StepTypeRun, true},
		{StepTypePython, true},
		{StepTypeTypeScript, false},
	}
	for _, tt := range tests {
		if got := tt.st.IsImplemented(); got != tt.want {
			t.Errorf("StepType(%q).IsImplemented() = %v, want %v", tt.st, got, tt.want)
		}
	}
}
