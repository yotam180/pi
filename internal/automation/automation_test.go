package automation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ValidBashStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "up.yaml", `
name: docker/up
description: Start all containers

steps:
  - bash: docker-compose up -d
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Name != "docker/up" {
		t.Errorf("name = %q, want %q", a.Name, "docker/up")
	}
	if a.Description != "Start all containers" {
		t.Errorf("description = %q, want %q", a.Description, "Start all containers")
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeBash)
	}
	if a.Steps[0].Value != "docker-compose up -d" {
		t.Errorf("step value = %q, want %q", a.Steps[0].Value, "docker-compose up -d")
	}
}

func TestLoad_MultipleSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "deploy.yaml", `
name: deploy
description: Build and deploy

steps:
  - bash: make build
  - run: docker/push
  - bash: kubectl apply -f deploy.yaml
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(a.Steps))
	}

	want := []struct {
		t StepType
		v string
	}{
		{StepTypeBash, "make build"},
		{StepTypeRun, "docker/push"},
		{StepTypeBash, "kubectl apply -f deploy.yaml"},
	}
	for i, w := range want {
		if a.Steps[i].Type != w.t {
			t.Errorf("step[%d].Type = %q, want %q", i, a.Steps[i].Type, w.t)
		}
		if a.Steps[i].Value != w.v {
			t.Errorf("step[%d].Value = %q, want %q", i, a.Steps[i].Value, w.v)
		}
	}
}

func TestLoad_PipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "logs.yaml", `
name: logs
description: Stream and format logs

steps:
  - bash: docker-compose logs -f
    pipe_to: next
  - bash: format-logs.sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Steps[0].PipeTo != "next" {
		t.Errorf("step[0].PipeTo = %q, want %q", a.Steps[0].PipeTo, "next")
	}
	if a.Steps[1].PipeTo != "" {
		t.Errorf("step[1].PipeTo = %q, want empty", a.Steps[1].PipeTo)
	}
}

func TestLoad_InlineBashMultiline(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "setup.yaml", `
name: setup/install-uv
description: Install uv if missing

steps:
  - bash: |
      if ! command -v uv &> /dev/null; then
        curl -LsSf https://astral.sh/uv/install.sh | sh
      fi
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeBash)
	}
	if !strings.Contains(a.Steps[0].Value, "command -v uv") {
		t.Errorf("step value should contain multiline bash, got: %q", a.Steps[0].Value)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestLoad_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
description: No name field

steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

func TestLoad_NoSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty.yaml", `
name: empty
description: No steps
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no steps")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("error should mention 'at least one step', got: %v", err)
	}
}

func TestLoad_NoStepType(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "nostep.yaml", `
name: bad
steps:
  - pipe_to: next
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for step with no type")
	}
	if !strings.Contains(err.Error(), "must specify one of") {
		t.Errorf("error should mention 'must specify one of', got: %v", err)
	}
}

func TestLoad_MultipleStepTypes(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "multi.yaml", `
name: bad
steps:
  - bash: echo hello
    run: other/thing
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for multiple step types")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Errorf("error should mention 'exactly one', got: %v", err)
	}
}

func TestLoad_PythonStep_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "py.yaml", `
name: test
steps:
  - python: script.py
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypePython {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypePython)
	}
	if a.Steps[0].Value != "script.py" {
		t.Errorf("step value = %q, want %q", a.Steps[0].Value, "script.py")
	}
}

func TestLoad_TypeScriptStep_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "ts.yaml", `
name: test
steps:
  - typescript: script.ts
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypeTypeScript {
		t.Errorf("step type = %q, want %q", a.Steps[0].Type, StepTypeTypeScript)
	}
	if a.Steps[0].Value != "script.ts" {
		t.Errorf("step value = %q, want %q", a.Steps[0].Value, "script.ts")
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: test
steps: [[[invalid
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention 'parsing', got: %v", err)
	}
}

func TestLoad_EmptyStepValue(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "empty-val.yaml", `
name: test
steps:
  - bash: ""
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty step value")
	}
	if !strings.Contains(err.Error(), "empty value") {
		t.Errorf("error should mention 'empty value', got: %v", err)
	}
}

func TestStepType_IsImplemented(t *testing.T) {
	tests := []struct {
		st   StepType
		want bool
	}{
		{StepTypeBash, true},
		{StepTypeRun, true},
		{StepTypePython, true},
		{StepTypeTypeScript, true},
	}
	for _, tt := range tests {
		if got := tt.st.IsImplemented(); got != tt.want {
			t.Errorf("StepType(%q).IsImplemented() = %v, want %v", tt.st, got, tt.want)
		}
	}
}

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

func boolPtr(b bool) *bool { return &b }

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
	if !found["PI_INPUT_SERVICE=api"] {
		t.Error("missing PI_INPUT_SERVICE=api")
	}
	if !found["PI_INPUT_TAIL=200"] {
		t.Error("missing PI_INPUT_TAIL=200")
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
	if len(vars) != 1 || vars[0] != "PI_INPUT_MY_INPUT=val" {
		t.Errorf("expected PI_INPUT_MY_INPUT=val, got: %v", vars)
	}
}

// --- if: field tests ---

func TestLoad_StepWithIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "cond.yaml", `
name: conditional
description: Test conditional step
steps:
  - bash: echo hello
    if: os.macos
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	if a.Steps[0].If != "os.macos" {
		t.Errorf("step.If = %q, want %q", a.Steps[0].If, "os.macos")
	}
}

func TestLoad_StepWithoutIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-cond.yaml", `
name: normal
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].If != "" {
		t.Errorf("step.If should be empty, got %q", a.Steps[0].If)
	}
}

func TestLoad_StepWithComplexIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "complex-cond.yaml", `
name: complex
steps:
  - bash: echo hello
    if: os.macos and not command.brew
  - bash: echo world
    if: os.linux or os.macos
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(a.Steps))
	}
	if a.Steps[0].If != "os.macos and not command.brew" {
		t.Errorf("step[0].If = %q, want %q", a.Steps[0].If, "os.macos and not command.brew")
	}
	if a.Steps[1].If != "os.linux or os.macos" {
		t.Errorf("step[1].If = %q, want %q", a.Steps[1].If, "os.linux or os.macos")
	}
}

func TestLoad_StepWithIfAndPipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "pipe-cond.yaml", `
name: piped
steps:
  - bash: echo data
    pipe_to: next
    if: os.macos
  - bash: cat
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].If != "os.macos" {
		t.Errorf("step[0].If = %q, want %q", a.Steps[0].If, "os.macos")
	}
	if a.Steps[0].PipeTo != "next" {
		t.Errorf("step[0].PipeTo = %q, want %q", a.Steps[0].PipeTo, "next")
	}
}

func TestLoad_StepWithInvalidIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad-if.yaml", `
name: bad-if
steps:
  - bash: echo hello
    if: "and and and"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid if expression")
	}
	if !strings.Contains(err.Error(), "invalid if expression") {
		t.Errorf("error should mention 'invalid if expression', got: %v", err)
	}
}

func TestLoad_StepWithFuncCallIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "func-if.yaml", `
name: func-cond
steps:
  - bash: echo hello
    if: file.exists(".env")
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].If != `file.exists(".env")` {
		t.Errorf("step.If = %q, want %q", a.Steps[0].If, `file.exists(".env")`)
	}
}
