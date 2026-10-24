// Package watcher contains the iam-watcher plugins and their shared
// dependencies.
package watcher

import (
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/watcher/options"
)

// AggregateConfig is injected into every watcher that needs shared
// dependencies.
type AggregateConfig struct {
	Store   store.IStore
	Options *options.WatcherOptions
}
