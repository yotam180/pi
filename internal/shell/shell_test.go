package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func TestGenerateShellFile_BasicShortcuts(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "myproject",
		Shortcuts: map[string]config.Shortcut{
			"up":   {Run: "docker/up"},
			"down": {Run: "docker/down"},
		},
	}

	content := GenerateShellFile(cfg, "/usr/local/bin/pi", "/home/dev/myproject")

	if !strings.Contains(content, "# PI shortcuts for myproject") {
		t.Error("expected header with project name")
	}
	if !strings.Contains(content, "function up()") {
		t.Error("expected function up()")
	}
	if !strings.Contains(content, "function down()") {
		t.Error("expected function down()")
	}
	if !strings.Contains(content, `cd "/home/dev/myproject"`) {
		t.Error("expected cd to repo root")
	}
	if !strings.Contains(content, "pi run docker/up") {
		t.Error("expected pi run docker/up")
	}
}

func TestGenerateShellFile_AnywhereShortcut(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "myproject",
		Shortcuts: map[string]config.Shortcut{
			"deploy": {Run: "deploy/push", Anywhere: true},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/home/dev/myproject")

	if strings.Contains(content, "(cd") {
		t.Error("anywhere shortcut should not cd")
	}
	if !strings.Contains(content, "--repo") {
		t.Error("anywhere shortcut should use --repo flag")
	}
	if !strings.Contains(content, `deploy/push "$@"`) {
		t.Error("expected args forwarding")
	}
}

func TestGenerateShellFile_Empty(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project:   "empty",
		Shortcuts: map[string]config.Shortcut{},
	}

	content := GenerateShellFile(cfg, "pi", "/root")
	if content != "" {
		t.Errorf("expected empty output for no shortcuts, got:\n%s", content)
	}
}

func TestGenerateShellFile_Sorted(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "sorted",
		Shortcuts: map[string]config.Shortcut{
			"zzz": {Run: "z"},
			"aaa": {Run: "a"},
			"mmm": {Run: "m"},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/root")
	aaaIdx := strings.Index(content, "function aaa()")
	mmmIdx := strings.Index(content, "function mmm()")
	zzzIdx := strings.Index(content, "function zzz()")

	if aaaIdx > mmmIdx || mmmIdx > zzzIdx {
		t.Error("functions should be in alphabetical order")
	}
}

func TestGenerateFunction_Default(t *testing.T) {
	sc := config.Shortcut{Run: "docker/up"}
	fn := generateFunction("dup", sc, "/usr/local/bin/pi", "/home/dev/repo")

	expected := `function dup() {
  (cd "/home/dev/repo" && /usr/local/bin/pi run docker/up "$@")
}
`
	if fn != expected {
		t.Errorf("unexpected function:\ngot:\n%s\nwant:\n%s", fn, expected)
	}
}

func TestGenerateFunction_Anywhere(t *testing.T) {
	sc := config.Shortcut{Run: "deploy/push", Anywhere: true}
	fn := generateFunction("deploy", sc, "/usr/local/bin/pi", "/home/dev/repo")

	expected := `function deploy() {
  /usr/local/bin/pi run --repo "/home/dev/repo" deploy/push "$@"
}
`
	if fn != expected {
		t.Errorf("unexpected function:\ngot:\n%s\nwant:\n%s", fn, expected)
	}
}

func TestInstallAndUninstall(t *testing.T) {
	// Override home dir for this test
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create .zshrc
	zshrc := filepath.Join(tmpHome, ".zshrc")
	os.WriteFile(zshrc, []byte("# existing config\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	// Install
	shellPath, err := Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify shell file
	data, err := os.ReadFile(shellPath)
	if err != nil {
		t.Fatalf("reading shell file: %v", err)
	}
	if !strings.Contains(string(data), "function hello()") {
		t.Error("shell file should contain hello function")
	}

	// Verify source line
	zshData, _ := os.ReadFile(zshrc)
	if !strings.Contains(string(zshData), "# Added by PI") {
		t.Error(".zshrc should contain source line")
	}

	// Install again — idempotent
	_, err = Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}
	zshData, _ = os.ReadFile(zshrc)
	if strings.Count(string(zshData), "# Added by PI") != 1 {
		t.Error("source line should appear only once after re-install")
	}

	// Uninstall
	if err := Uninstall("testproj"); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	if _, err := os.Stat(shellPath); !os.IsNotExist(err) {
		t.Error("shell file should be removed after uninstall")
	}

	// Source line should be cleaned up (no repos left)
	zshData, _ = os.ReadFile(zshrc)
	if strings.Contains(string(zshData), "# Added by PI") {
		t.Error("source line should be removed when no repos remain")
	}
}

func TestInstall_NoShortcuts(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &config.ProjectConfig{
		Project:   "empty",
		Shortcuts: map[string]config.Shortcut{},
	}

	_, err := Install(cfg, "pi", "/repo")
	if err == nil {
		t.Fatal("expected error for empty shortcuts")
	}
	if !strings.Contains(err.Error(), "no shortcuts") {
		t.Errorf("expected 'no shortcuts' error, got: %v", err)
	}
}

func TestListInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, "proj-a.sh"), []byte("# a"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "proj-b.sh"), []byte("# b"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0] != "proj-a" || projects[1] != "proj-b" {
		t.Errorf("unexpected projects: %v", projects)
	}
}

