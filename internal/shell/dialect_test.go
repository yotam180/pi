package shell

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/config"
)

func TestBashDialect_EvalWrapperFunc(t *testing.T) {
	d := BashDialect{}
	result := d.EvalWrapperFunc("myfunc", `echo "hello"`)

	if !strings.Contains(result, "function myfunc()") {
		t.Errorf("expected function myfunc(), got:\n%s", result)
	}
	if !strings.Contains(result, `_pi_eval_file="$(mktemp)"`) {
		t.Error("expected mktemp call")
	}
	if !strings.Contains(result, `echo "hello"`) {
		t.Error("expected inner command")
	}
	if !strings.Contains(result, "local _pi_exit=$?") {
		t.Error("expected exit code capture")
	}
	if !strings.Contains(result, `source "$_pi_eval_file"`) {
		t.Error("expected source of eval file")
	}
	if !strings.Contains(result, `rm -f "$_pi_eval_file"`) {
		t.Error("expected cleanup")
	}
	if !strings.Contains(result, "return $_pi_exit") {
		t.Error("expected return with exit code")
	}
}

func TestBashDialect_InRepoCmd(t *testing.T) {
	d := BashDialect{}
	result := d.InRepoCmd("/home/dev/project", "pi run docker/up")

	if !strings.HasPrefix(result, `(cd "/home/dev/project"`) {
		t.Errorf("expected subshell with cd, got: %s", result)
	}
	if !strings.Contains(result, "&& pi run docker/up)") {
		t.Errorf("expected command inside subshell, got: %s", result)
	}
}

func TestBashDialect_AnywhereCmd(t *testing.T) {
	d := BashDialect{}
	result := d.AnywhereCmd("/home/dev/project", "pi", "run deploy/push")

	if !strings.Contains(result, `--repo "/home/dev/project"`) {
		t.Errorf("expected --repo flag, got: %s", result)
	}
	if !strings.Contains(result, "run deploy/push") {
		t.Errorf("expected run args, got: %s", result)
	}
}

func TestBashDialect_AllArgs(t *testing.T) {
	d := BashDialect{}
	if d.AllArgs() != `"$@"` {
		t.Errorf("expected \"$@\", got: %s", d.AllArgs())
	}
}

func TestBashDialect_PositionalArg(t *testing.T) {
	d := BashDialect{}
	if d.PositionalArg("1") != `"$1"` {
		t.Errorf("expected \"$1\", got: %s", d.PositionalArg("1"))
	}
	if d.PositionalArg("3") != `"$3"` {
		t.Errorf("expected \"$3\", got: %s", d.PositionalArg("3"))
	}
}

func TestBashDialect_FileHeader(t *testing.T) {
	d := BashDialect{}
	header := d.FileHeader("myproject", "/home/dev/myproject")

	if !strings.Contains(header, "# PI shortcuts for myproject") {
		t.Error("expected project name in header")
	}
	if !strings.Contains(header, "# Repo: /home/dev/myproject") {
		t.Error("expected repo path in header")
	}
	if !strings.Contains(header, "do not edit manually") {
		t.Error("expected auto-generated warning")
	}
}

func TestDefaultDialect_IsBash(t *testing.T) {
	d := DefaultDialect()
	_, ok := d.(BashDialect)
	if !ok {
		t.Errorf("DefaultDialect() should return BashDialect, got %T", d)
	}
}

func TestCustomDialect_PlugsIntoGeneration(t *testing.T) {
	mock := &mockDialect{}

	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"up": {Run: "docker/up"},
		},
	}

	result := GenerateShellFileWithDialect(cfg, "pi", "/repo", mock)

	if !strings.Contains(result, "[HEADER:test:/repo]") {
		t.Errorf("expected mock header, got:\n%s", result)
	}
	if !strings.Contains(result, "[WRAP:") {
		t.Errorf("expected mock wrapper, got:\n%s", result)
	}
}

func TestCustomDialect_GlobalWrapper(t *testing.T) {
	mock := &mockDialect{}
	result := GenerateGlobalWrapperWithDialect("pi", mock)

	if !strings.Contains(result, "[WRAP:pi:") {
		t.Errorf("expected mock wrapper for pi function, got:\n%s", result)
	}
}

type mockDialect struct{}

func (d *mockDialect) EvalWrapperFunc(funcName, innerCmd string) string {
	return fmt.Sprintf("[WRAP:%s:%s]\n", funcName, innerCmd)
}

func (d *mockDialect) InRepoCmd(repoRoot, piCmd string) string {
	return fmt.Sprintf("[INREPO:%s:%s]", repoRoot, piCmd)
}

func (d *mockDialect) AnywhereCmd(repoRoot, piBinary, piArgs string) string {
	return fmt.Sprintf("[ANYWHERE:%s:%s:%s]", repoRoot, piBinary, piArgs)
}

