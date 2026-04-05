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
		t.Fatal("expected error for no steps or install")
	}
	if !strings.Contains(err.Error(), "must have") {
		t.Errorf("error should mention missing steps/install, got: %v", err)
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

func TestStepType_IsValid(t *testing.T) {
	tests := []struct {
		st   StepType
		want bool
	}{
		{StepTypeBash, true},
		{StepTypeRun, true},
		{StepTypePython, true},
		{StepTypeTypeScript, true},
		{"unknown", false},
	}
	for _, tt := range tests {
		if got := tt.st.IsValid(); got != tt.want {
			t.Errorf("StepType(%q).IsValid() = %v, want %v", tt.st, got, tt.want)
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

func TestInputEnvVars_DeterministicOrder(t *testing.T) {
	input := map[string]string{
		"zebra":   "z",
		"alpha":   "a",
		"middle":  "m",
		"beta":    "b",
	}
	for i := 0; i < 20; i++ {
		vars := InputEnvVars(input)
		if len(vars) != 4 {
			t.Fatalf("expected 4 vars, got %d", len(vars))
		}
		if vars[0] != "PI_INPUT_ALPHA=a" {
			t.Errorf("iteration %d: expected PI_INPUT_ALPHA=a at [0], got %s", i, vars[0])
		}
		if vars[1] != "PI_INPUT_BETA=b" {
			t.Errorf("iteration %d: expected PI_INPUT_BETA=b at [1], got %s", i, vars[1])
		}
		if vars[2] != "PI_INPUT_MIDDLE=m" {
			t.Errorf("iteration %d: expected PI_INPUT_MIDDLE=m at [2], got %s", i, vars[2])
		}
		if vars[3] != "PI_INPUT_ZEBRA=z" {
			t.Errorf("iteration %d: expected PI_INPUT_ZEBRA=z at [3], got %s", i, vars[3])
		}
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

func TestLoad_StepWithEnv(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "env-step.yaml", `
name: build-with-env
steps:
  - bash: go build ./...
    env:
      GOOS: linux
      GOARCH: amd64
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if len(a.Steps[0].Env) != 2 {
		t.Fatalf("env count = %d, want 2", len(a.Steps[0].Env))
	}
	if a.Steps[0].Env["GOOS"] != "linux" {
		t.Errorf("env[GOOS] = %q, want %q", a.Steps[0].Env["GOOS"], "linux")
	}
	if a.Steps[0].Env["GOARCH"] != "amd64" {
		t.Errorf("env[GOARCH] = %q, want %q", a.Steps[0].Env["GOARCH"], "amd64")
	}
}

func TestLoad_StepWithoutEnv(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-env.yaml", `
name: plain
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps[0].Env) != 0 {
		t.Errorf("env should be empty, got %v", a.Steps[0].Env)
	}
}

func TestLoad_StepEnvWithIfAndPipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "env-combo.yaml", `
name: combo
steps:
  - bash: echo hello
    env:
      FOO: bar
    if: os.macos
    pipe_to: next
  - bash: cat
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Env["FOO"] != "bar" {
		t.Errorf("env[FOO] = %q, want %q", a.Steps[0].Env["FOO"], "bar")
	}
	if a.Steps[0].If != "os.macos" {
		t.Errorf("If = %q, want %q", a.Steps[0].If, "os.macos")
	}
	if a.Steps[0].PipeTo != "next" {
		t.Errorf("PipeTo = %q, want %q", a.Steps[0].PipeTo, "next")
	}
}

func TestLoad_AutomationWithIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "cond-auto.yaml", `
name: macos-only
description: Only runs on macOS
if: os.macos
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != "os.macos" {
		t.Errorf("If = %q, want %q", a.If, "os.macos")
	}
	if a.Name != "macos-only" {
		t.Errorf("Name = %q, want %q", a.Name, "macos-only")
	}
}

func TestLoad_AutomationWithoutIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-cond-auto.yaml", `
name: always-run
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != "" {
		t.Errorf("If = %q, want empty string", a.If)
	}
}

func TestLoad_AutomationWithComplexIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "complex-auto.yaml", `
name: complex-cond
if: os.macos and not command.brew
steps:
  - bash: echo installing
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != "os.macos and not command.brew" {
		t.Errorf("If = %q, want %q", a.If, "os.macos and not command.brew")
	}
}

func TestLoad_AutomationWithInvalidIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad-auto.yaml", `
name: bad-if
if: "and and and"
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid if expression")
	}
	if !strings.Contains(err.Error(), "invalid if expression") {
		t.Errorf("error should mention 'invalid if expression', got: %v", err)
	}
}

func TestLoad_AutomationWithFuncCallIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "func-auto.yaml", `
name: env-check
if: file.exists(".env")
steps:
  - bash: source .env && echo loaded
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != `file.exists(".env")` {
		t.Errorf("If = %q, want %q", a.If, `file.exists(".env")`)
	}
}

func TestLoad_AutomationIfWithStepIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "both-if.yaml", `
name: both-cond
if: os.macos
steps:
  - bash: brew install jq
    if: not command.jq
  - bash: echo done
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != "os.macos" {
		t.Errorf("automation If = %q, want %q", a.If, "os.macos")
	}
	if a.Steps[0].If != "not command.jq" {
		t.Errorf("step[0] If = %q, want %q", a.Steps[0].If, "not command.jq")
	}
	if a.Steps[1].If != "" {
		t.Errorf("step[1] If = %q, want empty", a.Steps[1].If)
	}
}

