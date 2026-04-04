package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating directories for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

const validAutomation = `name: %s
description: %s
steps:
  - bash: echo hello
`

func makeAutomation(name, desc string) string {
	return strings.Replace(
		strings.Replace(validAutomation, "%s", name, 1),
		"%s", desc, 1)
}

func TestDiscover_EmptyPiDir(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Automations) != 0 {
		t.Errorf("expected 0 automations, got %d", len(result.Automations))
	}
}

func TestDiscover_NonExistentPiDir(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Automations) != 0 {
		t.Errorf("expected 0 automations, got %d", len(result.Automations))
	}
}

func TestDiscover_FlatYAML(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))
	writeFile(t, filepath.Join(piDir, "docker", "down.yaml"),
		makeAutomation("docker/down", "Stop containers"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Automations) != 2 {
		t.Fatalf("expected 2 automations, got %d", len(result.Automations))
	}

	if _, ok := result.Automations["docker/up"]; !ok {
		t.Error("missing automation 'docker/up'")
	}
	if _, ok := result.Automations["docker/down"]; !ok {
		t.Error("missing automation 'docker/down'")
	}
}

func TestDiscover_DirectoryAutomation(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "setup", "cursor", "automation.yaml"),
		makeAutomation("setup/cursor", "Install cursor extensions"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Automations) != 1 {
		t.Fatalf("expected 1 automation, got %d", len(result.Automations))
	}
	if _, ok := result.Automations["setup/cursor"]; !ok {
		t.Error("missing automation 'setup/cursor'")
	}
}

func TestDiscover_MixedFormats(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))
	writeFile(t, filepath.Join(piDir, "setup", "cursor", "automation.yaml"),
		makeAutomation("setup/cursor", "Install cursor extensions"))
	writeFile(t, filepath.Join(piDir, "build.yaml"),
		makeAutomation("build", "Build project"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Automations) != 3 {
		t.Fatalf("expected 3 automations, got %d", len(result.Automations))
	}

	expected := []string{"build", "docker/up", "setup/cursor"}
	names := result.Names()
	if len(names) != len(expected) {
		t.Fatalf("expected names %v, got %v", expected, names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d]: expected %q, got %q", i, name, names[i])
		}
	}
}

func TestDiscover_NameCollision(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	// Both resolve to "docker/up"
	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Flat form"))
	writeFile(t, filepath.Join(piDir, "docker", "up", "automation.yaml"),
		makeAutomation("docker/up", "Dir form"))

	_, err := Discover(piDir)
	if err == nil {
		t.Fatal("expected name collision error, got nil")
	}
	if !strings.Contains(err.Error(), "collision") {
		t.Errorf("expected collision error, got: %v", err)
	}
}

func TestDiscover_SkipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))
	writeFile(t, filepath.Join(piDir, "docker", "helper.sh"),
		"#!/bin/bash\necho hi")
	writeFile(t, filepath.Join(piDir, "docker", "helper.py"),
		"print('hi')")
	writeFile(t, filepath.Join(piDir, "notes.txt"),
		"just notes")

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Automations) != 1 {
		t.Errorf("expected 1 automation (should skip non-yaml), got %d", len(result.Automations))
	}
}

func TestDiscover_NameNormalization(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "Docker", "Up.yaml"),
		makeAutomation("Docker/Up", "Start containers"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name should be lowercased
	if _, ok := result.Automations["docker/up"]; !ok {
		t.Errorf("expected normalized name 'docker/up', got keys: %v", result.Names())
	}
}

func TestFind_ExistingAutomation(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a, err := result.Find("docker/up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "docker/up" {
		t.Errorf("expected name 'docker/up', got %q", a.Name)
	}
}

func TestFind_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a, err := result.Find("Docker/Up")
	if err != nil {
		t.Fatalf("unexpected error looking up 'Docker/Up': %v", err)
	}
	if a == nil {
		t.Fatal("expected automation, got nil")
	}
}

func TestFind_NotFound_WithAvailable(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))
	writeFile(t, filepath.Join(piDir, "build.yaml"),
		makeAutomation("build", "Build it"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = result.Find("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent automation")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "not found") {
		t.Errorf("error should mention 'not found': %v", err)
	}
	if !strings.Contains(errStr, "docker/up") {
		t.Errorf("error should list available automation 'docker/up': %v", err)
	}
	if !strings.Contains(errStr, "build") {
		t.Errorf("error should list available automation 'build': %v", err)
	}
}

func TestFind_NotFound_Empty(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = result.Find("anything")
	if err == nil {
		t.Fatal("expected error for nonexistent automation")
	}
	if !strings.Contains(err.Error(), "no automations discovered") {
		t.Errorf("error should mention no automations: %v", err)
	}
}

func TestFind_TrimsSlashes(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "docker", "up.yaml"),
		makeAutomation("docker/up", "Start containers"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a, err := result.Find("/docker/up/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected automation, got nil")
	}
}

func TestDiscover_DeeplyNested(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "infra", "k8s", "deploy.yaml"),
		makeAutomation("infra/k8s/deploy", "Deploy to k8s"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result.Automations["infra/k8s/deploy"]; !ok {
		t.Errorf("expected 'infra/k8s/deploy', got: %v", result.Names())
	}
}

