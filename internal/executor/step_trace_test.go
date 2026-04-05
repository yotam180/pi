package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestStepTrace_DefaultBehavior(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test",
		bashStep("echo hello"),
		bashStep("echo world"),
	)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "→ bash: echo hello") {
		t.Errorf("expected trace line for first step, got: %q", got)
	}
	if !strings.Contains(got, "→ bash: echo world") {
		t.Errorf("expected trace line for second step, got: %q", got)
	}
}

func TestStepTrace_RunStep(t *testing.T) {
	dir := t.TempDir()
	child := newAutomation("child", bashStep("echo child"))
	a := newAutomation("parent", runStep("child"))

	disc := newDiscovery(map[string]*automation.Automation{
		"child": child,
	})
	exec, _, stderr := newExecutorWithCapture(dir, disc)
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "→ run: child") {
		t.Errorf("expected trace for run step, got: %q", got)
	}
	if !strings.Contains(got, "→ bash: echo child") {
		t.Errorf("expected trace for child bash step, got: %q", got)
	}
}

func TestStepTrace_SilentStep(t *testing.T) {
	dir := t.TempDir()
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo silent-output",
		Silent: true,
	}
	normalStep := bashStep("echo normal-output")
	a := newAutomation("test", silentStep, normalStep)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if strings.Contains(stderrOut, "echo silent-output") {
		t.Errorf("silent step should not print trace, got: %q", stderrOut)
	}
	if !strings.Contains(stderrOut, "→ bash: echo normal-output") {
		t.Errorf("normal step should print trace, got: %q", stderrOut)
	}

	stdoutOut := stdout.String()
	if strings.Contains(stdoutOut, "silent-output") {
		t.Errorf("silent step should suppress stdout, got: %q", stdoutOut)
	}
	if !strings.Contains(stdoutOut, "normal-output") {
		t.Errorf("normal step should print stdout, got: %q", stdoutOut)
	}
}

func TestStepTrace_LoudOverridesSilent(t *testing.T) {
	dir := t.TempDir()
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo loud-output",
		Silent: true,
	}
	a := newAutomation("test", silentStep)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	exec.Loud = true
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "→ bash: echo loud-output") {
		t.Errorf("loud should override silent and print trace, got: %q", stderrOut)
	}

	stdoutOut := stdout.String()
	if !strings.Contains(stdoutOut, "loud-output") {
		t.Errorf("loud should override silent and print output, got: %q", stdoutOut)
	}
}

func TestStepTrace_SilentStepStillExecutes(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "proof.txt")
	silentStep := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo done > " + outFile,
		Silent: true,
	}
	a := newAutomation("test", silentStep)

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("silent step should still execute, but output file missing: %v", err)
	}
}

func TestStepTrace_SilentPipeCapture(t *testing.T) {
	dir := t.TempDir()
	silentPiped := automation.Step{
		Type:   automation.StepTypeBash,
		Value:  "echo piped-data",
		Silent: true,
		Pipe: true,
	}
	receiver := bashStep("cat")
	a := newAutomation("test", silentPiped, receiver)

	exec, stdout, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrOut := stderr.String()
	if strings.Contains(stderrOut, "echo piped-data") {
		t.Errorf("silent piped step should not print trace, got: %q", stderrOut)
	}

	stdoutOut := strings.TrimSpace(stdout.String())
	if stdoutOut != "piped-data" {
		t.Errorf("pipe should still work for silent steps, got: %q", stdoutOut)
	}
}

