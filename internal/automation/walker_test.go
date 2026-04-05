package automation

import (
	"testing"
)

func TestWalkSteps_RegularSteps(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "echo hello"},
			{Type: StepTypePython, Value: "print('hi')"},
		},
	}

	var visited []StepLocation
	var values []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		visited = append(visited, loc)
		values = append(values, step.Value)
	})

	if len(visited) != 2 {
		t.Fatalf("expected 2 visits, got %d", len(visited))
	}
	if values[0] != "echo hello" || values[1] != "print('hi')" {
		t.Errorf("unexpected values: %v", values)
	}
	if visited[0].Index != 0 || visited[0].FirstIndex != -1 || visited[0].Phase != "" {
		t.Errorf("step[0] location: %+v", visited[0])
	}
	if visited[1].Index != 1 || visited[1].FirstIndex != -1 || visited[1].Phase != "" {
		t.Errorf("step[1] location: %+v", visited[1])
	}
}

func TestWalkSteps_FirstBlock(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "before"},
			{
				First: []Step{
					{Type: StepTypeBash, Value: "first-a", If: "os.macos"},
					{Type: StepTypeBash, Value: "first-b"},
				},
			},
			{Type: StepTypeBash, Value: "after"},
		},
	}

	var visited []StepLocation
	var values []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		visited = append(visited, loc)
		values = append(values, step.Value)
	})

	if len(visited) != 4 {
		t.Fatalf("expected 4 visits, got %d", len(visited))
	}

	// step[0]: regular
	if values[0] != "before" || visited[0].Index != 0 || visited[0].FirstIndex != -1 {
		t.Errorf("step[0]: value=%q loc=%+v", values[0], visited[0])
	}
	// step[1].first[0]
	if values[1] != "first-a" || visited[1].Index != 1 || visited[1].FirstIndex != 0 {
		t.Errorf("step[1].first[0]: value=%q loc=%+v", values[1], visited[1])
	}
	// step[1].first[1]
	if values[2] != "first-b" || visited[2].Index != 1 || visited[2].FirstIndex != 1 {
		t.Errorf("step[1].first[1]: value=%q loc=%+v", values[2], visited[2])
	}
	// step[2]: regular
	if values[3] != "after" || visited[3].Index != 2 || visited[3].FirstIndex != -1 {
		t.Errorf("step[2]: value=%q loc=%+v", values[3], visited[3])
	}
}

func TestWalkSteps_InstallerScalar(t *testing.T) {
	a := &Automation{
		Install: &InstallSpec{
			Test:    InstallPhase{IsScalar: true, Scalar: "command -v go"},
			Run:     InstallPhase{IsScalar: true, Scalar: "brew install go"},
			Version: "go version",
		},
	}

	var visited []StepLocation
	var values []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		visited = append(visited, loc)
		values = append(values, step.Value)
	})

	if len(visited) != 2 {
		t.Fatalf("expected 2 visits (test + run), got %d", len(visited))
	}

	if values[0] != "command -v go" || visited[0].Phase != "test" || !visited[0].IsScalar {
		t.Errorf("test phase: value=%q loc=%+v", values[0], visited[0])
	}
	if values[1] != "brew install go" || visited[1].Phase != "run" || !visited[1].IsScalar {
		t.Errorf("run phase: value=%q loc=%+v", values[1], visited[1])
	}

	for i, loc := range visited {
		if !loc.IsScalar {
			t.Errorf("visit[%d]: expected IsScalar=true", i)
		}
	}
}

func TestWalkSteps_InstallerStepList(t *testing.T) {
	a := &Automation{
		Install: &InstallSpec{
			Test: InstallPhase{
				Steps: []Step{
					{Type: StepTypeBash, Value: "python3 --version"},
				},
			},
			Run: InstallPhase{
				Steps: []Step{
					{
						First: []Step{
							{Type: StepTypeBash, Value: "mise install python", If: "command.mise"},
							{Type: StepTypeBash, Value: "brew install python"},
						},
					},
				},
			},
			Verify: &InstallPhase{
				Steps: []Step{
					{Type: StepTypeBash, Value: "python3 --version"},
				},
			},
			Version: "python3 --version | awk '{print $2}'",
		},
	}

	var visited []StepLocation
	var phases []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		visited = append(visited, loc)
		phases = append(phases, loc.Phase)
	})

	if len(visited) != 4 {
		t.Fatalf("expected 4 visits (1 test + 2 run first: sub-steps + 1 verify), got %d", len(visited))
	}

	if phases[0] != "test" {
		t.Errorf("expected test phase first, got %q", phases[0])
	}
	if phases[1] != "run" || visited[1].FirstIndex != 0 {
		t.Errorf("expected run first[0], got phase=%q loc=%+v", phases[1], visited[1])
	}
	if phases[2] != "run" || visited[2].FirstIndex != 1 {
		t.Errorf("expected run first[1], got phase=%q loc=%+v", phases[2], visited[2])
	}
	if phases[3] != "verify" {
		t.Errorf("expected verify phase last, got %q", phases[3])
	}
}

func TestWalkSteps_EmptyAutomation(t *testing.T) {
	a := &Automation{}

	var count int
	WalkSteps(a, func(step Step, loc StepLocation) {
		count++
	})
	if count != 0 {
		t.Errorf("expected 0 visits for empty automation, got %d", count)
	}
}

