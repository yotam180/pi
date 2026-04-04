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