func TestExpandTraceVars(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		inputEnv       []string
		automationEnv  map[string]string
		stepEnv        map[string]string
		want           string
	}{
		{
			name:  "no variables",
			value: "echo hello",
			want:  "echo hello",
		},
		{
			name:     "input var expanded",
			value:    "cargo build --profile $PI_IN_PROFILE",
			inputEnv: []string{"PI_IN_PROFILE=release"},
			want:     "cargo build --profile release",
		},
		{
			name:     "braced input var expanded",
			value:    "cargo build --profile ${PI_IN_PROFILE}",
			inputEnv: []string{"PI_IN_PROFILE=release"},
			want:     "cargo build --profile release",
		},
		{
			name:     "multiple input vars",
			value:    "$PI_IN_CMD --flag $PI_IN_FLAG",
			inputEnv: []string{"PI_IN_CMD=build", "PI_IN_FLAG=verbose"},
			want:     "build --flag verbose",
		},
		{
			name:          "automation env expanded",
			value:         "go build -o $OUTPUT",
			automationEnv: map[string]string{"OUTPUT": "bin/app"},
			want:          "go build -o bin/app",
		},
		{
			name:    "step env expanded",
			value:   "echo $GREETING",
			stepEnv: map[string]string{"GREETING": "world"},
			want:    "echo world",
		},
		{
			name:          "step env overrides automation env",
			value:         "echo $MODE",
			automationEnv: map[string]string{"MODE": "dev"},
			stepEnv:       map[string]string{"MODE": "prod"},
			want:          "echo prod",
		},
		{
			name:  "unknown vars pass through",
			value: "echo $HOME $UNKNOWN",
			want:  "echo $HOME $UNKNOWN",
		},
		{
			name:     "mixed known and unknown vars",
			value:    "echo $HOME $PI_IN_PROFILE",
			inputEnv: []string{"PI_IN_PROFILE=release"},
			want:     "echo $HOME release",
		},
		{
			name:  "no dollar sign returns unchanged",
			value: "simple command",
			want:  "simple command",
		},
		{
			name:     "deprecated PI_INPUT_ prefix",
			value:    "echo $PI_INPUT_VERSION",
			inputEnv: []string{"PI_IN_VERSION=3.13", "PI_INPUT_VERSION=3.13"},
			want:     "echo 3.13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandTraceVars(tt.value, tt.inputEnv, tt.automationEnv, tt.stepEnv)
			if got != tt.want {
				t.Errorf("expandTraceVars() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStepTrace_InputVarsExpanded(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("build",
		map[string]automation.InputSpec{
			"profile": {Type: "string", Default: "dev"},
		},
		[]string{"profile"},
		bashStep("echo build --profile $PI_IN_PROFILE"),
	)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"profile": "release"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "echo build --profile release") {
		t.Errorf("trace should show expanded variable, got: %q", got)
	}
	if strings.Contains(got, "$PI_IN_PROFILE") {
		t.Errorf("trace should not show raw variable, got: %q", got)
	}
}

func TestStepTrace_AutomationEnvExpanded(t *testing.T) {
	dir := t.TempDir()
	a := &automation.Automation{
		Name:     "build",
		Env:      map[string]string{"MY_OS": "linux", "MY_ARCH": "amd64"},
		Steps:    []automation.Step{bashStep("echo build -o $MY_OS-$MY_ARCH")},
		FilePath: "/fake/path/automation.yaml",
	}

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "echo build -o linux-amd64") {
		t.Errorf("trace should show expanded automation env, got: %q", got)
	}
}

func TestStepTrace_StepEnvExpanded(t *testing.T) {
	dir := t.TempDir()
	step := automation.Step{
		Type:  automation.StepTypeBash,
		Value: "echo $GREETING",
		Env:   map[string]string{"GREETING": "world"},
	}
	a := newAutomation("test", step)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.Run(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "echo world") {
		t.Errorf("trace should show expanded step env, got: %q", got)
	}
}

func TestStepTrace_MultipleInputVarsExpanded(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("build",
		map[string]automation.InputSpec{
			"os":   {Type: "string", Default: "linux"},
			"arch": {Type: "string", Default: "amd64"},
		},
		[]string{"os", "arch"},
		bashStep("echo GOOS=$PI_IN_OS GOARCH=$PI_IN_ARCH"),
	)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"os": "darwin", "arch": "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "echo GOOS=darwin GOARCH=arm64") {
		t.Errorf("trace should show all expanded variables, got: %q", got)
	}
}

func TestStepTrace_FirstBlockVarsExpanded(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("build",
		map[string]automation.InputSpec{
			"tool": {Type: "string", Default: "make"},
		},
		[]string{"tool"},
		automation.Step{
			Type: "",
			First: []automation.Step{
				bashStep("echo using $PI_IN_TOOL"),
			},
		},
	)

	exec, _, stderr := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"tool": "cmake"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stderr.String()
	if !strings.Contains(got, "echo using cmake") {
		t.Errorf("first: block trace should show expanded variable, got: %q", got)
	}
}
