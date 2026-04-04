package cli

import (
	"bytes"
	"strings"
	"testing"
)

func executeCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd := NewRootCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestRootHelp(t *testing.T) {
	out, err := executeCmd("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PI") {
		t.Errorf("expected help to mention PI, got: %s", out)
	}
	for _, sub := range []string{"run", "list", "setup", "shell"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected help to list %q subcommand, got: %s", sub, out)
		}
	}
}

func TestRunHelp(t *testing.T) {
	out, err := executeCmd("run", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "automation") {
		t.Errorf("expected run help to mention automation, got: %s", out)
	}
}

func TestRunRequiresArg(t *testing.T) {
	_, err := executeCmd("run")
	if err == nil {
		t.Fatal("expected error when no automation name given")
	}
}

func TestListHelp(t *testing.T) {
	out, err := executeCmd("list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "automation") {
		t.Errorf("expected list help to mention automations, got: %s", out)
	}
}

func TestSetupHelp(t *testing.T) {
	out, err := executeCmd("setup", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "setup") {
		t.Errorf("expected setup help, got: %s", out)
	}
	if !strings.Contains(out, "--no-shell") {
		t.Errorf("expected --no-shell flag in help, got: %s", out)
	}
}

func TestShellHelp(t *testing.T) {
	out, err := executeCmd("shell", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "shortcut") {
		t.Errorf("expected shell help to mention shortcuts, got: %s", out)
	}
}

func TestVersion(t *testing.T) {
	out, err := executeCmd("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("expected version to contain 'dev', got: %s", out)
	}
}
