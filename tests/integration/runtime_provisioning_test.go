package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeProvisioning_ListAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "runtime-provisioning")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"needs-python", "no-provision"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestRuntimeProvisioning_NoRequirements(t *testing.T) {
	dir := filepath.Join(examplesDir(), "runtime-provisioning")
	out, code := runPi(t, dir, "run", "no-provision")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "no runtimes needed") {
		t.Errorf("expected 'no runtimes needed' in output, got:\n%s", out)
	}
}

func TestRuntimeProvisioning_PythonAlreadyInstalled(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "runtime-provisioning")
	out, code := runPi(t, dir, "run", "needs-python")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "python is available") {
		t.Errorf("expected 'python is available' in output, got:\n%s", out)
	}
	if strings.Contains(out, "[provisioned]") {
		t.Errorf("should not provision when runtime is already available, got:\n%s", out)
	}
}

func TestRuntimeProvisioning_NeverModeErrors(t *testing.T) {
	dir := filepath.Join(examplesDir(), "runtime-provisioning-never")
	out, code := runPi(t, dir, "run", "needs-impossible")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Missing requirements:") {
		t.Errorf("expected 'Missing requirements:' in output, got:\n%s", out)
	}
}

func TestRuntimeProvisioning_ConfigParsedCorrectly(t *testing.T) {
	dir := filepath.Join(examplesDir(), "runtime-provisioning")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "needs-python") {
		t.Errorf("expected automation in list, got:\n%s", out)
	}
}

func TestRuntimeProvisioning_RuntimesConfig_Auto(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)

	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`
project: test-auto
runtimes:
  provision: auto
  manager: direct
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`
name: hello
description: Test automation
steps:
  - bash: echo "hello"
`), 0644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected 'hello' in output, got:\n%s", out)
	}
}

func TestRuntimeProvisioning_RuntimesConfig_Ask(t *testing.T) {
	dir := t.TempDir()
	piDir := filepath.Join(dir, ".pi")
	os.MkdirAll(piDir, 0755)

	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte(`
project: test-ask
runtimes:
  provision: ask
  manager: mise
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(piDir, "hello.yaml"), []byte(`
name: hello
description: Test automation
steps:
  - bash: echo "hello"
`), 0644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
}
