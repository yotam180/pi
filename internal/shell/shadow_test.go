package shell

import (
	"strings"
	"testing"
)

func TestCheckShadowedNames_NoShadows(t *testing.T) {
	warnings := CheckShadowedNames([]string{"vpup", "vpdown", "myapp"})
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestCheckShadowedNames_ShellBuiltin(t *testing.T) {
	warnings := CheckShadowedNames([]string{"test", "vpup"})
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Name != "test" {
		t.Errorf("expected name 'test', got %q", warnings[0].Name)
	}
	if warnings[0].Kind != "shell builtin" {
		t.Errorf("expected kind 'shell builtin', got %q", warnings[0].Kind)
	}
}

func TestCheckShadowedNames_CommonCommand(t *testing.T) {
	warnings := CheckShadowedNames([]string{"git", "vpup"})
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Name != "git" {
		t.Errorf("expected name 'git', got %q", warnings[0].Name)
	}
	if warnings[0].Kind != "common command" {
		t.Errorf("expected kind 'common command', got %q", warnings[0].Kind)
	}
}

func TestCheckShadowedNames_Multiple(t *testing.T) {
	warnings := CheckShadowedNames([]string{"test", "echo", "ls", "vpup"})
	if len(warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d", len(warnings))
	}
	// Sorted alphabetically
	if warnings[0].Name != "echo" {
		t.Errorf("expected first warning 'echo', got %q", warnings[0].Name)
	}
	if warnings[1].Name != "ls" {
		t.Errorf("expected second warning 'ls', got %q", warnings[1].Name)
	}
	if warnings[2].Name != "test" {
		t.Errorf("expected third warning 'test', got %q", warnings[2].Name)
	}
}

func TestCheckShadowedNames_Empty(t *testing.T) {
	warnings := CheckShadowedNames(nil)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for nil input, got %d", len(warnings))
	}
}

func TestCheckShadowedNames_CaseInsensitive(t *testing.T) {
	warnings := CheckShadowedNames([]string{"TEST", "Echo"})
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(warnings))
	}
}

func TestCheckShadowedNames_AllBuiltins(t *testing.T) {
	builtins := []string{
		"alias", "bg", "cd", "command", "echo", "eval", "exec", "exit",
		"export", "false", "fg", "getopts", "hash", "jobs", "kill", "local",
		"printf", "pwd", "read", "readonly", "return", "set", "shift",
		"source", "test", "time", "times", "trap", "true", "type",
		"ulimit", "umask", "unalias", "unset", "wait", "which",
	}
	for _, b := range builtins {
		warnings := CheckShadowedNames([]string{b})
		if len(warnings) != 1 {
			t.Errorf("expected %q to be detected as shell builtin", b)
			continue
		}
		if warnings[0].Kind != "shell builtin" {
			t.Errorf("expected %q kind 'shell builtin', got %q", b, warnings[0].Kind)
		}
	}
}

func TestCheckShadowedNames_AllCommonCommands(t *testing.T) {
	commands := []string{
		"cat", "cp", "curl", "chmod", "chown", "diff", "find", "git",
		"grep", "head", "less", "ln", "ls", "make", "man", "mkdir",
		"mv", "rm", "rmdir", "run", "sed", "sort", "ssh", "sudo",
		"tail", "tar", "touch", "wc", "wget", "xargs",
	}
	for _, c := range commands {
		warnings := CheckShadowedNames([]string{c})
		if len(warnings) != 1 {
			t.Errorf("expected %q to be detected as common command", c)
			continue
		}
		if warnings[0].Kind != "common command" {
			t.Errorf("expected %q kind 'common command', got %q", c, warnings[0].Kind)
		}
	}
}

func TestCheckShadowedNames_Suggestion(t *testing.T) {
	warnings := CheckShadowedNames([]string{"test"})
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Suggestion == "" {
		t.Error("expected a suggestion for shadowed name")
	}
	if !strings.HasPrefix(warnings[0].Suggestion, "pi-") {
		t.Errorf("expected suggestion starting with 'pi-', got %q", warnings[0].Suggestion)
	}
}

func TestFormatWarning_Builtin(t *testing.T) {
	w := ShadowWarning{Name: "test", Kind: "shell builtin", Suggestion: "pi-test"}
	msg := FormatWarning(w)
	if !strings.Contains(msg, `shortcut "test"`) {
		t.Errorf("expected shortcut name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "shell builtin") {
		t.Errorf("expected kind in message, got: %s", msg)
	}
	if !strings.Contains(msg, `"pi-test"`) {
		t.Errorf("expected suggestion in message, got: %s", msg)
	}
}

func TestFormatWarning_Command(t *testing.T) {
	w := ShadowWarning{Name: "git", Kind: "common command", Suggestion: "pi-git"}
	msg := FormatWarning(w)
	if !strings.Contains(msg, `shortcut "git"`) {
		t.Errorf("expected shortcut name in message, got: %s", msg)
	}
	if !strings.Contains(msg, "common command") {
		t.Errorf("expected kind in message, got: %s", msg)
	}
}

func TestFormatWarning_NoSuggestion(t *testing.T) {
	w := ShadowWarning{Name: "test", Kind: "shell builtin", Suggestion: ""}
	msg := FormatWarning(w)
	if strings.Contains(msg, "consider renaming") {
		t.Errorf("should not suggest renaming when no suggestion, got: %s", msg)
	}
}

func TestCheckShadowedNames_SortedOutput(t *testing.T) {
	warnings := CheckShadowedNames([]string{"wait", "cd", "ls", "echo"})
	if len(warnings) != 4 {
		t.Fatalf("expected 4 warnings, got %d", len(warnings))
	}
	for i := 1; i < len(warnings); i++ {
		if warnings[i].Name < warnings[i-1].Name {
			t.Errorf("warnings not sorted: %q appears after %q", warnings[i].Name, warnings[i-1].Name)
		}
	}
}
