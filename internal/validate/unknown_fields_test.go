package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/suggest"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// --- checkUnknownFields tests ---

func TestCheck_UnknownFields_ValidFile(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Run the tests
bash: go test ./...
`
	yamlPath := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"test": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid file, got: %v", errs)
	}
}

func TestCheck_UnknownFields_ValidMultiStep(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Build and deploy
env:
  GOOS: linux
steps:
  - bash: go build ./...
    env:
      CGO_ENABLED: "0"
    dir: services/api
    timeout: 30s
    description: Build the binary
  - run: deploy/push
    with:
      env: prod
    if: command.docker
    silent: true
`
	yamlPath := filepath.Join(dir, "build.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"build": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid multi-step file, got: %v", errs)
	}
}

func TestCheck_UnknownFields_UnknownTopLevel(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: Run the tests
bash: go test ./...
`
	yamlPath := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"test": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "descrption") {
		t.Errorf("expected error to mention 'descrption', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "unknown field") {
		t.Errorf("expected error to mention 'unknown field', got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_TopLevelSuggestion(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: Run the tests
bash: go test ./...
`
	yamlPath := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"test": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") {
		t.Errorf("expected suggestion, got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "description") {
		t.Errorf("expected 'description' suggestion, got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_UnknownStepLevel(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Build
steps:
  - bash: go build ./...
    descrption: Build the binary
`
	yamlPath := filepath.Join(dir, "build.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"build": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "step[0]") {
		t.Errorf("expected error to mention 'step[0]', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "descrption") {
		t.Errorf("expected error to mention 'descrption', got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_MultipleUnknowns(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: Build
nmae: wrong
bash: go build ./...
`
	yamlPath := filepath.Join(dir, "build.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"build": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestCheck_UnknownFields_SkipsBuiltins(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: typo
bash: echo hello
`
	yamlPath := filepath.Join(dir, "hello.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	builtinAutos := map[string]*automation.Automation{
		"pi:hello": {Name: "pi:hello", FilePath: yamlPath},
	}
	builtinResult := discovery.NewResult(builtinAutos, []string{"pi:hello"})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	disc.MergeBuiltins(builtinResult)

	ctx := &Context{
		Root:      dir,
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: disc,
	}

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for builtins, got: %v", errs)
	}
}

func TestCheck_UnknownFields_InstallerFields(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Install Go
install:
  test: command -v go
  run: brew install go
  version: go version | awk '{print $3}'
`
	yamlPath := filepath.Join(dir, "install-go.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"install-go": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid installer, got: %v", errs)
	}
}

func TestCheck_UnknownFields_UnknownInstallField(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Install Go
install:
  test: command -v go
  run: brew install go
  vresion: go version
`
	yamlPath := filepath.Join(dir, "install-go.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"install-go": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "install:") {
		t.Errorf("expected error to mention 'install:', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "vresion") {
		t.Errorf("expected error to mention 'vresion', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "version") {
		t.Errorf("expected 'version' suggestion, got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_FirstBlockSubSteps(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Install tool
steps:
  - first:
      - bash: mise install go
        iff: command.mise
      - bash: brew install go
`
	yamlPath := filepath.Join(dir, "install.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"install": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "first[0]") {
		t.Errorf("expected error to mention 'first[0]', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "iff") {
		t.Errorf("expected error to mention 'iff', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "if") {
		t.Errorf("expected 'if' suggestion, got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_MissingFile(t *testing.T) {
	dir := t.TempDir()
	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"missing": {filePath: filepath.Join(dir, "nonexistent.yaml")},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for missing file, got: %v", errs)
	}
}

func TestCheck_UnknownFields_EmptyFilePath(t *testing.T) {
	dir := t.TempDir()
	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"nopath": {filePath: ""},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty file path, got: %v", errs)
	}
}

func TestCheck_UnknownFields_ShorthandWithAllModifiers(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Cross-compile
bash: go build -o bin/app ./...
env:
  GOOS: linux
dir: services/api
timeout: 30s
silent: true
if: os.linux
`
	yamlPath := filepath.Join(dir, "build.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"build": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid shorthand, got: %v", errs)
	}
}

func TestCheck_UnknownFields_StepCommonTypo(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Build
steps:
  - bash: echo hello
    timout: 30s
`
	yamlPath := filepath.Join(dir, "build.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"build": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "timout") {
		t.Errorf("expected error to mention 'timout', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "timeout") {
		t.Errorf("expected 'timeout' suggestion, got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_AutomationNameInError(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: typo
bash: echo hello
`
	yamlPath := filepath.Join(dir, "docker", "up.yaml")
	os.MkdirAll(filepath.Join(dir, "docker"), 0755)
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"docker/up": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "docker/up") {
		t.Errorf("expected error to mention automation name 'docker/up', got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_ValidInputsField(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Deploy
inputs:
  env:
    type: string
    required: true
bash: deploy.sh $PI_IN_ENV
`
	yamlPath := filepath.Join(dir, "deploy.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"deploy": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid inputs, got: %v", errs)
	}
}

func TestCheck_UnknownFields_ValidRequiresField(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Test
requires:
  - python >= 3.9
bash: python3 -m pytest
`
	yamlPath := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"test": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid requires, got: %v", errs)
	}
}

func TestCheck_UnknownFields_InstallerWithVerify(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Install tool
install:
  test: command -v mytool
  run: curl -fsSL https://example.com/install.sh | sh
  verify: mytool --version
  version: mytool --version
`
	yamlPath := filepath.Join(dir, "install-tool.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"install-tool": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid installer with verify, got: %v", errs)
	}
}

func TestCheck_UnknownFields_ParentShellAndPipe(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Activate venv
steps:
  - bash: source venv/bin/activate
    parent_shell: true
  - bash: echo data
    pipe: true
  - python: process.py
`
	yamlPath := filepath.Join(dir, "activate.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"activate": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_UnknownFields_PipeTo(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Legacy pipe
steps:
  - bash: echo data
    pipe_to: next
  - bash: cat
`
	yamlPath := filepath.Join(dir, "legacy.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"legacy": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for pipe_to (deprecated but known), got: %v", errs)
	}
}

func TestCheck_UnknownFields_CompletelyUnknownNoSuggestion(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `description: Test
bash: echo hello
zzzzxxx: something
`
	yamlPath := filepath.Join(dir, "test.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"test": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "zzzzxxx") {
		t.Errorf("expected error to mention 'zzzzxxx', got: %s", errs[0])
	}
	if strings.Contains(errs[0], "did you mean") {
		t.Errorf("expected no suggestion for completely unrelated field, got: %s", errs[0])
	}
}

func TestCheck_UnknownFields_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `{{{invalid yaml`
	yamlPath := filepath.Join(dir, "broken.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"broken": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unparseable YAML (other checks catch this), got: %v", errs)
	}
}

func TestCheck_UnknownFields_StepAndTopLevelMixed(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `descrption: typo at top
steps:
  - bash: echo hello
    slent: true
`
	yamlPath := filepath.Join(dir, "mixed.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	ctx := newUnknownFieldsContext(t, dir, map[string]automationEntry{
		"mixed": {filePath: yamlPath},
	})

	errs := checkUnknownFields(ctx)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}

	found := map[string]bool{}
	for _, err := range errs {
		if strings.Contains(err, "descrption") {
			found["top"] = true
		}
		if strings.Contains(err, "slent") {
			found["step"] = true
		}
	}
	if !found["top"] {
		t.Error("expected error for top-level 'descrption'")
	}
	if !found["step"] {
		t.Error("expected error for step-level 'slent'")
	}
}

