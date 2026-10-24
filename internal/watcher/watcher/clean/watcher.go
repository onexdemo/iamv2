// Package clean implements the policy-audit retention watcher.
package clean

import (
	"context"
	"log/slog"

	"github.com/onexstack/onexstack/pkg/watch/registry"

	watcherpkg "github.com/onexdemo/iamv2/internal/watcher/watcher"
)

// watcherSpec runs once a day.
const watcherSpec = "@daily"

type cleanWatcher struct {
	config *watcherpkg.AggregateConfig
}

var (
	_ registry.Watcher                = (*cleanWatcher)(nil)
	_ registry.ISpec                  = (*cleanWatcher)(nil)
	_ watcherpkg.WantsAggregateConfig = (*cleanWatcher)(nil)
)

// SetAggregateConfig injects the shared dependencies.
func (w *cleanWatcher) SetAggregateConfig(config *watcherpkg.AggregateConfig) {
	w.config = config
}

// Run removes policy audit rows older than the configured retention window.
func (w *cleanWatcher) Run() {
	if w.config == nil || w.config.Store == nil {
		slog.Warn("clean watcher is not initialized")

		return
	}

	ctx := context.Background()
	rows, err := w.config.Store.PolicyAudit().ClearOutdated(ctx, w.config.Options.Clean.MaxReserveDays)
	if err != nil {
		slog.ErrorContext(ctx, "failed to clean policy audit data", "error", err)

		return
	}

	slog.InfoContext(ctx, "cleaned policy audit data", "rows_affected", rows)
}

// Spec returns the cron schedule.
func (*cleanWatcher) Spec() string {
	return watcherSpec
}

func init() {
	registry.Register("clean", &cleanWatcher{})
}