// --- Install block tests ---

func TestLoad_InstallScalar(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-brew
description: Install Homebrew

install:
  test: command -v brew >/dev/null 2>&1
  run: /bin/bash -c "$(curl -fsSL https://example.com/install.sh)"
  version: brew --version | head -1
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if len(a.Steps) != 0 {
		t.Errorf("expected no steps, got %d", len(a.Steps))
	}

	inst := a.Install
	if !inst.Test.IsScalar {
		t.Error("expected scalar test phase")
	}
	if inst.Test.Scalar != "command -v brew >/dev/null 2>&1" {
		t.Errorf("test scalar = %q", inst.Test.Scalar)
	}
	if !inst.Run.IsScalar {
		t.Error("expected scalar run phase")
	}
	if inst.HasVerify() {
		t.Error("expected no explicit verify phase")
	}
	if inst.Version != "brew --version | head -1" {
		t.Errorf("version = %q", inst.Version)
	}
}

func TestLoad_InstallStepList(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-python
description: Install Python

install:
  test:
    - bash: python3 --version 2>&1 | grep -q "Python 3.13"
  run:
    - bash: mise install python@3.13
      if: command.mise
    - bash: brew install python@3.13
      if: not command.mise
  verify:
    - bash: python3 --version 2>&1 | grep -q "Python 3.13"
  version: python3 --version | awk '{print $2}'
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}

	inst := a.Install
	if inst.Test.IsScalar {
		t.Error("expected step list for test phase")
	}
	if len(inst.Test.Steps) != 1 {
		t.Fatalf("expected 1 test step, got %d", len(inst.Test.Steps))
	}
	if inst.Test.Steps[0].Type != StepTypeBash {
		t.Errorf("test step type = %q, want bash", inst.Test.Steps[0].Type)
	}

	if inst.Run.IsScalar {
		t.Error("expected step list for run phase")
	}
	if len(inst.Run.Steps) != 2 {
		t.Fatalf("expected 2 run steps, got %d", len(inst.Run.Steps))
	}
	if inst.Run.Steps[0].If != "command.mise" {
		t.Errorf("run step[0] If = %q", inst.Run.Steps[0].If)
	}
	if inst.Run.Steps[1].If != "not command.mise" {
		t.Errorf("run step[1] If = %q", inst.Run.Steps[1].If)
	}

	if !inst.HasVerify() {
		t.Error("expected explicit verify phase")
	}
	if len(inst.Verify.Steps) != 1 {
		t.Fatalf("expected 1 verify step, got %d", len(inst.Verify.Steps))
	}
}

func TestLoad_InstallMixed(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-node
description: Install Node

install:
  test: command -v node >/dev/null 2>&1
  run:
    - bash: brew install node
      if: os.macos
    - bash: apt-get install -y nodejs
      if: os.linux
  version: node --version
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inst := a.Install
	if !inst.Test.IsScalar {
		t.Error("expected scalar test phase")
	}
	if inst.Run.IsScalar {
		t.Error("expected step list for run phase")
	}
	if len(inst.Run.Steps) != 2 {
		t.Fatalf("expected 2 run steps, got %d", len(inst.Run.Steps))
	}
	if inst.HasVerify() {
		t.Error("expected no explicit verify (should default to test)")
	}
}

