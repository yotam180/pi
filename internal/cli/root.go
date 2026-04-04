package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vyper-tooling/pi/internal/executor"
)

var version = "dev"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "pi",
		Short: "PI — developer automation platform",
		Long: `PI is a developer automation platform for teams managing complex repositories.
It replaces shell shortcut files and setup scripts with a structured, polyglot,
and shareable automation model.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newRunCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newSetupCmd())
	root.AddCommand(newShellCmd())

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		if exitErr, ok := err.(*executor.ExitError); ok {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
