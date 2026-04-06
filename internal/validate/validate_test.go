package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/discovery"
)

// --- Runner tests ---

func TestRunner_EmptyChecks(t *testing.T) {
	r := NewRunner()
	if r.Checks() != 0 {
		t.Errorf("expected 0 checks, got %d", r.Checks())
	}

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	ctx := &Context{Root: "/tmp", Config: cfg, Discovery: disc}

	result := r.Run(ctx)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestRunner_Register(t *testing.T) {
	r := NewRunner()
	r.Register(CheckFunc{CheckName: "test-check", Fn: func(ctx *Context) []string {
		return nil
	}})
	if r.Checks() != 1 {
		t.Errorf("expected 1 check, got %d", r.Checks())
	}
}

func TestRunner_RunAggregatesErrors(t *testing.T) {
	r := NewRunner()
	r.Register(CheckFunc{CheckName: "check-a", Fn: func(ctx *Context) []string {
		return []string{"error from A"}
	}})
	r.Register(CheckFunc{CheckName: "check-b", Fn: func(ctx *Context) []string {
		return []string{"error from B1", "error from B2"}
	}})
	r.Register(CheckFunc{CheckName: "check-c", Fn: func(ctx *Context) []string {
		return nil
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	ctx := &Context{Root: "/tmp", Config: cfg, Discovery: disc}

	result := r.Run(ctx)
	if len(result.Errors) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", len(result.Errors), result.Errors)
	}
	if result.Errors[0] != "error from A" {
		t.Errorf("first error = %q, want %q", result.Errors[0], "error from A")
	}
}

func TestRunner_CountsFromContext(t *testing.T) {
	r := NewRunner()

	autos := map[string]*automation.Automation{
		"hello": {Name: "hello"},
		"world": {Name: "world"},
	}
	disc := discovery.NewResult(autos, []string{"hello", "world"})
	cfg := &config.ProjectConfig{
		Project: "test",
		Shortcuts: map[string]config.Shortcut{
			"h": {Run: "hello"},
		},
		Setup: []config.SetupEntry{
			{Run: "hello"},
			{Run: "world"},
		},
	}
	ctx := &Context{Root: "/tmp", Config: cfg, Discovery: disc}

	result := r.Run(ctx)
	if result.AutomationCount != 2 {
		t.Errorf("AutomationCount = %d, want 2", result.AutomationCount)
	}
	if result.ShortcutCount != 1 {
		t.Errorf("ShortcutCount = %d, want 1", result.ShortcutCount)
	}
	if result.SetupCount != 2 {
		t.Errorf("SetupCount = %d, want 2", result.SetupCount)
	}
}

func TestRunner_ChecksRunInOrder(t *testing.T) {
	var order []string
	r := NewRunner()
	r.Register(CheckFunc{CheckName: "first", Fn: func(ctx *Context) []string {
		order = append(order, "first")
		return nil
	}})
	r.Register(CheckFunc{CheckName: "second", Fn: func(ctx *Context) []string {
		order = append(order, "second")
		return nil
	}})
	r.Register(CheckFunc{CheckName: "third", Fn: func(ctx *Context) []string {
		order = append(order, "third")
		return nil
	}})

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	r.Run(&Context{Root: "/tmp", Config: cfg, Discovery: disc})

	if len(order) != 3 || order[0] != "first" || order[1] != "second" || order[2] != "third" {
		t.Errorf("checks ran in wrong order: %v", order)
	}
}

func TestDefaultRunner_Has10Checks(t *testing.T) {
	r := DefaultRunner()
	if r.Checks() != 10 {
		t.Errorf("DefaultRunner should have 10 checks, got %d", r.Checks())
	}
}

// --- CheckFunc tests ---

func TestCheckFunc_Name(t *testing.T) {
	c := CheckFunc{CheckName: "my-check"}
	if c.Name() != "my-check" {
		t.Errorf("Name() = %q, want %q", c.Name(), "my-check")
	}
}

func TestCheckFunc_Run(t *testing.T) {
	called := false
	c := CheckFunc{
		CheckName: "test",
		Fn: func(ctx *Context) []string {
			called = true
			return []string{"err"}
		},
	}

	disc := discovery.NewResult(map[string]*automation.Automation{}, nil)
	cfg := &config.ProjectConfig{Project: "test"}
	errs := c.Run(&Context{Root: "/tmp", Config: cfg, Discovery: disc})
	if !called {
		t.Error("Fn was not called")
	}
	if len(errs) != 1 || errs[0] != "err" {
		t.Errorf("unexpected errors: %v", errs)
	}
}

// --- CheckWithInputs tests ---

func TestCheckWithInputs_NoWith(t *testing.T) {
	a := &automation.Automation{Name: "test"}
	msgs := CheckWithInputs(nil, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors for nil with, got: %v", msgs)
	}
	msgs = CheckWithInputs(map[string]string{}, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors for empty with, got: %v", msgs)
	}
}

func TestCheckWithInputs_AllValid(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}, "arch": {}},
		InputKeys: []string{"version", "arch"},
	}
	msgs := CheckWithInputs(map[string]string{"version": "3.13", "arch": "arm64"}, a)
	if len(msgs) != 0 {
		t.Errorf("expected no errors, got: %v", msgs)
	}
}

