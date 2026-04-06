package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestDoctorHelp(t *testing.T) {
	out, err := executeCmd("doctor", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "requirement") {
		t.Errorf("expected help to mention requirements, got: %s", out)
	}
	if !strings.Contains(out, "doctor") {
		t.Errorf("expected help to mention doctor, got: %s", out)
	}
}

func TestDoctorInRootHelp(t *testing.T) {
	out, err := executeCmd("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "doctor") {
		t.Errorf("expected root help to list doctor subcommand, got: %s", out)
	}
}

func TestDoctor_NoLocalAutomations(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: empty\n"), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Built-in automations are always present but have no requires:, so
	// doctor should report no requirements to check.
	if !strings.Contains(buf.String(), "No automations have requirements") {
		t.Errorf("expected 'No automations have requirements', got: %s", buf.String())
	}
}

func TestDoctor_NoRequirements(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: A simple automation
steps:
  - bash: echo hello
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No automations have requirements") {
		t.Errorf("expected 'No automations have requirements', got: %s", buf.String())
	}
}

func TestDoctor_SatisfiedRequirement(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "needs-bash.yaml"), []byte(`name: needs-bash
description: Needs bash
requires:
  - command: bash
steps:
  - bash: echo ok
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "needs-bash") {
		t.Errorf("expected automation name in output, got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("expected check mark in output, got: %s", output)
	}
	if !strings.Contains(output, "command: bash") {
		t.Errorf("expected 'command: bash' label, got: %s", output)
	}
}

func TestDoctor_MissingRequirement(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "needs-missing.yaml"), []byte(`name: needs-missing
description: Needs impossible tool
requires:
  - command: pi-nonexistent-tool-xyz
steps:
  - bash: echo never
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err == nil {
		t.Fatal("expected error for missing requirement")
	}
	output := buf.String()
	if !strings.Contains(output, "needs-missing") {
		t.Errorf("expected automation name, got: %s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected ✗ for missing requirement, got: %s", output)
	}
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' message, got: %s", output)
	}
}

func TestDoctor_ExitCode1OnMissing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "bad.yaml"), []byte(`name: bad
description: Missing tool
requires:
  - command: pi-nonexistent-tool-xyz
steps:
  - bash: echo never
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDoctor_MixedSatisfiedAndMissing(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)

	os.WriteFile(filepath.Join(piDir, "good.yaml"), []byte(`name: good
description: All satisfied
requires:
  - command: bash
steps:
  - bash: echo ok
`), 0644)

	os.WriteFile(filepath.Join(piDir, "bad.yaml"), []byte(`name: bad
description: Missing tool
requires:
  - command: pi-nonexistent-tool-xyz
steps:
  - bash: echo never
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err == nil {
		t.Fatal("expected error when any requirement is missing")
	}

	output := buf.String()
	if !strings.Contains(output, "good") {
		t.Errorf("expected 'good' automation listed, got: %s", output)
	}
	if !strings.Contains(output, "bad") {
		t.Errorf("expected 'bad' automation listed, got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("expected ✓ for satisfied, got: %s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected ✗ for missing, got: %s", output)
	}
}

func TestFormatDoctorLabel(t *testing.T) {
	tests := []struct {
		name string
		req  automation.Requirement
		want string
	}{
		{
			name: "command without version",
			req:  automation.Requirement{Kind: automation.RequirementCommand, Name: "jq"},
			want: "command: jq",
		},
		{
			name: "command with version",
			req:  automation.Requirement{Kind: automation.RequirementCommand, Name: "curl", MinVersion: "7.0"},
			want: "command: curl >= 7.0",
		},
		{
			name: "runtime without version",
			req:  automation.Requirement{Kind: automation.RequirementRuntime, Name: "python"},
			want: "python",
		},
		{
			name: "runtime with version",
			req:  automation.Requirement{Kind: automation.RequirementRuntime, Name: "node", MinVersion: "18"},
			want: "node >= 18",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDoctorLabel(tt.req)
			if got != tt.want {
				t.Errorf("formatDoctorLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDoctor_CommandWithVersion(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "needs-bash-ver.yaml"), []byte("description: Needs bash with version\nrequires:\n  - command: bash >= 1.0\nsteps:\n  - bash: echo ok\n"), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "command: bash >= 1.0") {
		t.Errorf("expected 'command: bash >= 1.0' label, got: %s", output)
	}
}

func TestDoctor_RuntimeRequirementLabel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "needs-go.yaml"), []byte("description: Needs go\nrequires:\n  - go >= 1.20\nsteps:\n  - bash: echo ok\n"), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "go >= 1.20") {
		t.Errorf("expected 'go >= 1.20' label, got: %s", output)
	}
	if strings.Contains(output, "command: go") {
		t.Errorf("runtime requirement should not have 'command:' prefix, got: %s", output)
	}
}

func TestDoctor_SkipsAutomationsWithoutRequires(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)

	os.WriteFile(filepath.Join(piDir, "no-req.yaml"), []byte(`name: no-req
description: No requirements
steps:
  - bash: echo hello
`), 0644)

	os.WriteFile(filepath.Join(piDir, "has-req.yaml"), []byte(`name: has-req
description: Has requirements
requires:
  - command: bash
steps:
  - bash: echo hello
`), 0644)

	var buf bytes.Buffer
	err := runDoctor(dir, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "no-req") {
		t.Errorf("should not show automation without requires, got: %s", output)
	}
	if !strings.Contains(output, "has-req") {
		t.Errorf("expected 'has-req' automation listed, got: %s", output)
	}
}
