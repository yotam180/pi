package automation

import (
	"strings"
	"testing"
)

func TestLoad_InputsBlock(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "greet.yaml", `
name: greet
description: Greet someone

inputs:
  name:
    type: string
    required: true
    description: Who to greet
  greeting:
    type: string
    required: false
    default: hello
    description: The greeting word

steps:
  - bash: echo "$PI_INPUT_GREETING $PI_INPUT_NAME"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(a.Inputs))
	}
	if len(a.InputKeys) != 2 {
		t.Fatalf("expected 2 input keys, got %d", len(a.InputKeys))
	}
	if a.InputKeys[0] != "name" || a.InputKeys[1] != "greeting" {
		t.Errorf("input keys order wrong: %v", a.InputKeys)
	}

	nameSpec := a.Inputs["name"]
	if !nameSpec.IsRequired() {
		t.Error("name input should be required")
	}
	if nameSpec.Description != "Who to greet" {
		t.Errorf("name description = %q", nameSpec.Description)
	}

	greetSpec := a.Inputs["greeting"]
	if greetSpec.IsRequired() {
		t.Error("greeting input should not be required")
	}
	if greetSpec.Default != "hello" {
		t.Errorf("greeting default = %q, want %q", greetSpec.Default, "hello")
	}
}

func TestLoad_NoInputs(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "simple.yaml", `
name: simple
steps:
  - bash: echo hi
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Inputs != nil {
		t.Errorf("expected nil inputs, got %v", a.Inputs)
	}
	if a.InputKeys != nil {
		t.Errorf("expected nil input keys, got %v", a.InputKeys)
	}
}

