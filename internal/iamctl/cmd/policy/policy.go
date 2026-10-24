// Package policy implements the iamctl policy commands.
package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ory/ladon"
	"github.com/spf13/cobra"

	cmdutil "github.com/onexdemo/iamv2/internal/iamctl/cmd/util"
	"github.com/onexdemo/iamv2/internal/iamctl/util/table"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// NewCmdPolicy creates the policy command group.
func NewCmdPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage iam authorization policies",
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
	return &cobra.Command{
		Use:     "create POLICY_NAME POLICY_JSON",
		Short:   "Create an authorization policy",
		Args:    cobra.ExactArgs(2),
		Example: `  iamctl policy create foo '{"description":"allow reads","subjects":["users:foo"],"actions":["get"],"effect":"allow","resources":["resources:articles:*"]}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			policy, err := decodePolicy(args[1])
			if err != nil {
				return err
			}

			resp, err := cmdutil.Client().CreatePolicy(context.Background(), &v1.CreatePolicyRequest{
				Name:   args[0],
				Policy: *policy,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "policy/%s created\n", resp.Policy.Name)

			return nil
		},
	}
}

func newCmdGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get POLICY_NAME",
		Short: "Display an authorization policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := cmdutil.Client().GetPolicy(context.Background(), args[0])
			if err != nil {
				return err
			}

			printPolicy(cmd, resp.Policy)

			return nil
		},
	}
}

func newCmdList() *cobra.Command {
	var offset, limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display all policies owned by the authenticated user",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := cmdutil.Client().ListPolicies(context.Background(), offset, limit)
			if err != nil {
				return err
			}

			rows := make([][]string, 0, len(resp.Items))
			for _, policy := range resp.Items {
				rows = append(rows, []string{
					policy.Name,
					policy.Policy.Effect,
					fmt.Sprintf("%d", len(policy.Policy.Subjects)),
					fmt.Sprintf("%d", len(policy.Policy.Resources)),
					fmt.Sprintf("%d", len(policy.Policy.Actions)),
					table.FormatTime(policy.CreatedAt),
				})
			}

			table.Print(cmd.OutOrStdout(), []string{"NAME", "EFFECT", "SUBJECTS", "RESOURCES", "ACTIONS", "CREATED"}, rows)

			return nil
		},
	}

	cmd.Flags().Int64VarP(&offset, "offset", "o", 0, "Specify the offset of the first row to be returned.")
	cmd.Flags().Int64VarP(&limit, "limit", "l", 1000, "Specify the number of records to be returned.")

	return cmd
}

func newCmdUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update POLICY_NAME POLICY_JSON",
		Short: "Update an authorization policy",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			policy, err := decodePolicy(args[1])
			if err != nil {
				return err
			}

			resp, err := cmdutil.Client().UpdatePolicy(context.Background(), &v1.UpdatePolicyRequest{
				Name:   args[0],
				Policy: *policy,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "policy/%s updated\n", resp.Policy.Name)

			return nil
		},
	}
}

func newCmdDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete POLICY_NAME",
		Short: "Delete an authorization policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Client().DeletePolicy(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "policy/%s deleted\n", args[0])

			return nil
		},
	}
}

func decodePolicy(raw string) (*ladon.DefaultPolicy, error) {
	var policy ladon.DefaultPolicy

	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return nil, fmt.Errorf("decode policy: %w", err)
	}

	return &policy, nil
}

func printPolicy(cmd *cobra.Command, policy *v1.Policy) {
	data, err := json.MarshalIndent(policy.Policy, "", "  ")
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "policy/%s\n", policy.Name)

		return
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Name:     %s\n", policy.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Username: %s\n", policy.Username)
	fmt.Fprintf(cmd.OutOrStdout(), "Policy:\n%s\n", data)
}