func (d *mockDialect) AllArgs() string {
	return "[ARGS]"
}

func (d *mockDialect) PositionalArg(n string) string {
	return fmt.Sprintf("[ARG:%s]", n)
}

func (d *mockDialect) FileHeader(project, repoRoot string) string {
	return fmt.Sprintf("[HEADER:%s:%s]\n", project, repoRoot)
}

func TestBashDialect_EvalWrapperFunc_ComplexInnerCmd(t *testing.T) {
	d := BashDialect{}
	inner := `(cd "/my project" && PI_PARENT_EVAL_FILE="$_pi_eval_file" pi run docker/up "$@")`
	result := d.EvalWrapperFunc("vpup", inner)

	if !strings.Contains(result, "function vpup()") {
		t.Error("expected function vpup()")
	}
	if !strings.Contains(result, inner) {
		t.Errorf("inner command should be preserved verbatim, got:\n%s", result)
	}
}

func TestBashDialect_InRepoCmd_QuotesSpaces(t *testing.T) {
	d := BashDialect{}
	result := d.InRepoCmd("/home/dev/my project", "pi run test")

	if !strings.Contains(result, `"/home/dev/my project"`) {
		t.Errorf("path with spaces should be quoted, got: %s", result)
	}
}

func TestBashDialect_AnywhereCmd_QuotesSpaces(t *testing.T) {
	d := BashDialect{}
	result := d.AnywhereCmd("/home/dev/my project", "pi", "run deploy")

	if !strings.Contains(result, `--repo "/home/dev/my project"`) {
		t.Errorf("repo path with spaces should be quoted, got: %s", result)
	}
}

func TestGenerateShellFileWithDialect_EmptyShortcuts(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project:   "empty",
		Shortcuts: map[string]config.Shortcut{},
	}

	result := GenerateShellFileWithDialect(cfg, "pi", "/repo", DefaultDialect())
	if result != "" {
		t.Errorf("expected empty output for no shortcuts, got:\n%s", result)
	}
}

func TestBuildWithArgs(t *testing.T) {
	tests := []struct {
		name     string
		with     map[string]string
		contains []string
	}{
		{
			name:     "positional ref",
			with:     map[string]string{"service": "$1"},
			contains: []string{`--with service="$1"`},
		},
		{
			name:     "literal value",
			with:     map[string]string{"count": "50"},
			contains: []string{`--with count="50"`},
		},
		{
			name:     "mixed",
			with:     map[string]string{"env": "$1", "region": "us-east-1"},
			contains: []string{`--with env="$1"`, `--with region="us-east-1"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWithArgs(tt.with, DefaultDialect())
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("expected %q in result %q", want, result)
				}
			}
		})
	}
}

func TestBuildWithArgs_Sorted(t *testing.T) {
	with := map[string]string{"z": "3", "a": "1", "m": "2"}
	result := buildWithArgs(with, DefaultDialect())

	aIdx := strings.Index(result, "--with a=")
	mIdx := strings.Index(result, "--with m=")
	zIdx := strings.Index(result, "--with z=")

	if aIdx > mIdx || mIdx > zIdx {
		t.Errorf("with args should be sorted alphabetically, got: %s", result)
	}
}

func TestBuildWithArgs_FishDialect(t *testing.T) {
	tests := []struct {
		name     string
		with     map[string]string
		contains []string
	}{
		{
			name:     "positional ref uses argv",
			with:     map[string]string{"service": "$1"},
			contains: []string{`--with service=$argv[1]`},
		},
		{
			name:     "literal value unchanged",
			with:     map[string]string{"count": "50"},
			contains: []string{`--with count="50"`},
		},
		{
			name:     "mixed",
			with:     map[string]string{"env": "$1", "region": "us-east-1"},
			contains: []string{`--with env=$argv[1]`, `--with region="us-east-1"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWithArgs(tt.with, FishDialect{})
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("expected %q in result %q", want, result)
				}
			}
		})
	}
}

// --- FishDialect tests ---

func TestFishDialect_EvalWrapperFunc(t *testing.T) {
	d := FishDialect{}
	result := d.EvalWrapperFunc("myfunc", `echo "hello"`)

	if !strings.Contains(result, "function myfunc") {
		t.Errorf("expected function myfunc, got:\n%s", result)
	}
	if strings.Contains(result, "function myfunc()") {
		t.Error("fish functions should not have () after name")
	}
	if !strings.Contains(result, "_pi_eval_file (mktemp)") {
		t.Error("expected mktemp call with fish syntax")
	}
	if !strings.Contains(result, `echo "hello"`) {
		t.Error("expected inner command")
	}
	if !strings.Contains(result, "set -l _pi_exit $status") {
		t.Error("expected exit code capture with fish syntax")
	}
	if !strings.Contains(result, "eval (cat $_pi_eval_file)") {
		t.Error("expected fish eval of eval file")
	}
	if !strings.Contains(result, "rm -f $_pi_eval_file") {
		t.Error("expected cleanup")
	}
	if !strings.Contains(result, "return $_pi_exit") {
		t.Error("expected return with exit code")
	}
	if !strings.Contains(result, "end\n") {
		t.Error("expected 'end' to close function")
	}
}

