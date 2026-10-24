// Package options contains the configuration consumed by iam-watcher plugins.
package options

import "github.com/spf13/pflag"

// CleanOptions configures the policy-audit clean watcher.
type CleanOptions struct {
	// MaxReserveDays is the number of days policy audit rows are retained.
	MaxReserveDays int `json:"max-reserve-days" mapstructure:"max-reserve-days"`
}

// TaskOptions configures the account-maintenance watcher.
type TaskOptions struct {
	// MaxInactiveDays disables accounts that have not logged in for this many
	// days. Zero disables the check.
	MaxInactiveDays int `json:"max-inactive-days" mapstructure:"max-inactive-days"`
}

// WatcherOptions groups the configuration consumed by the registered watchers.
type WatcherOptions struct {
	Clean CleanOptions `json:"clean" mapstructure:"clean"`
	Task  TaskOptions  `json:"task" mapstructure:"task"`
}

// NewWatcherOptions returns the default watcher options.
func NewWatcherOptions() *WatcherOptions {
	return &WatcherOptions{
		Clean: CleanOptions{MaxReserveDays: 180},
		Task:  TaskOptions{MaxInactiveDays: 0},
	}
}

// AddFlags binds the watcher options to the given flag set.
func (o *WatcherOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(
		&o.Clean.MaxReserveDays,
		"watcher.clean.max-reserve-days",
		o.Clean.MaxReserveDays,
		"Policy audit log maximum retention days.",
	)
	fs.IntVar(
		&o.Task.MaxInactiveDays,
		"watcher.task.max-inactive-days",
		o.Task.MaxInactiveDays,
		"Maximum account inactivity in days before the account is disabled (0 disables the check).",
	)
}
