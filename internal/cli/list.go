package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/project"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available automations in the project",
		Long: `Scan the .pi/ folder and list all discovered automations with their names
and descriptions. PI walks up from the current directory to find pi.yaml,
so this works from any subdirectory of the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			return listAutomations(cwd, cmd.OutOrStdout())
		},
	}
}

// listAutomations discovers and prints all automations. Extracted for testability.
func listAutomations(startDir string, out io.Writer) error {
	root, err := project.FindRoot(startDir)
	if err != nil {
		return err
	}

	piDir := filepath.Join(root, discovery.PiDir)
	result, err := discovery.Discover(piDir)
	if err != nil {
		return fmt.Errorf("discovering automations: %w", err)
	}

	names := result.Names()
	if len(names) == 0 {
		fmt.Fprintln(out, "No automations found. Create .pi/*.yaml files to get started.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tINPUTS\n")
	for _, name := range names {
		a := result.Automations[name]
		desc := a.Description
		if desc == "" {
			desc = "-"
		}
		inputs := formatInputsSummary(a)
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, desc, inputs)
	}
	return w.Flush()
}

func formatInputsSummary(a *automation.Automation) string {
	if len(a.InputKeys) == 0 {
		return "-"
	}
	var parts []string
	for _, key := range a.InputKeys {
		spec := a.Inputs[key]
		if spec.IsRequired() {
			parts = append(parts, key)
		} else {
			parts = append(parts, key+"?")
		}
	}
	return strings.Join(parts, ", ")
}
