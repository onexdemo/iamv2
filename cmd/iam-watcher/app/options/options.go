package options

import (
	"github.com/onexstack/onexstack/pkg/app"
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/watch"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/onexdemo/iamv2/internal/watcher"
	watcheroptions "github.com/onexdemo/iamv2/internal/watcher/options"
)

// Ensure ServerOptions implements the app.FlagSetOptions interface.
var _ app.FlagSetOptions = (*ServerOptions)(nil)

// ServerOptions contains the configuration options for the watcher.
type ServerOptions struct {
	WatchOptions   *watch.Options                    `json:"watch" mapstructure:"watch"`
	MySQLOptions   *genericoptions.MySQLOptions      `json:"coredb" mapstructure:"coredb"`
	WatcherOptions *watcheroptions.WatcherOptions    `json:"watcher" mapstructure:"watcher"`
	SlogOptions    *genericoptions.SlogOptions       `json:"slog" mapstructure:"slog"`
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		WatchOptions:   watch.NewOptions(),
		MySQLOptions:   genericoptions.NewMySQLOptions(),
		WatcherOptions: watcheroptions.NewWatcherOptions(),
		SlogOptions:    genericoptions.NewSlogOptions(),
	}
}

// AddFlags binds the options to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.WatchOptions.AddFlags(fs, "watch")
	o.MySQLOptions.AddFlags(fs, "coredb")
	o.WatcherOptions.AddFlags(fs)
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
	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.SlogOptions.Validate()...)

	return utilerrors.NewAggregate(errs)
}

// Config builds a watcher.Config from the options.
func (o *ServerOptions) Config() (*watcher.Config, error) {
	return &watcher.Config{
		WatchOptions:   o.WatchOptions,
		MySQLOptions:   o.MySQLOptions,
		WatcherOptions: o.WatcherOptions,
	}, nil
}
