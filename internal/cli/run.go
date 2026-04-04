package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/discovery"
	"github.com/vyper-tooling/pi/internal/executor"
	"github.com/vyper-tooling/pi/internal/project"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <automation> [args...]",
		Short: "Run an automation by name",
		Long: `Run a PI automation by its name. The automation is resolved from the .pi/ folder
in the current project. PI walks up from the current directory to find pi.yaml,
so this works from any subdirectory of the project.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			return runAutomation(cwd, args[0], args[1:], os.Stdout, os.Stderr)
		},
	}
}

// runAutomation resolves and executes an automation. Extracted for testability.
// Returns *executor.ExitError for non-zero exit codes (caller decides whether to os.Exit).
func runAutomation(startDir, name string, args []string, stdout, stderr io.Writer) error {
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

	stdoutFile, stdoutIsFile := stdout.(*os.File)
	stderrFile, stderrIsFile := stderr.(*os.File)

	exec := &executor.Executor{
		RepoRoot:  root,
		Discovery: result,
	}
	if stdoutIsFile {
		exec.Stdout = stdoutFile
	}
	if stderrIsFile {
		exec.Stderr = stderrFile
	}

	return exec.Run(a, args)
}
