package analytics

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// Options controls how authorization audit records are buffered and written to
// Redis.
type Options struct {
	Enable                  bool          `json:"enable" mapstructure:"enable"`
	PoolSize                int           `json:"pool-size" mapstructure:"pool-size"`
	RecordsBufferSize       uint64        `json:"records-buffer-size" mapstructure:"records-buffer-size"`
	FlushInterval           uint64        `json:"flush-interval" mapstructure:"flush-interval"`
	EnableDetailedRecording bool          `json:"enable-detailed-recording" mapstructure:"enable-detailed-recording"`
	StorageExpirationTime   time.Duration `json:"storage-expiration-time" mapstructure:"storage-expiration-time"`
}

// NewOptions returns the default analytics options.
func NewOptions() *Options {
	return &Options{
		Enable:                  true,
		PoolSize:                50,
		RecordsBufferSize:       2000,
		FlushInterval:           200,
		EnableDetailedRecording: true,
		StorageExpirationTime:   24 * time.Hour,
	}
}

// AddFlags binds the options to the given flag set.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enable, "analytics.enable", o.Enable, "Record authorization audit data.")
	fs.IntVar(&o.PoolSize, "analytics.pool-size", o.PoolSize, "Number of analytics workers.")
	fs.Uint64Var(&o.RecordsBufferSize, "analytics.records-buffer-size", o.RecordsBufferSize, "Buffer size for analytics workers.")
	fs.Uint64Var(&o.FlushInterval, "analytics.flush-interval", o.FlushInterval, "Flush interval in milliseconds (1-1000).")
	fs.BoolVar(
		&o.EnableDetailedRecording,
		"analytics.enable-detailed-recording",
		o.EnableDetailedRecording,
		"Record detailed policy and decider information.",
	)
	fs.DurationVar(
		&o.StorageExpirationTime,
		"analytics.storage-expiration-time",
		o.StorageExpirationTime,
		"Retention of analytics data in Redis.",
	)
}

// Validate checks the analytics options.
func (o *Options) Validate() []error {
	if o == nil || !o.Enable {
		return nil
	}

	errs := []error{}
	if o.FlushInterval < 1 || o.FlushInterval > 1000 {
		errs = append(errs, fmt.Errorf("--analytics.flush-interval %v must be between 1 and 1000", o.FlushInterval))
	}

	if o.PoolSize <= 0 {
		errs = append(errs, fmt.Errorf("--analytics.pool-size must be greater than 0"))
	}

	if o.RecordsBufferSize < uint64(o.PoolSize) {
		errs = append(errs, fmt.Errorf("--analytics.records-buffer-size must be greater than or equal to pool size"))
	}

	return errs
}
