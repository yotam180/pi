package executor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/reqcheck"
	"github.com/vyper-tooling/pi/internal/runtimes"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Python 3.13.0", "3.13.0"},
		{"Python 3.9.7", "3.9.7"},
		{"v20.11.0", "20.11.0"},
		{"node v22.0.0", "22.0.0"},
		{"jq-1.7.1", "1.7.1"},
		{"docker version 24.0.5, build ced0996", "24.0.5"},
		{"kubectl v1.28.3", "1.28.3"},
		{"Go 1.22.0", "1.22.0"},
		{"tsx v4.19.0", "4.19.0"},
		{"no version here", ""},
		{"", ""},
		{"just 42", ""},          // single number without dot isn't matched
		{"version 3.13", "3.13"}, // two components
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := reqcheck.ExtractVersion(tt.input)
			if got != tt.want {
				t.Errorf("ExtractVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"3.13.0", "3.11", 1},
		{"3.9.7", "3.11", -1},
		{"3.11.0", "3.11", 0},
		{"3.11", "3.11.0", 0},
		{"20.11.0", "20.11.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.28.3", "1.28", 1},
		{"1.28", "1.28.3", -1},
		{"22.0.0", "20.0.0", 1},
		{"3.10", "3.9", 1},
		{"3.9", "3.10", -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.a, tt.b), func(t *testing.T) {
			got, err := reqcheck.CompareVersions(tt.a, tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareVersions_Invalid(t *testing.T) {
	_, err := reqcheck.CompareVersions("abc", "1.0")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}

	_, err = reqcheck.CompareVersions("1.0", "abc")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}
}

func TestRuntimeCommand(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"python", "python3"},
		{"node", "node"},
		{"rust", "rustc"},
		{"go", "go"},
	}
	for _, tt := range tests {
		if got := reqcheck.RuntimeCommand(tt.name); got != tt.want {
			t.Errorf("RuntimeCommand(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func mockEnv(lookPath map[string]bool, versionOutput map[string]string) *conditions.RuntimeEnv {
	return &conditions.RuntimeEnv{
		GOOS:   "darwin",
		GOARCH: "arm64",
		Getenv: func(s string) string { return "" },
		LookPath: func(name string) (string, error) {
			if lookPath[name] {
				return "/usr/bin/" + name, nil
			}
			return "", fmt.Errorf("not found: %s", name)
		},
		Stat: func(name string) (os.FileInfo, error) {
			return nil, fmt.Errorf("not found")
		},
		ExecOutput: func(cmd string, args ...string) string {
			if out, ok := versionOutput[cmd]; ok {
				return out
			}
			return ""
		},
	}
}

func TestCheckRequirement_CommandFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{"docker": true},
		nil,
	)
	req := automation.Requirement{Name: "docker", Kind: automation.RequirementCommand}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_CommandNotFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{},
		nil,
	)
	req := automation.Requirement{Name: "kubectl", Kind: automation.RequirementCommand}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for missing command")
	}
	if result.Error != "not found" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckRequirement_RuntimeFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{"python3": true},
		nil,
	)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_RuntimeNotFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{},
		nil,
	)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for missing runtime")
	}
	if result.Error != "not found" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckRequirement_VersionSatisfied(t *testing.T) {
	env := mockEnv(
		map[string]bool{"python3": true},
		map[string]string{"python3": "Python 3.13.0"},
	)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "3.13.0" {
		t.Errorf("expected detected version 3.13.0, got %q", result.DetectedVersion)
	}
}

func TestCheckRequirement_VersionTooLow(t *testing.T) {
	env := mockEnv(
		map[string]bool{"python3": true},
		map[string]string{"python3": "Python 3.9.7"},
	)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for version too low")
	}
	if !strings.Contains(result.Error, "3.9.7") || !strings.Contains(result.Error, "3.11") {
		t.Errorf("error should mention both versions, got: %s", result.Error)
	}
}

func TestCheckRequirement_RustRuntimeFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{"rustc": true},
		nil,
	)
	req := automation.Requirement{Name: "rust", Kind: automation.RequirementRuntime}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_RustRuntimeNotFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{},
		nil,
	)
	req := automation.Requirement{Name: "rust", Kind: automation.RequirementRuntime}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for missing rust runtime")
	}
	if result.Error != "not found" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckRequirement_RustVersionSatisfied(t *testing.T) {
	env := mockEnv(
		map[string]bool{"rustc": true},
		map[string]string{"rustc": "rustc 1.80.0 (051478957 2024-07-21)"},
	)
	req := automation.Requirement{Name: "rust", Kind: automation.RequirementRuntime, MinVersion: "1.75"}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "1.80.0" {
		t.Errorf("expected detected version 1.80.0, got %q", result.DetectedVersion)
	}
}

