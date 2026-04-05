package executor

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestStepDir_BashInline(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	os.Mkdir(sub, 0o755)

	var stdout bytes.Buffer
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd",
		Dir:   "subdir",
	}
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != sub {
		t.Errorf("got %q, want %q", got, sub)
	}
}

func TestStepDir_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	absDir := t.TempDir()

	var stdout bytes.Buffer
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd",
		Dir:   absDir,
	}
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != absDir {
		t.Errorf("got %q, want %q", got, absDir)
	}
}

func TestStepDir_NoDir_UsesRepoRoot(t *testing.T) {
	dir := t.TempDir()

	var stdout bytes.Buffer
	step := bashStep("pwd")
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
	}
}

func TestStepDir_MissingDirectory(t *testing.T) {
	dir := t.TempDir()
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "echo hello",
		Dir:   "nonexistent",
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error should mention 'does not exist', got: %s", err.Error())
	}
}

func TestStepDir_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "afile.txt")
	os.WriteFile(filePath, []byte("hello"), 0o644)

	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "echo hello",
		Dir:   "afile.txt",
	}
	a := newAutomation("test", step)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for non-directory path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error should mention 'not a directory', got: %s", err.Error())
	}
}

func TestStepDir_PythonStep(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	sub := filepath.Join(dir, "pydir")
	os.Mkdir(sub, 0o755)

	var stdout bytes.Buffer
	step := automation.Step{
		Type:  automation.StepTypePython,
		Value: "import os; print(os.getcwd())",
		Dir:   "pydir",
	}
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	wantSub, _ := filepath.EvalSymlinks(sub)
	if got != sub && got != wantSub {
		t.Errorf("got %q, want %q (or %q)", got, sub, wantSub)
	}
}

func TestStepDir_PerStepIsolation(t *testing.T) {
	dir := t.TempDir()
	sub1 := filepath.Join(dir, "dir1")
	sub2 := filepath.Join(dir, "dir2")
	os.Mkdir(sub1, 0o755)
	os.Mkdir(sub2, 0o755)

	outFile1 := filepath.Join(dir, "out1.txt")
	outFile2 := filepath.Join(dir, "out2.txt")

	step1 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd > " + outFile1,
		Dir:   "dir1",
	}
	step2 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd > " + outFile2,
		Dir:   "dir2",
	}
	a := newAutomation("test", step1, step2)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got1, _ := os.ReadFile(outFile1)
	got2, _ := os.ReadFile(outFile2)
	if strings.TrimSpace(string(got1)) != sub1 {
		t.Errorf("step1 dir: got %q, want %q", strings.TrimSpace(string(got1)), sub1)
	}
	if strings.TrimSpace(string(got2)) != sub2 {
		t.Errorf("step2 dir: got %q, want %q", strings.TrimSpace(string(got2)), sub2)
	}
}

func TestStepDir_MixedWithNoDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0o755)

	outFile1 := filepath.Join(dir, "out1.txt")
	outFile2 := filepath.Join(dir, "out2.txt")

	step1 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd > " + outFile1,
		Dir:   "sub",
	}
	step2 := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "pwd > " + outFile2,
	}
	a := newAutomation("test", step1, step2)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got1, _ := os.ReadFile(outFile1)
	got2, _ := os.ReadFile(outFile2)
	if strings.TrimSpace(string(got1)) != sub {
		t.Errorf("step1 (dir:sub): got %q, want %q", strings.TrimSpace(string(got1)), sub)
	}
	if strings.TrimSpace(string(got2)) != dir {
		t.Errorf("step2 (no dir): got %q, want %q", strings.TrimSpace(string(got2)), dir)
	}
}

func TestStepDir_WithEnvAndIf(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "combo")
	os.Mkdir(sub, 0o755)

	var stdout bytes.Buffer
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: `echo "$MY_VAR $(pwd)"`,
		Dir:   "combo",
		Env:   map[string]string{"MY_VAR": "hello"},
	}
	a := newAutomation("test", step)
	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &stdout,
		Stderr:    io.Discard,
	}

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	subResolved, _ := filepath.EvalSymlinks(sub)
	want1 := "hello " + sub
	want2 := "hello " + subResolved
	if got != want1 && got != want2 {
		t.Errorf("got %q, want %q (or %q)", got, want1, want2)
	}
}

func TestResolveStepDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "mydir")
	os.Mkdir(sub, 0o755)

	t.Run("empty returns repo root", func(t *testing.T) {
		got, err := resolveStepDir(dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != dir {
			t.Errorf("got %q, want %q", got, dir)
		}
	})

	t.Run("relative path", func(t *testing.T) {
		got, err := resolveStepDir(dir, "mydir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != sub {
			t.Errorf("got %q, want %q", got, sub)
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		got, err := resolveStepDir(dir, sub)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != sub {
			t.Errorf("got %q, want %q", got, sub)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := resolveStepDir(dir, "nope")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error should mention 'does not exist', got: %s", err.Error())
		}
	})

	t.Run("not a directory", func(t *testing.T) {
		f := filepath.Join(dir, "file.txt")
		os.WriteFile(f, []byte("x"), 0o644)
		_, err := resolveStepDir(dir, "file.txt")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("error should mention 'not a directory', got: %s", err.Error())
		}
	})
}
