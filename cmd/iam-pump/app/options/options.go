package options

import (
	"github.com/onexstack/onexstack/pkg/app"
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/watch"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/onexdemo/iamv2/internal/pump"
)

// Ensure ServerOptions implements the app.FlagSetOptions interface.
var _ app.FlagSetOptions = (*ServerOptions)(nil)

// ServerOptions contains the configuration options for the pump.
type ServerOptions struct {
	WatchOptions *watch.Options              `json:"watch" mapstructure:"watch"`
	RedisOptions *genericoptions.RedisOptions `json:"redis" mapstructure:"redis"`
	PumpOptions  *pump.Options                `json:"pump" mapstructure:"pump"`
	SlogOptions  *genericoptions.SlogOptions  `json:"slog" mapstructure:"slog"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		WatchOptions: watch.NewOptions(),
		RedisOptions: genericoptions.NewRedisOptions(),
		PumpOptions:  pump.NewOptions(),
		SlogOptions:  genericoptions.NewSlogOptions(),
	}
}

// AddFlags binds the options to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.WatchOptions.AddFlags(fs, "watch")
	o.RedisOptions.AddFlags(fs, "redis")
	o.PumpOptions.AddFlags(fs)
	o.SlogOptions.AddFlags(fs, "slog")
}

// Complete completes all required options.
func (o *ServerOptions) Complete() error {
	return nil
}

// Validate checks whether the options are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}
	errs = append(errs, o.WatchOptions.Validate()...)
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.PumpOptions.Validate()...)
	errs = append(errs, o.SlogOptions.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// Config builds a pump.Config from the options.
func (o *ServerOptions) Config() (*pump.Config, error) {
	return &pump.Config{
		WatchOptions: o.WatchOptions,
		RedisOptions: o.RedisOptions,
		PumpOptions:  o.PumpOptions,
	}, nil
}
