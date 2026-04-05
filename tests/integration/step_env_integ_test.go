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
	if !strings.Contains(strings.ToLower(out), "env:") {
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

func TestAutoEnv_AppliesToAllSteps(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "run", "auto-env")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "step1=from_automation") {
		t.Errorf("expected step1 to see automation env, got: %s", out)
	}
	if !strings.Contains(out, "step2=from_automation") {
		t.Errorf("expected step2 to see automation env, got: %s", out)
	}
}

func TestAutoEnv_StepOverride(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "run", "auto-env-override")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "auto=from_automation") {
		t.Errorf("expected first step to use automation env, got: %s", out)
	}
	if !strings.Contains(out, "override=from_step") {
		t.Errorf("expected second step to use step override, got: %s", out)
	}
}

func TestAutoEnv_Shorthand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "run", "auto-env-shorthand")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "value=shorthand_works") {
		t.Errorf("expected shorthand automation-level env to work, got: %s", out)
	}
}

func TestAutoEnv_InfoShowsEnv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "info", "auto-env")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Env:") {
		t.Errorf("expected Env: line in info output, got: %s", out)
	}
	if !strings.Contains(out, "SHARED_VAR") {
		t.Errorf("expected SHARED_VAR in env list, got: %s", out)
	}
}

func TestAutoEnv_ListShowsNewAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "step-env")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "auto-env") {
		t.Errorf("expected auto-env in list, got: %s", out)
	}
	if !strings.Contains(out, "auto-env-override") {
		t.Errorf("expected auto-env-override in list, got: %s", out)
	}
	if !strings.Contains(out, "auto-env-shorthand") {
		t.Errorf("expected auto-env-shorthand in list, got: %s", out)
	}
}
