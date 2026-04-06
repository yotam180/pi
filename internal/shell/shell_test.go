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
	if !strings.Contains(content, "PI_PARENT_EVAL_FILE") {
		t.Error("expected PI_PARENT_EVAL_FILE eval wrapper")
	}
	if !strings.Contains(content, "function pi-setup-myproject()") {
		t.Error("expected pi-setup helper function")
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

	// The deploy function itself should not cd (pi-setup helper does cd, which is fine)
	deployFn := extractFunction(content, "deploy")
	if strings.Contains(deployFn, "(cd") {
		t.Error("anywhere shortcut should not cd")
	}
	if !strings.Contains(deployFn, "--repo") {
		t.Error("anywhere shortcut should use --repo flag")
	}
	if !strings.Contains(deployFn, `deploy/push "$@"`) {
		t.Error("expected args forwarding")
	}
}

// extractFunction extracts a single function definition from shell content by name.
func extractFunction(content, name string) string {
	marker := "function " + name + "()"
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx:]
	end := strings.Index(rest[1:], "\nfunction ")
	if end < 0 {
		return rest
	}
	return rest[:end+1]
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
	fn := generateFunction("dup", sc, "/usr/local/bin/pi", "/home/dev/repo", DefaultDialect())

	if !strings.Contains(fn, "function dup()") {
		t.Errorf("expected function dup(), got:\n%s", fn)
	}
	if !strings.Contains(fn, `cd "/home/dev/repo"`) {
		t.Errorf("expected cd to repo root, got:\n%s", fn)
	}
	if !strings.Contains(fn, `/usr/local/bin/pi run docker/up "$@"`) {
		t.Errorf("expected pi run command, got:\n%s", fn)
	}
	if !strings.Contains(fn, "PI_PARENT_EVAL_FILE") {
		t.Errorf("expected PI_PARENT_EVAL_FILE eval wrapper, got:\n%s", fn)
	}
	if !strings.Contains(fn, `source "$_pi_eval_file"`) {
		t.Errorf("expected source of eval file, got:\n%s", fn)
	}
	if !strings.Contains(fn, `rm -f "$_pi_eval_file"`) {
		t.Errorf("expected cleanup of eval file, got:\n%s", fn)
	}
}

