package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var piBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "pi-integration-*")
	if err != nil {
		panic("creating temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmp)

	piBinary = filepath.Join(tmp, "pi")
	if runtime.GOOS == "windows" {
		piBinary += ".exe"
	}

	repoRoot := findRepoRoot()
	cmd := exec.Command("go", "build", "-o", piBinary, "./cmd/pi/")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("building pi binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func findRepoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repo root (go.mod)")
		}
		dir = parent
	}
}

func examplesDir() string {
	return filepath.Join(findRepoRoot(), "examples")
}

func runPi(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v\n%s", args, err, string(out))
		}
	}
	return string(out), exitCode
}

func runPiWithEnv(t *testing.T, dir string, env []string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v\n%s", args, err, string(out))
		}
	}
	return string(out), exitCode
}

// --- Basic example tests ---

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

// --- Docker project tests ---

func TestDocker_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"docker/up", "docker/down", "docker/logs", "docker/build", "docker/build-and-up"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestDocker_Up(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "run", "docker/up")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "All containers started") {
		t.Errorf("expected up output, got:\n%s", out)
	}
}

func TestDocker_Down(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "run", "docker/down")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "All containers stopped") {
		t.Errorf("expected down output, got:\n%s", out)
	}
}

func TestDocker_Logs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "run", "docker/logs")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "[api]") {
		t.Errorf("expected log output, got:\n%s", out)
	}
}

func TestDocker_LogsWithArg(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "run", "docker/logs", "api")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Showing logs for: api") {
		t.Errorf("expected filtered logs, got:\n%s", out)
	}
}

func TestDocker_BuildAndUp(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-project")
	out, code := runPi(t, dir, "run", "docker/build-and-up")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "All images built") {
		t.Errorf("expected build output from run: chain, got:\n%s", out)
	}
	if !strings.Contains(out, "All containers started") {
		t.Errorf("expected up output from run: chain, got:\n%s", out)
	}

	buildIdx := strings.Index(out, "All images built")
	startIdx := strings.Index(out, "All containers started")
	if buildIdx > startIdx {
		t.Error("expected build to happen before up")
	}
}

// --- Pipe example tests ---

func TestPipe_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"upper", "count-lines"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestPipe_Upper(t *testing.T) {
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPi(t, dir, "run", "upper")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "HELLO FROM PIPE") {
		t.Errorf("expected uppercased output, got:\n%s", out)
	}
}

func TestPipe_CountLines(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "pipe")
	out, code := runPi(t, dir, "run", "count-lines")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "3" {
		t.Errorf("expected line count '3', got %q", trimmed)
	}
}

// --- Version tests ---

func TestVersion_Flag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "--version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "pi ") {
		t.Errorf("expected output starting with 'pi ', got %q", trimmed)
	}
	if len(strings.TrimPrefix(trimmed, "pi ")) == 0 {
		t.Error("expected non-empty version string after 'pi '")
	}
}

func TestVersion_Subcommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "pi ") {
		t.Errorf("expected output starting with 'pi ', got %q", trimmed)
	}
}

func TestVersion_FlagAndSubcommandMatch(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	flagOut, flagCode := runPi(t, dir, "--version")
	subOut, subCode := runPi(t, dir, "version")
	if flagCode != 0 || subCode != 0 {
		t.Fatalf("expected exit 0 for both, got flag=%d sub=%d", flagCode, subCode)
	}
	if strings.TrimSpace(flagOut) != strings.TrimSpace(subOut) {
		t.Errorf("--version and version subcommand differ: %q vs %q", flagOut, subOut)
	}
}

// --- Inputs tests ---

func TestInputs_PositionalArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "greet", "alice")
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
	out, code := runPi(t, dir, "run", "greet", "bob", "hi")
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
	out, code := runPi(t, dir, "run", "greet", "--with", "name=charlie", "--with", "greeting=hey")
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
	out, code := runPi(t, dir, "run", "greet", "--with", "name=dave")
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
	out, code := runPi(t, dir, "run", "greet", "--with", "typo=val")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown input, got 0: %s", out)
	}
	if !strings.Contains(out, "unknown input") {
		t.Errorf("expected 'unknown input' in error, got: %s", out)
	}
}

func TestInputs_RunStepWithWith(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "run", "caller")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed != "hey world" {
		t.Errorf("output = %q, want %q", trimmed, "hey world")
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

// --- Info command tests ---

func TestInfo_BasicAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Name:") {
		t.Errorf("expected Name header, got:\n%s", out)
	}
	if !strings.Contains(out, "greet") {
		t.Errorf("expected automation name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "No inputs.") {
		t.Errorf("expected 'No inputs.' for automation without inputs, got:\n%s", out)
	}
}

