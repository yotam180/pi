package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Run all setup automations",
		Long:  "Run all automations listed in the setup section of pi.yaml sequentially.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "pi setup: not implemented yet")
			return nil
		},
	}
}
