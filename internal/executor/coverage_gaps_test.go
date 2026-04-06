package executor

import (
	"bytes"
	"errors"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/display"
)

// --- interpolateWith (0% → 100%) ---

func TestInterpolateWith_Empty(t *testing.T) {
	e := &Executor{}
	got := e.interpolateWith(nil)
	if got != nil {
		t.Errorf("interpolateWith(nil) = %v, want nil", got)
	}
}

func TestInterpolateWith_EmptyMap(t *testing.T) {
	e := &Executor{}
	got := e.interpolateWith(map[string]string{})
	if len(got) != 0 {
		t.Errorf("interpolateWith(empty) = %v, want empty", got)
	}
}

func TestInterpolateWith_OutputsLast(t *testing.T) {
	e := &Executor{stepOutputs: []string{"hello", "world"}}
	got := e.interpolateWith(map[string]string{"key": "outputs.last"})
	if got["key"] != "world" {
		t.Errorf("interpolateWith outputs.last = %q, want %q", got["key"], "world")
	}
}

func TestInterpolateWith_Literal(t *testing.T) {
	e := &Executor{}
	got := e.interpolateWith(map[string]string{"key": "literal-value"})
	if got["key"] != "literal-value" {
		t.Errorf("interpolateWith literal = %q, want %q", got["key"], "literal-value")
	}
}

func TestInterpolateWith_MultipleKeys(t *testing.T) {
	e := &Executor{stepOutputs: []string{"v1", "v2"}}
	got := e.interpolateWith(map[string]string{
		"a": "outputs.0",
		"b": "outputs.last",
		"c": "plain",
	})
	if got["a"] != "v1" {
		t.Errorf("key a = %q, want %q", got["a"], "v1")
	}
	if got["b"] != "v2" {
		t.Errorf("key b = %q, want %q", got["b"], "v2")
	}
	if got["c"] != "plain" {
		t.Errorf("key c = %q, want %q", got["c"], "plain")
	}
}

// --- stdout() / stderr() nil fallback (66.7% → 100%) ---

func TestStdout_NilFallback(t *testing.T) {
	e := &Executor{}
	got := e.stdout()
	if got != os.Stdout {
		t.Errorf("stdout() with nil = %v, want os.Stdout", got)
	}
}

func TestStdout_SetExplicitly(t *testing.T) {
	buf := &bytes.Buffer{}
	e := &Executor{Stdout: buf}
	got := e.stdout()
	if got != buf {
		t.Error("stdout() with explicit writer should return that writer")
	}
}

func TestStderr_NilFallback(t *testing.T) {
	e := &Executor{}
	got := e.stderr()
	if got != os.Stderr {
		t.Errorf("stderr() with nil = %v, want os.Stderr", got)
	}
}

func TestStderr_SetExplicitly(t *testing.T) {
	buf := &bytes.Buffer{}
	e := &Executor{Stderr: buf}
	got := e.stderr()
	if got != buf {
		t.Error("stderr() with explicit writer should return that writer")
	}
}

func TestStdin_NilFallback(t *testing.T) {
	e := &Executor{}
	got := e.stdin()
	if got != os.Stdin {
		t.Errorf("stdin() with nil = %v, want os.Stdin", got)
	}
}

// --- printer() lazy creation (66.7% → 100%) ---

func TestPrinter_ExplicitPrinter(t *testing.T) {
	p := display.New(io.Discard)
	e := &Executor{Printer: p}
	got := e.printer()
	if got != p {
		t.Error("printer() with explicit Printer should return it")
	}
}

func TestPrinter_LazyCreation(t *testing.T) {
	buf := &bytes.Buffer{}
	e := &Executor{Stderr: buf}
	got := e.printer()
	if got == nil {
		t.Fatal("printer() should not return nil")
	}
	got.StepTrace("bash", "echo test")
	if !strings.Contains(buf.String(), "bash") {
		t.Errorf("lazy printer should write to stderr, got: %q", buf.String())
	}
}

// --- evaluateCondition parse error branch (76.9% → 100%) ---

func TestEvaluateCondition_ParseError(t *testing.T) {
	e := &Executor{RepoRoot: t.TempDir()}
	_, err := e.evaluateCondition("invalid(")
	if err == nil {
		t.Fatal("expected error for invalid expression")
	}
}

