package runtimes

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/runtimeinfo"
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

func TestProvisionGoDirect_RunnerCalled(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			return fmt.Errorf("download failed")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("go", "1.23")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "downloading go") {
		t.Errorf("error should mention downloading, got: %v", err)
	}

	if len(mock.runCalls) < 1 {
		t.Fatal("expected at least 1 run call")
	}
	if mock.runCalls[0].Bin != "bash" {
		t.Errorf("bin = %q, want bash", mock.runCalls[0].Bin)
	}
	// Verify the script contains the correct download URL
	script := mock.runCalls[0].Args[1]
	if !strings.Contains(script, "dl.google.com/go/go1.23") {
		t.Errorf("script should reference go download URL, got: %s", script)
	}
}

func TestProvisionGoDirect_ScriptContainsCorrectURL(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer
	var capturedScript string

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			if bin == "bash" && len(args) >= 2 {
				capturedScript = args[1]
			}
			return fmt.Errorf("fail")
		},
	}

	p := &Provisioner{
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
	}

	p.provisionGoDirect("1.22.5")
	if !strings.Contains(capturedScript, "go1.22.5") {
		t.Errorf("script should contain version, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "dl.google.com/go") {
		t.Errorf("script should use Google CDN, got: %s", capturedScript)
	}
}

func TestProvisionRustDirect_RunnerCalled(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			return fmt.Errorf("download failed")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("rust", "stable")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "installing rust") {
		t.Errorf("error should mention installing, got: %v", err)
	}

	if len(mock.runCalls) < 1 {
		t.Fatal("expected at least 1 run call")
	}
	if mock.runCalls[0].Bin != "bash" {
		t.Errorf("bin = %q, want bash", mock.runCalls[0].Bin)
	}
	script := mock.runCalls[0].Args[1]
	if !strings.Contains(script, "rustup.rs") {
		t.Errorf("script should reference rustup.rs, got: %s", script)
	}
}

func TestProvisionRustDirect_ScriptContainsVersion(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer
	var capturedScript string

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			if bin == "bash" && len(args) >= 2 {
				capturedScript = args[1]
			}
			return fmt.Errorf("fail")
		},
	}

	p := &Provisioner{
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
	}

	p.provisionRustDirect("1.78.0")
	if !strings.Contains(capturedScript, "1.78.0") {
		t.Errorf("script should contain version, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "RUSTUP_HOME") {
		t.Errorf("script should set RUSTUP_HOME, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "CARGO_HOME") {
		t.Errorf("script should set CARGO_HOME, got: %s", capturedScript)
	}
	if !strings.Contains(capturedScript, "--no-modify-path") {
		t.Errorf("script should use --no-modify-path, got: %s", capturedScript)
	}
}

func TestProvisionDirect_AllKnownRuntimesSupported(t *testing.T) {
	for _, name := range []string{"python", "node", "go", "rust"} {
		desc := runtimeinfo.Find(name)
		if desc == nil {
			t.Errorf("runtime %q should be known", name)
			continue
		}
		if !desc.DirectDownload {
			t.Errorf("runtime %q should support direct download", name)
		}
	}
}

func TestProvision_UnknownManager(t *testing.T) {
	base := t.TempDir()
	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: "unknown-manager",
		BaseDir: base,
	}

	_, err := p.Provision("python", "3.13")
	if err == nil {
		t.Fatal("expected error for unknown runtime manager")
	}
	if !strings.Contains(err.Error(), "unknown runtime manager") {
		t.Errorf("error should mention unknown runtime manager, got: %v", err)
	}
}

func TestBinDirFor_DefaultVersion(t *testing.T) {
	base := t.TempDir()
	binDir := filepath.Join(base, "python", "3.13", "bin")
	os.MkdirAll(binDir, 0755)
	os.WriteFile(filepath.Join(binDir, "python3"), []byte("#!/bin/sh\necho ok"), 0755)

	p := &Provisioner{BaseDir: base}

	got := p.BinDirFor("python", "")
	if got != binDir {
		t.Errorf("BinDirFor with empty version should use default, got %q, want %q", got, binDir)
	}
}

