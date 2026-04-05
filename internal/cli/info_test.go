package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func setupInfoWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "simple.yaml"), []byte(`name: simple
description: A simple automation
steps:
  - bash: echo hello
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "no-desc.yaml"), []byte(`name: no-desc
steps:
  - bash: echo x
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "with-inputs.yaml"), []byte(`name: with-inputs
description: Automation with inputs
inputs:
  service:
    type: string
    required: true
    description: The service to target
  tail:
    type: string
    default: "200"
    description: Number of lines to tail
  verbose:
    type: string
    required: false
steps:
  - bash: echo hi
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "multi-step.yaml"), []byte(`name: multi-step
description: Multiple steps
steps:
  - bash: echo one
  - bash: echo two
  - bash: echo three
`), 0o644)

	return root
}

func TestShowAutomationInfo_Simple(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "simple", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name:         simple") {
		t.Errorf("expected name, got:\n%s", out)
	}
	if !strings.Contains(out, "Description:  A simple automation") {
		t.Errorf("expected description, got:\n%s", out)
	}
	if !strings.Contains(out, "Steps:        1") {
		t.Errorf("expected step count, got:\n%s", out)
	}
	if !strings.Contains(out, "No inputs.") {
		t.Errorf("expected 'No inputs.' for automation without inputs, got:\n%s", out)
	}
}

func TestShowAutomationInfo_NoDescription(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "no-desc", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "(no description)") {
		t.Errorf("expected '(no description)' placeholder, got:\n%s", out)
	}
}

func TestShowAutomationInfo_WithInputs(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-inputs", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "Inputs:") {
		t.Errorf("expected Inputs header, got:\n%s", out)
	}
	if !strings.Contains(out, "service (string, required) → $PI_IN_SERVICE") {
		t.Errorf("expected required input with env var, got:\n%s", out)
	}
	if !strings.Contains(out, "The service to target") {
		t.Errorf("expected input description, got:\n%s", out)
	}
	if !strings.Contains(out, `tail (string, optional, default: "200") → $PI_IN_TAIL`) {
		t.Errorf("expected optional input with default and env var, got:\n%s", out)
	}
	if !strings.Contains(out, "verbose (string, optional) → $PI_IN_VERBOSE") {
		t.Errorf("expected optional input without default with env var, got:\n%s", out)
	}
}

func TestShowAutomationInfo_RequiredDistinguished(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-inputs", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "required") {
		t.Error("expected 'required' to appear for required inputs")
	}
	if !strings.Contains(out, "optional") {
		t.Error("expected 'optional' to appear for optional inputs")
	}
}

func TestShowAutomationInfo_MultiStep(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "multi-step", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Steps:        3") {
		t.Errorf("expected 3 steps, got:\n%s", buf.String())
	}
}

func TestShowAutomationInfo_NotFound(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for unknown automation")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "simple") {
		t.Errorf("expected available automations listed in error, got: %v", err)
	}
}

func TestShowAutomationInfo_NoPiYaml(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := showAutomationInfo(dir, "anything", &buf)
	if err == nil {
		t.Fatal("expected error when no pi.yaml found")
	}
	if !strings.Contains(err.Error(), "pi.yaml") {
		t.Errorf("expected error to mention pi.yaml, got: %v", err)
	}
}

func TestShowAutomationInfo_FromSubdirectory(t *testing.T) {
	root := setupInfoWorkspace(t)
	sub := filepath.Join(root, "src", "deep")
	os.MkdirAll(sub, 0o755)

	var buf bytes.Buffer
	err := showAutomationInfo(sub, "simple", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "simple") {
		t.Error("expected info output when called from subdirectory")
	}
}

func TestShowAutomationInfo_DefaultValues(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-inputs", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `default: "200"`) {
		t.Errorf("expected default value shown, got:\n%s", out)
	}
}

func TestShowAutomationInfo_AutomationLevelIf(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "cond.yaml"), []byte(`name: cond
description: Conditional automation
if: os.macos
steps:
  - bash: echo hello
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "cond", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Condition:    os.macos") {
		t.Errorf("expected Condition line, got:\n%s", out)
	}
}

func TestShowAutomationInfo_NoConditionWhenAbsent(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "simple", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(buf.String(), "Condition:") {
		t.Errorf("expected no Condition line for unconditional automation, got:\n%s", buf.String())
	}
}

func TestShowAutomationInfo_StepLevelIf(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "stepif.yaml"), []byte(`name: stepif
description: Steps with conditions
steps:
  - bash: echo macos
    if: os.macos
  - bash: echo linux
    if: os.linux
  - bash: echo always
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "stepif", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected Step details section, got:\n%s", out)
	}
	if !strings.Contains(out, "[if: os.macos]") {
		t.Errorf("expected step condition shown, got:\n%s", out)
	}
	if !strings.Contains(out, "[if: os.linux]") {
		t.Errorf("expected step condition shown, got:\n%s", out)
	}
	if !strings.Contains(out, "3. bash: echo always") {
		t.Errorf("expected unconditional step without [if:], got:\n%s", out)
	}
}

func TestShowAutomationInfo_StepDir(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "with-dir.yaml"), []byte(`name: with-dir
description: Steps with dir override
steps:
  - bash: go test ./...
    dir: src
  - bash: echo done
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-dir", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected Step details section, got:\n%s", out)
	}
	if !strings.Contains(out, "[dir: src]") {
		t.Errorf("expected dir annotation, got:\n%s", out)
	}
}

