package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletionHelp(t *testing.T) {
	out, err := executeCmd("completion", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "completion") {
		t.Errorf("expected help to mention completion, got: %s", out)
	}
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		if !strings.Contains(out, shell) {
			t.Errorf("expected help to mention %q, got: %s", shell, out)
		}
	}
}

func TestCompletionRequiresArg(t *testing.T) {
	_, err := executeCmd("completion")
	if err == nil {
		t.Fatal("expected error when no shell specified")
	}
}

func TestCompletionBash(t *testing.T) {
	out, err := executeCmd("completion", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bash completion") {
		t.Errorf("expected bash completion script, got: %s", out[:min(len(out), 200)])
	}
}

func TestCompletionZsh(t *testing.T) {
	out, err := executeCmd("completion", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "compdef") {
		t.Errorf("expected zsh compdef in completion script, got: %s", out[:min(len(out), 200)])
	}
}

func TestCompletionFish(t *testing.T) {
	out, err := executeCmd("completion", "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "fish") {
		t.Errorf("expected fish completion script, got: %s", out[:min(len(out), 200)])
	}
}

func TestCompletionPowershell(t *testing.T) {
	out, err := executeCmd("completion", "powershell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "powershell") || !strings.Contains(out, "Register") {
		t.Errorf("expected powershell completion script, got: %s", out[:min(len(out), 200)])
	}
}

func TestAutomationCompleter_ReturnsAutomations(t *testing.T) {
	tmp := t.TempDir()
	piDir := filepath.Join(tmp, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(tmp, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: Say hello\nbash: echo hello\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte("bash: go build ./...\n"), 0o644)

	completer := automationCompleter()
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	completions, _ := completer(nil, nil, "")

	if len(completions) < 2 {
		t.Fatalf("expected at least 2 completions, got %d: %v", len(completions), completions)
	}

	var found []string
	for _, c := range completions {
		name := strings.SplitN(c, "\t", 2)[0]
		found = append(found, name)
	}

	hasHello := false
	hasBuild := false
	for _, n := range found {
		if n == "hello" {
			hasHello = true
		}
		if n == "build" {
			hasBuild = true
		}
	}
	if !hasHello || !hasBuild {
		t.Errorf("expected hello and build in completions, got: %v", found)
	}
}

func TestAutomationCompleter_IncludesDescription(t *testing.T) {
	tmp := t.TempDir()
	piDir := filepath.Join(tmp, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(tmp, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("description: Say hello\nbash: echo hello\n"), 0o644)

	completer := automationCompleter()
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	completions, _ := completer(nil, nil, "")

	hasDesc := false
	for _, c := range completions {
		if strings.Contains(c, "\tSay hello") {
			hasDesc = true
		}
	}
	if !hasDesc {
		t.Errorf("expected completion with description tab, got: %v", completions)
	}
}

func TestAutomationCompleter_ExcludesBuiltins(t *testing.T) {
	tmp := t.TempDir()
	piDir := filepath.Join(tmp, ".pi")
	os.MkdirAll(piDir, 0o755)
	os.WriteFile(filepath.Join(tmp, "pi.yaml"), []byte("project: test\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte("bash: echo hello\n"), 0o644)

	completer := automationCompleter()
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	completions, _ := completer(nil, nil, "")

	for _, c := range completions {
		name := strings.SplitN(c, "\t", 2)[0]
		if strings.HasPrefix(name, "pi:") {
			t.Errorf("builtins should be excluded from completion, found: %s", name)
		}
	}
}

func TestAutomationCompleter_NoArgsAfterFirst(t *testing.T) {
	completer := automationCompleter()
	completions, _ := completer(nil, []string{"some-automation"}, "")
	if len(completions) != 0 {
		t.Errorf("expected no completions after first arg, got: %v", completions)
	}
}

func TestAutomationCompleter_GracefulOnMissingProject(t *testing.T) {
	tmp := t.TempDir()

	completer := automationCompleter()
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	completions, _ := completer(nil, nil, "")
	if len(completions) != 0 {
		t.Errorf("expected empty completions for missing project, got: %v", completions)
	}
}