func TestCheckRequirement_GoRuntimeFound(t *testing.T) {
	env := mockEnv(
		map[string]bool{"go": true},
		nil,
	)
	req := automation.Requirement{Name: "go", Kind: automation.RequirementRuntime}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_GoVersionSatisfied(t *testing.T) {
	env := mockEnv(
		map[string]bool{"go": true},
		map[string]string{"go": "go version go1.23.0 darwin/arm64"},
	)
	req := automation.Requirement{Name: "go", Kind: automation.RequirementRuntime, MinVersion: "1.22"}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "1.23.0" {
		t.Errorf("expected detected version 1.23.0, got %q", result.DetectedVersion)
	}
}

func TestCheckRequirement_CommandWithVersion(t *testing.T) {
	env := mockEnv(
		map[string]bool{"kubectl": true},
		map[string]string{"kubectl": "kubectl v1.28.3"},
	)
	req := automation.Requirement{Name: "kubectl", Kind: automation.RequirementCommand, MinVersion: "1.28"}
	result := reqcheck.CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_CommandVersionTooLow(t *testing.T) {
	env := mockEnv(
		map[string]bool{"kubectl": true},
		map[string]string{"kubectl": "kubectl v1.27.0"},
	)
	req := automation.Requirement{Name: "kubectl", Kind: automation.RequirementCommand, MinVersion: "1.28"}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for version too low")
	}
}

func TestCheckRequirement_VersionUndetectable(t *testing.T) {
	env := mockEnv(
		map[string]bool{"sometool": true},
		map[string]string{"sometool": "no version info here"},
	)
	req := automation.Requirement{Name: "sometool", Kind: automation.RequirementCommand, MinVersion: "1.0"}
	result := reqcheck.CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied when version is undetectable")
	}
	if !strings.Contains(result.Error, "unable to detect version") {
		t.Errorf("expected 'unable to detect version' error, got: %s", result.Error)
	}
}

func TestValidateRequirements_AllSatisfied(t *testing.T) {
	dir := t.TempDir()
	env := mockEnv(
		map[string]bool{"python3": true, "docker": true},
		map[string]string{"python3": "Python 3.13.0"},
	)

	a := &automation.Automation{
		Name:     "test-automation",
		FilePath: "/fake/path/automation.yaml",
		Steps:    []automation.Step{bashStep("echo ok")},
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"},
			{Name: "docker", Kind: automation.RequirementCommand},
		},
	}

	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  newDiscovery(nil),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		RuntimeEnv: env,
	}

	err := exec.ValidateRequirements(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRequirements_SomeMissing(t *testing.T) {
	dir := t.TempDir()
	env := mockEnv(
		map[string]bool{"python3": true},
		map[string]string{"python3": "Python 3.9.7"},
	)

	a := &automation.Automation{
		Name:     "format-logs",
		FilePath: "/fake/path/automation.yaml",
		Steps:    []automation.Step{bashStep("echo ok")},
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"},
			{Name: "jq", Kind: automation.RequirementCommand},
		},
	}

	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  newDiscovery(nil),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		RuntimeEnv: env,
	}

	err := exec.ValidateRequirements(a)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var ve *reqcheck.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *reqcheck.ValidationError, got %T", err)
	}

	if len(ve.Results) != 2 {
		t.Fatalf("expected 2 failed results, got %d", len(ve.Results))
	}

	if ve.Results[0].Requirement.Name != "python" {
		t.Errorf("first result should be python, got %s", ve.Results[0].Requirement.Name)
	}
	if ve.Results[1].Requirement.Name != "jq" {
		t.Errorf("second result should be jq, got %s", ve.Results[1].Requirement.Name)
	}
}