func TestLoad_InstallAndStepsMutualExclusion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: both
description: Has both

steps:
  - bash: echo hello

install:
  test: command -v foo
  run: install foo
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for both steps and install")
	}
	if !strings.Contains(err.Error(), "cannot have both") {
		t.Errorf("error should mention mutual exclusion, got: %v", err)
	}
}

func TestLoad_InstallEmptyTest(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
description: Empty test

install:
  test: ""
  run: install foo
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty test")
	}
	if !strings.Contains(err.Error(), "install.test must have content") {
		t.Errorf("error = %v", err)
	}
}

func TestLoad_InstallEmptyRun(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
description: Empty run

install:
  test: command -v foo
  run: ""
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty run")
	}
	if !strings.Contains(err.Error(), "install.run must have content") {
		t.Errorf("error = %v", err)
	}
}

func TestLoad_InstallVerifyDefaultsToTest(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-brew
description: Install Homebrew

install:
  test: command -v brew >/dev/null 2>&1
  run: curl install.sh | sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Install.HasVerify() {
		t.Error("expected verify to be nil (defaults to test)")
	}
}

func TestLoad_InstallWithExplicitVerify(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-python
description: Install Python

install:
  test: python3 --version | grep -q "3.13"
  run: mise install python@3.13
  verify: python3 --version | grep -q "3.13"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.Install.HasVerify() {
		t.Error("expected explicit verify phase")
	}
	if !a.Install.Verify.IsScalar {
		t.Error("expected scalar verify phase")
	}
}

func TestLoad_InstallWithIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-brew
description: Install Homebrew
if: os.macos

install:
  test: command -v brew
  run: curl install.sh | sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.If != "os.macos" {
		t.Errorf("If = %q, want %q", a.If, "os.macos")
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
}

func TestLoad_InstallWithInputs(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-python
description: Install Python

inputs:
  version:
    type: string
    required: true

install:
  test: python3 --version | grep -q "Python $PI_INPUT_VERSION"
  run: mise install "python@$PI_INPUT_VERSION"
  version: python3 --version | awk '{print $2}'
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if len(a.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(a.Inputs))
	}
	spec, ok := a.Inputs["version"]
	if !ok {
		t.Fatal("expected 'version' input")
	}
	if !spec.IsRequired() {
		t.Error("expected 'version' to be required")
	}
}

func TestLoad_InstallStepWithInvalidIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad
description: Bad if

install:
  test: command -v foo
  run:
    - bash: install foo
      if: "!invalid"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid if in install step")
	}
	if !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "if") {
		t.Errorf("error = %v", err)
	}
}

func TestLoad_InstallRunStepWithRunRef(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-node
description: Install Node

install:
  test: command -v node
  run:
    - run: pi:install-homebrew
      if: os.macos and not command.brew
    - bash: brew install node
      if: os.macos
  version: node --version
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}

	steps := a.Install.Run.Steps
	if len(steps) != 2 {
		t.Fatalf("expected 2 run steps, got %d", len(steps))
	}
	if steps[0].Type != StepTypeRun {
		t.Errorf("expected run step type, got %q", steps[0].Type)
	}
	if steps[0].Value != "pi:install-homebrew" {
		t.Errorf("expected run step value 'pi:install-homebrew', got %q", steps[0].Value)
	}
}

