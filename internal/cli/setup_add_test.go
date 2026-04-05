package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func writeTestPiDir(t *testing.T, dir, name, content string) {
	t.Helper()
	writeTestFile(t, dir, filepath.Join(".pi", name), content)
}

func TestRunSetupAdd_BareString(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-uv", nil, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-uv" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-uv")
	}

	if !strings.Contains(stdout.String(), "Added to setup") {
		t.Errorf("stdout should contain 'Added to setup', got: %q", stdout.String())
	}
}

func TestRunSetupAdd_ShortFormExpansion(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "python", nil, "3.13", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-python" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-python")
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q, want %q", cfg.Setup[0].With["version"], "3.13")
	}

	if !strings.Contains(stdout.String(), "Resolved 'python'") {
		t.Errorf("stdout should show resolution, got: %q", stdout.String())
	}
}

func TestRunSetupAdd_PiPrefixExpansion(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:go", nil, "1.23", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].Run != "pi:install-go" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-go")
	}
}

func TestRunSetupAdd_WithIfFlag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-homebrew", nil, "", "os.macos", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].If != "os.macos" {
		t.Errorf("if = %q, want %q", cfg.Setup[0].If, "os.macos")
	}
}

func TestRunSetupAdd_KeyValueArgs(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:cursor/install-extensions", []string{"file=.pi/cursor/extensions.txt"}, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].With["file"] != ".pi/cursor/extensions.txt" {
		t.Errorf("with.file = %q, want %q", cfg.Setup[0].With["file"], ".pi/cursor/extensions.txt")
	}
}

func TestRunSetupAdd_InvalidKeyValue(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-uv", []string{"badarg"}, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid key=value")
	}
	if !strings.Contains(err.Error(), "invalid key=value") {
		t.Errorf("error should mention 'invalid key=value', got: %v", err)
	}
}

func TestRunSetupAdd_Duplicate(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", `project: test
setup:
  - pi:install-uv
`)

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-uv", nil, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error (idempotent): %v", err)
	}

	if !strings.Contains(stdout.String(), "Already in pi.yaml") {
		t.Errorf("stdout should say 'Already in pi.yaml', got: %q", stdout.String())
	}
}

func TestRunSetupAdd_NoPiYaml_YesFlag(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-uv", nil, "", "", "", "", true, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Project == "" {
		t.Error("project should be initialized")
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-uv" {
		t.Errorf("run = %q", cfg.Setup[0].Run)
	}

	if !strings.Contains(stdout.String(), "Initialized project") {
		t.Errorf("stdout should mention initialization, got: %q", stdout.String())
	}
}

func TestRunSetupAdd_NoPiYaml_NonInteractive(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-uv", nil, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Project == "" {
		t.Error("project should be initialized (non-interactive auto-accepts)")
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
}

func TestRunSetupAdd_LocalAutomation(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "setup/install-deps", nil, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].Run != "setup/install-deps" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "setup/install-deps")
	}

	if strings.Contains(stdout.String(), "Resolved") {
		t.Errorf("should not show resolution for local automation, got: %q", stdout.String())
	}
}

func TestRunSetupAdd_SourceFlag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:git/install-hooks", nil, "", "", ".pi/hooks", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].With["source"] != ".pi/hooks" {
		t.Errorf("with.source = %q, want %q", cfg.Setup[0].With["source"], ".pi/hooks")
	}
}

func TestRunSetupAdd_GroupsFlag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:uv/sync", nil, "", "", "", "dev,local", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].With["groups"] != "dev,local" {
		t.Errorf("with.groups = %q, want %q", cfg.Setup[0].With["groups"], "dev,local")
	}
}

func TestRunSetupAdd_ReplaceSameRunTarget(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\nsetup:\n  - pi:install-node\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "pi:install-node", nil, "22", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Replaced in pi.yaml") {
		t.Errorf("stdout should say 'Replaced in pi.yaml', got: %q", stdout.String())
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1 (should replace, not append)", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-node" {
		t.Errorf("run = %q, want pi:install-node", cfg.Setup[0].Run)
	}
	if cfg.Setup[0].With["version"] != "22" {
		t.Errorf("with.version = %q, want 22", cfg.Setup[0].With["version"])
	}
}

func TestRunSetupAdd_InvokesBeforeWriting(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")
	writeTestPiDir(t, dir, "greet.yaml", "description: Say hello\nbash: echo hello\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "greet", nil, "", "", "", "", false, false, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Added to setup") {
		t.Errorf("stdout should say 'Added to setup', got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "hello") {
		t.Errorf("stdout should contain automation output 'hello', got: %q", stdout.String())
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
}

func TestRunSetupAdd_InvokeFailure_NoPiYamlModification(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")
	writeTestPiDir(t, dir, "fail.yaml", "description: Failing automation\nbash: exit 1\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "fail", nil, "", "", "", "", false, false, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for failing automation")
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Setup) != 0 {
		t.Errorf("setup count = %d, want 0 (pi.yaml should not be modified on failure)", len(cfg.Setup))
	}
}

func TestRunSetupAdd_OnlyAddSkipsExecution(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "nonexistent/automation", nil, "", "", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error with --only-add: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "nonexistent/automation" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "nonexistent/automation")
	}
}

func TestRunSetupAdd_InvokeNotFound_NoPiYamlModification(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "nonexistent/automation", nil, "", "", "", "", false, false, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for not-found automation")
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Setup) != 0 {
		t.Errorf("setup count = %d, want 0 (pi.yaml should not be modified when automation not found)", len(cfg.Setup))
	}
}

func TestRunSetupAdd_CombinedFlags(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runSetupAdd(dir, "python", nil, "3.13", "os.macos", "", "", false, true, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if cfg.Setup[0].Run != "pi:install-python" {
		t.Errorf("run = %q", cfg.Setup[0].Run)
	}
	if cfg.Setup[0].If != "os.macos" {
		t.Errorf("if = %q", cfg.Setup[0].If)
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q", cfg.Setup[0].With["version"])
	}
}
