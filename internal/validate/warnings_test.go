package validate

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// --- WarnCheckFunc tests ---

func TestWarnCheckFunc_Name(t *testing.T) {
	c := WarnCheckFunc{CheckName: "my-warn"}
	if c.Name() != "my-warn" {
		t.Errorf("Name() = %q, want %q", c.Name(), "my-warn")
	}
}

func TestWarnCheckFunc_Run(t *testing.T) {
	called := false
	c := WarnCheckFunc{
		CheckName: "test-warn",
		Fn: func(ctx *Context) []string {
			called = true
			return []string{"warning!"}
		},
	}

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	warns := c.Run(&Context{Root: "/tmp", Config: cfg, Discovery: disc})
	if !called {
		t.Error("Fn was not called")
	}
	if len(warns) != 1 || warns[0] != "warning!" {
		t.Errorf("unexpected warnings: %v", warns)
	}
}

// --- Runner warning integration ---

func TestRunner_WarningsNotIncludedByDefault(t *testing.T) {
	r := NewRunner()
	r.RegisterWarn(WarnCheckFunc{CheckName: "test-warn", Fn: func(ctx *Context) []string {
		return []string{"a warning"}
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	ctx := &Context{Root: "/tmp", Config: cfg, Discovery: disc}

	result := r.Run(ctx)
	if len(result.Warnings) != 0 {
		t.Errorf("Run() should not include warnings, got: %v", result.Warnings)
	}
}

func TestRunner_WarningsIncludedWithOpts(t *testing.T) {
	r := NewRunner()
	r.RegisterWarn(WarnCheckFunc{CheckName: "test-warn", Fn: func(ctx *Context) []string {
		return []string{"a warning"}
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	ctx := &Context{Root: "/tmp", Config: cfg, Discovery: disc}

	result := r.RunWithOpts(ctx, true)
	if len(result.Warnings) != 1 {
		t.Fatalf("RunWithOpts(true) should include 1 warning, got %d", len(result.Warnings))
	}
	if result.Warnings[0] != "a warning" {
		t.Errorf("unexpected warning: %q", result.Warnings[0])
	}
}

func TestRunner_MultipleWarnChecksAggregate(t *testing.T) {
	r := NewRunner()
	r.RegisterWarn(WarnCheckFunc{CheckName: "warn-a", Fn: func(ctx *Context) []string {
		return []string{"warning A"}
	}})
	r.RegisterWarn(WarnCheckFunc{CheckName: "warn-b", Fn: func(ctx *Context) []string {
		return []string{"warning B1", "warning B2"}
	}})
	r.RegisterWarn(WarnCheckFunc{CheckName: "warn-c", Fn: func(ctx *Context) []string {
		return nil
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	result := r.RunWithOpts(&Context{Root: "/tmp", Config: cfg, Discovery: disc}, true)

	if len(result.Warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestRunner_WarnChecksCount(t *testing.T) {
	r := NewRunner()
	if r.WarnChecks() != 0 {
		t.Errorf("new runner should have 0 warn checks, got %d", r.WarnChecks())
	}
	r.RegisterWarn(WarnCheckFunc{CheckName: "w", Fn: func(ctx *Context) []string { return nil }})
	if r.WarnChecks() != 1 {
		t.Errorf("expected 1 warn check, got %d", r.WarnChecks())
	}
}

func TestRunner_ErrorsAndWarningsBothPopulated(t *testing.T) {
	r := NewRunner()
	r.Register(CheckFunc{CheckName: "err", Fn: func(ctx *Context) []string {
		return []string{"an error"}
	}})
	r.RegisterWarn(WarnCheckFunc{CheckName: "warn", Fn: func(ctx *Context) []string {
		return []string{"a warning"}
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	result := r.RunWithOpts(&Context{Root: "/tmp", Config: cfg, Discovery: disc}, true)

	if len(result.Errors) != 1 || result.Errors[0] != "an error" {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "a warning" {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}
}

// --- warnMissingDescription ---

func TestWarnMissingDescription_WithDescription(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"hello": {Name: "hello", Description: "Says hello"},
		},
	)
	warns := warnMissingDescription(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %v", warns)
	}
}

func TestWarnMissingDescription_MissingDescription(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	warns := warnMissingDescription(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "hello") {
		t.Errorf("expected warning to mention 'hello', got: %s", warns[0])
	}
	if !strings.Contains(warns[0], "missing description") {
		t.Errorf("expected 'missing description', got: %s", warns[0])
	}
}

func TestWarnMissingDescription_MixedDescriptions(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"with-desc":    {Name: "with-desc", Description: "Has a description"},
			"without-desc": {Name: "without-desc"},
		},
	)
	warns := warnMissingDescription(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "without-desc") {
		t.Errorf("expected warning about 'without-desc', got: %s", warns[0])
	}
}

func TestWarnMissingDescription_SkipsBuiltins(t *testing.T) {
	localAutos := map[string]*automation.Automation{
		"local": {Name: "local"},
	}
	disc := discovery.NewResult(localAutos, []string{"local"})

	builtinAutos := map[string]*automation.Automation{
		"hello": {Name: "hello"},
	}
	builtinDisc := discovery.NewResult(builtinAutos, []string{"hello"})
	disc.MergeBuiltins(builtinDisc)

	ctx := &Context{
		Root:      t.TempDir(),
		Config:    &config.ProjectConfig{Project: "test"},
		Discovery: disc,
	}
	warns := warnMissingDescription(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning (local only), got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "local") {
		t.Errorf("expected warning about 'local', got: %s", warns[0])
	}
}

// --- warnUnusedAutomations ---

func TestWarnUnusedAutomations_ReferencedByShortcut(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"h": {Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings for referenced automation, got: %v", warns)
	}
}

func TestWarnUnusedAutomations_ReferencedBySetup(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup:   []config.SetupEntry{{Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings for setup-referenced automation, got: %v", warns)
	}
}

func TestWarnUnusedAutomations_ReferencedByRunStep(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "callee"},
				},
			},
			"callee": {Name: "callee"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	// "caller" is unreferenced but "callee" is referenced
	found := false
	for _, w := range warns {
		if strings.Contains(w, "callee") {
			t.Errorf("callee should not be warned about, got: %s", w)
		}
		if strings.Contains(w, "caller") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about unreferenced 'caller', got: %v", warns)
	}
}

func TestWarnUnusedAutomations_Unreferenced(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"orphan": {Name: "orphan"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "orphan") {
		t.Errorf("expected warning about 'orphan', got: %s", warns[0])
	}
	if !strings.Contains(warns[0], "not referenced") {
		t.Errorf("expected 'not referenced', got: %s", warns[0])
	}
}

func TestWarnUnusedAutomations_MultipleUnreferencedSorted(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"zzz-orphan": {Name: "zzz-orphan"},
			"aaa-orphan": {Name: "aaa-orphan"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	if len(warns) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "aaa-orphan") {
		t.Errorf("expected first warning about 'aaa-orphan' (sorted), got: %s", warns[0])
	}
	if !strings.Contains(warns[1], "zzz-orphan") {
		t.Errorf("expected second warning about 'zzz-orphan' (sorted), got: %s", warns[1])
	}
}

func TestWarnUnusedAutomations_SkipsBuiltins(t *testing.T) {
	localAutos := map[string]*automation.Automation{
		"local-used": {Name: "local-used"},
	}
	disc := discovery.NewResult(localAutos, []string{"local-used"})

	builtinAutos := map[string]*automation.Automation{
		"hello": {Name: "hello"},
	}
	builtinDisc := discovery.NewResult(builtinAutos, []string{"hello"})
	disc.MergeBuiltins(builtinDisc)

	ctx := &Context{
		Root: t.TempDir(),
		Config: &config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"lu": {Run: "local-used"}},
		},
		Discovery: disc,
	}
	warns := warnUnusedAutomations(ctx)
	for _, w := range warns {
		if strings.Contains(w, "hello") {
			t.Errorf("builtins should be skipped, got warning: %s", w)
		}
	}
}

func TestWarnUnusedAutomations_AllReferenced(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"a": {Run: "alpha"}},
			Setup:     []config.SetupEntry{{Run: "beta"}},
		},
		map[string]*automation.Automation{
			"alpha": {Name: "alpha"},
			"beta":  {Name: "beta"},
		},
	)
	warns := warnUnusedAutomations(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings when all referenced, got: %v", warns)
	}
}

// --- warnShortcutShadowing ---

func TestWarnShortcutShadowing_NoShadow(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"vpup": {Run: "docker/up"}},
		},
		map[string]*automation.Automation{
			"docker/up": {Name: "docker/up"},
		},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings, got: %v", warns)
	}
}

func TestWarnShortcutShadowing_ShellBuiltin(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"test": {Run: "tests/run"}},
		},
		map[string]*automation.Automation{
			"tests/run": {Name: "tests/run"},
		},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "test") {
		t.Errorf("expected warning about 'test', got: %s", warns[0])
	}
	if !strings.Contains(warns[0], "shell builtin") {
		t.Errorf("expected 'shell builtin', got: %s", warns[0])
	}
}