func TestDiscover_InvalidAutomationFile(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "bad.yaml"), "not: valid: yaml: [")

	_, err := Discover(piDir)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestNames_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "a.yaml"),
		makeAutomation("a", "A"))
	writeFile(t, filepath.Join(piDir, "b.yaml"),
		makeAutomation("b", "B"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatal(err)
	}

	names := result.Names()
	names[0] = "mutated"

	original := result.Names()
	if original[0] == "mutated" {
		t.Error("Names() should return a copy, not the internal slice")
	}
}

func TestDiscover_RootLevelAutomation(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	writeFile(t, filepath.Join(piDir, "build.yaml"),
		makeAutomation("build", "Build the project"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result.Automations["build"]; !ok {
		t.Error("missing root-level automation 'build'")
	}
}

func TestMergeBuiltins_AddsBuiltins(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "local.yaml"),
		makeAutomation("local", "Local automation"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builtinAutomations := map[string]*automation.Automation{
		"builtin-one": {Name: "builtin-one", Description: "Built-in one", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo one"}}},
		"builtin-two": {Name: "builtin-two", Description: "Built-in two", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo two"}}},
	}
	builtinResult := NewResult(builtinAutomations, []string{"builtin-one", "builtin-two"})

	result.MergeBuiltins(builtinResult)

	if len(result.Automations) != 3 {
		t.Errorf("expected 3 automations after merge, got %d", len(result.Automations))
	}

	names := result.Names()
	expected := []string{"builtin-one", "builtin-two", "local"}
	if len(names) != len(expected) {
		t.Fatalf("expected names %v, got %v", expected, names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d]: expected %q, got %q", i, name, names[i])
		}
	}
}

func TestMergeBuiltins_LocalTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "shared.yaml"),
		makeAutomation("shared", "Local version"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builtinAutomations := map[string]*automation.Automation{
		"shared": {Name: "shared", Description: "Built-in version", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo builtin"}}},
	}
	builtinResult := NewResult(builtinAutomations, []string{"shared"})

	result.MergeBuiltins(builtinResult)

	a := result.Automations["shared"]
	if a.Description != "Local version" {
		t.Errorf("expected local version to take precedence, got description: %q", a.Description)
	}

	if result.IsBuiltin("shared") {
		t.Error("expected 'shared' NOT to be marked as built-in (local shadows it)")
	}
}

func TestIsBuiltin_CorrectlyMarked(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "local.yaml"),
		makeAutomation("local", "Local"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builtinAutomations := map[string]*automation.Automation{
		"builtin": {Name: "builtin", Description: "Built-in", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo b"}}},
	}
	builtinResult := NewResult(builtinAutomations, []string{"builtin"})
	result.MergeBuiltins(builtinResult)

	if result.IsBuiltin("local") {
		t.Error("expected 'local' NOT to be marked built-in")
	}
	if !result.IsBuiltin("builtin") {
		t.Error("expected 'builtin' to be marked built-in")
	}
}

func TestFind_BuiltinPrefix(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "shared.yaml"),
		makeAutomation("shared", "Local version"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	builtinAutomations := map[string]*automation.Automation{
		"shared":       {Name: "shared", Description: "Built-in version", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo builtin"}}},
		"builtin-only": {Name: "builtin-only", Description: "Only in built-in", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo only"}}},
	}
	builtinResult := NewResult(builtinAutomations, []string{"builtin-only", "shared"})
	result.MergeBuiltins(builtinResult)

	// Find("shared") → local
	a, err := result.Find("shared")
	if err != nil {
		t.Fatalf("Find('shared') error: %v", err)
	}
	if a.Description != "Local version" {
		t.Errorf("Find('shared') should return local, got: %q", a.Description)
	}

	// Find("pi:shared") → built-in
	a, err = result.Find("pi:shared")
	if err != nil {
		t.Fatalf("Find('pi:shared') error: %v", err)
	}
	if a.Description != "Built-in version" {
		t.Errorf("Find('pi:shared') should return built-in, got: %q", a.Description)
	}

	// Find("pi:builtin-only") → built-in
	a, err = result.Find("pi:builtin-only")
	if err != nil {
		t.Fatalf("Find('pi:builtin-only') error: %v", err)
	}
	if a.Description != "Only in built-in" {
		t.Errorf("Find('pi:builtin-only') should return built-in, got: %q", a.Description)
	}
}

func TestFind_BuiltinPrefix_NotFound(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "local.yaml"),
		makeAutomation("local", "Local"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No builtins merged — pi: prefix should fail
	_, err = result.Find("pi:nonexistent")
	if err == nil {
		t.Fatal("expected error for pi:nonexistent")
	}
	if !strings.Contains(err.Error(), "built-in automation") {
		t.Errorf("expected built-in error message, got: %v", err)
	}
}

func TestMergeBuiltins_NilIsNoOp(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)
	writeFile(t, filepath.Join(piDir, "local.yaml"),
		makeAutomation("local", "Local"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result.MergeBuiltins(nil)

	if len(result.Automations) != 1 {
		t.Errorf("expected 1 automation after nil merge, got %d", len(result.Automations))
	}
}

func TestDiscover_AutomationYAMLAtRoot(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, PiDir)

	// automation.yaml directly in .pi/ — should be skipped because
	// its directory name would be ".", which isn't a valid automation name
	writeFile(t, filepath.Join(piDir, "automation.yaml"),
		makeAutomation("root", "Root automation"))

	result, err := Discover(piDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Automations) != 0 {
		t.Errorf("expected automation.yaml at .pi/ root to be skipped, got %d automations", len(result.Automations))
	}
}
