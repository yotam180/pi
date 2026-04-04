package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRoot_InProjectDir(t *testing.T) {
	dir := t.TempDir()
	writePiYaml(t, dir)

	root, err := FindRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestFindRoot_FromSubdirectory(t *testing.T) {
	root := t.TempDir()
	writePiYaml(t, root)

	sub := filepath.Join(root, "src", "nested", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != root {
		t.Errorf("expected %s, got %s", root, found)
	}
}

func TestFindRoot_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := FindRoot(dir)
	if err == nil {
		t.Fatal("expected error when pi.yaml not found")
	}
	if !strings.Contains(err.Error(), "pi.yaml") {
		t.Errorf("expected error to mention pi.yaml, got: %v", err)
	}
}

func TestFindRoot_PicksClosest(t *testing.T) {
	outer := t.TempDir()
	writePiYaml(t, outer)

	inner := filepath.Join(outer, "subproject")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	writePiYaml(t, inner)

	sub := filepath.Join(inner, "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != inner {
		t.Errorf("expected closest root %s, got %s", inner, found)
	}
}

func writePiYaml(t *testing.T, dir string) {
	t.Helper()
	content := []byte("project: test\n")
	if err := os.WriteFile(filepath.Join(dir, "pi.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}
