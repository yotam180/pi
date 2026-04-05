package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate a shell completion script for the specified shell.

To load completions:

Bash:
  $ source <(pi completion bash)
  # Or permanently:
  $ pi completion bash > /etc/bash_completion.d/pi

Zsh:
  $ source <(pi completion zsh)
  # Or permanently (requires compinit):
  $ pi completion zsh > "${fpath[1]}/_pi"

Fish:
  $ pi completion fish | source
  # Or permanently:
  $ pi completion fish > ~/.config/fish/completions/pi.fish

PowerShell:
  PS> pi completion powershell | Out-String | Invoke-Expression
  # Or permanently:
  PS> pi completion powershell > pi.ps1  # and source it from your profile`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return cmd.Help()
			}
		},
	}
	return cmd
}

// automationCompleter returns a ValidArgsFunction that completes automation names.
// It silently returns empty on any error (completion should never crash).
func automationCompleter() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cwd, err := os.Getwd()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		pc, err := resolveProject(cwd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		result, err := pc.Discover(nil)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		names := result.Names()
		var completions []string
		for _, name := range names {
			if result.IsBuiltin(name) {
				continue
			}
			a := result.Automations[name]
			if a.Description != "" {
				completions = append(completions, name+"\t"+a.Description)
			} else {
				completions = append(completions, name)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
