package executor

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
)

func TestOutputsLast_PassedViaWith(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "received: $PI_IN_MSG"`),
	)

	a := newAutomation("caller",
		bashStep(`echo "42"`),
		runStepWith("target", map[string]string{"msg": "outputs.last"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "received: 42") {
		t.Errorf("output = %q, want to contain %q", output, "received: 42")
	}
}

func TestOutputsLast_MultiStepChain(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "got: $PI_IN_MSG"`),
	)

	a := newAutomation("caller",
		bashStep(`echo "first"`),
		bashStep(`echo "second"`),
		bashStep(`echo "third"`),
		runStepWith("target", map[string]string{"msg": "outputs.last"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "got: third") {
		t.Errorf("output = %q, want to contain %q", output, "got: third")
	}
}

func TestOutputsLast_PipeStillCaptured(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "got: $PI_IN_MSG"`),
	)

	a := newAutomation("caller",
		pipedBashStep(`echo "piped data"`),
		bashStep(`cat`),
		runStepWith("target", map[string]string{"msg": "outputs.last"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "got: piped data") {
		t.Errorf("output = %q, want to contain %q", output, "got: piped data")
	}
}

func TestOutputsLast_SilentStepStillCaptured(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "got: $PI_IN_MSG"`),
	)

	silentStep := bashStep(`echo "secret"`)
	silentStep.Silent = true

	a := newAutomation("caller",
		silentStep,
		runStepWith("target", map[string]string{"msg": "outputs.last"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "got: secret") {
		t.Errorf("output = %q, want to contain %q", output, "got: secret")
	}
}

func TestOutputsLast_IndexedOutput(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "got: $PI_IN_MSG"`),
	)

	a := newAutomation("caller",
		bashStep(`echo "first"`),
		bashStep(`echo "second"`),
		runStepWith("target", map[string]string{"msg": "outputs.0"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "got: first") {
		t.Errorf("output = %q, want to contain %q", output, "got: first")
	}
}

func TestOutputsLast_InputsInterpolation(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"ver": {Type: "string"},
		},
		[]string{"ver"},
		bashStep(`echo "version: $PI_IN_VER"`),
	)

	caller := automationWithInputs("caller",
		map[string]automation.InputSpec{
			"version": {Type: "string"},
		},
		[]string{"version"},
		runStepWith("target", map[string]string{"ver": "inputs.version"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": caller,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.RunWithInputs(caller, nil, map[string]string{"version": "22"}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "version: 22") {
		t.Errorf("output = %q, want to contain %q", output, "version: 22")
	}
}

func TestOutputsLast_CombinedOutputAndInputs(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"actual":   {Type: "string"},
			"required": {Type: "string"},
		},
		[]string{"actual", "required"},
		bashStep(`echo "actual=$PI_IN_ACTUAL required=$PI_IN_REQUIRED"`),
	)

	caller := automationWithInputs("caller",
		map[string]automation.InputSpec{
			"version": {Type: "string"},
		},
		[]string{"version"},
		bashStep(`echo "22.3.1"`),
		runStepWith("target", map[string]string{
			"actual":   "outputs.last",
			"required": "inputs.version",
		}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": caller,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.RunWithInputs(caller, nil, map[string]string{"version": "22"}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "actual=22.3.1") {
		t.Errorf("output = %q, want to contain %q", output, "actual=22.3.1")
	}
	if !strings.Contains(output, "required=22") {
		t.Errorf("output = %q, want to contain %q", output, "required=22")
	}
}

func TestOutputsLast_LiteralNotInterpolated(t *testing.T) {
	dir := t.TempDir()

	target := automationWithInputs("target",
		map[string]automation.InputSpec{
			"msg": {Type: "string"},
		},
		[]string{"msg"},
		bashStep(`echo "got: $PI_IN_MSG"`),
	)

	a := newAutomation("caller",
		runStepWith("target", map[string]string{"msg": "hello world"}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller": a,
		"target": target,
	})

	e, stdout, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(a, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "got: hello world") {
		t.Errorf("output = %q, want to contain %q", output, "got: hello world")
	}
}

func TestGoFunc_VersionSatisfies(t *testing.T) {
	dir := t.TempDir()

	vsAutomation := &automation.Automation{
		Name: "version-satisfies",
		Inputs: map[string]automation.InputSpec{
			"version":  {Type: "string"},
			"required": {Type: "string"},
		},
		InputKeys: []string{"version", "required"},
		GoFunc: func(inputs map[string]string) error {
			if inputs["version"] == "22.3.1" && inputs["required"] == "22" {
				return nil
			}
			return &ExitError{Code: 1}
		},
	}

	disc := newDiscovery(map[string]*automation.Automation{
		"version-satisfies": vsAutomation,
	})

	e, _, _ := newExecutorWithCapture(dir, disc)

	err := e.RunWithInputs(vsAutomation, nil, map[string]string{
		"version":  "22.3.1",
		"required": "22",
	})
	if err != nil {
		t.Fatalf("Run with matching version: %v", err)
	}

	err = e.RunWithInputs(vsAutomation, nil, map[string]string{
		"version":  "18.0.0",
		"required": "22",
	})
	if err == nil {
		t.Fatal("Run with non-matching version: expected error, got nil")
	}
}

func TestGoFunc_CalledViaRunStep(t *testing.T) {
	dir := t.TempDir()

	vsAutomation := &automation.Automation{
		Name: "version-satisfies",
		Inputs: map[string]automation.InputSpec{
			"version":  {Type: "string"},
			"required": {Type: "string"},
		},
		InputKeys: []string{"version", "required"},
		GoFunc: func(inputs map[string]string) error {
			if inputs["version"] != "42" {
				return &ExitError{Code: 1}
			}
			return nil
		},
	}

	caller := newAutomation("caller",
		bashStep(`echo "42"`),
		runStepWith("version-satisfies", map[string]string{
			"version":  "outputs.last",
			"required": "42",
		}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller":            caller,
		"version-satisfies": vsAutomation,
	})

	e, _, _ := newExecutorWithCapture(dir, disc)

	if err := e.Run(caller, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestGoFunc_CalledViaRunStep_Failure(t *testing.T) {
	dir := t.TempDir()

	vsAutomation := &automation.Automation{
		Name: "version-satisfies",
		Inputs: map[string]automation.InputSpec{
			"version":  {Type: "string"},
			"required": {Type: "string"},
		},
		InputKeys: []string{"version", "required"},
		GoFunc: func(inputs map[string]string) error {
			if inputs["version"] != "99" {
				return &ExitError{Code: 1}
			}
			return nil
		},
	}

	caller := newAutomation("caller",
		bashStep(`echo "42"`),
		runStepWith("version-satisfies", map[string]string{
			"version":  "outputs.last",
			"required": "99",
		}),
	)

	disc := newDiscovery(map[string]*automation.Automation{
		"caller":            caller,
		"version-satisfies": vsAutomation,
	})

	e, _, _ := newExecutorWithCapture(dir, disc)

	err := e.Run(caller, nil)
	if err == nil {
		t.Fatal("expected error from version-satisfies, got nil")
	}
}

func TestGoFunc_SkippedByCondition(t *testing.T) {
	dir := t.TempDir()

	called := false
	goFunc := &automation.Automation{
		Name: "gofunc",
		If:   "os.windows",
		GoFunc: func(inputs map[string]string) error {
			called = true
			return nil
		},
	}

	disc := newDiscovery(map[string]*automation.Automation{"gofunc": goFunc})
	env := fakeRuntimeEnv("linux")
	e, _, _ := newExecutorWithEnv(dir, disc, env)

	if err := e.Run(goFunc, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if called {
		t.Error("GoFunc was called despite condition being false")
	}
}

func TestInterpolateValue_NoMatch(t *testing.T) {
	e := &Executor{}

	if got := e.interpolateValue("hello", nil); got != "hello" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "hello", got, "hello")
	}

	if got := e.interpolateValue("outputs.last", nil); got != "" {
		t.Errorf("interpolateValue(%q) = %q, want empty", "outputs.last", got)
	}
}

func TestInterpolateValue_OutputsLast(t *testing.T) {
	e := &Executor{stepOutputs: []string{"first", "second"}}

	if got := e.interpolateValue("outputs.last", nil); got != "second" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "outputs.last", got, "second")
	}
}

func TestInterpolateValue_OutputsIndexed(t *testing.T) {
	e := &Executor{stepOutputs: []string{"zero", "one", "two"}}

	if got := e.interpolateValue("outputs.0", nil); got != "zero" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "outputs.0", got, "zero")
	}
	if got := e.interpolateValue("outputs.2", nil); got != "two" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "outputs.2", got, "two")
	}
	if got := e.interpolateValue("outputs.99", nil); got != "outputs.99" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "outputs.99", got, "outputs.99")
	}
}

func TestInterpolateValue_Inputs(t *testing.T) {
	e := &Executor{}
	inputEnv := []string{"PI_IN_VERSION=22", "PI_INPUT_VERSION=22"}

	if got := e.interpolateValue("inputs.version", inputEnv); got != "22" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "inputs.version", got, "22")
	}

	if got := e.interpolateValue("inputs.unknown", inputEnv); got != "inputs.unknown" {
		t.Errorf("interpolateValue(%q) = %q, want %q", "inputs.unknown", got, "inputs.unknown")
	}
}