func TestWalkSteps_NoVerifyPhase(t *testing.T) {
	a := &Automation{
		Install: &InstallSpec{
			Test: InstallPhase{IsScalar: true, Scalar: "test -f /usr/bin/foo"},
			Run:  InstallPhase{IsScalar: true, Scalar: "apt install foo"},
		},
	}

	var phases []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		phases = append(phases, loc.Phase)
	})

	if len(phases) != 2 {
		t.Fatalf("expected 2 phases (no verify), got %d", len(phases))
	}
	if phases[0] != "test" || phases[1] != "run" {
		t.Errorf("unexpected phases: %v", phases)
	}
}

func TestWalkSteps_MixedStepsAndInstallNeverBoth(t *testing.T) {
	// In practice, automation.validate() prevents having both steps and install.
	// But the walker should handle the theoretical case correctly.
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "echo from-steps"},
		},
		Install: &InstallSpec{
			Test: InstallPhase{IsScalar: true, Scalar: "test-cmd"},
			Run:  InstallPhase{IsScalar: true, Scalar: "run-cmd"},
		},
	}

	var values []string
	WalkSteps(a, func(step Step, loc StepLocation) {
		values = append(values, step.Value)
	})

	// Walker visits both (validation prevents this in practice)
	if len(values) != 3 {
		t.Fatalf("expected 3 visits, got %d", len(values))
	}
}

func TestWalkStepsUntil_StopsEarly(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "first"},
			{Type: StepTypeBash, Value: "second"},
			{Type: StepTypeBash, Value: "third"},
		},
	}

	var values []string
	WalkStepsUntil(a, func(step Step, loc StepLocation) bool {
		values = append(values, step.Value)
		return step.Value == "second"
	})

	if len(values) != 2 {
		t.Fatalf("expected 2 visits (stopped at second), got %d: %v", len(values), values)
	}
}

func TestWalkStepsUntil_StopsInFirstBlock(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{
				First: []Step{
					{Type: StepTypeBash, Value: "first-a"},
					{Type: StepTypeBash, Value: "first-b"},
				},
			},
			{Type: StepTypeBash, Value: "after-first"},
		},
	}

	var values []string
	WalkStepsUntil(a, func(step Step, loc StepLocation) bool {
		values = append(values, step.Value)
		return step.Value == "first-a"
	})

	if len(values) != 1 {
		t.Fatalf("expected 1 visit, got %d: %v", len(values), values)
	}
}

func TestWalkStepsUntil_StopsInInstallPhase(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "regular"},
		},
		Install: &InstallSpec{
			Test: InstallPhase{IsScalar: true, Scalar: "test-cmd"},
			Run:  InstallPhase{IsScalar: true, Scalar: "run-cmd"},
		},
	}

	var values []string
	WalkStepsUntil(a, func(step Step, loc StepLocation) bool {
		values = append(values, step.Value)
		return step.Value == "test-cmd"
	})

	if len(values) != 2 {
		t.Fatalf("expected 2 visits (regular + test), got %d: %v", len(values), values)
	}
}

func TestWalkStepsUntil_CompletesWhenNeverTrue(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeBash, Value: "one"},
			{Type: StepTypeBash, Value: "two"},
		},
	}

	var count int
	WalkStepsUntil(a, func(step Step, loc StepLocation) bool {
		count++
		return false
	})

	if count != 2 {
		t.Fatalf("expected 2 visits, got %d", count)
	}
}

func TestStepLocation_FormatPath_Regular(t *testing.T) {
	loc := StepLocation{Phase: "", Index: 2, FirstIndex: -1}
	got := loc.FormatPath("my-auto")
	want := "my-auto: step[2]"
	if got != want {
		t.Errorf("FormatPath = %q, want %q", got, want)
	}
}

func TestStepLocation_FormatPath_FirstBlock(t *testing.T) {
	loc := StepLocation{Phase: "", Index: 1, FirstIndex: 0}
	got := loc.FormatPath("deploy")
	want := "deploy: step[1].first[0]"
	if got != want {
		t.Errorf("FormatPath = %q, want %q", got, want)
	}
}

func TestStepLocation_FormatPath_InstallPhase(t *testing.T) {
	loc := StepLocation{Phase: "run", Index: 0, FirstIndex: -1}
	got := loc.FormatPath("install-go")
	want := "install-go: install.run step[0]"
	if got != want {
		t.Errorf("FormatPath = %q, want %q", got, want)
	}
}

func TestStepLocation_FormatPath_InstallPhaseFirstBlock(t *testing.T) {
	loc := StepLocation{Phase: "run", Index: 0, FirstIndex: 1}
	got := loc.FormatPath("install-python")
	want := "install-python: install.run step[0].first[1]"
	if got != want {
		t.Errorf("FormatPath = %q, want %q", got, want)
	}
}

func TestStepLocation_FormatPath_Scalar(t *testing.T) {
	loc := StepLocation{Phase: "test", Index: 0, FirstIndex: -1, IsScalar: true}
	got := loc.FormatPath("install-uv")
	want := "install-uv: install.test"
	if got != want {
		t.Errorf("FormatPath = %q, want %q", got, want)
	}
}

func TestWalkSteps_PreservesStepFields(t *testing.T) {
	a := &Automation{
		Steps: []Step{
			{Type: StepTypeRun, Value: "other/auto", If: "os.macos", Silent: true},
		},
	}

	WalkSteps(a, func(step Step, loc StepLocation) {
		if step.Type != StepTypeRun {
			t.Errorf("expected run type, got %q", step.Type)
		}
		if step.If != "os.macos" {
			t.Errorf("expected if: os.macos, got %q", step.If)
		}
		if !step.Silent {
			t.Error("expected silent: true")
		}
	})
}