func TestShowAutomationInfo_StepTimeout(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "with-timeout.yaml"), []byte(`name: with-timeout
description: Steps with timeout
steps:
  - bash: go build ./...
    timeout: 30s
  - bash: echo done
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-timeout", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected Step details section, got:\n%s", out)
	}
	if !strings.Contains(out, "[timeout: 30s]") {
		t.Errorf("expected timeout annotation, got:\n%s", out)
	}
}

func TestShowAutomationInfo_StepDescription(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "with-desc.yaml"), []byte(`name: with-desc
description: Automation with step descriptions
steps:
  - bash: docker-compose up -d
    description: Start all containers
  - bash: sleep 2
  - python: check.py
    description: Verify services are healthy
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-desc", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected Step details section when descriptions present, got:\n%s", out)
	}
	if !strings.Contains(out, "Start all containers") {
		t.Errorf("expected step description shown, got:\n%s", out)
	}
	if !strings.Contains(out, "Verify services are healthy") {
		t.Errorf("expected step description shown, got:\n%s", out)
	}
}

func TestShowAutomationInfo_StepDescriptionWithAnnotations(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "desc-annot.yaml"), []byte(`name: desc-annot
description: Description with annotations
steps:
  - bash: go test ./...
    description: Run test suite
    dir: src
    timeout: 30s
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "desc-annot", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[dir: src; timeout: 30s]") {
		t.Errorf("expected annotations on step line, got:\n%s", out)
	}
	if !strings.Contains(out, "Run test suite") {
		t.Errorf("expected description below step line, got:\n%s", out)
	}
}

func TestShowAutomationInfo_AutomationLevelEnv(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "with-auto-env.yaml"), []byte(`description: Build for Linux
env:
  CGO_ENABLED: "0"
  GOARCH: amd64
  GOOS: linux
steps:
  - bash: go build -o bin/app ./...
  - bash: sha256sum bin/app
`), 0o644)

	var buf bytes.Buffer
	err := showAutomationInfo(root, "with-auto-env", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Env:") {
		t.Errorf("expected Env: line in output, got:\n%s", out)
	}
	if !strings.Contains(out, "CGO_ENABLED") {
		t.Errorf("expected CGO_ENABLED in env list, got:\n%s", out)
	}
	if !strings.Contains(out, "GOARCH") {
		t.Errorf("expected GOARCH in env list, got:\n%s", out)
	}
	if !strings.Contains(out, "GOOS") {
		t.Errorf("expected GOOS in env list, got:\n%s", out)
	}
}

func TestStepAnnotations_Empty(t *testing.T) {
	s := automation.Step{Type: automation.StepTypeBash, Value: "echo hi"}
	annotations := stepAnnotations(s)
	if len(annotations) != 0 {
		t.Errorf("expected no annotations for plain step, got: %v", annotations)
	}
}

func TestStepAnnotations_AllFields(t *testing.T) {
	s := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "echo hi",
		If:          "os.macos",
		Pipe:        true,
		Silent:      true,
		ParentShell: true,
		Dir:         "src",
		Timeout:     30_000_000_000,
		TimeoutRaw:  "30s",
		Env:         map[string]string{"FOO": "bar", "BAZ": "qux"},
	}
	annotations := stepAnnotations(s)
	expected := []string{
		"if: os.macos",
		"pipe",
		"silent",
		"parent_shell",
		"dir: src",
		"timeout: 30s",
		"env: BAZ, FOO",
	}
	if len(annotations) != len(expected) {
		t.Fatalf("expected %d annotations, got %d: %v", len(expected), len(annotations), annotations)
	}
	for i, want := range expected {
		if annotations[i] != want {
			t.Errorf("annotation[%d] = %q, want %q", i, annotations[i], want)
		}
	}
}

func TestStepAnnotations_SubsetFields(t *testing.T) {
	s := automation.Step{
		Type:       automation.StepTypeBash,
		Value:      "go test",
		Dir:        "backend",
		Timeout:    60_000_000_000,
		TimeoutRaw: "1m",
	}
	annotations := stepAnnotations(s)
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d: %v", len(annotations), annotations)
	}
	if annotations[0] != "dir: backend" {
		t.Errorf("expected 'dir: backend', got %q", annotations[0])
	}
	if annotations[1] != "timeout: 1m" {
		t.Errorf("expected 'timeout: 1m', got %q", annotations[1])
	}
}

func TestShowAutomationInfo_NoStepDetailsWithoutConditions(t *testing.T) {
	root := setupInfoWorkspace(t)
	var buf bytes.Buffer
	err := showAutomationInfo(root, "multi-step", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(buf.String(), "Step details:") {
		t.Errorf("expected no Step details for steps without conditions, got:\n%s", buf.String())
	}
}
