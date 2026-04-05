package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_ValidProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-valid")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success marker, got: %s", stdout)
	}
	if !strings.Contains(stdout, "shortcut") {
		t.Errorf("expected shortcut count, got: %s", stdout)
	}
	if !strings.Contains(stdout, "setup") {
		t.Errorf("expected setup count, got: %s", stdout)
	}
	if !strings.Contains(stdout, "automation") {
		t.Errorf("expected automation count, got: %s", stdout)
	}
}

func TestValidate_InvalidProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-invalid")
	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "nonexistent-automation") {
		t.Errorf("expected broken shortcut error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "also-nonexistent") {
		t.Errorf("expected broken setup error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "ghost-automation") {
		t.Errorf("expected broken run step error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "3 error") {
		t.Errorf("expected 3 errors counted, got: %s", stderr)
	}
}

func TestValidate_InvalidProject_AllErrorsReported(t *testing.T) {
	dir := filepath.Join(examplesDir(), "validate-invalid")
	_, stderr, code := runPiSplit(t, dir, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	errorCount := strings.Count(stderr, "✗")
	if errorCount != 3 {
		t.Errorf("expected 3 error lines (✗), got %d in:\n%s", errorCount, stderr)
	}
}

func TestValidate_BasicProject(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for basic project, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success for basic project, got: %s", stdout)
	}
}

func TestValidate_BuiltinRefsValid(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	stdout, _, code := runPiSplit(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected exit 0 for builtins project, got %d", code)
	}
	if !strings.Contains(stdout, "✓") {
		t.Errorf("expected success for builtins project, got: %s", stdout)
	}
}
