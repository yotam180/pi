package executor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPythonInline_Success(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := `with open("` + outFile + `", "w") as f: f.write("hello\n")`
	a := newAutomation("test", pythonStep(script))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "hello" {
		t.Errorf("output = %q, want %q", got, "hello")
	}
}

func TestPythonInline_Failure(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	a := newAutomation("test", pythonStep("import sys; sys.exit(42)"))
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

func TestPythonInline_WithArgs(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := `import sys; open("` + outFile + `", "w").write(" ".join(sys.argv[1:]) + "\n")`
	a := newAutomation("test", pythonStep(script))
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

func TestPythonFile_Success(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptDir := filepath.Join(dir, ".pi", "test")
	os.MkdirAll(scriptDir, 0755)
	scriptPath := filepath.Join(scriptDir, "run.py")
	os.WriteFile(scriptPath, []byte(`with open("`+outFile+`", "w") as f: f.write("from-file\n")`+"\n"), 0755)

	a := newAutomationInDir("test", scriptDir, pythonStep("run.py"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "from-file" {
		t.Errorf("output = %q, want %q", got, "from-file")
	}
}

func TestPythonFile_WithArgs(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptPath := filepath.Join(dir, "greet.py")
	os.WriteFile(scriptPath, []byte("import sys\nwith open(\""+outFile+"\", \"w\") as f: f.write(\" \".join(sys.argv[1:]) + \"\\n\")\n"), 0755)

	a := newAutomationInDir("test", dir, pythonStep("greet.py"))
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

func TestPythonFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationInDir("test", dir, pythonStep("nonexistent.py"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for missing script file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestPythonInline_Multiline(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := "import os\nx = 'multiline'\nwith open('" + outFile + "', 'w') as f:\n    f.write(x + '\\n')"
	a := newAutomation("test", pythonStep(script))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "multiline" {
		t.Errorf("output = %q, want %q", got, "multiline")
	}
}

func TestPythonVenvDetection(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	venvDir := filepath.Join(dir, "fakevenv")
	venvBinDir := filepath.Join(venvDir, "bin")
	os.MkdirAll(venvBinDir, 0755)

	fakePython := filepath.Join(venvBinDir, "python")
	os.WriteFile(fakePython, []byte("#!/bin/bash\n"+
		"echo venv-used > "+outFile+"\n"+
		"exec python3 \"$@\"\n"), 0755)

	t.Setenv("VIRTUAL_ENV", venvDir)

	a := newAutomation("test", pythonStep("print('ok')"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "venv-used" {
		t.Errorf("expected venv python to be used, marker file contains %q", got)
	}
}

func TestMixedSteps_BashAndPython(t *testing.T) {
	requirePython(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	a := newAutomation("test",
		bashStep("echo from-bash >> "+outFile),
		pythonStep(`with open("`+outFile+`", "a") as f: f.write("from-python\n")`),
		bashStep("echo from-bash-again >> "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"from-bash", "from-python", "from-bash-again"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(lines), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], w)
		}
	}
}