func TestValidateRequirements_NoRequirements(t *testing.T) {
	dir := t.TempDir()
	a := &automation.Automation{
		Name:     "test",
		FilePath: "/fake/path/automation.yaml",
		Steps:    []automation.Step{bashStep("echo ok")},
	}

	exec := &Executor{
		RepoRoot:  dir,
		Discovery: newDiscovery(nil),
		Stdout:    &bytes.Buffer{},
		Stderr:    &bytes.Buffer{},
	}

	err := exec.ValidateRequirements(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatValidationError(t *testing.T) {
	ve := &reqcheck.ValidationError{
		AutomationName: "format-logs",
		Results: []reqcheck.CheckResult{
			{
				Requirement:     automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"},
				Satisfied:       false,
				DetectedVersion: "3.9.7",
				Error:           "found 3.9.7 but need >= 3.11",
			},
			{
				Requirement: automation.Requirement{Name: "jq", Kind: automation.RequirementCommand},
				Satisfied:   false,
				Error:       "not found",
			},
		},
	}

	output := reqcheck.FormatValidationError(ve)

	if !strings.Contains(output, "✗ pi run format-logs") {
		t.Errorf("output should contain automation name, got:\n%s", output)
	}
	if !strings.Contains(output, "Missing requirements:") {
		t.Errorf("output should contain 'Missing requirements:', got:\n%s", output)
	}
	if !strings.Contains(output, "python >= 3.11") {
		t.Errorf("output should contain 'python >= 3.11', got:\n%s", output)
	}
	if !strings.Contains(output, "found 3.9.7 but need >= 3.11") {
		t.Errorf("output should contain version comparison, got:\n%s", output)
	}
	if !strings.Contains(output, "command: jq") {
		t.Errorf("output should contain 'command: jq', got:\n%s", output)
	}
	if !strings.Contains(output, "brew install jq") {
		t.Errorf("output should contain install hint for jq, got:\n%s", output)
	}
	if !strings.Contains(output, "brew install python3") {
		t.Errorf("output should contain install hint for python, got:\n%s", output)
	}
}

func TestInstallHint(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"python", "brew install python3"},
		{"node", "brew install node"},
		{"docker", "brew install --cask docker"},
		{"jq", "brew install jq"},
		{"rust", "rustup"},
		{"go", "brew install go"},
		{"unknown-tool", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := automation.Requirement{Name: tt.name}
			got := reqcheck.InstallHintFor(req)
			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Errorf("InstallHintFor(%q) = %q, want to contain %q", tt.name, got, tt.want)
			}
			if tt.want == "" && got != "" {
				t.Errorf("InstallHintFor(%q) = %q, want empty", tt.name, got)
			}
		})
	}
}

func TestRunWithInputs_FailsOnMissingRequirement(t *testing.T) {
	dir := t.TempDir()
	env := mockEnv(map[string]bool{}, nil)

	a := &automation.Automation{
		Name:     "needs-python",
		FilePath: "/fake/path/automation.yaml",
		Steps:    []automation.Step{bashStep("echo ok")},
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime},
		},
	}

	var stderr bytes.Buffer
	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  &discovery.Result{Automations: map[string]*automation.Automation{"needs-python": a}},
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
		RuntimeEnv: env,
	}

	err := exec.RunWithInputs(a, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing requirement")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}

	if !strings.Contains(stderr.String(), "Missing requirements:") {
		t.Errorf("stderr should contain 'Missing requirements:', got:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "python") {
		t.Errorf("stderr should mention python, got:\n%s", stderr.String())
	}
}

func TestRunWithInputs_PassesOnSatisfiedRequirement(t *testing.T) {
	dir := t.TempDir()
	env := mockEnv(
		map[string]bool{"python3": true},
		nil,
	)

	a := &automation.Automation{
		Name:     "has-python",
		FilePath: "/fake/path/automation.yaml",
		Steps:    []automation.Step{bashStep("true")},
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime},
		},
	}

	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  &discovery.Result{Automations: map[string]*automation.Automation{"has-python": a}},
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		RuntimeEnv: env,
	}

	err := exec.RunWithInputs(a, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunWithInputs_InstallerWithMissingRequirement(t *testing.T) {
	dir := t.TempDir()
	env := mockEnv(map[string]bool{}, nil)

	a := &automation.Automation{
		Name:     "install-tool",
		FilePath: "/fake/path/automation.yaml",
		Install: &automation.InstallSpec{
			Test: automation.InstallPhase{IsScalar: true, Scalar: "command -v tool"},
			Run:  automation.InstallPhase{IsScalar: true, Scalar: "echo installing"},
		},
		Requires: []automation.Requirement{
			{Name: "docker", Kind: automation.RequirementCommand},
		},
	}

	var stderr bytes.Buffer
	exec := &Executor{
		RepoRoot:   dir,
		Discovery:  &discovery.Result{Automations: map[string]*automation.Automation{"install-tool": a}},
		Stdout:     &bytes.Buffer{},
		Stderr:     &stderr,
		RuntimeEnv: env,
	}

	err := exec.RunWithInputs(a, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing requirement on installer")
	}

	if !strings.Contains(stderr.String(), "Missing requirements:") {
		t.Errorf("stderr should contain validation error, got:\n%s", stderr.String())
	}
}

func TestCheckRequirementForDoctor_DetectsVersionWithoutConstraint(t *testing.T) {
	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
		ExecOutput: func(cmd string, args ...string) string {
			if cmd == "bash" {
				return "GNU bash, version 5.2.37(1)-release"
			}
			return ""
		},
	}

	req := automation.Requirement{Name: "bash", Kind: automation.RequirementCommand}
	result := reqcheck.CheckRequirementForDoctor(req, env)

	if !result.Satisfied {
		t.Fatalf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "5.2.37" {
		t.Errorf("expected version 5.2.37, got %q", result.DetectedVersion)
	}
}

func TestCheckRequirementForDoctor_NotFound(t *testing.T) {
	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "", fmt.Errorf("not found") },
	}

	req := automation.Requirement{Name: "missing", Kind: automation.RequirementCommand}
	result := reqcheck.CheckRequirementForDoctor(req, env)

	if result.Satisfied {
		t.Fatal("expected not satisfied for missing command")
	}
	if result.Error != "not found" {
		t.Errorf("expected 'not found' error, got %q", result.Error)
	}
}

