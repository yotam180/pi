package reqcheck

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
)

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

// --- ExtractVersion ---

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
		{"just 42", ""},
		{"version 3.13", "3.13"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExtractVersion(tt.input)
			if got != tt.want {
				t.Errorf("ExtractVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- CompareVersions ---

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
			got, err := CompareVersions(tt.a, tt.b)
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
	_, err := CompareVersions("abc", "1.0")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}

	_, err = CompareVersions("1.0", "abc")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}
}

// --- RuntimeCommand ---

func TestRuntimeCommand(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"python", "python3"},
		{"node", "node"},
		{"rust", "rustc"},
		{"go", "go"},
		{"docker", "docker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuntimeCommand(tt.name); got != tt.want {
				t.Errorf("RuntimeCommand(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// --- CheckRequirement ---

func TestCheckRequirement_CommandFound(t *testing.T) {
	env := mockEnv(map[string]bool{"docker": true}, nil)
	req := automation.Requirement{Name: "docker", Kind: automation.RequirementCommand}
	result := CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_CommandNotFound(t *testing.T) {
	env := mockEnv(map[string]bool{}, nil)
	req := automation.Requirement{Name: "kubectl", Kind: automation.RequirementCommand}
	result := CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for missing command")
	}
	if result.Error != "not found" {
		t.Errorf("unexpected error: %s", result.Error)
	}
}

func TestCheckRequirement_RuntimeFound(t *testing.T) {
	env := mockEnv(map[string]bool{"python3": true}, nil)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	result := CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_RuntimeNotFound(t *testing.T) {
	env := mockEnv(map[string]bool{}, nil)
	req := automation.Requirement{Name: "python", Kind: automation.RequirementRuntime}
	result := CheckRequirement(req, env)

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
	result := CheckRequirement(req, env)

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
	result := CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied for version too low")
	}
	if !strings.Contains(result.Error, "3.9.7") || !strings.Contains(result.Error, "3.11") {
		t.Errorf("error should mention both versions, got: %s", result.Error)
	}
}

func TestCheckRequirement_RustRuntime(t *testing.T) {
	env := mockEnv(map[string]bool{"rustc": true}, nil)
	req := automation.Requirement{Name: "rust", Kind: automation.RequirementRuntime}
	result := CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
}

func TestCheckRequirement_RustVersionSatisfied(t *testing.T) {
	env := mockEnv(
		map[string]bool{"rustc": true},
		map[string]string{"rustc": "rustc 1.80.0 (051478957 2024-07-21)"},
	)
	req := automation.Requirement{Name: "rust", Kind: automation.RequirementRuntime, MinVersion: "1.75"}
	result := CheckRequirement(req, env)

	if !result.Satisfied {
		t.Errorf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "1.80.0" {
		t.Errorf("expected detected version 1.80.0, got %q", result.DetectedVersion)
	}
}

func TestCheckRequirement_GoVersion(t *testing.T) {
	env := mockEnv(
		map[string]bool{"go": true},
		map[string]string{"go": "go version go1.23.0 darwin/arm64"},
	)
	req := automation.Requirement{Name: "go", Kind: automation.RequirementRuntime, MinVersion: "1.22"}
	result := CheckRequirement(req, env)

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
	result := CheckRequirement(req, env)

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
	result := CheckRequirement(req, env)

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
	result := CheckRequirement(req, env)

	if result.Satisfied {
		t.Error("expected not satisfied when version is undetectable")
	}
	if !strings.Contains(result.Error, "unable to detect version") {
		t.Errorf("expected 'unable to detect version' error, got: %s", result.Error)
	}
}

// --- CheckRequirementForDoctor ---

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
	result := CheckRequirementForDoctor(req, env)

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
	result := CheckRequirementForDoctor(req, env)

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
	result := CheckRequirementForDoctor(req, env)

	if !result.Satisfied {
		t.Fatalf("expected satisfied, got error: %s", result.Error)
	}
	if result.DetectedVersion != "3.13.0" {
		t.Errorf("expected version 3.13.0, got %q", result.DetectedVersion)
	}
}

// --- DetectVersion ---

func TestDetectVersion_WithMockedEnv(t *testing.T) {
	env := &conditions.RuntimeEnv{
		ExecOutput: func(cmd string, args ...string) string {
			if cmd == "mytool" && args[0] == "--version" {
				return "mytool version 2.5.1"
			}
			return ""
		},
	}

	got := DetectVersion("mytool", env)
	if got != "2.5.1" {
		t.Errorf("DetectVersion = %q, want %q", got, "2.5.1")
	}
}

func TestDetectVersion_FallsBackToVersionSubcommand(t *testing.T) {
	env := &conditions.RuntimeEnv{
		ExecOutput: func(cmd string, args ...string) string {
			if cmd == "go" && args[0] == "version" {
				return "go version go1.23.0 darwin/arm64"
			}
			return ""
		},
	}

	got := DetectVersion("go", env)
	if got != "1.23.0" {
		t.Errorf("DetectVersion = %q, want %q", got, "1.23.0")
	}
}

func TestDetectVersion_NilExecOutput(t *testing.T) {
	env := &conditions.RuntimeEnv{
		ExecOutput: nil,
	}
	// When ExecOutput is nil, it falls back to actually executing commands.
	// We just verify it doesn't panic.
	_ = DetectVersion("nonexistent-tool-xyz", env)
}

// --- FormatValidationError ---

func TestFormatValidationError(t *testing.T) {
	ve := &ValidationError{
		AutomationName: "format-logs",
		Results: []CheckResult{
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

	output := FormatValidationError(ve)

	if !strings.Contains(output, "✗ pi run format-logs") {
		t.Errorf("output should contain automation name, got:\n%s", output)
	}
	if !strings.Contains(output, "Missing requirements:") {
		t.Errorf("output should contain 'Missing requirements:', got:\n%s", output)
	}
	if !strings.Contains(output, "python >= 3.11") {
		t.Errorf("output should contain 'python >= 3.11', got:\n%s", output)
	}
	if !strings.Contains(output, "command: jq") {
		t.Errorf("output should contain 'command: jq', got:\n%s", output)
	}
	if !strings.Contains(output, "brew install jq") {
		t.Errorf("output should contain install hint for jq, got:\n%s", output)
	}
}

func TestFormatValidationError_SkipsSatisfied(t *testing.T) {
	ve := &ValidationError{
		AutomationName: "test",
		Results: []CheckResult{
			{
				Requirement: automation.Requirement{Name: "docker", Kind: automation.RequirementCommand},
				Satisfied:   true,
			},
			{
				Requirement: automation.Requirement{Name: "jq", Kind: automation.RequirementCommand},
				Satisfied:   false,
				Error:       "not found",
			},
		},
	}

	output := FormatValidationError(ve)
	if strings.Contains(output, "docker") {
		t.Errorf("output should not contain satisfied requirement 'docker', got:\n%s", output)
	}
	if !strings.Contains(output, "jq") {
		t.Errorf("output should contain unsatisfied requirement 'jq', got:\n%s", output)
	}
}

// --- FormatRequirementLabel ---

func TestFormatRequirementLabel(t *testing.T) {
	tests := []struct {
		name string
		req  automation.Requirement
		want string
	}{
		{
			"runtime with version",
			automation.Requirement{Name: "python", Kind: automation.RequirementRuntime, MinVersion: "3.11"},
			"python >= 3.11",
		},
		{
			"command without version",
			automation.Requirement{Name: "jq", Kind: automation.RequirementCommand},
			"command: jq",
		},
		{
			"runtime without version",
			automation.Requirement{Name: "python", Kind: automation.RequirementRuntime},
			"python",
		},
		{
			"command with version",
			automation.Requirement{Name: "kubectl", Kind: automation.RequirementCommand, MinVersion: "1.28"},
			"kubectl >= 1.28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRequirementLabel(tt.req)
			if got != tt.want {
				t.Errorf("FormatRequirementLabel = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- InstallHintFor ---

func TestInstallHintFor_KnownTools(t *testing.T) {
	tools := []string{"python", "node", "docker", "jq", "go", "rust", "tsx", "mise", "uv"}
	for _, tool := range tools {
		req := automation.Requirement{Name: tool}
		hint := InstallHintFor(req)
		if hint == "" {
			t.Errorf("InstallHintFor(%q) should return a hint", tool)
		}
	}
}

func TestInstallHintFor_UnknownTool(t *testing.T) {
	req := automation.Requirement{Name: "unknown-tool-xyz"}
	hint := InstallHintFor(req)
	if hint != "" {
		t.Errorf("InstallHintFor(unknown) = %q, want empty", hint)
	}
}

func TestInstallHintFor_RustFamily(t *testing.T) {
	for _, name := range []string{"rust", "rustc", "cargo", "rustup"} {
		req := automation.Requirement{Name: name}
		hint := InstallHintFor(req)
		if !strings.Contains(hint, "rustup") {
			t.Errorf("InstallHintFor(%q) = %q, expected to contain 'rustup'", name, hint)
		}
	}
}

// --- ValidationError.Error() ---

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{AutomationName: "build"}
	got := ve.Error()
	if !strings.Contains(got, "build") {
		t.Errorf("Error() = %q, should contain automation name", got)
	}
	if !strings.Contains(got, "unsatisfied requirements") {
		t.Errorf("Error() = %q, should mention unsatisfied requirements", got)
	}
}

// --- parseVersionParts ---

func TestParseVersionParts(t *testing.T) {
	parts, err := parseVersionParts("3.13.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 3 || parts[0] != 3 || parts[1] != 13 || parts[2] != 0 {
		t.Errorf("parseVersionParts(3.13.0) = %v, want [3 13 0]", parts)
	}
}

func TestParseVersionParts_Invalid(t *testing.T) {
	_, err := parseVersionParts("abc")
	if err == nil {
		t.Fatal("expected error for non-numeric version")
	}
}

// --- Version comparison edge cases ---

func TestCompareVersions_Equal(t *testing.T) {
	got, err := CompareVersions("1.0.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestCompareVersions_DifferentLengths(t *testing.T) {
	got, err := CompareVersions("1.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("1.0 vs 1.0.0 should be equal, got %d", got)
	}

	got, err = CompareVersions("1.0.1", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1 {
		t.Errorf("1.0.1 vs 1.0 should be 1, got %d", got)
	}
}
