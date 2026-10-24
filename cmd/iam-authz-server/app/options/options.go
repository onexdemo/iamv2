package options

import (
	"errors"
	"time"

	"github.com/onexstack/onexstack/pkg/app"
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/onexdemo/iamv2/internal/authz"
	"github.com/onexdemo/iamv2/internal/authz/pkg/analytics"
)

// Ensure ServerOptions implements the app.FlagSetOptions interface.
var _ app.FlagSetOptions = (*ServerOptions)(nil)

// ServerOptions contains the configuration options for the authz server.
type ServerOptions struct {
	SecureServingOptions   *genericoptions.SecureServingOptions   `json:"secure" mapstructure:"secure"`
	InsecureServingOptions *genericoptions.InsecureServingOptions `json:"insecure" mapstructure:"insecure"`
	RedisOptions           *genericoptions.RedisOptions           `json:"redis" mapstructure:"redis"`
	AnalyticsOptions       *analytics.Options                     `json:"analytics" mapstructure:"analytics"`
	OTelOptions            *genericoptions.OTelOptions            `json:"otel" mapstructure:"otel"`

	APIServerAddr       string        `json:"apiserver-addr" mapstructure:"apiserver-addr"`
	APIServerUsername   string        `json:"apiserver-username" mapstructure:"apiserver-username"`
	APIServerPassword   string        `json:"apiserver-password" mapstructure:"apiserver-password"`
	APIServerTimeout    time.Duration `json:"apiserver-timeout" mapstructure:"apiserver-timeout"`
	CacheReloadInterval time.Duration `json:"cache-reload-interval" mapstructure:"cache-reload-interval"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		SecureServingOptions:   genericoptions.NewSecureServingOptions(),
		InsecureServingOptions: genericoptions.NewInsecureServingOptions(),
		RedisOptions:           genericoptions.NewRedisOptions(),
		AnalyticsOptions:       analytics.NewOptions(),
		OTelOptions:            genericoptions.NewOTelOptions(),
		APIServerAddr:          "http://127.0.0.1:8080",
		APIServerTimeout:       10 * time.Second,
		CacheReloadInterval:    time.Minute,
	}
	opts.InsecureServingOptions.Addr = ":8081"

	return opts
}

// AddFlags binds the options to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs, "secure")
	o.InsecureServingOptions.AddFlags(fs, "insecure")
	o.RedisOptions.AddFlags(fs, "redis")
	o.AnalyticsOptions.AddFlags(fs)
	o.OTelOptions.AddFlags(fs, "otel")

	fs.StringVar(&o.APIServerAddr, "apiserver-addr", o.APIServerAddr, "Base URL of iam-apiserver.")
	fs.StringVar(&o.APIServerUsername, "apiserver-username", o.APIServerUsername, "Administrator username used to fetch cache snapshots.")
	fs.StringVar(&o.APIServerPassword, "apiserver-password", o.APIServerPassword, "Administrator password used to fetch cache snapshots.")
	fs.DurationVar(&o.APIServerTimeout, "apiserver-timeout", o.APIServerTimeout, "Timeout for cache snapshot requests.")
	fs.DurationVar(&o.CacheReloadInterval, "cache-reload-interval", o.CacheReloadInterval, "Fallback cache reload interval.")
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
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.AnalyticsOptions.Validate()...)
	errs = append(errs, o.OTelOptions.Validate()...)

	if o.APIServerAddr == "" {
		errs = append(errs, errAPIServerAddrRequired)
	}

	return utilerrors.NewAggregate(errs)
}

// Config builds an authz.Config from the options.
func (o *ServerOptions) Config() (*authz.Config, error) {
	return &authz.Config{
		SecureServingOptions:   o.SecureServingOptions,
		InsecureServingOptions: o.InsecureServingOptions,
		RedisOptions:           o.RedisOptions,
		AnalyticsOptions:       o.AnalyticsOptions,
		APIServerAddr:          o.APIServerAddr,
		APIServerUsername:      o.APIServerUsername,
		APIServerPassword:      o.APIServerPassword,
		APIServerTimeout:       o.APIServerTimeout,
		CacheReloadInterval:    o.CacheReloadInterval,
	}, nil
}

var errAPIServerAddrRequired = errors.New("apiserver-addr must not be empty")
