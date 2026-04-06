package executor

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
)

func TestDryRun_BasicBashStep(t *testing.T) {
	a := newAutomation("test", automation.Step{Type: automation.StepTypeBash, Value: "echo hello"})
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "bash") {
		t.Errorf("expected 'bash' in output, got: %q", got)
	}
	if !strings.Contains(got, "echo hello") {
		t.Errorf("expected 'echo hello' in output, got: %q", got)
	}
}

func TestDryRun_DoesNotExecuteCommands(t *testing.T) {
	a := newAutomation("test", automation.Step{Type: automation.StepTypeBash, Value: "exit 1"})
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("dry-run should not execute commands and should not fail: %v", err)
	}
}

func TestDryRun_ConditionalStepSkipped(t *testing.T) {
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo never", If: "os.windows"},
	)
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:   t.TempDir(),
		Discovery:  discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:     &stderr,
		DryRun:     true,
		RuntimeEnv: fakeRuntimeEnv("darwin"),
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "skipped") {
		t.Errorf("expected 'skipped' in output for false condition, got: %q", got)
	}
	if !strings.Contains(got, "os.windows") {
		t.Errorf("expected condition text in output, got: %q", got)
	}
}

func TestDryRun_ConditionalStepExecuted(t *testing.T) {
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo yes", If: "os.macos"},
	)
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:   t.TempDir(),
		Discovery:  discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:     &stderr,
		DryRun:     true,
		RuntimeEnv: fakeRuntimeEnv("darwin"),
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if strings.Contains(got, "skipped") {
		t.Errorf("should not be skipped on macOS, got: %q", got)
	}
	if !strings.Contains(got, "echo yes") {
		t.Errorf("expected step value in output, got: %q", got)
	}
}

func TestDryRun_RunStepRecurses(t *testing.T) {
	target := newAutomation("target", automation.Step{Type: automation.StepTypeBash, Value: "echo inner"})
	caller := newAutomation("caller", automation.Step{Type: automation.StepTypeRun, Value: "target"})

	all := map[string]*automation.Automation{"caller": caller, "target": target}
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(all, []string{"caller", "target"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(caller, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "run") {
		t.Errorf("expected 'run' step type in output, got: %q", got)
	}
	if !strings.Contains(got, "echo inner") {
		t.Errorf("expected target step in output (recursion), got: %q", got)
	}
}

func TestDryRun_MultipleSteps(t *testing.T) {
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo step1"},
		automation.Step{Type: automation.StepTypeBash, Value: "echo step2"},
		automation.Step{Type: automation.StepTypeBash, Value: "echo step3"},
	)
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "step1") || !strings.Contains(got, "step2") || !strings.Contains(got, "step3") {
		t.Errorf("expected all three steps in output, got: %q", got)
	}
}

func TestDryRun_FirstBlock(t *testing.T) {
	a := newAutomation("test", automation.Step{
		First: []automation.Step{
			{Type: automation.StepTypeBash, Value: "echo windows", If: "os.windows"},
			{Type: automation.StepTypeBash, Value: "echo macos", If: "os.macos"},
			{Type: automation.StepTypeBash, Value: "echo fallback"},
		},
	})
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:   t.TempDir(),
		Discovery:  discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:     &stderr,
		DryRun:     true,
		RuntimeEnv: fakeRuntimeEnv("darwin"),
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "first") {
		t.Errorf("expected 'first' block indicator, got: %q", got)
	}
	if !strings.Contains(got, "match") {
		t.Errorf("expected match indicator for macos branch, got: %q", got)
	}
	if !strings.Contains(got, "skipped") {
		t.Errorf("expected 'skipped' for windows branch, got: %q", got)
	}
	if !strings.Contains(got, "not reached") {
		t.Errorf("expected 'not reached' for fallback, got: %q", got)
	}
}