func TestCheckWithInputs_UnknownKey(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}},
		InputKeys: []string{"version"},
	}
	msgs := CheckWithInputs(map[string]string{"vrsion": "3.13"}, a)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "vrsion") {
		t.Errorf("expected error to mention 'vrsion', got: %s", msgs[0])
	}
	if !strings.Contains(msgs[0], "version") {
		t.Errorf("expected error to list available inputs, got: %s", msgs[0])
	}
}

func TestCheckWithInputs_NoInputsOnTarget(t *testing.T) {
	a := &automation.Automation{Name: "hello"}
	msgs := CheckWithInputs(map[string]string{"name": "world"}, a)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "no declared inputs") {
		t.Errorf("expected 'no declared inputs' message, got: %s", msgs[0])
	}
}

func TestCheckWithInputs_MultipleUnknownSorted(t *testing.T) {
	a := &automation.Automation{
		Name:      "test",
		Inputs:    map[string]automation.InputSpec{"version": {}},
		InputKeys: []string{"version"},
	}
	msgs := CheckWithInputs(map[string]string{"platform": "linux", "arch": "arm64"}, a)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "arch") {
		t.Errorf("expected first error (sorted) to be about 'arch', got: %s", msgs[0])
	}
	if !strings.Contains(msgs[1], "platform") {
		t.Errorf("expected second error (sorted) to be about 'platform', got: %s", msgs[1])
	}
}

// --- DetectCycles tests ---

func TestDetectCycles_NoCycles(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": nil,
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got: %v", cycles)
	}
}

func TestDetectCycles_DirectCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	cycle := cycles[0]
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node, got: %v", cycle)
	}
}