func TestGenerateFunction_Anywhere(t *testing.T) {
	sc := config.Shortcut{Run: "deploy/push", Anywhere: true}
	fn := generateFunction("deploy", sc, "/usr/local/bin/pi", "/home/dev/repo", DefaultDialect())

	if !strings.Contains(fn, "function deploy()") {
		t.Errorf("expected function deploy(), got:\n%s", fn)
	}
	if strings.Contains(fn, "(cd") {
		t.Errorf("anywhere shortcut should not cd, got:\n%s", fn)
	}
	if !strings.Contains(fn, `--repo "/home/dev/repo"`) {
		t.Errorf("expected --repo flag, got:\n%s", fn)
	}
	if !strings.Contains(fn, "PI_PARENT_EVAL_FILE") {
		t.Errorf("expected PI_PARENT_EVAL_FILE eval wrapper, got:\n%s", fn)
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
	// The dlogs function itself should not contain "$@" — only the pi-setup helper uses it
	dlogsFn := extractFunction(content, "dlogs")
	if strings.Contains(dlogsFn, "\"$@\"") {
		t.Errorf("with-mapped shortcut should not include $@ passthrough, got:\n%s", dlogsFn)
	}
}

func TestGenerateShellFile_SetupHelper(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "myproject",
		Shortcuts: map[string]config.Shortcut{
			"up": {Run: "docker/up"},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/home/dev/myproject")

	if !strings.Contains(content, "function pi-setup-myproject()") {
		t.Error("expected pi-setup-myproject helper function")
	}
	setupFn := extractFunction(content, "pi-setup-myproject")
	if !strings.Contains(setupFn, "PI_PARENT_EVAL_FILE") {
		t.Error("setup helper should use PI_PARENT_EVAL_FILE")
	}
	if !strings.Contains(setupFn, "pi setup") {
		t.Error("setup helper should call pi setup")
	}
	if !strings.Contains(setupFn, `source "$_pi_eval_file"`) {
		t.Error("setup helper should source eval file")
	}
}

func TestGenerateShellFile_EvalWrapperExitCode(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"build": {Run: "build/run"},
		},
	}

	content := GenerateShellFile(cfg, "pi", "/repo")
	buildFn := extractFunction(content, "build")
	if !strings.Contains(buildFn, "return $_pi_exit") {
		t.Error("eval wrapper should preserve exit code")
	}
	if !strings.Contains(buildFn, "local _pi_exit=$?") {
		t.Error("eval wrapper should capture exit code")
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

func TestGenerateGlobalWrapper(t *testing.T) {
	content := GenerateGlobalWrapper("/usr/local/bin/pi")

	if !strings.Contains(content, "function pi()") {
		t.Error("expected function pi()")
	}
	if !strings.Contains(content, "command /usr/local/bin/pi") {
		t.Error("expected 'command' keyword to avoid recursion")
	}
	if !strings.Contains(content, "PI_PARENT_EVAL_FILE") {
		t.Error("expected PI_PARENT_EVAL_FILE in wrapper")
	}
	if !strings.Contains(content, `source "$_pi_eval_file"`) {
		t.Error("expected source of eval file")
	}
	if !strings.Contains(content, `rm -f "$_pi_eval_file"`) {
		t.Error("expected cleanup of eval file")
	}
	if !strings.Contains(content, `return $_pi_exit`) {
		t.Error("expected exit code preservation")
	}
}

func TestInstall_CreatesGlobalWrapper(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	_, err := Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	wrapperPath := filepath.Join(tmpHome, piShellDir, piWrapperFile)
	data, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("reading wrapper file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "function pi()") {
		t.Error("wrapper file should contain pi() function")
	}
	if !strings.Contains(content, "command pi") {
		t.Error("wrapper file should use 'command' to avoid recursion")
	}
}

func TestUninstall_RemovesWrapperWhenLastProject(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	cfg := &config.ProjectConfig{
		Project: "only-proj",
		Shortcuts: map[string]config.Shortcut{
			"a": {Run: "auto/a"},
		},
	}

	Install(cfg, "pi", "/repo")
	wrapperPath := filepath.Join(tmpHome, piShellDir, piWrapperFile)
	if _, err := os.Stat(wrapperPath); err != nil {
		t.Fatalf("wrapper should exist after install: %v", err)
	}

	Uninstall("only-proj")
	if _, err := os.Stat(wrapperPath); !os.IsNotExist(err) {
		t.Error("wrapper should be removed when last project is uninstalled")
	}
}

func TestUninstall_KeepsWrapperIfOtherProjectsExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

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

	Uninstall("proj1")

	wrapperPath := filepath.Join(tmpHome, piShellDir, piWrapperFile)
	if _, err := os.Stat(wrapperPath); err != nil {
		t.Error("wrapper should remain when other projects still installed")
	}
}

func TestListInstalled_ExcludesWrapper(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, piWrapperFile), []byte("# wrapper"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "myproj.sh"), []byte("# shortcuts"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d: %v", len(projects), projects)
	}
	if projects[0] != "myproj" {
		t.Errorf("expected myproj, got %s", projects[0])
	}
}

func TestGenerateFunction_EvalInsideSubshell(t *testing.T) {
	sc := config.Shortcut{Run: "docker/up"}
	fn := generateFunction("dup", sc, "/usr/local/bin/pi", "/home/dev/repo", DefaultDialect())

	// PI_PARENT_EVAL_FILE must be inside the subshell, not prefixing it.
	// Correct: (cd "/path" && PI_PARENT_EVAL_FILE="$_pi_eval_file" /usr/local/bin/pi ...)
	// Wrong:   PI_PARENT_EVAL_FILE="$_pi_eval_file" (cd "/path" && ...)
	if strings.Contains(fn, `PI_PARENT_EVAL_FILE="$_pi_eval_file" (cd`) {
		t.Errorf("PI_PARENT_EVAL_FILE must be inside the subshell, not outside:\n%s", fn)
	}
	if !strings.Contains(fn, `&& PI_PARENT_EVAL_FILE="$_pi_eval_file"`) {
		t.Errorf("PI_PARENT_EVAL_FILE should be inside subshell as command prefix:\n%s", fn)
	}
}