func TestEvaluateCondition_UnknownPredicate(t *testing.T) {
	e := &Executor{RepoRoot: t.TempDir()}
	_, err := e.evaluateCondition("unknown.predicate")
	if err == nil {
		t.Fatal("expected error for unknown predicate")
	}
	if !strings.Contains(err.Error(), "unknown predicate") {
		t.Errorf("error should mention 'unknown predicate', got: %v", err)
	}
}

func TestEvaluateCondition_AutomationLevelConditionParseError(t *testing.T) {
	dir := t.TempDir()
	a := newAutomationWithIf("test", "invalid(", bashStep("echo hello"))
	e := newExecutor(dir, newDiscovery(nil))

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for invalid automation-level if expression")
	}
	if !strings.Contains(err.Error(), "if:") {
		t.Errorf("error should reference 'if:', got: %v", err)
	}
}

func TestEvaluateCondition_StepLevelConditionParseError(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", bashStepIf("echo hello", "invalid("))
	e := newExecutor(dir, newDiscovery(nil))

	err := e.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for invalid step-level if expression")
	}
	if !strings.Contains(err.Error(), "step[0] if:") {
		t.Errorf("error should reference 'step[0] if:', got: %v", err)
	}
}

// --- AppendToParentEval error paths (71.4% → 100%) ---

func TestAppendToParentEval_InvalidPath(t *testing.T) {
	err := AppendToParentEval("/nonexistent/deeply/nested/path/eval.sh", "command")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
	if !strings.Contains(err.Error(), "opening parent eval file") {
		t.Errorf("error should reference opening, got: %v", err)
	}
}

func TestAppendToParentEval_ReadOnlyFile(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")
	os.WriteFile(evalFile, []byte("existing"), 0o444)
	defer os.Chmod(evalFile, 0o644)

	if os.Getuid() == 0 {
		t.Skip("root can write to read-only files")
	}

	err := AppendToParentEval(evalFile, "command")
	if err != nil {
		if !strings.Contains(err.Error(), "opening parent eval file") {
			t.Errorf("error should reference opening, got: %v", err)
		}
	}
}

// --- registry() custom Runners path (80% → 100%) ---

func TestRegistry_CustomRunners(t *testing.T) {
	reg := NewRegistry()
	reg.Register(automation.StepTypeBash, NewBashRunner())

	e := &Executor{Runners: reg}
	got := e.registry()
	if got != reg {
		t.Error("registry() should return custom Runners when set")
	}
}

func TestRegistry_CachedDefaultRegistry(t *testing.T) {
	e := &Executor{}
	first := e.registry()
	second := e.registry()
	if first != second {
		t.Error("registry() should return the same cached default on subsequent calls")
	}
	if e.cachedRegistry == nil {
		t.Error("cachedRegistry should be set after first call")
	}
}

func TestRegistry_CustomTakesPrecedenceOverCached(t *testing.T) {
	e := &Executor{}
	_ = e.registry()
	if e.cachedRegistry == nil {
		t.Fatal("cachedRegistry should be set")
	}

	custom := NewRegistry()
	e.Runners = custom
	got := e.registry()
	if got != custom {
		t.Error("registry() should return Runners when set, even if cachedRegistry exists")
	}
}

// --- execFirstBlock no-match + capturePipe (81% → improved) ---

func TestFirstBlock_NoneMatch_PipeCapture_EmptyBuffer(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStepPiped(
			bashStepIf("echo wrong", "os.linux"),
		),
		bashStep("echo after"),
	)
	exec, stdout, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "after" {
		t.Errorf("output = %q, want %q", got, "after")
	}
}

func TestFirstBlock_ParentShellInSubStep(t *testing.T) {
	dir := t.TempDir()
	evalFile := filepath.Join(dir, "eval.sh")

	parentSub := automation.Step{
		Type:        automation.StepTypeBash,
		Value:       "export FOO=bar",
		ParentShell: true,
	}
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo wrong", "os.linux"),
			parentSub,
		),
	)
	exec, _, _ := newExecutorWithEnv(dir, newDiscovery(nil), fakeRuntimeEnv("darwin"))
	exec.ParentEvalFile = evalFile

	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(evalFile)
	if err != nil {
		t.Fatalf("reading eval file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "export FOO=bar" {
		t.Errorf("eval file = %q, want %q", string(data), "export FOO=bar")
	}
}

