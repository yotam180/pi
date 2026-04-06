package config

import (
	"errors"
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

	var dup *DuplicatePackageError
	if !errors.As(err, &dup) {
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

	var dup *DuplicateSetupEntryError
	if !errors.As(err, &dup) {
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
	var dupTarget *DuplicateSetupEntryError
	if !errors.As(err, &dupTarget) {
		t.Fatalf("expected *DuplicateSetupEntryError, got %T: %v", err, err)
	}
}

func TestAddSetupEntry_SameRunDifferentVersion_Replaces(t *testing.T) {
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
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected ReplacedSetupEntryError")
	}
	var replTarget *ReplacedSetupEntryError
	if !errors.As(err, &replTarget) {
		t.Fatalf("expected *ReplacedSetupEntryError, got %T: %v", err, err)
	}

	cfg, err2 := Load(dir)
	if err2 != nil {
		t.Fatalf("failed to reload: %v", err2)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1 (replaced in-place)", len(cfg.Setup))
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q, want %q", cfg.Setup[0].With["version"], "3.13")
	}
}

func TestAddSetupEntry_SameRunBareToVersioned_Replaces(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - pi:install-node
`)

	entry := SetupEntry{
		Run:  "pi:install-node",
		With: map[string]string{"version": "22"},
	}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected ReplacedSetupEntryError")
	}
	var replTarget *ReplacedSetupEntryError
	if !errors.As(err, &replTarget) {
		t.Fatalf("expected *ReplacedSetupEntryError, got %T: %v", err, err)
	}

	cfg, err2 := Load(dir)
	if err2 != nil {
		t.Fatalf("failed to reload: %v", err2)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-node" {
		t.Errorf("run = %q, want %q", cfg.Setup[0].Run, "pi:install-node")
	}
	if cfg.Setup[0].With["version"] != "22" {
		t.Errorf("with.version = %q, want %q", cfg.Setup[0].With["version"], "22")
	}
}

func TestAddSetupEntry_ReplacePreservesPosition(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - pi:install-uv
  - pi:install-node
  - pi:install-python
`)

	entry := SetupEntry{
		Run:  "pi:install-node",
		With: map[string]string{"version": "22"},
	}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected ReplacedSetupEntryError")
	}
	var replTarget *ReplacedSetupEntryError
	if !errors.As(err, &replTarget) {
		t.Fatalf("expected *ReplacedSetupEntryError, got %T: %v", err, err)
	}

	cfg, err2 := Load(dir)
	if err2 != nil {
		t.Fatalf("failed to reload: %v", err2)
	}
	if len(cfg.Setup) != 3 {
		t.Fatalf("setup count = %d, want 3", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-uv" {
		t.Errorf("setup[0].Run = %q, want pi:install-uv", cfg.Setup[0].Run)
	}
	if cfg.Setup[1].Run != "pi:install-node" {
		t.Errorf("setup[1].Run = %q, want pi:install-node", cfg.Setup[1].Run)
	}
	if cfg.Setup[1].With["version"] != "22" {
		t.Errorf("setup[1].with.version = %q, want 22", cfg.Setup[1].With["version"])
	}
	if cfg.Setup[2].Run != "pi:install-python" {
		t.Errorf("setup[2].Run = %q, want pi:install-python", cfg.Setup[2].Run)
	}
}

