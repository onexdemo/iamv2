package watcher

import (
	"github.com/onexstack/onexstack/pkg/watch/initializer"
	"github.com/onexstack/onexstack/pkg/watch/registry"
)

type initializerImpl struct {
	*Config
}

var _ initializer.WatcherInitializer = (*initializerImpl)(nil)

// NewInitializer creates an initializer that injects pump dependencies.
func NewInitializer(config *Config) *initializerImpl {
	return &initializerImpl{Config: config}
}

// Initialize configures a watcher that implements WantsConfig.
func (i *initializerImpl) Initialize(w registry.Watcher) {
	if wants, ok := w.(WantsConfig); ok {
		wants.SetPumpConfig(i.Config)
	}
}
