package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddPackage_GitHubSimple(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
`)

	entry := PackageEntry{Source: "yotam180/pi-common@v1.2"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "yotam180/pi-common@v1.2" {
		t.Errorf("source = %q, want %q", cfg.Packages[0].Source, "yotam180/pi-common@v1.2")
	}
}

func TestAddPackage_FileWithAlias(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
`)

	entry := PackageEntry{Source: "file:~/my-automations", As: "mytools"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "file:~/my-automations" {
		t.Errorf("source = %q, want %q", cfg.Packages[0].Source, "file:~/my-automations")
	}
	if cfg.Packages[0].As != "mytools" {
		t.Errorf("as = %q, want %q", cfg.Packages[0].As, "mytools")
	}
}

func TestAddPackage_IdempotentDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
packages:
  - yotam180/pi-common@v1.2
`)

	entry := PackageEntry{Source: "yotam180/pi-common@v1.2"}
	err := AddPackage(dir, entry)
	if err == nil {
		t.Fatal("expected DuplicatePackageError")
	}

	dup, ok := err.(*DuplicatePackageError)
	if !ok {
		t.Fatalf("expected *DuplicatePackageError, got %T: %v", err, err)
	}
	if dup.Source != "yotam180/pi-common@v1.2" {
		t.Errorf("dup.Source = %q, want %q", dup.Source, "yotam180/pi-common@v1.2")
	}
}

func TestAddPackage_AppendsToExistingBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

packages:
  - existing-org/existing-pkg@v1.0
`)

	entry := PackageEntry{Source: "yotam180/pi-common@v1.2"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Packages) != 2 {
		t.Fatalf("packages count = %d, want 2", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "existing-org/existing-pkg@v1.0" {
		t.Errorf("packages[0].Source = %q, want %q", cfg.Packages[0].Source, "existing-org/existing-pkg@v1.0")
	}
	if cfg.Packages[1].Source != "yotam180/pi-common@v1.2" {
		t.Errorf("packages[1].Source = %q, want %q", cfg.Packages[1].Source, "yotam180/pi-common@v1.2")
	}
}

func TestAddPackage_CreatesPackagesBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

shortcuts:
  up: docker/up
`)

	entry := PackageEntry{Source: "yotam180/pi-common@v1.2"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "packages:") {
		t.Error("file should contain 'packages:'")
	}
	if !strings.Contains(content, "yotam180/pi-common@v1.2") {
		t.Error("file should contain the added package source")
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if len(cfg.Packages) != 1 {
		t.Fatalf("packages count = %d, want 1", len(cfg.Packages))
	}

	if len(cfg.Shortcuts) != 1 {
		t.Errorf("shortcuts count = %d, want 1 (preserved)", len(cfg.Shortcuts))
	}
}

func TestAddPackage_MissingPiYaml(t *testing.T) {
	dir := t.TempDir()

	entry := PackageEntry{Source: "yotam180/pi-common@v1.2"}
	err := AddPackage(dir, entry)
	if err == nil {
		t.Fatal("expected error for missing pi.yaml")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestAddPackage_PreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	original := `project: my-project

shortcuts:
  up: docker/up
  down: docker/down

setup:
  - setup/install-deps
`
	writeFile(t, dir, "pi.yaml", original)

	entry := PackageEntry{Source: "org/pkg@v1.0"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "project: my-project") {
		t.Error("project field should be preserved")
	}
	if !strings.Contains(content, "up: docker/up") {
		t.Error("shortcuts should be preserved")
	}
	if !strings.Contains(content, "setup/install-deps") {
		t.Error("setup entries should be preserved")
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if cfg.Project != "my-project" {
		t.Errorf("project = %q, want %q", cfg.Project, "my-project")
	}
}

func TestAddPackage_MultipleAdds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
`)

	first := PackageEntry{Source: "org-a/pkg-a@v1.0"}
	if err := AddPackage(dir, first); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	second := PackageEntry{Source: "org-b/pkg-b@v2.0"}
	if err := AddPackage(dir, second); err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	third := PackageEntry{Source: "file:~/shared", As: "shared"}
	if err := AddPackage(dir, third); err != nil {
		t.Fatalf("third add failed: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Packages) != 3 {
		t.Fatalf("packages count = %d, want 3", len(cfg.Packages))
	}
	if cfg.Packages[0].Source != "org-a/pkg-a@v1.0" {
		t.Errorf("packages[0] = %q", cfg.Packages[0].Source)
	}
	if cfg.Packages[1].Source != "org-b/pkg-b@v2.0" {
		t.Errorf("packages[1] = %q", cfg.Packages[1].Source)
	}
	if cfg.Packages[2].Source != "file:~/shared" {
		t.Errorf("packages[2] = %q", cfg.Packages[2].Source)
	}
	if cfg.Packages[2].As != "shared" {
		t.Errorf("packages[2].As = %q, want %q", cfg.Packages[2].As, "shared")
	}
}

func TestFormatPackageEntry_Simple(t *testing.T) {
	entry := PackageEntry{Source: "org/repo@v1.0"}
	got := formatPackageEntry(entry)
	want := "  - org/repo@v1.0"
	if got != want {
		t.Errorf("formatPackageEntry() = %q, want %q", got, want)
	}
}

func TestFormatPackageEntry_WithAlias(t *testing.T) {
	entry := PackageEntry{Source: "file:~/path", As: "mytools"}
	got := formatPackageEntry(entry)
	want := "  - source: file:~/path\n    as: mytools"
	if got != want {
		t.Errorf("formatPackageEntry() = %q, want %q", got, want)
	}
}

func TestInsertPackageEntry_NoExistingBlock(t *testing.T) {
	content := "project: test\n"
	entry := PackageEntry{Source: "org/repo@v1.0"}
	got, err := insertPackageEntry(content, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "packages:\n  - org/repo@v1.0") {
		t.Errorf("expected packages block, got:\n%s", got)
	}
}

func TestInsertPackageEntry_ExistingBlock(t *testing.T) {
	content := "project: test\n\npackages:\n  - existing@v1.0\n"
	entry := PackageEntry{Source: "new@v2.0"}
	got, err := insertPackageEntry(content, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "- existing@v1.0") {
		t.Error("existing entry should be preserved")
	}
	if !strings.Contains(got, "- new@v2.0") {
		t.Error("new entry should be present")
	}
}

func TestDuplicatePackageError_Message(t *testing.T) {
	err := &DuplicatePackageError{Source: "org/repo@v1.0"}
	got := err.Error()
	if !strings.Contains(got, "org/repo@v1.0") {
		t.Errorf("error message should contain source, got: %s", got)
	}
	if !strings.Contains(got, "already declared") {
		t.Errorf("error message should say 'already declared', got: %s", got)
	}
}

func TestAddPackage_ExistingBlockWithFollowingContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

packages:
  - existing@v1.0

shortcuts:
  up: docker/up
`)

	entry := PackageEntry{Source: "new@v2.0"}
	if err := AddPackage(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Packages) != 2 {
		t.Fatalf("packages count = %d, want 2", len(cfg.Packages))
	}
	if cfg.Packages[1].Source != "new@v2.0" {
		t.Errorf("packages[1].Source = %q, want %q", cfg.Packages[1].Source, "new@v2.0")
	}

	if len(cfg.Shortcuts) != 1 {
		t.Errorf("shortcuts preserved = %d, want 1", len(cfg.Shortcuts))
	}
}