func TestAddSetupEntry_ReplaceVersionedToBare(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - run: pi:install-node
    with:
      version: "22"
`)

	entry := SetupEntry{Run: "pi:install-node"}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected ReplacedSetupEntryError")
	}
	var replTarget *ReplacedSetupEntryError
	if !errors.As(err, &replTarget) {
		t.Fatalf("expected *ReplacedSetupEntryError, got %T: %v", err, err)
	}

	cfg, err2 := Load(dir)
	if err2 != nil {
		t.Fatalf("failed to reload: %v", err2)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].Run != "pi:install-node" {
		t.Errorf("run = %q", cfg.Setup[0].Run)
	}
	if len(cfg.Setup[0].With) != 0 {
		t.Errorf("with should be empty, got %v", cfg.Setup[0].With)
	}
}

func TestReplacedSetupEntryError_Message(t *testing.T) {
	err := &ReplacedSetupEntryError{Run: "pi:install-node"}
	got := err.Error()
	if !strings.Contains(got, "pi:install-node") {
		t.Errorf("error should contain run value, got: %s", got)
	}
	if !strings.Contains(got, "replaced") {
		t.Errorf("error should say 'replaced', got: %s", got)
	}
}

func TestReplaceSetupEntry_MultiLineToMultiLine(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pi.yaml", `project: test
setup:
  - run: pi:install-python
    with:
      version: "3.12"

shortcuts:
  up: docker/up
`)

	entry := SetupEntry{
		Run:  "pi:install-python",
		With: map[string]string{"version": "3.13"},
	}
	err := AddSetupEntry(dir, entry)
	if err == nil {
		t.Fatal("expected ReplacedSetupEntryError")
	}
	var replTarget *ReplacedSetupEntryError
	if !errors.As(err, &replTarget) {
		t.Fatalf("expected *ReplacedSetupEntryError, got %T: %v", err, err)
	}

	cfg, err2 := Load(dir)
	if err2 != nil {
		t.Fatalf("failed to reload: %v", err2)
	}
	if len(cfg.Setup) != 1 {
		t.Fatalf("setup count = %d, want 1", len(cfg.Setup))
	}
	if cfg.Setup[0].With["version"] != "3.13" {
		t.Errorf("with.version = %q, want 3.13", cfg.Setup[0].With["version"])
	}
	if len(cfg.Shortcuts) != 1 {
		t.Errorf("shortcuts count = %d, want 1 (preserved)", len(cfg.Shortcuts))
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

// --- Shared YAML block helper unit tests ---

func TestFindBlockIndex_Found(t *testing.T) {
	lines := []string{"project: test", "", "setup:", "  - pi:install-uv"}
	idx := findBlockIndex(lines, "setup")
	if idx != 2 {
		t.Errorf("findBlockIndex = %d, want 2", idx)
	}
}

func TestFindBlockIndex_NotFound(t *testing.T) {
	lines := []string{"project: test", "", "shortcuts:", "  up: docker/up"}
	idx := findBlockIndex(lines, "setup")
	if idx != -1 {
		t.Errorf("findBlockIndex = %d, want -1", idx)
	}
}

func TestFindBlockIndex_MultipleBlocks(t *testing.T) {
	lines := []string{"project: test", "setup:", "  - a", "packages:", "  - b"}
	idx := findBlockIndex(lines, "packages")
	if idx != 3 {
		t.Errorf("findBlockIndex = %d, want 3", idx)
	}
}

func TestAppendToBlock_Simple(t *testing.T) {
	lines := []string{"setup:", "  - existing"}
	got := appendToBlock(lines, 0, "  - new-entry")
	if !strings.Contains(got, "- existing") {
		t.Error("should preserve existing entry")
	}
	if !strings.Contains(got, "- new-entry") {
		t.Error("should contain new entry")
	}
}

func TestAppendToBlock_WithTrailingBlanks(t *testing.T) {
	lines := []string{"setup:", "  - existing", "", "shortcuts:"}
	got := appendToBlock(lines, 0, "  - new-entry")
	if !strings.Contains(got, "- new-entry") {
		t.Error("should contain new entry")
	}
	if !strings.Contains(got, "shortcuts:") {
		t.Error("should preserve following block")
	}
}

func TestAppendNewBlock(t *testing.T) {
	content := "project: test\n"
	got := appendNewBlock(content, "setup", "  - pi:install-uv")
	if !strings.Contains(got, "setup:\n  - pi:install-uv") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestAppendNewBlock_StripsTrailingNewlines(t *testing.T) {
	content := "project: test\n\n\n"
	got := appendNewBlock(content, "packages", "  - org/repo@v1.0")
	if strings.HasPrefix(got, "project: test\n\n\n") {
		t.Error("should strip trailing newlines before appending")
	}
	if !strings.Contains(got, "packages:\n  - org/repo@v1.0") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestInsertIntoBlock_ExistingBlock(t *testing.T) {
	content := "project: test\n\npackages:\n  - existing@v1.0\n"
	got := insertIntoBlock(content, "packages", "  - new@v2.0")
	if !strings.Contains(got, "- existing@v1.0") {
		t.Error("should preserve existing entry")
	}
	if !strings.Contains(got, "- new@v2.0") {
		t.Error("should contain new entry")
	}
}

func TestInsertIntoBlock_NoBlock(t *testing.T) {
	content := "project: test\n"
	got := insertIntoBlock(content, "packages", "  - org/repo@v1.0")
	if !strings.Contains(got, "packages:\n  - org/repo@v1.0") {
		t.Errorf("should create block, got: %q", got)
	}
}
