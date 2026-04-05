package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/project"
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
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			return showAutomationInfo(cwd, args[0], cmd.OutOrStdout())
		},
	}
}

// showAutomationInfo resolves an automation by name and prints its details.
func showAutomationInfo(startDir, name string, out io.Writer) error {
	root, err := project.FindRoot(startDir)
	if err != nil {
		return err
	}

	result, err := discoverAll(root)
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

func printStepsDetail(steps []automation.Step, out io.Writer) {
	hasDetails := false
	for _, s := range steps {
		if s.If != "" || len(s.Env) > 0 || s.Silent || s.ParentShell || s.Dir != "" || s.Timeout > 0 || s.Description != "" {
			hasDetails = true
			break
		}
	}
	if !hasDetails {
		return
	}

	fmt.Fprintf(out, "\nStep details:\n")
	for i, s := range steps {
		label := fmt.Sprintf("%s: %s", s.Type, truncateValue(s.Value, 40))
		var annotations []string
		if s.If != "" {
			annotations = append(annotations, fmt.Sprintf("if: %s", s.If))
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
	fmt.Fprintf(out, "  %s (%s)\n", name, meta)

	if spec.Description != "" {
		fmt.Fprintf(out, "      %s\n", spec.Description)
	}
}