func TestDetectCycles_SelfLoop(t *testing.T) {
	graph := map[string][]string{
		"x": {"x"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	if len(cycles[0]) != 2 {
		t.Errorf("self-loop cycle should have 2 elements (x → x), got: %v", cycles[0])
	}
}

func TestDetectCycles_ThreeNodeCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_DiamondNoCycle(t *testing.T) {
	graph := map[string][]string{
		"top":    {"left", "right"},
		"left":   {"bottom"},
		"right":  {"bottom"},
		"bottom": nil,
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("diamond should have no cycles, got: %v", cycles)
	}
}

func TestDetectCycles_MultipleCycles(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"},
		"c": {"d"},
		"d": {"c"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 2 {
		t.Fatalf("expected 2 cycles, got %d: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_DisconnectedGraphWithCycle(t *testing.T) {
	graph := map[string][]string{
		"a":      {"b"},
		"b":      {"a"},
		"island": nil,
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle (island should not affect detection), got %d: %v", len(cycles), cycles)
	}
}

// --- NormalizeCycleKey tests ---

func TestNormalizeCycleKey_Rotation(t *testing.T) {
	k1 := NormalizeCycleKey([]string{"b", "c", "a", "b"})
	k2 := NormalizeCycleKey([]string{"a", "b", "c", "a"})
	if k1 != k2 {
		t.Errorf("rotated cycles should normalize to same key: %q vs %q", k1, k2)
	}
}

func TestNormalizeCycleKey_SelfLoop(t *testing.T) {
	key := NormalizeCycleKey([]string{"x", "x"})
	if key != "x" {
		t.Errorf("self-loop key = %q, want %q", key, "x")
	}
}

func TestNormalizeCycleKey_SingleNode(t *testing.T) {
	key := NormalizeCycleKey([]string{"only"})
	if key != "only" {
		t.Errorf("single node key = %q, want %q", key, "only")
	}
}

func TestNormalizeCycleKey_Empty(t *testing.T) {
	key := NormalizeCycleKey(nil)
	if key != "" {
		t.Errorf("empty cycle key = %q, want %q", key, "")
	}
}

// --- Individual check smoke tests (via full runner on synthetic data) ---

func newTestContext(t *testing.T, cfg *config.ProjectConfig, autos map[string]*automation.Automation) *Context {
	t.Helper()
	names := make([]string, 0, len(autos))
	for n := range autos {
		names = append(names, n)
	}
	disc := discovery.NewResult(autos, names)
	return &Context{
		Root:      t.TempDir(),
		Config:    cfg,
		Discovery: disc,
	}
}

func TestCheck_ShortcutRefs_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"greet": {Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	errs := checkShortcutRefs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_ShortcutRefs_Broken(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project:   "test",
			Shortcuts: map[string]config.Shortcut{"bad": {Run: "nonexistent"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	errs := checkShortcutRefs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "nonexistent") {
		t.Errorf("expected error to mention 'nonexistent', got: %s", errs[0])
	}
}

func TestCheck_SetupRefs_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup:   []config.SetupEntry{{Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	errs := checkSetupRefs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_SetupRefs_Broken(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup:   []config.SetupEntry{{Run: "missing"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	errs := checkSetupRefs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "setup[0]") {
		t.Errorf("expected error to mention 'setup[0]', got: %s", errs[0])
	}
}

func TestCheck_RunStepRefs_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "target"},
				},
			},
			"target": {Name: "target"},
		},
	)
	errs := checkRunStepRefs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_RunStepRefs_Broken(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "ghost"},
				},
			},
		},
	)
	errs := checkRunStepRefs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "ghost") {
		t.Errorf("expected error to mention 'ghost', got: %s", errs[0])
	}
}

func TestCheck_CircularDeps_NoCycle(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"a": {
				Name: "a",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "b"},
				},
			},
			"b": {
				Name:  "b",
				Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo done"}},
			},
		},
	)
	errs := checkCircularDeps(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_CircularDeps_Cycle(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"a": {
				Name: "a",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "b"},
				},
			},
			"b": {
				Name: "b",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "a"},
				},
			},
		},
	)
	errs := checkCircularDeps(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "circular dependency") {
		t.Errorf("expected 'circular dependency', got: %s", errs[0])
	}
}

func TestCheck_Conditions_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"build": {
				Name: "build",
				If:   "os.macos",
				Steps: []automation.Step{
					{Type: automation.StepTypeBash, Value: "echo test", If: "command.docker"},
				},
			},
		},
	)
	errs := checkConditions(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_Conditions_UnknownPredicate(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"build": {
				Name: "build",
				Steps: []automation.Step{
					{Type: automation.StepTypeBash, Value: "echo test", If: "os.macoss"},
				},
			},
		},
	)
	errs := checkConditions(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "unknown predicate") {
		t.Errorf("expected 'unknown predicate', got: %s", errs[0])
	}
}

