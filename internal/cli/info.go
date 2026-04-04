package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
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

	piDir := filepath.Join(root, discovery.PiDir)
	result, err := discovery.Discover(piDir)
	if err != nil {
		return fmt.Errorf("discovering automations: %w", err)
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

	fmt.Fprintf(out, "Steps:        %d\n", len(a.Steps))

	printStepsDetail(a.Steps, out)

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

func printStepsDetail(steps []automation.Step, out io.Writer) {
	hasConditions := false
	for _, s := range steps {
		if s.If != "" {
			hasConditions = true
			break
		}
	}
	if !hasConditions {
		return
	}

	fmt.Fprintf(out, "\nStep details:\n")
	for i, s := range steps {
		label := fmt.Sprintf("%s: %s", s.Type, truncateValue(s.Value, 40))
		if s.If != "" {
			fmt.Fprintf(out, "  %d. %s  [if: %s]\n", i+1, label, s.If)
		} else {
			fmt.Fprintf(out, "  %d. %s\n", i+1, label)
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