func TestListInstalled_Empty(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestUninstall_KeepsSourceLineIfOtherReposExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	zshrc := filepath.Join(tmpHome, ".zshrc")
	os.WriteFile(zshrc, []byte("# existing\n"), 0o644)

	cfg1 := &config.ProjectConfig{
		Project: "proj1",
		Shortcuts: map[string]config.Shortcut{
			"a": {Run: "auto/a"},
		},
	}
	cfg2 := &config.ProjectConfig{
		Project: "proj2",
		Shortcuts: map[string]config.Shortcut{
			"b": {Run: "auto/b"},
		},
	}

	Install(cfg1, "pi", "/repo1")
	Install(cfg2, "pi", "/repo2")

	// Uninstall proj1 — proj2 still exists
	Uninstall("proj1")

	zshData, _ := os.ReadFile(zshrc)
	if !strings.Contains(string(zshData), "# Added by PI") {
		t.Error("source line should remain when other repos still installed")
	}
}

func TestGenerateShellFile_WithMapping(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"dlogs": {
				Run: "docker/logs",
				With: map[string]string{
					"service": "$1",
					"tail":    "$2",
				},
			},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/repo")
	if !strings.Contains(content, "--with service=\"$1\"") {
		t.Errorf("expected --with service=\"$1\" in output, got:\n%s", content)
	}
	if !strings.Contains(content, "--with tail=\"$2\"") {
		t.Errorf("expected --with tail=\"$2\" in output, got:\n%s", content)
	}
	if strings.Contains(content, "\"$@\"") {
		t.Errorf("with-mapped shortcut should not include $@ passthrough, got:\n%s", content)
	}
}

func TestGenerateShellFile_WithLiteralValue(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"dlogs-short": {
				Run: "docker/logs",
				With: map[string]string{
					"tail":    "50",
					"service": "$1",
				},
			},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/repo")
	if !strings.Contains(content, "--with service=\"$1\"") {
		t.Errorf("expected --with service=\"$1\" in output, got:\n%s", content)
	}
	if !strings.Contains(content, `--with tail="50"`) {
		t.Errorf("expected --with tail=\"50\" in output, got:\n%s", content)
	}
}

func TestGenerateShellFile_WithAnywhereAndWith(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"deploy": {
				Run:      "deploy/push",
				Anywhere: true,
				With:     map[string]string{"env": "$1"},
			},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/repo")
	if !strings.Contains(content, "--repo") {
		t.Errorf("anywhere shortcut with 'with' should have --repo, got:\n%s", content)
	}
	if !strings.Contains(content, "--with env=\"$1\"") {
		t.Errorf("expected --with env=\"$1\", got:\n%s", content)
	}
}
