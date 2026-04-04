package builtins

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
)

//go:embed all:embed_pi
var embeddedFS embed.FS

const (
	embedRoot          = "embed_pi"
	automationFileName = "automation.yaml"
	yamlExt            = ".yaml"
	Prefix             = "pi:"
)

// Discover walks the embedded .pi/ directory and returns all built-in automations
// as a discovery.Result. Names do NOT include the "pi:" prefix in the result map —
// the prefix is handled during resolution.
func Discover() (*discovery.Result, error) {
	automations := make(map[string]*automation.Automation)
	sources := make(map[string]string)

	err := fs.WalkDir(embeddedFS, embedRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != yamlExt {
			return nil
		}

		name, ok := deriveName(embedRoot, path)
		if !ok {
			return nil
		}

		name = normalizeName(name)

		if existingPath, exists := sources[name]; exists {
			return fmt.Errorf("built-in automation name collision: %q from both %s and %s", name, existingPath, path)
		}

		data, err := embeddedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded %s: %w", path, err)
		}

		a, err := automation.LoadFromBytes(data, path)
		if err != nil {
			return fmt.Errorf("loading built-in %s: %w", path, err)
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

	return discovery.NewResult(automations, names), nil
}

func deriveName(root, path string) (string, bool) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", false
	}

	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)

	if base == automationFileName {
		dir := filepath.Dir(rel)
		if dir == "." {
			return "", false
		}
		return filepath.ToSlash(dir), true
	}

	if filepath.Ext(base) == yamlExt {
		return strings.TrimSuffix(rel, yamlExt), true
	}

	return "", false
}

func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.Trim(name, "/")
	return filepath.ToSlash(name)
}
