package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <automation>",
		Short: "Show details about an automation",
		Long: `Print the name, description, and input documentation for a given automation.
PI walks up from the current directory to find pi.yaml, so this works from
any subdirectory of the project.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return showAutomationInfo(cwd, args[0], cmd.OutOrStdout())
		},
	}
}

// showAutomationInfo resolves an automation by name and prints its details.
func showAutomationInfo(startDir, name string, out io.Writer) error {
	pc, err := resolveProject(startDir)
	if err != nil {
		return err
	}

	result, err := pc.Discover(nil)
	if err != nil {
		return err
	}

	a, err := result.Find(name)
	if err != nil {
		return err
	}

	printAutomationInfo(a, out)
	return nil
}

func printAutomationInfo(a *automation.Automation, out io.Writer) {
	fmt.Fprintf(out, "Name:         %s\n", a.Name)

	desc := a.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Fprintf(out, "Description:  %s\n", desc)

	if a.If != "" {
		fmt.Fprintf(out, "Condition:    %s\n", a.If)
	}

	if len(a.Env) > 0 {
		envKeys := make([]string, 0, len(a.Env))
		for k := range a.Env {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)
		fmt.Fprintf(out, "Env:          %s\n", strings.Join(envKeys, ", "))
	}

	if a.IsInstaller() {
		fmt.Fprintf(out, "Type:         installer\n")
		printInstallDetail(a.Install, out)
	} else {
		fmt.Fprintf(out, "Steps:        %d\n", len(a.Steps))
		printStepsDetail(a.Steps, out)
	}

	if len(a.InputKeys) == 0 {
		fmt.Fprintf(out, "\nNo inputs.\n")
		return
	}

	fmt.Fprintf(out, "\nInputs:\n")
	for _, key := range a.InputKeys {
		spec := a.Inputs[key]
		printInputSpec(key, spec, out)
	}
}

func printInstallDetail(inst *automation.InstallSpec, out io.Writer) {
	printPhaseInfo := func(name string, phase *automation.InstallPhase) {
		if phase.IsScalar {
			fmt.Fprintf(out, "  %s: %s\n", name, truncateValue(phase.Scalar, 60))
		} else {
			fmt.Fprintf(out, "  %s: %d step(s)\n", name, len(phase.Steps))
		}
	}

	fmt.Fprintf(out, "\nInstall lifecycle:\n")
	printPhaseInfo("test", &inst.Test)
	printPhaseInfo("run", &inst.Run)
	if inst.HasVerify() {
		printPhaseInfo("verify", inst.Verify)
	} else {
		fmt.Fprintf(out, "  verify: (re-runs test)\n")
	}
	if inst.Version != "" {
		fmt.Fprintf(out, "  version: %s\n", truncateValue(inst.Version, 60))
	}
}

// stepAnnotations returns the display annotations for a step (if, pipe, silent,
// parent_shell, dir, timeout, env). Used by both printStepsDetail and
// printFirstBlockDetail to avoid duplicating the annotation-building logic.
func stepAnnotations(s automation.Step) []string {
	var annotations []string
	if s.If != "" {
		annotations = append(annotations, fmt.Sprintf("if: %s", s.If))
	}
	if s.Pipe {
		annotations = append(annotations, "pipe")
	}
	if s.Silent {
		annotations = append(annotations, "silent")
	}
	if s.ParentShell {
		annotations = append(annotations, "parent_shell")
	}
	if s.Dir != "" {
		annotations = append(annotations, fmt.Sprintf("dir: %s", s.Dir))
	}
	if s.Timeout > 0 {
		annotations = append(annotations, fmt.Sprintf("timeout: %s", s.TimeoutRaw))
	}
	if len(s.Env) > 0 {
		envKeys := make([]string, 0, len(s.Env))
		for k := range s.Env {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)
		annotations = append(annotations, fmt.Sprintf("env: %s", strings.Join(envKeys, ", ")))
	}
	return annotations
}

func printStepsDetail(steps []automation.Step, out io.Writer) {
	hasDetails := false
	for _, s := range steps {
		if s.IsFirst() || s.If != "" || len(s.Env) > 0 || s.Silent || s.ParentShell || s.Dir != "" || s.Timeout > 0 || s.Description != "" || s.Pipe {
			hasDetails = true
			break
		}
	}
	if !hasDetails {
		return
	}

	fmt.Fprintf(out, "\nStep details:\n")
	for i, s := range steps {
		if s.IsFirst() {
			printFirstBlockDetail(i, s, out)
			continue
		}
		label := fmt.Sprintf("%s: %s", s.Type, truncateValue(s.Value, 40))
		annotations := stepAnnotations(s)
		if len(annotations) > 0 {
			fmt.Fprintf(out, "  %d. %s  [%s]\n", i+1, label, strings.Join(annotations, "; "))
		} else {
			fmt.Fprintf(out, "  %d. %s\n", i+1, label)
		}
		if s.Description != "" {
			fmt.Fprintf(out, "     %s\n", s.Description)
		}
	}
}

func printFirstBlockDetail(index int, step automation.Step, out io.Writer) {
	header := "first"
	var blockAnnotations []string
	if step.If != "" {
		blockAnnotations = append(blockAnnotations, fmt.Sprintf("if: %s", step.If))
	}
	if step.Pipe {
		blockAnnotations = append(blockAnnotations, "pipe")
	}
	if len(blockAnnotations) > 0 {
		fmt.Fprintf(out, "  %d. %s  [%s]\n", index+1, header, strings.Join(blockAnnotations, "; "))
	} else {
		fmt.Fprintf(out, "  %d. %s\n", index+1, header)
	}
	if step.Description != "" {
		fmt.Fprintf(out, "     %s\n", step.Description)
	}
	for j, sub := range step.First {
		label := fmt.Sprintf("%s: %s", sub.Type, truncateValue(sub.Value, 36))
		subAnnotations := stepAnnotations(sub)
		if len(subAnnotations) > 0 {
			fmt.Fprintf(out, "     %c. %s  [%s]\n", 'a'+j, label, strings.Join(subAnnotations, "; "))
		} else {
			fmt.Fprintf(out, "     %c. %s\n", 'a'+j, label)
		}
		if sub.Description != "" {
			fmt.Fprintf(out, "        %s\n", sub.Description)
		}
	}
}

func truncateValue(s string, maxLen int) string {
	s = strings.SplitN(s, "\n", 2)[0]
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func printInputSpec(name string, spec automation.InputSpec, out io.Writer) {
	var parts []string

	if spec.Type != "" {
		parts = append(parts, spec.Type)
	}

	if spec.IsRequired() {
		parts = append(parts, "required")
	} else {
		parts = append(parts, "optional")
	}

	if spec.Default != "" {
		parts = append(parts, fmt.Sprintf("default: %q", spec.Default))
	}

	meta := strings.Join(parts, ", ")
	envName := "PI_IN_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	fmt.Fprintf(out, "  %s (%s) → $%s\n", name, meta, envName)

	if spec.Description != "" {
		fmt.Fprintf(out, "      %s\n", spec.Description)
	}
}
