package watcher

import (
	"github.com/onexstack/onexstack/pkg/watch/registry"
)

// WantsAggregateConfig is implemented by watchers that need shared iam
// dependencies.
type WantsAggregateConfig interface {
	registry.Watcher
	SetAggregateConfig(config *AggregateConfig)
}
