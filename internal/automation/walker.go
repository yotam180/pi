package automation

import "fmt"

// StepLocation describes where a step lives within an automation.
type StepLocation struct {
	// Phase is empty for regular steps, or "test"/"run"/"verify" for install phases.
	Phase string

	// Index is the 0-based step index within the step list or install phase.
	Index int

	// FirstIndex is the 0-based sub-step index within a first: block,
	// or -1 if the step is not inside a first: block.
	FirstIndex int

	// IsScalar is true when the step was synthesized from a scalar install phase.
	IsScalar bool
}

// InFirstBlock returns true if this location is inside a first: block.
func (loc StepLocation) InFirstBlock() bool {
	return loc.FirstIndex >= 0
}

// FormatPath returns a human-readable location string for error messages.
// Examples: "my-auto: step[2]", "my-auto: step[1].first[0]",
// "my-auto: install.run step[0]", "my-auto: install.test".
func (loc StepLocation) FormatPath(automationName string) string {
	if loc.IsScalar {
		return fmt.Sprintf("%s: install.%s", automationName, loc.Phase)
	}
	var prefix string
	if loc.Phase != "" {
		prefix = fmt.Sprintf("%s: install.%s ", automationName, loc.Phase)
	} else {
		prefix = fmt.Sprintf("%s: ", automationName)
	}
	if loc.InFirstBlock() {
		return fmt.Sprintf("%sstep[%d].first[%d]", prefix, loc.Index, loc.FirstIndex)
	}
	return fmt.Sprintf("%sstep[%d]", prefix, loc.Index)
}

// StepVisitor is called for each step encountered during a walk.
// It receives the step and its location within the automation.
type StepVisitor func(step Step, loc StepLocation)

// WalkSteps visits every step in an automation, including steps inside
// first: blocks and install phases. For scalar install phases, a synthetic
// Step{Type: StepTypeBash, Value: phase.Scalar} is generated.
//
// Visit order: regular steps (with first: sub-steps expanded inline),
// then install phases in test → run → verify order.
func WalkSteps(a *Automation, fn StepVisitor) {
	walkStepSlice(a.Steps, "", fn)

	if a.Install == nil {
		return
	}
	walkInstallPhase(&a.Install.Test, "test", fn)
	walkInstallPhase(&a.Install.Run, "run", fn)
	if a.Install.Verify != nil {
		walkInstallPhase(a.Install.Verify, "verify", fn)
	}
}

// WalkStepsUntil is like WalkSteps but stops early when the visitor returns true.
func WalkStepsUntil(a *Automation, fn func(step Step, loc StepLocation) bool) {
	if walkStepSliceUntil(a.Steps, "", fn) {
		return
	}

	if a.Install == nil {
		return
	}
	if walkInstallPhaseUntil(&a.Install.Test, "test", fn) {
		return
	}
	if walkInstallPhaseUntil(&a.Install.Run, "run", fn) {
		return
	}
	if a.Install.Verify != nil {
		walkInstallPhaseUntil(a.Install.Verify, "verify", fn)
	}
}

func walkStepSlice(steps []Step, phase string, fn StepVisitor) {
	for i, step := range steps {
		if step.IsFirst() {
			for j, sub := range step.First {
				fn(sub, StepLocation{Phase: phase, Index: i, FirstIndex: j})
			}
			continue
		}
		fn(step, StepLocation{Phase: phase, Index: i, FirstIndex: -1})
	}
}

func walkInstallPhase(phase *InstallPhase, name string, fn StepVisitor) {
	if phase.IsScalar {
		fn(
			Step{Type: StepTypeBash, Value: phase.Scalar},
			StepLocation{Phase: name, Index: 0, FirstIndex: -1, IsScalar: true},
		)
		return
	}
	walkStepSlice(phase.Steps, name, fn)
}

func walkStepSliceUntil(steps []Step, phase string, fn func(Step, StepLocation) bool) bool {
	for i, step := range steps {
		if step.IsFirst() {
			for j, sub := range step.First {
				if fn(sub, StepLocation{Phase: phase, Index: i, FirstIndex: j}) {
					return true
				}
			}
			continue
		}
		if fn(step, StepLocation{Phase: phase, Index: i, FirstIndex: -1}) {
			return true
		}
	}
	return false
}

func walkInstallPhaseUntil(phase *InstallPhase, name string, fn func(Step, StepLocation) bool) bool {
	if phase.IsScalar {
		return fn(
			Step{Type: StepTypeBash, Value: phase.Scalar},
			StepLocation{Phase: name, Index: 0, FirstIndex: -1, IsScalar: true},
		)
	}
	return walkStepSliceUntil(phase.Steps, name, fn)
}