func TestGenerateCompletionScript(t *testing.T) {
	content := GenerateCompletionScript("/usr/local/bin/pi")

	if !strings.Contains(content, "ZSH_VERSION") {
		t.Error("expected ZSH_VERSION check for zsh completion")
	}
	if !strings.Contains(content, "BASH_VERSION") {
		t.Error("expected BASH_VERSION check for bash completion")
	}
	if !strings.Contains(content, "/usr/local/bin/pi completion zsh") {
		t.Error("expected pi completion zsh command")
	}
	if !strings.Contains(content, "/usr/local/bin/pi completion bash") {
		t.Error("expected pi completion bash command")
	}
	if !strings.Contains(content, "do not edit manually") {
		t.Error("expected auto-generated warning")
	}
}

func TestInstall_CreatesCompletionScript(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	_, err := Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	completionPath := filepath.Join(tmpHome, piShellDir, piCompletionFile)
	data, err := os.ReadFile(completionPath)
	if err != nil {
		t.Fatalf("reading completion file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "completion zsh") {
		t.Error("completion file should contain zsh completion setup")
	}
	if !strings.Contains(content, "completion bash") {
		t.Error("completion file should contain bash completion setup")
	}
}

func TestUninstall_RemovesCompletionWhenLastProject(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)

	cfg := &config.ProjectConfig{
		Project: "only-proj",
		Shortcuts: map[string]config.Shortcut{
			"a": {Run: "auto/a"},
		},
	}

	Install(cfg, "pi", "/repo")
	completionPath := filepath.Join(tmpHome, piShellDir, piCompletionFile)
	if _, err := os.Stat(completionPath); err != nil {
		t.Fatalf("completion script should exist after install: %v", err)
	}

	Uninstall("only-proj")
	if _, err := os.Stat(completionPath); !os.IsNotExist(err) {
		t.Error("completion script should be removed when last project is uninstalled")
	}
}

func TestListInstalled_ExcludesCompletion(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, piCompletionFile), []byte("# completion"), 0o644)
	os.WriteFile(filepath.Join(shellDir, piWrapperFile), []byte("# wrapper"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "myproj.sh"), []byte("# shortcuts"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d: %v", len(projects), projects)
	}
	if projects[0] != "myproj" {
		t.Errorf("expected myproj, got %s", projects[0])
	}
}

func TestGenerateFunction_WithInputs_EvalInsideSubshell(t *testing.T) {
	sc := config.Shortcut{
		Run:  "docker/logs",
		With: map[string]string{"service": "$1"},
	}
	fn := generateFunction("dlogs", sc, "/usr/local/bin/pi", "/home/dev/repo", DefaultDialect())

	if strings.Contains(fn, `PI_PARENT_EVAL_FILE="$_pi_eval_file" (cd`) {
		t.Errorf("PI_PARENT_EVAL_FILE must be inside the subshell, not outside:\n%s", fn)
	}
	if !strings.Contains(fn, `&& PI_PARENT_EVAL_FILE="$_pi_eval_file"`) {
		t.Errorf("PI_PARENT_EVAL_FILE should be inside subshell as command prefix:\n%s", fn)
	}
}

// --- Fish shell integration tests ---

