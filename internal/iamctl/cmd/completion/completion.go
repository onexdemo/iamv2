// Package completion implements the iamctl completion command.
package completion

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the completion command.
func NewCmdCompletion() *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell %q: use bash, zsh, fish or powershell", shell)
			}
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "bash", "Shell type: bash, zsh, fish or powershell.")

	return cmd
}