func TestLoad_InstallNoVersionField(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "inst.yaml", `
name: install-tool
description: Install some tool

install:
  test: command -v tool
  run: install-tool.sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Install.Version != "" {
		t.Errorf("expected empty version, got %q", a.Install.Version)
	}
}

func TestLoadFromBytes_InstallBlock(t *testing.T) {
	yaml := []byte(`
name: install-brew
description: Install Homebrew

install:
  test: command -v brew
  run: curl install.sh | sh
  version: brew --version
`)

	a, err := LoadFromBytes(yaml, "builtin://install-brew")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if !a.Install.Test.IsScalar {
		t.Error("expected scalar test")
	}
	if a.Install.Version != "brew --version" {
		t.Errorf("version = %q", a.Install.Version)
	}
}

// --- requires: block tests ---

func TestLoad_RequiresRuntimeBare(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python
  - node
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "python" || a.Requires[0].Kind != RequirementRuntime || a.Requires[0].MinVersion != "" {
		t.Errorf("req[0] = %+v, want python runtime with no version", a.Requires[0])
	}
	if a.Requires[1].Name != "node" || a.Requires[1].Kind != RequirementRuntime || a.Requires[1].MinVersion != "" {
		t.Errorf("req[1] = %+v, want node runtime with no version", a.Requires[1])
	}
}

func TestLoad_RequiresRuntimeWithVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11
  - node >= 18
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "python" || a.Requires[0].Kind != RequirementRuntime || a.Requires[0].MinVersion != "3.11" {
		t.Errorf("req[0] = %+v, want python >= 3.11", a.Requires[0])
	}
	if a.Requires[1].Name != "node" || a.Requires[1].Kind != RequirementRuntime || a.Requires[1].MinVersion != "18" {
		t.Errorf("req[1] = %+v, want node >= 18", a.Requires[1])
	}
}

func TestLoad_RequiresCommandBare(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - command: docker
  - command: jq
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "docker" || a.Requires[0].Kind != RequirementCommand || a.Requires[0].MinVersion != "" {
		t.Errorf("req[0] = %+v, want command:docker no version", a.Requires[0])
	}
	if a.Requires[1].Name != "jq" || a.Requires[1].Kind != RequirementCommand || a.Requires[1].MinVersion != "" {
		t.Errorf("req[1] = %+v, want command:jq no version", a.Requires[1])
	}
}

func TestLoad_RequiresCommandWithVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - command: kubectl >= 1.28
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "kubectl" || a.Requires[0].Kind != RequirementCommand || a.Requires[0].MinVersion != "1.28" {
		t.Errorf("req[0] = %+v, want command:kubectl >= 1.28", a.Requires[0])
	}
}

func TestLoad_RequiresMixed(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11
  - command: docker
  - command: jq
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Kind != RequirementRuntime {
		t.Errorf("req[0].Kind = %q, want runtime", a.Requires[0].Kind)
	}
	if a.Requires[1].Kind != RequirementCommand {
		t.Errorf("req[1].Kind = %q, want command", a.Requires[1].Kind)
	}
	if a.Requires[2].Kind != RequirementCommand {
		t.Errorf("req[2].Kind = %q, want command", a.Requires[2].Kind)
	}
}

func TestLoad_RequiresThreePartVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
description: Test
requires:
  - python >= 3.11.2
  - command: kubectl >= 1.28.0
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Requires[0].MinVersion != "3.11.2" {
		t.Errorf("req[0].MinVersion = %q, want 3.11.2", a.Requires[0].MinVersion)
	}
	if a.Requires[1].MinVersion != "1.28.0" {
		t.Errorf("req[1].MinVersion = %q, want 1.28.0", a.Requires[1].MinVersion)
	}
}

func TestLoad_RequiresNoBlock(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 0 {
		t.Errorf("expected 0 requirements, got %d", len(a.Requires))
	}
}

func TestLoad_RequiresOnInstaller(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: install-tool
description: Install tool
requires:
  - command: curl
install:
  test: command -v tool
  run: curl install.sh | sh
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if len(a.Requires) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "curl" || a.Requires[0].Kind != RequirementCommand {
		t.Errorf("req[0] = %+v, want command:curl", a.Requires[0])
	}
}

func TestLoad_RequiresUnknownRuntime(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - ruby >= 3.0
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown runtime")
	}
	if !strings.Contains(err.Error(), "unknown runtime") {
		t.Errorf("expected 'unknown runtime' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "command:") {
		t.Errorf("expected hint about command:, got: %v", err)
	}
}

func TestLoad_RequiresEmptyEntry(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - ""
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty requires entry")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestLoad_RequiresEmptyCommand(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - command: ""
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty command value")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestLoad_RequiresBadVersionSyntax(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >= abc
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad version syntax")
	}
	if !strings.Contains(err.Error(), "non-numeric") {
		t.Errorf("expected non-numeric error, got: %v", err)
	}
}

func TestLoad_RequiresMissingVersionAfterGte(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >=
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing version after >=")
	}
	if !strings.Contains(err.Error(), "missing version") {
		t.Errorf("expected 'missing version' error, got: %v", err)
	}
}

func TestLoad_RequiresInvalidMappingKey(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - runtime: python
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown mapping key")
	}
	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("expected 'unknown key' error, got: %v", err)
	}
}

func TestLoad_RequiresVersionWithEmptyComponent(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python >= 3..11
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for version with empty component")
	}
	if !strings.Contains(err.Error(), "empty component") {
		t.Errorf("expected 'empty component' error, got: %v", err)
	}
}

func TestLoad_RequiresScalarWithSpaces(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - python something
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' error, got: %v", err)
	}
}

func TestLoad_RequiresCommandVersionBadSyntax(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "req.yaml", `
name: test
requires:
  - command: kubectl >= x.y
steps:
  - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad command version")
	}
	if !strings.Contains(err.Error(), "non-numeric") {
		t.Errorf("expected non-numeric error, got: %v", err)
	}
}

