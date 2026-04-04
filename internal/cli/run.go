package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <automation> [args...]",
		Short: "Run an automation by name",
		Long:  "Run a PI automation by its name. The automation is resolved from the .pi/ folder in the current project.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "pi run: not implemented yet (automation=%q)\n", args[0])
			return nil
		},
	}
}
