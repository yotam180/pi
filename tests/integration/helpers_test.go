package integration

import (
	"os/exec"
	"testing"
)

func requirePython(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
}

func requireNode(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not available")
	}
}

func requireTsx(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tsx"); err != nil {
		t.Skip("tsx not available")
	}
}
