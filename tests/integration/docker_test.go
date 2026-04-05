package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

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
