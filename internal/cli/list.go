package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/discovery"
)

func newListCmd() *cobra.Command {
	var showAll bool
	var showBuiltins bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available automations in the project",
		Long: `Scan the .pi/ folder and list all discovered automations with their names
and descriptions. PI walks up from the current directory to find pi.yaml,
so this works from any subdirectory of the project.

By default, built-in automations (pi:*) are hidden. Use --builtins to include them.
Use --all to browse automations from all declared packages.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return listAutomations(cwd, cmd.OutOrStdout(), showAll, showBuiltins)
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show automations from all declared packages")
	cmd.Flags().BoolVarP(&showBuiltins, "builtins", "b", false, "Include built-in (pi:*) automations")

	return cmd
}

// listAutomations discovers and prints all automations. Extracted for testability.
func listAutomations(startDir string, out io.Writer, showAll bool, showBuiltins bool) error {
	pc, err := resolveProject(startDir)
	if err != nil {
		return err
	}

	result, err := pc.Discover(nil)
	if err != nil {
		return err
	}

	names := result.Names()
	if len(names) == 0 {
		fmt.Fprintln(out, "No automations found. Create .pi/*.yaml files to get started.")
		return nil
	}

	// Filter names for default view (exclude builtins unless --builtins)
	var filtered []string
	for _, name := range names {
		if result.IsBuiltin(name) && !showBuiltins {
			continue
		}
		filtered = append(filtered, name)
	}

	if len(filtered) == 0 {
		fmt.Fprintln(out, "No automations found. Create .pi/*.yaml files to get started.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tSOURCE\tDESCRIPTION\tINPUTS\n")
	for _, name := range filtered {
		a := result.Automations[name]
		desc := a.Description
		if desc == "" {
			desc = "-"
		}
		source := automationSource(result, name)
		inputs := formatInputsSummary(a)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, source, desc, inputs)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if showAll {
		return printPackageAutomations(out, result)
	}

	return nil
}

// automationSource returns the source indicator for an automation name.
func automationSource(r *discovery.Result, name string) string {
	if r.IsBuiltin(name) {
		return "[built-in]"
	}
	if r.IsPackage(name) {
		source := r.PackageSource(name)
		// Check if the package has an alias
		for _, pkg := range r.Packages() {
			if pkg.Source == source && pkg.Alias != "" {
				return pkg.Alias
			}
		}
		return source
	}
	return "[workspace]"
}

// printPackageAutomations prints grouped sections for each declared package,
// showing all automations available in that package (not just those merged
// into the main list).
func printPackageAutomations(out io.Writer, result *discovery.Result) error {
	pkgs := result.Packages()
	if len(pkgs) == 0 {
		return nil
	}

	for _, pkg := range pkgs {
		autos := result.PackageAutomations(pkg.Source)
		if len(autos) == 0 {
			continue
		}

		// Print package header
		fmt.Fprintf(out, "\n── %s ", pkg.Source)
		if pkg.Alias != "" {
			fmt.Fprintf(out, "(alias: %s) ", pkg.Alias)
		}
		headerLen := len(pkg.Source) + 4
		if pkg.Alias != "" {
			headerLen += len(pkg.Alias) + 10
		}
		padding := 60 - headerLen
		if padding < 3 {
			padding = 3
		}
		fmt.Fprintln(out, strings.Repeat("─", padding))

		// Sort and print automations in this package
		names := sortedKeys(autos)
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		for _, name := range names {
			a := autos[name]
			desc := a.Description
			if desc == "" {
				desc = "-"
			}
			source := pkg.Alias
			if source == "" {
				source = pkg.Source
			}
			inputs := formatInputsSummary(a)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, source, desc, inputs)
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

// sortedKeys returns sorted keys from a map.
func sortedKeys(m map[string]*automation.Automation) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
