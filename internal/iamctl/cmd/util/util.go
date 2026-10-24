// Package util contains helpers shared by iamctl commands.
package util

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/onexdemo/iamv2/internal/iamctl/pkg/client"
)

var defaultClient *client.Client

// SetClient stores the process-wide apiserver client.
func SetClient(c *client.Client) {
	defaultClient = c
}

// Client returns the process-wide apiserver client.
func Client() *client.Client {
	return defaultClient
}

// CheckErr prints the error and exits with a non-zero status.
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// UsageErrorf returns a usage error for the given command.
func UsageErrorf(cmd *cobra.Command, format string, args ...any) error {
	return fmt.Errorf("%s\nSee '%s -h' for help and examples", fmt.Sprintf(format, args...), cmd.CommandPath())
}