func TestInstall_CreatesFishFilesWhenFishConfigExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create .zshrc and fish config
	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)
	fishDir := filepath.Join(tmpHome, ".config", "fish")
	os.MkdirAll(fishDir, 0o755)
	os.WriteFile(filepath.Join(fishDir, "config.fish"), []byte("# fish config\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	_, err := Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify fish shortcut file
	fishPath := filepath.Join(tmpHome, piShellDir, "testproj.fish")
	data, err := os.ReadFile(fishPath)
	if err != nil {
		t.Fatalf("reading fish file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "function hello") {
		t.Error("fish file should contain hello function")
	}
	if strings.Contains(content, "function hello()") {
		t.Error("fish file should not use () syntax")
	}
	if !strings.Contains(content, "$argv") {
		t.Error("fish file should use $argv")
	}
	if !strings.Contains(content, "end\n") {
		t.Error("fish file should use 'end' to close functions")
	}

	// Verify fish wrapper
	fishWrapperPath := filepath.Join(tmpHome, piShellDir, piFishWrapperFile)
	wrapperData, err := os.ReadFile(fishWrapperPath)
	if err != nil {
		t.Fatalf("reading fish wrapper: %v", err)
	}
	if !strings.Contains(string(wrapperData), "function pi") {
		t.Error("fish wrapper should contain pi function")
	}

	// Verify fish completion
	fishCompletionPath := filepath.Join(tmpHome, piShellDir, piFishCompletionFile)
	compData, err := os.ReadFile(fishCompletionPath)
	if err != nil {
		t.Fatalf("reading fish completion: %v", err)
	}
	if !strings.Contains(string(compData), "completion fish") {
		t.Error("fish completion should reference fish completion command")
	}

	// Verify fish source line in config.fish
	fishCfgData, _ := os.ReadFile(filepath.Join(fishDir, "config.fish"))
	if !strings.Contains(string(fishCfgData), "# Added by PI") {
		t.Error("config.fish should contain source line")
	}
	if !strings.Contains(string(fishCfgData), "*.fish") {
		t.Error("config.fish source line should source .fish files")
	}
}

func TestInstall_SkipsFishWhenNoFishConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte("# existing\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	_, err := Install(cfg, "pi", "/repo")
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Fish file should NOT exist
	fishPath := filepath.Join(tmpHome, piShellDir, "testproj.fish")
	if _, err := os.Stat(fishPath); !os.IsNotExist(err) {
		t.Error("fish file should not be created when no fish config exists")
	}
}

func TestUninstall_RemovesFishFiles(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)
	fishDir := filepath.Join(tmpHome, ".config", "fish")
	os.MkdirAll(fishDir, 0o755)
	os.WriteFile(filepath.Join(fishDir, "config.fish"), []byte("# fish\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "only-proj",
		Shortcuts: map[string]config.Shortcut{
			"a": {Run: "auto/a"},
		},
	}

	Install(cfg, "pi", "/repo")

	// Verify fish files exist
	fishPath := filepath.Join(tmpHome, piShellDir, "only-proj.fish")
	if _, err := os.Stat(fishPath); err != nil {
		t.Fatalf("fish file should exist after install: %v", err)
	}

	Uninstall("only-proj")

	// Fish shortcut file should be removed
	if _, err := os.Stat(fishPath); !os.IsNotExist(err) {
		t.Error("fish file should be removed after uninstall")
	}

	// Fish wrapper should be removed (last project)
	fishWrapperPath := filepath.Join(tmpHome, piShellDir, piFishWrapperFile)
	if _, err := os.Stat(fishWrapperPath); !os.IsNotExist(err) {
		t.Error("fish wrapper should be removed when last project is uninstalled")
	}

	// Fish completion should be removed (last project)
	fishCompletionPath := filepath.Join(tmpHome, piShellDir, piFishCompletionFile)
	if _, err := os.Stat(fishCompletionPath); !os.IsNotExist(err) {
		t.Error("fish completion should be removed when last project is uninstalled")
	}

	// Fish source line should be removed
	fishCfgData, _ := os.ReadFile(filepath.Join(fishDir, "config.fish"))
	if strings.Contains(string(fishCfgData), "# Added by PI") {
		t.Error("fish source line should be removed when no repos remain")
	}
}

func TestUninstall_KeepsFishFilesIfOtherProjectsExist(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)
	fishDir := filepath.Join(tmpHome, ".config", "fish")
	os.MkdirAll(fishDir, 0o755)
	os.WriteFile(filepath.Join(fishDir, "config.fish"), []byte("# fish\n"), 0o644)

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

	Uninstall("proj1")

	// proj2's fish file should still exist
	fishPath2 := filepath.Join(tmpHome, piShellDir, "proj2.fish")
	if _, err := os.Stat(fishPath2); err != nil {
		t.Error("proj2 fish file should remain when other projects still installed")
	}

	// Fish wrapper should remain
	fishWrapperPath := filepath.Join(tmpHome, piShellDir, piFishWrapperFile)
	if _, err := os.Stat(fishWrapperPath); err != nil {
		t.Error("fish wrapper should remain when other projects still installed")
	}
}

func TestListInstalled_IncludesFishOnlyProjects(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, "proj-a.sh"), []byte("# a"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "proj-a.fish"), []byte("# a fish"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "proj-b.fish"), []byte("# b fish"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d: %v", len(projects), projects)
	}
	if projects[0] != "proj-a" || projects[1] != "proj-b" {
		t.Errorf("unexpected projects: %v", projects)
	}
}

