package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_ValidFull(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: my-app

shortcuts:
  up:    docker/up
  down:  docker/down
  deploy:
    run: deploy/push
    anywhere: true

setup:
  - run: setup/install-deps
  - run: pi:install-python
    with:
      version: "3.13"
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project != "my-app" {
		t.Errorf("project = %q, want %q", cfg.Project, "my-app")
	}

	if len(cfg.Shortcuts) != 3 {
		t.Fatalf("shortcuts count = %d, want 3", len(cfg.Shortcuts))
	}

	up := cfg.Shortcuts["up"]
	if up.Run != "docker/up" || up.Anywhere {
		t.Errorf("up shortcut = %+v, want {Run:docker/up Anywhere:false}", up)
	}

	deploy := cfg.Shortcuts["deploy"]
	if deploy.Run != "deploy/push" || !deploy.Anywhere {
		t.Errorf("deploy shortcut = %+v, want {Run:deploy/push Anywhere:true}", deploy)
	}

	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "setup/install-deps" {
		t.Errorf("setup[0].Run = %q, want %q", cfg.Setup[0].Run, "setup/install-deps")
	}
	if cfg.Setup[1].With["version"] != "3.13" {
		t.Errorf("setup[1].With[version] = %q, want %q", cfg.Setup[1].With["version"], "3.13")
	}
}

func TestLoad_MinimalValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: minimal`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project != "minimal" {
		t.Errorf("project = %q, want %q", cfg.Project, "minimal")
	}
	if cfg.Shortcuts != nil && len(cfg.Shortcuts) != 0 {
		t.Errorf("shortcuts should be empty, got %v", cfg.Shortcuts)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing pi.yaml")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestLoad_MissingProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
shortcuts:
  up: docker/up
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing project field")
	}
	if !strings.Contains(err.Error(), "project") {
		t.Errorf("error should mention 'project', got: %v", err)
	}
}

func TestLoad_EmptyShortcutRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
shortcuts:
  bad:
    run: ""
    anywhere: true
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty shortcut run")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention 'empty', got: %v", err)
	}
}

func TestLoad_EmptySetupRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - run: ""
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty setup run")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention 'empty', got: %v", err)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
shortcuts: [[[invalid
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention 'parsing', got: %v", err)
	}
}

func TestLoad_ShortcutStringAndObject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
shortcuts:
  simple: docker/up
  complex:
    run: deploy/push
    anywhere: true
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	simple := cfg.Shortcuts["simple"]
	if simple.Run != "docker/up" {
		t.Errorf("simple.Run = %q, want %q", simple.Run, "docker/up")
	}
	if simple.Anywhere {
		t.Error("simple.Anywhere should be false")
	}

	complex := cfg.Shortcuts["complex"]
	if complex.Run != "deploy/push" {
		t.Errorf("complex.Run = %q, want %q", complex.Run, "deploy/push")
	}
	if !complex.Anywhere {
		t.Error("complex.Anywhere should be true")
	}
}

func TestLoad_SetupEntryWithIf(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - run: setup/install-brew
    if: os.macos
  - run: setup/install-uv
    if: not command.uv
  - run: setup/install-node
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Setup) != 3 {
		t.Fatalf("setup count = %d, want 3", len(cfg.Setup))
	}
	if cfg.Setup[0].If != "os.macos" {
		t.Errorf("setup[0].If = %q, want %q", cfg.Setup[0].If, "os.macos")
	}
	if cfg.Setup[1].If != "not command.uv" {
		t.Errorf("setup[1].If = %q, want %q", cfg.Setup[1].If, "not command.uv")
	}
	if cfg.Setup[2].If != "" {
		t.Errorf("setup[2].If = %q, want empty", cfg.Setup[2].If)
	}
}

func TestLoad_SetupEntryWithInvalidIf(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - run: setup/install-brew
    if: "and and"
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid if expression")
	}
	if !strings.Contains(err.Error(), "setup[0]") {
		t.Errorf("error should reference setup[0], got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid if expression") {
		t.Errorf("error should mention invalid if expression, got: %v", err)
	}
}

func TestLoad_ShortcutWithMapping(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
shortcuts:
  dlogs:
    run: docker/logs
    with:
      service: $1
      tail: $2
  dlogs-short:
    run: docker/logs
    with:
      tail: "50"
      service: $1
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dlogs := cfg.Shortcuts["dlogs"]
	if dlogs.With["service"] != "$1" {
		t.Errorf("dlogs.With[service] = %q, want %q", dlogs.With["service"], "$1")
	}
	if dlogs.With["tail"] != "$2" {
		t.Errorf("dlogs.With[tail] = %q, want %q", dlogs.With["tail"], "$2")
	}

	short := cfg.Shortcuts["dlogs-short"]
	if short.With["tail"] != "50" {
		t.Errorf("dlogs-short.With[tail] = %q, want %q", short.With["tail"], "50")
	}
	if short.With["service"] != "$1" {
		t.Errorf("dlogs-short.With[service] = %q, want %q", short.With["service"], "$1")
	}
}
