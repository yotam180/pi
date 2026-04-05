package automation

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vyper-tooling/pi/internal/conditions"
	"gopkg.in/yaml.v3"
)

// Automation represents a parsed automation YAML file.
type Automation struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`

	// Install defines the structured installer lifecycle. Mutually exclusive with Steps.
	Install *InstallSpec `yaml:"-"`

	// Requires declares tools and runtimes the automation needs before execution.
	Requires []Requirement `yaml:"-"`

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

// IsInstaller returns true if this automation uses the install: block schema.
func (a *Automation) IsInstaller() bool {
	return a.Install != nil
}

// Dir returns the directory containing this automation's YAML file.
func (a *Automation) Dir() string {
	return filepath.Dir(a.FilePath)
}

// UnmarshalYAML implements custom unmarshalling for Automation to handle
// the polymorphic step type, ordered inputs, install: blocks, and
// single-step shorthand (top-level bash:/python:/typescript:/run: keys).
func (a *Automation) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Name        string           `yaml:"name"`
		Description string           `yaml:"description"`
		Steps       []stepRaw        `yaml:"steps"`
		Install     *InstallSpec     `yaml:"install"`
		Inputs      *inputsRaw       `yaml:"inputs"`
		If          string           `yaml:"if"`
		Requires    []requirementRaw `yaml:"requires"`

		// Single-step shorthand: top-level step type keys
		Bash       *string           `yaml:"bash"`
		Python     *string           `yaml:"python"`
		TypeScript *string           `yaml:"typescript"`
		Run        *string           `yaml:"run"`
		Env        map[string]string `yaml:"env"`
		Dir        string            `yaml:"dir"`
		Timeout    string            `yaml:"timeout"`
		Silent     bool              `yaml:"silent"`
		PipeTo     string            `yaml:"pipe_to"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	a.Name = raw.Name
	a.Description = raw.Description
	a.If = raw.If
	a.Install = raw.Install

	if raw.Inputs != nil {
		a.Inputs = raw.Inputs.Specs
		a.InputKeys = raw.Inputs.Keys
	}

	for _, rr := range raw.Requires {
		a.Requires = append(a.Requires, rr.Requirement)
	}

	// Detect single-step shorthand (top-level bash:/python:/typescript:/run:)
	shorthand, err := buildShorthandStep(raw.Bash, raw.Python, raw.TypeScript, raw.Run)
	if err != nil {
		return err
	}
	if shorthand != nil {
		if len(raw.Steps) > 0 {
			return fmt.Errorf("automation cannot have both a top-level step key (bash/python/typescript/run) and \"steps\"")
		}
		if raw.Install != nil {
			return fmt.Errorf("automation cannot have both a top-level step key (bash/python/typescript/run) and \"install\"")
		}
		shorthand.Env = raw.Env
		shorthand.Dir = raw.Dir
		shorthand.Timeout = raw.Timeout
		shorthand.Silent = raw.Silent
		shorthand.PipeTo = raw.PipeTo
		step, err := shorthand.toStep(0)
		if err != nil {
			return err
		}
		a.Steps = append(a.Steps, step)
		return nil
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

// buildShorthandStep checks for top-level step keys and returns a stepRaw if
// exactly one is present. Returns (nil, nil) if none are present. Returns an
// error if multiple top-level step keys are present.
func buildShorthandStep(bash, python, typescript, run *string) (*stepRaw, error) {
	var count int
	if bash != nil {
		count++
	}
	if python != nil {
		count++
	}
	if typescript != nil {
		count++
	}
	if run != nil {
		count++
	}

	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		return nil, fmt.Errorf("automation has multiple top-level step keys (bash/python/typescript/run); use exactly one")
	}

	sr := &stepRaw{
		Bash:       bash,
		Python:     python,
		TypeScript: typescript,
		Run:        run,
	}
	return sr, nil
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

// LoadFromBytes parses automation YAML from raw bytes, using filePath as the
// automation's logical file path (for Dir() resolution and error messages).
func LoadFromBytes(data []byte, filePath string) (*Automation, error) {
	a := &Automation{}
	if err := yaml.Unmarshal(data, a); err != nil {
		return nil, fmt.Errorf("parsing automation %s: %w", filePath, err)
	}

	a.FilePath = filePath

	if err := a.validate(filePath); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Automation) validate(path string) error {
	hasSteps := len(a.Steps) > 0
	hasInstall := a.Install != nil

	if hasSteps && hasInstall {
		return fmt.Errorf("%s: automation cannot have both \"steps\" and \"install\"", path)
	}

	if !hasSteps && !hasInstall {
		return fmt.Errorf("%s: automation must have \"steps\", \"install\", or a top-level step key (bash/python/typescript/run)", path)
	}

	if a.If != "" {
		if _, err := conditions.Predicates(a.If); err != nil {
			return fmt.Errorf("%s: invalid if expression: %w", path, err)
		}
	}

	if hasInstall {
		return validateInstall(path, a.Install)
	}

	return validateSteps(path, a.Steps)
}
