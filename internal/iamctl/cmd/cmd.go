// Package cmd builds the iamctl command tree.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/onexstack/onexstack/pkg/cli/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/onexdemo/iamv2/internal/iamctl/cmd/color"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/completion"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/info"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/jwt"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/options"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/policy"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/secret"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/user"
	cmdutil "github.com/onexdemo/iamv2/internal/iamctl/cmd/util"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/validate"
	"github.com/onexdemo/iamv2/internal/iamctl/cmd/version"
	"github.com/onexdemo/iamv2/internal/iamctl/pkg/client"
)

const (
	// defaultHomeDir is the directory that holds iamctl configuration.
	defaultHomeDir = ".iamv2"
	// defaultConfigName is the default configuration file name.
	defaultConfigName = "iamctl.yaml"
)

// NewDefaultCommand creates the iamctl command tree.
func NewDefaultCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iamctl",
		Short: "iamctl controls the iam platform",
		Long: templates.LongDesc(`
		iamctl is the command-line client for the iam platform.

		Find more information at:
			https://github.com/marmotedu/iam`),
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return loadClient(cmd)
		},
	}

	flags := cmd.PersistentFlags()
	flags.String("config", defaultConfigPath(), "Path to the iamctl configuration file.")
	flags.String("server", "", "Base URL of iam-apiserver.")
	flags.String("username", "", "Username used for HTTP Basic authentication.")
	flags.String("password", "", "Password used for HTTP Basic authentication.")
	flags.String("token", "", "Bearer token used for authentication.")
	flags.String("secret-id", "", "Access key ID used for authentication.")
	flags.String("secret-key", "", "Access key secret used for authentication.")
	flags.Duration("timeout", 10*time.Second, "Request timeout.")
	flags.Bool("insecure-skip-tls-verify", false, "Skip TLS certificate verification.")
	flags.String("certificate-authority", "", "Path to the CA certificate used to verify the server.")

	bindFlag(flags, "config", "config")
	bindFlag(flags, "server", "server.address")
	bindFlag(flags, "username", "user.username")
	bindFlag(flags, "password", "user.password")
	bindFlag(flags, "token", "user.token")
	bindFlag(flags, "secret-id", "user.secret-id")
	bindFlag(flags, "secret-key", "user.secret-key")
	bindFlag(flags, "timeout", "server.timeout")
	bindFlag(flags, "insecure-skip-tls-verify", "server.insecure-skip-tls-verify")
	bindFlag(flags, "certificate-authority", "server.certificate-authority")

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				info.NewCmdInfo(),
				color.NewCmdColor(),
			},
		},
		{
			Message: "Identity and Access Management Commands:",
			Commands: []*cobra.Command{
				user.NewCmdUser(),
				secret.NewCmdSecret(),
				policy.NewCmdPolicy(),
				jwt.NewCmdJWT(),
			},
		},
		{
			Message: "Troubleshooting and Debugging Commands:",
			Commands: []*cobra.Command{
				validate.NewCmdValidate(),
			},
		},
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				completion.NewCmdCompletion(),
				options.NewCmdOptions(),
				version.NewCmdVersion(),
			},
		},
	}
	groups.Add(cmd)

	templates.ActsAsRootCommand(cmd, []string{"options"}, groups...)

	return cmd
}

func loadClient(_ *cobra.Command) error {
	viper.SetConfigFile(viper.GetString("config"))
	viper.SetConfigType("yaml")
	viper.SetDefault("server.address", "http://127.0.0.1:8080")
	viper.SetDefault("server.timeout", 10*time.Second)
	viper.SetDefault("server.insecure-skip-tls-verify", false)

	if configFile := viper.GetString("config"); configFile != "" {
		if _, err := os.Stat(configFile); err == nil {
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("read config file: %w", err)
			}
		}
	}

	c, err := client.New(client.Config{
		Server:                viper.GetString("server.address"),
		Username:              viper.GetString("user.username"),
		Password:              viper.GetString("user.password"),
		Token:                 viper.GetString("user.token"),
		SecretID:              viper.GetString("user.secret-id"),
		SecretKey:             viper.GetString("user.secret-key"),
		Timeout:               viper.GetDuration("server.timeout"),
		InsecureSkipTLSVerify: viper.GetBool("server.insecure-skip-tls-verify"),
		CertificateAuthority:  viper.GetString("server.certificate-authority"),
	})
	if err != nil {
		return err
	}

	cmdutil.SetClient(c)

	return nil
}

func bindFlag(flags *pflag.FlagSet, name, key string) {
	flag := flags.Lookup(name)
	if flag == nil {
		return
	}

	if err := viper.BindPFlag(key, flag); err != nil {
		cobra.CheckErr(err)
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfigName
	}

	return fmt.Sprintf("%s/%s/%s", home, defaultHomeDir, defaultConfigName)
}