func TestInfo_WithInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "inputs")
	out, code := runPi(t, dir, "info", "greet")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Inputs:") {
		t.Errorf("expected Inputs header, got:\n%s", out)
	}
	if !strings.Contains(out, "name (string, required)") {
		t.Errorf("expected required input with type, got:\n%s", out)
	}
	if !strings.Contains(out, "Who to greet") {
		t.Errorf("expected input description, got:\n%s", out)
	}
	if !strings.Contains(out, `default: "hello"`) {
		t.Errorf("expected default value shown, got:\n%s", out)
	}
	if !strings.Contains(out, "optional") {
		t.Errorf("expected 'optional' for optional input, got:\n%s", out)
	}
}

func TestInfo_NotFound(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "info", "nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown automation, got 0: %s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in error, got: %s", out)
	}
}

func TestInfo_NoArgs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	_, code := runPi(t, dir, "info")
	if code == 0 {
		t.Fatal("expected non-zero exit when no argument provided")
	}
}

// --- Conditional step tests ---

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
	out, code := runPi(t, dir, "run", "platform-info")
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
	out, code := runPi(t, dir, "run", "skip-all")
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
	out, code := runPi(t, dir, "run", "pipe-conditional")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	trimmed := strings.TrimSpace(out)
	// The middle step (tr a-z A-Z) should be skipped (condition is os.windows),
	// so the output should be lowercase "hello world" passed through unchanged.
	if trimmed != "hello world" {
		t.Errorf("output = %q, want %q (pipe should pass through skipped step)", trimmed, "hello world")
	}
}

// --- Automation-level if: tests ---

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

// --- Conditional: env predicate tests ---

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
	// Run without PI_TEST_VAR — use a clean env to ensure it's unset
	cmd := exec.Command(piBinary, "run", "env-check")
	cmd.Dir = dir
	// Build a minimal env without PI_TEST_VAR
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
		if exitErr, ok := err.(*exec.ExitError); ok {
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

// --- Conditional: command predicate tests ---

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

// --- Conditional: file/dir predicate tests ---

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

// --- Conditional: complex boolean tests ---

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

// --- Conditional: automation with both automation-level and step-level if ---

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

// --- Conditional: pi info shows if: conditions ---

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

// --- Built-in automation tests ---

func TestBuiltins_List_ShowsBuiltinMarker(t *testing.T) {
	dir := filepath.Join(examplesDir(), "basic")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "[built-in]") {
		t.Errorf("expected [built-in] marker in list output, got:\n%s", out)
	}
}

func TestBuiltins_RunWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected built-in output, got:\n%s", out)
	}
}

func TestBuiltins_LocalShadowsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from local override") {
		t.Errorf("expected local override, got:\n%s", out)
	}
}

func TestBuiltins_PiPrefixAlwaysGetsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected built-in despite local shadow, got:\n%s", out)
	}
}

func TestBuiltins_RunStepCallsBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "call-builtin")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected run step to resolve pi:hello to built-in, got:\n%s", out)
	}
}

func TestBuiltins_InfoWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "info", "pi:hello")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Name:") {
		t.Errorf("expected info output, got:\n%s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected automation name in info, got:\n%s", out)
	}
}

func TestBuiltins_ListShadowed(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	// "hello" exists locally, should NOT be marked [built-in]
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "hello") && strings.Contains(line, "[built-in]") {
			t.Errorf("expected local 'hello' to NOT have [built-in] marker, got:\n%s", line)
		}
	}
}

func TestBuiltins_SetupWithPiPrefix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPiWithEnv(t, dir, []string{"CI=true"}, "setup")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "hello from built-in") {
		t.Errorf("expected setup to run pi:hello built-in, got:\n%s", out)
	}
	if !strings.Contains(out, "hello from local") {
		t.Errorf("expected setup to run local-hello, got:\n%s", out)
	}
}

func TestBuiltins_PiPrefixNotFound(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit for pi:nonexistent")
	}
	if !strings.Contains(out, "built-in automation") {
		t.Errorf("expected built-in not found error, got:\n%s", out)
	}
}

// --- Built-in Docker automations ---

