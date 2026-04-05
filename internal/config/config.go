package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/conditions"
	"gopkg.in/yaml.v3"
)

const FileName = "pi.yaml"

// ProvisionMode controls whether PI auto-provisions missing runtimes.
type ProvisionMode string

const (
	ProvisionNever ProvisionMode = "never"
	ProvisionAsk   ProvisionMode = "ask"
	ProvisionAuto  ProvisionMode = "auto"
)

// RuntimeManager selects which backend provisions runtimes.
type RuntimeManager string

const (
	RuntimeManagerMise   RuntimeManager = "mise"
	RuntimeManagerDirect RuntimeManager = "direct"
)

// RuntimesConfig holds the runtimes: block from pi.yaml.
type RuntimesConfig struct {
	Provision ProvisionMode  `yaml:"provision"`
	Manager   RuntimeManager `yaml:"manager"`
}

// PackageEntry represents one entry in the packages: list.
// It can be either a simple string ("org/repo@v1.2") or an object with
// source: and optional as: keys.
type PackageEntry struct {
	Source string `yaml:"source"`
	As     string `yaml:"as"`
}

// UnmarshalYAML implements custom unmarshalling for PackageEntry so it can
// accept both a plain string ("org/repo@v1.2") and an object ({source: ..., as: ...}).
func (p *PackageEntry) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		p.Source = value.Value
		return nil
	}

	if value.Kind == yaml.MappingNode {
		type packageAlias PackageEntry
		var alias packageAlias
		if err := value.Decode(&alias); err != nil {
			return err
		}
		*p = PackageEntry(alias)
		return nil
	}

	return fmt.Errorf("line %d: package entry must be a string or object", value.Line)
}

// IsFileSource returns true if the package source starts with "file:".
func (p *PackageEntry) IsFileSource() bool {
	return strings.HasPrefix(p.Source, "file:")
}

// FilePath returns the filesystem path for a file: source, with ~ expanded.
// Returns empty string for non-file sources.
func (p *PackageEntry) FilePath() string {
	if !p.IsFileSource() {
		return ""
	}
	path := strings.TrimPrefix(p.Source, "file:")
	return expandTilde(path)
}

// expandTilde replaces a leading "~/" with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// ProjectConfig represents the top-level pi.yaml file.
type ProjectConfig struct {
	Project   string                `yaml:"project"`
	Shortcuts map[string]Shortcut   `yaml:"shortcuts"`
	Setup     []SetupEntry          `yaml:"setup"`
	Packages  []PackageEntry        `yaml:"packages"`
	Runtimes  *RuntimesConfig       `yaml:"runtimes"`
}

// Shortcut can be either a simple string (automation name) or an object with
// additional options like "anywhere" and "with" for input mapping.
type Shortcut struct {
	Run      string            `yaml:"run"`
	Anywhere bool              `yaml:"anywhere"`
	With     map[string]string `yaml:"with"`
}

// SetupEntry represents one entry in the setup list.
// It can be either a bare string ("setup/install-go") or an object with run:, if:, with: keys.
type SetupEntry struct {
	Run  string            `yaml:"run"`
	With map[string]string `yaml:"with"`
	If   string            `yaml:"if"`
}

// UnmarshalYAML implements custom unmarshalling for SetupEntry so it can accept
// both a plain string ("setup/install-go") and an object ({run: ..., if: ...}).
func (e *SetupEntry) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		e.Run = value.Value
		return nil
	}

	if value.Kind == yaml.MappingNode {
		type setupAlias SetupEntry
		var alias setupAlias
		if err := value.Decode(&alias); err != nil {
			return err
		}
		*e = SetupEntry(alias)
		return nil
	}

	return fmt.Errorf("line %d: setup entry must be a string or object", value.Line)
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
		if entry.If != "" {
			if _, err := conditions.Predicates(entry.If); err != nil {
				return fmt.Errorf("%s: setup[%d] invalid if expression: %w", path, i, err)
			}
		}
	}

	if err := c.validatePackages(path); err != nil {
		return err
	}

	if c.Runtimes != nil {
		if err := c.Runtimes.validate(path); err != nil {
			return err
		}
	}

	return nil
}

func (c *ProjectConfig) validatePackages(path string) error {
	aliases := make(map[string]int) // alias → index of first occurrence

	for i, pkg := range c.Packages {
		if pkg.Source == "" {
			return fmt.Errorf("%s: packages[%d] has empty source", path, i)
		}

		if pkg.As != "" {
			if prev, exists := aliases[pkg.As]; exists {
				return fmt.Errorf("%s: packages[%d] alias %q duplicates packages[%d]", path, i, pkg.As, prev)
			}
			aliases[pkg.As] = i

			if strings.Contains(pkg.As, "/") {
				return fmt.Errorf("%s: packages[%d] alias %q must not contain \"/\"", path, i, pkg.As)
			}
		}
	}

	return nil
}

func (r *RuntimesConfig) validate(path string) error {
	switch r.Provision {
	case "", ProvisionNever, ProvisionAsk, ProvisionAuto:
	default:
		return fmt.Errorf("%s: runtimes.provision must be one of: never, ask, auto (got %q)", path, r.Provision)
	}

	switch r.Manager {
	case "", RuntimeManagerMise, RuntimeManagerDirect:
	default:
		return fmt.Errorf("%s: runtimes.manager must be one of: mise, direct (got %q)", path, r.Manager)
	}

	return nil
}

// PackageAliases returns a set of all package alias names declared in packages:.
// Used by discovery to detect alias-prefixed automation references.
func (c *ProjectConfig) PackageAliases() map[string]bool {
	aliases := make(map[string]bool)
	for _, pkg := range c.Packages {
		if pkg.As != "" {
			aliases[pkg.As] = true
		}
	}
	return aliases
}

// EffectiveProvisionMode returns the provision mode, defaulting to "never".
func (c *ProjectConfig) EffectiveProvisionMode() ProvisionMode {
	if c.Runtimes == nil || c.Runtimes.Provision == "" {
		return ProvisionNever
	}
	return c.Runtimes.Provision
}

// EffectiveRuntimeManager returns the runtime manager, defaulting to "mise".
func (c *ProjectConfig) EffectiveRuntimeManager() RuntimeManager {
	if c.Runtimes == nil || c.Runtimes.Manager == "" {
		return RuntimeManagerMise
	}
	return c.Runtimes.Manager
}
