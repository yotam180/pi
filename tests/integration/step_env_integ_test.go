package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStepEnv_RunBuild(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "run", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "building for linux/amd64") {
		t.Errorf("expected env vars in output, got: %s", out)
	}
}

func TestStepEnv_MultiEnvIsolation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "run", "multi-env")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "step1=alpha") {
		t.Errorf("expected step1=alpha, got: %s", out)
	}
	if !strings.Contains(out, "step2=unset") {
		t.Errorf("expected step2=unset (env not leaked), got: %s", out)
	}
}

func TestStepEnv_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "build") {
		t.Errorf("expected 'build' in list, got: %s", out)
	}
}

func TestStepEnv_Info(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "info", "build")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "env:") {
		t.Errorf("expected env annotation in info output, got: %s", out)
	}
}

func TestStepEnv_InfoWithCondition(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "info", "env-with-if")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "if:") {
		t.Errorf("expected 'if:' in info output, got: %s", out)
	}
	if !strings.Contains(out, "env:") {
		t.Errorf("expected 'env:' in info output, got: %s", out)
	}
}
