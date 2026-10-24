// Package user implements the iamctl user commands.
package user

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	cmdutil "github.com/onexdemo/iamv2/internal/iamctl/cmd/util"
	"github.com/onexdemo/iamv2/internal/iamctl/util/table"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// NewCmdUser creates the user command group.
func NewCmdUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage iam users",
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
		newCmdChangePassword(),
	)

	return cmd
}

func newCmdCreate() *cobra.Command {
	var nickname, phone string

	cmd := &cobra.Command{
		Use:   "create USERNAME PASSWORD EMAIL",
		Short: "Create a user resource",
		Args:  cobra.ExactArgs(3),
		Example: `  # Create user foo
  iamctl user create foo Foo@2020 foo@foxmail.com

  # Create user with extra profile fields
  iamctl user create foo Foo@2020 foo@foxmail.com --nickname=colin --phone=18128845xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if nickname == "" {
				nickname = args[0]
			}

			resp, err := cmdutil.Client().CreateUser(context.Background(), &v1.CreateUserRequest{
				Name:     args[0],
				Password: args[1],
				Nickname: nickname,
				Email:    args[2],
				Phone:    phone,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "user/%s created\n", resp.User.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&nickname, "nickname", "", "The nickname of the user.")
	cmd.Flags().StringVar(&phone, "phone", "", "The phone number of the user.")

	return cmd
}

func newCmdGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get USERNAME",
		Short: "Display a user resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := cmdutil.Client().GetUser(context.Background(), args[0])
			if err != nil {
				return err
			}

			user := resp.User
			table.Print(cmd.OutOrStdout(), []string{"NAME", "NICKNAME", "EMAIL", "PHONE", "STATUS", "CREATED"},
				[][]string{{
					user.Name,
					user.Nickname,
					user.Email,
					user.Phone,
					fmt.Sprintf("%d", user.Status),
					table.FormatTime(user.CreatedAt),
				}})

			return nil
		},
	}
}

func newCmdList() *cobra.Command {
	var offset, limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display all users in the iam platform",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := cmdutil.Client().ListUsers(context.Background(), offset, limit)
			if err != nil {
				return err
			}

			rows := make([][]string, 0, len(resp.Items))
			for _, user := range resp.Items {
				rows = append(rows, []string{
					user.Name,
					user.Nickname,
					user.Email,
					user.Phone,
					fmt.Sprintf("%d", user.TotalPolicy),
					table.FormatTime(user.CreatedAt),
				})
			}

			table.Print(cmd.OutOrStdout(), []string{"NAME", "NICKNAME", "EMAIL", "PHONE", "POLICIES", "CREATED"}, rows)

			return nil
		},
	}

	cmd.Flags().Int64VarP(&offset, "offset", "o", 0, "Specify the offset of the first row to be returned.")
	cmd.Flags().Int64VarP(&limit, "limit", "l", 1000, "Specify the number of records to be returned.")

	return cmd
}

func newCmdUpdate() *cobra.Command {
	var nickname, email, phone string

	cmd := &cobra.Command{
		Use:   "update USERNAME",
		Short: "Update a user resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdutil.Client()
			current, err := c.GetUser(context.Background(), args[0])
			if err != nil {
				return err
			}

			rq := &v1.UpdateUserRequest{
				Name:     args[0],
				Nickname: current.User.Nickname,
				Email:    current.User.Email,
				Phone:    current.User.Phone,
			}
			if nickname != "" {
				rq.Nickname = nickname
			}

			if email != "" {
				rq.Email = email
			}

			if phone != "" {
				rq.Phone = phone
			}

			resp, err := c.UpdateUser(context.Background(), rq)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "user/%s updated\n", resp.User.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&nickname, "nickname", "", "The nickname of the user.")
	cmd.Flags().StringVar(&email, "email", "", "The email of the user.")
	cmd.Flags().StringVar(&phone, "phone", "", "The phone number of the user.")

	return cmd
}

func newCmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete USERNAME",
		Short: "Delete a user resource from the iam platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Client().DeleteUser(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "user/%s deleted\n", args[0])

			return nil
		},
	}
}

func newCmdChangePassword() *cobra.Command {
	var oldPassword, newPassword string

	cmd := &cobra.Command{
		Use:   "change-password USERNAME",
		Short: "Change the password of the authenticated user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Client().ChangePassword(context.Background(), args[0], &v1.ChangePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: newPassword,
			}); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "user/%s password changed\n", args[0])

			return nil
		},
	}

	cmd.Flags().StringVar(&oldPassword, "old-password", "", "The current password.")
	cmd.Flags().StringVar(&newPassword, "new-password", "", "The new password.")

	_ = cmd.MarkFlagRequired("old-password")
	_ = cmd.MarkFlagRequired("new-password")

	return cmd
}
