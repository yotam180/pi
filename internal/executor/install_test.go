package executor

import (
	"bytes"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestExecInstall_AlreadyInstalled(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test:    automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Run:     automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Version: "echo 1.2.3",
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "already installed") {
		t.Errorf("expected 'already installed', got: %q", output)
	}
	if !strings.Contains(output, "1.2.3") {
		t.Errorf("expected version '1.2.3' in output, got: %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("expected '✓' icon, got: %q", output)
	}
}

func TestExecInstall_FreshInstall(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "tool-installed")
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test:    automation.InstallPhase{IsScalar: true, Scalar: "test -f " + marker},
		Run:     automation.InstallPhase{IsScalar: true, Scalar: "touch " + marker},
		Version: "echo 2.0.0",
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "installing...") {
		t.Errorf("expected 'installing...' status, got: %q", output)
	}
	if !strings.Contains(output, "installed") {
		t.Errorf("expected 'installed' status, got: %q", output)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Errorf("expected version '2.0.0', got: %q", output)
	}
}

func TestExecInstall_RunFails(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run:  automation.InstallPhase{IsScalar: true, Scalar: "echo 'install error' >&2; exit 1"},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed run")
	}
	output := stderr.String()
	if !strings.Contains(output, "failed") {
		t.Errorf("expected 'failed' status, got: %q", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected '✗' icon, got: %q", output)
	}
	if !strings.Contains(output, "install error") {
		t.Errorf("expected stderr from run in output, got: %q", output)
	}
}

func TestExecInstall_VerifyFails(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	verifyPhase := automation.InstallPhase{IsScalar: true, Scalar: "exit 1"}
	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test:   automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run:    automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Verify: &verifyPhase,
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed verify")
	}
	output := stderr.String()
	if !strings.Contains(output, "failed") {
		t.Errorf("expected 'failed' status, got: %q", output)
	}
}

func TestExecInstall_VerifyDefaultsToTest(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "installed")
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test:    automation.InstallPhase{IsScalar: true, Scalar: "test -f " + marker},
		Run:     automation.InstallPhase{IsScalar: true, Scalar: "touch " + marker},
		Version: "echo 1.0.0",
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "installed") && !strings.Contains(output, "✓") {
		t.Errorf("expected successful install, got: %q", output)
	}
}

func TestExecInstall_NoVersion(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Run:  automation.InstallPhase{IsScalar: true, Scalar: "true"},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if strings.Contains(output, "(") {
		t.Errorf("expected no version parenthetical, got: %q", output)
	}
}

func TestExecInstall_Silent(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test:    automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Run:     automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Version: "echo 1.0.0",
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
		Silent:    true,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stderr.String() != "" {
		t.Errorf("expected no output in silent mode, got: %q", stderr.String())
	}
}

func TestExecInstall_SilentStillShowsStderrOnFailure(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run:  automation.InstallPhase{IsScalar: true, Scalar: "echo 'error msg' >&2; exit 1"},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
		Silent:    true,
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed run")
	}
	output := stderr.String()
	if !strings.Contains(output, "error msg") {
		t.Errorf("expected stderr from run to be streamed even in silent mode, got: %q", output)
	}
}

func TestExecInstall_RunFailsScalarStderrStreamed(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run:  automation.InstallPhase{IsScalar: true, Scalar: "echo 'detailed failure reason' >&2; exit 1"},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed run")
	}
	output := stderr.String()
	if !strings.Contains(output, "detailed failure reason") {
		t.Errorf("expected stderr from scalar run phase to be streamed, got: %q", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected '✗' icon, got: %q", output)
	}
}

func TestExecInstall_RunFailsStepListStderrStreamed(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run: automation.InstallPhase{
			IsScalar: false,
			Steps: []automation.Step{
				{Type: automation.StepTypeBash, Value: "echo 'step list failure' >&2; exit 1"},
			},
		},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed run")
	}
	output := stderr.String()
	if !strings.Contains(output, "step list failure") {
		t.Errorf("expected stderr from step list run phase to be streamed, got: %q", output)
	}
}

func TestExecInstall_StepListWithConditionals(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "tool-installed")
	var stderr bytes.Buffer

	a := newInstallerAutomation("test-tool", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "test -f " + marker},
		Run: automation.InstallPhase{
			IsScalar: false,
			Steps: []automation.Step{
				{Type: automation.StepTypeBash, Value: "exit 0", If: "os.windows"},
				{Type: automation.StepTypeBash, Value: "touch " + marker},
			},
		},
		Version: "echo 3.0.0",
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
		RuntimeEnv: &RuntimeEnv{
			GOOS:     "darwin",
			GOARCH:   "arm64",
			Getenv:   func(string) string { return "" },
			LookPath: func(string) (string, error) { return "", osexec.ErrNotFound },
			Stat:     func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		},
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "installed") {
		t.Errorf("expected 'installed' status, got: %q", output)
	}
}

