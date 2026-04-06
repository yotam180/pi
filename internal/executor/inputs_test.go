package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestRunWithInputs_EnvVarsInjected(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"name": {Description: "who"},
		},
		[]string{"name"},
		bashStep(`echo "hello $PI_INPUT_NAME"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"name": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "hello world" {
		t.Errorf("output = %q, want %q", got, "hello world")
	}
}

func TestRunWithInputs_Positional(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"x": {},
		},
		[]string{"x"},
		bashStep(`echo "$PI_INPUT_X"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"42"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "42" {
		t.Errorf("output = %q, want %q", got, "42")
	}
}

func TestRunWithInputs_Defaults(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"greeting": {Default: "hi"},
		},
		[]string{"greeting"},
		bashStep(`echo "$PI_INPUT_GREETING"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "hi" {
		t.Errorf("output = %q, want %q", got, "hi")
	}
}

func TestRunWithInputs_MissingRequired(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"required_arg": {Required: boolPtr(true)},
		},
		[]string{"required_arg"},
		bashStep("echo should not run"),
	)

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing required input")
	}
	if !strings.Contains(err.Error(), "required input") {
		t.Errorf("expected 'required input' error, got: %v", err)
	}
}

func TestRunWithInputs_MixingError(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"x": {},
		},
		[]string{"x"},
		bashStep("echo should not run"),
	)

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"pos"}, map[string]string{"x": "with"})
	if err == nil {
		t.Fatal("expected error for mixing positional and --with")
	}
	if !strings.Contains(err.Error(), "cannot mix") {
		t.Errorf("expected 'cannot mix' error, got: %v", err)
	}
}

func TestRunWithInputs_ExcessPositionalArgs(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"target": {Description: "build target"},
		},
		[]string{"target"},
		bashStep("echo should not run"),
	)

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"release", "--verbose"}, nil)
	if err == nil {
		t.Fatal("expected error for excess positional args")
	}
	if !strings.Contains(err.Error(), "too many arguments") {
		t.Errorf("expected 'too many arguments' error, got: %v", err)
	}
}

func TestRunWithInputs_NoInputsPassesArgsThrough(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	a := newAutomation("test", bashStep(`echo "$1" > `+outFile))

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"passed"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "passed" {
		t.Errorf("output = %q, want %q", got, "passed")
	}
}

func TestRunWithInputs_ShortPrefixEnvVarsInjected(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"name": {Description: "who"},
		},
		[]string{"name"},
		bashStep(`echo "hello $PI_IN_NAME"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"name": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "hello world" {
		t.Errorf("output = %q, want %q", got, "hello world")
	}
}

func TestRunWithInputs_BothPrefixesAvailable(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"val": {},
		},
		[]string{"val"},
		bashStep(`echo "$PI_IN_VAL:$PI_INPUT_VAL"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, map[string]string{"val": "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "42:42" {
		t.Errorf("output = %q, want %q (both prefixes should resolve)", got, "42:42")
	}
}

func TestRunWithInputs_RunStepWithWith(t *testing.T) {
	dir := t.TempDir()
	inner := automationWithInputs("inner",
		map[string]automation.InputSpec{
			"msg": {},
		},
		[]string{"msg"},
		bashStep(`echo "$PI_INPUT_MSG"`),
	)

	outer := newAutomation("outer", automation.Step{
		Type:  automation.StepTypeRun,
		Value: "inner",
		With:  map[string]string{"msg": "from-outer"},
	})

	disc := newDiscovery(map[string]*automation.Automation{
		"inner": inner,
		"outer": outer,
	})
	exec, stdout, _ := newExecutorWithCapture(dir, disc)

	err := exec.Run(outer, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "from-outer" {
		t.Errorf("output = %q, want %q", got, "from-outer")
	}
}

func TestRunWithInputs_PIArgsSetForNoInputAutomation(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", bashStep(`echo "$PI_ARGS"`))

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"--release", "--verbose"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "--release --verbose" {
		t.Errorf("PI_ARGS = %q, want %q", got, "--release --verbose")
	}
}

func TestRunWithInputs_PIArgsNotSetWhenNoArgs(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	a := newAutomation("test", bashStep(`echo ">${PI_ARGS}<" > `+outFile))

	exec := newExecutor(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(outFile)
	got := strings.TrimSpace(string(data))
	if got != "><" {
		t.Errorf("output = %q, want %q (PI_ARGS should be empty)", got, "><")
	}
}

func TestRunWithInputs_PIArgsNotSetWhenInputsConsumeArgs(t *testing.T) {
	dir := t.TempDir()
	a := automationWithInputs("test",
		map[string]automation.InputSpec{
			"target": {Description: "build target"},
		},
		[]string{"target"},
		bashStep(`echo "target=$PI_IN_TARGET args=$PI_ARGS"`),
	)

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"release"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "target=release args=" {
		t.Errorf("output = %q, want %q", got, "target=release args=")
	}
}

func TestRunWithInputs_PIArgsSingleArg(t *testing.T) {
	dir := t.TempDir()
	a := newAutomation("test", bashStep(`echo "cmd $PI_ARGS"`))

	exec, stdout, _ := newExecutorWithCapture(dir, newDiscovery(nil))
	err := exec.RunWithInputs(a, []string{"--ignored"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "cmd --ignored" {
		t.Errorf("output = %q, want %q", got, "cmd --ignored")
	}
}