func TestFirstBlock_ConditionErrorInSubStep(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		firstStep(
			bashStepIf("echo hello", "invalid("),
		),
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for invalid condition in first: sub-step")
	}
	if !strings.Contains(err.Error(), "first[0] if:") {
		t.Errorf("error should reference 'first[0] if:', got: %v", err)
	}
}

// --- resolveScriptPath absolute path branch (66.7% → 100%) ---

func TestResolveScriptPath_AbsolutePath(t *testing.T) {
	absPath := "/usr/local/bin/script.sh"
	got := resolveScriptPath("/some/automation/dir", absPath)
	if got != absPath {
		t.Errorf("resolveScriptPath(%q) = %q, want %q", absPath, got, absPath)
	}
}

func TestResolveScriptPath_RelativePath(t *testing.T) {
	got := resolveScriptPath("/some/automation/dir", "script.sh")
	want := "/some/automation/dir/script.sh"
	if got != want {
		t.Errorf("resolveScriptPath(%q) = %q, want %q", "script.sh", got, want)
	}
}

// --- Unimplemented step type error ---

func TestExecStep_UnimplementedStepType(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		automation.Step{Type: "unknown_type", Value: "something"},
	)
	exec := newExecutor(dir, newDiscovery(nil))

	err := exec.Run(a, nil)
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented', got: %v", err)
	}
}

// --- resolveStepDir branches ---

func TestResolveStepDir_Empty(t *testing.T) {
	got, err := resolveStepDir("/repo/root", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/repo/root" {
		t.Errorf("resolveStepDir(%q) = %q, want %q", "", got, "/repo/root")
	}
}

func TestResolveStepDir_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveStepDir("/repo/root", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("resolveStepDir(%q) = %q, want %q", dir, got, dir)
	}
}

func TestResolveStepDir_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	os.WriteFile(filePath, []byte("hi"), 0o644)

	_, err := resolveStepDir(dir, "file.txt")
	if err == nil {
		t.Fatal("expected error for non-directory")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error should mention 'not a directory', got: %v", err)
	}
}

// --- isCommandNotFound ---

