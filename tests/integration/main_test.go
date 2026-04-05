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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v\n%s", args, err, string(out))
		}
	}
	return string(out), exitCode
}

func runPiStdout(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	exitCode := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v", args, err)
		}
	}
	return stdout.String(), exitCode
}

func runPiSplit(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	exitCode = 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v", args, err)
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

func runPiWithEnv(t *testing.T, dir string, env []string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(piBinary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("running pi %v: %v\n%s", args, err, string(out))
		}
	}
	return string(out), exitCode
}