func TestBuiltins_DockerAutomationsInList(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DockerAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"docker/up", "docker/down", "docker/logs"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DockerInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	for _, name := range []string{"pi:docker/up", "pi:docker/down", "pi:docker/logs"} {
		t.Run(name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, "Description:") {
				t.Errorf("expected Description: in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_DockerRunStepResolvesBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "call-docker-up") {
		t.Errorf("expected call-docker-up in list output, got:\n%s", out)
	}
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected docker/up built-in in list output, got:\n%s", out)
	}
}

func TestBuiltins_DockerRunStepInfoResolvesBuiltin(t *testing.T) {
	dir := filepath.Join(examplesDir(), "docker-builtins")
	out, code := runPi(t, dir, "info", "pi:docker/up")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "docker/up") {
		t.Errorf("expected docker/up in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "Docker Compose") {
		t.Errorf("expected description mentioning Docker Compose, got:\n%s", out)
	}
}

// --- Installer built-in automations ---

func TestBuiltins_InstallerAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"install-homebrew", "install-python", "install-node", "install-go", "install-rust", "install-uv", "install-tsx"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_InstallerInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	installers := []struct {
		name   string
		substr string
	}{
		{"pi:install-homebrew", "Homebrew"},
		{"pi:install-python", "Python"},
		{"pi:install-node", "Node.js"},
		{"pi:install-rust", "Rust"},
		{"pi:install-uv", "uv"},
		{"pi:install-tsx", "tsx"},
	}
	for _, tc := range installers {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, tc.substr) {
				t.Errorf("expected %q in info output, got:\n%s", tc.substr, out)
			}
		})
	}
}

func TestBuiltins_InstallerInfoShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	for _, name := range []string{"pi:install-python", "pi:install-node", "pi:install-rust"} {
		t.Run(name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "version") {
				t.Errorf("expected 'version' input in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_InstallerHomebrewShowsCondition(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "info", "pi:install-homebrew")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "os.macos") {
		t.Errorf("expected 'os.macos' condition in info output, got:\n%s", out)
	}
}

func TestBuiltins_InstallTsxIdempotent(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "run", "pi:install-tsx")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") && !strings.Contains(out, "installed") {
		t.Errorf("expected 'already installed' or 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' status icon in output, got:\n%s", out)
	}
}

func TestBuiltins_InstallerListShowsInputsColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "install-python") && strings.Contains(line, "[built-in]") {
			if !strings.Contains(line, "version") {
				t.Errorf("expected install-python list line to show 'version' input, got:\n%s", line)
			}
			break
		}
	}
}

// --- Dev tool built-in automations ---

func TestBuiltins_DevToolAutomationsInList(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"cursor/install-extensions", "git/install-hooks"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DevToolAutomationsMarkedBuiltIn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"cursor/install-extensions", "git/install-hooks"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, name) && strings.Contains(line, "[built-in]") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to have [built-in] marker in list output, got:\n%s", name, out)
		}
	}
}

func TestBuiltins_DevToolInfoShowsDetails(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	tests := []struct {
		name        string
		description string
	}{
		{"pi:cursor/install-extensions", "Install missing Cursor extensions"},
		{"pi:git/install-hooks", "Install git hooks from a source directory"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, "Name:") {
				t.Errorf("expected Name: in info output, got:\n%s", out)
			}
			if !strings.Contains(out, "Description:") {
				t.Errorf("expected Description: in info output, got:\n%s", out)
			}
		})
	}
}

func TestBuiltins_DevToolInfoShowsInputs(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	tests := []struct {
		name  string
		input string
	}{
		{"pi:cursor/install-extensions", "extensions"},
		{"pi:git/install-hooks", "source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runPi(t, dir, "info", tc.name)
			if code != 0 {
				t.Fatalf("expected exit 0, got %d: %s", code, out)
			}
			if !strings.Contains(out, tc.input) {
				t.Errorf("expected %q input in info output, got:\n%s", tc.input, out)
			}
		})
	}
}

func TestBuiltins_DevToolListShowsInputsColumn(t *testing.T) {
	dir := filepath.Join(examplesDir(), "builtins")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	tests := []struct {
		name  string
		input string
	}{
		{"cursor/install-extensions", "extensions"},
		{"git/install-hooks", "source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, tc.name) && strings.Contains(line, "[built-in]") {
					if !strings.Contains(line, tc.input) {
						t.Errorf("expected %s list line to show %q input, got:\n%s", tc.name, tc.input, line)
					}
					return
				}
			}
			t.Errorf("did not find %s in list output:\n%s", tc.name, out)
		})
	}
}

// --- Conditional: list shows all conditional automations ---

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

// --- Installer schema integration tests ---

func TestInstallerSchema_ListShowsInstallerAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"check-ready", "install-marker", "install-conditional", "install-no-version", "steps-automation"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestInstallerSchema_AlreadyInstalled(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' icon in output, got:\n%s", out)
	}
}

