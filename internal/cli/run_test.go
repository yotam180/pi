package cli

import (
	"errors"
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
	var exitErr *executor.ExitError
	if !errors.As(err, &exitErr) {
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

func TestRunAutomation_PIArgsAvailable(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build with extra args
bash: echo "args=$PI_ARGS"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "build", []string{"--release", "--verbose"}, nil, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "args=--release --verbose" {
		t.Errorf("output = %q, want %q", got, "args=--release --verbose")
	}
}

func TestRunAutomation_ExcessPositionalArgs(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`description: Greet with input
inputs:
  name:
    type: string
    required: true
steps:
  - bash: echo "hello $PI_INPUT_NAME"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "greet", []string{"alice", "--extra", "stuff"}, nil, false, false, &buf, os.Stderr)
	if err == nil {
		t.Fatal("expected error for excess positional args")
	}
	if !strings.Contains(err.Error(), "too many arguments") {
		t.Errorf("expected 'too many arguments' in error, got: %v", err)
	}
}

func TestRunAutomation_PositionalInputsMapByOrder(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "deploy.yaml"), []byte(`description: Deploy to environment
inputs:
  env:
    type: string
    required: true
  region:
    type: string
    default: us-east-1
steps:
  - bash: echo "$PI_IN_ENV $PI_IN_REGION"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "deploy", []string{"prod", "eu-west-1"}, nil, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "prod eu-west-1" {
		t.Errorf("output = %q, want %q", got, "prod eu-west-1")
	}
}

func TestRunAutomation_PositionalWithDefaults(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "deploy.yaml"), []byte(`description: Deploy to environment
inputs:
  env:
    type: string
    required: true
  region:
    type: string
    default: us-east-1
steps:
  - bash: echo "$PI_IN_ENV $PI_IN_REGION"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "deploy", []string{"staging"}, nil, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "staging us-east-1" {
		t.Errorf("output = %q, want %q", got, "staging us-east-1")
	}
}

func TestRunAutomation_FlagLikeArgsForwardedAsPI_ARGS(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build with extra args
bash: echo "args=$PI_ARGS"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "build", []string{"--release", "--target=linux"}, nil, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "args=--release --target=linux" {
		t.Errorf("output = %q, want %q", got, "args=--release --target=linux")
	}
}

func TestRunAutomation_FlagLikeArgsAsPositionalInputs(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
inputs:
  profile:
    type: string
    default: dev
steps:
  - bash: echo "profile=$PI_IN_PROFILE"
`), 0o644)

	var buf strings.Builder
	err := runAutomation(root, "build", []string{"release"}, nil, false, false, &buf, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "profile=release" {
		t.Errorf("output = %q, want %q", got, "profile=release")
	}
}

func TestRunCmd_FlagLikeArgsAfterAutomation_NotParsedAsPiFlags(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
bash: echo "args=$PI_ARGS"
`), 0o644)

	var stdout strings.Builder
	err := runAutomation(root, "build", []string{"--silent", "--loud"}, nil, false, false, &stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "args=--silent --loud" {
		t.Errorf("expected flag-like args forwarded via PI_ARGS, got: %q", got)
	}
}

func TestRunCmd_CobraDoesNotParsePostNameFlags(t *testing.T) {
	cmd := newRunCmd()
	cmd.SetArgs([]string{"--repo", "/nonexistent", "build", "--silent", "--loud"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error (nonexistent repo), but no error means flags were parsed correctly")
	}
	if strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("--silent/--loud after automation name should not be parsed as pi flags: %v", err)
	}
}

func TestRunCmd_PiFlagsBeforeAutomation(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
bash: echo "running"
`), 0o644)

	err := runAutomation(root, "build", nil, nil, true, false, os.Stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCmd_DoubleDashStillWorks(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte(`description: Build
bash: echo "args=$PI_ARGS"
`), 0o644)

	var stdout strings.Builder
	err := runAutomation(root, "build", []string{"--release"}, nil, false, false, &stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "args=--release" {
		t.Errorf("expected args forwarded via PI_ARGS, got: %q", got)
	}
}

func TestRunCmd_WithFlagBeforeAutomation(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(root, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(piDir, "greet.yaml"), []byte(`description: Greet
inputs:
  name:
    type: string
    required: true
steps:
  - bash: echo "hello $PI_IN_NAME"
`), 0o644)

	var stdout strings.Builder
	err := runAutomation(root, "greet", nil, map[string]string{"name": "world"}, false, false, &stdout, os.Stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "hello world" {
		t.Errorf("expected --with to pass input, got: %q", got)
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