func TestWarnShortcutShadowing_CommonCommand(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"git": {Run: "git/status"}},
		},
		map[string]*automation.Automation{
			"git/status": {Name: "git/status"},
		},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.Contains(warns[0], "common command") {
		t.Errorf("expected 'common command', got: %s", warns[0])
	}
}

func TestWarnShortcutShadowing_MultipleShadows(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Shortcuts: map[string]config.Shortcut{
				"test": {Run: "tests/run"},
				"git":  {Run: "git/status"},
				"vpup": {Run: "docker/up"},
			},
		},
		map[string]*automation.Automation{
			"tests/run":  {Name: "tests/run"},
			"git/status": {Name: "git/status"},
			"docker/up":  {Name: "docker/up"},
		},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 2 {
		t.Fatalf("expected 2 warnings (test + git), got %d: %v", len(warns), warns)
	}
}

func TestWarnShortcutShadowing_NoShortcuts(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings for no shortcuts, got: %v", warns)
	}
}

func TestWarnShortcutShadowing_IncludesPiYamlPrefix(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"echo": {Run: "tools/echo"}},
		},
		map[string]*automation.Automation{
			"tools/echo": {Name: "tools/echo"},
		},
	)
	warns := warnShortcutShadowing(ctx)
	if len(warns) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warns), warns)
	}
	if !strings.HasPrefix(warns[0], "pi.yaml:") {
		t.Errorf("expected warning to start with 'pi.yaml:', got: %s", warns[0])
	}
}