func TestInstallerSchema_FreshInstall(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	marker := filepath.Join(t.TempDir(), "test-marker")
	out, code := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "installing...") {
		t.Errorf("expected 'installing...' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "installed") {
		t.Errorf("expected 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected version '1.0.0' in output, got:\n%s", out)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Errorf("expected marker file to exist at %s", marker)
	}
}

func TestInstallerSchema_FreshInstallThenAlreadyInstalled(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	marker := filepath.Join(t.TempDir(), "test-marker")

	// First run: fresh install
	out1, code1 := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code1 != 0 {
		t.Fatalf("first run: expected exit 0, got %d: %s", code1, out1)
	}
	if !strings.Contains(out1, "installing...") {
		t.Errorf("first run: expected 'installing...', got:\n%s", out1)
	}

	// Second run: already installed
	out2, code2 := runPi(t, dir, "run", "install-marker", "--with", "path="+marker)
	if code2 != 0 {
		t.Fatalf("second run: expected exit 0, got %d: %s", code2, out2)
	}
	if !strings.Contains(out2, "already installed") {
		t.Errorf("second run: expected 'already installed', got:\n%s", out2)
	}
}

func TestInstallerSchema_NoVersion(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "install-no-version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed', got:\n%s", out)
	}
	// No version parenthetical should appear
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "install-no-version") && strings.Contains(line, "(") {
			t.Errorf("expected no version parenthetical, got:\n%s", line)
		}
	}
}

func TestInstallerSchema_InfoShowsInstallerType(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Type:") {
		t.Errorf("expected 'Type:' in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "installer") {
		t.Errorf("expected 'installer' in info output, got:\n%s", out)
	}
	if !strings.Contains(out, "Install lifecycle") {
		t.Errorf("expected 'Install lifecycle' in info output, got:\n%s", out)
	}
}

func TestInstallerSchema_InfoShowsStepsForRegularAutomation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "info", "steps-automation")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Steps:") {
		t.Errorf("expected 'Steps:' in info output, got:\n%s", out)
	}
	if strings.Contains(out, "Type:         installer") {
		t.Errorf("unexpected 'Type: installer' in info output for steps-based automation, got:\n%s", out)
	}
}

func TestInstallerSchema_ConditionalRunSteps(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "install-conditional")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "installed") {
		t.Errorf("expected 'installed' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2.0.0") {
		t.Errorf("expected version '2.0.0' in output, got:\n%s", out)
	}
}

func TestInstallerSchema_SilentFlag(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "--silent", "check-ready")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if strings.Contains(out, "already installed") {
		t.Errorf("expected no status output with --silent, got:\n%s", out)
	}
	if strings.Contains(out, "✓") {
		t.Errorf("expected no status icon with --silent, got:\n%s", out)
	}
}

func TestInstallerSchema_RegularAutomationUnaffectedBySilent(t *testing.T) {
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "--silent", "steps-automation")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "I am a regular automation") {
		t.Errorf("expected regular automation output even with --silent, got:\n%s", out)
	}
}

func TestInstallerSchema_BuiltinInstallerAlreadyInstalled(t *testing.T) {
	requireTsx(t)
	dir := filepath.Join(examplesDir(), "installer-schema")
	out, code := runPi(t, dir, "run", "pi:install-tsx")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected '✓' icon for already-installed tsx, got:\n%s", out)
	}
	if !strings.Contains(out, "already installed") {
		t.Errorf("expected 'already installed' for tsx, got:\n%s", out)
	}
}

// --- Requires validation tests ---

func TestRequiresValidation_List(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	for _, name := range []string{"needs-bash", "needs-impossible", "needs-impossible-version", "needs-python", "no-requires"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output, got:\n%s", name, out)
		}
	}
}

func TestRequiresValidation_SatisfiedCommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-bash")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "bash is available") {
		t.Errorf("expected 'bash is available' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_SatisfiedRuntime(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-python")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "python is available") {
		t.Errorf("expected 'python is available' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_MissingCommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Missing requirements:") {
		t.Errorf("expected 'Missing requirements:' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "pi-nonexistent-tool-xyz") {
		t.Errorf("expected tool name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' in output, got:\n%s", out)
	}
	if strings.Contains(out, "this should never run") {
		t.Errorf("automation steps should not execute when requirements fail, got:\n%s", out)
	}
}

func TestRequiresValidation_ImpossibleVersion(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible-version")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Missing requirements:") {
		t.Errorf("expected 'Missing requirements:' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "python >= 99.0") {
		t.Errorf("expected 'python >= 99.0' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "99.0") {
		t.Errorf("expected version requirement in output, got:\n%s", out)
	}
	if strings.Contains(out, "this should never run") {
		t.Errorf("automation steps should not execute when requirements fail, got:\n%s", out)
	}
}

func TestRequiresValidation_NoRequirements(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "no-requires")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "no requirements needed") {
		t.Errorf("expected 'no requirements needed' in output, got:\n%s", out)
	}
}

