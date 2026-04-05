package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