func TestLoadFromBytes_RequiresBlock(t *testing.T) {
	data := []byte(`
name: test
description: Test
requires:
  - node >= 18
  - command: docker
steps:
  - bash: echo hello
`)

	a, err := LoadFromBytes(data, "builtin://test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Requires) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(a.Requires))
	}
	if a.Requires[0].Name != "node" || a.Requires[0].MinVersion != "18" {
		t.Errorf("req[0] = %+v, want node >= 18", a.Requires[0])
	}
	if a.Requires[1].Name != "docker" || a.Requires[1].Kind != RequirementCommand {
		t.Errorf("req[1] = %+v, want command:docker", a.Requires[1])
	}
}

func TestParseNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{"docker", "docker", "", false},
		{"kubectl >= 1.28", "kubectl", "1.28", false},
		{"python >= 3.11.2", "python", "3.11.2", false},
		{"node >= 18", "node", "18", false},
		{">= 1.0", "", "", true},
		{"python >=", "", "", true},
		{"python >= abc", "", "", true},
		{"python >= 3..1", "", "", true},
		{"python something", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version, err := parseNameVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNameVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if name != tt.wantName {
					t.Errorf("name = %q, want %q", name, tt.wantName)
				}
				if version != tt.wantVersion {
					t.Errorf("version = %q, want %q", version, tt.wantVersion)
				}
			}
		})
	}
}

func TestStep_SilentField(t *testing.T) {
	yaml := `name: test
description: Test silent field
steps:
  - bash: echo visible
  - bash: echo hidden
    silent: true
  - bash: echo also visible
`
	a, err := LoadFromBytes([]byte(yaml), "/fake/test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(a.Steps))
	}
	if a.Steps[0].Silent {
		t.Error("step 0 should not be silent")
	}
	if !a.Steps[1].Silent {
		t.Error("step 1 should be silent")
	}
	if a.Steps[2].Silent {
		t.Error("step 2 should not be silent")
	}
}

func TestStep_SilentFalseExplicit(t *testing.T) {
	yaml := `name: test
description: Test explicit silent false
steps:
  - bash: echo hello
    silent: false
`
	a, err := LoadFromBytes([]byte(yaml), "/fake/test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Silent {
		t.Error("step with silent: false should not be silent")
	}
}

func TestLoad_ParentShellStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "activate.yaml", `
name: activate-venv
description: Activate virtualenv in parent shell

steps:
  - bash: source venv/bin/activate
    parent_shell: true
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if !a.Steps[0].ParentShell {
		t.Error("expected parent_shell to be true")
	}
	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step type = %q, want bash", a.Steps[0].Type)
	}
}

func TestLoad_ParentShellOnNonBashStep_Error(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad-parent-shell
description: Invalid parent_shell on python

steps:
  - python: print("hello")
    parent_shell: true
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for parent_shell on non-bash step")
	}
	if !strings.Contains(err.Error(), "parent_shell") {
		t.Errorf("error should mention parent_shell, got: %v", err)
	}
}

func TestLoad_ParentShellWithPipeTo_Error(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad.yaml", `
name: bad-pipe-parent
description: Invalid parent_shell with pipe_to

steps:
  - bash: echo test
    parent_shell: true
    pipe_to: next
  - bash: cat
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for parent_shell with pipe_to")
	}
	if !strings.Contains(err.Error(), "pipe_to") {
		t.Errorf("error should mention pipe_to, got: %v", err)
	}
}

