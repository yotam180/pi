package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunAdd_FileSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "file:~/my-automations", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "file:~/my-automations" {
		t.Errorf("source = %q, want %q", cfg.Packages[0].Source, "file:~/my-automations")
	}

	if !strings.Contains(stderr.String(), "added") {
		t.Errorf("stderr should contain 'added', got: %q", stderr.String())
	}
}

func TestRunAdd_FileSourceWithAlias(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "file:~/my-automations", "mytools", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}
	if cfg.Packages[0].As != "mytools" {
		t.Errorf("as = %q, want %q", cfg.Packages[0].As, "mytools")
	}
}

func TestRunAdd_IdempotentDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", `project: test
packages:
  - file:~/my-automations
`)

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "file:~/my-automations", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stderr.String(), "already in") {
		t.Errorf("stderr should say 'already in', got: %q", stderr.String())
	}

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if len(cfg.Packages) != 1 {
		t.Errorf("should not duplicate, got %d packages", len(cfg.Packages))
	}
}

func TestRunAdd_NoVersionError(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "yotam180/pi-common", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "version required") {
		t.Errorf("error should mention 'version required', got: %v", err)
	}
}

func TestRunAdd_InvalidSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "just-a-name", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	if !strings.Contains(err.Error(), "invalid package source") {
		t.Errorf("error should mention 'invalid package source', got: %v", err)
	}
}

func TestRunAdd_NoPiYaml(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "file:~/path", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing pi.yaml")
	}
}

func TestRunAdd_NoArgs(t *testing.T) {
	cmd := newAddCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no args")
	}
}

func TestRunAdd_BuiltinRefError(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pi.yaml", "project: test\n")

	var stdout, stderr bytes.Buffer
	err := runAdd(dir, "pi:hello", "", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for builtin ref")
	}
	if !strings.Contains(err.Error(), "invalid package source") {
		t.Errorf("error should mention 'invalid package source', got: %v", err)
	}
}
