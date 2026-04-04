package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = "pi.yaml"

// ProjectConfig represents the top-level pi.yaml file.
type ProjectConfig struct {
	Project   string                `yaml:"project"`
	Shortcuts map[string]Shortcut   `yaml:"shortcuts"`
	Setup     []SetupEntry          `yaml:"setup"`
}

// Shortcut can be either a simple string (automation name) or an object with
// additional options like "anywhere".
type Shortcut struct {
	Run      string `yaml:"run"`
	Anywhere bool   `yaml:"anywhere"`
}

// SetupEntry represents one entry in the setup list.
type SetupEntry struct {
	Run  string            `yaml:"run"`
	With map[string]string `yaml:"with"`
}

// UnmarshalYAML implements custom unmarshalling for Shortcut so it can accept
// both a plain string ("docker/up") and an object ({run: ..., anywhere: true}).
func (s *Shortcut) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		s.Run = value.Value
		s.Anywhere = false
		return nil
	}

	if value.Kind == yaml.MappingNode {
		type shortcutAlias Shortcut
		var alias shortcutAlias
		if err := value.Decode(&alias); err != nil {
			return err
		}
		*s = Shortcut(alias)
		return nil
	}

	return fmt.Errorf("line %d: shortcut must be a string or object", value.Line)
}

// Load reads and parses pi.yaml from the given directory.
// Returns an error if the file is missing, malformed, or fails validation.
func Load(dir string) (*ProjectConfig, error) {
	path := filepath.Join(dir, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s not found in %s", FileName, dir)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	cfg := &ProjectConfig{}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if err := cfg.validate(path); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ProjectConfig) validate(path string) error {
	if c.Project == "" {
		return fmt.Errorf("%s: \"project\" is required", path)
	}

	for name, shortcut := range c.Shortcuts {
		if shortcut.Run == "" {
			return fmt.Errorf("%s: shortcut %q has empty \"run\" field", path, name)
		}
	}

	for i, entry := range c.Setup {
		if entry.Run == "" {
			return fmt.Errorf("%s: setup[%d] has empty \"run\" field", path, i)
		}
	}

	return nil
}
