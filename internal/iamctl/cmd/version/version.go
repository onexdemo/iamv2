// Package version implements the iamctl version command.
package version

import (
	"fmt"

	"github.com/onexstack/onexstack/pkg/version"
	"github.com/spf13/cobra"
)

// NewCmdVersion creates the version command.
func NewCmdVersion() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the iamctl version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := version.Get()

			switch output {
			case "json":
				fmt.Fprintln(cmd.OutOrStdout(), info.ToJSON())
			default:
				fmt.Fprintln(cmd.OutOrStdout(), info.Text())
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format: json or text.")

	return cmd
}
