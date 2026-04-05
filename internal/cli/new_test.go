package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func TestRunNew_CreatesBasicAutomation(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "build", "", "", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "build.yaml"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "description:") {
		t.Errorf("should contain description, got: %q", content)
	}
	if !strings.Contains(content, "bash:") {
		t.Errorf("should contain bash step, got: %q", content)
	}
	if !strings.Contains(content, "hello from build") {
		t.Errorf("should contain default command, got: %q", content)
	}
}

func TestRunNew_WithBashFlag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "test", "go test ./...", "", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "test.yaml"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "bash: go test ./...") {
		t.Errorf("should contain provided bash command, got: %q", content)
	}
}

func TestRunNew_WithPythonFlag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "transform", "", "transform.py", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "transform.yaml"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "python: transform.py") {
		t.Errorf("should contain python step, got: %q", content)
	}
}

func TestRunNew_WithDescription(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "build", "make build", "", "Build the project", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "build.yaml"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "description: Build the project") {
		t.Errorf("should contain provided description, got: %q", content)
	}
}

func TestRunNew_NestedPath(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "setup/install-deps", "npm install", "", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "setup", "install-deps.yaml"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "bash: npm install") {
		t.Errorf("should contain bash command, got: %q", content)
	}
}

func TestRunNew_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	writeTestFile(t, dir, ".pi/build.yaml", "description: existing\nbash: make\n")

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "build", "", "", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %q", err.Error())
	}
}

func TestRunNew_NoProject(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "build", "", "", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when no project exists")
	}
	if !strings.Contains(err.Error(), "pi init") {
		t.Errorf("error should suggest 'pi init', got: %q", err.Error())
	}
}

func TestRunNew_StripYamlExtension(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "build.yaml", "make", "", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".pi", "build.yaml")); err != nil {
		t.Error("should create build.yaml, not build.yaml.yaml")
	}
}

func TestRunNew_OutputShowsNextSteps(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "deploy", "", "", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Created") {
		t.Errorf("should show 'Created', got: %q", output)
	}
	if !strings.Contains(output, "pi run deploy") {
		t.Errorf("should suggest 'pi run deploy', got: %q", output)
	}
	if !strings.Contains(output, "pi info deploy") {
		t.Errorf("should suggest 'pi info deploy', got: %q", output)
	}
}

func TestRunNew_CreatesValidAutomation(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, config.FileName, "project: test\n")
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	var stdout, stderr bytes.Buffer
	err := runNew(dir, "check", "go vet ./...", "", "Run go vet", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".pi", "check.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	got := string(content)
	want := "description: Run go vet\nbash: go vet ./...\n"
	if got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestGenerateAutomationYAML_DefaultCommand(t *testing.T) {
	got := generateAutomationYAML("my-tool", "", "", "")
	if !strings.Contains(got, "description: TODO") {
		t.Errorf("should contain TODO in description, got: %q", got)
	}
	if !strings.Contains(got, `bash: echo "hello from my-tool"`) {
		t.Errorf("should contain default bash command, got: %q", got)
	}
}

func TestGenerateAutomationYAML_DescriptionWithColon(t *testing.T) {
	got := generateAutomationYAML("test", "go test", "", "Run tests: unit and integration")
	if !strings.Contains(got, `description: "Run tests: unit and integration"`) {
		t.Errorf("description with colon should be quoted, got: %q", got)
	}
}

func TestGenerateAutomationYAML_NestedName(t *testing.T) {
	got := generateAutomationYAML("docker/logs", "", "", "")
	if !strings.Contains(got, `bash: echo "hello from logs"`) {
		t.Errorf("default command should use base name, got: %q", got)
	}
}

func TestInitProject_CreatesExampleAutomation(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	err := initProject(dir, "test-proj", &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".pi", "hello.yaml"))
	if err != nil {
		t.Fatalf("hello.yaml not created: %v", err)
	}

	content := string(data)
	if content != ExampleAutomationContent {
		t.Errorf("hello.yaml content = %q, want %q", content, ExampleAutomationContent)
	}

	output := stdout.String()
	if !strings.Contains(output, "hello.yaml") {
		t.Errorf("output should mention hello.yaml, got: %q", output)
	}
}

func TestInitProject_NextStepsMentionNew(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	err := initProject(dir, "test-proj", &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "pi run hello") {
		t.Errorf("next steps should mention 'pi run hello', got: %q", output)
	}
	if !strings.Contains(output, "pi new") {
		t.Errorf("next steps should mention 'pi new', got: %q", output)
	}
}
