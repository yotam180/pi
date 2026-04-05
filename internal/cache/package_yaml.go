package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const packageYAMLFile = "pi-package.yaml"

// PackageYAML represents the optional pi-package.yaml file in a package repo.
type PackageYAML struct {
	MinPIVersion string `yaml:"min_pi_version"`
}

// checkPackageYAML reads pi-package.yaml from the given directory (if it exists)
// and validates min_pi_version against the running PI version.
func (c *Cache) checkPackageYAML(dir string) error {
	path := filepath.Join(dir, packageYAMLFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // absent is fine
		}
		return fmt.Errorf("reading %s: %w", packageYAMLFile, err)
	}

	var pkg PackageYAML
	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("parsing %s: %w", packageYAMLFile, err)
	}

	if pkg.MinPIVersion == "" {
		return nil
	}

	if c.PIVersion == "" || c.PIVersion == "dev" {
		return nil // dev builds skip version checks
	}

	if !versionSatisfies(c.PIVersion, pkg.MinPIVersion) {
		return fmt.Errorf("this package requires PI >= %s, but you are running PI %s — upgrade PI to use this package",
			pkg.MinPIVersion, c.PIVersion)
	}

	return nil
}

// versionSatisfies returns true if running >= required.
// Versions are compared as dot-separated numeric components.
// Leading "v" prefix is stripped.
func versionSatisfies(running, required string) bool {
	running = strings.TrimPrefix(running, "v")
	required = strings.TrimPrefix(required, "v")

	runParts := parseVersionParts(running)
	reqParts := parseVersionParts(required)

	maxLen := len(runParts)
	if len(reqParts) > maxLen {
		maxLen = len(reqParts)
	}

	for i := 0; i < maxLen; i++ {
		r := 0
		if i < len(runParts) {
			r = runParts[i]
		}
		q := 0
		if i < len(reqParts) {
			q = reqParts[i]
		}
		if r > q {
			return true
		}
		if r < q {
			return false
		}
	}

	return true // equal
}

func parseVersionParts(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			break // stop at non-numeric component
		}
		result = append(result, n)
	}
	return result
}
