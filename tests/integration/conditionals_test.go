package integration

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConditional_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"platform-info", "skip-all", "pipe-conditional"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestConditional_PlatformInfo(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPiStdout(t, dir, "run", "platform-info")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	trimmed := strings.TrimSpace(out)
	lines := strings.Split(trimmed, "\n")

	if runtime.GOOS == "darwin" {
		if !strings.Contains(trimmed, "Running on macOS") {
			t.Errorf("on macOS, expected 'Running on macOS' in output, got:\n%s", trimmed)
		}
		if strings.Contains(trimmed, "Running on Linux") {
			t.Error("on macOS, should not contain 'Running on Linux'")
		}
	} else if runtime.GOOS == "linux" {
		if !strings.Contains(trimmed, "Running on Linux") {
			t.Errorf("on Linux, expected 'Running on Linux' in output, got:\n%s", trimmed)
		}
		if strings.Contains(trimmed, "Running on macOS") {
			t.Error("on Linux, should not contain 'Running on macOS'")
		}
	}

	lastLine := lines[len(lines)-1]
	if lastLine != "Done" {
		t.Errorf("last line = %q, want %q", lastLine, "Done")
	}
}

func TestConditional_SkipAll(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPiStdout(t, dir, "run", "skip-all")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "Always runs" {
		t.Errorf("output = %q, want %q", trimmed, "Always runs")
	}
}

func TestConditional_PipePassesThroughSkipped(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPiStdout(t, dir, "run", "pipe-conditional")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "hello world" {
		t.Errorf("output = %q, want %q (pipe should pass through skipped step)", trimmed, "hello world")
	}
}

func TestConditional_AutomationLevelIf_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"macos-only", "impossible", "call-conditional"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestConditional_AutomationLevelIf_Impossible(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "impossible")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "[skipped] impossible") {
		t.Errorf("expected skip message, got:\n%s", out)
	}
	if strings.Contains(out, "This should never run") {
		t.Errorf("impossible automation should not have executed")
	}
}

func TestConditional_AutomationLevelIf_MacOSOnly(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "macos-only")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if runtime.GOOS == "darwin" {
		if !strings.Contains(out, "macOS automation executed") {
			t.Errorf("on macOS, expected 'macOS automation executed', got:\n%s", out)
		}
	} else {
		if !strings.Contains(out, "[skipped] macos-only") {
			t.Errorf("on non-macOS, expected skip message, got:\n%s", out)
		}
	}
}

func TestConditional_AutomationLevelIf_RunStepCallsSkipped(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "call-conditional")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "before") {
		t.Errorf("expected 'before' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "after") {
		t.Errorf("expected 'after' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "[skipped] impossible") {
		t.Errorf("expected skip message for impossible, got:\n%s", out)
	}
	if strings.Contains(out, "This should never run") {
		t.Errorf("impossible automation should not have executed")
	}
}

func TestConditional_EnvCheck_WithVar(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPiWithEnv(t, dir, []string{"PI_TEST_VAR=1"}, "run", "env-check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "CI environment detected") {
		t.Errorf("expected env.PI_TEST_VAR step to run, got:\n%s", out)
	}
	if strings.Contains(out, "Not in CI") {
		t.Errorf("expected 'not env.PI_TEST_VAR' step to be skipped when var is set, got:\n%s", out)
	}
	if !strings.Contains(out, "env-check done") {
		t.Errorf("expected unconditional step, got:\n%s", out)
	}
}

func TestConditional_EnvCheck_WithoutVar(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	cmd := exec.Command(piBinary, "run", "env-check")
	cmd.Dir = dir
	var cleanEnv []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PI_TEST_VAR=") {
			cleanEnv = append(cleanEnv, e)
		}
	}
	cmd.Env = cleanEnv
	raw, err := cmd.CombinedOutput()
	out := string(raw)
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi: %v\n%s", err, out)
		}
	}
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: %s", exitCode, out)
	}
	if strings.Contains(out, "CI environment detected") {
		t.Errorf("expected env.PI_TEST_VAR step to be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "Not in CI") {
		t.Errorf("expected 'not env.PI_TEST_VAR' step to run, got:\n%s", out)
	}
}

