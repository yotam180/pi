package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBasic_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"greet", "build/compile", "deploy"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "DESCRIPTION") {
		t.Errorf("expected table headers, got:\n%s", out)
	}
}

func TestBasic_Greet(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Hello, World!") {
		t.Errorf("expected greeting, got:\n%s", out)
	}
}

func TestBasic_GreetWithArg(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "greet", "Alice")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Hello, Alice!") {
		t.Errorf("expected personalized greeting, got:\n%s", out)
	}
}

func TestBasic_BuildCompile(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "build/compile")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Compiling") || !strings.Contains(out, "Build complete") {
		t.Errorf("expected build output, got:\n%s", out)
	}
}

func TestBasic_Deploy(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "deploy")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Compiling") {
		t.Errorf("expected build step output (run: chaining), got:\n%s", out)
	}
	if !strings.Contains(out, "Deploy complete") {
		t.Errorf("expected deploy step output, got:\n%s", out)
	}
}

func TestBasic_NotFound(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "nonexistent")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown automation")
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' error, got:\n%s", out)
	}
}

func TestBasic_NotFound_DidYouMean(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "gret")
	if code == 0 {
		t.Fatal("expected non-zero exit for misspelled automation")
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' error, got:\n%s", out)
	}
	if !strings.Contains(out, "Did you mean?") {
		t.Errorf("expected 'Did you mean?' suggestion, got:\n%s", out)
	}
	if !strings.Contains(out, "greet") {
		t.Errorf("expected 'greet' in suggestions, got:\n%s", out)
	}
}

func TestBasic_NotFound_DidYouMean_NoSuggestions(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "run", "zzzzzzzzzzzzzzzzzzzzz")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown automation")
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' error, got:\n%s", out)
	}
	if strings.Contains(out, "Did you mean?") {
		t.Errorf("should NOT show 'Did you mean?' for distant names, got:\n%s", out)
	}
}

func TestBasic_FromSubdirectory(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0o755)
	defer os.RemoveAll(sub)

	out, code := runPi(t, sub, "list")
	if code != 0 {
		t.Fatalf("expected exit 0 from subdir, got %d: %s", code, out)
	}
	if !strings.Contains(out, "greet") {
		t.Errorf("expected automations listed from subdir, got:\n%s", out)
	}
}
