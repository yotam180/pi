package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available automations in the project",
		Long:  "Scan the .pi/ folder and list all discovered automations with their names and descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "pi list: not implemented yet")
			return nil
		},
	}
}