func TestIsCommandNotFound_True(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"executable file not found in $PATH", true},
		{"no such file or directory", true},
		{"permission denied", false},
		{"exit status 1", false},
	}
	for _, tt := range tests {
		err := errors.New(tt.msg)
		got := isCommandNotFound(err)
		if got != tt.want {
			t.Errorf("isCommandNotFound(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}

// --- WriteTempScript ---

func TestWriteTempScript_Success(t *testing.T) {
	path, cleanup, err := writeTempScript("echo hello", ".sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading temp script: %v", err)
	}
	if string(data) != "echo hello" {
		t.Errorf("content = %q, want %q", string(data), "echo hello")
	}
	if !strings.HasSuffix(path, ".sh") {
		t.Errorf("path should end in .sh, got %q", path)
	}
}

// --- resolvePythonBin ---

func TestResolvePythonBin_NoVenv(t *testing.T) {
	old := os.Getenv("VIRTUAL_ENV")
	os.Unsetenv("VIRTUAL_ENV")
	defer func() {
		if old != "" {
			os.Setenv("VIRTUAL_ENV", old)
		}
	}()

	got := resolvePythonBin()
	if got != "python3" {
		t.Errorf("resolvePythonBin() = %q, want %q", got, "python3")
	}
}

func TestResolvePythonBin_WithVenv(t *testing.T) {
	old := os.Getenv("VIRTUAL_ENV")
	os.Setenv("VIRTUAL_ENV", "/tmp/myvenv")
	defer func() {
		if old != "" {
			os.Setenv("VIRTUAL_ENV", old)
		} else {
			os.Unsetenv("VIRTUAL_ENV")
		}
	}()

	got := resolvePythonBin()
	if got != "/tmp/myvenv/bin/python" {
		t.Errorf("resolvePythonBin() = %q, want %q", got, "/tmp/myvenv/bin/python")
	}
}

// --- InstallHintFor exported wrapper ---

func TestInstallHintFor_Known(t *testing.T) {
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	got := InstallHintFor(req)
	if got == "" {
		t.Error("InstallHintFor(python) should return a hint")
	}
}

func TestInstallHintFor_Unknown(t *testing.T) {
	req := automation.Requirement{Name: "unknown-tool-xyz", Kind: automation.RequirementCommand}
	got := InstallHintFor(req)
	if got != "" {
		t.Errorf("InstallHintFor(unknown) = %q, want empty", got)
	}
}

// --- ValidationError formatting ---

func TestFormatValidationError_WithHint(t *testing.T) {
	ve := &ValidationError{
		AutomationName: "build",
		Results: []CheckResult{
			{
				Requirement: automation.Requirement{Name: "python", Kind: automation.RequirementRuntime},
				Satisfied:   false,
				Error:       "not found",
			},
		},
	}
	got := FormatValidationError(ve)
	if !strings.Contains(got, "python") {
		t.Errorf("should contain requirement name, got: %q", got)
	}
	if !strings.Contains(got, "install:") {
		t.Errorf("should contain install hint, got: %q", got)
	}
}

func TestFormatValidationError_WithoutHint(t *testing.T) {
	ve := &ValidationError{
		AutomationName: "build",
		Results: []CheckResult{
			{
				Requirement: automation.Requirement{Name: "obscure-tool", Kind: automation.RequirementCommand},
				Satisfied:   false,
				Error:       "not found",
			},
		},
	}
	got := FormatValidationError(ve)
	if strings.Contains(got, "install:") {
		t.Errorf("should not contain install hint for unknown tool, got: %q", got)
	}
}

func TestFormatValidationError_SkipsSatisfied(t *testing.T) {
	ve := &ValidationError{
		AutomationName: "build",
		Results: []CheckResult{
			{
				Requirement: automation.Requirement{Name: "bash", Kind: automation.RequirementCommand},
				Satisfied:   true,
			},
			{
				Requirement: automation.Requirement{Name: "python", Kind: automation.RequirementRuntime},
				Satisfied:   false,
				Error:       "not found",
			},
		},
	}
	got := FormatValidationError(ve)
	if strings.Contains(got, "bash") {
		t.Errorf("should not include satisfied requirements, got: %q", got)
	}
}

// --- formatRequirementLabel branches ---

func TestFormatRequirementLabel_WithVersion(t *testing.T) {
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"}
	got := formatRequirementLabel(req)
	if got != "python >= 3.11" {
		t.Errorf("formatRequirementLabel = %q, want %q", got, "python >= 3.11")
	}
}

func TestFormatRequirementLabel_CommandNoVersion(t *testing.T) {
	req := automation.Requirement{Name: "jq", Kind: automation.RequirementCommand}
	got := formatRequirementLabel(req)
	if got != "command: jq" {
		t.Errorf("formatRequirementLabel = %q, want %q", got, "command: jq")
	}
}

func TestFormatRequirementLabel_RuntimeNoVersion(t *testing.T) {
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	got := formatRequirementLabel(req)
	if got != "python" {
		t.Errorf("formatRequirementLabel = %q, want %q", got, "python")
	}
}

// --- runtimeCommand branches ---

func TestRuntimeCommand_Python(t *testing.T) {
	got := runtimeCommand("python")
	if got != "python3" {
		t.Errorf("runtimeCommand(python) = %q, want %q", got, "python3")
	}
}

func TestRuntimeCommand_Rust(t *testing.T) {
	got := runtimeCommand("rust")
	if got != "rustc" {
		t.Errorf("runtimeCommand(rust) = %q, want %q", got, "rustc")
	}
}

func TestRuntimeCommand_Default(t *testing.T) {
	got := runtimeCommand("node")
	if got != "node" {
		t.Errorf("runtimeCommand(node) = %q, want %q", got, "node")
	}
}

// --- detectVersion with mocked env ---

func TestDetectVersion_MockedEnv_VersionFlag(t *testing.T) {
	env := &RuntimeEnv{
		GOOS:   "darwin",
		GOARCH: "arm64",
		Getenv: func(s string) string { return "" },
		LookPath: func(s string) (string, error) {
			if s == "mytool" {
				return "/usr/bin/mytool", nil
			}
			return "", osexec.ErrNotFound
		},
		Stat: os.Stat,
		ExecOutput: func(cmd string, args ...string) string {
			if cmd == "mytool" && len(args) > 0 && args[0] == "--version" {
				return "mytool version 2.5.1"
			}
			return ""
		},
	}
	got := detectVersion("mytool", env)
	if got != "2.5.1" {
		t.Errorf("detectVersion = %q, want %q", got, "2.5.1")
	}
}

func TestDetectVersion_MockedEnv_FallbackToVersionSubcommand(t *testing.T) {
	env := &RuntimeEnv{
		GOOS:   "darwin",
		GOARCH: "arm64",
		Getenv: func(s string) string { return "" },
		LookPath: func(s string) (string, error) {
			return "/usr/bin/" + s, nil
		},
		Stat: os.Stat,
		ExecOutput: func(cmd string, args ...string) string {
			if args[0] == "--version" {
				return ""
			}
			if args[0] == "version" {
				return "go version go1.23.0 darwin/arm64"
			}
			return ""
		},
	}
	got := detectVersion("go", env)
	if got != "1.23.0" {
		t.Errorf("detectVersion = %q, want %q", got, "1.23.0")
	}
}

// --- extractVersion patterns ---

func TestExtractVersion_Various(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Python 3.13.0", "3.13.0"},
		{"v20.11.0", "20.11.0"},
		{"jq-1.7.1", "1.7.1"},
		{"node v22.0.0", "22.0.0"},
		{"go version go1.23.0 darwin/arm64", "1.23.0"},
		{"no version here", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := extractVersion(tt.input)
		if got != tt.want {
			t.Errorf("extractVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- compareVersions ---

func TestCompareVersions_Various(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"3.13.0", "3.11.0", 1},
		{"3.9.7", "3.11.0", -1},
		{"1.0.0", "1.0.0", 0},
		{"22.0", "22.0.0", 0},
		{"2.0", "1.0", 1},
	}
	for _, tt := range tests {
		got, err := compareVersions(tt.a, tt.b)
		if err != nil {
			t.Fatalf("compareVersions(%q, %q) error: %v", tt.a, tt.b, err)
		}
		if got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCompareVersions_ParseError(t *testing.T) {
	_, err := compareVersions("abc", "1.0")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}
	_, err = compareVersions("1.0", "abc")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}
}

// --- BuildStepEnv ---

func TestBuildStepEnv_AllEmpty(t *testing.T) {
	got := BuildStepEnv(nil, nil, nil, nil)
	if got != nil {
		t.Errorf("BuildStepEnv all empty = %v, want nil", got)
	}
}

func TestBuildStepEnv_RuntimePathsOnly(t *testing.T) {
	got := BuildStepEnv([]string{"/my/bin"}, nil, nil, nil)
	if got == nil {
		t.Fatal("expected non-nil env with runtime paths")
	}
	found := false
	for _, entry := range got {
		if strings.HasPrefix(entry, "PATH=") && strings.Contains(entry, "/my/bin") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected PATH to contain /my/bin")
	}
}

// --- stepExecCtx construction and use ---

func TestStepExecCtx_Roundtrip(t *testing.T) {
	a := newAutomation("test", bashStep("echo hi"))
	ctx := &stepExecCtx{
		automation: a,
		args:       []string{"arg1"},
		inputEnv:   []string{"PI_IN_X=1"},
	}
	if ctx.automation.Name != "test" {
		t.Errorf("ctx.automation.Name = %q, want %q", ctx.automation.Name, "test")
	}
	if len(ctx.args) != 1 || ctx.args[0] != "arg1" {
		t.Errorf("ctx.args = %v, want [arg1]", ctx.args)
	}
	if len(ctx.inputEnv) != 1 || ctx.inputEnv[0] != "PI_IN_X=1" {
		t.Errorf("ctx.inputEnv = %v, want [PI_IN_X=1]", ctx.inputEnv)
	}
}