func TestLoad_RunStepWithWith(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "caller.yaml", `
name: caller
steps:
  - run: greet
    with:
      name: world
      greeting: hi
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].With["name"] != "world" {
		t.Errorf("with[name] = %q, want %q", a.Steps[0].With["name"], "world")
	}
	if a.Steps[0].With["greeting"] != "hi" {
		t.Errorf("with[greeting] = %q, want %q", a.Steps[0].With["greeting"], "hi")
	}
}

func TestLoad_WithOnNonRunStep_Error(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
steps:
  - bash: echo hello
    with:
      key: value
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for 'with' on non-run step")
	}
	if !strings.Contains(err.Error(), "only valid on 'run' steps") {
		t.Errorf("error should mention 'run' steps, got: %v", err)
	}
}

func TestInputSpec_IsRequired_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name string
		spec InputSpec
		want bool
	}{
		{"explicit required true", InputSpec{Required: boolPtr(true)}, true},
		{"explicit required false", InputSpec{Required: boolPtr(false)}, false},
		{"no required, no default", InputSpec{}, true},
		{"no required, has default", InputSpec{Default: "foo"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.IsRequired(); got != tt.want {
				t.Errorf("IsRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveInputs_Positional(t *testing.T) {
	a := &Automation{
		Inputs: map[string]InputSpec{
			"service": {Description: "service name"},
			"tail":    {Default: "200", Description: "lines"},
		},
		InputKeys: []string{"service", "tail"},
	}

	resolved, err := a.ResolveInputs(nil, []string{"api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved["service"] != "api" {
		t.Errorf("service = %q, want %q", resolved["service"], "api")
	}
	if resolved["tail"] != "200" {
		t.Errorf("tail = %q, want %q (default)", resolved["tail"], "200")
	}
}

func TestResolveInputs_WithArgs(t *testing.T) {
	a := &Automation{
		Inputs: map[string]InputSpec{
			"service": {Description: "service name"},
			"tail":    {Default: "200"},
		},
		InputKeys: []string{"service", "tail"},
	}

	resolved, err := a.ResolveInputs(map[string]string{"service": "web", "tail": "500"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved["service"] != "web" {
		t.Errorf("service = %q, want %q", resolved["service"], "web")
	}
	if resolved["tail"] != "500" {
		t.Errorf("tail = %q, want %q", resolved["tail"], "500")
	}
}

func TestResolveInputs_MissingRequired(t *testing.T) {
	a := &Automation{
		Inputs: map[string]InputSpec{
			"service": {Required: boolPtr(true)},
		},
		InputKeys: []string{"service"},
	}

	_, err := a.ResolveInputs(nil, nil)
	if err == nil {
		t.Fatal("expected error for missing required input")
	}
	if !strings.Contains(err.Error(), "required input \"service\"") {
		t.Errorf("expected error about required service, got: %v", err)
	}
}

func TestResolveInputs_UnknownWith(t *testing.T) {
	a := &Automation{
		Inputs:    map[string]InputSpec{"service": {}},
		InputKeys: []string{"service"},
	}

	_, err := a.ResolveInputs(map[string]string{"typo": "val"}, nil)
	if err == nil {
		t.Fatal("expected error for unknown input key")
	}
	if !strings.Contains(err.Error(), "unknown input \"typo\"") {
		t.Errorf("expected unknown input error, got: %v", err)
	}
}

func TestResolveInputs_MixingError(t *testing.T) {
	a := &Automation{
		Inputs:    map[string]InputSpec{"x": {}},
		InputKeys: []string{"x"},
	}

	_, err := a.ResolveInputs(map[string]string{"x": "1"}, []string{"2"})
	if err == nil {
		t.Fatal("expected error for mixing --with and positional")
	}
	if !strings.Contains(err.Error(), "cannot mix") {
		t.Errorf("expected 'cannot mix' error, got: %v", err)
	}
}

func TestResolveInputs_AllDefaults(t *testing.T) {
	a := &Automation{
		Inputs: map[string]InputSpec{
			"a": {Default: "1"},
			"b": {Default: "2"},
		},
		InputKeys: []string{"a", "b"},
	}

	resolved, err := a.ResolveInputs(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved["a"] != "1" || resolved["b"] != "2" {
		t.Errorf("defaults not applied: %v", resolved)
	}
}

func TestResolveInputs_NoInputs(t *testing.T) {
	a := &Automation{}
	resolved, err := a.ResolveInputs(nil, []string{"ignored"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != nil {
		t.Errorf("expected nil resolved, got: %v", resolved)
	}
}

func TestInputEnvVars(t *testing.T) {
	vars := InputEnvVars(map[string]string{
		"service": "api",
		"tail":    "200",
	})
	found := make(map[string]bool)
	for _, v := range vars {
		found[v] = true
	}
	if !found["PI_IN_SERVICE=api"] {
		t.Error("missing PI_IN_SERVICE=api")
	}
	if !found["PI_INPUT_SERVICE=api"] {
		t.Error("missing PI_INPUT_SERVICE=api (deprecated)")
	}
	if !found["PI_IN_TAIL=200"] {
		t.Error("missing PI_IN_TAIL=200")
	}
	if !found["PI_INPUT_TAIL=200"] {
		t.Error("missing PI_INPUT_TAIL=200 (deprecated)")
	}
}

func TestInputEnvVars_Empty(t *testing.T) {
	vars := InputEnvVars(nil)
	if vars != nil {
		t.Errorf("expected nil for empty inputs, got: %v", vars)
	}
}

func TestInputEnvVars_HyphenToUnderscore(t *testing.T) {
	vars := InputEnvVars(map[string]string{"my-input": "val"})
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars (PI_IN_ + PI_INPUT_), got: %v", vars)
	}
	found := make(map[string]bool)
	for _, v := range vars {
		found[v] = true
	}
	if !found["PI_IN_MY_INPUT=val"] {
		t.Error("missing PI_IN_MY_INPUT=val")
	}
	if !found["PI_INPUT_MY_INPUT=val"] {
		t.Error("missing PI_INPUT_MY_INPUT=val (deprecated)")
	}
}

func TestInputEnvVars_DeterministicOrder(t *testing.T) {
	input := map[string]string{
		"zebra":  "z",
		"alpha":  "a",
		"middle": "m",
		"beta":   "b",
	}
	for i := 0; i < 20; i++ {
		vars := InputEnvVars(input)
		if len(vars) != 8 {
			t.Fatalf("expected 8 vars (4 PI_IN_ + 4 PI_INPUT_), got %d", len(vars))
		}
		expected := []string{
			"PI_IN_ALPHA=a",
			"PI_INPUT_ALPHA=a",
			"PI_IN_BETA=b",
			"PI_INPUT_BETA=b",
			"PI_IN_MIDDLE=m",
			"PI_INPUT_MIDDLE=m",
			"PI_IN_ZEBRA=z",
			"PI_INPUT_ZEBRA=z",
		}
		for j, want := range expected {
			if vars[j] != want {
				t.Errorf("iteration %d: expected %s at [%d], got %s", i, want, j, vars[j])
			}
		}
	}
}

