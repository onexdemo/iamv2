// Package color implements the iamctl color command.
package color

import (
	"fmt"

	"github.com/spf13/cobra"
)

var palette = []string{
	"\033[31m", // red
	"\033[32m", // green
	"\033[33m", // yellow
	"\033[34m", // blue
	"\033[35m", // magenta
	"\033[36m", // cyan
}

const reset = "\033[0m"

// NewCmdColor creates the color command.
func NewCmdColor() *cobra.Command {
	return &cobra.Command{
		Use:   "color STRING...",
		Short: "Print strings in rotating terminal colors",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, arg := range args {
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n", palette[i%len(palette)], arg, reset)
			}

			return nil
		},
	}
}
