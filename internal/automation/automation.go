package automation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/conditions"
	"gopkg.in/yaml.v3"
)

// StepType enumerates the supported step types.
type StepType string

const (
	StepTypeBash       StepType = "bash"
	StepTypeRun        StepType = "run"
	StepTypePython     StepType = "python"
	StepTypeTypeScript StepType = "typescript"
)

var supportedStepTypes = map[StepType]bool{
	StepTypeBash:       true,
	StepTypeRun:        true,
	StepTypePython:     true,
	StepTypeTypeScript: true,
}

var implementedStepTypes = map[StepType]bool{
	StepTypeBash:       true,
	StepTypeRun:        true,
	StepTypePython:     true,
	StepTypeTypeScript: true,
}

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

// Step represents a single step within an automation.
type Step struct {
	Type   StepType          `yaml:"-"`
	Value  string            `yaml:"-"`
	PipeTo string            `yaml:"pipe_to"`
	With   map[string]string `yaml:"-"`
	If     string            `yaml:"-"`
}

// Automation represents a parsed automation YAML file.
type Automation struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`

	// If contains a boolean condition expression. When present, the entire
	// automation is skipped if the expression evaluates to false.
	If string `yaml:"-"`

	// Inputs declares the parameters this automation accepts.
	// Keys are ordered by insertion (YAML parse order) for positional mapping.
	Inputs    map[string]InputSpec `yaml:"-"`
	InputKeys []string             `yaml:"-"`

	// FilePath is the absolute path to the YAML file this automation was loaded from.
	// Set by Load(), not parsed from YAML.
	FilePath string `yaml:"-"`
}

// stepRaw is the intermediate representation used during YAML unmarshalling.
// Each step is a mapping that may contain one of the step type keys.
type stepRaw struct {
	Bash       *string           `yaml:"bash"`
	Run        *string           `yaml:"run"`
	Python     *string           `yaml:"python"`
	TypeScript *string           `yaml:"typescript"`
	PipeTo     string            `yaml:"pipe_to"`
	With       map[string]string `yaml:"with"`
	If         string            `yaml:"if"`
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

// UnmarshalYAML implements custom unmarshalling for Automation to handle
// the polymorphic step type and ordered inputs.
func (a *Automation) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Name        string     `yaml:"name"`
		Description string     `yaml:"description"`
		Steps       []stepRaw  `yaml:"steps"`
		Inputs      *inputsRaw `yaml:"inputs"`
		If          string     `yaml:"if"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	a.Name = raw.Name
	a.Description = raw.Description
	a.If = raw.If

	if raw.Inputs != nil {
		a.Inputs = raw.Inputs.Specs
		a.InputKeys = raw.Inputs.Keys
	}

	for i, sr := range raw.Steps {
		step, err := sr.toStep(i)
		if err != nil {
			return err
		}
		a.Steps = append(a.Steps, step)
	}

	return nil
}

func (sr *stepRaw) toStep(index int) (Step, error) {
	var found []struct {
		t StepType
		v string
	}

	if sr.Bash != nil {
		found = append(found, struct {
			t StepType
			v string
		}{StepTypeBash, *sr.Bash})
	}
	if sr.Run != nil {
		found = append(found, struct {
			t StepType
			v string
		}{StepTypeRun, *sr.Run})
	}
	if sr.Python != nil {
		found = append(found, struct {
			t StepType
			v string
		}{StepTypePython, *sr.Python})
	}
	if sr.TypeScript != nil {
		found = append(found, struct {
			t StepType
			v string
		}{StepTypeTypeScript, *sr.TypeScript})
	}

	if len(found) == 0 {
		return Step{}, fmt.Errorf("step[%d]: must specify one of: bash, run, python, typescript", index)
	}
	if len(found) > 1 {
		return Step{}, fmt.Errorf("step[%d]: must specify exactly one step type, found multiple", index)
	}

	s := found[0]
	if !supportedStepTypes[s.t] {
		return Step{}, fmt.Errorf("step[%d]: unknown step type %q", index, s.t)
	}

	step := Step{
		Type:   s.t,
		Value:  s.v,
		PipeTo: sr.PipeTo,
		If:     sr.If,
	}
	if len(sr.With) > 0 {
		if s.t != StepTypeRun {
			return Step{}, fmt.Errorf("step[%d]: 'with' is only valid on 'run' steps", index)
		}
		step.With = sr.With
	}
	return step, nil
}

// Load reads and parses an automation YAML file at the given path.
func Load(path string) (*Automation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("automation file not found: %s", path)
		}
		return nil, fmt.Errorf("reading automation file %s: %w", path, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path for %s: %w", path, err)
	}

	a := &Automation{}
	if err := yaml.Unmarshal(data, a); err != nil {
		return nil, fmt.Errorf("parsing automation file %s: %w", path, err)
	}

	a.FilePath = absPath

	if err := a.validate(path); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Automation) validate(path string) error {
	if a.Name == "" {
		return fmt.Errorf("%s: \"name\" is required", path)
	}

	if len(a.Steps) == 0 {
		return fmt.Errorf("%s: automation must have at least one step", path)
	}

	if a.If != "" {
		if _, err := conditions.Predicates(a.If); err != nil {
			return fmt.Errorf("%s: invalid if expression: %w", path, err)
		}
	}

	for i, step := range a.Steps {
		if step.Value == "" {
			return fmt.Errorf("%s: step[%d] has empty value", path, i)
		}
		if !implementedStepTypes[step.Type] {
			return fmt.Errorf("%s: step[%d] type %q is recognized but not yet implemented", path, i, step.Type)
		}
		if step.If != "" {
			if _, err := conditions.Predicates(step.If); err != nil {
				return fmt.Errorf("%s: step[%d] invalid if expression: %w", path, i, err)
			}
		}
	}

	return nil
}

// Dir returns the directory containing this automation's YAML file.
func (a *Automation) Dir() string {
	return filepath.Dir(a.FilePath)
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
	}

	return resolved, nil
}

// InputEnvVars converts resolved input values to PI_INPUT_<NAME> env var format.
func InputEnvVars(resolved map[string]string) []string {
	if len(resolved) == 0 {
		return nil
	}
	vars := make([]string, 0, len(resolved))
	for k, v := range resolved {
		envKey := "PI_INPUT_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		vars = append(vars, envKey+"="+v)
	}
	return vars
}

// IsImplemented returns true if the step type is currently implemented in the engine.
func (s StepType) IsImplemented() bool {
	return implementedStepTypes[s]
}
