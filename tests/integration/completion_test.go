package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletion_Bash(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "completion", "bash")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "bash completion") {
		t.Errorf("expected bash completion script, got:\n%s", out[:min(len(out), 200)])
	}
}

func TestCompletion_Zsh(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "completion", "zsh")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "compdef") {
		t.Errorf("expected zsh compdef in output, got:\n%s", out[:min(len(out), 200)])
	}
}

func TestCompletion_Fish(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "completion", "fish")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "fish") {
		t.Errorf("expected fish completion script, got:\n%s", out[:min(len(out), 200)])
	}
}

func TestCompletion_NoArg(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	_, code := runPi(t, dir, "completion")
	if code == 0 {
		t.Fatal("expected non-zero exit for missing shell argument")
	}
}

func TestCompletion_WorksWithoutPiYaml(t *testing.T) {
	tmp := t.TempDir()
	out, code := runPi(t, tmp, "completion", "zsh")
	if code != 0 {
		t.Fatalf("expected exit 0 even without pi.yaml, got %d: %s", code, out)
	}
	if !strings.Contains(out, "compdef") {
		t.Errorf("expected valid zsh completion script, got:\n%s", out[:min(len(out), 200)])
	}
}

func TestShell_InstallCreatesCompletion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	tmpHome := t.TempDir()
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)

	out, code := runPiWithHome(t, dir, tmpHome, "shell")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	completionFile := filepath.Join(tmpHome, ".pi", "shell", "_pi-completion.sh")
	data, err := os.ReadFile(completionFile)
	if err != nil {
		t.Fatalf("completion file should be created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "completion zsh") {
		t.Error("completion file should set up zsh completion")
	}
	if !strings.Contains(content, "completion bash") {
		t.Error("completion file should set up bash completion")
	}
}

func TestCompletion_DynamicRunCompletion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")

	out, code := runPi(t, dir, "__complete", "run", "")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "greet") {
		t.Errorf("expected 'greet' in run completions, got:\n%s", out)
	}
	if !strings.Contains(out, "deploy") {
		t.Errorf("expected 'deploy' in run completions, got:\n%s", out)
	}
}

func TestCompletion_DynamicInfoCompletion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")

	out, code := runPi(t, dir, "__complete", "info", "")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if !strings.Contains(out, "greet") {
		t.Errorf("expected 'greet' in info completions, got:\n%s", out)
	}
}

func TestCompletion_ExcludesBuiltins(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")

	out, code := runPi(t, dir, "__complete", "run", "")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "pi:") {
			t.Errorf("builtins should not appear in completion, found: %s", line)
		}
	}
}
