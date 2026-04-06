package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList_BasicOutput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	for _, header := range []string{"NAME", "SOURCE", "DESCRIPTION", "INPUTS"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected table header %q, got:\n%s", header, out)
		}
	}

	for _, name := range []string{"greet", "build/compile", "deploy"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected automation %q in output, got:\n%s", name, out)
		}
	}
}

func TestList_SourceColumnShowsWorkspace(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "[workspace]") {
		t.Errorf("expected [workspace] source for local automations, got:\n%s", out)
	}
}

func TestList_InputsColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "name, greeting?") {
		t.Errorf("expected inputs summary 'name, greeting?' for greet automation, got:\n%s", out)
	}
}

func TestList_BuiltinsHiddenByDefault(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if strings.Contains(out, "pi:") {
		t.Errorf("builtins should be hidden by default, got:\n%s", out)
	}
}

func TestList_BuiltinsFlag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list", "--builtins")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "[built-in]") {
		t.Errorf("expected [built-in] source with --builtins flag, got:\n%s", out)
	}
	if !strings.Contains(out, "install-python") {
		t.Errorf("expected install-python builtin, got:\n%s", out)
	}
	if !strings.Contains(out, "install-node") {
		t.Errorf("expected install-node builtin, got:\n%s", out)
	}
}

func TestList_BuiltinsShortFlag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list", "-b")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "[built-in]") {
		t.Errorf("expected [built-in] source with -b flag, got:\n%s", out)
	}
}

func TestList_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: empty\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".pi"), 0o755)

	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "No automations found") {
		t.Errorf("expected 'No automations found' for empty project, got:\n%s", out)
	}
}

func TestList_PackageSourceColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "packages")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "mytools") {
		t.Errorf("expected package alias 'mytools' in SOURCE column, got:\n%s", out)
	}
}

func TestList_AllFlagShowsPackageSections(t *testing.T) {
	dir := filepath.Join(examplesDir(), "packages")
	out, code := runPi(t, dir, "list", "--all")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "──") {
		t.Errorf("expected package section separator with --all, got:\n%s", out)
	}
	if !strings.Contains(out, "mytools") {
		t.Errorf("expected package alias in section header, got:\n%s", out)
	}
}

func TestList_FromSubdirectory(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	sub := filepath.Join(dir, "subdir-list-test")
	os.MkdirAll(sub, 0o755)
	defer os.RemoveAll(sub)

	out, code := runPi(t, sub, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "greet") {
		t.Errorf("expected automations found from subdirectory, got:\n%s", out)
	}
}

func TestList_NoPiYaml(t *testing.T) {
	dir := t.TempDir()
	_, code := runPi(t, dir, "list")
	if code == 0 {
		t.Fatal("expected non-zero exit when no pi.yaml found")
	}
}

func TestList_DescriptionColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	if !strings.Contains(out, "Greet someone") {
		t.Errorf("expected automation description in output, got:\n%s", out)
	}
}

func TestList_NoInputsShowsDash(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", code, out)
	}

	lines := strings.Split(out, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "greet") {
			found = true
			if !strings.Contains(line, "-") {
				t.Errorf("expected '-' for no inputs on greet line, got: %q", line)
			}
			break
		}
	}
	if !found {
		t.Errorf("greet automation not found in output:\n%s", out)
	}
}