func TestDryRun_Installer(t *testing.T) {
	a := &automation.Automation{
		Name: "install-foo",
		Install: &automation.InstallSpec{
			Test:    automation.InstallPhase{IsScalar: true, Scalar: "command -v foo"},
			Run:     automation.InstallPhase{IsScalar: true, Scalar: "curl install.foo.sh | sh"},
			Version: "foo --version",
		},
	}
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"install-foo": a}, []string{"install-foo"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "install") {
		t.Errorf("expected 'install' in output, got: %q", got)
	}
	if !strings.Contains(got, "test") {
		t.Errorf("expected 'test' phase in output, got: %q", got)
	}
	if !strings.Contains(got, "run") {
		t.Errorf("expected 'run' phase in output, got: %q", got)
	}
	if !strings.Contains(got, "verify") {
		t.Errorf("expected 'verify' phase in output, got: %q", got)
	}
	if !strings.Contains(got, "version") {
		t.Errorf("expected 'version' in output, got: %q", got)
	}
}

func TestDryRun_StepAnnotations(t *testing.T) {
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "echo loud", Silent: true},
		automation.Step{Type: automation.StepTypeBash, Value: "echo piped", Pipe: true},
		automation.Step{Type: automation.StepTypeBash, Value: "echo received"},
	)
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "silent") {
		t.Errorf("expected 'silent' annotation, got: %q", got)
	}
	if !strings.Contains(got, "pipe") {
		t.Errorf("expected 'pipe' annotation, got: %q", got)
	}
}

func TestDryRun_GoFunc(t *testing.T) {
	a := &automation.Automation{
		Name:        "go-func-test",
		Description: "test go func",
		GoFunc: func(inputs map[string]string) error {
			return nil
		},
	}
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"go-func-test": a}, []string{"go-func-test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "go-func") {
		t.Errorf("expected 'go-func' in output, got: %q", got)
	}
}

func TestDryRun_AutomationLevelIf(t *testing.T) {
	a := &automation.Automation{
		Name: "conditional",
		If:   "os.windows",
		Steps: []automation.Step{
			{Type: automation.StepTypeBash, Value: "echo never"},
		},
	}
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:   t.TempDir(),
		Discovery:  discovery.NewResult(map[string]*automation.Automation{"conditional": a}, []string{"conditional"}),
		Stderr:     &stderr,
		DryRun:     true,
		RuntimeEnv: fakeRuntimeEnv("darwin"),
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if strings.Contains(got, "echo never") {
		t.Errorf("skipped automation should not show steps, got: %q", got)
	}
}

func TestDryRun_CircularDependencyHandled(t *testing.T) {
	a := newAutomation("self", automation.Step{Type: automation.StepTypeRun, Value: "self"})
	all := map[string]*automation.Automation{"self": a}
	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(all, []string{"self"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	err := e.Run(a, nil)
	if err != nil {
		t.Fatalf("dry-run should handle circular deps gracefully, got: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "circular") {
		t.Errorf("expected circular dependency mention, got: %q", got)
	}
}

func TestDryRun_DirAndTimeoutAnnotations(t *testing.T) {
	a := newAutomation("test",
		automation.Step{Type: automation.StepTypeBash, Value: "make build", Dir: "services/api", TimeoutRaw: "30s"},
	)
	// need to give it a positive timeout duration too
	a.Steps[0].Timeout = 30 * 1e9

	var stderr strings.Builder
	e := &Executor{
		RepoRoot:  t.TempDir(),
		Discovery: discovery.NewResult(map[string]*automation.Automation{"test": a}, []string{"test"}),
		Stderr:    &stderr,
		DryRun:    true,
	}

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "dir: services/api") {
		t.Errorf("expected dir annotation, got: %q", got)
	}
	if !strings.Contains(got, "timeout: 30s") {
		t.Errorf("expected timeout annotation, got: %q", got)
	}
}

func TestDryRunTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 80, "short"},
		{"line1\nline2", 80, "line1..."},
		{strings.Repeat("x", 100), 80, strings.Repeat("x", 77) + "..."},
	}
	for _, tt := range tests {
		got := dryRunTruncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("dryRunTruncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
