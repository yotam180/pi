package validate

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
)

func emptyResult() *discovery.Result {
	return discovery.NewResult(map[string]*automation.Automation{}, nil)
}

func newPiYamlCtx(t *testing.T, piYaml string) *Context {
	t.Helper()
	dir := t.TempDir()
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: test\nbash: echo hi\n"), 0644)
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(piYaml), 0644)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}
	disc, err := discovery.Discover(piDir, nil)
	if err != nil {
		t.Fatalf("unexpected discovery error: %v", err)
	}
	return &Context{Root: dir, Config: cfg, Discovery: disc}
}

func TestCheckPiYamlUnknownFields_ValidFile(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
shortcuts:
  hi: hello
setup:
  - hello
packages: []
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid pi.yaml, got: %v", errs)
	}
}

func TestCheckPiYamlUnknownFields_ValidWithRuntimes(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
runtimes:
  provision: auto
  manager: mise
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheckPiYamlUnknownFields_UnknownTopLevel(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nfoobarbaz: true\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "unknown field") || !strings.Contains(errs[0], "foobarbaz") {
		t.Errorf("error should mention unknown field foobarbaz, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_SuggestionShortcutz(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nshortcutz:\n  hi: hello\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "shortcuts") {
		t.Errorf("error should suggest shortcuts, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_SuggestionPakages(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\npakages: []\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "packages") {
		t.Errorf("error should suggest packages, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_SuggestionSetup(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nsetpu: []\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "setup") {
		t.Errorf("error should suggest setup, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_MultipleUnknown(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nalpha: 1\nbeta: 2\ngamma: 3\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestCheckPiYamlUnknownFields_NoSuggestionForCompletelyUnrelated(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nxyzzy: plugh\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if strings.Contains(errs[0], "did you mean") {
		t.Errorf("should not suggest for completely unrelated field, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_RuntimesUnknownField(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
runtimes:
  provision: auto
  maneger: mise
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "runtimes") || !strings.Contains(errs[0], "maneger") {
		t.Errorf("error should mention runtimes and maneger, got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "manager") {
		t.Errorf("error should suggest manager, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_RuntimesCompletelyUnknown(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
runtimes:
  provision: auto
  foobar: baz
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "foobar") {
		t.Errorf("error should mention foobar, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_MissingPiYaml(t *testing.T) {
	ctx := &Context{
		Root:      t.TempDir(),
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: emptyResult(),
	}
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("should return no errors when pi.yaml doesn't exist, got: %v", errs)
	}
}

func TestCheckPiYamlUnknownFields_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("::invalid::yaml::["), 0644)
	ctx := &Context{
		Root:      dir,
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: emptyResult(),
	}
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("should return no errors for invalid YAML (handled by config.Load), got: %v", errs)
	}
}

func TestCheckPiYamlUnknownFields_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(""), 0644)
	ctx := &Context{
		Root:      dir,
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: emptyResult(),
	}
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("should return no errors for empty file, got: %v", errs)
	}
}

func TestCheckPiYamlUnknownFields_SuggestionRuntimes(t *testing.T) {
	ctx := newPiYamlCtx(t, "project: test\nruntime: {}\n")
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "runtimes") {
		t.Errorf("error should suggest runtimes, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_SuggestionProject(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\nprojetc: test2\n"), 0644)

	cfg := &config.ProjectConfig{Project: "test"}
	disc := emptyResult()

	ctx := &Context{Root: dir, Config: cfg, Discovery: disc}
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "did you mean") || !strings.Contains(errs[0], "project") {
		t.Errorf("error should suggest project, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_MixedValidAndInvalid(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
shortcuts:
  hi: hello
unknownfield: value
setup:
  - hello
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for the unknown field, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "unknownfield") {
		t.Errorf("error should mention unknownfield, got: %s", errs[0])
	}
}

func TestCheckPiYamlUnknownFields_IntegrationWithRunner(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: test\nbash: echo hi\n"), 0644)
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\nshortcutz:\n  hi: hello\n"), 0644)

	cfg := &config.ProjectConfig{
		Project: "test",
	}
	disc, _ := discovery.Discover(piDir, nil)

	ctx := &Context{Root: dir, Config: cfg, Discovery: disc}
	runner := DefaultRunner()
	result := runner.Run(ctx)

	found := false
	for _, e := range result.Errors {
		if strings.Contains(e, "pi.yaml: unknown field") && strings.Contains(e, "shortcutz") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DefaultRunner should detect unknown pi.yaml field 'shortcutz', errors: %v", result.Errors)
	}
}

func TestKnownPiYamlKeys_Complete(t *testing.T) {
	expected := []string{"project", "shortcuts", "setup", "packages", "runtimes"}
	sort.Strings(expected)
	got := sortedKeys(knownPiYamlKeys)
	if len(got) != len(expected) {
		t.Fatalf("knownPiYamlKeys has %d entries, want %d: %v", len(got), len(expected), got)
	}
	for i, key := range expected {
		if got[i] != key {
			t.Errorf("knownPiYamlKeys[%d] = %q, want %q", i, got[i], key)
		}
	}
}

func TestKnownRuntimesKeys_Complete(t *testing.T) {
	expected := []string{"provision", "manager"}
	sort.Strings(expected)
	got := sortedKeys(knownRuntimesKeys)
	if len(got) != len(expected) {
		t.Fatalf("knownRuntimesKeys has %d entries, want %d: %v", len(got), len(expected), got)
	}
	for i, key := range expected {
		if got[i] != key {
			t.Errorf("knownRuntimesKeys[%d] = %q, want %q", i, got[i], key)
		}
	}
}

func TestCheckPiYamlUnknownFields_TopLevelAndRuntimesCombined(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
shortcutz:
  hi: hello
runtimes:
  provison: auto
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors (top-level + runtimes), got %d: %v", len(errs), errs)
	}

	hasTopLevel := false
	hasRuntimes := false
	for _, e := range errs {
		if strings.Contains(e, "shortcutz") {
			hasTopLevel = true
		}
		if strings.Contains(e, "provison") {
			hasRuntimes = true
		}
	}
	if !hasTopLevel {
		t.Error("should report unknown top-level field 'shortcutz'")
	}
	if !hasRuntimes {
		t.Error("should report unknown runtimes field 'provison'")
	}
}

func TestCheckRuntimesNodeUnknownFields_ValidNode(t *testing.T) {
	ctx := newPiYamlCtx(t, `project: test
runtimes:
  provision: auto
  manager: mise
`)
	errs := checkPiYamlUnknownFields(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid runtimes, got: %v", errs)
	}
}
