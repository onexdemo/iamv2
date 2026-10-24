// Package secret implements the iamctl secret commands.
package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cmdutil "github.com/onexdemo/iamv2/internal/iamctl/cmd/util"
	"github.com/onexdemo/iamv2/internal/iamctl/util/table"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// NewCmdSecret creates the secret command group.
func NewCmdSecret() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage iam access keys",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(
		newCmdCreate(),
		newCmdGet(),
		newCmdList(),
		newCmdUpdate(),
		newCmdDelete(),
	)

	return cmd
}

func newCmdCreate() *cobra.Command {
	description := ""
	expires := time.Now().Add(144 * time.Hour).Unix()

	cmd := &cobra.Command{
		Use:   "create SECRET_NAME",
		Short: "Create an access key pair",
		Args:  cobra.ExactArgs(1),
		Example: `  # Create a secret that expires in 144 hours
  iamctl secret create foo

  # Create a secret with an explicit expiry and description
  iamctl secret create foo --expires=1988121600 --description="secret for iam"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := cmdutil.Client().CreateSecret(context.Background(), &v1.CreateSecretRequest{
				Name:        args[0],
				Expires:     expires,
				Description: description,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "secret/%s created\n", resp.Secret.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  secretID:  %s\n", resp.Secret.SecretID)
			fmt.Fprintf(cmd.OutOrStdout(), "  secretKey: %s\n", resp.Secret.SecretKey)

			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "The description of the secret.")
	cmd.Flags().Int64Var(&expires, "expires", expires, "The expiry time of the secret as a UNIX timestamp.")

	return cmd
}

func newCmdGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get SECRET_NAME",
		Short: "Display an access key pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := cmdutil.Client().GetSecret(context.Background(), args[0])
			if err != nil {
				return err
			}

			secret := resp.Secret
			fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", secret.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "SecretID:    %s\n", secret.SecretID)
			fmt.Fprintf(cmd.OutOrStdout(), "SecretKey:   %s\n", secret.SecretKey)
			fmt.Fprintf(cmd.OutOrStdout(), "Expires:     %s\n", formatExpiry(secret.Expires))
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", secret.Description)

			return nil
		},
	}
}

func newCmdList() *cobra.Command {
	var offset, limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display all access keys owned by the authenticated user",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := cmdutil.Client().ListSecrets(context.Background(), offset, limit)
			if err != nil {
				return err
			}

			rows := make([][]string, 0, len(resp.Items))
			for _, secret := range resp.Items {
				rows = append(rows, []string{
					secret.Name,
					secret.SecretID,
					formatExpiry(secret.Expires),
					secret.Description,
					table.FormatTime(secret.CreatedAt),
				})
			}

			table.Print(cmd.OutOrStdout(), []string{"NAME", "SECRET ID", "EXPIRES", "DESCRIPTION", "CREATED"}, rows)

			return nil
		},
	}

	cmd.Flags().Int64VarP(&offset, "offset", "o", 0, "Specify the offset of the first row to be returned.")
	cmd.Flags().Int64VarP(&limit, "limit", "l", 1000, "Specify the number of records to be returned.")

	return cmd
}

func newCmdUpdate() *cobra.Command {
	var (
		description string
		expires     int64
	)

	cmd := &cobra.Command{
		Use:   "update SECRET_NAME",
		Short: "Update an access key pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdutil.Client()
			current, err := c.GetSecret(context.Background(), args[0])
			if err != nil {
				return err
			}

			rq := &v1.UpdateSecretRequest{
				Name:        args[0],
				Expires:     current.Secret.Expires,
				Description: current.Secret.Description,
			}
			if cmd.Flags().Changed("expires") {
				rq.Expires = expires
			}

			if cmd.Flags().Changed("description") {
				rq.Description = description
			}

			resp, err := c.UpdateSecret(context.Background(), rq)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "secret/%s updated\n", resp.Secret.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "The description of the secret.")
	cmd.Flags().Int64Var(&expires, "expires", 0, "The expiry time of the secret as a UNIX timestamp.")

	return cmd
}

func newCmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete SECRET_NAME",
		Short: "Delete an access key pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Client().DeleteSecret(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "secret/%s deleted\n", args[0])

			return nil
		},
	}
}

func formatExpiry(expires int64) string {
	if expires <= 0 {
		return "never"
	}

	return time.Unix(expires, 0).Format(time.DateTime)
}
