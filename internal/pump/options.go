package pump

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/onexdemo/iamv2/internal/pump/pumps"
)

// Options controls how analytics records are drained from Redis and fanned out
// to the configured pumps.
type Options struct {
	// PurgeDelay is the interval in seconds between drain cycles.
	PurgeDelay int `json:"purge-delay" mapstructure:"purge-delay"`
	// OmitDetailedRecording strips policy details from every record.
	OmitDetailedRecording bool `json:"omit-detailed-recording" mapstructure:"omit-detailed-recording"`
	// Pumps maps a pump key to its configuration.
	Pumps map[string]pumps.Config `json:"pumps" mapstructure:"pumps"`
}

// NewOptions returns the default pump options.
func NewOptions() *Options {
	return &Options{
		PurgeDelay: 10,
		Pumps:      map[string]pumps.Config{},
	}
}

// AddFlags binds the options to the given flag set.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.PurgeDelay, "purge-delay", o.PurgeDelay, "Interval in seconds between analytics drain cycles.")
	fs.BoolVar(
		&o.OmitDetailedRecording,
		"omit-detailed-recording",
		o.OmitDetailedRecording,
		"Strip policy and decider details from records before writing.",
	)
}

// Validate checks the pump options.
func (o *Options) Validate() []error {
	errs := []error{}
	if o.PurgeDelay <= 0 {
		errs = append(errs, fmt.Errorf("--purge-delay must be greater than 0"))
	}

	return errs
}