func TestExecInstall_WithInputs(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "python-installed")
	var stderr bytes.Buffer

	a := &automation.Automation{
		Name: "install-python",
		Install: &automation.InstallSpec{
			Test:    automation.InstallPhase{IsScalar: true, Scalar: "test -f " + marker},
			Run:     automation.InstallPhase{IsScalar: true, Scalar: "touch " + marker},
			Version: "echo $PI_INPUT_VERSION",
		},
		Inputs:    map[string]automation.InputSpec{"version": {Type: "string"}},
		InputKeys: []string{"version"},
		FilePath:  "/fake/path/automation.yaml",
	}

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.RunWithInputs(a, nil, map[string]string{"version": "3.13"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "installed") {
		t.Errorf("expected 'installed' status, got: %q", output)
	}
	if !strings.Contains(output, "3.13") {
		t.Errorf("expected version '3.13' in output, got: %q", output)
	}
}

func TestExecInstall_FirstBlockFailStderrSurfaced(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("install-node", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "false"},
		Run: automation.InstallPhase{
			IsScalar: false,
			Steps: []automation.Step{
				{
					Type: automation.StepTypeBash,
					First: []automation.Step{
						{Type: automation.StepTypeBash, Value: "echo 'mise path' >&2; exit 1", If: "command.mise"},
						{Type: automation.StepTypeBash, Value: "echo 'brew path' >&2; exit 1", If: "command.brew"},
						{Type: automation.StepTypeBash, Value: "echo 'no suitable installer found (tried mise, brew)' >&2\nexit 1"},
					},
				},
			},
		},
	})

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
		RuntimeEnv: &RuntimeEnv{
			GOOS:     "darwin",
			GOARCH:   "arm64",
			Getenv:   func(string) string { return "" },
			LookPath: func(string) (string, error) { return "", osexec.ErrNotFound },
			Stat:     func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		},
	}

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error from failed install")
	}
	output := stderr.String()
	if !strings.Contains(output, "✗") {
		t.Errorf("expected '✗' icon, got: %q", output)
	}
	if !strings.Contains(output, "no suitable installer found") {
		t.Errorf("expected stderr from first: block fallback to be surfaced, got: %q", output)
	}
}

func TestExecInstall_ScalarPhaseUsesAutomationEnv(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "env-marker")
	var stderr bytes.Buffer

	a := &automation.Automation{
		Name: "install-with-env",
		Env:  map[string]string{"MY_INSTALL_VAR": "hello-from-env"},
		Install: &automation.InstallSpec{
			Test:    automation.InstallPhase{IsScalar: true, Scalar: "test -f " + marker},
			Run:     automation.InstallPhase{IsScalar: true, Scalar: "echo $MY_INSTALL_VAR > " + marker},
			Version: "echo 1.0.0",
		},
		FilePath: "/fake/path/automation.yaml",
	}

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, readErr := os.ReadFile(marker)
	if readErr != nil {
		t.Fatalf("marker file not created: %v", readErr)
	}
	if got := strings.TrimSpace(string(content)); got != "hello-from-env" {
		t.Errorf("expected 'hello-from-env' in marker, got: %q", got)
	}
}

func TestExecInstall_VersionCaptureUsesAutomationEnv(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := &automation.Automation{
		Name: "install-version-env",
		Env:  map[string]string{"MY_VERSION": "42.0.0"},
		Install: &automation.InstallSpec{
			Test:    automation.InstallPhase{IsScalar: true, Scalar: "true"},
			Run:     automation.InstallPhase{IsScalar: true, Scalar: "true"},
			Version: "echo $MY_VERSION",
		},
		FilePath: "/fake/path/automation.yaml",
	}

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "42.0.0") {
		t.Errorf("expected version '42.0.0' from automation env, got: %q", output)
	}
}

func TestExecInstall_WithAutomationLevelIf(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer

	a := newInstallerAutomation("install-brew", &automation.InstallSpec{
		Test: automation.InstallPhase{IsScalar: true, Scalar: "true"},
		Run:  automation.InstallPhase{IsScalar: true, Scalar: "true"},
	})
	a.If = "os.windows"

	e := &Executor{
		RepoRoot:  root,
		Discovery: newDiscovery(nil),
		Stdout:    io.Discard,
		Stderr:    &stderr,
		RuntimeEnv: &RuntimeEnv{
			GOOS:     "darwin",
			GOARCH:   "arm64",
			Getenv:   func(string) string { return "" },
			LookPath: func(string) (string, error) { return "", osexec.ErrNotFound },
			Stat:     func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		},
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stderr.String()
	if !strings.Contains(output, "[skipped]") {
		t.Errorf("expected '[skipped]' message, got: %q", output)
	}
}