// --- Integration: RunWithOpts with real warning checks ---

func TestRunWithOpts_WarningsWithCleanProject(t *testing.T) {
	r := DefaultRunner()
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"h": {Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {
				Name:        "hello",
				Description: "Says hello",
				Steps:       []automation.Step{{Type: automation.StepTypeBash, Value: "echo hello"}},
			},
		},
	)

	result := r.RunWithOpts(ctx, true)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings for clean project, got: %v", result.Warnings)
	}
}

func TestRunWithOpts_AllWarningTypes(t *testing.T) {
	r := DefaultRunner()
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Shortcuts: map[string]config.Shortcut{
				"test": {Run: "no-desc"},
			},
		},
		map[string]*automation.Automation{
			"no-desc": {
				Name:  "no-desc",
				Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo ok"}},
			},
			"orphan": {
				Name:        "orphan",
				Description: "An orphaned automation",
				Steps:       []automation.Step{{Type: automation.StepTypeBash, Value: "echo ok"}},
			},
		},
	)

	result := r.RunWithOpts(ctx, true)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}

	hasDescWarn := false
	hasUnusedWarn := false
	hasShadowWarn := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "missing description") {
			hasDescWarn = true
		}
		if strings.Contains(w, "not referenced") {
			hasUnusedWarn = true
		}
		if strings.Contains(w, "shell builtin") {
			hasShadowWarn = true
		}
	}
	if !hasDescWarn {
		t.Error("expected missing-description warning")
	}
	if !hasUnusedWarn {
		t.Error("expected unused-automations warning")
	}
	if !hasShadowWarn {
		t.Error("expected shortcut-shadowing warning")
	}
}