func TestProvision_AskMode_VersionInPrompt(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer
	var promptMsg string

	p := &Provisioner{
		Mode:    config.ProvisionAsk,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		PromptFunc: func(msg string) bool {
			promptMsg = msg
			return false
		},
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	p.Provision("node", "20")
	if !strings.Contains(promptMsg, "node") {
		t.Errorf("prompt should mention node, got: %s", promptMsg)
	}
	if !strings.Contains(promptMsg, ">= 20") {
		t.Errorf("prompt should mention version >= 20, got: %s", promptMsg)
	}
}

func TestProvision_AskMode_NoVersionInPrompt(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer
	var promptMsg string

	p := &Provisioner{
		Mode:    config.ProvisionAsk,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		PromptFunc: func(msg string) bool {
			promptMsg = msg
			return false
		},
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	p.Provision("python", "")
	if !strings.Contains(promptMsg, "python") {
		t.Errorf("prompt should mention python, got: %s", promptMsg)
	}
	if strings.Contains(promptMsg, ">=") {
		t.Errorf("prompt should not mention >= when no version specified, got: %s", promptMsg)
	}
}

func TestProvisionDirect_UnknownRuntime(t *testing.T) {
	p := &Provisioner{
		BaseDir: t.TempDir(),
	}
	err := p.provisionDirect("unknown", "1.0")
	if err == nil {
		t.Fatal("expected error for unknown runtime")
	}
	if !strings.Contains(err.Error(), "direct provisioning not supported") {
		t.Errorf("error should mention unsupported, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mise") {
		t.Errorf("error should suggest mise, got: %v", err)
	}
}

func TestBinDir_DefaultVersionPath(t *testing.T) {
	base := t.TempDir()
	p := &Provisioner{BaseDir: base}

	got := p.binDir("python", "")
	want := filepath.Join(base, "python", "3.13", "bin")
	if got != want {
		t.Errorf("binDir with empty version: got %q, want %q", got, want)
	}

	got = p.binDir("node", "")
	want = filepath.Join(base, "node", "20", "bin")
	if got != want {
		t.Errorf("binDir with empty version: got %q, want %q", got, want)
	}

	got = p.binDir("go", "")
	want = filepath.Join(base, "go", "1.23", "bin")
	if got != want {
		t.Errorf("binDir with empty version: got %q, want %q", got, want)
	}

	got = p.binDir("rust", "")
	want = filepath.Join(base, "rust", "stable", "bin")
	if got != want {
		t.Errorf("binDir with empty version: got %q, want %q", got, want)
	}
}

func TestStderr_Default(t *testing.T) {
	p := &Provisioner{}
	w := p.stderr()
	if w != os.Stderr {
		t.Error("stderr() should return os.Stderr when Stderr field is nil")
	}
}

func TestProvision_AutoMode_MiseInstalled(t *testing.T) {
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

	python3 := filepath.Join(result.BinDir, "python3")
	if _, err := os.Stat(python3); err != nil {
		t.Errorf("python3 binary should exist at %s: %v", python3, err)
	}

	if !strings.Contains(stderr.String(), "[provisioned]") {
		t.Errorf("stderr should contain [provisioned], got: %s", stderr.String())
	}
}

// --- CmdRunner interface tests ---

type mockCmdRunner struct {
	runCalls    []mockRunCall
	outputCalls []mockOutputCall
	runErr      func(bin string, args []string) error
	outputFunc  func(bin string, args []string) (string, error)
}

type mockRunCall struct {
	Bin  string
	Args []string
}

type mockOutputCall struct {
	Bin  string
	Args []string
}

func (m *mockCmdRunner) Run(bin string, args []string, stdout, stderr io.Writer) error {
	m.runCalls = append(m.runCalls, mockRunCall{Bin: bin, Args: args})
	if m.runErr != nil {
		return m.runErr(bin, args)
	}
	return nil
}

func (m *mockCmdRunner) Output(bin string, args []string, stderr io.Writer) (string, error) {
	m.outputCalls = append(m.outputCalls, mockOutputCall{Bin: bin, Args: args})
	if m.outputFunc != nil {
		return m.outputFunc(bin, args)
	}
	return "", nil
}

func TestRunner_DefaultsToExec(t *testing.T) {
	p := &Provisioner{}
	r := p.runner()
	if _, ok := r.(*execCmdRunner); !ok {
		t.Errorf("runner() should return *execCmdRunner when Runner is nil, got %T", r)
	}
}

func TestRunner_UsesCustomRunner(t *testing.T) {
	mock := &mockCmdRunner{}
	p := &Provisioner{Runner: mock}
	r := p.runner()
	if r != mock {
		t.Error("runner() should return the custom Runner when set")
	}
}

func TestExecCmdRunner_Run(t *testing.T) {
	r := &execCmdRunner{}
	var out bytes.Buffer
	err := r.Run("echo", []string{"hello"}, &out, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("stdout should contain 'hello', got: %q", out.String())
	}
}

func TestExecCmdRunner_Output(t *testing.T) {
	r := &execCmdRunner{}
	out, err := r.Output("echo", []string{"hello"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Errorf("output = %q, want %q", out, "hello")
	}
}

func TestExecCmdRunner_RunError(t *testing.T) {
	r := &execCmdRunner{}
	err := r.Run("false", nil, io.Discard, io.Discard)
	if err == nil {
		t.Error("expected error from 'false' command")
	}
}

func TestExecCmdRunner_OutputError(t *testing.T) {
	r := &execCmdRunner{}
	_, err := r.Output("false", nil, io.Discard)
	if err == nil {
		t.Error("expected error from 'false' command")
	}
}

// --- provisionWithMise mock tests ---

func TestProvisionWithMise_Success(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	// Create a fake mise bin directory with binaries
	fakeMiseBinDir := filepath.Join(t.TempDir(), "mise-installs", "python", "3.13", "bin")
	os.MkdirAll(fakeMiseBinDir, 0755)
	os.WriteFile(filepath.Join(fakeMiseBinDir, "python3"), []byte("#!/bin/sh\necho ok"), 0755)
	os.WriteFile(filepath.Join(fakeMiseBinDir, "pip3"), []byte("#!/bin/sh\necho ok"), 0755)

	mock := &mockCmdRunner{
		outputFunc: func(bin string, args []string) (string, error) {
			if len(args) >= 2 && args[0] == "where" {
				return filepath.Dir(fakeMiseBinDir), nil
			}
			return "", nil
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	result, err := p.Provision("python", "3.13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Provisioned {
		t.Error("should have provisioned")
	}
	if result.Version != "3.13" {
		t.Errorf("version = %q, want %q", result.Version, "3.13")
	}

	// Verify mise install was called
	if len(mock.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(mock.runCalls))
	}
	call := mock.runCalls[0]
	if call.Bin != "/usr/local/bin/mise" {
		t.Errorf("install bin = %q, want /usr/local/bin/mise", call.Bin)
	}
	if len(call.Args) != 2 || call.Args[0] != "install" || call.Args[1] != "python@3.13" {
		t.Errorf("install args = %v, want [install python@3.13]", call.Args)
	}

	// Verify mise where was called
	if len(mock.outputCalls) != 1 {
		t.Fatalf("expected 1 output call, got %d", len(mock.outputCalls))
	}
	outCall := mock.outputCalls[0]
	if outCall.Args[0] != "where" || outCall.Args[1] != "python@3.13" {
		t.Errorf("where args = %v, want [where python@3.13]", outCall.Args)
	}

	// Verify symlinks were created
	binDir := filepath.Join(base, "python", "3.13", "bin")
	for _, name := range []string{"python3", "pip3"} {
		link := filepath.Join(binDir, name)
		target, err := os.Readlink(link)
		if err != nil {
			t.Errorf("symlink %s should exist: %v", name, err)
			continue
		}
		expected := filepath.Join(fakeMiseBinDir, name)
		if target != expected {
			t.Errorf("symlink %s → %q, want %q", name, target, expected)
		}
	}

	if !strings.Contains(stderr.String(), "[provisioned]") {
		t.Errorf("stderr should contain [provisioned], got: %s", stderr.String())
	}
}

func TestProvisionWithMise_InstallFailure(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			if len(args) >= 1 && args[0] == "install" {
				return fmt.Errorf("exit status 1")
			}
			return nil
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("node", "20")
	if err == nil {
		t.Fatal("expected error when mise install fails")
	}
	if !strings.Contains(err.Error(), "mise install node@20") {
		t.Errorf("error should mention mise install, got: %v", err)
	}
}

func TestProvisionWithMise_WhereFailure(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		outputFunc: func(bin string, args []string) (string, error) {
			return "", fmt.Errorf("mise where failed")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("python", "3.13")
	if err == nil {
		t.Fatal("expected error when mise where fails")
	}
	if !strings.Contains(err.Error(), "mise where python@3.13") {
		t.Errorf("error should mention mise where, got: %v", err)
	}
}

func TestProvisionWithMise_SymlinkOverwrite(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	fakeMiseBinDir := filepath.Join(t.TempDir(), "mise-installs", "go", "1.23", "bin")
	os.MkdirAll(fakeMiseBinDir, 0755)
	os.WriteFile(filepath.Join(fakeMiseBinDir, "go"), []byte("#!/bin/sh\necho new"), 0755)

	// Pre-create a stale symlink
	binDir := filepath.Join(base, "go", "1.23", "bin")
	os.MkdirAll(binDir, 0755)
	os.Symlink("/nonexistent/old-go", filepath.Join(binDir, "go"))

	mock := &mockCmdRunner{
		outputFunc: func(bin string, args []string) (string, error) {
			return filepath.Dir(fakeMiseBinDir), nil
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	result, err := p.Provision("go", "1.23")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Provisioned {
		t.Error("should have provisioned")
	}

	// Verify old symlink was replaced
	target, err := os.Readlink(filepath.Join(binDir, "go"))
	if err != nil {
		t.Fatalf("symlink should exist: %v", err)
	}
	expected := filepath.Join(fakeMiseBinDir, "go")
	if target != expected {
		t.Errorf("symlink go → %q, want %q", target, expected)
	}
}

func TestProvisionWithMise_EmptyBinDir(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	fakeMiseDir := filepath.Join(t.TempDir(), "mise-installs", "node", "20")
	fakeBinDir := filepath.Join(fakeMiseDir, "bin")
	os.MkdirAll(fakeBinDir, 0755)

	mock := &mockCmdRunner{
		outputFunc: func(bin string, args []string) (string, error) {
			return fakeMiseDir, nil
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	// Empty mise bin dir — no binaries to symlink. Should still succeed.
	result, err := p.Provision("node", "20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Provisioned {
		t.Error("should report provisioned even with empty bin dir")
	}
}

func TestProvisionWithMise_NonexistentBinDir(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		outputFunc: func(bin string, args []string) (string, error) {
			return "/nonexistent/mise/path", nil
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerMise,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			if name == "mise" {
				return "/usr/local/bin/mise", nil
			}
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("python", "3.13")
	if err == nil {
		t.Fatal("expected error when mise bin directory doesn't exist")
	}
	if !strings.Contains(err.Error(), "reading mise bin directory") {
		t.Errorf("error should mention reading bin dir, got: %v", err)
	}
}

// --- Direct provisioning with mock runner ---

func TestProvisionNodeDirect_RunnerCalled(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			return fmt.Errorf("download failed")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("node", "20")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "downloading node") {
		t.Errorf("error should mention downloading, got: %v", err)
	}

	// Verify the runner was called with bash
	if len(mock.runCalls) < 1 {
		t.Fatal("expected at least 1 run call")
	}
	if mock.runCalls[0].Bin != "bash" {
		t.Errorf("bin = %q, want bash", mock.runCalls[0].Bin)
	}
}

func TestProvisionPythonDirect_RunnerCalled(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			return fmt.Errorf("download failed")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	_, err := p.Provision("python", "3.13")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "downloading python") {
		t.Errorf("error should mention downloading, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mise") {
		t.Errorf("error should suggest mise, got: %v", err)
	}

	if len(mock.runCalls) < 1 {
		t.Fatal("expected at least 1 run call")
	}
	if mock.runCalls[0].Bin != "bash" {
		t.Errorf("bin = %q, want bash", mock.runCalls[0].Bin)
	}
}

func TestProvision_AutoMode_DefaultVersion(t *testing.T) {
	base := t.TempDir()
	var stderr bytes.Buffer

	mock := &mockCmdRunner{
		runErr: func(bin string, args []string) error {
			return fmt.Errorf("fail")
		},
	}

	p := &Provisioner{
		Mode:    config.ProvisionAuto,
		Manager: config.RuntimeManagerDirect,
		BaseDir: base,
		Stderr:  &stderr,
		Runner:  mock,
		LookPath: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	// When version is empty, should use default version
	_, err := p.Provision("node", "")
	if err == nil {
		t.Fatal("expected error (mock fails)")
	}
	// The error will mention "node 20" because that's the default
	if !strings.Contains(err.Error(), "node 20") {
		t.Errorf("error should reference default version (20), got: %v", err)
	}
}

func TestKnownRuntimeList(t *testing.T) {
	list := runtimeinfo.SortedNames()
	for _, rt := range []string{"go", "node", "python", "rust"} {
		if !strings.Contains(list, rt) {
			t.Errorf("SortedNames should contain %q, got: %s", rt, list)
		}
	}
	if !strings.Contains(list, ", ") {
		t.Errorf("should be comma-separated, got: %s", list)
	}
}
