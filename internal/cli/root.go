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

	root.SetVersionTemplate("pi {{.Version}}\n")

	root.AddCommand(newRunCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newInfoCmd())
	root.AddCommand(newSetupCmd())
	root.AddCommand(newShellCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newValidateCmd())

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
