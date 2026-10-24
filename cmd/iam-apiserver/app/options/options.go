package options

import (
	"github.com/onexstack/onexstack/pkg/app"
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/onexdemo/iamv2/internal/apiserver"
)

// Ensure ServerOptions implements the app.FlagSetOptions interface.
var _ app.FlagSetOptions = (*ServerOptions)(nil)

// ServerOptions contains the configuration options for the apiserver.
type ServerOptions struct {
	SecureServingOptions   *genericoptions.SecureServingOptions   `json:"secure" mapstructure:"secure"`
	InsecureServingOptions *genericoptions.InsecureServingOptions `json:"insecure" mapstructure:"insecure"`
	MySQLOptions           *genericoptions.MySQLOptions           `json:"coredb" mapstructure:"coredb"`
	RedisOptions           *genericoptions.RedisOptions           `json:"redis" mapstructure:"redis"`
	JWTOptions             *genericoptions.JWTOptions             `json:"jwt" mapstructure:"jwt"`
	OTelOptions            *genericoptions.OTelOptions            `json:"otel" mapstructure:"otel"`

	AdminUsername string `json:"admin-username" mapstructure:"admin-username"`
	AdminPassword string `json:"admin-password" mapstructure:"admin-password"`
	AdminEmail    string `json:"admin-email" mapstructure:"admin-email"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		SecureServingOptions:   genericoptions.NewSecureServingOptions(),
		InsecureServingOptions: genericoptions.NewInsecureServingOptions(),
		MySQLOptions:           genericoptions.NewMySQLOptions(),
		RedisOptions:           genericoptions.NewRedisOptions(),
		JWTOptions:             genericoptions.NewJWTOptions(),
		OTelOptions:            genericoptions.NewOTelOptions(),
		AdminUsername:          "admin",
		AdminPassword:          "Admin@2020",
		AdminEmail:             "admin@foxmail.com",
	}
	opts.InsecureServingOptions.Addr = ":8080"

	return opts
}

// AddFlags binds the options to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs, "secure")
	o.InsecureServingOptions.AddFlags(fs, "insecure")
	o.MySQLOptions.AddFlags(fs, "coredb")
	o.RedisOptions.AddFlags(fs, "redis")
	o.JWTOptions.AddFlags(fs, "jwt")
	o.OTelOptions.AddFlags(fs, "otel")

	fs.StringVar(&o.AdminUsername, "admin-username", o.AdminUsername, "Bootstrap administrator username.")
	fs.StringVar(&o.AdminPassword, "admin-password", o.AdminPassword, "Bootstrap administrator password.")
	fs.StringVar(&o.AdminEmail, "admin-email", o.AdminEmail, "Bootstrap administrator email.")
}

// Complete completes all required options.
func (o *ServerOptions) Complete() error {
	return nil
}

// Validate checks whether the options are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}
	errs = append(errs, o.SecureServingOptions.Validate()...)
	errs = append(errs, o.InsecureServingOptions.Validate()...)
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.JWTOptions.Validate()...)
	errs = append(errs, o.OTelOptions.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// Config builds an apiserver.Config from the options.
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		SecureServingOptions:   o.SecureServingOptions,
		InsecureServingOptions: o.InsecureServingOptions,
		MySQLOptions:           o.MySQLOptions,
		RedisOptions:           o.RedisOptions,
		JWTOptions:             o.JWTOptions,
		AdminUsername:          o.AdminUsername,
		AdminPassword:          o.AdminPassword,
		AdminEmail:             o.AdminEmail,
	}, nil
}
