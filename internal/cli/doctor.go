package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/automation"
	"github.com/vyper-tooling/pi/internal/conditions"
	"github.com/vyper-tooling/pi/internal/display"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/reqcheck"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check requirement health for all automations",
		Long: `Scan all automations in the project, check their requires: entries,
and print a per-automation health table showing which requirements are
satisfied and which are missing.

Exits with code 0 if all requirements are met, 1 if any are missing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := getwd()
			if err != nil {
				return err
			}
			return runDoctor(cwd, cmd.OutOrStdout())
		},
	}
}

func runDoctor(startDir string, out io.Writer) error {
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
		fmt.Fprintln(out, "No automations found.")
		return nil
	}

	p := display.New(out)
	env := conditions.DefaultRuntimeEnv()
	anyFailed := false
	anyPrinted := false

	for _, name := range names {
		a := result.Automations[name]
		if len(a.Requires) == 0 {
			continue
		}

		if anyPrinted {
			fmt.Fprintln(out)
		}
		p.Bold("  %s\n", name)
		anyPrinted = true

		for _, req := range a.Requires {
			check := reqcheck.CheckRequirementForDoctor(req, env)
			if check.Satisfied {
				label := formatDoctorLabel(req)
				if check.DetectedVersion != "" {
					p.Green("    ✓ %-25s (%s)\n", label, check.DetectedVersion)
				} else {
					p.Green("    ✓ %s\n", label)
				}
			} else {
				label := formatDoctorLabel(req)
				hint := reqcheck.InstallHintFor(req)
				if hint != "" {
					p.Red("    ✗ %-25s %s → %s\n", label, check.Error, hint)
				} else {
					p.Red("    ✗ %-25s %s\n", label, check.Error)
				}
				anyFailed = true
			}
		}
	}

	if !anyPrinted {
		fmt.Fprintln(out, "No automations have requirements.")
		return nil
	}

	if anyFailed {
		return &executor.ExitError{Code: 1}
	}

	return nil
}

func formatDoctorLabel(req automation.Requirement) string {
	if req.MinVersion != "" {
		if req.Kind == automation.RequirementCommand {
			return fmt.Sprintf("command: %s >= %s", req.Name, req.MinVersion)
		}
		return fmt.Sprintf("%s >= %s", req.Name, req.MinVersion)
	}
	if req.Kind == automation.RequirementCommand {
		return fmt.Sprintf("command: %s", req.Name)
	}
	return req.Name
}
