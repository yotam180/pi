// Package runtimeinfo is the single source of truth for all runtime knowledge
// in PI. Every package that needs to know about supported runtimes (names,
// binaries, default versions) imports this package instead of maintaining its
// own copy.
package runtimeinfo

import (
	"sort"
	"strings"
)

// Descriptor describes everything PI knows about a supported runtime.
type Descriptor struct {
	Name           string // "python", "node", "go", "rust"
	Binary         string // "python3", "node", "go", "rustc"
	DefaultVersion string // "3.13", "20", "1.23", "stable"
	DirectDownload bool   // whether provisionDirect supports this runtime natively
	InstallHint    string // human-readable hint for pi doctor
}

// Runtimes is the canonical list of all supported runtimes.
var Runtimes = []Descriptor{
	{
		Name:           "python",
		Binary:         "python3",
		DefaultVersion: "3.13",
		DirectDownload: true,
		InstallHint:    "brew install python3  or  https://www.python.org/downloads/",
	},
	{
		Name:           "node",
		Binary:         "node",
		DefaultVersion: "20",
		DirectDownload: true,
		InstallHint:    "brew install node  or  https://nodejs.org/",
	},
	{
		Name:           "go",
		Binary:         "go",
		DefaultVersion: "1.23",
		DirectDownload: false,
		InstallHint:    "brew install go  or  https://go.dev/dl/",
	},
	{
		Name:           "rust",
		Binary:         "rustc",
		DefaultVersion: "stable",
		DirectDownload: false,
		InstallHint:    "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
	},
}

// Find returns the Descriptor for the given runtime name, or nil if unknown.
func Find(name string) *Descriptor {
	for i := range Runtimes {
		if Runtimes[i].Name == name {
			return &Runtimes[i]
		}
	}
	return nil
}

// KnownNames returns the set of known runtime names as a map (for validation).
func KnownNames() map[string]bool {
	m := make(map[string]bool, len(Runtimes))
	for _, r := range Runtimes {
		m[r.Name] = true
	}
	return m
}

// SortedNames returns a sorted, comma-separated list of known runtime names.
func SortedNames() string {
	names := make([]string, len(Runtimes))
	for i, r := range Runtimes {
		names[i] = r.Name
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Binary returns the CLI binary name for a runtime, or the name itself if unknown.
func Binary(name string) string {
	if d := Find(name); d != nil {
		return d.Binary
	}
	return name
}

// DefaultVersion returns the default version for a runtime, or "latest" if unknown.
func DefaultVersion(name string) string {
	if d := Find(name); d != nil {
		return d.DefaultVersion
	}
	return "latest"
}
