package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/config"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
	"github.com/vyper-tooling/pi/internal/runtimes"
)

func newRunCmd() *cobra.Command {
	var repoFlag string
	var withFlags []string
	var silent bool
	var loud bool

	cmd := &cobra.Command{
		Use:   "run <automation> [args...]",
		Short: "Run an automation by name",
		Long: `Run a PI automation by its name. The automation is resolved from the .pi/ folder
in the current project. PI walks up from the current directory to find pi.yaml,
so this works from any subdirectory of the project.

Use --repo to specify the project root explicitly (used by "anywhere" shortcuts).
Use --with key=value to pass named inputs (repeatable).
Use --silent to suppress PI status lines for installer automations.
Use --loud to force trace lines and output for all steps (overrides silent: true).`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir := repoFlag
			if startDir == "" {
				var err error
				startDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
			}

			withArgs, err := parseWithFlags(withFlags)
			if err != nil {
				return err
			}

			return runAutomation(startDir, args[0], args[1:], withArgs, silent, loud, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&repoFlag, "repo", "", "project root path (used by anywhere shortcuts)")
	cmd.Flags().StringArrayVar(&withFlags, "with", nil, "pass named input as key=value (repeatable)")
	cmd.Flags().BoolVar(&silent, "silent", false, "suppress PI status lines for installer automations")
	cmd.Flags().BoolVar(&loud, "loud", false, "force trace lines and output for all steps (overrides silent: true)")

	return cmd
}

// parseWithFlags converts ["key=value", ...] into map[string]string.
func parseWithFlags(flags []string) (map[string]string, error) {
	if len(flags) == 0 {
		return nil, nil
	}
	result := make(map[string]string, len(flags))
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid --with flag %q: must be key=value", f)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

// runAutomation resolves and executes an automation. Extracted for testability.
func runAutomation(startDir, name string, args []string, withArgs map[string]string, silent, loud bool, stdout, stderr io.Writer) error {
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

	exec := &executor.Executor{
		RepoRoot:  root,
		Discovery: result,
		Stdout:    stdout,
		Stderr:    stderr,
		Silent:    silent,
		Loud:      loud,
	}

	cfg, cfgErr := config.Load(root)
	if cfgErr == nil && cfg.EffectiveProvisionMode() != config.ProvisionNever {
		exec.Provisioner = runtimes.NewProvisioner(cfg, stderr)
	}

	return exec.RunWithInputs(a, args, withArgs)
}
