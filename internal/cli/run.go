package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
)

func newRunCmd() *cobra.Command {
	var repoFlag string
	var withFlags []string

	cmd := &cobra.Command{
		Use:   "run <automation> [args...]",
		Short: "Run an automation by name",
		Long: `Run a PI automation by its name. The automation is resolved from the .pi/ folder
in the current project. PI walks up from the current directory to find pi.yaml,
so this works from any subdirectory of the project.

Use --repo to specify the project root explicitly (used by "anywhere" shortcuts).
Use --with key=value to pass named inputs (repeatable).`,
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

			return runAutomation(startDir, args[0], args[1:], withArgs, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().StringVar(&repoFlag, "repo", "", "project root path (used by anywhere shortcuts)")
	cmd.Flags().StringArrayVar(&withFlags, "with", nil, "pass named input as key=value (repeatable)")

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
func runAutomation(startDir, name string, args []string, withArgs map[string]string, stdout, stderr io.Writer) error {
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

	exec := &executor.Executor{
		RepoRoot:  root,
		Discovery: result,
		Stdout:    stdout,
		Stderr:    stderr,
	}

	return exec.RunWithInputs(a, args, withArgs)
}