func TestCheck_Conditions_AutomationLevel(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"build": {
				Name:  "build",
				If:    "os.freebsd",
				Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo test"}},
			},
		},
	)
	errs := checkConditions(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "build") {
		t.Errorf("expected error to mention 'build', got: %s", errs[0])
	}
}

func TestCheck_ShortcutInputs_ValidInputs(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Shortcuts: map[string]config.Shortcut{
				"py": {Run: "install-python", With: map[string]string{"version": "3.13"}},
			},
		},
		map[string]*automation.Automation{
			"install-python": {
				Name:      "install-python",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkShortcutInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_ShortcutInputs_UnknownKey(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Shortcuts: map[string]config.Shortcut{
				"py": {Run: "install-python", With: map[string]string{"vrsion": "3.13"}},
			},
		},
		map[string]*automation.Automation{
			"install-python": {
				Name:      "install-python",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkShortcutInputs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "vrsion") {
		t.Errorf("expected error to mention 'vrsion', got: %s", errs[0])
	}
}

// --- BuildRunGraph tests ---

func TestBuildRunGraph_Basic(t *testing.T) {
	autos := map[string]*automation.Automation{
		"a": {Name: "a", Steps: []automation.Step{{Type: automation.StepTypeRun, Value: "b"}}},
		"b": {Name: "b", Steps: []automation.Step{{Type: automation.StepTypeBash, Value: "echo done"}}},
	}
	disc := discovery.NewResult(autos, []string{"a", "b"})
	graph := BuildRunGraph(disc)

	if len(graph["a"]) != 1 || graph["a"][0] != "b" {
		t.Errorf("expected a→b, got: %v", graph["a"])
	}
	if graph["b"] != nil {
		t.Errorf("expected b→nil, got: %v", graph["b"])
	}
}

func TestBuildRunGraph_DeduplicatesTargets(t *testing.T) {
	autos := map[string]*automation.Automation{
		"a": {Name: "a", Steps: []automation.Step{
			{Type: automation.StepTypeRun, Value: "b"},
			{Type: automation.StepTypeRun, Value: "b"},
		}},
		"b": {Name: "b"},
	}
	disc := discovery.NewResult(autos, []string{"a", "b"})
	graph := BuildRunGraph(disc)

	if len(graph["a"]) != 1 {
		t.Errorf("expected deduplicated edges, got: %v", graph["a"])
	}
}

// --- Setup input checks ---

func TestCheck_SetupInputs_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup: []config.SetupEntry{
				{Run: "install-python", With: map[string]string{"version": "3.13"}},
			},
		},
		map[string]*automation.Automation{
			"install-python": {
				Name:      "install-python",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkSetupInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_SetupInputs_UnknownKey(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup: []config.SetupEntry{
				{Run: "install-python", With: map[string]string{"vrsion": "3.13"}},
			},
		},
		map[string]*automation.Automation{
			"install-python": {
				Name:      "install-python",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkSetupInputs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "vrsion") {
		t.Errorf("expected error to mention 'vrsion', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "setup[0]") {
		t.Errorf("expected error to mention 'setup[0]', got: %s", errs[0])
	}
}

func TestCheck_SetupInputs_NoWith(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup:   []config.SetupEntry{{Run: "hello"}},
		},
		map[string]*automation.Automation{
			"hello": {Name: "hello"},
		},
	)
	errs := checkSetupInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for setup without with:, got: %v", errs)
	}
}

func TestCheck_SetupInputs_BrokenRefSkipsCheck(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{
			Project: "test",
			Setup: []config.SetupEntry{
				{Run: "nonexistent", With: map[string]string{"version": "3.13"}},
			},
		},
		map[string]*automation.Automation{},
	)
	errs := checkSetupInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no input errors when ref is broken, got: %v", errs)
	}
}

