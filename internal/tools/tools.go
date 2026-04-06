// Package tools provides a central registry of installable tools.
// Both the CLI setup-add short-name resolution and the reqcheck install
// hints derive their data from this registry, ensuring a single source
// of truth for tool metadata.
package tools

import (
	"sort"
	"strings"
)

// ToolDescriptor describes a tool that PI knows about.
type ToolDescriptor struct {
	// BuiltinName is the canonical pi: automation name (e.g. "pi:install-python").
	// Empty for tools that don't have a pi: builtin (e.g. pure command checks like docker, jq).
	BuiltinName string

	// ShortNames are the accepted short forms for `pi setup add`.
	// The first entry is the canonical short name (matching the install-<name> suffix).
	// Additional entries are aliases.
	ShortNames []string

	// InstallHint is the human-readable install instruction shown by pi doctor.
	InstallHint string
}

// Registry is the authoritative list of all tools PI knows about.
// Tools with a BuiltinName have a corresponding pi:install-* automation.
// Tools without a BuiltinName are command-only entries used for install hints.
var Registry = []ToolDescriptor{
	{
		BuiltinName: "pi:install-python",
		ShortNames:  []string{"python"},
		InstallHint: "brew install python3  or  https://www.python.org/downloads/",
	},
	{
		BuiltinName: "pi:install-node",
		ShortNames:  []string{"node", "nodejs"},
		InstallHint: "brew install node  or  https://nodejs.org/",
	},
	{
		BuiltinName: "pi:install-go",
		ShortNames:  []string{"go", "golang"},
		InstallHint: "brew install go  or  https://go.dev/dl/",
	},
	{
		BuiltinName: "pi:install-rust",
		ShortNames:  []string{"rust"},
		InstallHint: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
	},
	{
		// No builtin yet — pi:install-ruby doesn't exist.
		// When the builtin is created, set BuiltinName: "pi:install-ruby".
		ShortNames:  []string{"ruby"},
		InstallHint: "brew install ruby  or  https://www.ruby-lang.org/en/documentation/installation/",
	},
	{
		BuiltinName: "pi:install-uv",
		ShortNames:  []string{"uv"},
		InstallHint: "curl -LsSf https://astral.sh/uv/install.sh | sh",
	},
	{
		BuiltinName: "pi:install-tsx",
		ShortNames:  []string{"tsx"},
		InstallHint: "npm install -g tsx",
	},
	{
		BuiltinName: "pi:install-homebrew",
		ShortNames:  []string{"homebrew", "brew"},
		InstallHint: "https://brew.sh/",
	},
	{
		BuiltinName: "pi:install-terraform",
		ShortNames:  []string{"terraform", "tf"},
		InstallHint: "brew install terraform  or  https://developer.hashicorp.com/terraform/downloads",
	},
	{
		BuiltinName: "pi:install-kubectl",
		ShortNames:  []string{"kubectl", "k8s"},
		InstallHint: "brew install kubectl",
	},
	{
		BuiltinName: "pi:install-helm",
		ShortNames:  []string{"helm"},
		InstallHint: "brew install helm",
	},
	{
		BuiltinName: "pi:install-pnpm",
		ShortNames:  []string{"pnpm"},
		InstallHint: "npm install -g pnpm  or  https://pnpm.io/installation",
	},
	{
		BuiltinName: "pi:install-bun",
		ShortNames:  []string{"bun"},
		InstallHint: "curl -fsSL https://bun.sh/install | bash",
	},
	{
		BuiltinName: "pi:install-deno",
		ShortNames:  []string{"deno"},
		InstallHint: "curl -fsSL https://deno.land/install.sh | sh",
	},
	{
		BuiltinName: "pi:install-aws-cli",
		ShortNames:  []string{"aws-cli", "awscli", "aws"},
		InstallHint: "brew install awscli  or  https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html",
	},

	// Tools without a pi: builtin — used for install hints in pi doctor.
	{
		ShortNames:  []string{"docker"},
		InstallHint: "brew install --cask docker  or  https://docs.docker.com/get-docker/",
	},
	{
		ShortNames:  []string{"jq"},
		InstallHint: "brew install jq",
	},
	{
		ShortNames:  []string{"git"},
		InstallHint: "brew install git",
	},
	{
		ShortNames:  []string{"curl"},
		InstallHint: "brew install curl",
	},
	{
		ShortNames:  []string{"wget"},
		InstallHint: "brew install wget",
	},
	{
		ShortNames:  []string{"rustc", "cargo", "rustup"},
		InstallHint: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
	},
	{
		ShortNames:  []string{"make"},
		InstallHint: "xcode-select --install  (macOS)  or  apt install build-essential",
	},
	{
		ShortNames:  []string{"mise"},
		InstallHint: "curl https://mise.run | sh",
	},
}

// BuildShortNameMap returns a map from short names (and pi: prefix variants)
// to canonical pi:install-* automation names. Only entries with a non-empty
// BuiltinName are included.
func BuildShortNameMap() map[string]string {
	m := make(map[string]string)
	for _, td := range Registry {
		if td.BuiltinName == "" {
			continue
		}
		for _, short := range td.ShortNames {
			m[short] = td.BuiltinName
			m["pi:"+short] = td.BuiltinName
		}
	}
	return m
}

// InstallHintFor returns a human-readable install hint for the given tool name.
// Returns empty string if no hint is known.
func InstallHintFor(name string) string {
	for _, td := range Registry {
		for _, short := range td.ShortNames {
			if short == name {
				return td.InstallHint
			}
		}
	}
	return ""
}

// ToolResolutionHelp builds a human-readable table showing how short-form
// tool names resolve to their canonical pi:install-* names. Used by
// `pi setup add --help`. Only entries with builtins are shown.
//
// When multiple short names resolve to the same target, the canonical name
// is preferred (the one matching the pi:install-<name> suffix).
func ToolResolutionHelp() string {
	type pair struct {
		short    string
		resolved string
	}

	best := make(map[string]string)
	for _, td := range Registry {
		if td.BuiltinName == "" {
			continue
		}
		suffix := strings.TrimPrefix(td.BuiltinName, "pi:install-")
		chosen := td.ShortNames[0]
		for _, name := range td.ShortNames {
			if name == suffix {
				chosen = name
				break
			}
		}
		best[td.BuiltinName] = chosen
	}

	pairs := make([]pair, 0, len(best))
	for resolved, short := range best {
		pairs = append(pairs, pair{short, resolved})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].short < pairs[j].short
	})

	var b strings.Builder
	b.WriteString("Tool names are resolved automatically:\n")
	for _, p := range pairs {
		b.WriteString("  ")
		b.WriteString(p.short)
		padding := 12 - len(p.short)
		if padding < 1 {
			padding = 1
		}
		for i := 0; i < padding; i++ {
			b.WriteByte(' ')
		}
		b.WriteString("→  ")
		b.WriteString(p.resolved)
		b.WriteByte('\n')
	}
	return b.String()
}
