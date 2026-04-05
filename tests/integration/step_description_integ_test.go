package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStepDescription_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-description")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"build", "deploy", "no-desc"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got: %s", name, out)
		}
	}
}

func TestStepDescription_Run(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-description")
	stdout, _, code := runPiSplit(t, dir, "run", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "compiling...") {
		t.Errorf("expected 'compiling...' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "done") {
		t.Errorf("expected 'done' in output, got: %s", stdout)
	}
}

func TestStepDescription_InfoShowsDescriptions(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-description")
	out, code := runPi(t, dir, "info", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected 'Step details:' section, got: %s", out)
	}
	if !strings.Contains(out, "Compile source code") {
		t.Errorf("expected step description 'Compile source code', got: %s", out)
	}
	if !strings.Contains(out, "Link binaries") {
		t.Errorf("expected step description 'Link binaries', got: %s", out)
	}
}

func TestStepDescription_InfoWithAnnotations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-description")
	out, code := runPi(t, dir, "info", "deploy")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Build Docker image") {
		t.Errorf("expected description 'Build Docker image', got: %s", out)
	}
	if !strings.Contains(out, "silent") {
		t.Errorf("expected 'silent' annotation, got: %s", out)
	}
	if !strings.Contains(out, "Push to registry") {
		t.Errorf("expected description 'Push to registry', got: %s", out)
	}
	if !strings.Contains(out, "Verify deployment") {
		t.Errorf("expected description 'Verify deployment', got: %s", out)
	}
}

func TestStepDescription_InfoNoDescNoDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-description")
	out, code := runPi(t, dir, "info", "no-desc")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if strings.Contains(out, "Step details:") {
		t.Errorf("expected no 'Step details:' for automation without step descriptions or annotations, got: %s", out)
	}
}
