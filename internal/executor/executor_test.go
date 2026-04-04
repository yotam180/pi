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
	"github.com/vyper-tooling/pi/internal/discovery"
)

func newAutomation(name string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		Steps:    steps,
		FilePath: "/fake/path/automation.yaml",
	}
}

func newAutomationInDir(name, dir string, steps ...automation.Step) *automation.Automation {
	return &automation.Automation{
		Name:     name,
		Steps:    steps,
		FilePath: filepath.Join(dir, "automation.yaml"),
	}
}

func bashStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeBash, Value: value}
}

func runStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeRun, Value: value}
}

func pythonStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypePython, Value: value}
}

func typescriptStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeTypeScript, Value: value}
}

func newDiscovery(automations map[string]*automation.Automation) *discovery.Result {
	return &discovery.Result{Automations: automations}
}

func newExecutor(repoRoot string, disc *discovery.Result) *Executor {
	return &Executor{
		RepoRoot:  repoRoot,
		Discovery: disc,
		Stdout:    io.Discard,
		Stderr:    io.Discard,
	}
}

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

	exitErr, ok := err.(*ExitError)
	if !ok {
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

// --- Python step tests ---

func TestPythonInline_Success(t *testing.T) {
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
	dir := t.TempDir()
	a := newAutomation("test", pythonStep("import sys; sys.exit(42)"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 42 {
		t.Errorf("exit code = %d, want 42", exitErr.Code)
	}
}

func TestPythonInline_WithArgs(t *testing.T) {
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
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	// Create a fake venv with a python script that acts as the interpreter
	venvDir := filepath.Join(dir, "fakevenv")
	venvBinDir := filepath.Join(venvDir, "bin")
	os.MkdirAll(venvBinDir, 0755)

	// Create a fake python that writes a marker so we know the venv was used
	fakePython := filepath.Join(venvBinDir, "python")
	// The fake python is a bash script that runs real python3 but also writes a marker
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
		got := isFilePath(tt.input)
		if got != tt.want {
			t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.want)
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

// --- TypeScript step tests ---

func requireTsx(t *testing.T) {
	t.Helper()
	if _, err := osexec.LookPath("tsx"); err != nil {
		t.Skip("tsx not found in PATH, skipping TypeScript test")
	}
}

func TestTypeScriptInline_Success(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := `import { writeFileSync } from "fs"; writeFileSync("` + outFile + `", "hello\n");`
	a := newAutomation("test", typescriptStep(script))
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

func TestTypeScriptInline_Failure(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	a := newAutomation("test", typescriptStep("process.exit(42);"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 42 {
		t.Errorf("exit code = %d, want 42", exitErr.Code)
	}
}

func TestTypeScriptInline_WithArgs(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	script := `import { writeFileSync } from "fs"; writeFileSync("` + outFile + `", process.argv.slice(2).join(" ") + "\n");`
	a := newAutomation("test", typescriptStep(script))
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

func TestTypeScriptFile_Success(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptDir := filepath.Join(dir, ".pi", "test")
	os.MkdirAll(scriptDir, 0755)
	scriptPath := filepath.Join(scriptDir, "run.ts")
	os.WriteFile(scriptPath, []byte(`import { writeFileSync } from "fs"; writeFileSync("`+outFile+`", "from-file\n");`+"\n"), 0644)

	a := newAutomationInDir("test", scriptDir, typescriptStep("run.ts"))
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

func TestTypeScriptFile_WithArgs(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	scriptPath := filepath.Join(dir, "greet.ts")
	os.WriteFile(scriptPath, []byte(`import { writeFileSync } from "fs"; writeFileSync("`+outFile+`", process.argv.slice(2).join(" ") + "\n");`+"\n"), 0644)

	a := newAutomationInDir("test", dir, typescriptStep("greet.ts"))
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

func TestTypeScriptFile_NotFound(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	a := newAutomationInDir("test", dir, typescriptStep("nonexistent.ts"))
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for missing script file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestTypeScriptTsxNotFound(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", typescriptStep("console.log('hello');"))

	exec := newExecutor(dir, newDiscovery(nil))
	// Override PATH to ensure tsx cannot be found
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error when tsx is not found")
	}
	if !strings.Contains(err.Error(), "tsx not found") {
		t.Errorf("error should mention 'tsx not found', got: %v", err)
	}
}

func TestMixedSteps_BashAndTypeScript(t *testing.T) {
	requireTsx(t)
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")

	a := newAutomation("test",
		bashStep("echo from-bash >> "+outFile),
		typescriptStep(`import { appendFileSync } from "fs"; appendFileSync("`+outFile+`", "from-ts\n");`),
		bashStep("echo from-bash-again >> "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"from-bash", "from-ts", "from-bash-again"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(lines), lines)
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], w)
		}
	}
}

// --- Pipe support tests ---

func pipedBashStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypeBash, Value: value, PipeTo: "next"}
}

func pipedPythonStep(value string) automation.Step {
	return automation.Step{Type: automation.StepTypePython, Value: value, PipeTo: "next"}
}

func newExecutorWithCapture(repoRoot string, disc *discovery.Result) (*Executor, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return &Executor{
		RepoRoot:  repoRoot,
		Discovery: disc,
		Stdout:    stdout,
		Stderr:    stderr,
	}, stdout, stderr
}

func TestPipe_BashToBash(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo hello world"),
		bashStep("tr a-z A-Z"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO WORLD" {
		t.Errorf("output = %q, want %q", got, "HELLO WORLD")
	}
}

func TestPipe_BashToPython(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo hello world"),
		pythonStep("import sys; print(sys.stdin.read().strip().upper())"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO WORLD" {
		t.Errorf("output = %q, want %q", got, "HELLO WORLD")
	}
}

func TestPipe_ThreeStepChain(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo apple banana cherry"),
		pipedBashStep("tr ' ' '\\n'"),
		bashStep("sort"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	want := "apple\nbanana\ncherry"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPipe_FailureInMiddleStopsExecution(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	a := newAutomation("test",
		pipedBashStep("echo data"),
		pipedBashStep("exit 1"),
		bashStep("cat > "+outFile),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error when piped step fails")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.Code)
	}

	if _, err := os.Stat(outFile); err == nil {
		t.Error("third step should not have run, but output file exists")
	}
}

func TestPipe_StderrPassesThrough(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo stdout-data; echo stderr-data >&2"),
		bashStep("cat"),
	)
	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotStdout := strings.TrimSpace(stdout.String())
	if gotStdout != "stdout-data" {
		t.Errorf("stdout = %q, want %q", gotStdout, "stdout-data")
	}

	gotStderr := strings.TrimSpace(stderr.String())
	if gotStderr != "stderr-data" {
		t.Errorf("stderr = %q, want %q", gotStderr, "stderr-data")
	}
}

func TestPipe_NoPipeDefaultBehavior(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo step1"),
		bashStep("echo step2"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	want := "step1\nstep2"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestPipe_LastStepPipeToNextIsNoop(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep("echo piped-first"),
		automation.Step{Type: automation.StepTypeBash, Value: "echo last-step", PipeTo: "next"},
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "last-step" {
		t.Errorf("output = %q, want %q", got, "last-step")
	}
}

func TestPipe_PythonToBash(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedPythonStep("print('hello from python')"),
		bashStep("tr a-z A-Z"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO FROM PYTHON" {
		t.Errorf("output = %q, want %q", got, "HELLO FROM PYTHON")
	}
}

func TestPipe_ThroughRunStep(t *testing.T) {
	dir := t.TempDir()
	inner := newAutomation("upper", bashStep("tr a-z A-Z"))
	outer := newAutomation("test",
		pipedBashStep("echo hello"),
		runStep("upper"),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"upper": inner,
		"test":  outer,
	})
	exec, stdout, _ := newExecutorWithCapture(dir, disc)

	err := exec.Run(outer, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "HELLO" {
		t.Errorf("output = %q, want %q", got, "HELLO")
	}
}

func TestPipe_MultilineData(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		pipedBashStep(`printf "line1\nline2\nline3\n"`),
		bashStep("wc -l | tr -d ' '"),
	)
	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(stdout.String())
	if got != "3" {
		t.Errorf("output = %q, want %q", got, "3")
	}
}
