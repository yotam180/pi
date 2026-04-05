package automation

import (
	"fmt"
	"io"
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
// A step is either a regular step (Type/Value set) or a first: block (First non-nil).
type Step struct {
	Type        StepType          `yaml:"-"`
	Value       string            `yaml:"-"`
	Pipe        bool              `yaml:"-"`
	With        map[string]string `yaml:"-"`
	If          string            `yaml:"-"`
	Env         map[string]string `yaml:"-"`
	Silent      bool              `yaml:"-"`
	ParentShell bool              `yaml:"-"`
	Dir         string            `yaml:"-"`
	Timeout     time.Duration     `yaml:"-"`
	TimeoutRaw  string            `yaml:"-"` // original string for display (e.g. "30s")
	Description string            `yaml:"-"`

	// First holds the sub-steps of a first: block. When non-nil, this step
	// is a first-match block: the executor runs the first sub-step whose if:
	// condition passes and skips the rest. A sub-step without if: always matches.
	First []Step `yaml:"-"`
}

// IsFirst returns true if this step is a first: block.
func (s Step) IsFirst() bool {
	return s.First != nil
}


// stepRaw is the intermediate representation used during YAML unmarshalling.
// Each step is a mapping that may contain one of the step type keys, or a first: block.
// Piping is expressed as pipe: true and/or deprecated pipe_to: next; resolvePipe() normalizes them.
type stepRaw struct {
	Bash        *string           `yaml:"bash"`
	Run         *string           `yaml:"run"`
	Python      *string           `yaml:"python"`
	TypeScript  *string           `yaml:"typescript"`
	PipeTo      string            `yaml:"pipe_to"`
	Pipe        *bool             `yaml:"pipe"`
	With        map[string]string `yaml:"with"`
	If          string            `yaml:"if"`
	Env         map[string]string `yaml:"env"`
	Silent      bool              `yaml:"silent"`
	ParentShell bool              `yaml:"parent_shell"`
	Dir         string            `yaml:"dir"`
	Timeout     string            `yaml:"timeout"`
	Description string            `yaml:"description"`

	// First holds the sub-steps of a first: block (mutually exclusive with step type keys).
	First []stepRaw `yaml:"first"`
}

func (sr *stepRaw) toStep(index int, warnWriter io.Writer) (Step, error) {
	// Handle first: block — mutually exclusive with step type keys.
	// sr.First is non-nil when the YAML key was present (even for `first: []`).
	if sr.First != nil {
		return sr.toFirstStep(index, warnWriter)
	}

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
		return Step{}, fmt.Errorf("step[%d]: must specify one of: bash, run, python, typescript, first", index)
	}
	if len(found) > 1 {
		return Step{}, fmt.Errorf("step[%d]: must specify exactly one step type, found multiple", index)
	}

	s := found[0]
	if !validStepTypes[s.t] {
		return Step{}, fmt.Errorf("step[%d]: unknown step type %q", index, s.t)
	}

	pipe, err := sr.resolvePipe(index, warnWriter)
	if err != nil {
		return Step{}, err
	}

	step := Step{
		Type:        s.t,
		Value:       s.v,
		Pipe:        pipe,
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
		if pipe {
			return Step{}, fmt.Errorf("step[%d]: 'parent_shell' cannot be combined with 'pipe'", index)
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

// toFirstStep converts a stepRaw with a first: block into a Step.
func (sr *stepRaw) toFirstStep(index int, warnWriter io.Writer) (Step, error) {
	// first: is mutually exclusive with step type keys
	if sr.Bash != nil || sr.Run != nil || sr.Python != nil || sr.TypeScript != nil {
		return Step{}, fmt.Errorf("step[%d]: 'first' cannot be combined with a step type key (bash/run/python/typescript)", index)
	}
	// first: does not support env, dir, timeout, silent, parent_shell, with at the block level
	if len(sr.Env) > 0 {
		return Step{}, fmt.Errorf("step[%d]: 'env' is not valid on 'first' blocks (set env on individual sub-steps)", index)
	}
	if sr.Dir != "" {
		return Step{}, fmt.Errorf("step[%d]: 'dir' is not valid on 'first' blocks (set dir on individual sub-steps)", index)
	}
	if sr.Timeout != "" {
		return Step{}, fmt.Errorf("step[%d]: 'timeout' is not valid on 'first' blocks (set timeout on individual sub-steps)", index)
	}
	if sr.Silent {
		return Step{}, fmt.Errorf("step[%d]: 'silent' is not valid on 'first' blocks (set silent on individual sub-steps)", index)
	}
	if sr.ParentShell {
		return Step{}, fmt.Errorf("step[%d]: 'parent_shell' is not valid on 'first' blocks (set parent_shell on individual sub-steps)", index)
	}
	if len(sr.With) > 0 {
		return Step{}, fmt.Errorf("step[%d]: 'with' is not valid on 'first' blocks", index)
	}

	if len(sr.First) == 0 {
		return Step{}, fmt.Errorf("step[%d]: 'first' block must contain at least one sub-step", index)
	}

	var subSteps []Step
	for i, sub := range sr.First {
		s, err := sub.toStep(i, warnWriter)
		if err != nil {
			return Step{}, fmt.Errorf("step[%d].first: %w", index, err)
		}
		subSteps = append(subSteps, s)
	}

	pipe, err := sr.resolvePipe(index, warnWriter)
	if err != nil {
		return Step{}, err
	}

	return Step{
		First:       subSteps,
		Pipe:        pipe,
		If:          sr.If,
		Description: sr.Description,
	}, nil
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
			step, err := sr.toStep(i, nil)
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

// resolvePipe normalises pipe_to: next (deprecated) and pipe: true into a single bool.
func (sr *stepRaw) resolvePipe(index int, warnWriter io.Writer) (bool, error) {
	hasPipeTo := sr.PipeTo != ""
	hasPipe := sr.Pipe != nil

	if hasPipeTo && hasPipe {
		return false, fmt.Errorf("step[%d]: cannot specify both 'pipe' and 'pipe_to' (use 'pipe: true')", index)
	}

	if hasPipeTo {
		if sr.PipeTo != "next" {
			return false, fmt.Errorf("step[%d]: pipe_to must be \"next\", got %q", index, sr.PipeTo)
		}
		if warnWriter != nil {
			fmt.Fprintf(warnWriter, "warning: step[%d]: 'pipe_to: next' is deprecated, use 'pipe: true'\n", index)
		}
		return true, nil
	}

	if hasPipe {
		if !*sr.Pipe {
			return false, nil
		}
		return true, nil
	}

	return false, nil
}

// validateSteps checks that all steps in a slice have valid types and if: expressions.
func validateSteps(path string, steps []Step) error {
	for i, step := range steps {
		if step.IsFirst() {
			if err := validateFirstBlock(path, i, step); err != nil {
				return err
			}
			continue
		}
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

func validateFirstBlock(path string, index int, step Step) error {
	if step.If != "" {
		if _, err := conditions.Predicates(step.If); err != nil {
			return fmt.Errorf("%s: step[%d] invalid if expression: %w", path, index, err)
		}
	}
	for j, sub := range step.First {
		if sub.IsFirst() {
			return fmt.Errorf("%s: step[%d].first[%d]: nested first: blocks are not allowed", path, index, j)
		}
		if sub.Value == "" {
			return fmt.Errorf("%s: step[%d].first[%d] has empty value", path, index, j)
		}
		if !validStepTypes[sub.Type] {
			return fmt.Errorf("%s: step[%d].first[%d] type %q is not a valid step type", path, index, j, sub.Type)
		}
		if sub.If != "" {
			if _, err := conditions.Predicates(sub.If); err != nil {
				return fmt.Errorf("%s: step[%d].first[%d] invalid if expression: %w", path, index, j, err)
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
	prefix := fmt.Sprintf("%s: install.%s", path, phaseName)
	for i, step := range phase.Steps {
		if step.IsFirst() {
			if err := validateFirstBlock(prefix, i, step); err != nil {
				return err
			}
			continue
		}
		if step.Value == "" {
			return fmt.Errorf("%s step[%d] has empty value", prefix, i)
		}
		if !validStepTypes[step.Type] {
			return fmt.Errorf("%s step[%d] type %q is not a valid step type", prefix, i, step.Type)
		}
		if step.If != "" {
			if _, err := conditions.Predicates(step.If); err != nil {
				return fmt.Errorf("%s step[%d] invalid if expression: %w", prefix, i, err)
			}
		}
	}
	return nil
}
