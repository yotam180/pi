package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func TestRunInit_CreatesProjectFiles(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInit(dir, "test-project", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("pi.yaml not created: %v", err)
	}
	if string(data) != "project: test-project\n" {
		t.Errorf("pi.yaml content = %q, want %q", string(data), "project: test-project\n")
	}

	info, err := os.Stat(filepath.Join(dir, ".pi"))
	if err != nil {
		t.Fatalf(".pi/ not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf(".pi should be a directory")
	}

	if !strings.Contains(stdout.String(), "Initialized project") {
		t.Errorf("stdout should contain 'Initialized project', got: %q", stdout.String())
	}
}

func TestRunInit_NameFlag(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInit(dir, "custom-name", false, strings.NewReader("\n"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("pi.yaml not created: %v", err)
	}
	if !strings.Contains(string(data), "custom-name") {
		t.Errorf("pi.yaml should contain 'custom-name', got: %q", string(data))
	}
}

func TestRunInit_AlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: existing\n")

	var stdout, stderr bytes.Buffer
	err := runInit(dir, "", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Already initialized") {
		t.Errorf("stdout should contain 'Already initialized', got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "existing") {
		t.Errorf("stdout should mention project name 'existing', got: %q", stdout.String())
	}
}

func TestRunInit_AlreadyInitialized_ExitSuccess(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: myapp\n")

	var stdout, stderr bytes.Buffer
	err := runInit(dir, "", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("should exit successfully, got error: %v", err)
	}
}

func TestRunInit_PiDirExistsButNoPiYaml(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".pi"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := runInit(dir, "test-proj", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, config.FileName))
	if err != nil {
		t.Fatalf("pi.yaml not created: %v", err)
	}
	if !strings.Contains(string(data), "test-proj") {
		t.Errorf("pi.yaml should contain 'test-proj', got: %q", string(data))
	}
}

func TestRunInit_YesFlag_InfersFromDirName(t *testing.T) {
	dir := t.TempDir()
	// TempDir names are random, but we can still verify the flow works
	var stdout, stderr bytes.Buffer
	err := runInit(dir, "", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Initialized project") {
		t.Errorf("stdout should contain 'Initialized project', got: %q", stdout.String())
	}

	if _, err := os.Stat(filepath.Join(dir, config.FileName)); err != nil {
		t.Errorf("pi.yaml should exist after --yes init")
	}
}

func TestPromptProjectName_AcceptDefault(t *testing.T) {
	stdin := strings.NewReader("\n")
	var stdout bytes.Buffer

	name, err := promptProjectName("my-default", stdin, &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-default" {
		t.Errorf("name = %q, want %q", name, "my-default")
	}
	if !strings.Contains(stdout.String(), "Project name [my-default]") {
		t.Errorf("prompt should show default, got: %q", stdout.String())
	}
}

func TestPromptProjectName_CustomName(t *testing.T) {
	stdin := strings.NewReader("my-custom-app\n")
	var stdout bytes.Buffer

	name, err := promptProjectName("default-name", stdin, &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-custom-app" {
		t.Errorf("name = %q, want %q", name, "my-custom-app")
	}
}

func TestRunInit_NonInteractivePipedInput(t *testing.T) {
	dir := t.TempDir()
	stdin := strings.NewReader("ignored\n")

	var stdout, stderr bytes.Buffer
	err := runInit(dir, "", false, stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Piped input (not a terminal) should behave like --yes
	if !strings.Contains(stdout.String(), "Initialized project") {
		t.Errorf("stdout should contain 'Initialized project', got: %q", stdout.String())
	}
}

func TestRunInit_NextStepsShown(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runInit(dir, "proj", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Next steps:") {
		t.Errorf("should show 'Next steps:', got: %q", output)
	}
	if !strings.Contains(output, "pi setup add") {
		t.Errorf("should mention 'pi setup add', got: %q", output)
	}
	if !strings.Contains(output, "pi shell") {
		t.Errorf("should mention 'pi shell', got: %q", output)
	}
	if !strings.Contains(output, "pi run") {
		t.Errorf("should mention 'pi run', got: %q", output)
	}
}

func TestRunInit_AlreadyInitialized_ShowsNextSteps(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: existing\n")

	var stdout, stderr bytes.Buffer
	err := runInit(dir, "", true, nil, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Next steps:") {
		t.Errorf("already-initialized should show 'Next steps:', got: %q", output)
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MyProject", "myproject"},
		{"my_project", "my-project"},
		{"my project", "my-project"},
		{"My Cool App", "my-cool-app"},
		{"UPPER_CASE", "upper-case"},
		{"already-kebab", "already-kebab"},
		{"with.dots", "withdots"},
		{"  spaces  ", "spaces"},
		{"foo--bar", "foo--bar"},
		{"123-numbers", "123-numbers"},
	}

	for _, tt := range tests {
		got := toKebabCase(tt.input)
		if got != tt.want {
			t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInitProject_Reusable(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	err := initProject(dir, "reuse-test", &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("config load error: %v", err)
	}
	if cfg.Project != "reuse-test" {
		t.Errorf("project = %q, want %q", cfg.Project, "reuse-test")
	}
}
