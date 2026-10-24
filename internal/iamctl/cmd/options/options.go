// Package options implements the iamctl options command.
package options

import (
	"github.com/onexstack/onexstack/pkg/cli/templates"
	"github.com/spf13/cobra"
)

// NewCmdOptions creates the options command.
func NewCmdOptions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Print the flags inherited by all commands",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Usage()
		},
	}

	templates.UseOptionsTemplates(cmd)

	return cmd
}