// --- suggestField unit tests ---

func TestSuggestFieldName_ExactMatch(t *testing.T) {
	known := map[string]bool{"description": true, "bash": true}
	got := suggestFieldName("description", known)
	if got != "" {
		t.Errorf("exact match should return empty, got %q", got)
	}
}

func TestSuggestFieldName_CloseMatch(t *testing.T) {
	known := map[string]bool{"description": true, "bash": true}
	got := suggestFieldName("descrption", known)
	if got != "description" {
		t.Errorf("expected 'description', got %q", got)
	}
}

func TestSuggestFieldName_NoMatch(t *testing.T) {
	known := map[string]bool{"description": true, "bash": true}
	got := suggestFieldName("xyzabc", known)
	if got != "" {
		t.Errorf("expected empty for no match, got %q", got)
	}
}

func TestSuggestFieldName_MultipleCandidates(t *testing.T) {
	known := map[string]bool{"silent": true, "slent": true}
	got := suggestFieldName("sient", known)
	if got == "" {
		t.Error("expected a suggestion")
	}
}

func TestSuggestFieldName_ShortField(t *testing.T) {
	known := map[string]bool{"if": true, "env": true, "dir": true}
	got := suggestFieldName("iff", known)
	if got != "if" {
		t.Errorf("expected 'if', got %q", got)
	}
}

func TestLevenshtein_Identical(t *testing.T) {
	if d := suggest.Levenshtein("hello", "hello"); d != 0 {
		t.Errorf("expected 0, got %d", d)
	}
}

func TestLevenshtein_Empty(t *testing.T) {
	if d := suggest.Levenshtein("", "hello"); d != 5 {
		t.Errorf("expected 5, got %d", d)
	}
	if d := suggest.Levenshtein("hello", ""); d != 5 {
		t.Errorf("expected 5, got %d", d)
	}
}

func TestLevenshtein_OneDiff(t *testing.T) {
	if d := suggest.Levenshtein("hello", "helo"); d != 1 {
		t.Errorf("expected 1, got %d", d)
	}
}

func TestLevenshtein_Substitution(t *testing.T) {
	if d := suggest.Levenshtein("cat", "bat"); d != 1 {
		t.Errorf("expected 1, got %d", d)
	}
}

// --- helper ---

type automationEntry struct {
	filePath string
}

func newUnknownFieldsContext(t *testing.T, root string, entries map[string]automationEntry) *Context {
	t.Helper()
	autos := make(map[string]*automation.Automation, len(entries))
	names := make([]string, 0, len(entries))
	for name, entry := range entries {
		autos[name] = &automation.Automation{
			Name:     name,
			FilePath: entry.filePath,
		}
		names = append(names, name)
	}
	disc := discovery.NewResult(autos, names)
	return &Context{
		Root:      root,
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: disc,
	}
}
