package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
)

const (
	PiDir              = ".pi"
	automationFileName = "automation.yaml"
	yamlExt            = ".yaml"
)

const BuiltinPrefix = "pi:"

// Result holds all discovered automations and provides lookup.
type Result struct {
	Automations map[string]*automation.Automation
	Builtins    map[string]*automation.Automation // keyed without "pi:" prefix
	names       []string                          // sorted local names
	builtinSet  map[string]bool                   // tracks which names came from builtins
}

// Discover walks the given .pi/ directory, finds all automation YAML files,
// parses them, and returns a Result. It handles two resolution forms:
//   - .pi/docker/up.yaml          → name "docker/up"
//   - .pi/setup/cursor/automation.yaml → name "setup/cursor"
//
// Returns an error if two files resolve to the same automation name.
func Discover(piDir string) (*Result, error) {
	info, err := os.Stat(piDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Result{
				Automations: make(map[string]*automation.Automation),
				builtinSet:  make(map[string]bool),
			}, nil
		}
		return nil, fmt.Errorf("accessing %s: %w", piDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", piDir)
	}

	automations := make(map[string]*automation.Automation)
	sources := make(map[string]string) // name → file path (for collision error messages)

	err = filepath.Walk(piDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != yamlExt {
			return nil
		}

		name, ok := deriveName(piDir, path)
		if !ok {
			return nil
		}

		name = normalizeName(name)

		if existingPath, exists := sources[name]; exists {
			return fmt.Errorf("automation name collision: %q resolves from both %s and %s", name, existingPath, path)
		}

		a, err := automation.Load(path)
		if err != nil {
			return fmt.Errorf("loading %s: %w", path, err)
		}

		automations[name] = a
		sources[name] = path
		return nil
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(automations))
	for n := range automations {
		names = append(names, n)
	}
	sort.Strings(names)

	return &Result{
		Automations: automations,
		names:       names,
		builtinSet:  make(map[string]bool),
	}, nil
}

// NewResult creates a Result from a pre-built map and sorted name list.
func NewResult(automations map[string]*automation.Automation, names []string) *Result {
	return &Result{
		Automations: automations,
		names:       names,
		builtinSet:  make(map[string]bool),
	}
}

// MergeBuiltins incorporates built-in automations into this result.
// Built-ins are stored separately: Find("pi:hello") always resolves to the
// built-in, while Find("hello") resolves to local first, falling back to built-in.
// pi list shows built-ins that don't collide with local names.
func (r *Result) MergeBuiltins(builtinResult *Result) {
	if builtinResult == nil {
		return
	}

	r.Builtins = builtinResult.Automations

	for name, a := range builtinResult.Automations {
		if _, exists := r.Automations[name]; !exists {
			r.Automations[name] = a
			r.builtinSet[name] = true
		}
	}

	r.rebuildNames()
}

// deriveName converts a file path within the .pi/ directory into an automation name.
// Returns the name and true if the file is a valid automation, or ("", false) if
// it should be skipped (e.g., non-automation yaml files).
func deriveName(piDir, path string) (string, bool) {
	rel, err := filepath.Rel(piDir, path)
	if err != nil {
		return "", false
	}

	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)

	if base == automationFileName {
		// .pi/setup/cursor/automation.yaml → "setup/cursor"
		dir := filepath.Dir(rel)
		if dir == "." {
			return "", false
		}
		return filepath.ToSlash(dir), true
	}

	if filepath.Ext(base) == yamlExt {
		// .pi/docker/up.yaml → "docker/up"
		return strings.TrimSuffix(rel, yamlExt), true
	}

	return "", false
}

// normalizeName cleans up an automation name: lowercase, no leading/trailing slashes.
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.Trim(name, "/")
	return filepath.ToSlash(name)
}

// Find looks up an automation by name. If the name starts with "pi:", it
// resolves exclusively from built-ins. Otherwise, it checks local automations
// first, then falls back to built-ins.
func (r *Result) Find(name string) (*automation.Automation, error) {
	if strings.HasPrefix(name, BuiltinPrefix) {
		stripped := normalizeName(strings.TrimPrefix(name, BuiltinPrefix))
		if r.Builtins != nil {
			if a, ok := r.Builtins[stripped]; ok {
				return a, nil
			}
		}
		return nil, fmt.Errorf("built-in automation %q not found", name)
	}

	name = normalizeName(name)

	if a, ok := r.Automations[name]; ok {
		return a, nil
	}

	if len(r.names) == 0 {
		return nil, fmt.Errorf("automation %q not found (no automations discovered)", name)
	}

	return nil, fmt.Errorf("automation %q not found\n\nAvailable automations:\n%s", name, r.formatAvailable())
}

// Names returns a sorted list of all automation names (local + built-in).
func (r *Result) Names() []string {
	out := make([]string, len(r.names))
	copy(out, r.names)
	return out
}

// IsBuiltin returns true if the given name was provided by a built-in automation
// (and not shadowed by a local automation).
func (r *Result) IsBuiltin(name string) bool {
	return r.builtinSet[name]
}

func (r *Result) rebuildNames() {
	r.names = make([]string, 0, len(r.Automations))
	for n := range r.Automations {
		r.names = append(r.names, n)
	}
	sort.Strings(r.names)
}

func (r *Result) formatAvailable() string {
	var b strings.Builder
	for _, name := range r.names {
		a := r.Automations[name]
		if a.Description != "" {
			fmt.Fprintf(&b, "  %-30s %s\n", name, a.Description)
		} else {
			fmt.Fprintf(&b, "  %s\n", name)
		}
	}
	return b.String()
}