func TestCheckRequirementForDoctor_VersionConstraint(t *testing.T) {
	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "/usr/bin/" + file, nil },
		ExecOutput: func(cmd string, args ...string) string {
			return "Python 3.13.0"
		},
	}

	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"}
	result := reqcheck.CheckRequirementForDoctor(req, env)

	if !result.Satisfied {
		t.Fatalf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "3.13.0" {
		t.Errorf("expected version 3.13.0, got %q", result.DetectedVersion)
	}
}

func TestInstallHintFor_KnownTool(t *testing.T) {
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	hint := reqcheck.InstallHintFor(req)
	if hint == "" {
		t.Error("expected install hint for python")
	}
}

func TestInstallHintFor_UnknownTool(t *testing.T) {
	req := automation.Requirement{Name: "unknown-tool-xyz", Kind: automation.RequirementCommand}
	hint := reqcheck.InstallHintFor(req)
	if hint != "" {
		t.Errorf("expected empty hint for unknown tool, got %q", hint)
	}
}

func TestValidateRequirements_WithProvisioner_RuntimeProvisioned(t *testing.T) {
	base := t.TempDir()
	binDir := fmt.Sprintf("%s/python/3.11/bin", base)
	os.MkdirAll(binDir, 0755)
	os.WriteFile(fmt.Sprintf("%s/python3", binDir), []byte("#!/bin/sh"), 0755)

	prov := &runtimes.Provisioner{
		Mode:    "auto",
		Manager: "direct",
		BaseDir: base,
		Stderr:  &bytes.Buffer{},
	}

	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "", fmt.Errorf("not found") },
	}

	a := &automation.Automation{
		Name: "test",
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"},
		},
	}

	exec := &Executor{
		RepoRoot:    t.TempDir(),
		Discovery:   discovery.NewResult(nil, nil),
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		RuntimeEnv:  env,
		Provisioner: prov,
	}

	// Python not in PATH, but already provisioned locally
	err := exec.ValidateRequirements(a)
	if err != nil {
		t.Fatalf("expected provisioned runtime to satisfy requirement, got: %v", err)
	}

	if len(exec.runtimePaths) != 1 {
		t.Fatalf("expected 1 runtime path, got %d", len(exec.runtimePaths))
	}
	if exec.runtimePaths[0] != binDir {
		t.Errorf("runtime path = %q, want %q", exec.runtimePaths[0], binDir)
	}
}

