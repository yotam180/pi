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

// --- AddSetupEntry tests ---

func TestAddSetupEntry_BareString(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", "project: test\n")

	entry := SetupEntry{Run: "pi:install-uv"}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-uv" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-uv")
	}
}

func TestAddSetupEntry_WithVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", "project: test\n")

	entry := SetupEntry{
		Run:  "pi:install-python",
		With: map[string]string{"version": "3.13"},
	}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-python" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-python")
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q, want %q", cfg.Setup[0].With["version"], "3.13")
	}
}

func TestAddSetupEntry_WithIf(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", "project: test\n")

	entry := SetupEntry{
		Run: "pi:install-homebrew",
		If:  "os.macos",
	}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].If != "os.macos" {
		t.Errorf("if = %q, want %q", cfg.Setup[0].If, "os.macos")
	}
}

func TestAddSetupEntry_WithIfAndVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", "project: test\n")

	entry := SetupEntry{
		Run:  "pi:install-python",
		If:   "os.macos",
		With: map[string]string{"version": "3.13"},
	}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-python" {
		t.Errorf("run = %q", cfg.Setup[0].Run)
	}
	if cfg.Setup[0].If != "os.macos" {
		t.Errorf("if = %q", cfg.Setup[0].If)
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q", cfg.Setup[0].With["version"])
	}
}

func TestAddSetupEntry_Duplicate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - pi:install-uv
`)

	entry := SetupEntry{Run: "pi:install-uv"}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected DuplicateSetupEntryError")
	}

	dup, ok := err.(*DuplicateSetupEntryError)
	if !ok {
		t.Fatalf("expected *DuplicateSetupEntryError, got %T: %v", err, err)
	}
	if dup.Run != "pi:install-uv" {
		t.Errorf("dup.Run = %q, want %q", dup.Run, "pi:install-uv")
	}
}

func TestAddSetupEntry_DuplicateWithVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - run: pi:install-python
    with:
      version: "3.13"
`)

	entry := SetupEntry{
		Run:  "pi:install-python",
		With: map[string]string{"version": "3.13"},
	}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected DuplicateSetupEntryError")
	}
	if _, ok := err.(*DuplicateSetupEntryError); !ok {
		t.Fatalf("expected *DuplicateSetupEntryError, got %T: %v", err, err)
	}
}

func TestAddSetupEntry_SameRunDifferentVersion_NotDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - run: pi:install-python
    with:
      version: "3.12"
`)

	entry := SetupEntry{
		Run:  "pi:install-python",
		With: map[string]string{"version": "3.13"},
	}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
}

func TestAddSetupEntry_AppendsToExistingBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

setup:
  - setup/install-deps
`)

	entry := SetupEntry{Run: "pi:install-uv"}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "setup/install-deps" {
		t.Errorf("setup[0].Run = %q", cfg.Setup[0].Run)
	}
	if cfg.Setup[1].Run != "pi:install-uv" {
		t.Errorf("setup[1].Run = %q", cfg.Setup[1].Run)
	}
}

