package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputs_ChainWith_PassesLastOutputToRunStep(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "chain-with")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "received=captured-value") {
		t.Errorf("expected outputs.last to resolve in with: value, got stdout: %q", stdout)
	}
}

func TestOutputs_ChainEnv_OutputsLastInStepEnv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "chain-env")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "from_env=env-value-42") {
		t.Errorf("expected outputs.last in env: to resolve, got stdout: %q", stdout)
	}
}

func TestOutputs_Indexed_AccessByStepNumber(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "indexed")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "step0=first-output") {
		t.Errorf("expected outputs.0 to resolve, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "step1=second-output") {
		t.Errorf("expected outputs.1 to resolve, got stdout: %q", stdout)
	}
}

func TestOutputs_InputsInWith_ForwardsToSubAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "inputs-in-with", "hello-world")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "received=hello-world") {
		t.Errorf("expected inputs.msg to resolve in with: value, got stdout: %q", stdout)
	}
}

func TestOutputs_InputsInEnv_ForwardsToStepEnv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "inputs-in-env", "v2.0")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "tag=v2.0") {
		t.Errorf("expected inputs.tag to resolve in env: value, got stdout: %q", stdout)
	}
}

func TestOutputs_AutoEnvInputs_AutomationLevelEnvReferencesInput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "auto-env-inputs", "3.14")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "building=3.14") {
		t.Errorf("expected automation-level env to resolve inputs.version, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "still=3.14") {
		t.Errorf("expected second step to also see resolved env, got stdout: %q", stdout)
	}
}

func TestOutputs_Mixed_LiteralsAndInterpolation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "mixed", "test-prefix")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "received=dynamic-suffix") {
		t.Errorf("expected outputs.last in with: to resolve, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "prefix=test-prefix") {
		t.Errorf("expected input env var to resolve, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "literal=unchanged") {
		t.Errorf("expected literal value to pass through, got stdout: %q", stdout)
	}
}

func TestOutputs_PipeAndCapture_PipedStepRecordsOutput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "pipe-and-capture")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "piped-data") {
		t.Errorf("expected piped data to pass through, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "last=piped-data") {
		t.Errorf("expected outputs.1 to capture piped step output, got stdout: %q", stdout)
	}
}

func TestOutputs_ThreeStepChain_EachStepUsesLastOutput(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	stdout, code := runPiStdout(t, dir, "run", "three-step-chain")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "alpha-beta") {
		t.Errorf("expected step 2 to build on step 1 output, got stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "alpha-beta-gamma") {
		t.Errorf("expected step 3 to build on step 2 output, got stdout: %q", stdout)
	}
}

func TestOutputs_List_ShowsAllAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"chain-with", "chain-env", "indexed", "echo-input", "inputs-in-with", "inputs-in-env", "auto-env-inputs", "mixed", "pipe-and-capture", "three-step-chain"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got: %s", name, out)
		}
	}
}

func TestOutputs_Validate_AllAutomationsValid(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	out, code := runPi(t, dir, "validate")
	if code != 0 {
		t.Fatalf("expected validation to pass, got exit %d: %s", code, out)
	}
}

func TestOutputs_Info_ChainWith(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	out, code := runPi(t, dir, "info", "chain-with")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "chain-with") {
		t.Errorf("expected automation name in info, got: %s", out)
	}
	if !strings.Contains(out, "captures output") {
		t.Errorf("expected description in info, got: %s", out)
	}
}

func TestOutputs_DryRun_ChainWith(t *testing.T) {
	dir := filepath.Join(examplesDir(), "outputs")
	out, code := runPi(t, dir, "run", "--dry-run", "chain-with")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "bash") {
		t.Errorf("expected bash step in dry run output, got: %s", out)
	}
	if !strings.Contains(out, "run") {
		t.Errorf("expected run step in dry run output, got: %s", out)
	}
}