// --- Run step input checks ---

func TestCheck_RunStepInputs_Valid(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "target", With: map[string]string{"version": "3.13"}},
				},
			},
			"target": {
				Name:      "target",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkRunStepInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestCheck_RunStepInputs_UnknownKey(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "target", With: map[string]string{"vrsion": "3.13"}},
				},
			},
			"target": {
				Name:      "target",
				Inputs:    map[string]automation.InputSpec{"version": {}},
				InputKeys: []string{"version"},
			},
		},
	)
	errs := checkRunStepInputs(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "vrsion") {
		t.Errorf("expected error to mention 'vrsion', got: %s", errs[0])
	}
}

func TestCheck_RunStepInputs_NoWith(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "target"},
				},
			},
			"target": {Name: "target"},
		},
	)
	errs := checkRunStepInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for steps without with:, got: %v", errs)
	}
}

func TestCheck_RunStepInputs_BrokenRefSkipsCheck(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "ghost", With: map[string]string{"version": "3.13"}},
				},
			},
		},
	)
	errs := checkRunStepInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no input errors when ref is broken, got: %v", errs)
	}
}

func TestCheck_RunStepInputs_BashStepSkipped(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"hello": {
				Name: "hello",
				Steps: []automation.Step{
					{Type: automation.StepTypeBash, Value: "echo hello"},
				},
			},
		},
	)
	errs := checkRunStepInputs(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for bash steps, got: %v", errs)
	}
}

// --- File reference checks ---

func TestCheck_FileReferences_InlineNotFlagged(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"hello": {
				Name: "hello",
				Steps: []automation.Step{
					{Type: automation.StepTypeBash, Value: "echo hello world"},
					{Type: automation.StepTypePython, Value: "import sys; print('hi')"},
				},
			},
		},
	)
	errs := checkFileReferences(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for inline scripts, got: %v", errs)
	}
}

func TestCheck_FileReferences_RunStepSkipped(t *testing.T) {
	ctx := newTestContext(t,
		&config.ProjectConfig{Project: "test"},
		map[string]*automation.Automation{
			"caller": {
				Name: "caller",
				Steps: []automation.Step{
					{Type: automation.StepTypeRun, Value: "target"},
				},
			},
			"target": {Name: "target"},
		},
	)
	errs := checkFileReferences(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors for run: steps, got: %v", errs)
	}
}

func TestCheck_FileReferences_MissingFile(t *testing.T) {
	dir := t.TempDir()
	autos := map[string]*automation.Automation{
		"build": {
			Name:     "build",
			FilePath: filepath.Join(dir, "build.yaml"),
			Steps: []automation.Step{
				{Type: automation.StepTypeBash, Value: "compile.sh"},
			},
		},
	}
	names := []string{"build"}
	disc := discovery.NewResult(autos, names)
	ctx := &Context{Root: dir, Config: &config.ProjectConfig{Project: "test"}, Discovery: disc}

	errs := checkFileReferences(ctx)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "compile.sh") {
		t.Errorf("expected error to mention 'compile.sh', got: %s", errs[0])
	}
	if !strings.Contains(errs[0], "file not found") {
		t.Errorf("expected 'file not found', got: %s", errs[0])
	}
}

func TestCheck_FileReferences_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "compile.sh"), []byte("#!/bin/bash\necho build\n"), 0755)
	autos := map[string]*automation.Automation{
		"build": {
			Name:     "build",
			FilePath: filepath.Join(dir, "build.yaml"),
			Steps: []automation.Step{
				{Type: automation.StepTypeBash, Value: "compile.sh"},
			},
		},
	}
	names := []string{"build"}
	disc := discovery.NewResult(autos, names)
	ctx := &Context{Root: dir, Config: &config.ProjectConfig{Project: "test"}, Discovery: disc}

	errs := checkFileReferences(ctx)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
