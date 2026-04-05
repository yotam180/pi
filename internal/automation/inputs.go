package automation

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// InputSpec declares a single input parameter for an automation.
type InputSpec struct {
	Type        string `yaml:"type"`
	Required    *bool  `yaml:"required"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
}

// IsRequired returns true if the input is explicitly marked required or has no default.
// If Required is not set, it defaults to true when no Default is provided.
func (s InputSpec) IsRequired() bool {
	if s.Required != nil {
		return *s.Required
	}
	return s.Default == ""
}

// inputsRaw preserves declaration order for positional mapping.
type inputsRaw struct {
	Keys  []string
	Specs map[string]InputSpec
}

func (r *inputsRaw) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("inputs must be a mapping")
	}
	r.Specs = make(map[string]InputSpec)
	for i := 0; i < len(value.Content)-1; i += 2 {
		key := value.Content[i].Value
		var spec InputSpec
		if err := value.Content[i+1].Decode(&spec); err != nil {
			return fmt.Errorf("input %q: %w", key, err)
		}
		r.Keys = append(r.Keys, key)
		r.Specs[key] = spec
	}
	return nil
}

// ResolveInputs validates and resolves input values from the provided sources.
// withArgs are --with key=value pairs, positionalArgs are bare positional args.
// Returns a map of input name → resolved value. Only one of withArgs or positionalArgs
// may be non-empty; mixing them is an error.
func (a *Automation) ResolveInputs(withArgs map[string]string, positionalArgs []string) (map[string]string, error) {
	if len(a.Inputs) == 0 {
		return nil, nil
	}

	hasWith := len(withArgs) > 0
	hasPositional := len(positionalArgs) > 0

	if hasWith && hasPositional {
		return nil, fmt.Errorf("cannot mix --with flags and positional arguments")
	}

	resolved := make(map[string]string)

	if hasWith {
		for k := range withArgs {
			if _, ok := a.Inputs[k]; !ok {
				return nil, fmt.Errorf("unknown input %q (available: %s)", k, strings.Join(a.InputKeys, ", "))
			}
		}
		for _, key := range a.InputKeys {
			spec := a.Inputs[key]
			if v, ok := withArgs[key]; ok {
				resolved[key] = v
			} else if spec.Default != "" {
				resolved[key] = spec.Default
			} else if spec.IsRequired() {
				return nil, fmt.Errorf("required input %q is missing", key)
			}
		}
	} else {
		for i, key := range a.InputKeys {
			spec := a.Inputs[key]
			if i < len(positionalArgs) {
				resolved[key] = positionalArgs[i]
			} else if spec.Default != "" {
				resolved[key] = spec.Default
			} else if spec.IsRequired() {
				return nil, fmt.Errorf("required input %q is missing (position %d)", key, i+1)
			}
		}
		if len(positionalArgs) > len(a.InputKeys) {
			extra := positionalArgs[len(a.InputKeys):]
			return nil, fmt.Errorf("too many arguments: got %d, but %q only accepts %d input(s); extra: %s",
				len(positionalArgs), a.Name, len(a.InputKeys), strings.Join(extra, " "))
		}
	}

	return resolved, nil
}

// InputEnvVars converts resolved input values to env var format.
// Both PI_IN_<NAME> (canonical) and PI_INPUT_<NAME> (deprecated) are set
// for each input to maintain backward compatibility.
// Keys are sorted alphabetically for deterministic output.
func InputEnvVars(resolved map[string]string) []string {
	if len(resolved) == 0 {
		return nil
	}
	keys := make([]string, 0, len(resolved))
	for k := range resolved {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vars := make([]string, 0, len(resolved)*2)
	for _, k := range keys {
		suffix := strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		val := resolved[k]
		vars = append(vars, "PI_IN_"+suffix+"="+val)
		vars = append(vars, "PI_INPUT_"+suffix+"="+val)
	}
	return vars
}
