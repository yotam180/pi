package runtimes

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func TestNewProvisioner_Defaults(t *testing.T) {
	cfg := &config.ProjectConfig{Project: "test"}
	p := NewProvisioner(cfg, nil)

	if p.Mode != config.ProvisionNever {
		t.Errorf("mode = %q, want %q", p.Mode, config.ProvisionNever)
	}
	if p.Manager != config.RuntimeManagerMise {
		t.Errorf("manager = %q, want %q", p.Manager, config.RuntimeManagerMise)
	}
}

func TestNewProvisioner_FromConfig(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Runtimes: &config.RuntimesConfig{
			Provision: config.ProvisionAuto,
			Manager:   config.RuntimeManagerDirect,
		},
	}
	p := NewProvisioner(cfg, nil)

	if p.Mode != config.ProvisionAuto {
		t.Errorf("mode = %q, want %q", p.Mode, config.ProvisionAuto)
	}
	if p.Manager != config.RuntimeManagerDirect {
		t.Errorf("manager = %q, want %q", p.Manager, config.RuntimeManagerDirect)
	}
}

func TestProvision_NeverMode(t *testing.T) {
	p := &Provisioner{
		Mode:    config.ProvisionNever,
		BaseDir: t.TempDir(),
	}

	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provisioned {
		t.Error("should not provision in never mode")
	}
	if result.BinDir != "" {
		t.Errorf("binDir should be empty, got %q", result.BinDir)
	}
}

func TestProvision_UnknownRuntime(t *testing.T) {
	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		BaseDir: t.TempDir(),
	}

	_, err := p.Provision("ruby", "3.0")
	if err == nil {
		t.Fatal("expected error for unknown runtime")
	}
	if !strings.Contains(err.Error(), "unknown runtime") {
		t.Errorf("error should mention unknown runtime, got: %v", err)
	}
}

func TestProvision_AlreadyProvisioned(t *testing.T) {
	base := t.TempDir()
	binDir := filepath.Join(base, "python", "3.13", "bin")
	os.MkdirAll(binDir, 0755)
	os.WriteFile(filepath.Join(binDir, "python3"), []byte("#!/bin/sh\necho ok"), 0755)

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		BaseDir: base,
	}

	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Provisioned {
		t.Error("should detect already-provisioned runtime")
	}
	if result.BinDir != binDir {
		t.Errorf("binDir = %q, want %q", result.BinDir, binDir)
	}
}

func TestProvision_AskMode_NoPromptFunc(t *testing.T) {
	p := &Provisioner{
		Mode:    config.ProvisionAsk,
		BaseDir: t.TempDir(),
	}

	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provisioned {
		t.Error("should not provision when PromptFunc is nil (non-interactive)")
	}
}

func TestProvision_AskMode_UserDeclines(t *testing.T) {
	p := &Provisioner{
		Mode:       config.ProvisionAsk,
		BaseDir:    t.TempDir(),
		PromptFunc: func(msg string) bool { return false },
	}

	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Provisioned {
		t.Error("should not provision when user declines")
	}
}

func TestProvision_AskMode_UserAccepts(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	prompted := false
	p := &Provisioner{
		Mode:    config.ProvisionAsk,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		PromptFunc: func(msg string) bool {
			prompted = true
			if !strings.Contains(msg, "python") {
				t.Errorf("prompt should mention python, got: %s", msg)
			}
			return true
		},
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	// This will fail because direct download will fail in test env,
	// but we can verify the prompt was called
	_, _ = p.Provision("python", "3.13")
	if !prompted {
		t.Error("PromptFunc should have been called")
	}
}

func TestBinDirFor_NotProvisioned(t *testing.T) {
	p := &Provisioner{
		BaseDir: t.TempDir(),
	}

	binDir := p.BinDirFor("python", "3.13")
	if binDir != "" {
		t.Errorf("binDir should be empty for non-provisioned runtime, got %q", binDir)
	}
}

func TestBinDirFor_Provisioned(t *testing.T) {
	base := t.TempDir()
	binDir := filepath.Join(base, "node", "20", "bin")
	os.MkdirAll(binDir, 0755)
	os.WriteFile(filepath.Join(binDir, "node"), []byte("#!/bin/sh\necho ok"), 0755)

	p := &Provisioner{
		BaseDir: base,
	}

	got := p.BinDirFor("node", "20")
	if got != binDir {
		t.Errorf("binDir = %q, want %q", got, binDir)
	}
}

func TestPrependToPath(t *testing.T) {
	result := PrependToPath("/foo/bin")
	if !strings.HasPrefix(result, "/foo/bin"+string(os.PathListSeparator)) {
		t.Errorf("PATH should start with /foo/bin, got: %s", result)
	}
}

func TestDefaultVersion(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{"python", "3.13"},
		{"node", "20"},
		{"go", "1.23"},
		{"rust", "stable"},
		{"unknown", "latest"},
	}
	for _, tt := range tests {
		got := defaultVersion(tt.runtime)
		if got != tt.want {
			t.Errorf("defaultVersion(%q) = %q, want %q", tt.runtime, got, tt.want)
		}
	}
}

func TestRuntimeBinary(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{"python", "python3"},
		{"node", "node"},
		{"rust", "rustc"},
		{"go", "go"},
		{"other", "other"},
	}
	for _, tt := range tests {
		got := runtimeBinary(tt.runtime)
		if got != tt.want {
			t.Errorf("runtimeBinary(%q) = %q, want %q", tt.runtime, got, tt.want)
		}
	}
}

func TestKnownRuntimes(t *testing.T) {
	for _, rt := range []string{"python", "node", "go", "rust"} {
		if !KnownRuntimes[rt] {
			t.Errorf("%s should be a known runtime", rt)
		}
	}
	if KnownRuntimes["ruby"] {
		t.Error("ruby should not be a known runtime")
	}
}

func TestProvision_MiseFallbackToDirect(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	// When mise is not found, it should fall back to direct download
	// In CI/test, direct download may also fail, but the fallback path should be hit
	_, err := p.Provision("node", "20")
	// The error should come from the direct provisioner, not "mise not found"
	if err != nil && strings.Contains(err.Error(), "mise not found") {
		t.Errorf("should fall back to direct download when mise is not in PATH, got: %v", err)
	}
}

func TestProvision_AutoMode_MiseInstalled(t *testing.T) {
	// Only run when mise is actually installed
	misePath, err := exec.LookPath("mise")
	if err != nil {
		t.Skip("mise not installed, skipping mise integration test")
	}

	base := t.TempDir()
	var stderr bytes.Buffer

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return misePath, nil
			}
			return exec.LookPath(name)
		},
	}

	// Try provisioning python with mise
	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Skipf("mise provisioning failed (may need network): %v", err)
	}

	if !result.Provisioned {
		t.Error("should have provisioned python")
	}
	if result.BinDir == "" {
		t.Error("binDir should not be empty after provisioning")
	}

	// Verify the binary exists
	python3 := filepath.Join(result.BinDir, "python3")
	if _, err := os.Stat(python3); err != nil {
		t.Errorf("python3 binary should exist at %s: %v", python3, err)
	}

	// Check that the status line was printed
	if !strings.Contains(stderr.String(), "[provisioned]") {
		t.Errorf("stderr should contain [provisioned], got: %s", stderr.String())
	}
}
