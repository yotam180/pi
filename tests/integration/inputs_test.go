package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInputs_PositionalArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "alice")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hello alice" {
		t.Errorf("output = %q, want %q", trimmed, "hello alice")
	}
}

func TestInputs_PositionalBothArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "greet", "bob", "hi")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hi bob" {
		t.Errorf("output = %q, want %q", trimmed, "hi bob")
	}
}

func TestInputs_WithFlags(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "--with", "name=charlie", "--with", "greeting=hey", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hey charlie" {
		t.Errorf("output = %q, want %q", trimmed, "hey charlie")
	}
}

func TestInputs_DefaultApplied(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "--with", "name=dave", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hello dave" {
		t.Errorf("output = %q, want %q", trimmed, "hello dave")
	}
}

func TestInputs_MissingRequired(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "greet")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing required input, got 0: %s", out)
	}
	if !strings.Contains(out, "required input") {
		t.Errorf("expected 'required input' in error, got: %s", out)
	}
}

func TestInputs_UnknownInput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "--with", "typo=val", "greet")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown input, got 0: %s", out)
	}
	if !strings.Contains(out, "unknown input") {
		t.Errorf("expected 'unknown input' in error, got: %s", out)
	}
}

func TestInputs_RunStepWithWith(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPiStdout(t, dir, "run", "caller")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hey world" {
		t.Errorf("output = %q, want %q", trimmed, "hey world")
	}
}

func TestInputs_InfoShowsEnvVarPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "$PI_IN_NAME") {
		t.Errorf("expected $PI_IN_NAME in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "$PI_IN_GREETING") {
		t.Errorf("expected $PI_IN_GREETING in info output, got:\n%s", out)
	}
}

func TestInputs_List_ShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "INPUTS") {
		t.Error("expected INPUTS column in list output")
	}
	if !strings.Contains(out, "name, greeting?") {
		t.Errorf("expected 'name, greeting?' in list output, got:\n%s", out)
	}
}

func setupPIArgsWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: test\n"), 0o644)
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0o755)

	os.WriteFile(filepath.Join(piDir, "build.yaml"), []byte("description: Build with forwarded args\nbash: echo \"build $PI_ARGS\"\n"), 0o644)
	os.WriteFile(filepath.Join(piDir, "test.yaml"), []byte("description: Run tests with forwarded args\nbash: echo \"test $PI_ARGS\"\n"), 0o644)
	return dir
}

func TestInputs_PIArgs_ForwardedToNoInputAutomation(t *testing.T) {
	dir := setupPIArgsWorkspace(t)
	out, code := runPiStdout(t, dir, "run", "build", "--release", "--verbose")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "build --release --verbose" {
		t.Errorf("output = %q, want %q", trimmed, "build --release --verbose")
	}
}

func TestInputs_PIArgs_SingleArg(t *testing.T) {
	dir := setupPIArgsWorkspace(t)
	out, code := runPiStdout(t, dir, "run", "test", "--ignored")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "test --ignored" {
		t.Errorf("output = %q, want %q", trimmed, "test --ignored")
	}
}

func TestInputs_PIArgs_EmptyWhenNoArgs(t *testing.T) {
	dir := setupPIArgsWorkspace(t)
	out, code := runPiStdout(t, dir, "run", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "build" {
		t.Errorf("output = %q, want %q", trimmed, "build")
	}
}
