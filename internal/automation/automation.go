package automation

import (
	"fmt"
	"os"
	"path/filepath"

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
	StepTypeBash:   true,
	StepTypeRun:    true,
	StepTypePython: true,
}

// Step represents a single step within an automation.
type Step struct {
	Type   StepType `yaml:"-"`
	Value  string   `yaml:"-"`
	PipeTo string   `yaml:"pipe_to"`
}

// Automation represents a parsed automation YAML file.
type Automation struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`

	// FilePath is the absolute path to the YAML file this automation was loaded from.
	// Set by Load(), not parsed from YAML.
	FilePath string `yaml:"-"`
}

// stepRaw is the intermediate representation used during YAML unmarshalling.
// Each step is a mapping that may contain one of the step type keys.
type stepRaw struct {
	Bash       *string `yaml:"bash"`
	Run        *string `yaml:"run"`
	Python     *string `yaml:"python"`
	TypeScript *string `yaml:"typescript"`
	PipeTo     string  `yaml:"pipe_to"`
}

// UnmarshalYAML implements custom unmarshalling for Automation to handle
// the polymorphic step type.
func (a *Automation) UnmarshalYAML(value *yaml.Node) error {
	// Decode into a struct that captures everything except steps literally.
	var raw struct {
		Name        string    `yaml:"name"`
		Description string    `yaml:"description"`
		Steps       []stepRaw `yaml:"steps"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	a.Name = raw.Name
	a.Description = raw.Description

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

	return Step{
		Type:   s.t,
		Value:  s.v,
		PipeTo: sr.PipeTo,
	}, nil
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

	for i, step := range a.Steps {
		if step.Value == "" {
			return fmt.Errorf("%s: step[%d] has empty value", path, i)
		}
		if !implementedStepTypes[step.Type] {
			return fmt.Errorf("%s: step[%d] type %q is recognized but not yet implemented", path, i, step.Type)
		}
	}

	return nil
}

// Dir returns the directory containing this automation's YAML file.
func (a *Automation) Dir() string {
	return filepath.Dir(a.FilePath)
}

// IsImplemented returns true if the step type is currently implemented in the engine.
func (s StepType) IsImplemented() bool {
	return implementedStepTypes[s]
}