func TestAddSetupEntry_CreatesSetupBlock(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

shortcuts:
  up: docker/up
`)

	entry := SetupEntry{Run: "pi:install-uv"}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "setup:") {
		t.Error("file should contain 'setup:'")
	}
	if !strings.Contains(content, "pi:install-uv") {
		t.Error("file should contain the added setup entry")
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if len(cfg.Shortcuts) != 1 {
		t.Errorf("shortcuts count = %d, want 1 (preserved)", len(cfg.Shortcuts))
	}
}

func TestAddSetupEntry_PreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	original := `project: my-project

shortcuts:
  up: docker/up
  down: docker/down

packages:
  - org/pkg@v1.0
`
	writeFile(t, dir, "pi.yaml", original)

	entry := SetupEntry{Run: "pi:install-uv"}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "pi.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "project: my-project") {
		t.Error("project should be preserved")
	}
	if !strings.Contains(content, "up: docker/up") {
		t.Error("shortcuts should be preserved")
	}
	if !strings.Contains(content, "org/pkg@v1.0") {
		t.Error("packages should be preserved")
	}
}

func TestAddSetupEntry_MultipleAdds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", "project: test\n")

	entries := []SetupEntry{
		{Run: "pi:install-homebrew", If: "os.macos"},
		{Run: "pi:install-uv"},
		{Run: "pi:install-python", With: map[string]string{"version": "3.13"}},
	}

	for i, entry := range entries {
		if err := AddSetupEntry(dir, entry); err != nil {
			t.Fatalf("add %d failed: %v", i, err)
		}
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 3 {
		t.Fatalf("setup count = %d, want 3", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-homebrew" {
		t.Errorf("setup[0].Run = %q", cfg.Setup[0].Run)
	}
	if cfg.Setup[1].Run != "pi:install-uv" {
		t.Errorf("setup[1].Run = %q", cfg.Setup[1].Run)
	}
	if cfg.Setup[2].Run != "pi:install-python" {
		t.Errorf("setup[2].Run = %q", cfg.Setup[2].Run)
	}
}

func TestAddSetupEntry_MissingPiYaml(t *testing.T) {
	dir := t.TempDir()

	entry := SetupEntry{Run: "pi:install-uv"}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected error for missing pi.yaml")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestFormatSetupEntry_BareString(t *testing.T) {
	entry := SetupEntry{Run: "pi:install-uv"}
	got := FormatSetupEntry(entry)
	want := "  - pi:install-uv"
	if got != want {
		t.Errorf("FormatSetupEntry() = %q, want %q", got, want)
	}
}

func TestFormatSetupEntry_WithVersion(t *testing.T) {
	entry := SetupEntry{
		Run:  "pi:install-python",
		With: map[string]string{"version": "3.13"},
	}
	got := FormatSetupEntry(entry)
	if !strings.Contains(got, "run: pi:install-python") {
		t.Errorf("should contain run:, got: %q", got)
	}
	if !strings.Contains(got, "with:") {
		t.Errorf("should contain with:, got: %q", got)
	}
	if !strings.Contains(got, `version: "3.13"`) {
		t.Errorf("should contain version, got: %q", got)
	}
}

func TestFormatSetupEntry_WithIf(t *testing.T) {
	entry := SetupEntry{
		Run: "pi:install-homebrew",
		If:  "os.macos",
	}
	got := FormatSetupEntry(entry)
	if !strings.Contains(got, "run: pi:install-homebrew") {
		t.Errorf("should contain run:, got: %q", got)
	}
	if !strings.Contains(got, "if: os.macos") {
		t.Errorf("should contain if:, got: %q", got)
	}
}

func TestDuplicateSetupEntryError_Message(t *testing.T) {
	err := &DuplicateSetupEntryError{Run: "pi:install-uv"}
	got := err.Error()
	if !strings.Contains(got, "pi:install-uv") {
		t.Errorf("error should contain run value, got: %s", got)
	}
	if !strings.Contains(got, "already declared") {
		t.Errorf("error should say 'already declared', got: %s", got)
	}
}

func TestAddSetupEntry_ExistingBlockWithFollowingContent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test

setup:
  - setup/install-deps

shortcuts:
  up: docker/up
`)

	entry := SetupEntry{Run: "pi:install-uv"}
	if err := AddSetupEntry(dir, entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}

	if len(cfg.Setup) != 2 {
		t.Fatalf("setup count = %d, want 2", len(cfg.Setup))
	}
	if cfg.Setup[1].Run != "pi:install-uv" {
		t.Errorf("setup[1].Run = %q, want %q", cfg.Setup[1].Run, "pi:install-uv")
	}
	if len(cfg.Shortcuts) != 1 {
		t.Errorf("shortcuts preserved = %d, want 1", len(cfg.Shortcuts))
	}
}