func TestLoad_ParentShellFalse(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "normal.yaml", `
name: normal
description: Normal step

steps:
  - bash: echo hello
    parent_shell: false
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Steps[0].ParentShell {
		t.Error("expected parent_shell to be false")
	}
}

func TestLoad_StepWithDir(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "dir-step.yaml", `
name: build-in-subdir
steps:
  - bash: go build ./...
    dir: src
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if a.Steps[0].Dir != "src" {
		t.Errorf("dir = %q, want %q", a.Steps[0].Dir, "src")
	}
}

func TestLoad_StepWithoutDir(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-dir.yaml", `
name: plain
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Dir != "" {
		t.Errorf("dir should be empty, got %q", a.Steps[0].Dir)
	}
}

func TestLoad_StepDirWithOtherFields(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "dir-combo.yaml", `
name: combo
steps:
  - bash: go test ./...
    dir: src
    env:
      GOFLAGS: -race
    silent: true
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Dir != "src" {
		t.Errorf("dir = %q, want %q", a.Steps[0].Dir, "src")
	}
	if !a.Steps[0].Silent {
		t.Error("expected silent = true")
	}
	if a.Steps[0].Env["GOFLAGS"] != "-race" {
		t.Errorf("env[GOFLAGS] = %q, want %q", a.Steps[0].Env["GOFLAGS"], "-race")
	}
}

func TestLoad_StepWithTimeout(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "timeout.yaml", `
name: build-with-timeout
steps:
  - bash: go build ./...
    timeout: 30s
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(a.Steps))
	}
	if a.Steps[0].Timeout.Seconds() != 30 {
		t.Errorf("timeout = %v, want 30s", a.Steps[0].Timeout)
	}
	if a.Steps[0].TimeoutRaw != "30s" {
		t.Errorf("timeoutRaw = %q, want %q", a.Steps[0].TimeoutRaw, "30s")
	}
}

func TestLoad_StepWithTimeout_Minutes(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "timeout-min.yaml", `
name: long-build
steps:
  - bash: make all
    timeout: 5m
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Minutes() != 5 {
		t.Errorf("timeout = %v, want 5m", a.Steps[0].Timeout)
	}
}

func TestLoad_StepWithTimeout_ComplexDuration(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "timeout-complex.yaml", `
name: complex-timeout
steps:
  - bash: long-running-script.sh
    timeout: 1h30m
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Hours() != 1.5 {
		t.Errorf("timeout = %v, want 1h30m", a.Steps[0].Timeout)
	}
}

func TestLoad_StepWithoutTimeout(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "no-timeout.yaml", `
name: plain
steps:
  - bash: echo hello
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout != 0 {
		t.Errorf("timeout should be zero, got %v", a.Steps[0].Timeout)
	}
	if a.Steps[0].TimeoutRaw != "" {
		t.Errorf("timeoutRaw should be empty, got %q", a.Steps[0].TimeoutRaw)
	}
}

func TestLoad_StepTimeout_InvalidDuration(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "bad-timeout.yaml", `
name: bad
steps:
  - bash: echo hello
    timeout: not-a-duration
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid timeout duration")
	}
	if !strings.Contains(err.Error(), "invalid timeout") {
		t.Errorf("error = %q, expected to contain 'invalid timeout'", err.Error())
	}
}

func TestLoad_StepTimeout_NegativeDuration(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "neg-timeout.yaml", `
name: neg
steps:
  - bash: echo hello
    timeout: "-5s"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
	if !strings.Contains(err.Error(), "timeout must be positive") {
		t.Errorf("error = %q, expected to contain 'timeout must be positive'", err.Error())
	}
}

func TestLoad_StepTimeout_ZeroDuration(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "zero-timeout.yaml", `
name: zero
steps:
  - bash: echo hello
    timeout: "0s"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
	if !strings.Contains(err.Error(), "timeout must be positive") {
		t.Errorf("error = %q, expected to contain 'timeout must be positive'", err.Error())
	}
}

