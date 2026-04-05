package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/shell"
)

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Install shortcut functions into the current shell config",
		Long: `Reads shortcuts from pi.yaml and writes shell functions to ~/.pi/shell/<project>.sh.
A source line is added to .zshrc (and .bashrc if it exists) so shortcuts are
available in every new terminal.

Running pi shell again overwrites the file (idempotent).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShellInstall(cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	cmd.AddCommand(newShellUninstallCmd())
	cmd.AddCommand(newShellListCmd())

	return cmd
}

func newShellUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove shell shortcuts for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShellUninstall(cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

func newShellListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all installed shortcut files across all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShellList(cmd.OutOrStdout())
		},
	}
}

func runShellInstall(stdout, stderr io.Writer) error {
	cwd, err := getwd()
	if err != nil {
		return err
	}
	pc, err := resolveProjectStrict(cwd)
	if err != nil {
		return err
	}

	piBin, err := resolvePiBinary()
	if err != nil {
		return err
	}

	shellPath, err := shell.Install(pc.Config, piBin, pc.Root)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Installed %d shortcut(s) to %s\n", len(pc.Config.Shortcuts), shellPath)
	for name, sc := range pc.Config.Shortcuts {
		mode := "cd"
		if sc.Anywhere {
			mode = "anywhere"
		}
		fmt.Fprintf(stdout, "  %s → %s (%s)\n", name, sc.Run, mode)
	}
	fmt.Fprintln(stdout, "\nRestart your shell or run: source ~/.zshrc")
	return nil
}

func runShellUninstall(stdout, stderr io.Writer) error {
	cwd, err := getwd()
	if err != nil {
		return err
	}
	pc, err := resolveProjectStrict(cwd)
	if err != nil {
		return err
	}

	if err := shell.Uninstall(pc.Config.Project); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Removed shell shortcuts for %s\n", pc.Config.Project)
	return nil
}

func runShellList(out io.Writer) error {
	projects, err := shell.ListInstalled()
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Fprintln(out, "No shell shortcuts installed.")
		return nil
	}

	shellDir, _ := shell.ShellFileDir()
	fmt.Fprintf(out, "Installed shortcut files (%s):\n", shellDir)
	for _, p := range projects {
		fmt.Fprintf(out, "  %s.sh\n", p)
	}
	return nil
}

func resolvePiBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving pi binary: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe, nil
	}
	return resolved, nil
}