func TestRequiresValidation_ErrorShowsInstallHint(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "run", "needs-impossible-version")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}
	if !strings.Contains(out, "install:") {
		t.Errorf("expected install hint in output, got:\n%s", out)
	}
}

// ===== pi doctor tests =====

func TestDoctor_AllSatisfied(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")

	// The workspace has needs-impossible and needs-impossible-version which will fail,
	// so the exit code should be 1
	if code != 1 {
		t.Fatalf("expected exit 1 (some requirements missing), got %d: %s", code, out)
	}

	// Satisfied requirements should show ✓
	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ for satisfied requirements, got:\n%s", out)
	}

	// needs-bash should be satisfied
	if !strings.Contains(out, "needs-bash") {
		t.Errorf("expected needs-bash in output, got:\n%s", out)
	}
}

func TestDoctor_ShowsMissingRequirements(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	// needs-impossible should show ✗
	if !strings.Contains(out, "needs-impossible") {
		t.Errorf("expected needs-impossible in output, got:\n%s", out)
	}
	if !strings.Contains(out, "✗") {
		t.Errorf("expected ✗ for missing requirements, got:\n%s", out)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' for missing command, got:\n%s", out)
	}
}

func TestDoctor_ShowsVersionMismatch(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	// needs-impossible-version requires python >= 99.0
	if !strings.Contains(out, "needs-impossible-version") {
		t.Errorf("expected needs-impossible-version in output, got:\n%s", out)
	}
	if !strings.Contains(out, "python >= 99.0") {
		t.Errorf("expected 'python >= 99.0' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "need >= 99.0") {
		t.Errorf("expected version mismatch message, got:\n%s", out)
	}
}

func TestDoctor_SkipsNoRequiresAutomations(t *testing.T) {
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, _ := runPi(t, dir, "doctor")

	// no-requires automation should NOT appear in doctor output
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "no-requires" {
			t.Errorf("doctor should skip automations without requires:, but found 'no-requires' in output:\n%s", out)
		}
	}
}

func TestDoctor_ShowsDetectedVersion(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, _ := runPi(t, dir, "doctor")

	// needs-python requires python (any version) — should show detected version
	if !strings.Contains(out, "needs-python") {
		t.Errorf("expected needs-python in output, got:\n%s", out)
	}
	// The detected version should appear in parentheses (e.g., "(3.13.0)")
	if !strings.Contains(out, "(") || !strings.Contains(out, ")") {
		t.Errorf("expected version in parentheses, got:\n%s", out)
	}
}

func TestDoctor_ShowsInstallHint(t *testing.T) {
	requirePython(t)
	dir := filepath.Join(examplesDir(), "requires-validation")
	out, code := runPi(t, dir, "doctor")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d: %s", code, out)
	}

	// needs-impossible-version has python which has a known install hint
	if !strings.Contains(out, "install") || !strings.Contains(out, "python") {
		t.Errorf("expected install hint for python, got:\n%s", out)
	}
}

func TestDoctor_HealthyWorkspace(t *testing.T) {
	// Create a workspace where all requirements are satisfied
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), []byte("project: healthy\n"), 0644); err != nil {
		t.Fatal(err)
	}
	piDir := filepath.Join(dir, ".pi")
	if err := os.MkdirAll(piDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(piDir, "needs-bash.yaml"), []byte(`name: needs-bash
description: Needs bash
requires:
  - command: bash
steps:
  - bash: echo ok
`), 0644); err != nil {
		t.Fatal(err)
	}

	out, code := runPi(t, dir, "doctor")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}
	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ for satisfied requirements, got:\n%s", out)
	}
	if strings.Contains(out, "✗") {
		t.Errorf("should not have ✗ in healthy workspace, got:\n%s", out)
	}
}

// ===== Runtime provisioning tests =====

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
	// Python should be available on the test system
	if !strings.Contains(out, "python is available") {
		t.Errorf("expected 'python is available' in output, got:\n%s", out)
	}
	// Should NOT have provisioned since python is already installed
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
	// If config parsing was broken, pi list would fail
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

// --- Step-level env: tests ---

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
