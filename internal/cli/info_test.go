package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(out, "service (string, required)") {
		t.Errorf("expected required input, got:\n%s", out)
	}
	if !strings.Contains(out, "The service to target") {
		t.Errorf("expected input description, got:\n%s", out)
	}
	if !strings.Contains(out, `tail (string, optional, default: "200")`) {
		t.Errorf("expected optional input with default, got:\n%s", out)
	}
	if !strings.Contains(out, "verbose (string, optional)") {
		t.Errorf("expected optional input without default, got:\n%s", out)
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
