package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/executor"
)

func setupRunWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)

	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(filepath.Join(piDir, "docker"), 0o755)

	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`name: hello
description: Say hello
steps:
  - bash: echo hello world
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "docker", "up.yaml"), []byte(`name: docker/up
description: Start containers
steps:
  - bash: echo docker is up
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "fail.yaml"), []byte(`name: fail
description: Always fails
steps:
  - bash: exit 42
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "args.yaml"), []byte(`name: args
description: Echo args
steps:
  - bash: echo "got $1 $2"
`), 0o644)

	os.WriteFile(filepath.Join(piDir, "chain.yaml"), []byte(`name: chain
description: Chain to hello
steps:
  - run: hello
`), 0o644)

	return root
}

func TestRunAutomation_Success(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "hello", nil, nil, false, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAutomation_NestedName(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "docker/up", nil, nil, false, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAutomation_NotFound(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "nonexistent", nil, nil, false, false, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for unknown automation")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "hello") {
		t.Errorf("expected error to list available automations, got: %v", err)
	}
}

func TestRunAutomation_ExitCode(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "fail", nil, nil, false, false, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error for failed step")
	}
	exitErr, ok := err.(*executor.ExitError)
	if !ok {
		t.Fatalf("expected *executor.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 42 {
		t.Errorf("expected exit code 42, got %d", exitErr.Code)
	}
}

func TestRunAutomation_WithArgs(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "args", []string{"foo", "bar"}, nil, false, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAutomation_RunStep(t *testing.T) {
	root := setupRunWorkspace(t)
	err := runAutomation(root, "chain", nil, nil, false, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAutomation_FromSubdirectory(t *testing.T) {
	root := setupRunWorkspace(t)
	sub := filepath.Join(root, "src", "deep")
	os.MkdirAll(sub, 0o755)

	err := runAutomation(sub, "hello", nil, nil, false, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAutomation_NoPiYaml(t *testing.T) {
	dir := t.TempDir()
	err := runAutomation(dir, "hello", nil, nil, false, false, os.Stdout, os.Stderr)
	if err == nil {
		t.Fatal("expected error when no pi.yaml found")
	}
	if !strings.Contains(err.Error(), "pi.yaml") {
		t.Errorf("expected error to mention pi.yaml, got: %v", err)
	}
}

func TestRunAutomation_WithInputs(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`name: greet
description: Greet with input
inputs:
  name:
    type: string
    required: true
steps:
  - bash: echo "hello $PI_INPUT_NAME"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "greet", nil, map[string]string{"name": "alice"}, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "hello alice" {
		t.Errorf("output = %q, want %q", got, "hello alice")
	}
}

func TestParseWithFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    map[string]string
		wantErr bool
	}{
		{"empty", nil, nil, false},
		{"single", []string{"key=value"}, map[string]string{"key": "value"}, false},
		{"multiple", []string{"a=1", "b=2"}, map[string]string{"a": "1", "b": "2"}, false},
		{"value with equals", []string{"k=v=w"}, map[string]string{"k": "v=w"}, false},
		{"no equals", []string{"bad"}, nil, true},
		{"empty key", []string{"=value"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWithFlags(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseWithFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("got[%s] = %q, want %q", k, got[k], v)
					}
				}
			}
		})
	}
}
