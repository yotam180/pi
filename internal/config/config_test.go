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

func TestLoad_SetupBareString(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - setup/install-go
  - setup/install-ruby
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "setup/install-go" {
		t.Errorf("setup[0].Run = %q, want %q", cfg.Setup[0].Run, "setup/install-go")
	}
	if cfg.Setup[1].Run != "setup/install-ruby" {
		t.Errorf("setup[1].Run = %q, want %q", cfg.Setup[1].Run, "setup/install-ruby")
	}
	if cfg.Setup[0].If != "" {
		t.Errorf("setup[0].If = %q, want empty", cfg.Setup[0].If)
	}
}

func TestLoad_SetupMixedBareAndObject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - setup/install-go
  - run: setup/install-ruby
    if: os.macos
  - run: pi:install-python
    with:
      version: "3.13"
  - setup/install-uv
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Setup) != 4 {
		t.Fatalf("setup count = %d, want 4", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "setup/install-go" {
		t.Errorf("setup[0].Run = %q, want %q", cfg.Setup[0].Run, "setup/install-go")
	}
	if cfg.Setup[0].If != "" {
		t.Errorf("setup[0].If = %q, want empty", cfg.Setup[0].If)
	}
	if cfg.Setup[1].Run != "setup/install-ruby" {
		t.Errorf("setup[1].Run = %q, want %q", cfg.Setup[1].Run, "setup/install-ruby")
	}
	if cfg.Setup[1].If != "os.macos" {
		t.Errorf("setup[1].If = %q, want %q", cfg.Setup[1].If, "os.macos")
	}
	if cfg.Setup[2].Run != "pi:install-python" {
		t.Errorf("setup[2].Run = %q, want %q", cfg.Setup[2].Run, "pi:install-python")
	}
	if cfg.Setup[2].With["version"] != "3.13" {
		t.Errorf("setup[2].With[version] = %q, want %q", cfg.Setup[2].With["version"], "3.13")
	}
	if cfg.Setup[3].Run != "setup/install-uv" {
		t.Errorf("setup[3].Run = %q, want %q", cfg.Setup[3].Run, "setup/install-uv")
	}
}

func TestLoad_SetupBareStringEmpty(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - ""
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for empty bare setup entry")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention 'empty', got: %v", err)
	}
}

func TestLoad_SetupBareStringWithBuiltin(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
setup:
  - pi:install-python
  - setup/local-thing
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-python" {
		t.Errorf("setup[0].Run = %q, want %q", cfg.Setup[0].Run, "pi:install-python")
	}
	if cfg.Setup[1].Run != "setup/local-thing" {
		t.Errorf("setup[1].Run = %q, want %q", cfg.Setup[1].Run, "setup/local-thing")
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

func TestLoad_RuntimesConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
runtimes:
  provision: auto
  manager: mise
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Runtimes == nil {
		t.Fatal("runtimes should not be nil")
	}
	if cfg.Runtimes.Provision != ProvisionAuto {
		t.Errorf("provision = %q, want %q", cfg.Runtimes.Provision, ProvisionAuto)
	}
	if cfg.Runtimes.Manager != RuntimeManagerMise {
		t.Errorf("manager = %q, want %q", cfg.Runtimes.Manager, RuntimeManagerMise)
	}
}

func TestLoad_RuntimesConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Runtimes != nil {
		t.Error("runtimes should be nil when not specified")
	}
	if cfg.EffectiveProvisionMode() != ProvisionNever {
		t.Errorf("default provision = %q, want %q", cfg.EffectiveProvisionMode(), ProvisionNever)
	}
	if cfg.EffectiveRuntimeManager() != RuntimeManagerMise {
		t.Errorf("default manager = %q, want %q", cfg.EffectiveRuntimeManager(), RuntimeManagerMise)
	}
}

func TestLoad_RuntimesInvalidProvision(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
runtimes:
  provision: invalid
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid provision mode")
	}
	if !strings.Contains(err.Error(), "provision") {
		t.Errorf("error should mention 'provision', got: %v", err)
	}
}

func TestLoad_RuntimesInvalidManager(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
runtimes:
  manager: invalid
`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid manager")
	}
	if !strings.Contains(err.Error(), "manager") {
		t.Errorf("error should mention 'manager', got: %v", err)
	}
}

func TestLoad_RuntimesAllModes(t *testing.T) {
	for _, mode := range []ProvisionMode{ProvisionNever, ProvisionAsk, ProvisionAuto} {
		dir := t.TempDir()
		writeFile(t, dir, "pi.yaml", `
project: test
runtimes:
  provision: `+string(mode)+`
`)

		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("unexpected error for mode %q: %v", mode, err)
		}
		if cfg.Runtimes.Provision != mode {
			t.Errorf("provision = %q, want %q", cfg.Runtimes.Provision, mode)
		}
	}
}

func TestLoad_RuntimesDirectManager(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `
project: test
runtimes:
  provision: auto
  manager: direct
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Runtimes.Manager != RuntimeManagerDirect {
		t.Errorf("manager = %q, want %q", cfg.Runtimes.Manager, RuntimeManagerDirect)
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