func TestLoad_StepTimeout_OnRunStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "run-timeout.yaml", `
name: run-with-timeout
steps:
  - run: other/automation
    timeout: 30s
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for timeout on run step")
	}
	if !strings.Contains(err.Error(), "'timeout' is not valid on 'run' steps") {
		t.Errorf("error = %q, expected to mention timeout not valid on run steps", err.Error())
	}
}

func TestLoad_StepTimeout_OnParentShell(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "parent-timeout.yaml", `
name: parent-with-timeout
steps:
  - bash: source venv/bin/activate
    parent_shell: true
    timeout: 30s
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for timeout on parent_shell step")
	}
	if !strings.Contains(err.Error(), "'timeout' cannot be combined with 'parent_shell'") {
		t.Errorf("error = %q, expected to mention timeout+parent_shell", err.Error())
	}
}

func TestLoad_StepTimeout_WithOtherFields(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "timeout-combo.yaml", `
name: combo
steps:
  - bash: go test ./...
    timeout: 2m
    dir: src
    env:
      GOFLAGS: -race
    silent: true
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Minutes() != 2 {
		t.Errorf("timeout = %v, want 2m", a.Steps[0].Timeout)
	}
	if a.Steps[0].Dir != "src" {
		t.Errorf("dir = %q, want %q", a.Steps[0].Dir, "src")
	}
	if !a.Steps[0].Silent {
		t.Error("expected silent = true")
	}
}

func TestLoad_StepTimeout_PythonStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "py-timeout.yaml", `
name: py-timeout
steps:
  - python: print("hello")
    timeout: 10s
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Seconds() != 10 {
		t.Errorf("timeout = %v, want 10s", a.Steps[0].Timeout)
	}
}

func TestLoad_StepTimeout_TypeScriptStep(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "ts-timeout.yaml", `
name: ts-timeout
steps:
  - typescript: console.log("hello")
    timeout: 15s
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Seconds() != 15 {
		t.Errorf("timeout = %v, want 15s", a.Steps[0].Timeout)
	}
}

func TestLoad_StepTimeout_WithPipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "pipe-timeout.yaml", `
name: pipe-timeout
steps:
  - bash: echo hello
    pipe_to: next
    timeout: 5s
  - bash: cat
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Timeout.Seconds() != 5 {
		t.Errorf("timeout = %v, want 5s", a.Steps[0].Timeout)
	}
	if a.Steps[0].PipeTo != "next" {
		t.Errorf("pipe_to = %q, want %q", a.Steps[0].PipeTo, "next")
	}
}

func TestLoad_StepWithDescription(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "desc.yaml", `
name: with-desc
description: Automation with step descriptions
steps:
  - bash: docker-compose up -d
    description: Start all containers in the background
  - bash: sleep 2
  - python: check_health.py
    description: Verify services are healthy
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(a.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(a.Steps))
	}

	if a.Steps[0].Description != "Start all containers in the background" {
		t.Errorf("step[0].Description = %q, want %q", a.Steps[0].Description, "Start all containers in the background")
	}
	if a.Steps[1].Description != "" {
		t.Errorf("step[1].Description = %q, want empty", a.Steps[1].Description)
	}
	if a.Steps[2].Description != "Verify services are healthy" {
		t.Errorf("step[2].Description = %q, want %q", a.Steps[2].Description, "Verify services are healthy")
	}
}

func TestLoad_StepDescriptionWithOtherFields(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "desc-combo.yaml", `
name: desc-combo
steps:
  - bash: go test ./...
    description: Run tests in the API directory
    dir: services/api
    timeout: 5m
    silent: true
    env:
      GO_TEST_FLAGS: -v
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	step := a.Steps[0]
	if step.Description != "Run tests in the API directory" {
		t.Errorf("description = %q, want %q", step.Description, "Run tests in the API directory")
	}
	if step.Dir != "services/api" {
		t.Errorf("dir = %q, want %q", step.Dir, "services/api")
	}
	if step.Timeout.Minutes() != 5 {
		t.Errorf("timeout = %v, want 5m", step.Timeout)
	}
	if !step.Silent {
		t.Error("expected silent = true")
	}
	if step.Env["GO_TEST_FLAGS"] != "-v" {
		t.Errorf("env GO_TEST_FLAGS = %q, want %q", step.Env["GO_TEST_FLAGS"], "-v")
	}
}

func TestValidateVersionString(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"3", false},
		{"3.11", false},
		{"3.11.2", false},
		{"18", false},
		{"1.28.0", false},
		{"abc", true},
		{"3..11", true},
		{"3.11.", true},
		{".3.11", true},
		{"3.11a", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := validateVersionString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVersionString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
