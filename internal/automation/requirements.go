package automation

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RequirementKind distinguishes runtime requirements from command requirements.
type RequirementKind string

const (
	RequirementRuntime RequirementKind = "runtime"
	RequirementCommand RequirementKind = "command"
)

// knownRuntimes lists the runtime names that can appear as bare identifiers
// in a requires: entry (e.g. "python >= 3.11" or "node").
var knownRuntimes = map[string]bool{
	"python": true,
	"node":   true,
	"go":     true,
	"rust":   true,
}

// knownRuntimeNames returns a sorted, comma-separated list of known runtime names.
func knownRuntimeNames() string {
	names := make([]string, 0, len(knownRuntimes))
	for name := range knownRuntimes {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Requirement declares a tool or runtime that an automation needs.
type Requirement struct {
	Name       string
	Kind       RequirementKind
	MinVersion string // empty means any version
}

// requirementRaw handles YAML unmarshalling for a single requires: entry.
// Supports four forms:
//   - "python"             → runtime, any version
//   - "python >= 3.11"     → runtime, minimum version
//   - "command: docker"    → command, any version
//   - "command: jq >= 1.7" → command, minimum version
type requirementRaw struct {
	Requirement
}

func (r *requirementRaw) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		return r.parseScalar(value.Value)
	case yaml.MappingNode:
		return r.parseMapping(value)
	default:
		return fmt.Errorf("requires entry must be a string or a mapping, got %v", value.Kind)
	}
}

func (r *requirementRaw) parseScalar(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("requires entry cannot be empty")
	}

	name, version, err := parseNameVersion(s)
	if err != nil {
		return fmt.Errorf("requires entry %q: %w", s, err)
	}

	if !knownRuntimes[name] {
		return fmt.Errorf("requires entry %q: unknown runtime %q (known: %s); use \"command: %s\" for arbitrary commands", s, name, knownRuntimeNames(), name)
	}

	r.Kind = RequirementRuntime
	r.Name = name
	r.MinVersion = version
	return nil
}

func (r *requirementRaw) parseMapping(value *yaml.Node) error {
	if len(value.Content) != 2 {
		return fmt.Errorf("requires mapping must have exactly one key")
	}
	key := value.Content[0].Value
	val := strings.TrimSpace(value.Content[1].Value)

	if key != "command" {
		return fmt.Errorf("requires mapping: unknown key %q (expected \"command\")", key)
	}
	if val == "" {
		return fmt.Errorf("requires entry \"command:\" value cannot be empty")
	}

	name, version, err := parseNameVersion(val)
	if err != nil {
		return fmt.Errorf("requires entry \"command: %s\": %w", val, err)
	}

	r.Kind = RequirementCommand
	r.Name = name
	r.MinVersion = version
	return nil
}

// parseNameVersion splits "name >= version" or bare "name".
// Returns (name, version, error). version is empty for bare names.
func parseNameVersion(s string) (string, string, error) {
	if idx := strings.Index(s, ">="); idx != -1 {
		name := strings.TrimSpace(s[:idx])
		version := strings.TrimSpace(s[idx+2:])
		if name == "" {
			return "", "", fmt.Errorf("missing name before >=")
		}
		if version == "" {
			return "", "", fmt.Errorf("missing version after >=")
		}
		if err := validateVersionString(version); err != nil {
			return "", "", fmt.Errorf("invalid version %q: %w", version, err)
		}
		return name, version, nil
	}

	if strings.ContainsAny(s, " \t") {
		return "", "", fmt.Errorf("invalid format; use \"name >= version\" for version constraints")
	}

	return s, "", nil
}

// validateVersionString checks that a version string looks like a semver
// prefix: at least one numeric component, optionally dot-separated.
func validateVersionString(v string) error {
	parts := strings.Split(v, ".")
	for _, p := range parts {
		if p == "" {
			return fmt.Errorf("empty component in version")
		}
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return fmt.Errorf("non-numeric character %q in version component %q", string(ch), p)
			}
		}
	}
	return nil
}