func TestConditional_CommandCheck(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "command-check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "bash is available") {
		t.Errorf("expected command.bash step to run, got:\n%s", out)
	}
	if strings.Contains(out, "nonexistent-tool found") {
		t.Errorf("expected nonexistent command step to be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "command-check done") {
		t.Errorf("expected unconditional step, got:\n%s", out)
	}
}

func TestConditional_FileCheck(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "file-check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "pi.yaml exists") {
		t.Errorf("expected file.exists(pi.yaml) step to run, got:\n%s", out)
	}
	if strings.Contains(out, "missing file found") {
		t.Errorf("expected nonexistent file step to be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, ".pi dir exists") {
		t.Errorf("expected dir.exists(.pi) step to run, got:\n%s", out)
	}
	if strings.Contains(out, "missing dir found") {
		t.Errorf("expected nonexistent dir step to be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "file-check done") {
		t.Errorf("expected unconditional step, got:\n%s", out)
	}
}

func TestConditional_ComplexBool(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "complex-bool")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !strings.Contains(out, "macos or linux") {
			t.Errorf("expected 'macos or linux' on %s, got:\n%s", runtime.GOOS, out)
		}
	}
	if !strings.Contains(out, "has bash and pi.yaml") {
		t.Errorf("expected 'has bash and pi.yaml' step to run, got:\n%s", out)
	}
	if strings.Contains(out, "impossible combo") {
		t.Errorf("impossible combo should be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "not windows") {
		t.Errorf("expected 'not windows' step to run on non-Windows, got:\n%s", out)
	}
	if !strings.Contains(out, "complex-bool done") {
		t.Errorf("expected unconditional step, got:\n%s", out)
	}
}

func TestConditional_CombinedAutomationAndStepIf(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "run", "conditional-with-if")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	if runtime.GOOS == "darwin" {
		if !strings.Contains(out, "macos step") {
			t.Errorf("expected 'macos step' on darwin, got:\n%s", out)
		}
		if strings.Contains(out, "linux step") {
			t.Errorf("expected 'linux step' to be skipped on darwin, got:\n%s", out)
		}
	} else if runtime.GOOS == "linux" {
		if strings.Contains(out, "macos step") {
			t.Errorf("expected 'macos step' to be skipped on linux, got:\n%s", out)
		}
		if !strings.Contains(out, "linux step") {
			t.Errorf("expected 'linux step' on linux, got:\n%s", out)
		}
	}

	if !strings.Contains(out, "both platforms") {
		t.Errorf("expected unconditional step, got:\n%s", out)
	}
}

func TestConditional_Info_AutomationLevelIf(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "info", "impossible")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Condition:") {
		t.Errorf("expected Condition line in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "os.windows and os.linux") {
		t.Errorf("expected condition expression in info output, got:\n%s", out)
	}
}

func TestConditional_Info_StepLevelIf(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "info", "platform-info")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Step details:") {
		t.Errorf("expected Step details section, got:\n%s", out)
	}
	if !strings.Contains(out, "[if: os.macos]") {
		t.Errorf("expected step-level condition shown, got:\n%s", out)
	}
	if !strings.Contains(out, "[if: os.linux]") {
		t.Errorf("expected step-level condition shown, got:\n%s", out)
	}
}

func TestConditional_Info_NoCondition(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "info", "setup-always")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if strings.Contains(out, "Condition:") {
		t.Errorf("expected no Condition line for unconditional automation, got:\n%s", out)
	}
	if strings.Contains(out, "Step details:") {
		t.Errorf("expected no Step details for steps without conditions, got:\n%s", out)
	}
}

func TestConditional_List_AllAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "conditional")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{
		"call-conditional", "command-check", "complex-bool",
		"conditional-with-if", "env-check", "file-check",
		"impossible", "macos-only", "pipe-conditional",
		"platform-info", "setup-always", "setup-never",
		"setup-platform", "skip-all",
	} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}
