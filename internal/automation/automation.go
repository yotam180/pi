package automation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vyper-tooling/pi/internal/conditions"
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
		return fmt.Errorf("requires entry %q: unknown runtime %q (known: python, node); use \"command: %s\" for arbitrary commands", s, name, name)
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
	Type        StepType          `yaml:"-"`
	Value       string            `yaml:"-"`
	PipeTo      string            `yaml:"pipe_to"`
	With        map[string]string `yaml:"-"`
	If          string            `yaml:"-"`
	Env         map[string]string `yaml:"-"`
	Silent      bool              `yaml:"-"`
	ParentShell bool              `yaml:"-"`
	Dir         string            `yaml:"-"`
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
	Test    InstallPhase `yaml:"test"`
	Run     InstallPhase `yaml:"run"`
	Verify  *InstallPhase `yaml:"verify"`
	Version string       `yaml:"version"`
}

// HasVerify returns true if an explicit verify phase was declared.
func (s *InstallSpec) HasVerify() bool {
	return s.Verify != nil
}

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
// the polymorphic step type, ordered inputs, and install: blocks.
func (a *Automation) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Name        string           `yaml:"name"`
		Description string           `yaml:"description"`
		Steps       []stepRaw        `yaml:"steps"`
		Install     *InstallSpec     `yaml:"install"`
		Inputs      *inputsRaw       `yaml:"inputs"`
		If          string           `yaml:"if"`
		Requires    []requirementRaw `yaml:"requires"`
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
		Type:        s.t,
		Value:       s.v,
		PipeTo:      sr.PipeTo,
		If:          sr.If,
		Env:         sr.Env,
		Silent:      sr.Silent,
		ParentShell: sr.ParentShell,
		Dir:         sr.Dir,
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

	hasSteps := len(a.Steps) > 0
	hasInstall := a.Install != nil

	if hasSteps && hasInstall {
		return fmt.Errorf("%s: automation cannot have both \"steps\" and \"install\"", path)
	}

	if !hasSteps && !hasInstall {
		return fmt.Errorf("%s: automation must have \"steps\" or \"install\"", path)
	}

	if a.If != "" {
		if _, err := conditions.Predicates(a.If); err != nil {
			return fmt.Errorf("%s: invalid if expression: %w", path, err)
		}
	}

	if hasInstall {
		return a.validateInstall(path)
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

func (a *Automation) validateInstall(path string) error {
	inst := a.Install

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
		if !implementedStepTypes[step.Type] {
			return fmt.Errorf("%s: install.%s step[%d] type %q is not implemented", path, phaseName, i, step.Type)
		}
		if step.If != "" {
			if _, err := conditions.Predicates(step.If); err != nil {
				return fmt.Errorf("%s: install.%s step[%d] invalid if expression: %w", path, phaseName, i, err)
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

// IsImplemented returns true if the step type is currently implemented in the engine.
func (s StepType) IsImplemented() bool {
	return implementedStepTypes[s]
}
