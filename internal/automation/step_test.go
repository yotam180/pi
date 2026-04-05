package automation

import (
	"strings"
	"testing"
)

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

// --- silent, parent_shell, dir, timeout, description ---

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

// --- first: block tests ---

func TestLoad_FirstBlock_Basic(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first.yaml", `
description: Use first matching installer
steps:
  - first:
      - bash: echo "using mise"
        if: command.mise
      - bash: echo "using brew"
        if: command.brew
      - bash: echo "fallback"
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(a.Steps))
	}
	step := a.Steps[0]
	if !step.IsFirst() {
		t.Fatal("expected step to be a first: block")
	}
	if len(step.First) != 3 {
		t.Fatalf("expected 3 sub-steps, got %d", len(step.First))
	}
	if step.First[0].If != "command.mise" {
		t.Errorf("first[0].If = %q, want %q", step.First[0].If, "command.mise")
	}
	if step.First[1].If != "command.brew" {
		t.Errorf("first[1].If = %q, want %q", step.First[1].If, "command.brew")
	}
	if step.First[2].If != "" {
		t.Errorf("first[2].If = %q, want empty (fallback)", step.First[2].If)
	}
}

func TestLoad_FirstBlock_WithDescription(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-desc.yaml", `
description: Test
steps:
  - first:
      - bash: echo hello
    description: Pick the right installer
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].Description != "Pick the right installer" {
		t.Errorf("description = %q, want %q", a.Steps[0].Description, "Pick the right installer")
	}
}

func TestLoad_FirstBlock_WithPipeTo(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-pipe.yaml", `
description: Piped first
steps:
  - first:
      - bash: echo data
    pipe_to: next
  - bash: cat
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].PipeTo != "next" {
		t.Errorf("PipeTo = %q, want %q", a.Steps[0].PipeTo, "next")
	}
}

func TestLoad_FirstBlock_WithOuterIf(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-outerif.yaml", `
description: Conditional first block
steps:
  - first:
      - bash: echo hello
    if: os.macos
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].If != "os.macos" {
		t.Errorf("step.If = %q, want %q", a.Steps[0].If, "os.macos")
	}
}

func TestLoad_FirstBlock_EmptyError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-empty.yaml", `
description: Empty first block
steps:
  - first: []
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty first: block")
	}
	if !strings.Contains(err.Error(), "must contain at least one sub-step") {
		t.Errorf("error = %q, want mention of empty sub-steps", err.Error())
	}
}

func TestLoad_FirstBlock_WithStepTypeKeyError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-conflict.yaml", `
description: Conflict
steps:
  - first:
      - bash: echo hello
    bash: echo world
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for first: combined with bash:")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("error = %q, want mention of conflict", err.Error())
	}
}

func TestLoad_FirstBlock_EnvOnBlockError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-env.yaml", `
description: Env on block
steps:
  - first:
      - bash: echo hello
    env:
      FOO: bar
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for env on first: block")
	}
	if !strings.Contains(err.Error(), "env") && !strings.Contains(err.Error(), "first") {
		t.Errorf("error = %q, want mention of env on first", err.Error())
	}
}

func TestLoad_FirstBlock_DirOnBlockError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-dir.yaml", `
description: Dir on block
steps:
  - first:
      - bash: echo hello
    dir: some/dir
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for dir on first: block")
	}
	if !strings.Contains(err.Error(), "dir") {
		t.Errorf("error = %q, want mention of dir on first", err.Error())
	}
}

func TestLoad_FirstBlock_SubStepEnvOK(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-sub-env.yaml", `
description: Sub-step env
steps:
  - first:
      - bash: echo hello
        env:
          FOO: bar
        if: os.macos
      - bash: echo world
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Steps[0].First[0].Env["FOO"] != "bar" {
		t.Errorf("sub-step env FOO = %q, want %q", a.Steps[0].First[0].Env["FOO"], "bar")
	}
}

func TestLoad_FirstBlock_InInstallPhase(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-install.yaml", `
description: Installer with first block
install:
  test: command -v go
  run:
    - first:
        - bash: mise install go
          if: command.mise
        - bash: brew install go
          if: command.brew
        - bash: echo "no installer" && exit 1
  version: go version | awk '{print $3}'
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.IsInstaller() {
		t.Fatal("expected installer automation")
	}
	if len(a.Install.Run.Steps) != 1 {
		t.Fatalf("expected 1 run step, got %d", len(a.Install.Run.Steps))
	}
	runStep := a.Install.Run.Steps[0]
	if !runStep.IsFirst() {
		t.Fatal("expected first: block in install.run")
	}
	if len(runStep.First) != 3 {
		t.Fatalf("expected 3 sub-steps, got %d", len(runStep.First))
	}
}

func TestLoad_FirstBlock_NestedFirstError(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-nested.yaml", `
description: Nested first
steps:
  - first:
      - first:
          - bash: echo hello
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for nested first: blocks")
	}
}

func TestLoad_FirstBlock_MixedWithRegularSteps(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "first-mixed.yaml", `
description: Mixed steps
steps:
  - bash: echo before
  - first:
      - bash: echo first-a
        if: os.macos
      - bash: echo first-b
  - bash: echo after
`)

	a, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(a.Steps))
	}
	if a.Steps[0].Type != StepTypeBash {
		t.Errorf("step[0] should be bash, got %s", a.Steps[0].Type)
	}
	if !a.Steps[1].IsFirst() {
		t.Error("step[1] should be a first: block")
	}
	if a.Steps[2].Type != StepTypeBash {
		t.Errorf("step[2] should be bash, got %s", a.Steps[2].Type)
	}
}

