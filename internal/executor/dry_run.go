package executor

import (
	"fmt"
	"strings"

	"github.com/vyper-tooling/pi/internal/automation"
)

// dryRunAutomation prints the steps that would be executed without running them.
// Conditions are evaluated, run: targets are resolved, and installers show
// their lifecycle — but no commands run and no side effects occur.
func (e *Executor) dryRunAutomation(a *automation.Automation, args []string, inputEnv []string) error {
	if a.IsGoFunc() {
		e.dryRunPrint(0, "go-func", a.Name, "")
		return nil
	}

	if a.IsInstaller() {
		return e.dryRunInstaller(a, inputEnv, 0)
	}

	ctx := &stepExecCtx{
		automation:    a,
		args:          args,
		inputEnv:      inputEnv,
		automationEnv: e.interpolateEnv(a.Env, inputEnv),
	}

	for i, step := range a.Steps {
		if err := e.dryRunStep(ctx, step, i, 0); err != nil {
			return err
		}
	}
	return nil
}

// dryRunStep prints a single step (or first: block) at a given indentation depth.
func (e *Executor) dryRunStep(ctx *stepExecCtx, step automation.Step, index int, depth int) error {
	if step.If != "" {
		skip, err := e.evaluateCondition(step.If)
		if err != nil {
			return fmt.Errorf("automation %q step[%d] if: %w", ctx.automation.Name, index, err)
		}
		if skip {
			e.dryRunPrint(depth, string(step.Type), expandTraceVars(step.Value, ctx.inputEnv, ctx.automationEnv, step.Env), fmt.Sprintf("skipped (if: %s)", step.If))
			return nil
		}
	}

	if step.IsFirst() {
		return e.dryRunFirstBlock(ctx, step, index, depth)
	}

	traceValue := expandTraceVars(step.Value, ctx.inputEnv, ctx.automationEnv, step.Env)

	var annotations []string
	if step.If != "" {
		annotations = append(annotations, fmt.Sprintf("if: %s", step.If))
	}
	if step.Dir != "" {
		annotations = append(annotations, fmt.Sprintf("dir: %s", step.Dir))
	}
	if step.TimeoutRaw != "" {
		annotations = append(annotations, fmt.Sprintf("timeout: %s", step.TimeoutRaw))
	}
	if step.Silent {
		annotations = append(annotations, "silent")
	}
	if step.ParentShell {
		annotations = append(annotations, "parent_shell")
	}
	if step.Pipe {
		annotations = append(annotations, "pipe")
	}

	annotation := ""
	if len(annotations) > 0 {
		annotation = strings.Join(annotations, ", ")
	}

	stepType := string(step.Type)
	if step.ParentShell {
		stepType = "parent"
	}

	e.dryRunPrint(depth, stepType, traceValue, annotation)

	if step.Type == automation.StepTypeRun {
		return e.dryRunRunStep(ctx, step, depth)
	}

	return nil
}

// dryRunRunStep resolves and recurses into a run: target.
func (e *Executor) dryRunRunStep(ctx *stepExecCtx, step automation.Step, depth int) error {
	target, err := e.Discovery.Find(step.Value)
	if err != nil {
		return nil
	}

	if err := e.pushCall(target.Name); err != nil {
		e.dryRunPrint(depth+1, "", "(circular dependency — would fail at runtime)", "")
		e.popCall()
		return nil
	}

	var targetInputEnv []string
	if len(step.With) > 0 {
		with := e.interpolateWithCtx(step.With, ctx.inputEnv)
		resolvedInputs, err := target.ResolveInputs(with, nil)
		if err == nil {
			targetInputEnv = automation.InputEnvVars(resolvedInputs)
		}
	} else if len(ctx.args) > 0 && len(target.Inputs) > 0 {
		resolvedInputs, err := target.ResolveInputs(nil, ctx.args)
		if err == nil {
			targetInputEnv = automation.InputEnvVars(resolvedInputs)
		}
	}

	targetCtx := &stepExecCtx{
		automation:    target,
		args:          ctx.args,
		inputEnv:      targetInputEnv,
		automationEnv: e.interpolateEnv(target.Env, targetInputEnv),
	}

	if target.IsInstaller() {
		err := e.dryRunInstaller(target, targetInputEnv, depth+1)
		e.popCall()
		return err
	}

	if target.IsGoFunc() {
		e.dryRunPrint(depth+1, "go-func", target.Name, "")
		e.popCall()
		return nil
	}

	for i, s := range target.Steps {
		if err := e.dryRunStep(targetCtx, s, i, depth+1); err != nil {
			e.popCall()
			return err
		}
	}

	e.popCall()
	return nil
}