func TestListInstalled_ExcludesFishInfraFiles(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, piFishWrapperFile), []byte("# wrapper"), 0o644)
	os.WriteFile(filepath.Join(shellDir, piFishCompletionFile), []byte("# completion"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "myproj.fish"), []byte("# shortcuts"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d: %v", len(projects), projects)
	}
	if projects[0] != "myproj" {
		t.Errorf("expected myproj, got %s", projects[0])
	}
}

func TestListInstalled_DeduplicatesShAndFish(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	shellDir := filepath.Join(tmpHome, piShellDir)
	os.MkdirAll(shellDir, 0o755)
	os.WriteFile(filepath.Join(shellDir, "myproj.sh"), []byte("# sh"), 0o644)
	os.WriteFile(filepath.Join(shellDir, "myproj.fish"), []byte("# fish"), 0o644)

	projects, err := ListInstalled()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project (deduplicated), got %d: %v", len(projects), projects)
	}
	if projects[0] != "myproj" {
		t.Errorf("expected myproj, got %s", projects[0])
	}
}

func TestFishFilePath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	path, err := FishFilePath("myproject")
	if err != nil {
		t.Fatalf("FishFilePath failed: %v", err)
	}
	if !strings.HasSuffix(path, "myproject.fish") {
		t.Errorf("expected .fish extension, got: %s", path)
	}
	if !strings.Contains(path, piShellDir) {
		t.Errorf("expected path in %s, got: %s", piShellDir, path)
	}
}

func TestGenerateFishCompletionScript(t *testing.T) {
	content := GenerateFishCompletionScript("/usr/local/bin/pi")

	if !strings.Contains(content, "/usr/local/bin/pi completion fish") {
		t.Error("expected pi completion fish command")
	}
	if !strings.Contains(content, "| source") {
		t.Error("expected pipe to source")
	}
	if !strings.Contains(content, "do not edit manually") {
		t.Error("expected auto-generated warning")
	}
}

func TestInstall_FishSourceLineIdempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	os.WriteFile(filepath.Join(tmpHome, ".zshrc"), []byte(""), 0o644)
	fishDir := filepath.Join(tmpHome, ".config", "fish")
	os.MkdirAll(fishDir, 0o755)
	os.WriteFile(filepath.Join(fishDir, "config.fish"), []byte("# fish\n"), 0o644)

	cfg := &config.ProjectConfig{
		Project: "testproj",
		Shortcuts: map[string]config.Shortcut{
			"hello": {Run: "greet"},
		},
	}

	Install(cfg, "pi", "/repo")
	Install(cfg, "pi", "/repo")

	fishCfgData, _ := os.ReadFile(filepath.Join(fishDir, "config.fish"))
	if strings.Count(string(fishCfgData), "# Added by PI") != 1 {
		t.Error("fish source line should appear only once after re-install")
	}
}

func TestHasFishConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if hasFishConfig() {
		t.Error("hasFishConfig should return false when config.fish doesn't exist")
	}

	fishDir := filepath.Join(tmpHome, ".config", "fish")
	os.MkdirAll(fishDir, 0o755)
	os.WriteFile(filepath.Join(fishDir, "config.fish"), []byte(""), 0o644)

	if !hasFishConfig() {
		t.Error("hasFishConfig should return true when config.fish exists")
	}
}

func TestIsInfraFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{piWrapperFile, true},
		{piCompletionFile, true},
		{piFishWrapperFile, true},
		{piFishCompletionFile, true},
		{"myproj.sh", false},
		{"myproj.fish", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInfraFile(tt.name); got != tt.want {
				t.Errorf("isInfraFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
