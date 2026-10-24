// Package info implements the iamctl info command.
package info

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdInfo creates the info command.
func NewCmdInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display the iamctl connection configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Server:      %s\n", viper.GetString("server.address"))
			fmt.Fprintf(cmd.OutOrStdout(), "Username:    %s\n", viper.GetString("user.username"))
			fmt.Fprintf(cmd.OutOrStdout(), "SecretID:    %s\n", viper.GetString("user.secret-id"))
			fmt.Fprintf(cmd.OutOrStdout(), "Timeout:     %s\n", viper.GetDuration("server.timeout"))
			fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", viper.GetString("config"))

			return nil
		},
	}
}
