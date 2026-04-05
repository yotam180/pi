package executor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestBashInline_Success(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", bashStep("echo hello"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBashInline_WithArgs(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := `echo "$1 $2" > ` + outFile
	a := newAutomation("test", bashStep(script))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, []string{"hello", "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "hello world" {
		t.Errorf("output = %q, want %q", got, "hello world")
	}
}

func TestBashInline_Failure(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", bashStep("exit 42"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 42 {
		t.Errorf("exit code = %d, want 42", exitErr.Code)
	}
}

func TestBashInline_Multiline(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := "VAR=hello\necho $VAR > " + outFile
	a := newAutomation("test", bashStep(script))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "hello" {
		t.Errorf("output = %q, want %q", got, "hello")
	}
}

func TestBashFile_Success(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptDir := filepath.Join(dir, ".pi", "test")
	os.MkdirAll(scriptDir, 0755)
	scriptPath := filepath.Join(scriptDir, "run.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho file-step > "+outFile+"\n"), 0755)

	a := newAutomationInDir("test", scriptDir, bashStep("run.sh"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "file-step" {
		t.Errorf("output = %q, want %q", got, "file-step")
	}
}

func TestBashFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationInDir("test", dir, bashStep("nonexistent.sh"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for missing script file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestBashFile_WithArgs(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptPath := filepath.Join(dir, "greet.sh")
	os.WriteFile(scriptPath, []byte("#!/bin/bash\necho \"$1 $2\" > "+outFile+"\n"), 0755)

	a := newAutomationInDir("test", dir, bashStep("greet.sh"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, []string{"hi", "there"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "hi there" {
		t.Errorf("output = %q, want %q", got, "hi there")
	}
}

func TestRunStep_Chaining(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	inner := newAutomation("inner", bashStep("echo inner-ran >> "+outFile))
	outer := newAutomation("outer", runStep("inner"))

	disc := newDiscovery(map[string]*automation.Automation{
		"inner": inner,
		"outer": outer,
	})
	exec := newExecutor(dir, disc)

	err := exec.Run(outer, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "inner-ran" {
		t.Errorf("output = %q, want %q", got, "inner-ran")
	}
}

func TestRunStep_DeepChaining(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	c := newAutomation("c", bashStep("echo c >> "+outFile))
	b := newAutomation("b", runStep("c"))
	a := newAutomation("a", runStep("b"))

	disc := newDiscovery(map[string]*automation.Automation{
		"a": a,
		"b": b,
		"c": c,
	})
	exec := newExecutor(dir, disc)

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "c" {
		t.Errorf("output = %q, want %q", got, "c")
	}
}

func TestRunStep_ArgsForwarded(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	inner := newAutomation("inner", bashStep(`echo "$1" > `+outFile))
	outer := newAutomation("outer", runStep("inner"))

	disc := newDiscovery(map[string]*automation.Automation{
		"inner": inner,
		"outer": outer,
	})
	exec := newExecutor(dir, disc)

	err := exec.Run(outer, []string{"forwarded-arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "forwarded-arg" {
		t.Errorf("output = %q, want %q", got, "forwarded-arg")
	}
}

func TestRunStep_NotFound(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", runStep("nonexistent"))

	disc := newDiscovery(map[string]*automation.Automation{})
	exec := newExecutor(dir, disc)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for unknown run: target")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestCircularDependency_Direct(t *testing.T) {
	a := newAutomation("a", runStep("a"))

	disc := newDiscovery(map[string]*automation.Automation{
		"a": a,
	})
	exec := newExecutor(t.TempDir(), disc)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention 'circular', got: %v", err)
	}
}

func TestCircularDependency_Indirect(t *testing.T) {
	a := newAutomation("a", runStep("b"))
	b := newAutomation("b", runStep("c"))
	c := newAutomation("c", runStep("a"))

	disc := newDiscovery(map[string]*automation.Automation{
		"a": a,
		"b": b,
		"c": c,
	})
	exec := newExecutor(t.TempDir(), disc)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention 'circular', got: %v", err)
	}
	if !strings.Contains(err.Error(), "a → b → c → a") {
		t.Errorf("error should show the chain, got: %v", err)
	}
}

func TestMultipleSteps_StopsOnFailure(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	a := newAutomation("test",
		bashStep("echo step1 >> "+outFile),
		bashStep("exit 1"),
		bashStep("echo step3 >> "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "step1" {
		t.Errorf("output = %q, want only %q (step3 should not run)", got, "step1")
	}
}

func TestMultipleSteps_AllSucceed(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	a := newAutomation("test",
		bashStep("echo step1 >> "+outFile),
		bashStep("echo step2 >> "+outFile),
		bashStep("echo step3 >> "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	for i, want := range []string{"step1", "step2", "step3"} {
		if lines[i] != want {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

func TestWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	a := newAutomation("test", bashStep("pwd > "+outFile))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))

	// Resolve symlinks to handle macOS /private/var/folders... -> /var/folders...
	resolvedDir, _ := filepath.EvalSymlinks(dir)
	resolvedGot, _ := filepath.EvalSymlinks(got)
	if resolvedGot != resolvedDir {
		t.Errorf("working dir = %q, want %q", resolvedGot, resolvedDir)
	}
}

func TestMixedSteps_BashAndRun(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	inner := newAutomation("inner", bashStep("echo from-inner >> "+outFile))
	outer := newAutomation("outer",
		bashStep("echo from-outer >> "+outFile),
		runStep("inner"),
		bashStep("echo from-outer-again >> "+outFile),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"inner": inner,
		"outer": outer,
	})
	exec := newExecutor(dir, disc)

	err := exec.Run(outer, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"from-outer", "from-inner", "from-outer-again"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(lines), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], w)
		}
	}
}

func TestExitError_Message(t *testing.T) {
	e := &ExitError{Code: 42}
	if !strings.Contains(e.Error(), "42") {
		t.Errorf("error message should contain exit code, got: %v", e)
	}
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"run.sh", true},
		{"scripts/deploy.sh", true},
		{"run.py", true},
		{"scripts/format.py", true},
		{"run.ts", true},
		{"scripts/build.ts", true},
		{"echo hello", false},
		{"echo hello.sh", false},
		{"line1\nline2.sh", false},
		{"docker-compose up -d", false},
		{"import sys; print('hello')", false},
		{"print('hello.py')", false},
	}
	for _, tt := range tests {
		got := IsFilePath(tt.input)
		if got != tt.want {
			t.Errorf("IsFilePath(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestCallStackIsolation(t *testing.T) {
	dir := t.TempDir()

	a := newAutomation("a", bashStep("echo a"))
	b := newAutomation("b", bashStep("echo b"))

	disc := newDiscovery(map[string]*automation.Automation{
		"a": a,
		"b": b,
	})
	exec := newExecutor(dir, disc)

	if err := exec.Run(a, nil); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := exec.Run(b, nil); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(exec.callStack) != 0 {
		t.Errorf("call stack should be empty after runs, got %v", exec.callStack)
	}
}