// dryRunFirstBlock prints a first: block showing all sub-steps and which would match.
func (e *Executor) dryRunFirstBlock(ctx *stepExecCtx, step automation.Step, index int, depth int) error {
	e.dryRunPrint(depth, "first", "", "")
	matched := false
	for j, sub := range step.First {
		if sub.If != "" {
			skip, err := e.evaluateCondition(sub.If)
			if err != nil {
				return fmt.Errorf("automation %q step[%d].first[%d] if: %w", ctx.automation.Name, index, j, err)
			}
			if skip {
				e.dryRunPrint(depth+1, string(sub.Type), expandTraceVars(sub.Value, ctx.inputEnv, ctx.automationEnv, sub.Env), fmt.Sprintf("skipped (if: %s)", sub.If))
				continue
			}
		}

		if matched {
			e.dryRunPrint(depth+1, string(sub.Type), expandTraceVars(sub.Value, ctx.inputEnv, ctx.automationEnv, sub.Env), "not reached (earlier match)")
			continue
		}

		marker := "← match"
		if sub.If != "" {
			marker = fmt.Sprintf("← match (if: %s)", sub.If)
		}
		e.dryRunPrint(depth+1, string(sub.Type), expandTraceVars(sub.Value, ctx.inputEnv, ctx.automationEnv, sub.Env), marker)

		if sub.Type == automation.StepTypeRun {
			if err := e.dryRunRunStep(ctx, sub, depth+1); err != nil {
				return err
			}
		}

		matched = true
	}
	return nil
}

// dryRunInstaller prints the installer lifecycle phases.
func (e *Executor) dryRunInstaller(a *automation.Automation, inputEnv []string, depth int) error {
	inst := a.Install

	ctx := &stepExecCtx{
		automation:    a,
		inputEnv:      inputEnv,
		automationEnv: e.interpolateEnv(a.Env, inputEnv),
	}

	e.dryRunPrint(depth, "install", a.Name, "")

	e.dryRunPrint(depth+1, "test", "", "")
	if err := e.dryRunInstallPhase(ctx, &inst.Test, depth+2); err != nil {
		return err
	}

	e.dryRunPrint(depth+1, "run", "", "")
	if err := e.dryRunInstallPhase(ctx, &inst.Run, depth+2); err != nil {
		return err
	}

	if inst.Verify != nil {
		e.dryRunPrint(depth+1, "verify", "", "")
		if err := e.dryRunInstallPhase(ctx, inst.Verify, depth+2); err != nil {
			return err
		}
	} else {
		e.dryRunPrint(depth+1, "verify", "(re-runs test)", "")
	}

	if inst.Version != "" {
		e.dryRunPrint(depth+1, "version", inst.Version, "")
	}

	return nil
}

// dryRunInstallPhase prints steps within an install phase.
func (e *Executor) dryRunInstallPhase(ctx *stepExecCtx, phase *automation.InstallPhase, depth int) error {
	if phase.IsScalar {
		e.dryRunPrint(depth, "bash", phase.Scalar, "")
		return nil
	}
	for i, step := range phase.Steps {
		if err := e.dryRunStep(ctx, step, i, depth); err != nil {
			return err
		}
	}
	return nil
}

// dryRunPrint writes one dry-run output line at the given indentation depth.
func (e *Executor) dryRunPrint(depth int, stepType, value, annotation string) {
	indent := strings.Repeat("  ", depth)
	w := e.stderr()

	if stepType == "" && value != "" {
		fmt.Fprintf(w, "%s  %s\n", indent, value)
		return
	}

	truncated := dryRunTruncate(value, 80)

	if annotation != "" {
		if truncated != "" {
			fmt.Fprintf(w, "%s  → %s: %s  [%s]\n", indent, stepType, truncated, annotation)
		} else {
			fmt.Fprintf(w, "%s  → %s  [%s]\n", indent, stepType, annotation)
		}
	} else {
		if truncated != "" {
			fmt.Fprintf(w, "%s  → %s: %s\n", indent, stepType, truncated)
		} else {
			fmt.Fprintf(w, "%s  → %s\n", indent, stepType)
		}
	}
}

func dryRunTruncate(s string, maxLen int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i] + "..."
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