func TestValidateRequirements_WithProvisioner_CommandNotProvisioned(t *testing.T) {
	prov := &runtimes.Provisioner{
		Mode:    "auto",
		Manager: "direct",
		BaseDir: t.TempDir(),
		Stderr:  &bytes.Buffer{},
	}

	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "", fmt.Errorf("not found") },
	}

	a := &automation.Automation{
		Name: "test",
		Requires: []automation.Requirement{
			{Name: "docker", Kind: automation.RequirementCommand},
		},
	}

	exec := &Executor{
		RepoRoot:    t.TempDir(),
		Discovery:   discovery.NewResult(nil, nil),
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
		RuntimeEnv:  env,
		Provisioner: prov,
	}

	err := exec.ValidateRequirements(a)
	if err == nil {
		t.Fatal("expected error for missing command requirement (commands are not provisioned)")
	}
}

func TestValidateRequirements_NoProvisioner_FallsThrough(t *testing.T) {
	env := &conditions.RuntimeEnv{
		LookPath: func(file string) (string, error) { return "", fmt.Errorf("not found") },
	}

	a := &automation.Automation{
		Name: "test",
		Requires: []automation.Requirement{
			{Name: "python", Kind: automation.RequirementRuntime},
		},
	}

	exec := &Executor{
		RepoRoot:   t.TempDir(),
		Discovery:  discovery.NewResult(nil, nil),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		RuntimeEnv: env,
	}

	err := exec.ValidateRequirements(a)
	if err == nil {
		t.Fatal("expected error when no provisioner and runtime not found")
	}
}

func TestBuildStepEnv_NoRuntimePaths(t *testing.T) {
	env := BuildStepEnv(nil, nil, nil, nil)
	if env != nil {
		t.Error("expected nil env when no input vars and no runtime paths")
	}
}

func TestBuildStepEnv_WithRuntimePaths(t *testing.T) {
	env := BuildStepEnv([]string{"/provisioned/python/bin"}, nil, nil, nil)
	if env == nil {
		t.Fatal("expected non-nil env when runtime paths are set")
	}

	var pathEntry string
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			pathEntry = e
			break
		}
	}
	if pathEntry == "" {
		t.Fatal("expected PATH entry in env")
	}
	if !strings.Contains(pathEntry, "/provisioned/python/bin") {
		t.Errorf("PATH should contain provisioned bin dir, got: %s", pathEntry)
	}
}

func TestBuildStepEnv_WithInputsAndRuntimePaths(t *testing.T) {
	inputEnv := []string{"PI_INPUT_VERSION=3.13"}
	env := BuildStepEnv([]string{"/provisioned/node/bin"}, inputEnv, nil, nil)
	if env == nil {
		t.Fatal("expected non-nil env")
	}

	hasInput := false
	hasPath := false
	for _, e := range env {
		if e == "PI_INPUT_VERSION=3.13" {
			hasInput = true
		}
		if strings.HasPrefix(e, "PATH=") && strings.Contains(e, "/provisioned/node/bin") {
			hasPath = true
		}
	}
	if !hasInput {
		t.Error("expected PI_INPUT_VERSION in env")
	}
	if !hasPath {
		t.Error("expected provisioned PATH in env")
	}
}

func TestPrependPathInEnv(t *testing.T) {
	env := []string{"HOME=/home/user", "PATH=/usr/bin:/bin", "SHELL=/bin/zsh"}
	result := prependPathInEnv(env, []string{"/foo/bin", "/bar/bin"})

	var pathEntry string
	for _, e := range result {
		if strings.HasPrefix(e, "PATH=") {
			pathEntry = e
			break
		}
	}

	expected := "PATH=/foo/bin" + string(os.PathListSeparator) + "/bar/bin" + string(os.PathListSeparator) + "/usr/bin:/bin"
	if pathEntry != expected {
		t.Errorf("PATH = %q, want %q", pathEntry, expected)
	}
}

func TestPrependPathInEnv_NoPATH(t *testing.T) {
	env := []string{"HOME=/home/user", "SHELL=/bin/zsh"}
	result := prependPathInEnv(env, []string{"/foo/bin"})

	var found bool
	for _, e := range result {
		if strings.HasPrefix(e, "PATH=") {
			found = true
			if e != "PATH=/foo/bin" {
				t.Errorf("PATH = %q, want %q", e, "PATH=/foo/bin")
			}
		}
	}
	if !found {
		t.Error("expected PATH entry to be added")
	}
}
