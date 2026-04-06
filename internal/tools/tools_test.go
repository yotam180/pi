package tools

import (
	"strings"
	"testing"

	"github.com/vyper-tooling/pi/internal/builtins"
)

func TestBuildShortNameMap_AllEntriesPresent(t *testing.T) {
	m := BuildShortNameMap()

	cases := []struct {
		short    string
		expected string
	}{
		{"python", "pi:install-python"},
		{"node", "pi:install-node"},
		{"nodejs", "pi:install-node"},
		{"go", "pi:install-go"},
		{"golang", "pi:install-go"},
		{"rust", "pi:install-rust"},
		{"uv", "pi:install-uv"},
		{"tsx", "pi:install-tsx"},
		{"homebrew", "pi:install-homebrew"},
		{"brew", "pi:install-homebrew"},
		{"terraform", "pi:install-terraform"},
		{"tf", "pi:install-terraform"},
		{"kubectl", "pi:install-kubectl"},
		{"k8s", "pi:install-kubectl"},
		{"helm", "pi:install-helm"},
		{"pnpm", "pi:install-pnpm"},
		{"bun", "pi:install-bun"},
		{"deno", "pi:install-deno"},
		{"aws-cli", "pi:install-aws-cli"},
		{"awscli", "pi:install-aws-cli"},
		{"aws", "pi:install-aws-cli"},
	}

	for _, tc := range cases {
		got, ok := m[tc.short]
		if !ok {
			t.Errorf("short name %q missing from map", tc.short)
			continue
		}
		if got != tc.expected {
			t.Errorf("m[%q] = %q, want %q", tc.short, got, tc.expected)
		}
	}
}

func TestBuildShortNameMap_PiPrefixVariants(t *testing.T) {
	m := BuildShortNameMap()

	piPrefixed := []struct {
		short    string
		expected string
	}{
		{"pi:python", "pi:install-python"},
		{"pi:node", "pi:install-node"},
		{"pi:go", "pi:install-go"},
		{"pi:rust", "pi:install-rust"},
		{"pi:uv", "pi:install-uv"},
		{"pi:tsx", "pi:install-tsx"},
		{"pi:homebrew", "pi:install-homebrew"},
		{"pi:brew", "pi:install-homebrew"},
		{"pi:terraform", "pi:install-terraform"},
		{"pi:kubectl", "pi:install-kubectl"},
		{"pi:helm", "pi:install-helm"},
		{"pi:pnpm", "pi:install-pnpm"},
		{"pi:bun", "pi:install-bun"},
		{"pi:deno", "pi:install-deno"},
		{"pi:aws-cli", "pi:install-aws-cli"},
	}

	for _, tc := range piPrefixed {
		got, ok := m[tc.short]
		if !ok {
			t.Errorf("pi: prefix form %q missing from map", tc.short)
			continue
		}
		if got != tc.expected {
			t.Errorf("m[%q] = %q, want %q", tc.short, got, tc.expected)
		}
	}
}

func TestBuildShortNameMap_NoCommandOnlyEntries(t *testing.T) {
	m := BuildShortNameMap()

	commandOnly := []string{"docker", "jq", "git", "curl", "wget", "rustc", "cargo", "rustup", "make", "mise", "ruby"}
	for _, name := range commandOnly {
		if _, ok := m[name]; ok {
			t.Errorf("command-only tool %q should not be in short name map", name)
		}
	}
}

func TestInstallHintFor_KnownTools(t *testing.T) {
	tools := []string{"python", "node", "go", "rust", "docker", "jq", "git", "curl", "wget", "rustc", "cargo", "rustup", "make", "mise", "uv", "tsx"}
	for _, tool := range tools {
		hint := InstallHintFor(tool)
		if hint == "" {
			t.Errorf("InstallHintFor(%q) should return a hint", tool)
		}
	}
}

func TestInstallHintFor_UnknownTool(t *testing.T) {
	hint := InstallHintFor("unknown-tool-xyz")
	if hint != "" {
		t.Errorf("InstallHintFor(unknown) = %q, want empty", hint)
	}
}

func TestToolResolutionHelp_ContainsAllBuiltins(t *testing.T) {
	help := ToolResolutionHelp()

	expected := []string{
		"python", "pi:install-python",
		"node", "pi:install-node",
		"go", "pi:install-go",
		"rust", "pi:install-rust",
		"uv", "pi:install-uv",
		"homebrew", "pi:install-homebrew",
		"terraform", "pi:install-terraform",
		"kubectl", "pi:install-kubectl",
		"helm", "pi:install-helm",
		"pnpm", "pi:install-pnpm",
		"bun", "pi:install-bun",
		"deno", "pi:install-deno",
		"aws-cli", "pi:install-aws-cli",
	}
	for _, s := range expected {
		if !strings.Contains(help, s) {
			t.Errorf("help text missing %q:\n%s", s, help)
		}
	}
}

func TestToolResolutionHelp_Deterministic(t *testing.T) {
	first := ToolResolutionHelp()
	for i := 0; i < 10; i++ {
		got := ToolResolutionHelp()
		if got != first {
			t.Fatalf("non-deterministic output on iteration %d:\nfirst:\n%s\ngot:\n%s", i, first, got)
		}
	}
}

func TestToolResolutionHelp_PrefersCanonicalName(t *testing.T) {
	help := ToolResolutionHelp()

	if strings.Contains(help, "golang") {
		t.Errorf("help should show 'go' not 'golang':\n%s", help)
	}
	if strings.Contains(help, "nodejs") {
		t.Errorf("help should show 'node' not 'nodejs':\n%s", help)
	}
	if strings.Contains(help, "awscli") {
		t.Errorf("help should show 'aws-cli' not 'awscli':\n%s", help)
	}

	lines := strings.Split(help, "\n")
	foundHomebrew := false
	for _, line := range lines {
		if strings.Contains(line, "pi:install-homebrew") {
			if !strings.Contains(line, "homebrew") {
				t.Errorf("expected 'homebrew' for pi:install-homebrew entry, got: %s", line)
			}
			foundHomebrew = true
		}
	}
	if !foundHomebrew {
		t.Errorf("expected pi:install-homebrew in help text:\n%s", help)
	}
}

func TestRegistryCoversAllInstallBuiltins(t *testing.T) {
	result, err := builtins.Discover()
	if err != nil {
		t.Fatalf("builtins.Discover() error: %v", err)
	}

	shortMap := BuildShortNameMap()

	builtinTargets := make(map[string]bool)
	for _, target := range shortMap {
		builtinTargets[target] = true
	}

	for _, name := range result.Names() {
		if !strings.HasPrefix(name, "install-") {
			continue
		}
		fullName := "pi:" + name
		if !builtinTargets[fullName] {
			t.Errorf("builtin %q has no entry in tools.Registry", fullName)
		}
	}
}

func TestRegistryBuiltinNamesExist(t *testing.T) {
	result, err := builtins.Discover()
	if err != nil {
		t.Fatalf("builtins.Discover() error: %v", err)
	}

	builtinNames := make(map[string]bool)
	for _, name := range result.Names() {
		builtinNames[name] = true
	}

	for _, td := range Registry {
		if td.BuiltinName == "" {
			continue
		}
		name := strings.TrimPrefix(td.BuiltinName, "pi:")
		if !builtinNames[name] {
			t.Errorf("Registry entry %q references non-existent builtin (no %q in builtins)", td.BuiltinName, name)
		}
	}
}
