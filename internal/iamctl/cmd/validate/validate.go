// Package validate implements the iamctl validate command.
package validate

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cmdutil "github.com/onexdemo/iamv2/internal/iamctl/cmd/util"
)

// NewCmdValidate creates the validate command.
func NewCmdValidate() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the iamctl configuration and apiserver connectivity",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := cmdutil.Client().Healthz(ctx); err != nil {
				return fmt.Errorf("apiserver health check failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "configuration is valid; apiserver is healthy")

			return nil
		},
	}
}
