package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vyper-tooling/pi/internal/config"
)

// FindRoot walks up from startDir looking for pi.yaml.
// Returns the absolute path of the directory containing pi.yaml.
// This mimics how git finds .git/ — run pi from any subdirectory and it just works.
func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	for {
		candidate := filepath.Join(dir, config.FileName)
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find %s in %s or any parent directory", config.FileName, startDir)
		}
		dir = parent
	}
}
