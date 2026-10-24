package watcher

import (
	"github.com/onexstack/onexstack/pkg/watch/initializer"
	"github.com/onexstack/onexstack/pkg/watch/registry"
)

// WatcherInitializer injects iam-specific dependencies into watcher plugins.
type WatcherInitializer struct {
	*AggregateConfig
}

var _ initializer.WatcherInitializer = (*WatcherInitializer)(nil)

// NewInitializer creates a WatcherInitializer.
func NewInitializer(aggregate *AggregateConfig) *WatcherInitializer {
	return &WatcherInitializer{AggregateConfig: aggregate}
}

// Initialize configures a watcher that implements WantsAggregateConfig.
func (w *WatcherInitializer) Initialize(wc registry.Watcher) {
	if wants, ok := wc.(WantsAggregateConfig); ok {
		wants.SetAggregateConfig(w.AggregateConfig)
	}
}