func TestFishDialect_InRepoCmd(t *testing.T) {
	d := FishDialect{}
	result := d.InRepoCmd("/home/dev/project", "pi run docker/up")

	if !strings.Contains(result, `env -C "/home/dev/project"`) {
		t.Errorf("expected env -C with quoted path, got: %s", result)
	}
	if !strings.Contains(result, "pi run docker/up") {
		t.Errorf("expected command, got: %s", result)
	}
}

func TestFishDialect_AnywhereCmd(t *testing.T) {
	d := FishDialect{}
	result := d.AnywhereCmd("/home/dev/project", "pi", "run deploy/push")

	if !strings.Contains(result, `--repo "/home/dev/project"`) {
		t.Errorf("expected --repo flag, got: %s", result)
	}
	if !strings.Contains(result, "run deploy/push") {
		t.Errorf("expected run args, got: %s", result)
	}
}

func TestFishDialect_AllArgs(t *testing.T) {
	d := FishDialect{}
	if d.AllArgs() != "$argv" {
		t.Errorf("expected $argv, got: %s", d.AllArgs())
	}
}

func TestFishDialect_PositionalArg(t *testing.T) {
	d := FishDialect{}
	if d.PositionalArg("1") != "$argv[1]" {
		t.Errorf("expected $argv[1], got: %s", d.PositionalArg("1"))
	}
	if d.PositionalArg("3") != "$argv[3]" {
		t.Errorf("expected $argv[3], got: %s", d.PositionalArg("3"))
	}
}

func TestFishDialect_FileHeader(t *testing.T) {
	d := FishDialect{}
	header := d.FileHeader("myproject", "/home/dev/myproject")

	if !strings.Contains(header, "# PI shortcuts for myproject") {
		t.Error("expected project name in header")
	}
	if !strings.Contains(header, "# Repo: /home/dev/myproject") {
		t.Error("expected repo path in header")
	}
	if !strings.Contains(header, "do not edit manually") {
		t.Error("expected auto-generated warning")
	}
}

func TestFishDialect_InRepoCmd_QuotesSpaces(t *testing.T) {
	d := FishDialect{}
	result := d.InRepoCmd("/home/dev/my project", "pi run test")

	if !strings.Contains(result, `"/home/dev/my project"`) {
		t.Errorf("path with spaces should be quoted, got: %s", result)
	}
}

func TestFishDialect_AnywhereCmd_QuotesSpaces(t *testing.T) {
	d := FishDialect{}
	result := d.AnywhereCmd("/home/dev/my project", "pi", "run deploy")

	if !strings.Contains(result, `--repo "/home/dev/my project"`) {
		t.Errorf("repo path with spaces should be quoted, got: %s", result)
	}
}

func TestFishDialect_EvalWrapperFunc_ComplexInnerCmd(t *testing.T) {
	d := FishDialect{}
	inner := `env -C "/my project" PI_PARENT_EVAL_FILE=$_pi_eval_file pi run docker/up $argv`
	result := d.EvalWrapperFunc("vpup", inner)

	if !strings.Contains(result, "function vpup") {
		t.Error("expected function vpup")
	}
	if !strings.Contains(result, inner) {
		t.Errorf("inner command should be preserved verbatim, got:\n%s", result)
	}
}

func TestFishDialect_PlugsIntoGeneration(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"up": {Run: "docker/up"},
		},
	}

	result := GenerateShellFileWithDialect(cfg, "pi", "/repo", FishDialect{})

	if !strings.Contains(result, "function up") {
		t.Errorf("expected fish function up, got:\n%s", result)
	}
	if strings.Contains(result, "function up()") {
		t.Error("fish functions should not have () syntax")
	}
	if !strings.Contains(result, "$argv") {
		t.Errorf("expected $argv for arg forwarding, got:\n%s", result)
	}
	if !strings.Contains(result, "end\n") {
		t.Errorf("expected 'end' to close function, got:\n%s", result)
	}
}

func TestFishDialect_GlobalWrapper(t *testing.T) {
	result := GenerateGlobalWrapperWithDialect("pi", FishDialect{})

	if !strings.Contains(result, "function pi") {
		t.Error("expected function pi")
	}
	if strings.Contains(result, "function pi()") {
		t.Error("fish wrapper should not have () syntax")
	}
	if !strings.Contains(result, "command pi") {
		t.Error("expected 'command' keyword to avoid recursion")
	}
	if !strings.Contains(result, "$argv") {
		t.Error("expected $argv in fish wrapper")
	}
}
