package automation

import (
	"fmt"
	"time"

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

var validStepTypes = map[StepType]bool{
	StepTypeBash:       true,
	StepTypeRun:        true,
	StepTypePython:     true,
	StepTypeTypeScript: true,
}

// IsValid returns true if the step type is a recognized, valid step type.
func (s StepType) IsValid() bool {
	return validStepTypes[s]
}

// Step represents a single step within an automation.
type Step struct {
	Type        StepType          `yaml:"-"`
	Value       string            `yaml:"-"`
	PipeTo      string            `yaml:"pipe_to"`
	With        map[string]string `yaml:"-"`
	If          string            `yaml:"-"`
	Env         map[string]string `yaml:"-"`
	Silent      bool              `yaml:"-"`
	ParentShell bool              `yaml:"-"`
	Dir         string            `yaml:"-"`
	Timeout     time.Duration     `yaml:"-"`
	TimeoutRaw  string            `yaml:"-"` // original string for display (e.g. "30s")
	Description string            `yaml:"-"`
}

// stepRaw is the intermediate representation used during YAML unmarshalling.
// Each step is a mapping that may contain one of the step type keys.
type stepRaw struct {
	Bash        *string           `yaml:"bash"`
	Run         *string           `yaml:"run"`
	Python      *string           `yaml:"python"`
	TypeScript  *string           `yaml:"typescript"`
	PipeTo      string            `yaml:"pipe_to"`
	With        map[string]string `yaml:"with"`
	If          string            `yaml:"if"`
	Env         map[string]string `yaml:"env"`
	Silent      bool              `yaml:"silent"`
	ParentShell bool              `yaml:"parent_shell"`
	Dir         string            `yaml:"dir"`
	Timeout     string            `yaml:"timeout"`
	Description string            `yaml:"description"`
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
	if !validStepTypes[s.t] {
		return Step{}, fmt.Errorf("step[%d]: unknown step type %q", index, s.t)
	}

	step := Step{
		Type:        s.t,
		Value:       s.v,
		PipeTo:      sr.PipeTo,
		If:          sr.If,
		Env:         sr.Env,
		Silent:      sr.Silent,
		ParentShell: sr.ParentShell,
		Dir:         sr.Dir,
		Description: sr.Description,
	}
	if len(sr.With) > 0 {
		if s.t != StepTypeRun {
			return Step{}, fmt.Errorf("step[%d]: 'with' is only valid on 'run' steps", index)
		}
		step.With = sr.With
	}
	if sr.ParentShell {
		if s.t != StepTypeBash {
			return Step{}, fmt.Errorf("step[%d]: 'parent_shell' is only valid on 'bash' steps", index)
		}
		if sr.PipeTo != "" {
			return Step{}, fmt.Errorf("step[%d]: 'parent_shell' cannot be combined with 'pipe_to'", index)
		}
	}
	if sr.Timeout != "" {
		if s.t == StepTypeRun {
			return Step{}, fmt.Errorf("step[%d]: 'timeout' is not valid on 'run' steps (set timeouts on the target automation's steps instead)", index)
		}
		if sr.ParentShell {
			return Step{}, fmt.Errorf("step[%d]: 'timeout' cannot be combined with 'parent_shell'", index)
		}
		d, err := time.ParseDuration(sr.Timeout)
		if err != nil {
			return Step{}, fmt.Errorf("step[%d]: invalid timeout %q: %w", index, sr.Timeout, err)
		}
		if d <= 0 {
			return Step{}, fmt.Errorf("step[%d]: timeout must be positive, got %q", index, sr.Timeout)
		}
		step.Timeout = d
		step.TimeoutRaw = sr.Timeout
	}
	return step, nil
}

// InstallPhase represents a phase (test, run, verify) in an install: block.
// It can be either a single bash string or a list of steps.
type InstallPhase struct {
	IsScalar bool
	Scalar   string
	Steps    []Step
}

// UnmarshalYAML handles both scalar strings and step lists for install phases.
func (p *InstallPhase) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		p.IsScalar = true
		p.Scalar = value.Value
		return nil
	}
	if value.Kind == yaml.SequenceNode {
		p.IsScalar = false
		var rawSteps []stepRaw
		if err := value.Decode(&rawSteps); err != nil {
			return err
		}
		for i, sr := range rawSteps {
			step, err := sr.toStep(i)
			if err != nil {
				return err
			}
			p.Steps = append(p.Steps, step)
		}
		return nil
	}
	return fmt.Errorf("install phase must be a string or list of steps")
}

// InstallSpec defines the structured installer lifecycle.
type InstallSpec struct {
	Test    InstallPhase  `yaml:"test"`
	Run     InstallPhase  `yaml:"run"`
	Verify  *InstallPhase `yaml:"verify"`
	Version string        `yaml:"version"`
}

// HasVerify returns true if an explicit verify phase was declared.
func (s *InstallSpec) HasVerify() bool {
	return s.Verify != nil
}

// validateSteps checks that all steps in a slice have valid types and if: expressions.
func validateSteps(path string, steps []Step) error {
	for i, step := range steps {
		if step.Value == "" {
			return fmt.Errorf("%s: step[%d] has empty value", path, i)
		}
		if !validStepTypes[step.Type] {
			return fmt.Errorf("%s: step[%d] type %q is not a valid step type", path, i, step.Type)
		}
		if step.If != "" {
			if _, err := conditions.Predicates(step.If); err != nil {
				return fmt.Errorf("%s: step[%d] invalid if expression: %w", path, i, err)
			}
		}
	}
	return nil
}

// validateInstall validates the install: block of an automation.
func validateInstall(path string, inst *InstallSpec) error {
	if err := validateInstallPhase(path, "test", &inst.Test); err != nil {
		return err
	}
	if err := validateInstallPhase(path, "run", &inst.Run); err != nil {
		return err
	}
	if inst.Verify != nil {
		if err := validateInstallPhase(path, "verify", inst.Verify); err != nil {
			return err
		}
	}

	if !inst.Test.IsScalar && len(inst.Test.Steps) == 0 {
		return fmt.Errorf("%s: install.test must have content", path)
	}
	if inst.Test.IsScalar && inst.Test.Scalar == "" {
		return fmt.Errorf("%s: install.test must have content", path)
	}
	if !inst.Run.IsScalar && len(inst.Run.Steps) == 0 {
		return fmt.Errorf("%s: install.run must have content", path)
	}
	if inst.Run.IsScalar && inst.Run.Scalar == "" {
		return fmt.Errorf("%s: install.run must have content", path)
	}

	return nil
}

func validateInstallPhase(path, phaseName string, phase *InstallPhase) error {
	if phase.IsScalar {
		return nil
	}
	for i, step := range phase.Steps {
		if step.Value == "" {
			return fmt.Errorf("%s: install.%s step[%d] has empty value", path, phaseName, i)
		}
		if !validStepTypes[step.Type] {
			return fmt.Errorf("%s: install.%s step[%d] type %q is not a valid step type", path, phaseName, i, step.Type)
		}
		if step.If != "" {
			if _, err := conditions.Predicates(step.If); err != nil {
				return fmt.Errorf("%s: install.%s step[%d] invalid if expression: %w", path, phaseName, i, err)
			}
		}
	}
	return nil
}
