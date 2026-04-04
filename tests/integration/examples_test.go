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
