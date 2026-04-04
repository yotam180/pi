package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Polyglot example tests ---

func TestPolyglot_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	expected := []string{
		"data/format",
		"data/generate",
		"pipeline/etl",
		"pipeline/wordcount",
		"text/reverse",
		"text/transform",
	}
	for _, name := range expected {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestPolyglot_PythonInline(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "text/reverse")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "dlrow olleh") {
		t.Errorf("expected reversed default text, got:\n%s", out)
	}
}

func TestPolyglot_PythonInlineWithArg(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "text/reverse", "automation")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "noitamotua") {
		t.Errorf("expected reversed arg, got:\n%s", out)
	}
}

func TestPolyglot_PythonFile(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "text/transform")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "alpha") {
		t.Errorf("expected 'alpha' in box output, got:\n%s", out)
	}
	if !strings.Contains(out, "delta") {
		t.Errorf("expected 'delta' in box output, got:\n%s", out)
	}
	if !strings.Contains(out, "+") {
		t.Errorf("expected box border, got:\n%s", out)
	}
}

func TestPolyglot_TypeScriptInline(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "data/generate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice' in JSON output, got:\n%s", out)
	}
	if !strings.Contains(out, "95") {
		t.Errorf("expected score '95' in JSON output, got:\n%s", out)
	}
}

func TestPolyglot_TypeScriptFile(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "data/format")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Leaderboard") {
		t.Errorf("expected 'Leaderboard' header, got:\n%s", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice' in leaderboard, got:\n%s", out)
	}
	if !strings.Contains(out, "3 entries") {
		t.Errorf("expected '3 entries' footer, got:\n%s", out)
	}
}

func TestPolyglot_ETLPipeline(t *testing.T) {
	requirePython(t)
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "pipeline/etl")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "ETL Complete") {
		t.Errorf("expected 'ETL Complete' header, got:\n%s", out)
	}
	if !strings.Contains(out, "alice: 95") {
		t.Errorf("expected 'alice: 95' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Total: 3 records") {
		t.Errorf("expected 'Total: 3 records' in output, got:\n%s", out)
	}
}

func TestPolyglot_ETLPipeline_StepOrder(t *testing.T) {
	requirePython(t)
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "pipeline/etl")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "ETL Complete") || !strings.Contains(out, "Total:") {
		t.Fatalf("missing expected output sections:\n%s", out)
	}
	etlIdx := strings.Index(out, "ETL Complete")
	totalIdx := strings.Index(out, "Total:")
	if etlIdx > totalIdx {
		t.Error("expected 'ETL Complete' before 'Total:' in output")
	}
}

func TestPolyglot_WordCount(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "polyglot")
	out, code := runPi(t, dir, "run", "pipeline/wordcount")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "words: 9" {
		t.Errorf("expected 'words: 9', got %q", trimmed)
	}
}

func TestPolyglot_FromSubdirectory(t *testing.T) {
	dir := filepath.Join(examplesDir(), "polyglot")
	sub := filepath.Join(dir, "subdir")
	os.MkdirAll(sub, 0o755)
	defer os.RemoveAll(sub)

	out, code := runPi(t, sub, "list")
	if code != 0 {
		t.Fatalf("expected exit 0 from subdir, got %d: %s", code, out)
	}
	if !strings.Contains(out, "text/reverse") {
		t.Errorf("expected automations listed from subdir, got:\n%s", out)
	}
}
